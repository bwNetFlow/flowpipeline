// Receives flows from stdin in JSON format, as exported by the json segment.
// This segment can also read from a file with flows in json format per each line
package stdin

import (
	"bufio"
	"log"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
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
		case msg, ok := <-segment.In:
			if !ok {
				return
			}
			segment.Out <- msg
		case line := <-fromStdin:
			msg := &flow.FlowMessage{}
			err := protojson.Unmarshal(line, msg)
			if err != nil {
				log.Printf("[warning] StdIn: Skipping a flow, failed to recode input to protobuf: %v", err)
				continue
			}
			segment.Out <- msg
		}
	}
}

func init() {
	segment := &StdIn{}
	segments.RegisterSegment("stdin", segment)
}
