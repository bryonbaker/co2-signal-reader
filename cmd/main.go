// Copyright 2022 Bryon Baker

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"os-climate.org/carbon-intensity/pkg/data_publisher"
	"os-climate.org/carbon-intensity/pkg/data_source"
	"os-climate.org/carbon-intensity/pkg/reader"
	"os-climate.org/carbon-intensity/pkg/utils"

	"github.com/jessevdk/go-flags"
)

// App configuration details loaded from config file at boot.
var globalConfig struct {
	zones         []string
	dryRun        bool
	dataSource    string
	reader        string
	dataPublisher string
}

// Map that contains all of the possible publisher. A configuration determines which wil lbe instantiated.
var publisherMap = map[string]data_publisher.IDataPublisher{
	"console-publisher": &data_publisher.ConsolePublisher{},
	"kafka-publisher":   &data_publisher.KafkaPublisher{}}

// Map that contains all of the possible data sources. A configuration determines which wil lbe instantiated.
var readerMap = map[string]reader.IReader{
	//	"time-reader": &reader.TimerReader{},
	"one-shot": &reader.OneShotReader{}}

// Map that contains all of the possible data sources. A configuration determines which wil lbe instantiated.
var providerMap = map[string]data_source.IDataSource{
	// "simulator": &data_source.Simulator{},
	"co2-signal": &data_source.CO2SignalDataProvider{}}

func init() {
	log.Println("Initialising...")

	parseCommandLineArgs()

	// Load the configuration data from the configuration file
	// TODO: Add check to make sure the configuraiton item is valid
	config := utils.ReadConfig("./config/app-config.properties")
	globalConfig.dataSource = config["data-source"] // Which data source will the service use?
	globalConfig.reader = config["reader"]
	globalConfig.dataPublisher = config["data-publisher"] // Which publisher will the service use?

	if globalConfig.dryRun {
		// Override the configuration file if the command line switch is --dry-run
		log.Println("Running with --dry-run")
		globalConfig.dataPublisher = "console-publisher"
	}

	log.Printf("Loaded config: %v\n", globalConfig)
}

func main() {
	// Create a function thaty will be called on program exit so you cam close file handles etc.
	defer func() {
		cleanup()
	}()

	// Set up a channel for handling Ctrl-C, etc
	sigchan := make(chan os.Signal, 1)
	c := make(chan string) // Channel for passing pricing information
	quit := make(chan int) // Channel for sending quit signals.
	defer close(sigchan)
	defer close(c)
	defer close(quit)

	provider, exists := providerMap[globalConfig.dataSource]
	if !exists {
		optionList := ""
		for k := range providerMap {
			optionList += k + " "
		}
		var err error = fmt.Errorf("specified data source (%s) does not exist. Cannot instantiate the publisher. Options are: %s", globalConfig.dataSource, optionList)
		log.Fatal(err)
	}
	provider.Initialise()

	// Instantiate and initialise the Reader(s)
	// TODO: Add error handling
	reader, exists := readerMap[globalConfig.reader] // &reader.TimerReader{}
	if !exists {
		optionList := ""
		for k := range readerMap {
			optionList += k + " "
		}
		var err error = fmt.Errorf("specified reader (%s) does not exist. Cannot instantiate the  reader. Options are: %s", globalConfig.reader, optionList)
		log.Fatal(err)
	}

	reader.Initialise(c, quit)
	reader.SetDataProvider(provider)

	// Instantiate and initialise the Publisher fro the global configuration data
	publisher, exists := publisherMap[globalConfig.dataPublisher]
	if !exists {
		optionList := ""
		for k := range publisherMap {
			optionList += k + " "
		}
		var err error = fmt.Errorf("specified  publisher (%s) does not exist. Cannot instantiate the publisher. Options are: %s", globalConfig.dataPublisher, optionList)
		log.Fatal(err)
	}
	publisher.Initialise()

	// Start the reader thread
	globalConfig.zones = provider.GetAvailableZones()
	go reader.GetCarbonIntensity(globalConfig.zones)

	// Process messages
	run := true
loop:
	for run {
		select {
		case sig := <-sigchan:
			log.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			m := <-c         // Test the channel to see if the price getter has retrieved a quote
			if m == "done" { // Check if the reader is done.
				break loop
			} else if m != "" {
				SendToPublisher(publisher, m)
			}
		}
	}

	log.Printf("Exiting")
}

// Send the key/value to the instantiated Data Publisher
func SendToPublisher(publisher data_publisher.IDataPublisher, priceData string) {
	arr := strings.SplitN(priceData, ",", 2)

	// Check the data is formatted properly
	if len(arr) == 2 {
		publisher.PublishData(arr[0], arr[1])
	} else {
		log.Printf("ERROR: Badly formatted data in SendToPublisher. No comma separater: %s", priceData)
	}

}

// Called on program exit. Place any cleanup functions here
func cleanup() {

}

// isDryRun check is there is an os arg of "--dry-run". If there is then it returns tru. If not then it returns false.
func parseCommandLineArgs() {
	var opts struct {
		// Slice of bool will append 'true' each time the option
		// is encountered (can be set multiple times, like -vvv)
		DryRun bool `long:"dry-run" description:"Dry run - send output to console instead of the configured data publisher."`
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		log.Println("Invalid command-line options. Use --help for details.")
		os.Exit(1)
	}

	log.Printf("Dry run: %v\n", opts.DryRun)

	globalConfig.dryRun = opts.DryRun
}
