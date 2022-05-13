package kafkaconsumer

import (
	"context"
	"log"

	"github.com/Shopify/sarama"
	"github.com/bwNetFlow/flowpipeline/pb"
	"google.golang.org/protobuf/proto"
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
	for message := range claim.Messages() {
		session.MarkMessage(message, "")
		flowMsg := new(pb.EnrichedFlow)
		if err := proto.Unmarshal(message.Value, flowMsg); err == nil {
			h.flows <- flowMsg
		} else {
			log.Printf("[warning] KafkaConsumer: Error decoding flow, this might be due to the use of Goflow custom fields. Original error:\n  %s", err)
		}
	}
	return nil
}
