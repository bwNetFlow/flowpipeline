package kafkaconsumer

import (
	"context"
	"log"

	"github.com/Shopify/sarama"
	"github.com/bwNetFlow/flowpipeline/pb"
	oldpb "github.com/bwNetFlow/protobuf/go"
	"github.com/golang/protobuf/proto"
	goflowpb "github.com/netsampler/goflow2/pb"
)

// Handler represents a Sarama consumer group consumer
type Handler struct {
	ready  chan bool
	flows  chan *pb.EnrichedFlow
	cancel context.CancelFunc
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *Handler) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(h.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *Handler) Cleanup(sarama.ConsumerGroupSession) error {
	close(h.flows)
	return nil
}

func (h *Handler) Close() {
	h.cancel()
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	var loggedFormatDetails bool
	for message := range claim.Messages() {
		session.MarkMessage(message, "")
		flowMsg := new(pb.EnrichedFlow)
		goflowMsg := new(goflowpb.FlowMessage)
		oldFlowMsg := new(oldpb.FlowMessage)
		if err := proto.Unmarshal(message.Value, oldFlowMsg); err == nil {
			if !loggedFormatDetails {
				log.Println("[warning] KafkaConsumer: This Kafka cluster still uses the old flow format, consider updating its generator.")
				loggedFormatDetails = true
			}
			if err := proto.Unmarshal(message.Value, goflowMsg); err != nil {
				log.Println("[error] KafkaConsumer: Found old format that is not parsable using goflow format.")
				continue
			}
			h.flows <- &pb.EnrichedFlow{
				Core:          goflowMsg,
				Cid:           oldFlowMsg.Cid,
				CidString:     oldFlowMsg.CidString,
				SrcCid:        oldFlowMsg.SrcCid,
				DstCid:        oldFlowMsg.DstCid,
				Normalized:    pb.EnrichedFlow_NormalizedType(oldFlowMsg.Normalized),
				SrcIfName:     oldFlowMsg.SrcIfName,
				SrcIfDesc:     oldFlowMsg.SrcIfDesc,
				SrcIfSpeed:    oldFlowMsg.SrcIfSpeed,
				DstIfName:     oldFlowMsg.DstIfName,
				DstIfDesc:     oldFlowMsg.DstIfDesc,
				DstIfSpeed:    oldFlowMsg.DstIfSpeed,
				ProtoName:     oldFlowMsg.ProtoName,
				RemoteCountry: oldFlowMsg.RemoteCountry,
				SrcCountry:    oldFlowMsg.SrcCountry,
				DstCountry:    oldFlowMsg.DstCountry,
				RemoteAddr:    pb.EnrichedFlow_RemoteAddrType(oldFlowMsg.RemoteAddr),
				Note:          oldFlowMsg.Note,
			}
		} else if err := proto.Unmarshal(message.Value, goflowMsg); err == nil {
			if !loggedFormatDetails {
				log.Println("[info] KafkaConsumer: This Kafka cluster uses Goflow's flow format, converting.")
				loggedFormatDetails = true
			}
			h.flows <- &pb.EnrichedFlow{Core: goflowMsg}
		} else if err := proto.Unmarshal(message.Value, flowMsg); err == nil {
			h.flows <- flowMsg
		}
	}
	return nil
}
