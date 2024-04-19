package data_source

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/itchyny/gojq"
)

// The  structure that dummy market data should be returned in.
type co2SignalProviderResponse struct {
	Key                  string  `json:"key"`
	CountryCode          string  `json:"country_code"`
	Country              string  `json:"country_name"`
	Zone                 string  `json:"zone_name"`
	Status               string  `json:"status"`
	Datetime             string  `json:"datetime"`
	CarbonIntensity      float64 `json:"carbon_intensity"`
	FosselFuelPercentage float64 `json:"fossel_fuel_percentage"`
	UnitName             string  `json:"unit_name"`
	UnitValue            string  `json:"unit_value"`
}

const wsEntryPoint string = "https://api.co2signal.com"
const zonesURL string = "https://api.electricitymap.org/v3/zones"
const apiVersion string = "v1/latest"
const queryByCountryCode string = "countryCode="
const envVarName string = "CO2SIGNAL_API_KEY"
const getKeysJQuery string = "keys | .[]"

var dummyZoneList = [...]string{"US-AK", "US-CAL-BANC", "US-CAL-CISO", "US-CAL-IID", "US-CAL-LDWP", "US-CAR-CPLW", "US-CAR-DUK", "US-CAR-SC", "US-CAR-SCEG", "US-CAR-YAD", "US-CENT-SPA", "US-CENT-SWPP", "US-FLA-FMPP", "US-FLA-FPC", "US-FLA-FPL", "US-FLA-GVL", "US-FLA-SEC", "US-FLA-TAL", "US-FLA-TEC", "US-MIDA-PJM", "US-MIDW-AECI", "US-MIDW-GLHB", "US-MIDW-LGEE", "US-MIDW-MISO", "US-NE-ISNE", "US-NW-AVA", "US-NW-GCPD", "US-NW-GRID", "US-NW-GWA", "US-NW-IPCO"}

// const getLengthJQuery string = "length"

var authToken string

// CO2SignalDataProvider is an implementation of the DataProvider interface.
// It uses CO2 Signal as the data rovider for retireving carbon intensity of electricy generation.
type CO2SignalDataProvider struct {
}

// Initialise is used as a kibd of "constructor" to set up any internal properties.
// It should be called as soon as the CO2SignalDataProvider is instantiated.
func (r *CO2SignalDataProvider) Initialise() {
	val, ok := os.LookupEnv(envVarName)
	if !ok || val == "" {
		log.Fatalf("CO2SignalDataProvider::Initialise(). API-key environment variable (%s) not set.", envVarName)
	}
	authToken = val
}

func (r *CO2SignalDataProvider) GetAvailableZones() []string {

	var input map[string]interface{}
	r.getZones(&input)

	zoneList := jqZoneList(&input)
	log.Printf("All zones:\n%s", zoneList)

	// TODO: remove this line once a full API key is available
	zoneList = dummyZoneList[:]

	return zoneList
}

// GetZones retrieves the list of zones available form co2 signal and uses this to get the carbon intensity data.
func (r *CO2SignalDataProvider) getZones(input *map[string]interface{}) {
	log.Printf("CO2SignalDataProvider::getZones()")

	// TODO: Add a Context so it will time out.
	response, err := http.Get(zonesURL)
	if err != nil {
		log.Fatal(err)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	jsonResp := string(responseData)
	// log.Println(jsonResp)

	// Run all thew JQueries to extract the data
	json.Unmarshal([]byte(jsonResp), input)
}

// jqZoneList ueries a single path in a json message. It returns an interface because the caller
// understands the context and will need to cast it to the appropriate type.
// This can be used to search for a specific value at a path, or to return a subtree that
// can be parsed further. E.g. Return a float of an array.
func jqZoneList(input *map[string]interface{}) []string {
	var resp []string

	query, err := gojq.Parse(getKeysJQuery)
	if err != nil {
		log.Fatal(err)
	}

	iter := query.Run(*input) // or query.RunWithContext

	// While there are more items to fetch - fetch them.
	for value, more := iter.Next(); more; value, more = iter.Next() {
		// "more" and "value" are dual-purpose result codes. In the gojq code the bool result is the
		// inverse of "done". So result is false when done and true when not done. If true, the
		// Interface{} result can be cast to a value or an Error type. So if true you should check
		// if it was an error before checking for the result.
		if err, more := value.(error); more {
			log.Fatalln(err)
		} else if value == nil {
			log.Println("WARNING: JQuery returned no result: ", getKeysJQuery)
		} else {
			resp = append(resp, value.(string))
		}
	}

	return resp
}

// queryPath ueries a single path in a json message. It returns an interface because the caller
// understands the context and will need to cast it to the appropriate type.
// This can be used to search for a specific value at a path, or to return a subtree that
// can be parsed further. E.g. Return a float of an array.
func queryPath(input *map[string]interface{}, queryString string) interface{} {
	var resp interface{}

	query, err := gojq.Parse(queryString)
	if err != nil {
		log.Fatal(err)
	}

	iter := query.Run(*input) // or query.RunWithContext

	// While there are more items to fetch - fetch them.
	for value, more := iter.Next(); more; value, more = iter.Next() {
		// "more" and "value" are dual-purpose result codes. In the gojq code the bool result is the
		// inverse of "done". So result is false when done and true when not done. If true, the
		// Interface{} result can be cast to a value or an Error type. So if true you should check
		// if it was an error before checking for the result.
		if err, more := value.(error); more {
			log.Fatalln(err)
		} else if value == nil {
			log.Println("WARNING: JQuery returned no result: ", queryString)
		} else {
			resp = value
		}
	}

	return resp
}

// GetCarbonIntensity retrieves the carbon intensity of electricity for a given country code
// from co2signal.com.
func (r *CO2SignalDataProvider) GetCarbonIntensity(zone string) []DataSourceDetails {

	log.Printf("CO2SignalDataProvider::GetCarbonIntensity(%s)", zone)

	var resp []DataSourceDetails

	req := r.constructRequest(zone, "")
	jsonResp := r.requestData(req, authToken)

	log.Printf("CO2SignalDataProvider::GetCarbonIntensity(%s): %s", zone, jsonResp)

	var co2Result DataSourceDetails
	co2Result.Key, co2Result.ProviderResp = parseResponse(jsonResp)
	if co2Result.Key != "" {
		resp = append(resp, co2Result)
	}

	log.Printf("Parsed Response: %s : %s\n", co2Result.Key, co2Result.ProviderResp)

	return resp
}

// parseResponse extracts the FX details from the repsonse and stores it in a usable format
// that is base don the CSV formst you can download form ECB. Input params are:
// ecbJsonResp string: The json message returned from ECB
// Returns "","" if there was an error parsing the response.
func parseResponse(jsonResp string) (string, string) {
	var resp co2SignalProviderResponse

	// A list of all the JQueries that are used.
	queries := map[string]string{
		"country-code":           ".countryCode",
		"status":                 ".status",
		"datetime":               ".data.datetime",
		"carbon-intensity":       ".data.carbonIntensity",
		"fossel-fuel-percentage": ".data.fossilFuelPercentage",

		"unit-name":  ".units | keys[0]",
		"unit-value": ".units.carbonIntensity"}

	// Run all thew JQueries to extract the data
	var input map[string]interface{}
	json.Unmarshal([]byte(jsonResp), &input)

	var jsonVal interface{}

	jsonVal = queryPath(&input, queries["country-code"])
	if jsonVal != nil {
		resp.CountryCode = jsonVal.(string)
	} else {
		return "", ""
	}

	jsonVal = queryPath(&input, queries["status"])
	if jsonVal != nil {
		resp.Status = jsonVal.(string)
	} else {
		return "", ""
	}

	jsonVal = queryPath(&input, queries["datetime"])
	if jsonVal != nil {
		resp.Datetime = jsonVal.(string)
	} else {
		return "", ""
	}

	jsonVal = queryPath(&input, queries["carbon-intensity"])
	if jsonVal != nil {
		resp.CarbonIntensity = jsonVal.(float64)
	} else {
		return "", ""
	}

	jsonVal = queryPath(&input, queries["fossel-fuel-percentage"])
	if jsonVal != nil {
		resp.FosselFuelPercentage = jsonVal.(float64)
	} else {
		return "", ""
	}

	jsonVal = queryPath(&input, queries["unit-name"])
	if jsonVal != nil {
		resp.UnitName = jsonVal.(string)
	} else {
		return "", ""
	}

	jsonVal = queryPath(&input, queries["unit-value"])
	if jsonVal != nil {
		resp.UnitValue = jsonVal.(string)
	} else {
		return "", ""
	}

	// Construct the key
	resp.Key = resp.CountryCode

	// Format into a new JSON message
	convertedJsonMsg, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	return resp.Key, string(convertedJsonMsg)
}

// requestData sends the request to the data provider and returns the response as a string.
func (r *CO2SignalDataProvider) requestData(request string, token string) string {
	client := http.Client{}
	req, err := http.NewRequest("GET", request, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("auth-token", token)

	// TODO: Add a Context so it will time out.
	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(responseData)
}

// constructRequest formats the http request message for the market-data provider.
// https://api.co2signal.com/v1/latest?countryCode=FR
func (r *CO2SignalDataProvider) constructRequest(country string, authToken string) string {
	request := wsEntryPoint + "/" + apiVersion + "?" + queryByCountryCode + country

	return request
}
