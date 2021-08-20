package segments

import (
	"strconv"
	"sync"
)

type Normalize struct {
	BaseSegment
	Fallback uint64
}

func (segment Normalize) New(config map[string]string) Segment {
	var fallback uint64
	if parsedFallback, err := strconv.ParseUint(config["fallback"], 10, 32); err == nil {
		fallback = parsedFallback
	} else {
		fallback = 0
	}
	return &Normalize{
		Fallback: fallback,
	}
}

func (segment *Normalize) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.out)
		wg.Done()
	}()
	for msg := range segment.in {
		if msg.SamplingRate > 0 {
			msg.Bytes *= msg.SamplingRate
			msg.Packets *= msg.SamplingRate
			msg.Normalized = 1
		} else if segment.Fallback != 0 {
			msg.Bytes *= segment.Fallback
			msg.Packets *= segment.Fallback
			msg.SamplingRate = segment.Fallback
			msg.Normalized = 1
		}
		segment.out <- msg
	}
}

func init() {
	segment := &Normalize{}
	RegisterSegment("normalize", segment)
}
