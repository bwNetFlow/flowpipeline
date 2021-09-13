// Prints all incoming flows in a specific flowdump format.
// Add the protomap segment before this segment to enrich any flow message
// with human readable protocol names instead of the protocol numbers.
// This segment currently has no way to configure the output format (TODO).
package printflowdump

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
	"github.com/dustin/go-humanize"
)

type PrintFlowdump struct {
	segments.BaseSegment
}

func (segment *PrintFlowdump) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	for msg := range segment.In {
		fmt.Println(format_flow(msg))
		segment.Out <- msg
	}
}

func (segment PrintFlowdump) New(config map[string]string) segments.Segment {
	return &PrintFlowdump{}
}

func format_flow(flowmsg *flow.FlowMessage) string {
	timestamp := time.Unix(int64(flowmsg.TimeFlowEnd), 0).Format("15:04:05")
	src := net.IP(flowmsg.SrcAddr)
	dst := net.IP(flowmsg.DstAddr)
	router := net.IP(flowmsg.SamplerAddress)
	proto := flowmsg.ProtoName
	if flowmsg.ProtoName == "" {
		proto = fmt.Sprint(flowmsg.Proto)
	}
	duration := flowmsg.TimeFlowEnd - flowmsg.TimeFlowStart
	if duration == 0 {
		duration += 1
	}
	return fmt.Sprintf("%s: %s:%d -> %s:%d [%s -> %s, @%s], %s, %ds, %s, %s",
		timestamp, src, flowmsg.SrcPort, dst, flowmsg.DstPort,
		flowmsg.SrcIfDesc, flowmsg.DstIfDesc, router, proto,
		duration, humanize.SI(float64(flowmsg.Bytes*8/duration),
			"bps"), humanize.SI(float64(flowmsg.Packets/duration), "pps"))
}

func init() {
	segment := &PrintFlowdump{}
	segments.RegisterSegment("printflowdump", segment)
}
