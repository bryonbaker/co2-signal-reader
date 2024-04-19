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

import "os-climate.org/carbon-intensity/pkg/data_source"

type CoOrds struct {
}

// IReader defines an interface for reading carbon-intensity data from some data provider. The IReader
// is used to implement the method of triggering the read from the data source. For example, an IReader
// could be a scheduled read every 5 seconds, run once, a file, or trigger on an API POST to some HTTP endpoint.
// The IReader is intended to operate on a separate thread so it can read data asynchronously and post the result
// To a channel for processing in some other thread.
type IReader interface {
	// SetDataProvider assigns the DataProvider so this implementation can request the data to be retrieved.
	SetDataProvider(data_source.IDataSource)

	// Initialise configures all of the required runtime parameters and must be the first method called.
	Initialise(c chan string, quit chan int)

	// GetCarbonIntensity initiates the retrieval of the carbon-intensity data for the supplied list of countries
	// Data is published to the defined the go channel for processing in the main thread. There is a separate
	// channel for controlling shutdown.
	GetCarbonIntensity(countries []string)

	// GetCarbonIntensity initiates the retrieval of the carbon-intensity data for the supplied list of geo-corordinates
	// Data is published to the defined the go channel for processing in the main thread. There is a separate
	// channel for controlling shutdown.
	// GetCarbonIntensity(countries []CoOrds)
}
