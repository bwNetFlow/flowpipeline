// Collects and serves statistics about flows.
// Configuration options for this segments are:
// The endoint field can be configured:
// <host>:8080/flowdata
// <host>:8080/metrics
// The labels to be exported can be set in the configuration.
// Default labels are:
//   router,ipversion,application,protoname,direction,peer,remoteas,remotecountry
// Additional Labels are:
//  src_port,dst_port,src_ip,dst_ip
// The additional label should be used with care, because of infite quantity.
package prometheus

import (
	"log"
	"strings"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
)

type Prometheus struct {
	segments.BaseSegment
	Endpoint     string   // optional, default value is ":8080"
	MetricsPath  string   // optional, default is "/metrics"
	FlowdataPath string   // optional, default is "/flowdata"
	Labels       []string // optional, list of labels to be exported
}

func (segment Prometheus) New(config map[string]string) segments.Segment {
	var endpoint string = ":8080"
	if config["endpoint"] == "" {
		log.Println("[info] prometheus: Missing configuration parameter 'endpoint'. Using default port \":8080\"")
	} else {
		endpoint = config["endpoint"]
	}
	var metricsPath string = "/metrics"
	if config["metricsPath"] == "" {
		log.Println("[info] prometheus: Missing configuration parameter 'metricsPath'. Using default port \"/metrics\"")
	} else {
		metricsPath = config["metricsPath"]
	}
	var flowdataPath string = "/flowdata"
	if config["flowdataPath"] == "" {
		log.Println("[info] prometheus: Missing configuration parameter 'flowdataPath'. Using default port \"/metrics\"")
	} else {
		flowdataPath = config["flowdataPath"]
	}

	// set default labels if not configured
	var labels = []string{
		"router",
		"ipversion",
		"application",
		"protoname",
		"direction",
		"peer",
		"remoteas",
		"remotecountry",
	}
	if config["labels"] == "" {
		log.Println("[info] prometheus: Configuration parameter 'labels' not set. Using default labels to export")
	} else {
		// TODO: maybe validate for supported labels, anyway they will be empty
		labels = []string{}
		for _, lab := range strings.Split(config["labels"], ",") {
			log.Printf("[info] prometheus: custom label found: %s", lab)
			labels = append(labels, lab)
		}
	}

	return &Prometheus{
		Endpoint:     endpoint,
		MetricsPath:  metricsPath,
		FlowdataPath: flowdataPath,
		Labels:       labels,
	}
}

func (segment *Prometheus) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	var promExporter = Exporter{}
	promExporter.Initialize(segment.Labels)
	promExporter.ServeEndpoints(segment)

	for msg := range segment.In {
		segment.Out <- msg
		promExporter.Increment(msg)
	}
}

func init() {
	segment := &Prometheus{}
	segments.RegisterSegment("prometheus", segment)
}
