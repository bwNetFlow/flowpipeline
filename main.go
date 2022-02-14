// The flowpipeline utility unifies all bwNetFlow functionality and
// provides configurable pipelines to process flows in any manner.
//
// The main entrypoint accepts command line flags to point to a configuration
// file and to establish the log level.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"plugin"

	"github.com/bwNetFlow/flowpipeline/pipeline"
	"github.com/hashicorp/logutils"

	_ "github.com/bwNetFlow/flowpipeline/segments/alert/http"

	_ "github.com/bwNetFlow/flowpipeline/segments/controlflow/branch"

	_ "github.com/bwNetFlow/flowpipeline/segments/export/influx"
	_ "github.com/bwNetFlow/flowpipeline/segments/export/prometheus"

	_ "github.com/bwNetFlow/flowpipeline/segments/filter/drop"
	_ "github.com/bwNetFlow/flowpipeline/segments/filter/elephant"
	_ "github.com/bwNetFlow/flowpipeline/segments/filter/flowfilter"

	_ "github.com/bwNetFlow/flowpipeline/segments/input/bpf"
	_ "github.com/bwNetFlow/flowpipeline/segments/input/goflow"
	_ "github.com/bwNetFlow/flowpipeline/segments/input/kafkaconsumer"
	_ "github.com/bwNetFlow/flowpipeline/segments/input/stdin"

	_ "github.com/bwNetFlow/flowpipeline/segments/modify/addcid"
	_ "github.com/bwNetFlow/flowpipeline/segments/modify/anonymize"
	_ "github.com/bwNetFlow/flowpipeline/segments/modify/dropfields"
	_ "github.com/bwNetFlow/flowpipeline/segments/modify/geolocation"
	_ "github.com/bwNetFlow/flowpipeline/segments/modify/normalize"
	_ "github.com/bwNetFlow/flowpipeline/segments/modify/protomap"
	_ "github.com/bwNetFlow/flowpipeline/segments/modify/remoteaddress"
	_ "github.com/bwNetFlow/flowpipeline/segments/modify/snmp"

	_ "github.com/bwNetFlow/flowpipeline/segments/pass"

	_ "github.com/bwNetFlow/flowpipeline/segments/output/csv"
	_ "github.com/bwNetFlow/flowpipeline/segments/output/json"
	_ "github.com/bwNetFlow/flowpipeline/segments/output/kafkaproducer"
	_ "github.com/bwNetFlow/flowpipeline/segments/output/sqlite"

	_ "github.com/bwNetFlow/flowpipeline/segments/print/count"
	_ "github.com/bwNetFlow/flowpipeline/segments/print/printdots"
	_ "github.com/bwNetFlow/flowpipeline/segments/print/printflowdump"
)

var Version string

func main() {
	configfile := flag.String("c", "config.yml", "location of the config file in yml format")
	pluginpath := flag.String("p", "", "path to a segment plugin")
	loglevel := flag.String("l", "warning", "loglevel: one of 'debug', 'info', 'warning' or 'error'")
	version := flag.Bool("v", false, "print version")
	flag.Parse()

	if *version {
		fmt.Println(Version)
		return
	}

	log.SetOutput(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"debug", "info", "warning", "error"},
		MinLevel: logutils.LogLevel(*loglevel),
		Writer:   os.Stderr,
	})

	if *pluginpath != "" {
		_, err := plugin.Open(*pluginpath)
		if err != nil {
			log.Printf("[error] Problem loading the specified plugin: %s", err)
			return
		}
	}

	config, err := ioutil.ReadFile(*configfile)
	if err != nil {
		log.Printf("[error] reading config file: %s", err)
		return
	}
	pipe := pipeline.NewFromConfig(config)
	pipe.Start()
	pipe.AutoDrain()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs

	pipe.Close()
}
