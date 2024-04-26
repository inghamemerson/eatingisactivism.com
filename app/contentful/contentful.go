package contentful

import (
	"fmt"
	"io"
	"net/http"
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
	ID string `json:"sys.id"`
	Name string `json:"fields.name"`
	Slug string `json:"fields.slug"`
	Url string `json:"fields.url"`
	ShortDescription string `json:"fields.shortDescription"`
	LongDescription string `json:"fields.longDescription"`
	Lat string `json:"fields.coordinates.lat"`
	Lng string `json:"fields.coordinates.lon"`
	Standard ContentfulResponseLink `json:"fields.standard"`
	Tags []ContentfulResponseLink `json:"fields.tags"`
}

type ContentfulLocationStandard struct {
	ID string `json:"sys.id"`
	Name string `json:"fields.name"`
	Slug string `json:"fields.slug"`
}

type ContentfulLocationTag struct {
	ID string `json:"sys.id"`
	Name string `json:"fields.name"`
	Slug string `json:"fields.slug"`
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

func New(token string, spaceId string, environment string, baseURL string) *Contentful {
	return &Contentful{
		client: &http.Client{},
		token: token,
		BaseURL: baseURL,
		SpaceID: spaceId,
		Environment: environment,
	}
}

func (c *Contentful) getEntriesURL(contentType string) (string, error) {
	url := fmt.Sprintf("%s/spaces/%s/environments/%s/entries?access_token=%s&content_type=%s", c.BaseURL, c.SpaceID, c.Environment, c.token, contentType)

	return url, nil
}

func (c *Contentful) GetEntries(contentType string, limit int, offset int) (io.ReadCloser, error) {
	url, err := c.getEntriesURL(contentType)

	fmt.Println("URL: ", url)

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

	// lets debug and see what the response is
	// fmt.Println("Response: ", resp)

	return resp.Body, nil
}