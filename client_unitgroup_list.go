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

// ListUnitGroups lists unit groups.
func (c *Client) ListUnitGroups(
	ctx context.Context,
	request *maponv1.ListUnitGroupsRequest,
) (_ *maponv1.ListUnitGroupsResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list unit groups: %w", err)
		}
	}()

	params := url.Values{}
	if request.GetUnitId() != 0 {
		params.Add("unit_id", strconv.FormatInt(request.GetUnitId(), 10))
	}

	requestURL, err := url.Parse(c.baseURL + "/unit_groups/list.json")
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

	var responseBody jsonUnitGroupsResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	var groups []*maponv1.UnitGroup
	for _, g := range responseBody.Data {
		grp := &maponv1.UnitGroup{}
		grp.SetGroupId(g.ID)
		grp.SetName(g.Name)

		if g.ParentID != nil {
			strVal := fmt.Sprintf("%v", g.ParentID)
			if strVal != "" {
				pid, _ := strconv.ParseInt(strVal, 10, 64)
				grp.SetParentId(pid)
			}
		}
		groups = append(groups, grp)
	}

	resp := &maponv1.ListUnitGroupsResponse{}
	resp.SetGroups(groups)
	return resp, nil
}

type jsonUnitGroupsResponse struct {
	Data []struct {
		ID       int64       `json:"id"`
		Name     string      `json:"name"`
		ParentID interface{} `json:"parent_id"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
