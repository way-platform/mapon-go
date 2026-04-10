package mapon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// DataForwardEndpoint represents a configured data forwarding endpoint.
type DataForwardEndpoint struct {
	// ID is the Mapon endpoint ID.
	ID int64
	// URL is the webhook URL.
	URL string
	// Packs is the list of pack IDs configured for this endpoint.
	Packs []int32
}

// ListDataForwardsResponse is the response for [Client.ListDataForwards].
type ListDataForwardsResponse struct {
	Endpoints []DataForwardEndpoint
}

// ListDataForwards returns all registered push webhook endpoints for the API key.
func (c *Client) ListDataForwards(ctx context.Context, opts ...ClientOption) (_ *ListDataForwardsResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list data forwards: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	params.Add("key", cfg.apiKey)

	requestURL, err := url.Parse(c.baseURL + "/data_forward/list.json")
	if err != nil {
		return nil, fmt.Errorf("invalid request URL: %w", err)
	}
	requestURL.RawQuery = params.Encode()

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("User-Agent", getUserAgent())

	httpResponse, err := c.httpClient(cfg).Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return nil, newResponseError(httpResponse)
	}

	data, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	var responseBody jsonDataForwardListResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	var endpoints []DataForwardEndpoint
	for _, e := range responseBody.Data.Endpoints {
		endpoints = append(endpoints, DataForwardEndpoint{
			ID:    e.ID,
			URL:   e.URL,
			Packs: e.Packs,
		})
	}

	return &ListDataForwardsResponse{
		Endpoints: endpoints,
	}, nil
}

type jsonDataForwardListResponse struct {
	Data struct {
		Endpoints []struct {
			ID    int64   `json:"id"`
			URL   string  `json:"url"`
			Packs []int32 `json:"packs"`
		} `json:"endpoints"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
