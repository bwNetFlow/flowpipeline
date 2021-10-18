package influx

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	flow "github.com/bwNetFlow/protobuf/go"
	flow_helper "github.com/bwNetFlow/protobuf_helpers/go"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// Connector provides export features to Influx
type Connector struct {
	URL          string
	Org          string
	Bucket       string
	Token        string
	ExportFreq   int
	Batchsize    int
	influxClient influxdb2.Client
}

// Initialize a connection to Influxdb
func (c *Connector) Initialize() {
	c.influxClient = influxdb2.NewClientWithOptions(
		c.URL,
		c.Token,
		influxdb2.DefaultOptions().SetBatchSize(uint(c.Batchsize)))

	defer c.influxClient.Close() // TODO: needed? <--- das war blöd

	c.checkBucket()
}

// check if database exists
func (c *Connector) checkBucket() {
	bucket, err := c.influxClient.BucketsAPI().FindBucketByName(context.Background(), c.Bucket)
	if err != nil {
		// TODO: init bucket if not found? Maybe create one?
		log.Printf("[warning] influx: Given bucket %s not found", bucket.Name)
	} else {
		log.Printf("[info] influx: Bucket found with result: %s", bucket.Name)
	}
}

func (c *Connector) CreatePoint(flow *flow.FlowMessage) *write.Point {

	hflow := flow_helper.NewFlowHelper(flow)
	peer := hflow.Peer()
	// create point
	p := influxdb2.NewPoint(
		"flowdata",
		map[string]string{
			"id": "belwue",
		},
		map[string]interface{}{
			"router":    net.IP(flow.GetSamplerAddress()).String(),
			"ipversion": hflow.IPVersionString(),
			// "application":   application,
			"protoname": fmt.Sprint(flow.GetProtoName()),
			"direction": hflow.FlowDirectionString(),
			"peer":      peer,
			// "remoteas":      remoteAS,
			"remotecountry": flow.GetRemoteCountry(),
		}, time.Now())
	return p

}
