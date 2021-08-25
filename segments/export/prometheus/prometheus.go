// Collects and serves statistics about passing flows.
package prometheus

import (
	"log"
	"sync"

	"github.com/bwNetFlow/consumer_prometheus/exporter"
	"github.com/bwNetFlow/flowpipeline/segments"
)

type PrometheusExporter struct {
	segments.BaseSegment
	PromEndpoint string
}

func (segment PrometheusExporter) New(config map[string]string) segments.Segment {
	// TODO: provide default port
	if config["endpoint"] == "" {
		log.Println("[error] PrometheusExporter: Missing required configuration parameter 'endpoint'.")
		return nil
	}

	return &PrometheusExporter{
		PromEndpoint: config["endpoint"],
	}
}

func (segment *PrometheusExporter) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	var promExporter = exporter.Exporter{}
	promExporter.Initialize()
	promExporter.ServeEndpoints(segment.PromEndpoint)

	for msg := range segment.In {
		segment.Out <- msg
		promExporter.Increment(msg)
	}
}

func init() {
	segment := &PrometheusExporter{}
	segments.RegisterSegment("prometheusexporter", segment)
}
