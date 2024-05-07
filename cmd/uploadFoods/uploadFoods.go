package main

import (
	"net/http"
	"encoding/json"
	"os"
	"fmt"
	"time"
	"bytes"

	"eatingisactivism/app/seasons"

	"github.com/joho/godotenv"
	"github.com/charmbracelet/log"
	"golang.org/x/time/rate"
)

// This command is going to loop over all of the Foods and create and publish a new entry for each one.
// create url: https://api.contentful.com/spaces/dolzobiuxkfk/environments/master/entries
// Header: X-Contentful-Content-Type: food
// Header: Content-Type: application/vnd.contentful.management.v1+json
// Header: X-Contentful-Version: 1
// Example of fields:
// {
//    "fields": {
//      "name": {
//        "en-US": "Potato"
//      },
//      "slug": {
//        "en-US": "potato"
//      }
//    }
// }
// on success this returns a 201 status code and the new entry in the response body
// we want to take the sys.id from the response and publish the entry
// https://api.contentful.com/spaces/:spaceId/environments/master/entries/:id/published

var (
	contentfulCmaBaseUrl string = ""
	contentfulCmaToken string = ""
	contentfulSpaceId string = ""
	foodsMap map[string]string = make(map[string]string)
)

type Response struct {
	Sys struct {
		ID string `json:"id"`
	} `json:"sys"`
	Message string `json:"message"`
}

type NewEntry struct {
	Fields NewEntryFields `json:"fields"`
}

type NewEntryFields struct {
	Name NewEntryField `json:"name"`
	Slug NewEntryField `json:"slug"`
	Description NewEntryField `json:"description"`
}

type NewEntryField struct {
	EnUs string `json:"en-US"`
}

func main() {
	log.Info("Starting to upload foods to Contentful")
	godotenv.Load(".env")

	contentfulCmaBaseUrl = os.Getenv("CONTENTFUL_CMA_BASE_URL")
	contentfulCmaToken = os.Getenv("CONTENTFUL_CMA_TOKEN")
	contentfulSpaceId = os.Getenv("CONTENTFUL_SPACE_ID")

	if (contentfulCmaBaseUrl == "" || contentfulCmaToken == "" || contentfulSpaceId == "") {
		log.Fatal("Missing required environment variables")
	}

	foods := seasons.GetFoods()

	log.Info("Foods loaded")

	limiter := rate.NewLimiter(rate.Limit(time.Second * 5), 7)

	for _, food := range foods {
		if (!limiter.Allow()) {
			log.Warn("Rate limit hit, sleeping for a second")
			time.Sleep(time.Second)
		}

		createEntry(food)
	}
}

func createEntry(food seasons.Food) {
	url := fmt.Sprintf("%s/spaces/%s/environments/master/entries", contentfulCmaBaseUrl, contentfulSpaceId)
	client := &http.Client{}

	newEntry := &NewEntry{
		Fields: NewEntryFields{
			Name: NewEntryField{
				EnUs: food.Name,
			},
			Slug: NewEntryField{
				EnUs: food.Slug,
			},
			Description: NewEntryField{
				EnUs: food.Description,
			},
		},
	}

	marshalledEntry, err := json.Marshal(newEntry)

	payload := bytes.NewBuffer(marshalledEntry)

	if err != nil {
		log.Error("Error marshalling new entry", "err", err)
		return
	}

	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		log.Error("Error creating request", "err", err)
		return
	}

	req.Header.Add("X-Contentful-Content-Type", "foods")
	req.Header.Add("Content-Type", "application/vnd.contentful.management.v1+json")
	req.Header.Add("Authorization", "Bearer " + contentfulCmaToken)

	resp, err := client.Do(req)

	if err != nil {
		log.Error("Error creating entry", "err", err)
		return
	}

	defer resp.Body.Close()

	// parse the response
	var response Response
	json.NewDecoder(resp.Body).Decode(&response)

	// if the sys.id equals "", or "BadRequest" then we have an error
	if response.Sys.ID == "" || response.Sys.ID == "BadRequest" {
		log.Error("Error creating entry", "response", response)
		return
	}

	log.Info("Created entry with ID: " + response.Sys.ID)

	publishEntry(response.Sys.ID)
}

func publishEntry(id string) {
	url := fmt.Sprintf("https://api.contentful.com/spaces/dolzobiuxkfk/environments/master/entries/%s/published", id)
	client := &http.Client{}
	req, err := http.NewRequest("PUT", url, nil)

	if err != nil {
		log.Error("Error creating request", "err", err)
		return
	}

	req.Header.Add("X-Contentful-Version", fmt.Sprintf("%d", 1))
	req.Header.Add("Authorization", "Bearer " + contentfulCmaToken)

	resp, err := client.Do(req)

	if err != nil {
		log.Error("Error publishing entry", "err", err)
		return
	}

	defer resp.Body.Close()

	// if the status code is not 200, we have an error
	if resp.StatusCode != 200 {
		log.Error("Error publishing entry", "status", resp.StatusCode)
		return
	}

	log.Info("Published entry with ID: " + id)
}