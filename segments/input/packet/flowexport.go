//go:build linux
// +build linux

package packet

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type FlowKey struct {
	SrcAddr string
	DstAddr string
	SrcPort uint16
	DstPort uint16
	Proto   uint32
	IPTos   uint8
	InIface uint32
}

func NewFlowKey(packet gopacket.Packet) FlowKey {
	fkey := FlowKey{}
	fkey.InIface = uint32(packet.Metadata().InterfaceIndex)
	for _, layer := range packet.Layers() {
		switch layer.LayerType() {
		case layers.LayerTypeIPv4:
			ip, _ := layer.(*layers.IPv4)
			fkey.SrcAddr = string(ip.SrcIP.To16())
			fkey.DstAddr = string(ip.DstIP.To16())
			fkey.Proto = uint32(ip.Protocol)
			fkey.IPTos = ip.TOS
		case layers.LayerTypeIPv6:
			ip, _ := layer.(*layers.IPv6)
			fkey.SrcAddr = string(ip.SrcIP)
			fkey.DstAddr = string(ip.DstIP)
			fkey.Proto = uint32(ip.NextHeader)
			fkey.IPTos = ip.TrafficClass
		case layers.LayerTypeTCP:
			tcp, _ := layer.(*layers.TCP)
			fkey.SrcPort = uint16(tcp.SrcPort)
			fkey.DstPort = uint16(tcp.DstPort)
		case layers.LayerTypeUDP:
			udp, _ := layer.(*layers.UDP)
			fkey.SrcPort = uint16(udp.SrcPort)
			fkey.DstPort = uint16(udp.DstPort)
		}
	}
	return fkey
}

type FlowRecord struct {
	TimeReceived    time.Time
	LastUpdated     time.Time
	SamplerAddress  net.IP
	HardwareAddress net.HardwareAddr
	Packets         []gopacket.Packet
}

func BuildFlow(f *FlowRecord) *pb.EnrichedFlow {
	msg := &pb.EnrichedFlow{}
	msg.Type = pb.EnrichedFlow_EBPF
	msg.SamplerAddress = f.SamplerAddress
	msg.TimeReceived = uint64(f.TimeReceived.Unix())
	msg.TimeFlowStart = uint64(f.TimeReceived.Unix())
	msg.TimeFlowEnd = uint64(f.LastUpdated.Unix())
	for i, pkt := range f.Packets {
		if i == 0 {
			msg.InIf = uint32(pkt.Metadata().InterfaceIndex)
			for _, layer := range pkt.Layers() {
				switch layer.LayerType() {
				case layers.LayerTypeEthernet:
					eth, _ := layer.(*layers.Ethernet)
					msg.Etype = uint32(eth.EthernetType)
					if bytes.Equal(eth.SrcMAC, f.HardwareAddress) {
						msg.FlowDirection = 1                              // egress
						msg.RemoteAddr = pb.EnrichedFlow_RemoteAddrType(2) // src is remote
					}
					if bytes.Equal(eth.DstMAC, f.HardwareAddress) {
						msg.FlowDirection = 0                              // ingress
						msg.RemoteAddr = pb.EnrichedFlow_RemoteAddrType(1) // dst is remote
					}
				case layers.LayerTypeIPv4:
					ip, _ := layer.(*layers.IPv4)
					msg.SrcAddr = ip.SrcIP
					msg.DstAddr = ip.DstIP
					msg.Proto = uint32(ip.Protocol)
					msg.IPTos = uint32(ip.TOS)
					msg.IPTTL = uint32(ip.TTL)
				case layers.LayerTypeIPv6:
					ip, _ := layer.(*layers.IPv6)
					msg.SrcAddr = ip.SrcIP
					msg.DstAddr = ip.DstIP
					msg.Proto = uint32(ip.NextHeader)
					msg.IPTos = uint32(ip.TrafficClass)
					msg.IPTTL = uint32(ip.HopLimit)
					msg.IPv6FlowLabel = ip.FlowLabel
				case layers.LayerTypeTCP:
					tcp, _ := layer.(*layers.TCP)
					msg.SrcPort = uint32(tcp.SrcPort)
					msg.DstPort = uint32(tcp.DstPort)
					if tcp.URG {
						msg.TCPFlags = msg.TCPFlags | 0b100000
					}
					if tcp.ACK {
						msg.TCPFlags = msg.TCPFlags | 0b010000
					}
					if tcp.PSH {
						msg.TCPFlags = msg.TCPFlags | 0b001000
					}
					if tcp.RST {
						msg.TCPFlags = msg.TCPFlags | 0b000100
					}
					if tcp.SYN {
						msg.TCPFlags = msg.TCPFlags | 0b000010
					}
					if tcp.FIN {
						msg.TCPFlags = msg.TCPFlags | 0b000001
					}
				case layers.LayerTypeUDP:
					udp, _ := layer.(*layers.UDP)
					msg.SrcPort = uint32(udp.SrcPort)
					msg.DstPort = uint32(udp.DstPort)
				case layers.LayerTypeICMPv4:
					icmp, _ := layer.(*layers.ICMPv4)
					msg.IcmpType = uint32(icmp.TypeCode.Type())
					msg.IcmpCode = uint32(icmp.TypeCode.Code())
				case layers.LayerTypeICMPv6:
					icmp, _ := layer.(*layers.ICMPv6)
					msg.IcmpType = uint32(icmp.TypeCode.Type())
					msg.IcmpCode = uint32(icmp.TypeCode.Code())
				}
			}
		}
		// special handling
		msg.Bytes += uint64(pkt.Metadata().Length)
		msg.Packets += 1
	}
	return msg
}

type FlowExporter struct {
	activeTimeout   time.Duration
	inactiveTimeout time.Duration
	samplerAddress  net.IP
	hardwareAddress net.HardwareAddr

	Flows chan *pb.EnrichedFlow

	mutex *sync.RWMutex
	stop  chan bool
	cache map[FlowKey]*FlowRecord
}

func NewFlowExporter(activeTimeout string, inactiveTimeout string) (*FlowExporter, error) {
	activeTimeoutDuration, err := time.ParseDuration(activeTimeout)
	if err != nil {
		return nil, fmt.Errorf("active timeout misconfigured")
	}
	inactiveTimeoutDuration, err := time.ParseDuration(inactiveTimeout)
	if err != nil {
		return nil, fmt.Errorf("inactive timeout misconfigured")
	}

	fe := &FlowExporter{activeTimeout: activeTimeoutDuration, inactiveTimeout: inactiveTimeoutDuration}
	fe.Flows = make(chan *pb.EnrichedFlow)

	fe.mutex = &sync.RWMutex{}
	fe.cache = make(map[FlowKey]*FlowRecord)

	return fe, nil
}

func (f *FlowExporter) Start(samplerAddress net.IP, hardwareAddress net.HardwareAddr) {
	log.Println("[info] FlowExporter: Starting export goroutines.")

	f.samplerAddress = samplerAddress
	f.hardwareAddress = hardwareAddress
	f.stop = make(chan bool)
	go f.exportInactive()
	go f.exportActive()
}

func (f *FlowExporter) Stop() {
	log.Println("[info] FlowExporter: Stopping export goroutines.")
	close(f.stop)
}

func (f *FlowExporter) exportInactive() {
	ticker := time.NewTicker(f.inactiveTimeout)
	for {
		select {
		case <-ticker.C:
			now := time.Now()

			f.mutex.Lock()
			for key, record := range f.cache {
				if now.Sub(record.LastUpdated) > f.inactiveTimeout {
					f.export(key)
				}
			}
			f.mutex.Unlock()
		case <-f.stop:
			ticker.Stop()
			return
		}
	}
}

func (f *FlowExporter) exportActive() {
	ticker := time.NewTicker(f.activeTimeout)
	for {
		select {
		case <-ticker.C:
			now := time.Now()

			f.mutex.Lock()
			for key, record := range f.cache {
				if now.Sub(record.TimeReceived) > f.activeTimeout {
					f.export(key)
				}
			}
			f.mutex.Unlock()
		case <-f.stop:
			ticker.Stop()
			return
		}
	}
}

func (f *FlowExporter) Insert(pkt gopacket.Packet) {
	key := NewFlowKey(pkt)

	var record *FlowRecord
	var exists bool

	f.mutex.Lock()
	if record, exists = f.cache[key]; !exists {
		f.cache[key] = new(FlowRecord)
		f.cache[key].TimeReceived = time.Now()
		record = f.cache[key]
	}
	record.LastUpdated = time.Now()
	record.SamplerAddress = f.samplerAddress
	record.Packets = append(record.Packets, pkt)

	// shortcut flow export if we see TCP FIN
	if tcpLayer := pkt.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		if tcp.FIN {
			f.export(key)
		}
	}
	f.mutex.Unlock()
}

func (f *FlowExporter) ConsumeFrom(pkts chan gopacket.Packet) {
	for {
		select {
		case pkt, ok := <-pkts:
			if !ok {
				return
			}
			f.Insert(pkt)
		case <-f.stop:
			return
		}
	}
}

func (f *FlowExporter) export(key FlowKey) {
	flowRecord := f.cache[key]
	delete(f.cache, key)

	f.Flows <- BuildFlow(flowRecord)
}
