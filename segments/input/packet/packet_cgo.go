//go:build cgo
// +build cgo

package packet

import (
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

const cgoEnabled = true
const pfringEnabled = false

func getPcapHandle(source string, filter string) *pcap.Handle {
	inactive, err := pcap.NewInactiveHandle(source)
	if err != nil {
		log.Fatalf("[error]: Packet: Could not setup libpcap capture: %v", err)
	}
	defer inactive.CleanUp()

	handle, err := inactive.Activate()
	if err != nil {
		log.Fatalf("[error]: Packet: Could not initiate capture: %v", err)
	}

	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatalf("[error]: Packet: Could not set BPF filter: %v", err)
	}
	return handle
}

func getPcapFile(source string, filter string) *pcap.Handle {
	var handle *pcap.Handle
	var err error
	if handle, err = pcap.OpenOffline(source); err != nil {
		log.Fatalf("[error]: Packet: Could not setup pcap reader: %v", err)
	} else if err := handle.SetBPFFilter(filter); err != nil {
		log.Fatalf("[error]: Packet: Could not set BPF filter: %v", err)
	}
	if err != nil {
		log.Fatal(err)
	}
	return handle
}

type dummyHandle interface {
	LinkType() layers.LinkType
	Close()
	ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error)
}

var getPfringHandle func(source string, filter string) dummyHandle
