package main

import (
	"flag"
	"github.com/bwNetFlow/flowpipeline/pipeline"
	"github.com/hashicorp/logutils"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
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
	}
	pipeline := pipeline.NewFromConfig(config)
	pipeline.AutoDrain()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs

	pipeline.Close()
}
