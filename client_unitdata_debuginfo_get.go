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

type GetUnitDebugInfoRequest struct {
	UnitIDs []int64
}

type GetUnitDebugInfoResponse struct {
	Units []*maponv1.UnitDebugInfoData
}

// GetUnitDebugInfo returns various information about unit to help debug problems.
func (c *Client) GetUnitDebugInfo(ctx context.Context, request *GetUnitDebugInfoRequest, opts ...ClientOption) (_ *GetUnitDebugInfoResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: get unit debug info: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	for _, id := range request.UnitIDs {
		params.Add("unit_id[]", strconv.FormatInt(id, 10))
	}

	requestURL, err := url.Parse(c.baseURL + "/unit_data/debug_info.json")
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

	var responseBody jsonDebugInfoResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	res := &GetUnitDebugInfoResponse{}
	for _, u := range responseBody.Data.Units {
		udi := &maponv1.UnitDebugInfo{}

		if u.LastLocation != nil {
			loc := &maponv1.Location{}
			loc.SetLatitude(u.LastLocation.Lat)
			loc.SetLongitude(u.LastLocation.Lng)
			udi.SetLastLocation(loc)
			if t, err := time.Parse("2006-01-02 15:04:05", u.LastLocation.GMT); err == nil {
				udi.SetLastLocationTime(timestamppb.New(t))
			}
		}

		udi.SetFwVersion(u.FwVersion)

		if u.Tachograph != nil {
			if t, err := time.Parse("2006-01-02 15:04:05", u.Tachograph.LastTest); err == nil {
				udi.SetLastTachoTest(timestamppb.New(t))
			}
			if t, err := time.Parse("2006-01-02 15:04:05", u.Tachograph.LastID); err == nil {
				udi.SetLastTachoId(timestamppb.New(t))
			}
		}

		if u.Can != nil && u.Can.TotalDistance != nil {
			dist, _ := strconv.ParseFloat(u.Can.TotalDistance.Value, 64)
			udi.SetCanTotalDistanceKm(dist)
		}

		if u.GpsStatus != nil {
			udi.SetGpsStatus(u.GpsStatus.Value)
			if t, err := time.Parse("2006-01-02 15:04:05", u.GpsStatus.GMT); err == nil {
				udi.SetGpsStatusTime(timestamppb.New(t))
			}
		}

		data := &maponv1.UnitDebugInfoData{}
		data.SetUnitId(u.UnitID)
		data.SetDebugInfo(udi)
		res.Units = append(res.Units, data)
	}

	return res, nil
}

type jsonDebugInfoResponse struct {
	Data struct {
		Units []struct {
			UnitID       int64 `json:"unit_id"`
			LastLocation *struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
				GMT string  `json:"gmt"`
			} `json:"last_location"`
			FwVersion  string `json:"fw_version"`
			Tachograph *struct {
				LastTest string `json:"last_test"`
				LastID   string `json:"last_id"`
			} `json:"tachograph"`
			Can *struct {
				TotalDistance *struct {
					Value string `json:"value"`
				} `json:"total_distance"`
			} `json:"can"`
			GpsStatus *struct {
				Value string `json:"value"`
				GMT   string `json:"gmt"`
			} `json:"gpsstatus"`
		} `json:"units"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
