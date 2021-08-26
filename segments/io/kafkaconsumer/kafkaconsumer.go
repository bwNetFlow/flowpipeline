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
	Server string // required
	Topic  string // required
	Group  string // required
	User   string // required if auth is true
	Pass   string // required if auth is true
	Tls    bool   // optional, default is true
	Auth   bool   // optional, default is true
}

func (segment KafkaConsumer) New(config map[string]string) segments.Segment {
	if config["server"] == "" || config["topic"] == "" || config["group"] == "" {
		log.Println("[error] KafkaConsumer: Missing required configuration parameters.")
		return nil
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

	if auth && (config["user"] == "" || config["pass"] == "") {
		log.Println("[error] KafkaConsumer: Missing required configuration parameters for auth.")
		return nil
	}

	return &KafkaConsumer{
		Server: config["server"],
		Topic:  config["topic"],
		Group:  config["group"],
		User:   config["user"],
		Pass:   config["pass"],
		Tls:    tls,
		Auth:   auth,
	}
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

	err := kafkaConn.StartConsumer(segment.Server, strings.Split(segment.Topic, ","), segment.Group, -1)
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
