// Determines the remote address of flows based on different criteria.
//
// This segment basically runs flows through a switch case directive using its
// only config parameter, 'flowsrc'.
//
// 'border' assumes flows are exported on border interfaces: If a flow's
// direction is 'ingress' on such an interface, its remote address is the
// source address of the flow, whereas the local address of the flow would be
// its destination address inside our network. The same logic applies vice
// versa: 'egress' flows have a remote destination address.
//
// 'user' assumes flows are exported on user interfaces: If a flow's
// direction is 'ingress' on such an interface, its remote address is the
// destination address of the flow, whereas the local address of the flow would be
// its destination address inside our user's network. The same logic applies
// vice versa: 'egress' flows have a remote source address.
//
// 'mixed' assumes flows are exported whereever, and thus all remote address
// info is cleared in this case.
package remoteaddress

import (
	"log"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
)

type RemoteAddress struct {
	segments.BaseSegment
	FlowSrc string // required, 'border', 'user' and 'mixed' are available options, see above
}

func (segment RemoteAddress) New(config map[string]string) segments.Segment {
	if !(config["flowsrc"] == "border" || config["flowsrc"] == "user" || config["flowsrc"] == "mixed") {
		log.Println("[error] DropFields: The 'policy' parameter is required to be one of 'border', 'user', or 'mixed'.")
		return nil
	}
	return &RemoteAddress{
		FlowSrc: config["flowsrc"],
	}
}

func (segment *RemoteAddress) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	for msg := range segment.In {
		switch segment.FlowSrc {
		case "border":
			switch {
			case msg.FlowDirection == 0: // flow is ingress on border interface
				msg.RemoteAddr = 1 // thus, RemoteAddr should indicate SrcAddr
			case msg.FlowDirection == 1: // flow is egress on border interface
				msg.RemoteAddr = 2 // thus, RemoteAddr should indicate DstAddr
			}
		case "user":
			switch {
			case msg.FlowDirection == 0: // flow is ingress on user interface
				msg.RemoteAddr = 2 // thus, RemoteAddr should indicate DstAddr
			case msg.FlowDirection == 1: // flow is egress on user interface
				msg.RemoteAddr = 1 // thus, RemoteAddr should indicate SrcAddr
			}
		case "mixed":
			msg.RemoteAddr = 0 // reset previous info, we can't tell in a mixed env
		}
		segment.Out <- msg
	}
}

func init() {
	segment := &RemoteAddress{}
	segments.RegisterSegment("remoteaddress", segment)
}
