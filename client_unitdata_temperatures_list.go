package mapon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	maponv1 "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This API endpoint is documented in:
// docs/api/methods/09-method-unit_data.html

// ListTemperatures returns measured temperature points.
func (c *Client) ListTemperatures(
	ctx context.Context,
	request *maponv1.ListTemperaturesRequest,
) (_ *maponv1.ListTemperaturesResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list temperatures: %w", err)
		}
	}()

	params := url.Values{}
	for _, id := range request.GetUnitIds() {
		params.Add("unit_id[]", strconv.FormatInt(id, 10))
	}
	params.Add("from", request.GetFromTime().AsTime().UTC().Format(time.RFC3339))
	params.Add("till", request.GetToTime().AsTime().UTC().Format(time.RFC3339))

	requestURL, err := url.Parse(c.baseURL + "/unit_data/temperature.json")
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

	var responseBody jsonTemperatureResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	var units []*maponv1.UnitTemperatures
	for _, u := range responseBody.Data.Units {
		ut := &maponv1.UnitTemperatures{}
		ut.SetUnitId(u.UnitID)

		var sensors []*maponv1.UnitTemperatureSensor
		for _, s := range u.Sensors {
			uts := &maponv1.UnitTemperatureSensor{}
			uts.SetNumber(int32(s.No))

			var records []*maponv1.TemperatureRecord
			for _, t := range s.Temperatures {
				rec := &maponv1.TemperatureRecord{}
				rec.SetValueCelsius(t.Value)
				if tm, err := time.Parse("2006-01-02 15:04:05", t.GMT); err == nil {
					rec.SetTime(timestamppb.New(tm))
				}
				records = append(records, rec)
			}
			uts.SetTemperatures(records)
			sensors = append(sensors, uts)
		}
		ut.SetSensors(sensors)
		units = append(units, ut)
	}

	resp := &maponv1.ListTemperaturesResponse{}
	resp.SetUnits(units)
	return resp, nil
}

type jsonTemperatureResponse struct {
	Data struct {
		Units []struct {
			UnitID  int64 `json:"unit_id"`
			Sensors []struct {
				No           int `json:"no"`
				Temperatures []struct {
					GMT   string  `json:"gmt"`
					Value float64 `json:"value"`
				} `json:"temperatures"`
			} `json:"sensors"`
		} `json:"units"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
