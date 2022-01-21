// Consumes flows from a Kafka instance and passes them to the following
// segments. This segment is based on the kafkaconnector library:
// https://github.com/bwNetFlow/kafkaconnector
package kafkaconsumer

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
	kafka "github.com/bwNetFlow/kafkaconnector"
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

	var kafkaConn = kafka.Connector{}
	if !segment.Tls {
		kafkaConn.DisableTLS()
		log.Println("[info] KafkaConsumer: Disabled TLS, operating unencrypted.")
	}

	if !segment.Auth {
		kafkaConn.DisableAuth()
		log.Println("[info] KafkaConsumer: Disabled auth.")
	} else {
		kafkaConn.SetAuth(segment.User, segment.Pass)
		log.Printf("[info] KafkaConsumer: Authenticating as user '%s'.", segment.User)
	}

	err := kafkaConn.StartConsumer(segment.Server, strings.Split(segment.Topic, ","), segment.Group, segment.startingOffset)
	if err != nil {
		log.Fatalln("[error] KafkaConsumer: Error starting consumer, this usually indicates a misconfiguration (auth).")
	}

	// receive flows in a loop
	for {
		select {
		case msg := <-kafkaConn.ConsumerChannel():
			segment.Out <- msg
		case msg, ok := <-segment.In:
			if !ok {
				return
			}
			segment.Out <- msg
		}
	}
}

func init() {
	segment := &KafkaConsumer{}
	segments.RegisterSegment("kafkaconsumer", segment)
}
