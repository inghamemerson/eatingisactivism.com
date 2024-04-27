package locations

import (
	"encoding/json"
	"fmt"
	"os"
	// "time"

	"eatingisactivism/app/contentful"

	"github.com/joho/godotenv"
)

type Location struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Url string `json:"url"`
	ShortDescription string `json:"shortDescription"`
	LongDescription json.RawMessage `json:"longDescription"`
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Standard LocationStandard `json:"standard"`
	Tags []LocationTag `json:"tags"`
}

type LocationStandard struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Icon string `json:"icon"`
}

type LocationTag struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Icon string `json:"icon"`
}

type LocationMap map[string]Location
type LocationStandardMap map[string]LocationStandard
type LocationTagMap map[string]LocationTag

var (
	allLocations LocationMap
	allStandards LocationStandardMap
	allTags LocationTagMap
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

	allLocations = make(LocationMap)
	allStandards = make(LocationStandardMap)
	allTags = make(LocationTagMap)

	contentfulClient = contentful.New(contentfulApiKey, contentfulSpaceId, "master", contentfulApiBaseUrl)

	buildData()

	// go Poll()
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
// func poll() {
// 	for {
// 		buildData()

// 		time.Sleep(10 * time.Second)
// 	}
// }

func buildData() {
	newStandards := ContentfulStandards()

	if len(newStandards) > 0 {
		AddStandards(newStandards)
	}

	newTags := ContentfulTags()

	if len(newTags) > 0 {
		AddTags(newTags)
	}

	newLocations := ContentfulLocations()


	if len(newLocations) > 0 {
		AddLocations(newLocations)
	}
}

func removeEntry(contentType string, id string) {
	switch contentType {
		case "location":
			location := GetLocationByID(id)
			delete(allLocations, location.Slug)
		case "standard":
			standard := GetStandardByID(id)
			delete(allStandards, standard.Slug)
		case "tags":
			tag := GetTagByID(id)
			delete(allTags, tag.Slug)
	}
}

func getEntry(contentType string, id string) {
	switch contentType {
		case "location":
			location := ContentfulLocation(id)
			allLocations[location.Slug] = location
		case "standard":
			standard := ContentfulStandard(id)
			allStandards[standard.Slug] = standard
		case "tags":
			tag := ContentfulTag(id)
			allTags[tag.Slug] = tag
	}
}

func ContentfulLocations() []Location {
	locations := []Location{}

	entriesResponse, err := contentfulClient.GetEntries("location", 1000, 0, "")

	if err != nil {
		fmt.Println("Error: ", err)
		return locations
	}

	var response contentful.ContentfulLocationResponse

	err = json.Unmarshal(entriesResponse, &response)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	for _, location := range response.Items {
		standard := GetStandardByID(location.Fields.Standard.Sys.ID)
		tags := []LocationTag{}

		for _, tag := range location.Fields.Tags {
			foundTag := GetTagByID(tag.Sys.ID)

			tags = append(tags, foundTag)
		}

		locations = append(locations, Location{
			ID: location.Sys.ID,
			Name: location.Fields.Name,
			Slug: location.Fields.Slug,
			Url: location.Fields.Url,
			ShortDescription: location.Fields.ShortDescription,
			LongDescription: location.Fields.LongDescription,
			Lat: location.Fields.Coordinates.Lat,
			Lng: location.Fields.Coordinates.Lng,
			Standard: standard,
			Tags: tags,
		})
	}

	return locations
}

func ContentfulLocation(id string) Location {
	location := Location{}

	entriesResponse, err := contentfulClient.GetEntries("location", 1000, 0, id)

	if err != nil {
		fmt.Println("Error: ", err)
		return location
	}

	var response contentful.ContentfulLocation

	err = json.Unmarshal(entriesResponse, &response)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	if (response.Sys.ID == "") {
		return location
	}

	standard := GetStandardByID(response.Fields.Standard.Sys.ID)
	tags := []LocationTag{}

	for _, tag := range response.Fields.Tags {
		tags = append(tags, GetTagByID(tag.Sys.ID))
	}

	location = Location{
		ID: response.Sys.ID,
		Name: response.Fields.Name,
		Slug: response.Fields.Slug,
		Url: response.Fields.Url,
		ShortDescription: response.Fields.ShortDescription,
		LongDescription: response.Fields.LongDescription,
		Lat: response.Fields.Coordinates.Lat,
		Lng: response.Fields.Coordinates.Lng,
		Standard: standard,
		Tags: tags,
	}

	return location
}

func ContentfulStandards() []LocationStandard {
	standards := []LocationStandard{}

	entriesResponse, err := contentfulClient.GetEntries("standard", 1000, 0, "")

	if err != nil {
		fmt.Println("Error: ", err)
		return standards
	}

	var response contentful.ContentfulLocationStandardResponse

	err = json.Unmarshal(entriesResponse, &response)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	for _, standard := range response.Items {
		standards = append(standards, LocationStandard{
			ID: standard.Sys.ID,
			Name: standard.Fields.Title,
			Slug: standard.Fields.Slug,
			Icon: standard.Fields.Icon,
		})
	}

	return standards
}

func ContentfulStandard(id string) LocationStandard {
	standard := LocationStandard{}

	entriesResponse, err := contentfulClient.GetEntries("standard", 1000, 0, id)

	if err != nil {
		fmt.Println("Error: ", err)
		return standard
	}

	var response contentful.ContentfulLocationStandard

	err = json.Unmarshal(entriesResponse, &response)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	if (response.Sys.ID == "") {
		return standard
	}

	standard = LocationStandard{
		ID: response.Sys.ID,
		Name: response.Fields.Title,
		Slug: response.Fields.Slug,
		Icon: response.Fields.Icon,
	}

	return standard
}

func ContentfulTags() []LocationTag {
	tags := []LocationTag{}

	entriesResponse, err := contentfulClient.GetEntries("tags", 1000, 0, "")

	if err != nil {
		fmt.Println("Error: ", err)
		return tags
	}

	var response contentful.ContentfulLocationTagResponse

	err = json.Unmarshal(entriesResponse, &response)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	for _, tag := range response.Items {
		tags = append(tags, LocationTag{
			ID: tag.Sys.ID,
			Name: tag.Fields.Title,
			Slug: tag.Fields.Slug,
			Icon: tag.Fields.Icon,
		})
	}

	return tags
}

func ContentfulTag(id string) LocationTag {
	tag := LocationTag{}

	entriesResponse, err := contentfulClient.GetEntries("tags", 1000, 0, id)

	if err != nil {
		fmt.Println("Error: ", err)
		return tag
	}

	var response contentful.ContentfulLocationTag

	err = json.Unmarshal(entriesResponse, &response)

	if err != nil {
		fmt.Println("Error: ", err)
	}

	if (response.Sys.ID == "") {
		return tag
	}

	tag = LocationTag{
		ID: response.Sys.ID,
		Name: response.Fields.Title,
		Slug: response.Fields.Slug,
		Icon: response.Fields.Icon,
	}

	return tag
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

func GetLocationByID(id string) Location {
	for _, location := range allLocations {
		if location.ID == id {
			return location
		}
	}

	return Location{}
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

func HandleWebhook(webhookType string, data []byte) {
	var webhook contentful.ContentfulWebhook

	err := json.Unmarshal(data, &webhook)

	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	entryID := webhook.Sys.ID

	switch webhookType {
	case contentful.WebhookPublish:
		getEntry(webhook.Sys.ContentType.Sys.ID, entryID)
	case contentful.WebhookUnarchive:
		getEntry(webhook.Sys.ContentType.Sys.ID, entryID)
	case contentful.WebhookUnpublish:
		removeEntry(webhook.Sys.ContentType.Sys.ID, entryID)
	case contentful.WebhookArchive:
		removeEntry(webhook.Sys.ContentType.Sys.ID, entryID)
	case contentful.WebhookDelete:
		removeEntry(webhook.Sys.ContentType.Sys.ID, entryID)
	}
}

// filter function that return locations based on Standards, and Tags
func FilterLocations(standards []string, tags []string) LocationMap {
	locations := LocationMap{}

	for _, location := range allLocations {
		if len(standards) > 0 && !string_in_array(location.Standard.Slug, standards) {
			continue
		}

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
