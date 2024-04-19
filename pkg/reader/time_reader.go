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

package reader

import (
	"fmt"
	"time"

	"os-climate.org/carbon-intensity/pkg/data_source"
)

// TimeReader is am implementaiton of the IMarketReadethis. This implementation time-based reader of the market data.
// The TimeReader will request the market data from the IDataSource object every "n" seconds where n is defined
// as a configurable item.
type TimerReader struct {
	dataProvider data_source.IDataSource
	commsChannel chan string
	quitChannel  chan int
	timeDelay    int
}

// Initialise is an implementaiton of the base class and is used to set up the working variables for the readethis.
// Used to initialise the inter-process communication channels.
// This may not be required if the esign calls for all GetCarbonIntensity to be used as a go routine.
// In which case the channel initialisers move into the base class.
func (this *TimerReader) Initialise(c chan string, quit chan int) {
	this.commsChannel = c
	this.quitChannel = quit
	this.timeDelay = 120 // TODO: Replace this with a value read from the config file.
}

// SetDataProvider initialises the specific Market Provider that the market data will be rettirved from.
func (this *TimerReader) SetDataProvider(dp data_source.IDataSource) {
	this.dataProvider = dp
}

// GetCarbonIntensity implements the base class function. It uses a Go routine that retrieves pricing on
// scheduled intervals and puts the result on a channel for the main thread to pick up.
func (this *TimerReader) GetCarbonIntensity(countries []string) {
	fmt.Println("GetCarbonIntensity() request for TimerReader")

	if this.commsChannel == nil || this.quitChannel == nil {
		fmt.Println("ERROR: TimeReader::GetCarbonIntensity(): Channels not initialised.")
	} else if this.dataProvider == nil {
		fmt.Println("ERROR: TimeReader::GetCarbonIntensity(): DataProvider not initialised.")
	} else {
		this.GetCarbonIntensityFromProvider(countries) // Run it immediately before waiting for timethis.
		ticker := time.NewTicker(time.Duration(this.timeDelay) * time.Second)
		for _ = range ticker.C {
			// Timer has fired. Iterate through each currency and get the FX pricing.
			this.GetCarbonIntensityFromProvider(countries)
		}
		fmt.Printf("ERROR: GetCarbonIntensity() exiting the thread incorrectly")
	}
}

// getPricingFromMarketProvider calls the specific IMarketDataProvbider and processes the responses.
func (this *TimerReader) GetCarbonIntensityFromProvider(countries []string) {
	fmt.Println("GetCarbonIntensityFromProvider() request for TimerReader")
	// for _, country := range countries {
	// resp := this.dataProvider.GetCarbonIntensity(country)

	// // Iterate over the list of returned market prices and send each to the channel for processing in the main thread..
	// for _, v := range resp {
	// 	priceData := v.Fx_key + "," + v.Provider_resp // Comma-separated header

	// 	select {
	// 	case this.commsChannel <- priceData: // Send the pricing info to the main loop via the pricing channel.
	// 		continue
	// 	case <-this.quitChannel: // Check if a quit signal has been received. If so, tell the main loop that all thread-termination steps are done..
	// 		fmt.Printf("Received QUIT signal.\n")
	// 		this.commsChannel <- "done"
	// 		return
	// 	}
	// }

	// TODO: Update the lastGetTimestamp with now and persist it.
	// }
}
