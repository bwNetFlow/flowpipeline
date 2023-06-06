// Prints all flows to stdout or a given file in json format, for consumption by the stdin segment or for debugging.
package json

import (
	"bufio"
	"fmt"
	"github.com/klauspost/compress/zstd"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
	"google.golang.org/protobuf/encoding/protojson"
)

type Json struct {
	segments.BaseSegment
	writer *bufio.Writer

	FileName string // optional, default is empty which means stdout
}

func (segment Json) New(config map[string]string) segments.Segment {
	newsegment := &Json{}

	var filename string = "stdout"
	var file *os.File
	var err error
	if config["filename"] != "" {
		file, err = os.Create(config["filename"])
		if err != nil {
			log.Printf("[error] Json: File specified in 'filename' is not accessible: %s", err)
		}
		filename = config["filename"]
	} else {
		file = os.Stdout
		log.Println("[info] Json: 'filename' unset, using stdout.")
	}
	// configure zstd compression
	if config["zstd"] != "" {
		rawLevel, err := strconv.Atoi(config["zstd"])
		var level zstd.EncoderLevel
		if err != nil {
			log.Printf("[warning] Json: Unable to parse zstd option, using default: %s", err)
			level = zstd.SpeedDefault
		} else {
			level = zstd.EncoderLevelFromZstd(rawLevel)
		}
		encoder, err := zstd.NewWriter(file, zstd.WithEncoderLevel(level))
		if err != nil {
			log.Fatalf("[error] Json: error creating zstd encoder: %s", err)
		}
		newsegment.writer = bufio.NewWriter(encoder)
	} else {
		// no compression
		newsegment.writer = bufio.NewWriter(file)
	}
	newsegment.FileName = filename

	return newsegment
}

func (segment *Json) Run(wg *sync.WaitGroup) {
	defer func() {
		_ = segment.writer.Flush()
		close(segment.Out)
		wg.Done()
	}()
	for msg := range segment.In {
		data, err := protojson.Marshal(msg)
		if err != nil {
			log.Printf("[warning] Json: Skipping a flow, failed to recode protobuf as JSON: %v", err)
			continue
		}

		// use Fprintln because it adds an OS specific newline
		_, err = fmt.Fprintln(segment.writer, string(data))
		if err != nil {
			log.Printf("[warning] Json: Skipping a flow, failed to write to file %s: %v", segment.FileName, err)
			continue
		}
		// we need to flush here every time because we need full lines and can not wait
		// in case of using this output as in input for other instances consuming flow data
		_ = segment.writer.Flush()
		segment.Out <- msg
	}
}

func init() {
	segment := &Json{}
	segments.RegisterSegment("json", segment)
}
