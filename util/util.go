package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Response struct {
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
	Badge []string
	Tag []string
}

func GetLocations() []Location {
	// fmt.Println("Getting locations")
	godotenv.Load(".env")

	locations := []Location{}

	url := os.Getenv("URL")
	key := os.Getenv("KEY")

	if (url == "" || key == "") {
		fmt.Println("URL or KEY not found in .env")
		return locations
	}

	resp, err := http.Get(url + "?key=" + key)

	if err != nil {
		fmt.Println("Error: ", err)
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

	var response Response

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
			Badge: strings.Split(location[7], ","),
			Tag: strings.Split(location[8], ","),
		})
	}

	return locations
}

// get location from locations by slug, or return an error
func GetLocationBySlug(locations []Location, slug string) Location {
	for _, location := range locations {
		if location.Slug == slug {
			return location
		}
	}

	return Location{}
}
