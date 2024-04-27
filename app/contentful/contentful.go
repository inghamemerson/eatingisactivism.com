package contentful

import (
	"fmt"
	"io"
	"net/http"
	"encoding/json"
)

type Contentful struct {
	client *http.Client
	token string
	BaseURL string
	SpaceID string
	Environment string
}

type ContentfulResponseLink struct {
	Sys struct {
		Type string `json:"type"`
		LinkType string `json:"linkType"`
		ID string `json:"id"`
	} `json:"sys"`
}

type ContentfulLocation struct {
	Sys struct {
		ID string `json:"id"`
	} `json:"sys"`
	Fields struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
		Url string `json:"url"`
		ShortDescription string `json:"shortDescription"`
		LongDescription json.RawMessage `json:"longDescription"`
		Coordinates struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lon"`
		} `json:"coordinates"`
		Standard ContentfulResponseLink `json:"standard"`
		Tags []ContentfulResponseLink `json:"tags"`
	} `json:"fields"`
}

type ContentfulLocationStandard struct {
	Sys struct {
		ID string `json:"id"`
	} `json:"sys"`
	Fields struct {
		Title string `json:"title"`
		Slug string `json:"slug"`
		Icon string `json:"icon"`
	} `json:"fields"`
}

type ContentfulLocationTag struct {
	Sys struct {
		ID string `json:"id"`
	} `json:"sys"`
	Fields struct {
		Title string `json:"title"`
		Slug string `json:"slug"`
		Icon string `json:"icon"`
	} `json:"fields"`
}

type ContentfulLocationResponse struct {
	Items []ContentfulLocation `json:"items"`
	Message string `json:"message"`
}

type ContentfulLocationStandardResponse struct {
	Items []ContentfulLocationStandard `json:"items"`
	Message string `json:"message"`
}

type ContentfulLocationTagResponse struct {
	Items []ContentfulLocationTag `json:"items"`
	Message string `json:"message"`
}

type ContentfulWebhook struct {
	Sys struct {
		ID string `json:"id"`
		ContentType struct {
			Sys struct {
				ID string `json:"id"`
			}
		} `json:"contentType"`
	} `json:"sys"`
}

const (
	WebhookPublish string = "ContentManagement.Entry.publish"
	WebhookUnpublish string = "ContentManagement.Entry.unpublish"
	WebhookArchive string = "ContentManagement.Entry.archive"
	WebhookUnarchive string = "ContentManagement.Entry.unarchive"
	WebhookDelete string = "ContentManagement.Entry.delete"
)

func New(token string, spaceId string, environment string, baseURL string) *Contentful {
	return &Contentful{
		client: &http.Client{},
		token: token,
		BaseURL: baseURL,
		SpaceID: spaceId,
		Environment: environment,
	}
}

func (c *Contentful) getEntriesURL(contentType string, id string) (string, error) {
	url := fmt.Sprintf("%s/spaces/%s/environments/%s/entries", c.BaseURL, c.SpaceID, c.Environment)

	if id != "" {
		url = fmt.Sprintf("%s/%s", url, id)
	}

	url = fmt.Sprintf("%s?access_token=%s&content_type=%s", url, c.token, contentType)

	return url, nil
}

func (c *Contentful) GetEntries(contentType string, limit int, offset int, id string) ([]byte, error) {
	url, err := c.getEntriesURL(contentType, id)

	if err != nil {
		return nil, err
	}

	if limit > 0 {
		url = fmt.Sprintf("%s&limit=%d", url, limit)
	}

	if offset > 0 {
		url = fmt.Sprintf("%s&skip=%d", url, offset)
		url = fmt.Sprintf("%s&order=sys.createdAt", url)
	}

	resp, err := c.client.Get(url)

	if err != nil {
		return nil, err
	}

	// if the status code is 429, we are being rate limited
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("error: rate limited")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %d", resp.StatusCode)
	}

	if resp == nil {
		return nil, fmt.Errorf("error: no response")
	}

	defer resp.Body.Close()
	resBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return resBody, nil
}