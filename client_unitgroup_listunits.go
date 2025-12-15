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

// This API endpoint is documented in:
// docs/api/methods/10-method-unit_groups.html

type ListUnitsInGroupRequest struct {
	GroupID int64
}

type ListUnitsInGroupResponse struct {
	UnitIDs []int64
}

// ListUnitsInGroup lists units in a group.
func (c *Client) ListUnitsInGroup(ctx context.Context, request *ListUnitsInGroupRequest, opts ...ClientOption) (_ *ListUnitsInGroupResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list units in group: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	params.Add("id", strconv.FormatInt(request.GroupID, 10))

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

	var responseBody jsonUnitGroupUnitsResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	res := &ListUnitsInGroupResponse{}
	for _, u := range responseBody.Data.Units {
		res.UnitIDs = append(res.UnitIDs, u.ID)
	}

	return res, nil
}

type jsonUnitGroupUnitsResponse struct {
	Data struct {
		Units []struct {
			ID int64 `json:"id"`
		} `json:"units"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
