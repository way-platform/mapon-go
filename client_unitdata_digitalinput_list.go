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

type ListDigitalInputsRequest struct {
	UnitIDs []int64
	From    time.Time
	To      time.Time
}

type ListDigitalInputsResponse struct {
	Units []*maponv1.UnitDigitalInputs
}

// ListDigitalInputs returns switches states for selected period.
// Digital inputs are selected so that the digital inputs activity period is in selected period
// and digital inputs switched on time is no more than 15 days before selected period start.
func (c *Client) ListDigitalInputs(ctx context.Context, request *ListDigitalInputsRequest, opts ...ClientOption) (_ *ListDigitalInputsResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list digital inputs: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	for _, id := range request.UnitIDs {
		params.Add("unit_id[]", strconv.FormatInt(id, 10))
	}
	params.Add("from", request.From.UTC().Format(time.RFC3339))
	params.Add("till", request.To.UTC().Format(time.RFC3339))

	requestURL, err := url.Parse(c.baseURL + "/unit_data/digital_inputs.json")
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

	var responseBody jsonDigitalInputsResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	res := &ListDigitalInputsResponse{}
	for _, u := range responseBody.Data.Units {
		udi := &maponv1.UnitDigitalInputs{}
		udi.SetUnitId(u.UnitID)

		var inputDataList []*maponv1.DigitalInputData
		for _, inp := range u.DigitalInputs {
			did := &maponv1.DigitalInputData{}
			did.SetInputNumber(int32(inp.No))

			var events []*maponv1.DigitalInputEvent
			for _, st := range inp.States {
				evt := &maponv1.DigitalInputEvent{}
				if t, err := time.Parse("2006-01-02 15:04:05", st.On); err == nil {
					evt.SetOnTime(timestamppb.New(t))
				}
				if t, err := time.Parse("2006-01-02 15:04:05", st.Off); err == nil {
					evt.SetOffTime(timestamppb.New(t))
				}
				events = append(events, evt)
			}
			did.SetEvents(events)
			inputDataList = append(inputDataList, did)
		}
		udi.SetInputs(inputDataList)
		res.Units = append(res.Units, udi)
	}

	return res, nil
}

type jsonDigitalInputsResponse struct {
	Data struct {
		Units []struct {
			UnitID        int64 `json:"unit_id"`
			DigitalInputs []struct {
				No     int `json:"no"`
				States []struct {
					On  string `json:"on"`
					Off string `json:"off"`
				} `json:"states"`
			} `json:"digital_inputs"`
		} `json:"units"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
