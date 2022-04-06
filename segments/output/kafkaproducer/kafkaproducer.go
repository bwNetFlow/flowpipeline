// Produces all received flows to Kafka instance. This segment is based on the
// kafkaconnector library:
// https://github.com/bwNetFlow/kafkaconnector
package kafkaproducer

import (
	"log"
	"reflect"
	"strconv"
	"sync"

	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
	kafka "github.com/bwNetFlow/kafkaconnector"
)

// All configuration parameters are the same as in the kafkaconsumer segment,
// except for the 'topicsuffix' parameter. This parameter, if set, acts as a
// suffix that is appended to the topic that this segment will produce a given
// flow to. As a static suffix would not make much sense, it is interpreted as
// a flow message field name, which will be used to create different topics
// based on field contents. For instance, setting `topicsuffix: Proto` will
// yield separate topics for each different protocol number occuring in all
// flows. Usually, a sensible application is usage with customer ids (`Cid`).
//
// For more info, see examples/splitter in the repo.

// FIXME: use sarama directly here
type KafkaProducer struct {
	segments.BaseSegment
	Server      string // required
	Topic       string // required
	TopicSuffix string // optional, default is empty
	User        string // required if auth is true
	Pass        string // required if auth is true
	Tls         bool   // optional, default is true
	Auth        bool   // optional, default is true
}

func (segment KafkaProducer) New(config map[string]string) segments.Segment {
	if config["server"] == "" || config["topic"] == "" {
		log.Println("[error] KafkaProducer: Missing required configuration parameters.")
		return nil
	}

	if config["topicsuffix"] != "" {
		fmsg := reflect.ValueOf(pb.EnrichedFlow{})
		field := fmsg.FieldByName(config["topicsuffix"])
		if !field.IsValid() {
			log.Println("[error] KafkaProducer: The 'topicsuffix' is not a valid FlowMessage field.")
			return nil
		}
	} else {
		log.Println("[info] KafkaProducer: 'topicsuffix' set to default disabled.")
	}

	var tls bool = true
	if config["tls"] != "" {
		if parsedTls, err := strconv.ParseBool(config["tls"]); err == nil {
			tls = parsedTls
		} else {
			log.Println("[error] KafkaProducer: Could not parse 'tls' parameter, using default true.")
		}
	} else {
		log.Println("[info] KafkaProducer: 'tls' set to default true.")
	}

	var auth bool = true
	if config["auth"] != "" {
		if parsedAuth, err := strconv.ParseBool(config["auth"]); err == nil {
			auth = parsedAuth
		} else {
			log.Println("[error] KafkaProducer: Could not parse 'auth' parameter, using default true.")
		}
	} else {
		log.Println("[info] KafkaProducer: 'auth' set to default true.")
	}

	if auth && (config["user"] == "" || config["pass"] == "") {
		log.Println("[error] KafkaProducer: Missing required configuration parameters for auth.")
		return nil
	}

	return &KafkaProducer{
		Server:      config["server"],
		Topic:       config["topic"],
		TopicSuffix: config["topicsuffix"],
		User:        config["user"],
		Pass:        config["pass"],
		Tls:         tls,
		Auth:        auth,
	}
}

func (segment *KafkaProducer) Run(wg *sync.WaitGroup) {
	var kafkaConn = kafka.Connector{}
	defer func() {
		close(segment.Out)
		kafkaConn.Close()
		wg.Done()
	}()

	if !segment.Tls {
		kafkaConn.DisableTLS()
		log.Println("[info] KafkaProducer: Disabled TLS, operating unencrypted.")
	}

	if !segment.Auth {
		kafkaConn.DisableAuth()
		log.Println("[info] KafkaProducer: Disabled auth.")
	} else {
		kafkaConn.SetAuth(segment.User, segment.Pass)
		log.Printf("[info] KafkaProducer: Authenticating as user '%s'.", segment.User)
	}

	kafkaConn.StartProducer(segment.Server)

	producerChannel := kafkaConn.ProducerChannel(segment.Topic + segment.TopicSuffix)

	for msg := range segment.In {
		segment.Out <- msg
		producerChannel <- msg
	}
}

func init() {
	segment := &KafkaProducer{}
	segments.RegisterSegment("kafkaproducer", segment)
}
