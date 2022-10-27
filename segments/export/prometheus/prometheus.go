// Collects and serves statistics about flows.
// Configuration options for this segments are:
// The endoint field can be configured:
// <host>:8080/flowdata
// <host>:8080/metrics
// The labels to be exported can be set in the configuration.
// Default labels are:
//
//	router,ipversion,application,protoname,direction,peer,remoteas,remotecountry
//
// Additional Labels are:
//
//	src_port,dst_port,src_ip,dst_ip
//
// The additional label should be used with care, because of infite quantity.
package prometheus

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/bwNetFlow/flowpipeline/pb"
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
	if config["metricspath"] == "" {
		log.Println("[info] prometheus: Missing configuration parameter 'metricspath'. Using default path \"/metrics\"")
	} else {
		metricsPath = config["metricspath"]
	}
	var flowdataPath string = "/flowdata"
	if config["flowdatapath"] == "" {
		log.Println("[info] prometheus: Missing configuration parameter 'flowdatapath'. Using default path \"/flowdata\"")
	} else {
		flowdataPath = config["flowdatapath"]
	}

	newsegment := &Prometheus{
		Endpoint:     endpoint,
		MetricsPath:  metricsPath,
		FlowdataPath: flowdataPath,
	}

	// set default labels if not configured
	var labels []string
	if config["labels"] == "" {
		log.Println("[info] prometheus: Configuration parameter 'labels' not set. Using default labels 'Etype,Proto' to export")
		labels = strings.Split("Etype,Proto", ",")
	} else {
		labels = strings.Split(config["labels"], ",")
	}
	protofields := reflect.TypeOf(pb.EnrichedFlow{})
	for _, field := range labels {
		_, found := protofields.FieldByName(field)
		if !found {
			log.Printf("[error] Prometheus: Field '%s' specified in 'labels' does not exist.", field)
			return nil
		}
		newsegment.Labels = append(newsegment.Labels, field)
	}
	return newsegment
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
		labelset := make(map[string]string)
		values := reflect.ValueOf(msg).Elem()
		for _, fieldname := range segment.Labels {
			value := values.FieldByName(fieldname).Interface()
			switch value.(type) {
			case []uint8: // this is necessary for proper formatting
				ipstring := net.IP(value.([]uint8)).String()
				if ipstring == "<nil>" {
					ipstring = ""
				}
				labelset[fieldname] = ipstring
			case uint32: // this is because FormatUint is much faster than Sprint
				labelset[fieldname] = strconv.FormatUint(uint64(value.(uint32)), 10)
			case uint64: // this is because FormatUint is much faster than Sprint
				labelset[fieldname] = strconv.FormatUint(uint64(value.(uint64)), 10)
			case string: // this is because doing nothing is also much faster than Sprint
				labelset[fieldname] = value.(string)
			default:
				labelset[fieldname] = fmt.Sprint(value)
			}
		}
		promExporter.Increment(msg.Bytes, msg.Packets, labelset)
		segment.Out <- msg
	}
}

func init() {
	segment := &Prometheus{}
	segments.RegisterSegment("prometheus", segment)
}
