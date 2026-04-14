package mapon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	maponv1 "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1"
)

// SaveDataForward registers or updates a push webhook endpoint with Mapon.
// Returns the Mapon-assigned endpoint ID for later deregistration.
func (c *Client) SaveDataForward(
	ctx context.Context,
	request *maponv1.SaveDataForwardRequest,
) (_ *maponv1.SaveDataForwardResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: save data forward: %w", err)
		}
	}()

	inner := map[string]any{
		"url": request.GetUrl(),
	}
	if request.GetId() != 0 {
		inner["id"] = request.GetId()
	}
	body := map[string]any{
		"key":   c.config.apiKey,
		"data":  inner,
		"packs": request.GetPacks(),
	}
	if len(request.GetUnitIds()) > 0 {
		body["unit_ids"] = request.GetUnitIds()
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	requestURL, err := url.Parse(c.baseURL + "/data_forward/save.json")
	if err != nil {
		return nil, fmt.Errorf("invalid request URL: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("User-Agent", getUserAgent())

	httpResponse, err := c.httpClient(c.config).Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer func() { _ = httpResponse.Body.Close() }()

	if httpResponse.StatusCode != http.StatusOK {
		return nil, newResponseError(httpResponse)
	}

	respBytes, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}

	var responseBody jsonDataForwardSaveResponse
	if err := json.Unmarshal(respBytes, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	resp := &maponv1.SaveDataForwardResponse{}
	resp.SetEndpointId(responseBody.Data.ID)
	return resp, nil
}

type jsonDataForwardSaveResponse struct {
	Data struct {
		ID int64 `json:"id"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
