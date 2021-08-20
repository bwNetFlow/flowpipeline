package segments

import (
	"log"
	"os"
	"strconv"
	"sync"

	kafka "github.com/bwNetFlow/kafkaconnector"
)

type KafkaProducer struct {
	BaseSegment
	Server string
	Topic  string
	Group  string
	User   string
	Pass   string
	NoTLS  bool
	NoAuth bool
}

func (segment KafkaProducer) New(config map[string]string) Segment {
	notls, _ := strconv.ParseBool(config["notls"])
	noauth, _ := strconv.ParseBool(config["noauth"])
	return &KafkaProducer{
		Server: config["server"],
		Topic:  config["topic"],
		Group:  config["group"],
		User:   config["user"],
		Pass:   config["pass"],
		NoTLS:  notls,
		NoAuth: noauth,
	}
}

func (segment *KafkaProducer) Run(wg *sync.WaitGroup) {
	if (segment.Server == "" || segment.Topic == "" || segment.Group == "") ||
		(segment.NoAuth == false && (segment.User == "" || segment.Pass == "")) {
		log.Println("[error] KafkaProducer: Missing required configuration parameters.")
		os.Exit(1)
	}

	var kafkaConn = kafka.Connector{}
	defer func() {
		close(segment.Out)
		kafkaConn.Close()
		wg.Done()
	}()

	if segment.NoTLS {
		kafkaConn.DisableTLS()
		log.Println("[info] KafkaProducer: Disabled TLS, operating unencrypted.")
	}

	if segment.NoAuth {
		kafkaConn.DisableAuth()
		log.Println("[info] KafkaProducer: Disabled auth.")
	} else {
		kafkaConn.SetAuth(segment.User, segment.Pass)
		log.Printf("[info] KafkaProducer: Authenticating as user '%s'.", segment.User)
	}

	kafkaConn.StartProducer(segment.Server)

	producerChannel := kafkaConn.ProducerChannel(segment.Topic)

	for msg := range segment.In {
		segment.Out <- msg
		producerChannel <- msg
	}
}

func init() {
	segment := &KafkaProducer{}
	RegisterSegment("kafkaproducer", segment)
}
