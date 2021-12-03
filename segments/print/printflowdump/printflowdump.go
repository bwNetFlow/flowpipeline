// Prints all incoming flows in a specific flowdump format.
// Add the protomap segment before this segment to enrich any flow message
// with human readable protocol names instead of the protocol numbers.
// This segment currently has no way to configure the output format (TODO).
package printflowdump

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/bwNetFlow/flowpipeline/segments"
	"github.com/bwNetFlow/flowpipeline/segments/modify/protomap"
	flow "github.com/bwNetFlow/protobuf/go"
	"github.com/dustin/go-humanize"
)

type PrintFlowdump struct {
	segments.BaseSegment
	UseProtoname bool // optional, default is true
}

func (segment *PrintFlowdump) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	for msg := range segment.In {
		fmt.Println(segment.format_flow(msg))
		segment.Out <- msg
	}
}

func (segment PrintFlowdump) New(config map[string]string) segments.Segment {
	var useProtoname bool = true
	if config["useprotoname"] != "" {
		if parsedUseProtoname, err := strconv.ParseBool(config["useprotoname"]); err == nil {
			useProtoname = parsedUseProtoname
		} else {
			log.Println("[error] PrintFlowdump: Could not parse 'useprotoname' parameter, using default true.")
		}
	} else {
		log.Println("[info] PrintFlowdump: 'useprotoname' set to default true.")
	}
	return &PrintFlowdump{UseProtoname: useProtoname}
}

func (segment PrintFlowdump) format_flow(flowmsg *flow.FlowMessage) string {
	timestamp := time.Unix(int64(flowmsg.TimeFlowEnd), 0).Format("15:04:05")
	src := net.IP(flowmsg.SrcAddr)
	dst := net.IP(flowmsg.DstAddr)
	router := net.IP(flowmsg.SamplerAddress)
	var proto string
	if segment.UseProtoname {
		if flowmsg.ProtoName != "" {
			proto = flowmsg.ProtoName
		} else {
			// use function from another segment, as it is just a lookup.
			proto = protomap.ProtoNumToString(flowmsg.Proto)
		}
		if proto == "ICMP" && flowmsg.DstPort != 0 {
			proto = fmt.Sprintf("ICMP (type %d, code %d)", flowmsg.DstPort/256, flowmsg.DstPort%256)
		}
	} else {
		proto = fmt.Sprint(flowmsg.Proto)
	}

	duration := flowmsg.TimeFlowEnd - flowmsg.TimeFlowStart
	if duration == 0 {
		duration += 1
	}
	return fmt.Sprintf("%s: %s:%d -> %s:%d [%s → @%s → %s], %s, %ds, %s, %s",
		timestamp, src, flowmsg.SrcPort, dst, flowmsg.DstPort,
		flowmsg.SrcIfDesc, router, flowmsg.DstIfDesc, proto,
		duration, humanize.SI(float64(flowmsg.Bytes*8/duration),
			"bps"), humanize.SI(float64(flowmsg.Packets/duration), "pps"))
}

func init() {
	segment := &PrintFlowdump{}
	segments.RegisterSegment("printflowdump", segment)
}
