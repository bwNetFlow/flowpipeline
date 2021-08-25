package goflow

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
	"google.golang.org/protobuf/proto"

	formatter "github.com/netsampler/goflow2/format/protobuf"
	"github.com/netsampler/goflow2/transport"
	"github.com/netsampler/goflow2/utils"
)

type Goflow struct {
	segments.BaseSegment
	Port uint64

	goflow_in chan *flow.FlowMessage
}

func (segment Goflow) New(config map[string]string) segments.Segment {
	var port uint64 = 2055
	if config["port"] != "" {
		if parsedPort, err := strconv.ParseUint(config["port"], 10, 32); err == nil {
			port = parsedPort
		} else {
			log.Println("[error] Goflow: Could not parse 'port' parameter, using default 2055.")
		}
	} else {
		log.Println("[info] Goflow: 'port' set to default '2055'.")
	}
	return &Goflow{
		Port: port,
	}
}

func (segment *Goflow) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()
	segment.goflow_in = make(chan *flow.FlowMessage)
	go segment.startGoFlow(&channelDriver{segment.goflow_in})
	for {
		select {
		case msg, ok := <-segment.goflow_in:
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

type channelDriver struct {
	out chan *flow.FlowMessage
}

func (d *channelDriver) Send(key, data []byte) error {
	msg := &flow.FlowMessage{}
	if err := proto.Unmarshal(data, msg); err != nil {
		log.Println("[error] Goflow: Conversion error for received flow.")
		return nil
	}
	d.out <- msg
	return nil
}

func (d *channelDriver) Close(context.Context) error {
	close(d.out)
	return nil
}

func (segment *Goflow) startGoFlow(transport transport.TransportInterface) {
	formatter := &formatter.ProtobufDriver{}
	sNF := &utils.StateNetFlow{
		Format:    formatter,
		Transport: transport,
	}

	log.Printf("[info] Goflow: Listening for Netflow v9 on port %d...", segment.Port)
	err := sNF.FlowRoutine(1, "", int(segment.Port), false)
	if err != nil {
		log.Printf("[error] Goflow: Could not listen to UDP (%v)", err)
		os.Exit(1)
	}
}

func init() {
	segment := &Goflow{}
	segments.RegisterSegment("goflow", segment)
}
