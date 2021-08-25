// Prints a dot every n flows.
package printdots

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
)

type PrintDots struct {
	segments.BaseSegment
	FlowsPerDot uint64 // optional, default is 5000
}

func (segment PrintDots) New(config map[string]string) segments.Segment {
	var fpd uint64 = 5000
	if parsedFpd, err := strconv.ParseUint(config["flows_per_dot"], 10, 32); err == nil {
		fpd = parsedFpd
	} else {
		if config["flows_per_dot"] != "" {
			log.Println("[error] PrintDots: Could not parse 'flows_per_dot' parameter, using default 5000.")
		} else {
			log.Println("[info] PrintDots: 'flows_per_dot' set to default '5000'.")
		}
	}
	return &PrintDots{
		FlowsPerDot: fpd,
	}
}

func (segment *PrintDots) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	count := uint64(0)
	for msg := range segment.In {
		if count += 1; count >= segment.FlowsPerDot {
			fmt.Printf(".")
			count = 0
		}
		segment.Out <- msg
	}
}

func init() {
	segment := &PrintDots{}
	segments.RegisterSegment("printdots", segment)
}
