//go:build linux
// +build linux

package packet

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwNetFlow/flowpipeline/segments"
	"github.com/bwNetFlow/flowpipeline/segments/filter/aggregate"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	// "github.com/google/gopacket/pfring"
)

type Packet struct {
	segments.BaseSegment

	exporter *aggregate.FlowExporter

	Method          string // required, default is "pcap", one of the available capture methods "pcapgo|pcap|pfring|file"
	Source          string // required, the name of the source to capture from, depending on the method an interface or file name is required
	Filter          string // optional, a BPF filter which is applied when using a libpcap-based method
	ActiveTimeout   string // optional, default is 30m
	InactiveTimeout string // optional, default is 15s
}

type opt struct {
	Name    string
	Default string
	Options []string
	Type    string
}

type conf map[string]string

func (c conf) parseOption(o opt) (error, string) {
	switch o.Type {
	case "option":
		if c[o.Name] == "" {
			if o.Default != "" {
				log.Printf("[info] Packet: '%s' set to default '%s'.", o.Name, o.Default)
				return nil, o.Default
			} else {
				log.Printf("[error] Packet: Field '%s' is required.", o.Name)
				return fmt.Errorf("parse error"), ""
			}
		}
		for _, option := range o.Options {
			if c[o.Name] == option {
				return nil, option
			}
		}
		log.Printf("[error] Packet: Field '%s' must be set to a valid option: %s.", o.Name, strings.Join(o.Options, "|"))
		return fmt.Errorf("parse error"), ""
	case "duration":
		_, err := time.ParseDuration(c[o.Name])
		if err != nil {
			if c[o.Name] == "" {
				log.Printf("[info] Packet: '%s' set to default '%s'.", o.Name, o.Default)
			} else {
				log.Printf("[warning] Packet: '%s' was invalid, fallback to default '%s'.", o.Name, o.Default)
			}
			return nil, o.Default
		} else {
			log.Printf("[info] Packet: '%s' set to '%s'.", o.Name, c[o.Name])
			return nil, c[o.Name]
		}
	case "rfile":
		if _, err := os.Stat(c[o.Name]); err != nil {
			if o.Default != "" {
				if _, err := os.Stat(o.Default); err != nil {
					log.Printf("[info] Packet: '%s' set to default '%s'.", o.Name, o.Default)
					return nil, o.Default
				}
				return err, ""
			} else {
				log.Printf("[error] Packet: Field '%s' is required.", o.Name)
				return err, ""
			}
		}
		return nil, c[o.Name]
	case "iface":
		if c[o.Name] == "" {
			if o.Default != "" {
				log.Printf("[info] Packet: '%s' set to default '%s'.", o.Name, o.Default)
				return nil, o.Default
			} else {
				log.Printf("[error] Packet: Field '%s' is required.", o.Name)
				return fmt.Errorf("parse error"), ""
			}
		}
		if _, err := net.InterfaceByName(c[o.Name]); err == nil {
			return nil, c[o.Name]
		}
		log.Printf("[error] Packet: Field '%s' must be set to a valid interface.", o.Name)
		return fmt.Errorf("parse error"), ""
	}
	return fmt.Errorf("parse error"), ""
}

func (segment Packet) New(config map[string]string) segments.Segment {
	newsegment := &Packet{}

	// setup flow export
	var err error
	c := conf(config)
	if err, newsegment.ActiveTimeout = c.parseOption(opt{"activetimeout", "30m", []string{}, "duration"}); err != nil {
		return nil
	}
	if err, newsegment.InactiveTimeout = c.parseOption(opt{"inactivetimeout", "15s", []string{}, "duration"}); err != nil {
		return nil
	}
	if err, newsegment.Method = c.parseOption(opt{"method", "pcapgo", []string{"pcapgo", "pcap", "pfring", "file"}, "option"}); err != nil {
		return nil
	}
	if newsegment.Method == "file" {
		if err, newsegment.Source = c.parseOption(opt{"source", "", []string{}, "rfile"}); err != nil {
			return nil
		}
	} else {
		if err, newsegment.Source = c.parseOption(opt{"source", "", []string{}, "iface"}); err != nil {
			return nil
		}
	}

	if cgoEnabled && config["filter"] != "" {
		log.Printf("[info] Packet: Using BPF filter '%s' on packet stream, flows will be generated matches only.", config["filter"])
		newsegment.Filter = config["filter"] // this might be a Run()-time error later on
	} else if config["filter"] != "" {
		log.Println("[warning] Packet: Parameter 'filter' has been ignored as this requires a binary with CGO enabled.")
	}

	newsegment.exporter, err = aggregate.NewFlowExporter(newsegment.ActiveTimeout, newsegment.InactiveTimeout)
	if err != nil {
		log.Printf("[error] Packet: error setting up exporter: %s", err)
		return nil
	}
	return newsegment
}

func (segment *Packet) Run(wg *sync.WaitGroup) {
	var pktsrc *gopacket.PacketSource
	switch segment.Method {
	case "pcapgo":
		handle, err := pcapgo.NewEthernetHandle(segment.Source)
		if err != nil {
			log.Fatalf("[error]: Packet: Could not initiate capture: %v", err)
		}

		pktsrc = gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)

	case "pcap":
		if !cgoEnabled {
			log.Fatalln("[error]: Packet: CGO feature 'pcap' requested from binary compiled without CGO support.")
		} else {
			handle := getPcapHandle(segment.Source, segment.Filter)
			defer handle.Close()
			pktsrc = gopacket.NewPacketSource(handle, handle.LinkType())
		}
	case "pfring":
		if !cgoEnabled {
			log.Fatalln("[error]: Packet: CGO feature 'pcap' requested from binary compiled without CGO support.")
		} else {
			ring := getPfringHandle(segment.Source, segment.Filter)
			defer ring.Close()
			pktsrc = gopacket.NewPacketSource(ring, layers.LinkTypeEthernet)
		}
	case "file":
		f, err := os.Open(segment.Source)
		if err != nil {
			log.Fatalf("[error]: Packet: Could not open file: %v", err)
		}
		defer f.Close()

		if segment.Filter != "" && cgoEnabled {
			handle := getPcapFile(segment.Source, segment.Filter)
			pktsrc = gopacket.NewPacketSource(handle, handle.LinkType())
		} else if handle, err := pcapgo.NewNgReader(f, pcapgo.DefaultNgReaderOptions); err == nil {
			pktsrc = gopacket.NewPacketSource(handle, handle.LinkType())
		} else {
			if err.Error() == "Unknown magic a1b2c3d4" && cgoEnabled {
				if cgoEnabled {
					log.Printf("[warning]: Packet: Legacy pcap file detected, falling back to using libpcap instead of pure-Go implementation.")
					handle := getPcapFile(segment.Source, segment.Filter)
					pktsrc = gopacket.NewPacketSource(handle, handle.LinkType())
				} else {
					log.Fatalf("[error]: Packet: Could not read legacy pcap file, and classic libpcap is unavailable.")
				}
			} else {
				log.Fatalf("[error]: Packet: Could not read capture from file: %v", err)
			}
		}
	}

	if segment.Method == "file" {
		segment.exporter.Start(nil, nil)
	} else {
		iface, _ := net.InterfaceByName(segment.Source)
		var samplerAddress net.IP
		addrs, err := iface.Addrs()
		if err != nil {
			log.Fatalf("[error]: Packet: Could not determine sampler address: %v", err)
		}
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				samplerAddress = ipnet.IP
				break
			}
		}
		segment.exporter.Start(samplerAddress, iface.HardwareAddr)
	}

	go func() {
		segment.exporter.ConsumeFrom(pktsrc.Packets())
		if segment.Method == "file" {
			log.Println("[info] Packet: The pcap has ended.")
			os.Exit(0)
		} else {
			log.Fatalln("[error] Packet: The packet stream has ended for an unknown reason.")
		}
	}()

	defer func() {
		close(segment.Out)
		segment.exporter.Stop()
		wg.Done()
	}()
	for {
		select {
		case msg, ok := <-segment.exporter.Flows:
			if !ok {
				return
			}
			segment.Out <- msg
		case msg, ok := <-segment.In:
			if !ok {
				return
			}
			segment.Out <- msg
		}
	}
}

func init() {
	segment := &Packet{}
	segments.RegisterSegment("packet", segment)
}
