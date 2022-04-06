// Consumes flows from a Kafka instance and passes them to the following
// segments. This segment is based on the kafkaconnector library:
// https://github.com/bwNetFlow/kafkaconnector
package kafkaconsumer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/bwNetFlow/flowpipeline/segments"

	oldflow "github.com/bwNetFlow/protobuf/go"
)

type KafkaConsumer struct {
	segments.BaseSegment
	Server         string // required
	Topic          string // required
	Group          string // required
	User           string // required if auth is true
	Pass           string // required if auth is true
	Tls            bool   // optional, default is true
	Auth           bool   // optional, default is true
	StartAt        string // optional, one of "oldest" or "newest", default is "newest"
	startingOffset int64  // optional, default is -1
}

func (segment KafkaConsumer) New(config map[string]string) segments.Segment {
	newsegment := &KafkaConsumer{}
	if config["server"] == "" || config["topic"] == "" || config["group"] == "" {
		log.Println("[error] KafkaConsumer: Missing required configuration parameters.")
		return nil
	} else {
		newsegment.Server = config["server"]
		newsegment.Topic = config["topic"]
		newsegment.Group = config["group"]
	}

	var tls bool = true
	if config["tls"] != "" {
		if parsedTls, err := strconv.ParseBool(config["tls"]); err == nil {
			tls = parsedTls
		} else {
			log.Println("[error] KafkaConsumer: Could not parse 'tls' parameter, using default true.")
		}
	} else {
		log.Println("[info] KafkaConsumer: 'tls' set to default true.")
	}
	newsegment.Tls = tls

	var auth bool = true
	if config["auth"] != "" {
		if parsedAuth, err := strconv.ParseBool(config["auth"]); err == nil {
			auth = parsedAuth
		} else {
			log.Println("[error] KafkaConsumer: Could not parse 'auth' parameter, using default true.")
		}
	} else {
		log.Println("[info] KafkaConsumer: 'auth' set to default true.")
	}
	newsegment.Auth = auth

	if auth && (config["user"] == "" || config["pass"] == "") {
		log.Println("[error] KafkaConsumer: Missing required configuration parameters for auth.")
		return nil
	} else {
		newsegment.User = config["user"]
		newsegment.Pass = config["pass"]
	}

	startAt := "newest"
	var startingOffset int64 = -1 // see sarama const OffsetNewest
	if config["startat"] != "" {
		if strings.ToLower(config["startat"]) == "oldest" {
			startAt = "oldest"
			startingOffset = -2 // see sarama const OffsetOldest
		} else if strings.ToLower(config["startat"]) != "newest" {
			log.Println("[error] KafkaConsumer: Could not parse 'startat' parameter, using default 'newest'.")
		}
	} else {
		log.Println("[info] KafkaConsumer: 'startat' set to default 'newest'.")
	}
	newsegment.startingOffset = startingOffset
	newsegment.StartAt = startAt
	return newsegment
}

func (segment *KafkaConsumer) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.Out)
		wg.Done()
	}()

	var err error
	// TODO: move config into object
	config := sarama.NewConfig()

	// TODO: make configurable
	config.Version, err = sarama.ParseKafkaVersion("2.4.0")
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	if segment.Tls {
		rootCAs, err := x509.SystemCertPool()
		if err != nil {
			log.Panicf("TLS Error: %v", err)
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = &tls.Config{RootCAs: rootCAs}
		log.Println("[info] KafkaConsumer: Disabled TLS, operating unencrypted.")
	}

	if segment.Auth {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = segment.User
		config.Net.SASL.Password = segment.Pass
		log.Printf("[info] KafkaConsumer: Authenticating as user '%s'.", segment.User)
	} else {
		config.Net.SASL.Enable = false
		log.Println("[info] KafkaConsumer: Disabled auth.")
	}

	// TODO: make configurable
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	config.Consumer.Offsets.Initial = segment.startingOffset

	client, err := sarama.NewConsumerGroup(strings.Split(segment.Server, ","), segment.Group, config)
	if err != nil {
		if client == nil {
			log.Fatalf("[error] KafkaConsumer: Creating Kafka client failed, this indicates an unreachable server or a SSL problem. Original error:\n  %v", err)
		} else {
			log.Fatalf("[error] KafkaConsumer: Creating Kafka consumer group failed while the connection was okay. Original error:\n  %v", err)
		}
	}

	// log.Fatalln("[error] KafkaConsumer: Error starting consumer, this usually indicates a misconfiguration (auth).")
	handlerCtx, handlerCancel := context.WithCancel(context.Background())
	var handler = &Handler{
		ready: make(chan bool),
		flows: make(chan *oldflow.FlowMessage),
	}
	handlerWg := sync.WaitGroup{}
	handlerWg.Add(1)
	go func() {
		defer handlerWg.Done()
		for {
			// This loop ensures recreation of our consumer session when server-side rebalances happen.
			if err := client.Consume(handlerCtx, strings.Split(segment.Topic, ","), handler); err != nil {
				log.Printf("[error] KafkaConsumer: Could not create new consumer session. Original error:\n  %v", err)
				time.Sleep(5 * time.Second) // TODO: although this never occured for me, make configurable
				continue
			}
			// check if context was cancelled, signaling that the consumer should stop
			if handlerCtx.Err() != nil {
				return
			}
			handler.ready = make(chan bool) // TODO: this is from the official example, not sure it is necessary in out case
		}
	}()
	<-handler.ready
	log.Println("[info] KafkaConsumer: Connected and operational")

	defer func() {
		handlerWg.Wait()
		if err = client.Close(); err != nil {
			log.Panicf("[error] KafkaConsumer: Error closing Kafka client: %v", err)
		}
	}()

	// receive flows in a loop
	for {
		select {
		case msg, ok := <-handler.flows:
			if !ok {
				// This will occur when the handler calls its Cleanup method
				handlerCancel() // This is in case the channel was closed somehow else, which shouldn't happen
				return
			}
			segment.Out <- msg
		case msg, ok := <-segment.In:
			if !ok {
				handlerCancel() // Trigger handler shutdown and cleanup
			} else {
				segment.Out <- msg
			}
		}
	}
}

func init() {
	segment := &KafkaConsumer{}
	segments.RegisterSegment("kafkaconsumer", segment)
}
