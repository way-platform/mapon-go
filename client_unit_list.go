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

// ListUnits lists the units available for the current API key.
func (c *Client) ListUnits(
	ctx context.Context,
	request *maponv1.ListUnitsRequest,
) (_ *maponv1.ListUnitsResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list units: %w", err)
		}
	}()

	params := url.Values{}
	for _, id := range request.GetUnitIds() {
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

	units, err := ParseUnitsResponse(data)
	if err != nil {
		return nil, err
	}

	resp := &maponv1.ListUnitsResponse{}
	resp.SetUnits(units)
	return resp, nil
}
