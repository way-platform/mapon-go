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

type ListDigitalInputsExtendedRequest struct {
	UnitIDs  []int64
	InputIDs []int64
	From     time.Time
	To       time.Time
}

type ListDigitalInputsExtendedResponse struct {
	Units []*maponv1.UnitDigitalInputsExtended
}

// ListDigitalInputsExtended returns switches states details for selected period.
// Digital inputs are selected so that the digital inputs activity period is in selected period
// and digital inputs switched on time is no more than 15 days before selected period start.
func (c *Client) ListDigitalInputsExtended(ctx context.Context, request *ListDigitalInputsExtendedRequest, opts ...ClientOption) (_ *ListDigitalInputsExtendedResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list digital inputs extended: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	for _, id := range request.UnitIDs {
		params.Add("unit_id[]", strconv.FormatInt(id, 10))
	}
	for _, id := range request.InputIDs {
		params.Add("input_id[]", strconv.FormatInt(id, 10))
	}
	params.Add("from", request.From.UTC().Format(time.RFC3339))
	params.Add("till", request.To.UTC().Format(time.RFC3339))
	params.Add("include[]", "label")

	requestURL, err := url.Parse(c.baseURL + "/unit_data/digital_inputs_extended.json")
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

	var responseBody jsonDigitalInputsExtendedResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	res := &ListDigitalInputsExtendedResponse{}
	for _, u := range responseBody.Data.Units {
		udi := &maponv1.UnitDigitalInputsExtended{}
		udi.SetUnitId(u.UnitID)

		var extendedInputs []*maponv1.DigitalInputExtendedData
		for _, inp := range u.DigitalInputs {
			did := &maponv1.DigitalInputExtendedData{}
			did.SetInputId(inp.InputID)
			did.SetLabel(inp.Label)

			var events []*maponv1.DigitalInputExtendedEvent
			for _, st := range inp.States {
				evt := &maponv1.DigitalInputExtendedEvent{}
				if t, err := time.Parse("2006-01-02 15:04:05", st.GmtOn); err == nil {
					evt.SetOnTime(timestamppb.New(t))
				}
				if st.GmtOff != nil && *st.GmtOff != "" {
					if t, err := time.Parse("2006-01-02 15:04:05", *st.GmtOff); err == nil {
						evt.SetOffTime(timestamppb.New(t))
					}
				}

				locOn := &maponv1.Location{}
				locOn.SetLatitude(st.LatOn)
				locOn.SetLongitude(st.LngOn)
				locOn.SetAddress(st.PlaceOn)
				evt.SetOnLocation(locOn)

				locOff := &maponv1.Location{}
				locOff.SetLatitude(st.LatOff)
				locOff.SetLongitude(st.LngOff)
				locOff.SetAddress(st.PlaceOff)
				evt.SetOffLocation(locOff)

				evt.SetDistanceM(st.DistanceOn)
				evt.SetDriverId(st.DriverID)

				events = append(events, evt)
			}
			did.SetEvents(events)
			extendedInputs = append(extendedInputs, did)
		}
		udi.SetInputs(extendedInputs)
		res.Units = append(res.Units, udi)
	}

	return res, nil
}

type jsonDigitalInputsExtendedResponse struct {
	Data struct {
		Units []struct {
			UnitID        int64 `json:"unit_id"`
			DigitalInputs []struct {
				InputID int64  `json:"input_id"`
				Label   string `json:"label"`
				States  []struct {
					GmtOn      string  `json:"gmt_on"`
					LatOn      float64 `json:"lat_on"`
					LngOn      float64 `json:"lng_on"`
					PlaceOn    string  `json:"place_on"`
					GmtOff     *string `json:"gmt_off"`
					LatOff     float64 `json:"lat_off"`
					LngOff     float64 `json:"lng_off"`
					PlaceOff   string  `json:"place_off"`
					DistanceOn int64   `json:"distance_on"`
					DriverID   int64   `json:"driver_id"`
				} `json:"states"`
			} `json:"digital_inputs"`
		} `json:"units"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
