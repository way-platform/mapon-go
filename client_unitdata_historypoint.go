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

// GetHistoryPointData returns historical vehicle data at specific datetime.
func (c *Client) GetHistoryPointData(
	ctx context.Context,
	request *maponv1.GetHistoryPointDataRequest,
) (_ *maponv1.GetHistoryPointDataResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: get history point data: %w", err)
		}
	}()

	params := url.Values{}
	params.Add("unit_id", strconv.FormatInt(request.GetUnitId(), 10))
	params.Add("datetime", request.GetDatetime().AsTime().UTC().Format(time.RFC3339))
	for _, inc := range request.GetInclude() {
		params.Add("include[]", inc)
	}

	requestURL, err := url.Parse(c.baseURL + "/unit_data/history_point.json")
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

	var responseBody jsonHistoryPointResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	var units []*maponv1.UnitHistoryPoint
	for _, u := range responseBody.Data.Units {
		uhp := &maponv1.UnitHistoryPoint{}
		uhp.SetUnitId(u.UnitID)

		if u.CanTotalDistance != nil {
			val := &maponv1.CanMetricValue{}
			val.SetValue(parseCanFloat(u.CanTotalDistance.Value))
			if t, err := time.Parse("2006-01-02 15:04:05", u.CanTotalDistance.GMT); err == nil {
				val.SetTime(timestamppb.New(t))
			}
			uhp.SetCanTotalDistance(val)
		}

		if u.Mileage != nil {
			val := &maponv1.CanMetricValue{}
			val.SetValue(parseCanFloat(u.Mileage.Value))
			if t, err := time.Parse("2006-01-02 15:04:05", u.Mileage.GMT); err == nil {
				val.SetTime(timestamppb.New(t))
			}
			uhp.SetMileage(val)
		}

		if u.Position != nil {
			loc := &maponv1.Location{}
			loc.SetLatitude(u.Position.Value.Lat)
			loc.SetLongitude(u.Position.Value.Lng)
			uhp.SetPosition(loc)
			if t, err := time.Parse("2006-01-02 15:04:05", u.Position.GMT); err == nil {
				uhp.SetPositionTime(timestamppb.New(t))
			}
		}

		units = append(units, uhp)
	}

	resp := &maponv1.GetHistoryPointDataResponse{}
	resp.SetUnits(units)
	return resp, nil
}

type jsonHistoryPointResponse struct {
	Data struct {
		Units []struct {
			UnitID           int64         `json:"unit_id"`
			CanTotalDistance *jsonCanValue `json:"can_total_distance"`
			Mileage          *jsonCanValue `json:"mileage"`
			Position         *struct {
				GMT   string `json:"gmt"`
				Value struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"value"`
			} `json:"position"`
		} `json:"units"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
