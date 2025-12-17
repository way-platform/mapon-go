package mapon

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	maponv1 "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1"
)

// This API endpoint is documented in:
// docs/api/methods/08-method-unit.html

// ListUnitsRequest is the request for [Client.ListUnits].
type ListUnitsRequest struct {
	// UnitIDs is a list of unit IDs to filter by.
	UnitIDs []int64
}

// ListUnitsResponse is the response for [Client.ListUnits].
type ListUnitsResponse struct {
	// Units is the list of units returned by the API.
	Units []*maponv1.Unit
}

// ListUnits lists the units available for the current API key.
func (c *Client) ListUnits(ctx context.Context, request *ListUnitsRequest, opts ...ClientOption) (_ *ListUnitsResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list units: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	for _, id := range request.UnitIDs {
		params.Add("unit_id[]", strconv.FormatInt(id, 10))
	}
	// Always include all available data
	allIncludes := []string{
		"io_din",
		"fuel",
		"fuel_tank",
		"can",
		"reefer",
		"drivers",
		"temperature",
		"ambienttemp",
		"humidity",
		"device",
		"supply_voltage",
		"battery_voltage",
		"battery_level_percentage",
		"relays",
		"weights",
		"ignition",
		"tachograph",
		"altitude",
		"technical_details",
		"trailer_connections",
		"ev_values",
	}
	for _, inc := range allIncludes {
		params.Add("include[]", inc)
	}

	requestURL, err := url.Parse(c.baseURL + "/unit/list.json")
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

	units, err := ParseUnitsResponse(data)
	if err != nil {
		return nil, err
	}

	return &ListUnitsResponse{
		Units: units,
	}, nil
}
