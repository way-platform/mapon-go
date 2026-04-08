package mapon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// SaveDataForwardRequest is the request for [Client.SaveDataForward].
type SaveDataForwardRequest struct {
	// URL is the webhook endpoint URL to receive data packets.
	URL string
	// Packs is the list of pack IDs to forward (e.g., 1, 3, 5, 26, 55 for car packs).
	Packs []int32
	// UnitIDs is the list of unit IDs to forward (empty = all units for this API key).
	UnitIDs []int64
}

// SaveDataForwardResponse is the response for [Client.SaveDataForward].
type SaveDataForwardResponse struct {
	EndpointID int64
}

// SaveDataForward registers a push webhook endpoint with Mapon.
// Returns the Mapon-assigned endpoint ID for later deregistration.
func (c *Client) SaveDataForward(ctx context.Context, request *SaveDataForwardRequest, opts ...ClientOption) (_ int64, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: save data forward: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	params.Add("url", request.URL)

	for _, packID := range request.Packs {
		params.Add("packs[]", strconv.FormatInt(int64(packID), 10))
	}

	for _, unitID := range request.UnitIDs {
		params.Add("unit_ids[]", strconv.FormatInt(unitID, 10))
	}

	requestURL, err := url.Parse(c.baseURL + "/data_forward/insert_update.json")
	if err != nil {
		return 0, fmt.Errorf("invalid request URL: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), nil)
	if err != nil {
		return 0, err
	}
	httpRequest.PostForm = params
	httpRequest.Header.Set("User-Agent", getUserAgent())

	httpResponse, err := c.httpClient(cfg).Do(httpRequest)
	if err != nil {
		return 0, err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return 0, newResponseError(httpResponse)
	}

	data, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return 0, err
	}

	var responseBody jsonDataForwardSaveResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return 0, err
	}

	if responseBody.Error != nil {
		return 0, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	return responseBody.Data.ID, nil
}

type jsonDataForwardSaveResponse struct {
	Data struct {
		ID int64 `json:"id"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
