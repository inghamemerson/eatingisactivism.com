package locations

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type SheetResponse struct {
	Range          string       `json:"range"`
	MajorDimension string       `json:"majorDimension"`
	Values         [][]string   `json:"values"`
}

type Location struct {
	Name string
	Slug string
	Url string
	Description string
	Lat string
	Lng string
	Standard string
	Badges []string
	Tags []string
}

var (
	allLocations []Location
	sheetsUrl string
	sheetsKey string
)

func init() {
	godotenv.Load(".env")

	url := os.Getenv("URL")
	key := os.Getenv("KEY")

	if (url == "" || key == "") {
		fmt.Println("URL or KEY not found in .env")
	}

	sheetsUrl = url
	sheetsKey = key

	allLocations = SheetLocations()

	go PollLocations()
}

// GetLocations returns all locations
func GetLocations() []Location {
	return allLocations
}

// PollLocations polls the Google Sheets API for new locations every 5 seconds
func PollLocations() {
	for {
		newLocations := SheetLocations()

		if len(newLocations) > 0 {
			allLocations = newLocations
		}

		time.Sleep(10 * time.Second)
	}
}

// SheetLocations fetches the locations from the Google Sheets API
func SheetLocations() []Location {
	locations := []Location{}

	resp, err := http.Get(sheetsUrl + "?key=" + sheetsKey)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	// if we do not have a response, return empty locations
	if resp == nil {
		fmt.Println("Error: no response")
		return locations
	}

	// if status code is not 200, return empty locations
	if resp.StatusCode != 200 {
		fmt.Println("Error: ", resp.Status)
		return locations
	}

	defer resp.Body.Close()
	resBody, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	var response SheetResponse

	err = json.Unmarshal(resBody, &response)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	// the response body "values" is a JSON array of Locations that we need to parse
	for _, location := range response.Values {
		// ensure the location has all the necessary fields
		if len(location) < 9 {
			continue
		}

		locations = append(locations, Location{
			Name: location[0],
			Slug: location[1],
			Url: location[2],
			Description: location[3],
			Lat: location[4],
			Lng: location[5],
			Standard: location[6],
			Badges: strings.Split(location[7], ","),
			Tags: strings.Split(location[8], ","),
		})
	}

	return locations
}

func GetLocationBySlug(slug string) Location {
	for _, location := range allLocations {
		if location.Slug == slug {
			return location
		}
	}

	return Location{}
}

// filter function that return locations based on Standards, Badges, and Tags
func FilterLocations(standards []string, badges []string, tags []string) []Location {
	locations := []Location{}

	for _, location := range allLocations {
		if len(standards) > 0 && !string_in_array(location.Standard, standards) {
			continue
		}

		if len(badges) > 0 && !array_contains(location.Badges, badges) {
			continue
		}

		if len(tags) > 0 && !array_contains(location.Tags, tags) {
			continue
		}

		locations = append(locations, location)
	}

	return locations
}

func string_in_array(s string, arr []string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}

	return false
}

func array_contains(arr []string, arr2 []string) bool {
	for _, a := range arr {
		for _, b := range arr2 {
			if a == b {
				return true
			}
		}
	}

	return false
}


