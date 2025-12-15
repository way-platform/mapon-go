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

type GetDrivingTimeExtendedRequest struct {
	UnitID int64
}

type GetDrivingTimeExtendedResponse struct {
	Drivers []*maponv1.DrivingTimeInfo
}

// GetDrivingTimeExtended returns drivers information about driving time.
func (c *Client) GetDrivingTimeExtended(ctx context.Context, request *GetDrivingTimeExtendedRequest, opts ...ClientOption) (_ *GetDrivingTimeExtendedResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: get driving time extended: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	params.Add("unit_id", strconv.FormatInt(request.UnitID, 10))

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

	var responseBody jsonDrivingTimeResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	res := &GetDrivingTimeExtendedResponse{}

	var driverMap map[string]jsonDriverInfo
	if err := json.Unmarshal(responseBody.Data, &driverMap); err != nil {
		return nil, fmt.Errorf("failed to parse drivers map: %w", err)
	}

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

		res.Drivers = append(res.Drivers, dti)
	}

	return res, nil
}

type jsonDrivingTimeResponse struct {
	Data  json.RawMessage `json:"data"`
	Error *jsonError      `json:"error"`
}

type jsonDriverInfo struct {
	CurrentState string `json:"current_state"`
	DriverID     int64  `json:"driver_id"`
	DriverName   string `json:"driver_name"`
	DriverSurname string `json:"driver_surname"`
	Now *struct {
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
