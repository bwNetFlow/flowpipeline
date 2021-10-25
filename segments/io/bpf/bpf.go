package bpf

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/bwNetFlow/bpf_flowexport/flowexport"
	"github.com/bwNetFlow/bpf_flowexport/packetdump"
	"github.com/bwNetFlow/flowpipeline/segments"
)

type Bpf struct {
	segments.BaseSegment

	dumper   *packetdump.PacketDumper
	exporter *flowexport.FlowExporter

	Device          string // required, the name of the device to capture, e.g. "eth0"
	ActiveTimeout   string // optional, default is 30m
	InactiveTimeout string // optional, default is 15s
}

func (segment Bpf) New(config map[string]string) segments.Segment {
	newsegment := &Bpf{}

	var ok bool
	newsegment.Device, ok = config["device"]
	if !ok {
		log.Printf("[error] Bpf: setting the config parameter 'device' is required.")
		return nil
	}

	// setup bpf dumping
	newsegment.dumper = &packetdump.PacketDumper{}
	err := newsegment.dumper.Setup(newsegment.Device)
	if err != nil {
		log.Printf("[error] Bpf: error setting up BPF dumping: %s", err)
		return nil
	}

	// setup flow export
	_, err = time.ParseDuration(config["activetimeout"])
	if err != nil {
		if config["activetimeout"] == "" {
			log.Println("[info] Bpf: 'activetimeout' set to default '30m'.")
		} else {
			log.Println("[warning] Bpf: 'activetimeout' was invalid, fallback to default '30m'.")
		}
		newsegment.ActiveTimeout = "30m"
	}
	_, err = time.ParseDuration(config["inactivetimeout"])
	if err != nil {
		if config["inactivetimeout"] == "" {
			log.Println("[info] Bpf: 'inactivetimeout' set to default '15s'.")
		} else {
			log.Println("[warning] Bpf: 'inactivetimeout' was invalid, fallback to default '15s'.")
		}
		newsegment.InactiveTimeout = "15s"
	}
	newsegment.exporter, err = flowexport.NewFlowExporter(newsegment.ActiveTimeout, newsegment.InactiveTimeout)
	if err != nil {
		log.Printf("[error] Bpf: error setting up exporter: %s", err)
	}
	return newsegment
}

func (segment *Bpf) Run(wg *sync.WaitGroup) {
	err := segment.dumper.Start()
	if err != nil {
		log.Printf("[error] Bpf: error starting up BPF dumping: %s", err)
		os.Exit(1)
	}
	segment.exporter.Start()
	go segment.exporter.ConsumeFrom(segment.dumper.Packets())
	defer func() {
		close(segment.Out)
		segment.dumper.Stop()
		wg.Done()
	}()

	log.Printf("[info] Bpf: Startup finished, exporting flows from '%s'", segment.Device)
	for {
		select {
		case msg, ok := <-segment.exporter.Flows:
			if !ok {
				return
			}
			segment.Out <- msg
		case msg, ok := <-segment.In:
			if !ok {
				segment.exporter.Stop()
				return
			}
			segment.Out <- msg
		}
	}
}

func init() {
	segment := &Bpf{}
	segments.RegisterSegment("bpf", segment)
}
