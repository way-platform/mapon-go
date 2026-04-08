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

// DeleteDataForwardRequest is the request for [Client.DeleteDataForward].
type DeleteDataForwardRequest struct {
	// EndpointID is the Mapon data forwarding endpoint ID to delete.
	EndpointID int64
}

// DeleteDataForward deregisters a push webhook endpoint from Mapon.
func (c *Client) DeleteDataForward(ctx context.Context, request *DeleteDataForwardRequest, opts ...ClientOption) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: delete data forward: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	params.Add("id", strconv.FormatInt(request.EndpointID, 10))

	requestURL, err := url.Parse(c.baseURL + "/data_forward/delete.json")
	if err != nil {
		return fmt.Errorf("invalid request URL: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), nil)
	if err != nil {
		return err
	}
	httpRequest.PostForm = params
	httpRequest.Header.Set("User-Agent", getUserAgent())

	httpResponse, err := c.httpClient(cfg).Do(httpRequest)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return newResponseError(httpResponse)
	}

	data, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return err
	}

	var responseBody jsonDataForwardDeleteResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return err
	}

	if responseBody.Error != nil {
		return fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	return nil
}

type jsonDataForwardDeleteResponse struct {
	Error *jsonError `json:"error"`
}
