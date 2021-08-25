package stdin

import (
	"bufio"
	"bytes"
	"log"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
	"google.golang.org/protobuf/encoding/protojson"

	"io"
	"os"
	"sync"
)

type StdIn struct {
	segments.BaseSegment
}

func (segment StdIn) New(config map[string]string) segments.Segment {
	return &StdIn{}
}

func (segment *StdIn) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	fromStdin := make(chan []byte)
	go func() {
		for {
			rdr := bufio.NewReader(os.Stdin)
			line, err := rdr.ReadBytes('\n')
			if err != io.EOF {
				log.Printf("[warning] StdIn: Skipping a flow, could not read line from stdin: %v", err)
				continue
			}
			if len(line) == 0 {
				continue
			}
			fromStdin <- bytes.TrimSuffix(line, []byte("\n"))
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
				log.Printf("[warning] StdIn: Skipping a flow, failed to recode stdin to protobuf: %v", err)
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
