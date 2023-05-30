package addrstrings

import (
	"github.com/bwNetFlow/flowpipeline/segments"
	"sync"
)

type AddrStrings struct {
	segments.BaseSegment
}

func (segment AddrStrings) New(config map[string]string) segments.Segment {
	return &AddrStrings{}
}

func (segment *AddrStrings) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	for original := range segment.In {
		// SourceIP
		if original.SrcAddr != nil {
			original.SourceIP = original.SrcAddrObj().String()
		}
		// DestinationIP
		if original.DstAddr != nil {
			original.DestinationIP = original.DstAddrObj().String()
		}
		// NextHopIP
		if original.NextHop != nil {
			original.NextHopIP = original.NextHopObj().String()
		}
		// SamplerIP
		if original.SamplerAddress != nil {
			original.SamplerIP = original.SamplerAddressObj().String()
		}
		// SourceMAC
		if original.SrcMac != 0x0 {
			original.SourceIP = original.SrcMacString()
		}
		// DestinationMAC
		if original.DstMac != 0x0 {
			original.DestinationMAC = original.DstMacString()
		}
		segment.Out <- original
	}
}

// register segment
func init() {
	segment := &AddrStrings{}
	segments.RegisterSegment("addrstrings", segment)
}
