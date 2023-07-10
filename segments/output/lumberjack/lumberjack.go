// Send passing flows to one or more lumberjack (Elastic Beats) servers.
package lumberjack

import (
	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultQueueSize           = 65536
	defaultBatchSize           = 64
	defaultTimeout             = 5 * time.Second
	minimalBatchTimeout        = 50 * time.Millisecond
	defaultQueueStatusInterval = 0 * time.Second
	defaultReconnectWait       = 1 * time.Second
)

type ServerOptions struct {
	UseTLS            bool
	VerifyCertificate bool
	CompressionLevel  int
}

type Lumberjack struct {
	segments.BaseSegment
	Servers             map[string]ServerOptions
	BatchSize           int
	BatchTimeout        time.Duration
	BatchDebugPrintf    func(format string, v ...any)
	QueueStatusInterval time.Duration
	ReconnectWait       time.Duration
	LumberjackOut       chan *pb.EnrichedFlow
}

func NoDebugPrintf(format string, v ...any) {}
func DoDebugPrintf(format string, v ...any) {
	log.Printf(format, v...)
}

func (segment *Lumberjack) New(config map[string]string) segments.Segment {
	var (
		err                error
		buflen             int
		defaultCompression int
	)

	// parse default compression level
	defaultCompressionString := config["compression"]
	if defaultCompressionString == "" {
		defaultCompression = 0
	} else {
		defaultCompression, err = strconv.Atoi(defaultCompressionString)
		if err != nil {
			log.Fatalf("[error] Lumberjack: Failed to parse default compression level %s: %s", defaultCompressionString, err)
		}
		if defaultCompression < 0 || defaultCompression > 9 {
			log.Fatalf("[error] Lumberjack: Default compression level %d is out of range", defaultCompression)
		}
	}

	// parse server URLs
	rawServerStrings := strings.Split(config["servers"], ",")
	for idx, serverName := range rawServerStrings {
		rawServerStrings[idx] = strings.TrimSpace(serverName)
	}
	if len(rawServerStrings) == 0 {
		log.Fatal("[error] Lumberjack: No servers specified in 'servers' config option.")
	} else {
		segment.Servers = make(map[string]ServerOptions)
		for _, rawServerString := range rawServerStrings {
			serverURL, err := url.Parse(rawServerString)
			if err != nil {
				log.Fatalf("[error] Lumberjack: Failed to parse server URL %s: %s", rawServerString, err)
			}
			urlQueryParams := serverURL.Query()

			// parse TLS options
			var useTLS, verifyTLS bool
			switch serverURL.Scheme {
			case "tcp":
				useTLS = false
				verifyTLS = false
			case "tls":
				useTLS = true
				verifyTLS = true
			case "tlsnoverify":
				useTLS = true
				verifyTLS = false
			default:
				log.Fatalf("[error] Lumberjack: Unknown scheme %s in server URL %s", serverURL.Scheme, rawServerString)
			}

			// parse compression level
			var compressionLevel int
			compressionString := urlQueryParams.Get("compression")

			if compressionString == "" {
				// use global default if not specified
				compressionLevel = defaultCompression
			} else {
				compressionLevel, err = strconv.Atoi(compressionString)
				if err != nil {
					log.Fatalf("[error] Lumberjack: Failed to parse compression level %s for host %s: %s", compressionString, serverURL.Host, err)
				}
				if compressionLevel < 0 || compressionLevel > 9 {
					log.Fatalf("[error] Lumberjack: Compression level %d out of range for host %s", compressionLevel, serverURL.Host)
				}
			}

			segment.Servers[serverURL.Host] = ServerOptions{
				UseTLS:            useTLS,
				VerifyCertificate: verifyTLS,
				CompressionLevel:  compressionLevel,
			}
		}
	}

	// parse batchSize option
	segment.BatchSize = defaultBatchSize
	if config["batchsize"] != "" {
		segment.BatchSize, err = strconv.Atoi(config["batchsize"])
		if err != nil {
			log.Fatalf("[error] Lumberjack: Failed to parse batchsize config option: %s", err)
		}
	}
	if segment.BatchSize < 0 {
		segment.BatchSize = defaultBatchSize
	}
	// parse batchtimeout option
	segment.BatchTimeout = defaultTimeout
	if config["batchtimeout"] != "" {
		segment.BatchTimeout, err = time.ParseDuration(config["batchtimeout"])
		if err != nil {
			log.Fatalf("[error] Lumberjack: Failed to parse timeout config option: %s", err)
		}
	}

	if segment.BatchTimeout < minimalBatchTimeout {
		log.Printf("[error] Lumberjack: timeout %s too small, using default %s", segment.BatchTimeout.String(), defaultTimeout.String())
		segment.BatchTimeout = defaultTimeout
	}
	if segment.BatchTimeout > time.Minute {
		log.Printf("[error] Lumberjack: timeout %s too large, using default %s", segment.BatchTimeout.String(), defaultTimeout)
		segment.BatchTimeout = defaultTimeout
	}
	// parse batchdebug option
	if config["batchdebug"] != "" {
		batchDebug, err := strconv.ParseBool(config["batchdebug"])
		if err != nil {
			log.Fatalf("[error] Lumberjack: Failed to parse batchdebug config option: %s", err)
		}
		// set proper BatchDebugPrintf function
		if batchDebug {
			segment.BatchDebugPrintf = DoDebugPrintf
		} else {
			segment.BatchDebugPrintf = NoDebugPrintf
		}
	}

	// parse reconnectwait option
	segment.ReconnectWait = defaultReconnectWait
	if config["reconnectwait"] != "" {
		segment.ReconnectWait, err = time.ParseDuration(config["reconnectwait"])
		if err != nil {
			log.Fatalf("[error] Lumberjack: Failed to parse reconnectwait config option: %s", err)
		}
	}

	// parse queueStatusInterval option
	segment.QueueStatusInterval = defaultQueueStatusInterval
	if config["queuestatusinterval"] != "" {
		segment.QueueStatusInterval, err = time.ParseDuration(config["queuestatusinterval"])
		if err != nil {
			log.Fatalf("[error] Lumberjack: Failed to parse queuestatussnterval config option: %s", err)
		}
	}

	// create buffered channel
	if config["queuesize"] != "" {
		buflen, err = strconv.Atoi(config["queuesize"])
		if err != nil {
			log.Fatalf("[error] Lumberjack: Failed to parse queuesize config option: %s", err)
		}
	} else {
		buflen = defaultQueueSize
	}
	if buflen < 64 {
		log.Printf("[error] Lumberjack: queuesize too small, using default %d", defaultQueueSize)
		buflen = defaultQueueSize
	}
	segment.LumberjackOut = make(chan *pb.EnrichedFlow, buflen)

	return segment
}

func (segment *Lumberjack) Run(wg *sync.WaitGroup) {
	var writerWG sync.WaitGroup

	defer func() {
		close(segment.Out)
		writerWG.Wait()
		wg.Done()
		log.Println("[info] Lumberjack: All writer functions have stopped, exiting…")
	}()

	// print queue status information
	if segment.QueueStatusInterval > 0 {
		go func() {
			length := cap(segment.LumberjackOut)
			for {
				time.Sleep(segment.QueueStatusInterval)
				fill := len(segment.LumberjackOut)
				log.Printf("[debug] Lumberjack: Queue is %3.2f%% full (%d/%d)", float64(fill)/float64(length)*100, fill, length)
			}
		}()
	}

	// run goroutine for each lumberjack server
	for server, options := range segment.Servers {
		writerWG.Add(1)
		options := options
		go func(server string) {
			defer writerWG.Done()
			// connect to lumberjack server
			client := NewResilientClient(server, options, segment.ReconnectWait)
			defer client.Close()
			log.Printf("[info] Lumberjack: Connected to %s (TLS: %v, VerifyTLS: %v, Compression: %d)", server, options.UseTLS, options.VerifyCertificate, options.CompressionLevel)

			flowInterface := make([]interface{}, segment.BatchSize)
			idx := 0

			// see https://stackoverflow.com/questions/66037676/go-reset-a-timer-newtimer-within-select-loop for timer mechanics
			timer := time.NewTimer(segment.BatchTimeout)
			timer.Stop()
			defer timer.Stop()
			var timerSet bool

			for {
				select {
				case flow, isOpen := <-segment.LumberjackOut:
					// exit on closed channel
					if !isOpen {
						// send local buffer
						count, err := client.SendNoRetry(flowInterface[:idx])
						if err != nil {
							log.Printf("[error] Lumberjack: Failed to send final flow batch upon exit to %s: %s", server, err)
						} else {
							segment.BatchDebugPrintf("[debug] Lumberjack: %s Sent final batch (%d)", server, count)
						}
						return
					}

					// append flow to batch
					flowInterface[idx] = flow
					idx++

					// send batch if full
					if idx == segment.BatchSize {
						// We got an event, and timer was already set.
						// We need to stop the timer and drain the channel if needed,
						// so that we can safely reset it later.
						if timerSet {
							if !timer.Stop() {
								<-timer.C
							}
							timerSet = false
						}

						client.Send(flowInterface)
						segment.BatchDebugPrintf("[debug] Lumberjack: %s Sent full batch (%d)", server, segment.BatchSize)

						// reset idx
						idx = 0

						// If timer was not set, or it was stopped before, it's safe to reset it.
						if !timerSet {
							timerSet = true
							timer.Reset(segment.BatchTimeout)
						}
					}
				case <-timer.C:
					// timer expired, send batch
					if idx > 0 {
						segment.BatchDebugPrintf("[debug] Lumberjack: %s Sending incomplete batch (%d/%d)", server, idx, segment.BatchSize)
						client.Send(flowInterface[:idx])
						idx = 0
					} else {
						segment.BatchDebugPrintf("[debug] Lumberjack: %s Timer expired with empty batch", server)
					}

					timer.Reset(segment.BatchTimeout)
					timerSet = true
				}
			}
		}(server)
	}

	// forward flows to lumberjack servers and to the next segment
	for msg := range segment.In {
		segment.LumberjackOut <- msg
		segment.Out <- msg
	}
	close(segment.LumberjackOut)
}

// register segment
func init() {
	segment := &Lumberjack{}
	segments.RegisterSegment("lumberjack", segment)
}
