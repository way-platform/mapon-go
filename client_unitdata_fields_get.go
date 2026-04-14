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
// docs/api/methods/09-method-unit_data.html

// GetUnitFields returns additional data about unit.
func (c *Client) GetUnitFields(
	ctx context.Context,
	request *maponv1.GetUnitFieldsRequest,
) (_ *maponv1.GetUnitFieldsResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: get unit fields: %w", err)
		}
	}()

	params := url.Values{}
	params.Add("unit_id", strconv.FormatInt(request.GetUnitId(), 10))

	requestURL, err := url.Parse(c.baseURL + "/unit_data/fields.json")
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

	var responseBody jsonUnitFieldsResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	var units []*maponv1.UnitFields
	for _, u := range responseBody.Data.Units {
		uf := &maponv1.UnitFields{}
		uf.SetUnitId(u.UnitID)

		var fields []*maponv1.UnitField
		for k, v := range u.Fields {
			f := &maponv1.UnitField{}
			f.SetKey(k)
			f.SetValue(fmt.Sprintf("%v", v))
			fields = append(fields, f)
		}
		uf.SetFields(fields)
		units = append(units, uf)
	}

	resp := &maponv1.GetUnitFieldsResponse{}
	resp.SetUnits(units)
	return resp, nil
}

type jsonUnitFieldsResponse struct {
	Data struct {
		Units []struct {
			UnitID int64                  `json:"unit_id"`
			Fields map[string]interface{} `json:"fields"`
		} `json:"units"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
