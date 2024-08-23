// Copyright (c) 2024-2024 xplshn						[3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

const (
	defaultURL = "https://wttr.in"
)

var defaultClient = &http.Client{
	Timeout: 5 * time.Second,
}

type Options struct {
	City   string
	Params url.Values
}

type WeatherClient struct {
	httpClient *http.Client
	options    *Options
}

func NewWeatherClient(options *Options) *WeatherClient {
	if options.City == "" {
		fmt.Fprintln(os.Stderr, "error: the user did not specify a city")
		os.Exit(5)
	}
	if options.Params == nil {
		options.Params = url.Values{
			"q": {""},
			"0": {""},
		}
	}
	return &WeatherClient{
		httpClient: defaultClient,
		options:    options,
	}
}

func (wc *WeatherClient) GetWeather() error {
	queryURL := fmt.Sprintf("%s/%s", defaultURL, wc.options.City)
	request, err := http.NewRequest(http.MethodGet, queryURL, nil)
	if err != nil {
		return err
	}
	request.URL.RawQuery = wc.options.Params.Encode()

	response, err := wc.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("error: %s returned status %s", defaultURL, response.Status)
	}

	return displayWeather(response.Body)
}

func displayWeather(reader io.Reader) error {
	body, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	fmt.Print(string(body))
	return nil
}

func main() {
	location := flag.String("location", "", "Location for the weather report")
	onlyCurrent := flag.Bool("0", false, "Only current weather")     // Default is false
	superQuietVersion := flag.Bool("Q", false, "Superquiet version") // Default is false
	forceANSI := flag.Bool("na", false, "Request non-ANSI output format. (HTML)")
	currentAndToday := flag.Bool("1", false, "Current weather + today's forecast")
	currentTodayTomorrow := flag.Bool("2", false, "Current weather + today's + tomorrow's forecast")
	restrictGlyphs := flag.Bool("d", false, "Restrict output to standard console font glyphs")
	narrowVersion := flag.Bool("n", false, "Narrow version")
	quietVersion := flag.Bool("q", false, "Quiet version")
	noTermSeqs := flag.Bool("T", false, "Switch terminal sequences off")
	disableDefaults := flag.Bool("nd", false, "Disable default options")

	cmdInfo := ccmd.CmdInfo{
		Authors:     []string{"xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Name:        "wttr",
		Synopsis:    "<--location> <[-0|-na|-nd|-T|-q|-n|-d|-2|-1|-Q]>",
		Description: "Make a request to the online service wttr.in, which provides weather information",
		CustomFields: map[string]interface{}{
			"Notes": "The default flags are: `-0`, `-Q` when no other flags are set, and `-F` if `-na` is not set.",
		},
	}

	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error generating help page:", err)
		os.Exit(1)
	}

	flag.Usage = func() {
		fmt.Print(helpPage)
	}

	flag.Parse()

	params := url.Values{}
	// Apply defaults if no flags or only disableDefaults is set
	if !*disableDefaults && flag.NFlag() <= 1 {
		params.Add("0", "")
		params.Add("Q", "")
		params.Add("F", "")
	}

	// Add specific flags if set
	if *forceANSI {
		params.Add("na", "")
	} else {
		params.Add("A", "")
	}
	if *onlyCurrent {
		params.Add("0", "")
	}
	if *currentAndToday {
		params.Add("1", "")
	}
	if *currentTodayTomorrow {
		params.Add("2", "")
	}
	if *restrictGlyphs {
		params.Add("d", "")
	}
	if *narrowVersion {
		params.Add("n", "")
	}
	if *quietVersion {
		params.Add("q", "")
	}
	if *superQuietVersion {
		params.Add("Q", "")
	}
	if *noTermSeqs {
		params.Add("T", "")
	}

	clientOptions := &Options{
		City:   *location,
		Params: params,
	}
	client := NewWeatherClient(clientOptions)
	if err := client.GetWeather(); err != nil {
		log.Fatalln(err)
	}
}
