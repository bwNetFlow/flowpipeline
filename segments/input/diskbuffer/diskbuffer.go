package diskbuffer

import (
	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
	"github.com/dustin/go-humanize"
	"github.com/klauspost/compress/zstd"
	"google.golang.org/protobuf/encoding/protojson"
	"github.com/google/uuid"
	"golang.org/x/sys/unix"
	"log"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"fmt"
	"os"
	"bufio"
	"errors"
	"io"
)

const (
	defaultQueueSize           = 65536
	defaultBatchSize           = 128
	defaultQueueStatusInterval = 0 * time.Second
	defaultHighMemoryMark      = 70
	defaultLowMemoryMark       = 30
	defaultReadingMemoryMark   = 5
	defaultFileSize            = 50 * humanize.MByte
	defaultMaxCacheSize        = 1 * humanize.GByte
)

type DiskBuffer struct {
	segments.BaseSegment
	BatchSize           int
	BatchDebugPrintf    func(format string, v ...any)
	QueueStatusInterval time.Duration
	MemoryBuffer        chan *pb.EnrichedFlow
	FileSize            uint64
	BufferDir           string
	HighMemoryMark      int
	LowMemoryMark       int
	ReadingMemoryMark   int
	MaxCacheSize        uint64
	Capacity            int
}

func NoDebugPrintf(format string, v ...any) {}
func DoDebugPrintf(format string, v ...any) {
	log.Printf(format, v...)
}

func (segment *DiskBuffer) New(config map[string]string) segments.Segment {
	var (
		err                error
		buflen             int
	)

	segment.BufferDir = config["bufferdir"]
	if segment.BufferDir != "" {
		fi, err := os.Stat(segment.BufferDir)
		if err != nil {
			log.Fatalf("[error] Diskbuffer: Could not obtain file info for file %s", segment.BufferDir)
		}
		if !fi.IsDir() {
			log.Fatalf("[error] Diskbuffer: bufferdir %s must be a directory", segment.BufferDir)
		}
		if unix.Access(segment.BufferDir, unix.W_OK) != nil {
			log.Fatal("[error] Diskbuffer: bufferdir must be writeable")
		}
	} else {
		log.Fatal("[error] Diskbuffer: bufferdir must exist")
	}
	// parse HighMemoryMark option	
	segment.HighMemoryMark = defaultHighMemoryMark
	if config["highmemorymark"] != "" {
		segment.HighMemoryMark, err = strconv.Atoi(config["highmemorymark"])
		if err != nil {
			log.Fatalf("[error] Diskbuffer: Failed to parse highmemorymark config option: %s", err)
		}
		if segment.HighMemoryMark < 10 || segment.HighMemoryMark > 95 {
			log.Fatal("[error] Diskbuffer: HighMemoryMark must be between 10 and 95")
		}
	}

	segment.ReadingMemoryMark = defaultReadingMemoryMark
	if config["readingmemorymark"] != "" {
		segment.ReadingMemoryMark, err = strconv.Atoi(config["highmemorymark"])
		if err != nil {
			log.Fatalf("[error] Diskbuffer: Failed to parse readingmemorymark config option: %s", err)
		}
		if segment.ReadingMemoryMark < 1 || segment.ReadingMemoryMark > 50 {
			log.Fatal("[error] Diskbuffer: HighMemoryMark must be between 1 and 50")
		}

	}	

	// parse LowMemoryMark option	
	segment.LowMemoryMark = defaultLowMemoryMark
	if config["lowmemorymark"] != "" {
		segment.LowMemoryMark, err = strconv.Atoi(config["lowmemorymark"])
		if err != nil {
			log.Fatalf("[error] Diskbuffer: Failed to parse lowmemorymark config option: %s", err)
		}
		if segment.LowMemoryMark < 5 || segment.LowMemoryMark > 70 {
			log.Fatal("[error] Diskbuffer: HighMemoryMark must be between 5 and 70")
		}
	}

	//sanity check: lowmemorymark < highmemorymark
	if segment.LowMemoryMark > segment.HighMemoryMark {
		log.Fatal("[error] Diskbuffer: HighMemoryMark must be greater than LowMemoryMark")
	}
	if segment.ReadingMemoryMark > segment.LowMemoryMark {
		log.Fatal("[error] Diskbuffer: LowMemoryMark must be greater than ReadingMemoryMark")
	}
	
	segment.MaxCacheSize = defaultMaxCacheSize
	if config["maxcachesize"] != "" {
		segment.FileSize, err = humanize.ParseBytes(config["maxcachesize"])
		if err != nil {
			log.Fatalf("[error] Diskbuffer: Failed to parse maxcachesize config option: %s", err)
		}
	}	


	// parse filesize option	
	segment.FileSize = defaultFileSize
	if config["filesize"] != "" {
		segment.FileSize, err = humanize.ParseBytes(config["filesize"])
		if err != nil {
			log.Fatalf("[error] Diskbuffer: Failed to parse filesize config option: %s", err)
		}
	}

	// parse batchSize option
	segment.BatchSize = defaultBatchSize
	if config["batchsize"] != "" {
		segment.BatchSize, err = strconv.Atoi(config["batchsize"])
		if err != nil {
			log.Fatalf("[error] Diskbuffer: Failed to parse batchsize config option: %s", err)
		}
	}
	if segment.BatchSize < 0 {
		segment.BatchSize = defaultBatchSize
	}
	// parse batchdebug option
	if config["batchdebug"] != "" {
		batchDebug, err := strconv.ParseBool(config["batchdebug"])
		if err != nil {
			log.Fatalf("[error] Diskbuffer: Failed to parse batchdebug config option: %s", err)
		}
		// set proper BatchDebugPrintf function
		if batchDebug {
			segment.BatchDebugPrintf = DoDebugPrintf
		} else {
			segment.BatchDebugPrintf = NoDebugPrintf
		}
	}
	
	// parse queueStatusInterval option
	segment.QueueStatusInterval = defaultQueueStatusInterval
	if config["queuestatusinterval"] != "" {
		segment.QueueStatusInterval, err = time.ParseDuration(config["queuestatusinterval"])
		if err != nil {
			log.Fatalf("[error] Diskbuffer: Failed to parse queuestatussnterval config option: %s", err)
		}
	}

	// create buffered channel
	if config["queuesize"] != "" {
		buflen, err = strconv.Atoi(config["queuesize"])
		if err != nil {
			log.Fatalf("[error] Diskbuffer: Failed to parse queuesize config option: %s", err)
		}
	} else {
		buflen = defaultQueueSize
	}
	if buflen < 64 {
		log.Printf("[error] Diskbuffer: queuesize too small, using default %d", defaultQueueSize)
		buflen = defaultQueueSize
	}
	segment.MemoryBuffer = make(chan *pb.EnrichedFlow, buflen)
	segment.Capacity = cap(segment.MemoryBuffer)
	return segment
}

func WatchCacheFiles(segment *DiskBuffer, BufferWG *sync.WaitGroup, Signal chan struct{}, CacheFiles *[]string) {
	defer BufferWG.Done()
	var err error

	for {
		pattern := fmt.Sprintf("%s/*.json.zst", segment.BufferDir)
		*CacheFiles, err = filepath.Glob(pattern)
		if err != nil {
			log.Fatalf("[error] Diskbuffer: Failed with filepath glob: %s", err)
		}
		// sum sizes
		var CacheFilesSize int64 = 0
		for _, filename := range *CacheFiles {
			fi, err := os.Stat(filename)
			if err != nil {
				log.Printf("[warning] Diskbuffer: Could not obtain file info for file %s", filename)
			}
			CacheFilesSize += fi.Size()
		}

		select {
		case <- Signal:
			return
		case <- time.After(10 * time.Second):
		}
	}
}

func WriteWatchdogLowMemoryMark(segment *DiskBuffer, ReadWriteWG *sync.WaitGroup, Signal chan struct{}, StopDecider chan struct{}, StopWritingToDisk chan struct{}) {
	defer ReadWriteWG.Done()
	for {
		select {
		case <- Signal:
			return
		case <- StopDecider:
			time.Sleep(100 * time.Millisecond)
		default:	
			length := len(segment.MemoryBuffer)
			if length < segment.LowMemoryMark * segment.Capacity / 100 {
				close(StopWritingToDisk)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
func WriteToDisk(segment *DiskBuffer, ReadWriteWG *sync.WaitGroup, Signal chan struct{}, Watchdogs chan struct{}) {
	defer close(Watchdogs)
	defer ReadWriteWG.Done()
	
	log.Print("[debug] Diskbuffer: Started Writing to Disk")
	defer log.Print("[debug] Diskbuffer: Ended Writing to Disk")

	// we need a new filename
	filename := fmt.Sprintf("%s/%s.json.zst", segment.BufferDir, uuid.NewString())

	file, err := os.Create(filename)
	if err != nil {
		log.Printf("[error] Diskbuffer: File specified in 'filename' is not accessible: %s", err)
	}
	level := zstd.SpeedFastest
	encoder, err := zstd.NewWriter(file, zstd.WithEncoderLevel(level))
	if err != nil {
		log.Fatalf("[error] Diskbuffer: error creating zstd encoder: %s", err)
	}
	writer := bufio.NewWriterSize(encoder, 65536)

	defer file.Close()
	defer encoder.Close()
	defer writer.Flush() 

	for {
		select {
		case <-Signal:	
			return
		default:
			for i := 0; i < segment.BatchSize; i++ {
				select {
				case msg := <- segment.MemoryBuffer:
					data, err := protojson.Marshal(msg)
					if err != nil {
						log.Printf("[warning] Diskbuffer: Skipping a flow, failed to recode protobuf as JSON: %v", err)
						continue
					}
			
					// use Fprintln because it adds an OS specific newline
					_, err = fmt.Fprintln(writer, string(data))
					if err != nil {
						log.Printf("[warning] Diskbuffer: Skipping a flow, failed to write to file %s: %v", filename, err)
						continue
					}
				default:
					// MemoryBuffer is empty -> no need to write anyhing to disk
					return
				}
			}
			fi, err := file.Stat()
			if err != nil {
				log.Printf("[warning] Diskbuffer: Could not obtain file info for file %s", filename)
			}
			if uint64(fi.Size()) > segment.FileSize {
				log.Printf("[debug] Diskbuffer: File %s is bigger than %d, stopping write", filename, segment.FileSize)
				break
			}
		}
	}
}
func ReadWatchdogLowMemoryMark(segment *DiskBuffer, ReadWriteWG *sync.WaitGroup, Signal chan struct{}, StopReadingFromDisk chan struct{}) {
	defer ReadWriteWG.Done()
	for {
		select {
		case <- Signal:
			return
		default:
			length := len(segment.MemoryBuffer)
			if length > segment.LowMemoryMark * segment.Capacity / 100 {
				close(StopReadingFromDisk)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
func ReadWatchdogHighMemoryMark(segment *DiskBuffer, ReadWriteWG *sync.WaitGroup, Signal chan struct{}, StopDecider chan struct{}, EmergencyStopReadingFromDisk chan struct{}) {
	defer ReadWriteWG.Done()
	for {
		select {
		case <- Signal:
			return
		case <- StopDecider:
			close(EmergencyStopReadingFromDisk)
			return
		default:
			length := len(segment.MemoryBuffer)
			if length > segment.HighMemoryMark * segment.Capacity / 100 {
				close(EmergencyStopReadingFromDisk)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}	
	}
}
	
func ReadFromDisk(segment *DiskBuffer, ReadWriteWG *sync.WaitGroup, Signal chan struct{}, EmergencySignal chan struct{}, Watchdogs chan struct{}, CacheFiles *[]string) {
	defer close(Watchdogs)
	defer ReadWriteWG.Done()
	
	log.Print("[debug] Diskbuffer: Started Reading from Disk")
	defer log.Print("[debug] Diskbuffer: Ended Reading from Disk")

	// the ReadWriteWG ensures, that there are no Read and Write at the same time.
	// hence we can read from every file that exists
	fromReader := make(chan []byte)
	go func() {
		for _, filename := range *CacheFiles {
			file, err := os.Open(filename)
			
			if err != nil {
				log.Printf("[warning] Diskbuffer: Could not open file: %s, with error %s", filename, err)
				continue
			}

			var decoder, _ = zstd.NewReader(file)
			scanner := bufio.NewScanner(decoder)
			for { 
				scan := scanner.Scan()
				err := scanner.Err()
				if errors.Is(err, io.ErrUnexpectedEOF) {
					log.Printf("[warning] Diskbuffer: Unexpected EOF: %v", err)
					break
				}		
				if err != nil {
					log.Printf("[warning] Diskbuffer: Skipping a flow, could not read line from stdin: %v", err)
					continue
				}
				if !scan && scanner.Err() == nil {
					break
				}
				if len(scanner.Text()) == 0 {
					continue
				}
				fromReader <- []byte(scanner.Text())
			}
			// end-of-file: delete it
			file.Close()
			err = os.Remove(filename)
			if err != nil {
					log.Printf("[warning] Diskbuffer: Could not remove file %s with error %s", filename, err)
			}
			// check signal channel, if we have a signal, do not read any new file
			select {
				case <- Signal:
					close(fromReader)
					return
				case <- EmergencySignal:
					close(fromReader)
					return
				default:
			}
		}
		close(fromReader)
	}()
	for line := range fromReader {
		select {
		case <- EmergencySignal:
			log.Print("[warning] Diskbuffer: While reading from disk, got into high watermark")
			// we are in high watermark again
			// write every line in a new file, then stop reading
			filename := fmt.Sprintf("%s/rest_%s.json.zst", segment.BufferDir, uuid.NewString())
			file, err := os.Create(filename)
			if err != nil {
					log.Printf("[error] Diskbuffer: File specified in 'filename' is not accessible: %s", err)
			}
			level := zstd.SpeedDefault
			encoder, err := zstd.NewWriter(file, zstd.WithEncoderLevel(level))
			if err != nil {
				log.Fatalf("[error] Diskbuffer: error creating zstd encoder: %s", err)
			}
			writer := bufio.NewWriter(encoder)

			defer file.Close()
			defer encoder.Close()
			defer writer.Flush() 

			for emerg_line := range fromReader {
				// use Fprintln because it adds an OS specific newline
				_, err = fmt.Fprintln(writer, emerg_line)
				if err != nil {
					log.Printf("[warning] Diskbuffer: Skipping a flow, failed to write to file %s: %v", filename, err)
					continue
				}
			}
		default:
			msg := &pb.EnrichedFlow{}
			err := protojson.Unmarshal(line, msg)
			if err != nil {
				log.Printf("[warning] Diskbuffer: Skipping a flow, failed to recode input to protobuf: %v", err)
				continue
			}
			select {
				case segment.Out <- msg:
				case <- time.After(10* time.Millisecond):
					segment.MemoryBuffer <- msg
			}
		}
	}
}
func QueueStatus(segment *DiskBuffer, BufferWG *sync.WaitGroup, StopQueueStatusInterval chan struct{}) {
	defer BufferWG.Done()	
	for {
		select {
		case <- StopQueueStatusInterval:
			return
		case <- time.After(segment.QueueStatusInterval):
			fill := len(segment.MemoryBuffer)
			log.Printf("[debug] Diskbuffer: Queue is %3.2f%% full (%d/%d)", float64(fill)/float64(segment.Capacity)*100, fill, segment.Capacity)
		}
	}
}

func (segment *DiskBuffer) Run(wg *sync.WaitGroup) {
	var BufferWG sync.WaitGroup
	var ReadWriteWG sync.WaitGroup
	var CacheFiles []string
	var CacheFilesSize int64 = 0
	StopDecider := make(chan struct{})

	FuncWatchCacheFiles := func(Signal chan struct{}) {
		WatchCacheFiles(segment, &BufferWG, Signal, &CacheFiles)
	}

	FuncWriteWatchdogLowMemoryMark := func(Signal chan struct{}, StopWritingToDisk chan struct{}) {
		WriteWatchdogLowMemoryMark(segment, &ReadWriteWG, Signal, StopDecider, StopWritingToDisk)
	}

	FuncWriteToDisk := func(Signal chan struct{}, Watchdogs chan struct{}) {
		WriteToDisk(segment, &ReadWriteWG, Signal, Watchdogs)
	}

	FuncReadWatchdogLowMemoryMark := func(Signal chan struct{}, StopReadingFromDisk chan struct{}) {
		ReadWatchdogLowMemoryMark(segment, &ReadWriteWG, Signal, StopReadingFromDisk)
	}
	FuncReadWatchdogHighMemoryMark := func(Signal chan struct{}, EmergencyStopReadingFromDisk chan struct{}) {
		ReadWatchdogHighMemoryMark(segment, &ReadWriteWG, Signal, StopDecider, EmergencyStopReadingFromDisk)
	}
	FuncReadFromDisk := func(Signal chan struct{}, EmergencySignal chan struct{}, Watchdogs chan struct{}) {
		ReadFromDisk(segment, &ReadWriteWG, Signal, EmergencySignal, Watchdogs, &CacheFiles)
	}
	defer func() {
		close(segment.Out)
		wg.Done()
		log.Println("[info] Diskbuffer: All writer functions have stopped, exiting…")
	}()
	defer BufferWG.Wait()	

	// print queue status information
	StopQueueStatusInterval := make(chan struct{})
	if segment.QueueStatusInterval > 0 {
		BufferWG.Add(1)
		go QueueStatus(segment, &BufferWG, StopQueueStatusInterval)
	}

	StopWritingNextSegment := make(chan struct{})
	StopCacheFileWatcher   := make(chan struct{})
	// read from input into memory buffer
	BufferWG.Add(1)
	go func() {
		defer BufferWG.Done()
		for msg := range segment.In {
			segment.MemoryBuffer <- msg
		}
		close(StopQueueStatusInterval)
		close(StopCacheFileWatcher)
		close(StopWritingNextSegment)
		close(StopDecider)
	}()

	// write into next segment
	BufferWG.Add(1)
	go func() {
		defer BufferWG.Done()
		for {
			select {
			case <- StopWritingNextSegment:
				return
			default:
				msg := <-segment.MemoryBuffer
				segment.Out <- msg
			}
		}
	}()

	BufferWG.Add(1)
	go FuncWatchCacheFiles(StopCacheFileWatcher)

	// decider if we should write into compressed files
	BufferWG.Add(1)
	go func() {
		defer BufferWG.Done()
		defer log.Print("[debug] Diskbuffer: Stopping Decider")
		log.Print("[debug] Diskbuffer: Starting Decider")
		for {
			select {
			case <- StopDecider:
				ReadWriteWG.Add(1)
				StopWritingToDisk := make(chan struct{})
				StopWatchdogs := make(chan struct{})
				go FuncWriteToDisk(StopWritingToDisk, StopWatchdogs)
				ReadWriteWG.Wait()
				return
			default:
				length := len(segment.MemoryBuffer)
				
				if length < segment.ReadingMemoryMark * segment.Capacity / 100 && len(CacheFiles) > 0 {
					ReadWriteWG.Wait()
					ReadWriteWG.Add(3)

					StopReadingFromDisk := make(chan struct{})
					EmergencyStopReadingFromDisk := make(chan struct{})
					StopWatchdogs := make(chan struct{})
					go FuncReadWatchdogLowMemoryMark(StopWatchdogs, StopReadingFromDisk)
					go FuncReadWatchdogHighMemoryMark(StopWatchdogs, EmergencyStopReadingFromDisk)
					go FuncReadFromDisk(StopReadingFromDisk, EmergencyStopReadingFromDisk, StopWatchdogs)
					ReadWriteWG.Wait()
				}
				if length > segment.HighMemoryMark * segment.Capacity / 100 && uint64(CacheFilesSize) < segment.MaxCacheSize {
					log.Print("[debug] Diskbuffer: Try to buffer to disk")
					// start new go routine
					ReadWriteWG.Wait()
					ReadWriteWG.Add(2)
					StopWritingToDisk := make(chan struct{})
					StopWatchdogs := make(chan struct{})
					go FuncWriteWatchdogLowMemoryMark(StopWatchdogs, StopWritingToDisk)	
					go FuncWriteToDisk(StopWritingToDisk, StopWatchdogs)
					ReadWriteWG.Wait()
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

// register segment
func init() {
	segment := &DiskBuffer{}
	segments.RegisterSegment("diskbuffer", segment)
}
