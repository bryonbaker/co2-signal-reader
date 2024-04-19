package reader

import (
	"fmt"
	"log"
	"time"

	"os-climate.org/carbon-intensity/pkg/data_source"
)

// TimeReader is am implementaiton of the IMarketReader. This implementation time-based reader of the market data.
// The TimeReader will request the market data from the IMarketDataSource object every "n" seconds where n is defined
// as a configurable item.
type OneShotReader struct {
	dataProvider data_source.IDataSource
	commsChannel chan string
	quitChannel  chan int
}

// Initialise is an implementaiton of the base class and is used to set up the working variables for the reader.
// Used to initialise the inter-process communication channels.
// This may not be required if the esign calls for all GetFxPricing to be used as a go routine.
// In which case the channel initialisers move into the base class.
func (r *OneShotReader) Initialise(c chan string, quit chan int) {
	r.commsChannel = c
	r.quitChannel = quit
}

// SetDataProvider assigns the MarketDataProvider so this implementation can request the data to be retrieved.
func (r *OneShotReader) SetDataProvider(ds data_source.IDataSource) {
	r.dataProvider = ds
}

// GetCarbonIntensity initiates the retrieval of the market data from the provider. It defined the go channel for providing the results,
// a separate channel for controlling shutdown, a list of currencies to retrieve the FX details for, the base Currency for the FX,
// and a date stamp to filter the FX data on.
func (r *OneShotReader) GetCarbonIntensity(countries []string) {
	log.Println("OneShotReader::GetCarbonIntensity()")

	r.GetCarbonIntensityFromProvider(countries)

	r.commsChannel <- "done"
}

func (r *OneShotReader) GetCarbonIntensityFromProvider(countries []string) {
	for _, country := range countries {
		resp := r.dataProvider.GetCarbonIntensity(country)

		// Iterate over the list of returned readings and send each to the channel for processing in the main thread..
		for _, v := range resp {
			priceData := v.Key + "," + v.ProviderResp // Comma-separated header

			select {
			case r.commsChannel <- priceData: // Send the pricing info to the main loop via the pricing channel.
				continue
			case <-r.quitChannel: // Check if a quit signal has been received. If so, tell the main loop that all thread-termination steps are done..
				fmt.Printf("Received QUIT signal.\n")
				r.commsChannel <- "done"
				return
			}
		}

		// The service has a rate limit of one request per second.
		time.Sleep(time.Second)
	}
}
