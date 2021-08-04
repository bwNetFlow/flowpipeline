package segments

import (
	"github.com/bwNetFlow/kafkaconnector"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type KafkaConsumer struct {
	BaseSegment
	Server string
	Topic  string
	Group  string
	User   string
	Pass   string
	NoTLS  bool
	NoAuth bool
}

func (segment KafkaConsumer) New(config map[string]string) Segment {
	notls, _ := strconv.ParseBool(config["notls"])
	noauth, _ := strconv.ParseBool(config["noauth"])
	return &KafkaConsumer{
		Server: config["server"],
		Topic:  config["topic"],
		Group:  config["group"],
		User:   config["user"],
		Pass:   config["pass"],
		NoTLS:  notls,
		NoAuth: noauth,
	}
}

func (segment *KafkaConsumer) Run(wg *sync.WaitGroup) {
	defer func() {
		close(segment.out)
		wg.Done()
	}()

	if (segment.Server == "" || segment.Topic == "" || segment.Group == "") ||
		(segment.NoAuth == false && (segment.User == "" || segment.Pass == "")) {
		log.Println("[error] KafkaConsumer: Missing required configuration parameters.")
		os.Exit(1)
	}

	var kafkaConn = kafka.Connector{}
	if segment.NoTLS {
		kafkaConn.DisableTLS()
		log.Println("[info] KafkaConsumer: Disabled TLS, operating unencrypted.")
	}

	if segment.NoAuth {
		kafkaConn.DisableAuth()
		log.Println("[info] KafkaConsumer: Disabled auth.")
	} else {
		kafkaConn.SetAuth(segment.User, segment.Pass)
		log.Printf("[info] KafkaConsumer: Authenticating as user '%s'.", segment.User)
	}

	kafkaConn.StartConsumer(segment.Server, strings.Split(segment.Topic, ","), segment.Group, -1)

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
