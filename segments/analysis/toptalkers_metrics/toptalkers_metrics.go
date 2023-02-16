package flowpipeline

import (
	"bufio"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bwNetFlow/flowpipeline/segments"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Record struct {
	FwdBytes    []uint64
	FwdPackets  []uint64
	DropBytes   []uint64
	DropPackets []uint64
	capacity    int
	pointer     int
	stopOnce    sync.Once
	stopC       chan struct{}
	sync.RWMutex
}

type ToptalkersMetrics struct {
	segments.BaseSegment
	writer   *bufio.Writer
	database *Database

	TrafficType     string // optional, default is "", name for the traffic type (included as label)
	Window          int    // optional, default is 60, sets the number of seconds used as a sliding window size
	ThresholdWindow int    // optional, use the last N buckets for calculation of averages, default: $window
	ReportWindow    int    // optional, use the last N buckets to calculate averages that are reported as result, default: $window
	ThresholdBps    uint64 // optional, default is 0, only log talkers with an average bits per second rate higher than this value
	ThresholdPps    uint64 // optional, default is 0, only log talkers with an average packets per second rate higher than this value
	Endpoint        string // optional, default value is ":8080"
	MetricsPath     string // optional, default is "/metrics"
	FlowdataPath    string // optional, default is "/flowdata"
	RelevantAddress string // optional, default is "destination", options are "destination", "source", "both"
}

type Database struct {
	database   map[string]*Record
	windowSize int
	sync.RWMutex
}

func (db *Database) GetRecord(key string) *Record {
	db.Lock()
	defer db.Unlock()
	record, found := db.database[key]
	if !found || record == nil {
		record = NewRecord(db.windowSize)
		db.database[key] = record
	}
	return record
}

func NewRecord(windowSize int) *Record {
	record := &Record{
		FwdBytes:    make([]uint64, windowSize),
		FwdPackets:  make([]uint64, windowSize),
		DropBytes:   make([]uint64, windowSize),
		DropPackets: make([]uint64, windowSize),
		capacity:    windowSize,
		pointer:     0,
		stopC:       make(chan struct{}),
	}
	go record.clock()
	return record
}

func (record *Record) Append(bytes uint64, packets uint64, statusFwd bool) {
	record.Lock()
	defer record.Unlock()
	if statusFwd == true {
		record.FwdBytes[record.pointer] += bytes
		record.FwdPackets[record.pointer] += packets
	} else {
		record.DropBytes[record.pointer] += bytes
		record.DropPackets[record.pointer] += packets
	}
}

func (record *Record) GetBytes(lastBuckets int) uint64 {
	// lastBuckets == 0 means "look at the whole window"
	if lastBuckets == 0 {
		lastBuckets = record.capacity
	}
	var sum uint64
	record.RLock()
	defer record.RUnlock()
	pos := record.pointer
	for i := 0; i < lastBuckets; i++ {
		if pos <= 0 {
			pos = record.capacity - 1
		} else {
			pos--
		}
		sum = sum + record.FwdBytes[pos] + record.DropBytes[pos]
	}
	return sum
}

func (record *Record) GetPackets(lastBuckets int) uint64 {
	// lastBuckets == 0 means "look at the whole window"
	if lastBuckets == 0 {
		lastBuckets = record.capacity
	}
	var sum uint64
	record.RLock()
	defer record.RUnlock()
	pos := record.pointer
	for i := 0; i < lastBuckets; i++ {
		if pos <= 0 {
			pos = record.capacity - 1
		} else {
			pos--
		}
		sum = sum + record.FwdPackets[pos] + record.DropPackets[pos]
	}
	return sum
}

func (record *Record) GetMetrics(lastBuckets int) (uint64, uint64, uint64, uint64) {
	// lastBuckets == 0 means "look at the whole window"
	if lastBuckets == 0 {
		lastBuckets = record.capacity
	}
	sumFwdBytes := uint64(0)
	sumFwdPackets := uint64(0)
	sumDropBytes := uint64(0)
	sumDropPackets := uint64(0)
	record.RLock()
	defer record.RUnlock()
	pos := record.pointer
	for i := 0; i < lastBuckets; i++ {
		if pos <= 0 {
			pos = record.capacity - 1
		} else {
			pos--
		}
		sumFwdBytes += record.FwdBytes[pos]
		sumFwdPackets += record.FwdPackets[pos]
		sumDropBytes += record.DropBytes[pos]
		sumDropPackets += record.DropPackets[pos]
	}
	return sumFwdBytes, sumFwdPackets, sumDropBytes, sumDropPackets
}

func (record *Record) clock() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			record.Lock()
			record.pointer++
			if record.pointer >= record.capacity {
				record.pointer = 0
			}
			record.FwdBytes[record.pointer] = 0
			record.FwdPackets[record.pointer] = 0
			record.DropBytes[record.pointer] = 0
			record.DropPackets[record.pointer] = 0
			record.Unlock()
		case <-record.stopC:
			return
		}
	}
}

func (db *Database) GetAllRecords() <-chan struct {
	key    string
	record *Record
} {
	out := make(chan struct {
		key    string
		record *Record
	})
	go func() {
		db.Lock()
		defer func() {
			db.Unlock()
			close(out)
		}()
		for key, record := range db.database {
			// if there were no packets seen for one window length, remove the key/address
			if record.GetPackets(0) == 0 {
				record.stopOnce.Do(func() {
					record.stopC <- struct{}{}
				})
				delete(db.database, key)
				continue
			}
			out <- struct {
				key    string
				record *Record
			}{key, record}
		}
	}()
	return out
}

// Exporter provides export features to Prometheus
type PrometheusExporter struct {
	MetaReg *prometheus.Registry
	FlowReg *prometheus.Registry

	kafkaMessageCount prometheus.Counter
}

// Initialize Prometheus Exporter
func (e *PrometheusExporter) Initialize(collector *PrometheusCollector) {
	// The Kafka metrics are added to the global registry.
	e.kafkaMessageCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_messages_total",
			Help: "Number of Kafka messages",
		})
	e.MetaReg = prometheus.NewRegistry()
	e.MetaReg.MustRegister(e.kafkaMessageCount)

	e.FlowReg = prometheus.NewRegistry()
	e.FlowReg.MustRegister(collector)
}

// listen on given endpoint addr with Handler for metricPath and flowdataPath
func (e *PrometheusExporter) ServeEndpoints(segment *ToptalkersMetrics) {
	mux := http.NewServeMux()
	mux.Handle(segment.MetricsPath, promhttp.HandlerFor(e.MetaReg, promhttp.HandlerOpts{}))
	mux.Handle(segment.FlowdataPath, promhttp.HandlerFor(e.FlowReg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
		http.ListenAndServe(segment.Endpoint, mux)
	}()
	log.Printf("Enabled metrics on %s and %s, listening at %s.", segment.MetricsPath, segment.FlowdataPath, segment.Endpoint)
}

func (segment ToptalkersMetrics) New(config map[string]string) segments.Segment {
	newsegment := &ToptalkersMetrics{
		Window:          60,
		ThresholdWindow: 60,
		ReportWindow:    60,
		TrafficType:     "",
		Endpoint:        ":8080",
		MetricsPath:     "/metrics",
		FlowdataPath:    "/flowdata",
		RelevantAddress: "destination",
	}

	if config["window"] != "" {
		if parsedWindow, err := strconv.ParseInt(config["window"], 10, 64); err == nil {
			newsegment.Window = int(parsedWindow)
			if newsegment.Window <= 0 {
				log.Println("[error] : Window has to be >0.")
				return nil
			}
		} else {
			log.Println("[error] ToptalkersMetrics: Could not parse 'window' parameter, using default 60.")
		}
	} else {
		log.Println("[info] ToptalkersMetrics: 'window' set to default 60.")
	}

	if config["thresholdwindow"] != "" {
		if parsedThresholdWindow, err := strconv.ParseInt(config["thresholdwindow"], 10, 64); err == nil {
			newsegment.ThresholdWindow = int(parsedThresholdWindow)
			if newsegment.ThresholdWindow <= 0 {
				log.Println("[error] : thresholdwindow has to be >0.")
				return nil
			}
		} else {
			log.Println("[error] ToptalkersMetrics: Could not parse 'thresholdwindow' parameter, using default (1 window).")
		}
	} else {
		log.Println("[info] ToptalkersMetrics: 'thresholdwindow' set to default (1 window).")
	}

	if config["reportwindow"] != "" {
		if parsedReportWindow, err := strconv.ParseInt(config["reportwindow"], 10, 64); err == nil {
			newsegment.ReportWindow = int(parsedReportWindow)
			if newsegment.ReportWindow <= 0 {
				log.Println("[error] : reportwindow has to be >0.")
				return nil
			}
		} else {
			log.Println("[error] ReportPrometheus: Could not parse 'reportwindow' parameter, using default (1 window).")
		}
	} else {
		log.Println("[info] ReportPrometheus: 'reportwindow' set to default (1 window).")
	}

	if config["traffictype"] != "" {
		newsegment.TrafficType = config["traffictype"]
	} else {
		log.Println("[info] ToptalkersMetrics: 'traffictype' is empty.")
	}

	if config["thresholdbps"] != "" {
		if parsedThresholdBps, err := strconv.ParseUint(config["thresholdbps"], 10, 32); err == nil {
			newsegment.ThresholdBps = parsedThresholdBps
		} else {
			log.Println("[error] ToptalkersMetrics: Could not parse 'thresholdbps' parameter, using default 0.")
		}
	} else {
		log.Println("[info] ToptalkersMetrics: 'thresholdbps' set to default '0'.")
	}

	if config["thresholdpps"] != "" {
		if parsedThresholdPps, err := strconv.ParseUint(config["thresholdpps"], 10, 32); err == nil {
			newsegment.ThresholdPps = parsedThresholdPps
		} else {
			log.Println("[error] ToptalkersMetrics: Could not parse 'thresholdpps' parameter, using default 0.")
		}
	} else {
		log.Println("[info] ToptalkersMetrics: 'thresholdpps' set to default '0'.")
	}

	if config["endpoint"] == "" {
		log.Println("[info] ToptalkersMetrics Missing configuration parameter 'endpoint'. Using default port \":8080\"")
	} else {
		newsegment.Endpoint = config["endpoint"]
	}

	if config["metricspath"] == "" {
		log.Println("[info] ToptalkersMetrics: Missing configuration parameter 'metricspath'. Using default path \"/metrics\"")
	} else {
		newsegment.FlowdataPath = config["metricspath"]
	}
	if config["flowdatapath"] == "" {
		log.Println("[info] ThresholdToptalkersMetrics: Missing configuration parameter 'flowdatapath'. Using default path \"/flowdata\"")
	} else {
		newsegment.FlowdataPath = config["flowdatapath"]
	}

	switch config["relevantaddress"] {
	case
		"destination",
		"source",
		"both":
		newsegment.RelevantAddress = config["relevantaddress"]
	case "":
		log.Println("[info] ToptalkersMetrics: 'relevantaddress' set to default 'destination'.")
	default:
		log.Println("[error] ToptalkersMetrics: Could not parse 'relevantaddress', using default value 'destination'.")
	}
	return newsegment
}

func (segment *ToptalkersMetrics) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	segment.database = &Database{
		database:   map[string]*Record{},
		windowSize: segment.Window,
	}

	var promExporter = PrometheusExporter{}
	collector := &PrometheusCollector{segment}
	promExporter.Initialize(collector)
	promExporter.ServeEndpoints(segment)

	for msg := range segment.In {
		promExporter.kafkaMessageCount.Inc()
		var keys []string
		if segment.RelevantAddress == "source" {
			keys = []string{msg.SrcAddrObj().String()}
		} else if segment.RelevantAddress == "destination" {
			keys = []string{msg.DstAddrObj().String()}
		} else if segment.RelevantAddress == "both" {
			keys = []string{msg.SrcAddrObj().String(), msg.DstAddrObj().String()}
		}
		for _, key := range keys {
			record := segment.database.GetRecord(key)
			record.Append(msg.Bytes, msg.Packets, msg.IsForwarded())
		}
		segment.Out <- msg
	}
}

var (
	trafficBpsDesc = prometheus.NewDesc(
		"traffic_bps",
		"Traffic volume in bits per second for a given address.",
		[]string{"traffic_type", "address", "forwarding_status"}, nil,
	)
	trafficPpsDesc = prometheus.NewDesc(
		"traffic_pps",
		"Traffic in packets per second for a given address.",
		[]string{"traffic_type", "address", "forwarding_status"}, nil,
	)
)

type PrometheusCollector struct {
	segment *ToptalkersMetrics
}

func (collector *PrometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- trafficBpsDesc
	ch <- trafficPpsDesc
}
func (collector *PrometheusCollector) Collect(ch chan<- prometheus.Metric) {
	for entry := range collector.segment.database.GetAllRecords() {
		key := entry.key
		record := entry.record
		// check if thresholds are exceeded
		rw := collector.segment.ReportWindow
		sumBps := float64(record.GetBytes(rw)*8) / float64(rw)
		sumPps := float64(record.GetPackets(rw)) / float64(rw)
		if (sumBps > float64(collector.segment.ThresholdBps)) && (sumPps > float64(collector.segment.ThresholdPps)) {
			sumFwdBytes, sumFwdPackets, sumDropBytes, sumDropPackets := record.GetMetrics(rw)
			ch <- prometheus.MustNewConstMetric(
				trafficBpsDesc,
				prometheus.GaugeValue,
				float64(sumFwdBytes*8)/float64(rw),
				collector.segment.TrafficType, key, "forwarded",
			)
			ch <- prometheus.MustNewConstMetric(
				trafficBpsDesc,
				prometheus.GaugeValue,
				float64(sumDropBytes*8)/float64(rw),
				collector.segment.TrafficType, key, "dropped",
			)
			ch <- prometheus.MustNewConstMetric(
				trafficPpsDesc,
				prometheus.GaugeValue,
				float64(sumFwdPackets)/float64(rw),
				collector.segment.TrafficType, key, "forwarded",
			)
			ch <- prometheus.MustNewConstMetric(
				trafficPpsDesc,
				prometheus.GaugeValue,
				float64(sumDropPackets)/float64(rw),
				collector.segment.TrafficType, key, "dropped",
			)
		}
	}
}

func init() {
	segment := &ToptalkersMetrics{}
	segments.RegisterSegment("toptalkers_metrics", segment)
}
