// Collects and exports all flows to influxdb for long term storage.
// Tags to configure for Influxdb are from the protobuf definition.
// Supported Tags are:
// Cid,ProtoName,RemoteCountry,SamplerAddress,SrcIfDesc,DstIfDesc
// If no Tags are provided 'ProtoName' will be the only Tag used by default.
package influx

import (
	"log"
	"net/url"
	"reflect"
	"strings"
	"sync"

	"github.com/bwNetFlow/flowpipeline/pb"
	"github.com/bwNetFlow/flowpipeline/segments"
)

type Influx struct {
	segments.BaseSegment
	Address string   // optional, URL for influxdb endpoint, default is http://127.0.0.1:8086
	Org     string   // required, Influx org name
	Bucket  string   // required, Influx bucket
	Token   string   // required, Influx access token
	Tags    []string // optional, list of Tags to be created. Recommended to keep this to a

	fieldNames []string
}

func (segment Influx) New(config map[string]string) segments.Segment {
	newsegment := &Influx{}

	// TODO: add paramteres for Influx endpoint and eval vars
	if config["address"] != "" {
		// check if a valid url has been passed
		_, err := url.Parse(config["address"])
		if err != nil {
			log.Printf("[error] Influx: error parsing given url: %e", err)
		}
		newsegment.Address = config["address"]
	} else {
		newsegment.Address = "http://127.0.0.1:8086"
		log.Println("[info] Influx: Missing configuration parameter 'address'. Using default address 'http://127.0.0.1:8086'")
	}

	if config["org"] == "" {
		log.Println("[error] Influx: Missing configuration parameter 'org'. Please set the organization to use.")
		return nil
	} else {
		newsegment.Org = config["org"]
	}

	if config["bucket"] == "" {
		log.Println("[error] Influx: Missing configuration parameter 'bucket'. Please set the bucket to use.")
		return nil
	} else {
		newsegment.Bucket = config["bucket"]
	}

	if config["token"] == "" {
		log.Println("[error] Influx: Missing configuration parameter 'token'. Please set the token to use.")
		return nil
	} else {
		newsegment.Token = config["token"]
	}

	// set default Tags if not configured
	if config["tags"] == "" {
		log.Println("[info] prometheus: Configuration parameter 'tags' not set. Using default tags to export.")
		config["tags"] = "ProtoName"
	}
	for _, tag := range strings.Split(config["tags"], ",") {
		log.Printf("[info] prometheus: custom tag found: %s", tag)
		newsegment.Tags = append(newsegment.Tags, tag)
	}
	protofields := reflect.TypeOf(pb.EnrichedFlow{})
	for _, field := range newsegment.Tags {
		_, found := protofields.FieldByName(field)
		if !found {
			log.Printf("[error] Prometheus: Field '%s' specified in 'tags' does not exist.", field)
			return nil
		}
	}

	return newsegment
}

func (segment *Influx) Run(wg *sync.WaitGroup) {
	// TODO: extend options
	var connector = Connector{
		Address:   segment.Address,
		Bucket:    segment.Bucket,
		Org:       segment.Org,
		Token:     segment.Token,
		Batchsize: 5000,
		Tags:      segment.Tags,
	}

	// initialize Influx endpoint
	connector.Initialize()
	writeAPI := connector.influxClient.WriteAPI(connector.Org, connector.Bucket)
	defer func() {
		close(segment.Out)
		// Force all unwritten data to be sent
		writeAPI.Flush()
		connector.influxClient.Close()
		wg.Done()
	}()

	for msg := range segment.In {
		segment.Out <- msg
		datapoint := connector.CreatePoint(msg)
		if datapoint == nil {
			// just ignore raised warnings if flow cannot be converted or unmarshalled
			continue
		}
		// async write
		writeAPI.WritePoint(datapoint)
	}
}

func init() {
	segment := &Influx{}
	segments.RegisterSegment("influx", segment)
}
