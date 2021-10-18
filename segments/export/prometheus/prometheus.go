// Collects and serves statistics about flows. Reuses the exporter package from
// https://github.com/bwNetFlow/consumer_prometheus
package prometheus

import (
	"log"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
)

type Prometheus struct {
	segments.BaseSegment
	Endpoint string // optional, default value is ":8080"
}

func (segment Prometheus) New(config map[string]string) segments.Segment {
	var endpoint string = ":8080"
	if config["endpoint"] == "" {
		log.Println("[info] prometheus: Missing configuration parameter 'endpoint'. Using default port \":8080\"")
	} else {
		endpoint = config["endpoint"]
	}

	return &Prometheus{
		Endpoint: endpoint,
	}
}

func (segment *Prometheus) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	var promExporter = Exporter{}
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
