package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PlaceSearchClient struct {
	apiKey  string
	baseURL string
	locale  string
	market  string
	client  *http.Client
}

type Place struct {
	EntityID          string `json:"entityId"`
	IataCode          string `json:"iataCode"`
	ParentID          string `json:"parentId,omitempty"`
	Name              string `json:"name"`
	CityName          string `json:"cityName"`
	CountryName       string `json:"countryName"`
	CountryID         string `json:"countryId,omitempty"`
	Type              string `json:"type"`
	Location          string `json:"location,omitempty"`
	Hierarchy         string `json:"hierarchy,omitempty"`
}

type autosuggestRequest struct {
	Query          autosuggestQuery `json:"query"`
	Limit          int              `json:"limit"`
	IsDestination  bool             `json:"isDestination"`
}

type autosuggestQuery struct {
	Market             string   `json:"market"`
	Locale             string   `json:"locale"`
	SearchTerm         string   `json:"searchTerm"`
	IncludedEntityTypes []string `json:"includedEntityTypes,omitempty"`
}

type autosuggestResponse struct {
	Places []Place `json:"places"`
}

func NewPlaceSearchClient(apiKey, baseURL string) *PlaceSearchClient {
	return &PlaceSearchClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		locale:  "pt-BR",
		market:  "BR",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *PlaceSearchClient) Search(ctx context.Context, query string) ([]Place, error) {
	body := autosuggestRequest{
		Query: autosuggestQuery{
			Market:     c.market,
			Locale:     c.locale,
			SearchTerm: query,
			IncludedEntityTypes: []string{
				"PLACE_TYPE_AIRPORT",
				"PLACE_TYPE_CITY",
			},
		},
		Limit:         10,
		IsDestination: false,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/autosuggest/flights", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("autosuggest API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result autosuggestResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Places, nil
}
