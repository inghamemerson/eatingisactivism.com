package locations

import (
	"encoding/json"
	"fmt"
	"os"
	"io"
	"time"

	"eatingisactivism/app/contentful"

	"github.com/joho/godotenv"
)

type SheetResponse struct {
	Range          string       `json:"range"`
	MajorDimension string       `json:"majorDimension"`
	Values         [][]string   `json:"values"`
}

type Location struct {
	ID string
	Name string
	Slug string
	Url string
	ShortDescription string
	LongDescription string
	Lat string
	Lng string
	Standard LocationStandard
	Tags []LocationTag
}

type LocationStandard struct {
	ID string
	Name string
	Slug string
}

type LocationTag struct {
	ID string
	Name string
	Slug string
}

type LocationMap map[string]Location
type LocationStandardMap map[string]LocationStandard
type LocationTagMap map[string]LocationTag

var (
	allLocations map[string]Location
	allStandards map[string]LocationStandard
	allTags map[string]LocationTag
	contentfulClient *contentful.Contentful
)

func init() {
	godotenv.Load(".env")

	contentfulApiKey := os.Getenv("CONTENTFUL_API_KEY")
	contentfulApiBaseUrl := os.Getenv("CONTENTFUL_API_BASE_URL")
	contentfulSpaceId := os.Getenv("CONTENTFUL_SPACE_ID")

	if (contentfulApiKey == "" || contentfulApiBaseUrl == "" || contentfulSpaceId == "") {
		fmt.Println("Error: missing Contentful API key, base URL, or space ID")
		return
	}

	allLocations = make(map[string]Location)
	allStandards = make(map[string]LocationStandard)
	allTags = make(map[string]LocationTag)

	contentfulClient = contentful.New(contentfulApiKey, contentfulSpaceId, "master", contentfulApiBaseUrl)

	buildData()

	go Poll()
}

// GetLocations returns all locations
func GetLocations() LocationMap {
	return allLocations
}

func GetStandards() LocationStandardMap {
	return allStandards
}

func GetTags() LocationTagMap {
	return allTags
}

// PollLocations polls the Google Sheets API for new locations every 5 seconds
func Poll() {
	for {
		buildData()

		time.Sleep(10 * time.Second)
	}
}

func buildData() {
	newStandards := ContentfulStandards()
	newTags := ContentfulTags()
	newLocations := ContentfulLocations()

	if len(newStandards) > 0 {
		AddStandards(newStandards)
	}

	if len(newTags) > 0 {
		AddTags(newTags)
	}

	if len(newLocations) > 0 {
		AddLocations(newLocations)
	}
}

func ContentfulLocations() []Location {
	locations := []Location{}

	entriesResponse, err := contentfulClient.GetEntries("location", 1000, 0)

	if err != nil {
		fmt.Println("Error: ", err)
		return locations
	}

	defer entriesResponse.Close()
	resBody, err := io.ReadAll(entriesResponse)

	if err != nil {
		fmt.Println("Error: ", err)
		return locations
	}

	var response contentful.ContentfulLocationResponse

	err = json.Unmarshal(resBody, &response)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	for _, location := range response.Items {
		standard := GetStandardByID(location.Standard.Sys.ID)
		tags := []LocationTag{}

		for _, tag := range location.Tags {
			tags = append(tags, GetTagByID(tag.Sys.ID))
		}

		locations = append(locations, Location{
			ID: location.ID,
			Name: location.Name,
			Slug: location.Slug,
			Url: location.Url,
			ShortDescription: location.ShortDescription,
			LongDescription: location.LongDescription,
			Lat: location.Lat,
			Lng: location.Lng,
			Standard: standard,
			Tags: tags,
		})
	}

	return locations
}

func ContentfulStandards() []LocationStandard {
	standards := []LocationStandard{}

	entriesResponse, err := contentfulClient.GetEntries("standard", 1000, 0)

	if err != nil {
		fmt.Println("Error: ", err)
		return standards
	}

	defer entriesResponse.Close()
	resBody, err := io.ReadAll(entriesResponse)

	if err != nil {
		fmt.Println("Error: ", err)
		return standards
	}

	var response contentful.ContentfulLocationStandardResponse

	err = json.Unmarshal(resBody, &response)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	for _, standard := range response.Items {
		standards = append(standards, LocationStandard{
			ID: standard.ID,
			Name: standard.Name,
			Slug: standard.Slug,
		})
	}

	return standards
}

func ContentfulTags() []LocationTag {
	tags := []LocationTag{}

	entriesResponse, err := contentfulClient.GetEntries("tags", 1000, 0)

	if err != nil {
		fmt.Println("Error: ", err)
		return tags
	}

	defer entriesResponse.Close()
	resBody, err := io.ReadAll(entriesResponse)

	if err != nil {
		fmt.Println("Error: ", err)
		return tags
	}

	var response contentful.ContentfulLocationTagResponse

	err = json.Unmarshal(resBody, &response)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	for _, tag := range response.Items {
		tags = append(tags, LocationTag{
			ID: tag.ID,
			Name: tag.Name,
			Slug: tag.Slug,
		})
	}

	return tags
}

func AddLocations(locations []Location) {
	for _, location := range locations {
		allLocations[location.Slug] = location
	}
}

func AddStandards(standards []LocationStandard) {
	for _, standard := range standards {
		allStandards[standard.Slug] = standard
	}
}

func AddTags(tags []LocationTag) {
	for _, tag := range tags {
		allTags[tag.Slug] = tag
	}
}

func GetLocationBySlug(slug string) Location {
	location, ok := allLocations[slug]

	if ok {
		return location
	}

	return Location{}
}

func GetStandardBySlug(slug string) LocationStandard {
	standard, ok := allStandards[slug]

	if ok {
		return standard
	}

	return LocationStandard{}
}

func GetTagBySlug(slug string) LocationTag {
	tag, ok := allTags[slug]

	if ok {
		return tag
	}

	return LocationTag{}
}

func GetTagByID(id string) LocationTag {
	for _, tag := range allTags {
		if tag.ID == id {
			return tag
		}
	}

	return LocationTag{}
}

func GetStandardByID(id string) LocationStandard {
	for _, standard := range allStandards {
		if standard.ID == id {
			return standard
		}
	}

	return LocationStandard{}
}

// filter function that return locations based on Standards, and Tags
func FilterLocations(standards []string, tags []string) LocationMap {
	locations := LocationMap{}

	for _, location := range allLocations {
		if len(standards) > 0 && !string_in_array(location.Standard.Slug, standards) {
			continue
		}

		// need to make an array of the location tag slugs
		tagSlugs := []string{}

		for _, tag := range location.Tags {
			tagSlugs = append(tagSlugs, tag.Slug)
		}

		if len(tags) > 0 && !array_contains(tagSlugs, tags) {
			continue
		}

		locations[location.Slug] = location
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
