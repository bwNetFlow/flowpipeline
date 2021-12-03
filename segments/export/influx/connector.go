package influx

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	flow "github.com/bwNetFlow/protobuf/go"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// Connector provides export features to Influx
type Connector struct {
	Address    string
	Org        string
	Bucket     string
	Token      string
	ExportFreq int
	Batchsize  int

	tags         []string
	influxClient influxdb2.Client
}

// Initialize a connection to Influxdb
func (c *Connector) Initialize() {
	c.influxClient = influxdb2.NewClientWithOptions(
		c.Address,
		c.Token,
		influxdb2.DefaultOptions().SetBatchSize(uint(c.Batchsize)))

	c.checkBucket()
}

// check if database exists
func (c *Connector) checkBucket() {
	bucket, err := c.influxClient.BucketsAPI().FindBucketByName(context.Background(), c.Bucket)
	if err != nil {
		// The bucket should be created by the Influxdb admin.
		log.Printf("[warning] Influx: Given bucket %s not found.", c.Bucket)
	} else {
		log.Printf("[info] Influx: Bucket found with result: %s", bucket.Name)
	}
}

func (c *Connector) CreatePoint(flow *flow.FlowMessage) *write.Point {
	// write tags for datapoint and drop them to not insert as fields
	// TODO: maybe we will add more fields from the Protobuf Definition to be used as Tags
	tags := map[string]string{}
	for index, t := range c.tags {
		switch t {
		case "Cid":
			tags[t] = fmt.Sprint(flow.Cid)
		case "ProtoName":
			tags[t] = fmt.Sprint(flow.GetProtoName())
		case "RemoteCountry":
			tags[t] = flow.GetRemoteCountry()
		case "SamplerAddress":
			tags[t] = net.IP(flow.GetSamplerAddress()).String()
		case "SrcIfDesc":
			tags[t] = fmt.Sprint(flow.SrcIfDesc)
		case "DstIfDesc":
			tags[t] = fmt.Sprint(flow.DstIfDesc)
		default:
			log.Printf("[info] Influx: Chosen Tag not supported. ignoring tag %s", t)
			// delete not supported tags
			c.tags = append(c.tags[:index], c.tags[index+1:]...)
		}
	}

	// marshall protobuf to json
	data, err := json.Marshal(flow)
	if err != nil {
		log.Printf("[warning] influx: Skipping a flow, failed to recode protobuf as JSON: %v", err)
		return nil
	}

	// convert json []byte to insert in influx
	fields := make(map[string]interface{})
	err = json.Unmarshal([]byte(data), &fields)
	if err != nil {
		log.Printf("[warning] influx: Skipping a flow, failed to unmarshall JSON: %v", err)
		return nil
	}

	//remove used tags from fields
	for _, tag := range c.tags {
		delete(fields, tag)
	}

	// create point
	p := influxdb2.NewPoint(
		"flowdata",
		tags,
		fields,
		time.Now())
	return p
}
