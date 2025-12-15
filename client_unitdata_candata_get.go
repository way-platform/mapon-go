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

type GetCanPointDataRequest struct {
	UnitID   int64
	Datetime time.Time
}

type GetCanPointDataResponse struct {
	Units []*maponv1.CanDataPoint
}

// GetCanDataPoint returns CAN data in specific datetime.
func (c *Client) GetCanDataPoint(ctx context.Context, request *GetCanPointDataRequest, opts ...ClientOption) (_ *GetCanPointDataResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: get can point data: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	params.Add("unit_id", strconv.FormatInt(request.UnitID, 10))
	params.Add("datetime", request.Datetime.UTC().Format(time.RFC3339))

	requestURL, err := url.Parse(c.baseURL + "/unit_data/can_point.json")
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

	var responseBody jsonCanPointResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	res := &GetCanPointDataResponse{}
	for _, u := range responseBody.Data.Units {
		cdp := &maponv1.CanDataPoint{}
		cdp.SetTime(timestamppb.New(request.Datetime))

		cdp.SetRpmAverage(int32(parseCanFloat(u.RpmAverage.Value)))
		cdp.SetRpmMax(int32(parseCanFloat(u.RpmMax.Value)))
		cdp.SetFuelLevelPercent(parseCanFloat(u.FuelLevel.Value))
		cdp.SetTotalDistanceKm(int64(parseCanFloat(u.TotalDistance.Value)))
		cdp.SetTotalFuelL(parseCanFloat(u.TotalFuel.Value))
		cdp.SetTotalEngineHours(parseCanFloat(u.TotalEngineHours.Value))
		cdp.SetAmbientTemperatureC(parseCanFloat(u.AmbientTemp.Value))

		var axes []*maponv1.CanDataPoint_AxisWeight
		for _, w := range u.WeightOnAxis {
			aw := &maponv1.CanDataPoint_AxisWeight{}
			aw.SetValueKg(w.Value)
			aw.SetAxisId(int32(w.Axis))
			aw.SetWheelId(int32(w.Wheel))
			axes = append(axes, aw)
		}
		cdp.SetAxisWeights(axes)

		res.Units = append(res.Units, cdp)
	}

	return res, nil
}

func parseCanFloat(v interface{}) float64 {
	f, _ := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	return f
}

type jsonCanPointResponse struct {
	Data struct {
		Units []struct {
			UnitID           int64               `json:"unit_id"`
			RpmAverage       jsonCanValue        `json:"rpm_average"`
			RpmMax           jsonCanValue        `json:"rpm_max"`
			FuelLevel        jsonCanValue        `json:"fuel_level"`
			TotalDistance    jsonCanValue        `json:"total_distance"`
			TotalFuel        jsonCanValue        `json:"total_fuel"`
			TotalEngineHours jsonCanValue        `json:"total_engine_hours"`
			AmbientTemp      jsonCanValue        `json:"ambient_temperature"`
			WeightOnAxis     []jsonCanAxisWeight `json:"weight_on_axis"`
		} `json:"units"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}

type jsonCanAxisWeight struct {
	Value float64 `json:"value"`
	Axis  int     `json:"axis"`
	Wheel int     `json:"wheel"`
}
