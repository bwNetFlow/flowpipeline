// Filters out the bulky average of flows.
package elephant

import (
	"log"
	"math"
	"sync"
	"time"

	"github.com/asecurityteam/rolling"
	"github.com/bwNetFlow/flowpipeline/segments"
)

type Elephant struct {
	segments.BaseSegment
}

func (segment Elephant) New(config map[string]string) segments.Segment {
	return &Elephant{}
}

func (segment *Elephant) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	var window = rolling.NewTimePolicy(rolling.NewWindow(300), time.Second) // 300s == 5min
	for msg := range segment.In {
		avg := window.Reduce(rolling.Avg)
		stddev := window.Reduce(StdDev(avg))
		threshold := avg + 3*stddev
		if float64(msg.Bytes) >= threshold {
			log.Printf("[info] Elephant: FORWARD, %d >= %f (%f + 1*%f)", msg.Bytes, threshold, avg, stddev)
			segment.Out <- msg
		}
		window.Append(float64(msg.Bytes))
	}
}

func StdDev(avg float64) func(w rolling.Window) float64 {
	return func(w rolling.Window) float64 {
		var result = 0.0
		var count = 0.0
		for _, bucket := range w {
			for _, value := range bucket {
				result = result + math.Pow(value-avg, 2.0)
				count = count + 1
			}
		}
		return math.Sqrt(result / count)
	}
}

func init() {
	segment := &Elephant{}
	segments.RegisterSegment("elephant", segment)
}
