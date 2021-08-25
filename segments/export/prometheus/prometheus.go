// Collects and serves statistics about flows. Reuses the exporter package from
// https://github.com/bwNetFlow/consumer_prometheus
package prometheus

import (
	"log"
	"sync"

	"github.com/bwNetFlow/consumer_prometheus/exporter"
	"github.com/bwNetFlow/flowpipeline/segments"
)

type Prometheus struct {
	segments.BaseSegment
	Endpoint string // required, no default
}

func (segment Prometheus) New(config map[string]string) segments.Segment {
	// TODO: provide default port
	if config["endpoint"] == "" {
		log.Println("[error] prometheus: Missing required configuration parameter 'endpoint'.")
		return nil
	}

	return &Prometheus{
		Endpoint: config["endpoint"],
	}
}

func (segment *Prometheus) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	var promExporter = exporter.Exporter{}
	promExporter.Initialize()
	promExporter.ServeEndpoints(segment.Endpoint)

	for msg := range segment.In {
		segment.Out <- msg
		promExporter.Increment(msg)
	}
}

func init() {
	segment := &Prometheus{}
	segments.RegisterSegment("prometheus", segment)
}
