package segments

import (
	"fmt"
	"strconv"
	"sync"
)

type PrintDots struct {
	BaseSegment
	FlowsPerDot uint64
}

func (segment PrintDots) New(config map[string]string) Segment {
	var fpd uint64
	if parsedFpd, err := strconv.ParseUint(config["flows_per_dot"], 10, 32); err == nil {
		fpd = parsedFpd
	} else {
		fpd = 5000
	}
	return &PrintDots{
		FlowsPerDot: fpd,
	}
}

func (segment *PrintDots) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.out)
		wg.Done()
	}()
	count := uint64(0)
	for msg := range segment.in {
		if count += 1; count >= segment.FlowsPerDot {
			fmt.Printf(".")
			count = 0
		}
		segment.out <- msg
	}
}

func init() {
	segment := &PrintDots{}
	RegisterSegment("printdots", segment)
}
