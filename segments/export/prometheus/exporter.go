package prometheus

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/bwNetFlow/flowpipeline/pb"
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

func (e *Exporter) Increment(flow *pb.EnrichedFlow) {
	var application string
	_, appGuess1 := filterPopularPorts(flow.GetSrcPort())
	_, appGuess2 := filterPopularPorts(flow.GetDstPort())
	if appGuess1 != "" {
		application = appGuess1
	} else if appGuess2 != "" {
		application = appGuess2
	}

	peer := flow.Peer()
	var remoteAS string
	if flow.GetFlowDirection() == 0 {
		remoteAS = nameThatAS(flow.GetSrcAS())
	} else {
		remoteAS = nameThatAS(flow.GetDstAS())
	}

	labels := prometheus.Labels{}
	for _, l := range e.labels {
		switch l {
		case "router":
			labels[l] = net.IP(flow.GetSamplerAddress()).String()
		case "ipversion":
			labels[l] = flow.IPVersionString()
		case "application":
			labels[l] = application
		case "protoname":
			labels[l] = fmt.Sprint(flow.GetProtoName())
		case "direction":
			labels[l] = flow.FlowDirectionString()
		case "peer":
			labels[l] = peer
		case "remoteas":
			labels[l] = remoteAS
		case "remotecountry":
			labels[l] = flow.GetRemoteCountry()
		case "src_port":
			labels[l] = fmt.Sprint(flow.SrcPort)
		case "dst_port":
			labels[l] = fmt.Sprint(flow.DstPort)
		case "src_addr":
			labels[l] = fmt.Sprint(flow.SrcAddr)
		case "dst_addr":
			labels[l] = fmt.Sprint(flow.DstAddr)
		default:
			labels[l] = ""
		}
	}

	e.kafkaMessageCount.Inc()
	// e.flowNumber.With(labels).Inc()
	// flowPackets.With(labels).Add(float64(flow.GetPackets()))
	e.flowBits.With(labels).Add(float64(flow.GetBytes()) * 8)
}

func (e *Exporter) IncrementCtrl(topic string, partition int32, offset int64) {
	labels := prometheus.Labels{
		"topic":     topic,
		"partition": fmt.Sprint(partition),
	}
	e.kafkaOffsets.With(labels).Add(float64(offset))
}

func filterPopularPorts(port uint32) (uint32, string) {
	switch port {
	case 80:
		return port, "http"
	case 443:
		return port, "https"
	case 20, 21:
		return port, "ftp"
	case 22:
		return port, "ssh"
	case 23:
		return port, "telnet"
	case 53:
		return port, "dns"
	case 25, 465:
		return port, "smtp"
	case 110, 995:
		return port, "pop3"
	case 143, 993:
		return port, "imap"
	}
	return 0, ""
}

func nameThatAS(asn uint32) string {
	asnmap := map[uint32]string{
		43:     "Brookhaven National Laboratory",
		70:     "National Library Medicine USA",
		174:    "Cogent",
		513:    "CERN",
		559:    "SWITCH",
		680:    "DFN",
		702:    "Verizon",
		714:    "Apple",
		786:    "JANET",
		1239:   "Sprint",
		1273:   "Vodafone",
		1297:   "CERN",
		1299:   "Telia",
		1754:   "DESY",
		2018:   "AFRINIC",
		2603:   "NORDUnet",
		2906:   "Netflix",
		2914:   "NTT",
		3209:   "Vodafone",
		3257:   "GTT",
		3303:   "Swisscom",
		3320:   "Deutsche Telekom",
		3356:   "CenturyLink",
		4356:   "Epic Games",
		5430:   "freenet",
		5501:   "Fraunhofer",
		5511:   "Orange",
		6185:   "Apple",
		6453:   "TATA",
		6507:   "Riot Games",
		6724:   "Strato",
		6735:   "sdt.net",
		6805:   "Telefonica",
		6830:   "Vodafone",
		6939:   "Hurricane Electric",
		7018:   "AT&T",
		8068:   "Microsoft",
		8075:   "Microsoft",
		8220:   "Colt",
		8403:   "Spotify",
		8560:   "1&1",
		8674:   "Netnod",
		8763:   "DENIC",
		8881:   "Versatel",
		9009:   "GLOBALAXS",
		10310:  "Yahoo",
		13030:  "Init7",
		13335:  "Cloudflare",
		15133:  "Verizon",
		15169:  "Google",
		16276:  "OVH",
		16509:  "Amazon",
		16591:  "Google Fiber",
		16625:  "Akamai",
		19679:  "Dropbox",
		20446:  "Highwinds",
		20504:  "RTL",
		20677:  "imos",
		20940:  "Akamai",
		22822:  "Limelight",
		24429:  "Alibaba",
		24940:  "Hetzner",
		30361:  "Swiftwill",
		31334:  "Kabel Deutschland",
		32590:  "Valve Steam",
		32934:  "Facebook",
		33915:  "Vodafone",
		35402:  "ecotel",
		36459:  "Github",
		36561:  "Google",
		37963:  "Alibaba",
		39702:  "S-IT",
		41552:  "Ebay",
		41690:  "Dailymotion",
		46489:  "Twitch",
		48918:  "Globalways",
		54113:  "Fastly",
		54994:  "QUANTIL",
		57976:  "Blizzard",
		58069:  "KIT",
		60781:  "Leaseweb",
		61339:  "LHC",
		197540: "Netcup",
		197602: "TV-9",
		206339: "Schuler Pressen",
	}
	if name, ok := asnmap[asn]; ok {
		return name
	} else {
		if asn == 0 {
			return "unset"
		}
		return "other"
	}
}
