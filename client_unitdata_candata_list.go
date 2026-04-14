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

// ListCanPeriodData returns CAN data for a given period.
func (c *Client) ListCanPeriodData(
	ctx context.Context,
	request *maponv1.ListCanPeriodDataRequest,
) (_ *maponv1.ListCanPeriodDataResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list can period data: %w", err)
		}
	}()

	params := url.Values{}
	params.Add("unit_id", strconv.FormatInt(request.GetUnitId(), 10))
	params.Add("from", request.GetFromTime().AsTime().UTC().Format(time.RFC3339))
	params.Add("till", request.GetToTime().AsTime().UTC().Format(time.RFC3339))
	for _, inc := range request.GetInclude() {
		params.Add("include[]", inc)
	}

	requestURL, err := url.Parse(c.baseURL + "/unit_data/can_period.json")
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

	var responseBody jsonCanPeriodResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	var units []*maponv1.UnitCanPeriodData
	for _, u := range responseBody.Data.Units {
		ucpd := &maponv1.UnitCanPeriodData{}
		ucpd.SetUnitId(u.UnitID)
		ucpd.SetRpmAverage(mapCanMetricList(u.RpmAverage))
		ucpd.SetRpmMax(mapCanMetricList(u.RpmMax))
		ucpd.SetFuelLevelPercent(mapCanMetricList(u.FuelLevel))
		ucpd.SetServiceDistanceKm(mapCanMetricList(u.ServiceDistance))
		ucpd.SetTotalDistanceKm(mapCanMetricList(u.TotalDistance))
		ucpd.SetTotalFuelL(mapCanMetricList(u.TotalFuel))
		ucpd.SetTotalEngineHours(mapCanMetricList(u.TotalEngineHours))
		ucpd.SetAmbientTemperatureC(mapCanMetricList(u.AmbientTemp))
		ucpd.SetWeightOnChassisTotalKg(mapCanMetricList(u.WeightOnChassisTotal))

		if u.EvValues != nil {
			ucpd.SetEvBatteryRelPercent(mapCanMetricList(u.EvValues.CanEvBatteryRel))
			ucpd.SetEvBatteryAbsKwh(mapCanMetricList(u.EvValues.CanEvBatteryAbs))
			ucpd.SetEvCharging(mapCanMetricList(u.EvValues.EvCharging))
		}

		ucpd.SetWeightOnAxis(mapAxisWeightList(u.WeightOnAxis))

		units = append(units, ucpd)
	}

	resp := &maponv1.ListCanPeriodDataResponse{}
	resp.SetUnits(units)
	return resp, nil
}

func mapCanMetricList(in []jsonCanValue) []*maponv1.CanMetricValue {
	var out []*maponv1.CanMetricValue
	for _, v := range in {
		mv := &maponv1.CanMetricValue{}
		val, _ := strconv.ParseFloat(fmt.Sprintf("%v", v.Value), 64)
		mv.SetValue(val)
		if t, err := time.Parse("2006-01-02 15:04:05", v.GMT); err == nil {
			mv.SetTime(timestamppb.New(t))
		}
		out = append(out, mv)
	}
	return out
}

func mapAxisWeightList(in []jsonAxisWeight) []*maponv1.AxisWeightMetricValue {
	var out []*maponv1.AxisWeightMetricValue
	for _, v := range in {
		mv := &maponv1.AxisWeightMetricValue{}
		val, _ := strconv.ParseFloat(fmt.Sprintf("%v", v.Value), 64)
		mv.SetValueKg(val)
		mv.SetAxisId(int32(v.Axis))
		mv.SetWheelId(int32(v.Wheel))
		if t, err := time.Parse("2006-01-02 15:04:05", v.GMT); err == nil {
			mv.SetTime(timestamppb.New(t))
		}
		out = append(out, mv)
	}
	return out
}

type jsonCanValue struct {
	GMT   string      `json:"gmt"`
	Value interface{} `json:"value"` // Can be string or number
}

type jsonAxisWeight struct {
	GMT   string      `json:"gmt"`
	Value interface{} `json:"value"`
	Axis  int         `json:"axis"`
	Wheel int         `json:"wheel"`
}

type jsonCanPeriodResponse struct {
	Data struct {
		Units []struct {
			UnitID               int64            `json:"unit_id"`
			RpmAverage           []jsonCanValue   `json:"rpm_average"`
			RpmMax               []jsonCanValue   `json:"rpm_max"`
			FuelLevel            []jsonCanValue   `json:"fuel_level"`
			ServiceDistance      []jsonCanValue   `json:"service_distance"`
			TotalDistance        []jsonCanValue   `json:"total_distance"`
			TotalFuel            []jsonCanValue   `json:"total_fuel"`
			TotalEngineHours     []jsonCanValue   `json:"total_engine_hours"`
			AmbientTemp          []jsonCanValue   `json:"ambient_temperature"`
			WeightOnChassisTotal []jsonCanValue   `json:"weight_on_chassis_total"`
			WeightOnAxis         []jsonAxisWeight `json:"weight_on_axis"`
			EvValues             *struct {
				CanEvBatteryRel []jsonCanValue `json:"can_ev_battery_rel"`
				CanEvBatteryAbs []jsonCanValue `json:"can_ev_battery_abs"`
				EvCharging      []jsonCanValue `json:"ev_charging"`
			} `json:"ev_values"`
		} `json:"units"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
