// Prints all flows to stdout, for consumption by the stdin segment or for debugging.
package stdout

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
	"google.golang.org/protobuf/encoding/protojson"
)

type StdOut struct {
	segments.BaseSegment
}

func (segment StdOut) New(config map[string]string) segments.Segment {
	return &StdOut{}
}

func (segment *StdOut) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	for msg := range segment.In {
		data, err := protojson.Marshal(msg)
		if err != nil {
			log.Printf("[warning] StdOut: Skipping a flow, failed to recode protobuf as JSON: %v", err)
			continue
		}
		fmt.Fprintln(os.Stdout, string(data))
		segment.Out <- msg
	}
}

func init() {
	segment := &StdOut{}
	segments.RegisterSegment("stdout", segment)
}
