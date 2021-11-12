// Collects and exports all flows to influxdb for long term storage.
// Tags to configure for Influxdb are from the protobuf definition.
// Supported Tags are:
// Cid,ProtoName,RemoteCountry,SamplerAddress,SrcIfDesc,DstIfDesc
// If no Tags are provided 'ProtoName' will be the only Tag used by default.
package influx

import (
	"log"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/bwNetFlow/flowpipeline/segments"
)

type Influx struct {
	segments.BaseSegment
	Address string   // URL for influxdb endpoint
	Org     string   // influx org name
	Bucket  string   // influx bucket
	Token   string   // influx access token
	Tags    []string // optional, list of Tags to be created. Recommended to keep this to a
	// 	Batchsize  uint32 // set the batch size for the writer
	// 	ExportFreq uint32 // set the frequency for the writer
}

func (segment Influx) New(config map[string]string) segments.Segment {
	// TODO: add paramteres for Influx endpoint and eval vars
	var address = "http://127.0.0.1:8086"
	if config["address"] != "" {
		address = config["address"]
		// check if a valid url has been passed
		_, err := url.Parse(address)
		if err != nil {
			log.Printf("[error] Influx: error parsing given url: %e", err)
		}
		address = config["address"]
	} else {
		log.Println("[info] Influx: Missing configuration parameter 'address'. Using default address 'http://127.0.0.1:8086'")
	}

	var org string
	if config["org"] == "" {
		log.Println("[error] Influx: Missing configuration parameter 'org'. Please set the organization to use.")
	} else {
		org = config["org"]
	}

	var bucket string
	if config["bucket"] == "" {
		log.Println("[error] Influx: Missing configuration parameter 'bucket'. Please set the bucket to use.")
	} else {
		bucket = config["bucket"]
	}

	var token string
	if config["token"] == "" {
		log.Println("[error] Influx: Missing configuration parameter 'token'. Please set the token to use.")
	} else {
		token = config["token"]
	}

	// set default Tags if not configured
	var tags = []string{
		"ProtoName",
	}
	if config["tags"] == "" {
		log.Println("[info] Influx: Configuration parameter 'tags' not set. Using default tag 'ProtoName' to export")
	} else {
		tags = []string{}
		for _, tag := range strings.Split(config["tags"], ",") {
			log.Printf("[info] Influx: custom tag found: %s", tag)
			tags = append(tags, tag)
		}
	}
	// sort tags in alphabetical order
	sort.Strings(tags)

	return &Influx{
		Address: address,
		Org:     org,
		Bucket:  bucket,
		Token:   token,
		Tags:    tags,
	}
}

func (segment *Influx) Run(wg *sync.WaitGroup) {
	// TODO: extend options
	var connector = Connector{
		Address:   segment.Address,
		Bucket:    segment.Bucket,
		Org:       segment.Org,
		Token:     segment.Token,
		Batchsize: 5000,
		tags:      segment.Tags,
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
