// The flowpipeline package unifies all bwNetFlow functionality and aims to
// replace all dedicated platform components. By providing configurable
// pipelines to process flows all usual processess can be recreated in a simple
// and streamlined manner.
//
// The main entrypoint accepts command line flags to point to a configuration
// file, establishes the log level, and ensures a smooth exit on SIGINT. The
// actual configuration is done using the 'pipeline' package by referencing the
// different types in the 'segments' package.
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"

	"github.com/bwNetFlow/flowpipeline/pipeline"
	"github.com/hashicorp/logutils"

	_ "github.com/bwNetFlow/flowpipeline/segments/addcid"
	_ "github.com/bwNetFlow/flowpipeline/segments/count"
	_ "github.com/bwNetFlow/flowpipeline/segments/dropfields"
	_ "github.com/bwNetFlow/flowpipeline/segments/flowfilter"
	_ "github.com/bwNetFlow/flowpipeline/segments/geolocation"
	_ "github.com/bwNetFlow/flowpipeline/segments/goflow"
	_ "github.com/bwNetFlow/flowpipeline/segments/kafkaconsumer"
	_ "github.com/bwNetFlow/flowpipeline/segments/kafkaproducer"
	_ "github.com/bwNetFlow/flowpipeline/segments/noop"
	_ "github.com/bwNetFlow/flowpipeline/segments/normalize"
	_ "github.com/bwNetFlow/flowpipeline/segments/printdots"
	_ "github.com/bwNetFlow/flowpipeline/segments/printflowdump"
	_ "github.com/bwNetFlow/flowpipeline/segments/prometheusexporter"
	_ "github.com/bwNetFlow/flowpipeline/segments/remoteaddress"
	_ "github.com/bwNetFlow/flowpipeline/segments/snmp"
	_ "github.com/bwNetFlow/flowpipeline/segments/stdin"
	_ "github.com/bwNetFlow/flowpipeline/segments/stdout"
)

func main() {
	configfile := flag.String("c", "config.yml", "location of the config file in yml format")
	loglevel := flag.String("l", "warning", "loglevel: one of 'info', 'warning' or 'error'")
	flag.Parse()

	log.SetOutput(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"info", "warning", "error"},
		MinLevel: logutils.LogLevel(*loglevel),
		Writer:   os.Stderr,
	})

	config, err := ioutil.ReadFile(*configfile)
	if err != nil {
		log.Printf("[error] reading config file: %s", err)
		return
	}
	pipeline := pipeline.NewFromConfig(config)
	pipeline.AutoDrain()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs

	pipeline.Close()
}
