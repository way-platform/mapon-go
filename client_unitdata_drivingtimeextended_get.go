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

// GetDrivingTimeExtended returns drivers information about driving time.
func (c *Client) GetDrivingTimeExtended(
	ctx context.Context,
	request *maponv1.GetDrivingTimeExtendedRequest,
) (_ *maponv1.GetDrivingTimeExtendedResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: get driving time extended: %w", err)
		}
	}()

	params := url.Values{}
	params.Add("unit_id", strconv.FormatInt(request.GetUnitId(), 10))

	requestURL, err := url.Parse(c.baseURL + "/unit_data/driving_time_extended.json")
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

	var responseBody jsonDrivingTimeResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	var driverMap map[string]jsonDriverInfo
	if err := json.Unmarshal(responseBody.Data, &driverMap); err != nil {
		return nil, fmt.Errorf("failed to parse drivers map: %w", err)
	}

	var drivers []*maponv1.DrivingTimeInfo
	for _, d := range driverMap {
		dti := &maponv1.DrivingTimeInfo{}
		dti.SetCurrentState(d.CurrentState)
		dti.SetDriverId(d.DriverID)
		dti.SetDriverName(d.DriverName)
		dti.SetDriverSurname(d.DriverSurname)

		if d.Now != nil {
			dti.SetNowDrivingS(d.Now.Driving)
			dti.SetNowDrivingRemainingS(d.Now.DrivingRemaining)
		}
		if d.Today != nil {
			dti.SetTodayDrivingS(d.Today.Driving)
			dti.SetTodayDrivingRemainingS(d.Today.DrivingRemaining)
		}
		if d.Week != nil {
			dti.SetWeekDrivingS(d.Week.Driving)
			dti.SetWeekDrivingRemainingS(d.Week.DrivingRemaining)
		}

		drivers = append(drivers, dti)
	}

	resp := &maponv1.GetDrivingTimeExtendedResponse{}
	resp.SetDrivers(drivers)
	return resp, nil
}

type jsonDrivingTimeResponse struct {
	Data  json.RawMessage `json:"data"`
	Error *jsonError      `json:"error"`
}

type jsonDriverInfo struct {
	CurrentState  string `json:"current_state"`
	DriverID      int64  `json:"driver_id"`
	DriverName    string `json:"driver_name"`
	DriverSurname string `json:"driver_surname"`
	Now           *struct {
		Driving          int64 `json:"driving"`
		DrivingRemaining int64 `json:"driving_remaining"`
	} `json:"now"`
	Today *struct {
		Driving          int64 `json:"driving"`
		DrivingRemaining int64 `json:"driving_remaining"`
	} `json:"today"`
	Week *struct {
		Driving          int64 `json:"driving"`
		DrivingRemaining int64 `json:"driving_remaining"`
	} `json:"week"`
}
