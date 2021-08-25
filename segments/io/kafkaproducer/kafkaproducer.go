// Produces all received flows to Kafka instance.
package kafkaproducer

import (
	"log"
	"reflect"
	"strconv"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
	kafka "github.com/bwNetFlow/kafkaconnector"
	flow "github.com/bwNetFlow/protobuf/go"
)

type KafkaProducer struct {
	segments.BaseSegment
	Server      string
	Topic       string
	TopicSuffix string
	User        string
	Pass        string
	UseTls      bool
	UseAuth     bool
}

func (segment KafkaProducer) New(config map[string]string) segments.Segment {
	if config["server"] == "" || config["topic"] == "" {
		log.Println("[error] KafkaProducer: Missing required configuration parameters.")
		return nil
	}

	if config["topic-suffix"] != "" {
		fmsg := reflect.ValueOf(flow.FlowMessage{})
		field := fmsg.Elem().FieldByName(config["topic-suffix"])
		if !field.IsValid() {
			log.Println("[error] KafkaProducer: The 'topic-suffix' is not a valid FlowMessage field. Disabling this feature.")
		}
	} else {
		log.Println("[info] KafkaProducer: 'topic-suffix' set to default disabled.")
	}

	var useTls bool = true
	if config["tls"] != "" {
		if parsedUseTls, err := strconv.ParseBool(config["tls"]); err == nil {
			useTls = parsedUseTls
		} else {
			log.Println("[error] KafkaProducer: Could not parse 'tls' parameter, using default true.")
		}
	} else {
		log.Println("[info] KafkaProducer: 'tls' set to default true.")
	}

	var useAuth bool = true
	if config["auth"] != "" {
		if parsedUseAuth, err := strconv.ParseBool(config["auth"]); err == nil {
			useAuth = parsedUseAuth
		} else {
			log.Println("[error] KafkaProducer: Could not parse 'auth' parameter, using default true.")
		}
	} else {
		log.Println("[info] KafkaProducer: 'auth' set to default true.")
	}

	if useAuth && (config["user"] == "" || config["pass"] == "") {
		log.Println("[error] KafkaProducer: Missing required configuration parameters for auth.")
		return nil
	}

	return &KafkaProducer{
		Server:      config["server"],
		Topic:       config["topic"],
		TopicSuffix: config["topic-suffix"],
		User:        config["user"],
		Pass:        config["pass"],
		UseTls:      useTls,
		UseAuth:     useAuth,
	}
}

func (segment *KafkaProducer) Run(wg *sync.WaitGroup) {
	var kafkaConn = kafka.Connector{}
	defer func() {
		close(segment.Out)
		kafkaConn.Close()
		wg.Done()
	}()

	if !segment.UseTls {
		kafkaConn.DisableTLS()
		log.Println("[info] KafkaProducer: Disabled TLS, operating unencrypted.")
	}

	if !segment.UseAuth {
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
