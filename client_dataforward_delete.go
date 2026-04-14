package mapon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	maponv1 "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1"
)

// DeleteDataForward deregisters a push webhook endpoint from Mapon.
func (c *Client) DeleteDataForward(
	ctx context.Context,
	request *maponv1.DeleteDataForwardRequest,
) (_ *maponv1.DeleteDataForwardResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: delete data forward: %w", err)
		}
	}()

	params := url.Values{}
	params.Add("key", c.config.apiKey)
	params.Add("id", strconv.FormatInt(request.GetEndpointId(), 10))

	requestURL, err := url.Parse(c.baseURL + "/data_forward/delete.json")
	if err != nil {
		return nil, fmt.Errorf("invalid request URL: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		requestURL.String(),
		strings.NewReader(params.Encode()),
	)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

	var responseBody jsonDataForwardDeleteResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	return &maponv1.DeleteDataForwardResponse{}, nil
}

type jsonDataForwardDeleteResponse struct {
	Error *jsonError `json:"error"`
}
