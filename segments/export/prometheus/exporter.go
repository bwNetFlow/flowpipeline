package prometheus

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Exporter provides export features to Prometheus
type Exporter struct {
	FlowReg *prometheus.Registry

	kafkaMessageCount prometheus.Counter
	kafkaOffsets      *prometheus.CounterVec
	flowBits          *prometheus.CounterVec

	labels []string
}

// Initialize Prometheus Exporter
func (e *Exporter) Initialize(labels []string) {
	e.labels = labels

	// The Kafka metrics are added to the global registry.
	e.kafkaMessageCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_messages_total",
			Help: "Number of Kafka messages",
		})
	e.kafkaOffsets = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_offset_current",
			Help: "Current Kafka Offset of the consumer",
		}, []string{"topic", "partition"})
	prometheus.MustRegister(e.kafkaMessageCount, e.kafkaOffsets)

	// Flows are stored in a separate Registry
	e.flowBits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "flow_bits",
			Help: "Number of Bits received across Flows.",
		}, labels)

	e.FlowReg = prometheus.NewRegistry()
	e.FlowReg.MustRegister(e.flowBits)

}

// listen on given endpoint addr with Handler for metricPath and flowdataPath
func (e *Exporter) ServeEndpoints(segment *Prometheus) {
	http.Handle(segment.MetricsPath, promhttp.Handler())
	http.Handle(segment.FlowdataPath, promhttp.HandlerFor(e.FlowReg, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Flow Exporter</title></head>
			<body>
			<h1>Flow Exporter</h1>
			<p><a href="` + segment.MetricsPath + `">Metrics</p>
			<p><a href="` + segment.FlowdataPath + `">Flow Data</p>
			</body>
		</html>`))
	})

	go func() {
		http.ListenAndServe(segment.Endpoint, nil)
	}()
	log.Printf("Enabled Prometheus %s and %s endpoints.", segment.MetricsPath, segment.FlowdataPath)
}

func (e *Exporter) Increment(bytes uint64, packets uint64, labelset prometheus.Labels) {
	e.kafkaMessageCount.Inc()
	// e.flowNumber.With(labels).Inc()
	// flowPackets.With(labels).Add(float64(flow.GetPackets()))
	e.flowBits.With(labelset).Add(float64(bytes) * 8)
}

func (e *Exporter) IncrementCtrl(topic string, partition int32, offset int64) {
	labels := prometheus.Labels{
		"topic":     topic,
		"partition": fmt.Sprint(partition),
	}
	e.kafkaOffsets.With(labels).Add(float64(offset))
}
