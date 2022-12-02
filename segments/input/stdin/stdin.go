// Receives flows from stdin in JSON format, as exported by the json segment.
// This segment can also read from a file with flows in json format per each line
package stdin

import (
	"bufio"
	"log"
	"time"

	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"

	"os"
	"sync"
)

type StdIn struct {
	segments.BaseSegment
	scanner *bufio.Scanner

	FileName string // optional, default is empty which means read from stdin
}

func (segment StdIn) New(config map[string]string) segments.Segment {
	newsegment := &StdIn{}

	var filename string = "stdout"
	var file *os.File
	var err error
	if config["filename"] != "" {
		file, err = os.Open(config["filename"])
		if err != nil {
			log.Printf("[error] StdIn: File specified in 'filename' is not accessible: %s", err)
			return nil
		}
		filename = config["filename"]
	} else {
		file = os.Stdin
		log.Println("[info] StdIn: 'filename' unset, using stdIn.")
	}
	newsegment.scanner = bufio.NewScanner(file)

	newsegment.FileName = filename

	return newsegment
}

func (segment *StdIn) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	fromStdin := make(chan []byte)
	go func() {
		for {
			segment.scanner.Scan()
			if err := segment.scanner.Err(); err != nil {
				log.Printf("[warning] StdIn: Skipping a flow, could not read line from stdin: %v", err)
				continue
			}
			if len(segment.scanner.Text()) == 0 {
				continue
			}
			// we need to get full representation of text and cast it to []byte
			// because scanner.Bytes doesn't return all content.
			fromStdin <- []byte(segment.scanner.Text())
		}
	}()
	for {
		select {
		case fc, ok := <-segment.In:
			if !ok {
				return
			}
			_, span := segment.Trace(fc)
			if span != nil {
				span.End()
			}
			segment.Out <- fc
			span.End()
		case line := <-fromStdin:
			ts := time.Now()
			msg := &pb.EnrichedFlow{}
			fc := pb.NewFlowContainer(msg, ts)
			parentCtx, span := segment.Trace(fc)

			// parse stdin
			var parseSpan trace.Span
			if parentCtx != nil {
				_, parseSpan = otel.Tracer("flowpipeline").Start(parentCtx, "parse")
			}
			err := protojson.Unmarshal(line, msg)
			if err != nil {
				log.Printf("[warning] StdIn: Skipping a flow, failed to recode input to protobuf: %v", err)
				span.AddEvent("parsing failed: " + err.Error())
				parseSpan.End()
				continue
			}
			if parseSpan != nil {
				parseSpan.End()
				span.AddEvent("generate")
				span.End()
			}
			segment.Out <- fc
		}
	}
}

func init() {
	segment := &StdIn{}
	segments.RegisterSegment("stdin", segment)
}
