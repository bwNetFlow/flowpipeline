// Captures Netflow v9 and feeds flows to the following segments.
package goflow

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
	flow "github.com/bwNetFlow/protobuf/go"
	"google.golang.org/protobuf/proto"

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
	// TODO: can we shave of this Unmarshal here and the Marshal in line 95
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

type myProtobufDriver struct {
}

func (d *myProtobufDriver) Format(data interface{}) ([]byte, []byte, error) {
	msg, ok := data.(proto.Message)
	if !ok {
		return nil, nil, fmt.Errorf("message is not protobuf")
	}
	// TODO: can we shave of this Marshal here and the Unmarshal in line 72
	b, err := proto.Marshal(msg)
	return nil, b, err
}

func (d *myProtobufDriver) Prepare() error             { return nil }
func (d *myProtobufDriver) Init(context.Context) error { return nil }

func (segment *Goflow) startGoFlow(transport transport.TransportInterface) {
	formatter := &myProtobufDriver{}
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
