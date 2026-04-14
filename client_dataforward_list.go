package mapon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	maponv1 "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1"
)

// ListDataForwards returns all registered push webhook endpoints for the API key.
func (c *Client) ListDataForwards(
	ctx context.Context,
	request *maponv1.ListDataForwardsRequest,
) (_ *maponv1.ListDataForwardsResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list data forwards: %w", err)
		}
	}()

	params := url.Values{}
	params.Add("key", c.config.apiKey)

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

	httpResponse, err := c.httpClient(c.config).Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer func() { _ = httpResponse.Body.Close() }()

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

	var endpoints []*maponv1.DataForwardEndpoint
	for _, e := range responseBody.Data.Endpoints {
		ep := &maponv1.DataForwardEndpoint{}
		ep.SetId(e.ID)
		ep.SetUrl(e.URL)
		ep.SetPacks(e.Packs)
		endpoints = append(endpoints, ep)
	}

	resp := &maponv1.ListDataForwardsResponse{}
	resp.SetEndpoints(endpoints)
	return resp, nil
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
