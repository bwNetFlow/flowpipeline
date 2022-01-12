// Filters out the bulky average of flows.
package elephant

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/asecurityteam/rolling"
	"github.com/bwNetFlow/flowpipeline/segments"
)

type Elephant struct {
	segments.BaseSegment
	Aspect     string  // optional, one of "bytes", "bps", "packets", or "pps", default is "bytes", determines which aspect qualifies a flow as an elephant
	Percentile float64 // optional, default is 99.00, determines the cutoff percentile for flows being dropped by this segment, i.e. 95.00 corresponds to outputting the top 5% only
	Exact      bool    // optional, default is false, determines whether to use percentiles that are exact or generated using the P-square estimation algorithm
}

func (segment Elephant) New(config map[string]string) segments.Segment {
	var aspect = "bytes"
	if config["aspect"] != "" {
		if strings.ToLower(config["aspect"]) == "bytes" || strings.ToLower(config["aspect"]) == "packets" || strings.ToLower(config["aspect"]) == "bps" || strings.ToLower(config["aspect"]) == "pps" {
			aspect = strings.ToLower(config["aspect"])
		} else {
			log.Println("[error] Elephant: Could not parse 'aspect' parameter, using default 'bytes'.")
		}
	} else {
		log.Println("[info] Elephant: 'aspect' set to default 'bytes'.")
	}

	var percentile = 99.00
	if config["percentile"] != "" {
		if parsedPercentile, err := strconv.ParseFloat(config["percentile"], 64); err == nil {
			percentile = parsedPercentile
			if percentile == 0 {
				log.Println("[error] Elephant: Using 0-Percentile corresponds to no-op. Remove this segment or use a higher value.")
				return nil
			}
		} else {
			log.Println("[error] Elephant: Could not parse 'percentile' parameter, using default 99.00.")
		}
	} else {
		log.Println("[info] Elephant: 'percentile' set to default 99.00.")
	}

	var exact = false
	if config["exact"] != "" {
		if parsedExact, err := strconv.ParseBool(config["exact"]); err == nil {
			exact = parsedExact
		} else {
			log.Println("[error] Elephant: Could not parse 'exact' parameter, using default false.")
		}
	} else {
		log.Println("[info] Elephant: 'exact' set to default false.")
	}

	return &Elephant{
		Aspect:     aspect,
		Percentile: percentile,
		Exact:      exact,
	}
}

func (segment *Elephant) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	var window = rolling.NewTimePolicy(rolling.NewWindow(60), time.Second) // 300s == 5min TODO
	for msg := range segment.In {
		var threshold float64
		if segment.Exact {
			threshold = window.Reduce(rolling.Percentile(segment.Percentile))
		} else {
			threshold = window.Reduce(rolling.FastPercentile(segment.Percentile))
		}
		var aspect float64
		switch segment.Aspect {
		case "bytes":
			aspect = float64(msg.Bytes)
		case "bps":
			duration := msg.TimeFlowEnd - msg.TimeFlowStart
			if duration == 0 {
				duration += 1
			}
			aspect = float64(msg.Bytes * 8.0 / duration)
		case "pps":
			duration := msg.TimeFlowEnd - msg.TimeFlowStart
			if duration == 0 {
				duration += 1
			}
			aspect = float64(msg.Packets / duration)
		case "packets":
			aspect = float64(msg.Packets)
		}
		if aspect >= threshold {
			log.Printf("[debug] Elephant: Found elephant with size %d (>=%f)", msg.Bytes, threshold)
			segment.Out <- msg
		}
		window.Append(aspect)
	}
}

func init() {
	segment := &Elephant{}
	segments.RegisterSegment("elephant", segment)
}
