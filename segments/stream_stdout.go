package segments

import (
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"os"
	"sync"
)

type StdOut struct {
	BaseSegment
}

func (segment StdOut) New(config map[string]string) Segment {
	return &StdOut{}
}

func (segment *StdOut) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.out)
		wg.Done()
	}()
	for msg := range segment.in {
		data, err := protojson.Marshal(msg)
		if err != nil {
			log.Printf("[warning] StdOut: Skipping a flow, failed to recode protobuf as JSON: %v", err)
			continue
		}
		fmt.Fprintln(os.Stdout, string(data))
		segment.out <- msg
	}
}
