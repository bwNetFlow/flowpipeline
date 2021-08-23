package segments

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	kafka "github.com/bwNetFlow/kafkaconnector"
)

type KafkaConsumer struct {
	BaseSegment
	Server  string
	Topic   string
	Group   string
	User    string
	Pass    string
	UseTls  bool
	UseAuth bool
}

func (segment KafkaConsumer) New(config map[string]string) Segment {
	if config["server"] == "" || config["topic"] == "" || config["group"] == "" {
		log.Println("[error] KafkaConsumer: Missing required configuration parameters.")
		return nil
	}

	var useTls bool = true
	if config["tls"] != "" {
		if parsedUseTls, err := strconv.ParseBool(config["tls"]); err == nil {
			useTls = parsedUseTls
		} else {
			log.Println("[error] KafkaConsumer: Could not parse 'tls' parameter, using default true.")
		}
	} else {
		log.Println("[info] KafkaConsumer: 'tls' set to default true.")
	}

	var useAuth bool = true
	if config["auth"] != "" {
		if parsedUseAuth, err := strconv.ParseBool(config["auth"]); err == nil {
			useAuth = parsedUseAuth
		} else {
			log.Println("[error] KafkaConsumer: Could not parse 'auth' parameter, using default true.")
		}
	} else {
		log.Println("[info] KafkaConsumer: 'auth' set to default true.")
	}

	if useAuth && (config["user"] == "" || config["pass"] == "") {
		log.Println("[error] KafkaConsumer: Missing required configuration parameters for auth.")
		return nil
	}

	return &KafkaConsumer{
		Server:  config["server"],
		Topic:   config["topic"],
		Group:   config["group"],
		User:    config["user"],
		Pass:    config["pass"],
		UseTls:  useTls,
		UseAuth: useAuth,
	}
}

func (segment *KafkaConsumer) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.out)
		wg.Done()
	}()

	var kafkaConn = kafka.Connector{}
	if !segment.UseTls {
		kafkaConn.DisableTLS()
		log.Println("[info] KafkaConsumer: Disabled TLS, operating unencrypted.")
	}

	if !segment.UseAuth {
		kafkaConn.DisableAuth()
		log.Println("[info] KafkaConsumer: Disabled auth.")
	} else {
		kafkaConn.SetAuth(segment.User, segment.Pass)
		log.Printf("[info] KafkaConsumer: Authenticating as user '%s'.", segment.User)
	}

	err := kafkaConn.StartConsumer(segment.Server, strings.Split(segment.Topic, ","), segment.Group, -1)
	if err != nil {
		log.Println("[error] KafkaConsumer: Error starting consumer, this usually indicates a misconfiguration (auth).")
		os.Exit(1)
	}

	// receive flows in a loop
	for {
		select {
		case msg := <-kafkaConn.ConsumerChannel():
			segment.out <- msg
		case msg, ok := <-segment.in:
			if !ok {
				return
			}
			segment.out <- msg
		}
	}
}

func init() {
	segment := &KafkaConsumer{}
	RegisterSegment("kafkaconsumer", segment)
}
