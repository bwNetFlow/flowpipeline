package segments

import (
	"log"
	"sync"
)

type RemoteAddress struct {
	BaseSegment
	FlowSrc string
}

func (segment RemoteAddress) New(config map[string]string) Segment {
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
		close(segment.out)
		wg.Done()
	}()

	for msg := range segment.in {
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
		segment.out <- msg
	}
}

func init() {
	segment := &RemoteAddress{}
	RegisterSegment("remoteaddress", segment)
}
