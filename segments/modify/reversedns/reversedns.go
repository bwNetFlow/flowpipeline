// Rewrites the Note field of passing flows to the remote addresses reverse DNS entry.
package reversedns

import (
	"log"
	"net"
	"strings"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
)

type ReverseDns struct {
	segments.BaseSegment
}

func (segment ReverseDns) New(config map[string]string) segments.Segment {
	return &ReverseDns{}
}

func (segment *ReverseDns) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	for msg := range segment.In {
		var remoteAddress string
		switch msg.RemoteAddr {
		case 1:
			remoteAddress = net.IP(msg.SrcAddr).String()
		case 2:
			remoteAddress = net.IP(msg.DstAddr).String()
		default:
			segment.Out <- msg
			continue
		}

		dnsName, err := net.LookupAddr(remoteAddress)
		if err == nil {
			msg.Note = strings.Join(dnsName, ",")
		} else {
			log.Println(err)
		}
		segment.Out <- msg
	}
}

func init() {
	segment := &ReverseDns{}
	segments.RegisterSegment("reversedns", segment)
}
