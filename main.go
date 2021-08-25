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

	_ "github.com/bwNetFlow/flowpipeline/segments/export/prometheusexporter"
	_ "github.com/bwNetFlow/flowpipeline/segments/filter"
	_ "github.com/bwNetFlow/flowpipeline/segments/io/goflow"
	_ "github.com/bwNetFlow/flowpipeline/segments/io/kafkaconsumer"
	_ "github.com/bwNetFlow/flowpipeline/segments/io/kafkaproducer"
	_ "github.com/bwNetFlow/flowpipeline/segments/io/stdin"
	_ "github.com/bwNetFlow/flowpipeline/segments/io/stdout"
	_ "github.com/bwNetFlow/flowpipeline/segments/modify/addcid"
	_ "github.com/bwNetFlow/flowpipeline/segments/modify/dropfields"
	_ "github.com/bwNetFlow/flowpipeline/segments/modify/geolocation"
	_ "github.com/bwNetFlow/flowpipeline/segments/modify/normalize"
	_ "github.com/bwNetFlow/flowpipeline/segments/modify/remoteaddress"
	_ "github.com/bwNetFlow/flowpipeline/segments/modify/snmp"
	_ "github.com/bwNetFlow/flowpipeline/segments/noop"
	_ "github.com/bwNetFlow/flowpipeline/segments/print/count"
	_ "github.com/bwNetFlow/flowpipeline/segments/print/printdots"
	_ "github.com/bwNetFlow/flowpipeline/segments/print/printflowdump"
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
