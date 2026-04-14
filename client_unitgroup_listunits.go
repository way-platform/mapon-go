package mapon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	maponv1 "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1"
)

// This API endpoint is documented in:
// docs/api/methods/10-method-unit_groups.html

// ListUnitsInGroup lists units in a group.
func (c *Client) ListUnitsInGroup(
	ctx context.Context,
	request *maponv1.ListUnitsInGroupRequest,
) (_ *maponv1.ListUnitsInGroupResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list units in group: %w", err)
		}
	}()

	params := url.Values{}
	params.Add("id", strconv.FormatInt(request.GetGroupId(), 10))

	requestURL, err := url.Parse(c.baseURL + "/unit_groups/list_units.json")
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

	var responseBody jsonUnitGroupUnitsResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	resp := &maponv1.ListUnitsInGroupResponse{}
	var unitIDs []int64
	for _, u := range responseBody.Data.Units {
		unitIDs = append(unitIDs, u.ID)
	}
	resp.SetUnitIds(unitIDs)

	return resp, nil
}

type jsonUnitGroupUnitsResponse struct {
	Data struct {
		Units []struct {
			ID int64 `json:"id"`
		} `json:"units"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
