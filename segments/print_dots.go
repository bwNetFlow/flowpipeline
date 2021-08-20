package segments

import (
	"fmt"
	"log"
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
		if config["flows_per_dot"] != "" {
			log.Println("[error] PrintDots: Could not parse 'flows_per_dot' parameter, using default 5000.")
		} else {
			log.Println("[info] PrintDots: 'flows_per_dot' set to default '5000'.")
		}
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
