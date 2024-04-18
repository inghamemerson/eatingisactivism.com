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
	Image bool
}

var (
	allLocations []Location
	sheetsUrl string
	sheetsKey string
)

type selectItem struct {
	Label string
	Value string
}

var LocationStandards []selectItem = []selectItem{
	{
		Label: "Gold",
		Value: "gold",
	},
	{
		Label: "Silver",
		Value: "silver",
	},
	{
		Label: "Bronze",
		Value: "bronze",
	},
}

var ValidStandards []string = map_values(LocationStandards, "Value")

var LocationBadges []selectItem = []selectItem{
	{
		Label: "Regenerative Organic Certified",
		Value: "roc",
	},
	{
		Label: "USDA Organic",
		Value: "usda_o",
	},
	{
		Label: "Certified Humane",
		Value: "hum",
	},
	{
		Label: "Patagonia Provisions",
		Value: "patagonia",
	},
}

var ValidBadges []string = map_values(LocationBadges, "Value")

var LocationTags []selectItem = []selectItem{
	{
		Label: "Beef",
		Value: "beef",
	},
	{
		Label: "Pork",
		Value: "pork",
	},
	{
		Label: "Fish",
		Value: "fish",
	},
	{
		Label: "Produce",
		Value: "produce",
	},
	{
		Label: "Poultry",
		Value: "poultry",
	},
	{
		Label: "Dairy",
		Value: "dairy",
	},
	{
		Label: "Grains",
		Value: "grains",
	},
	{
		Label: "Shellfish",
		Value: "shellfish",
	},
	{
		Label: "Honey",
		Value: "honey",
	},
	{
		Label: "Wine",
		Value: "wine",
	},
	{
		Label: "Beer",
		Value: "beer",
	},
}

var ValidTags []string = map_values(LocationTags, "Value")

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

		// split the badges by comma and clean out any whitespace
		badges := strings.Split(location[7], ",")
		locationBadges := []string{}
		for _, badge := range badges {
			badge = strings.ToLower(strings.TrimSpace(badge))

			if !string_in_array(badge, ValidBadges) {
				continue
			}

			locationBadges = append(locationBadges, badge)
		}

		tags := strings.Split(location[8], ",")
		locationTags := []string{}
		for _, tag := range tags {
			tag = strings.ToLower(strings.TrimSpace(tag))

			if !string_in_array(tag, ValidTags) {
				continue
			}

			locationTags = append(locationTags, tag)
		}

		locationStandard := ""
		standard := strings.ToLower(strings.TrimSpace(location[6]))
		if string_in_array(standard, ValidStandards) {
			locationStandard = standard
		}

		image := false
		if location[9] == "TRUE" {
			image = true
		}

		locations = append(locations, Location{
			Name: location[0],
			Slug: location[1],
			Url: location[2],
			Description: location[3],
			Lat: location[4],
			Lng: location[5],
			Standard: locationStandard,
			Badges: badges,
			Tags: tags,
			Image: image,
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

func map_values(arr []selectItem, key string) []string {
	values := []string{}

	for _, v := range arr {
		if key == "Label" {
			values = append(values, v.Label)
		} else if key == "Value" {
			values = append(values, v.Value)
		}
	}

	return values
}
