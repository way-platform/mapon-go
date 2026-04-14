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

// ListIgnitions returns ignition events for the specified units and period.
func (c *Client) ListIgnitions(
	ctx context.Context,
	request *maponv1.ListIgnitionsRequest,
) (_ *maponv1.ListIgnitionsResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list ignitions: %w", err)
		}
	}()

	params := url.Values{}
	for _, id := range request.GetUnitIds() {
		params.Add("unit_id[]", strconv.FormatInt(id, 10))
	}
	params.Add("from", request.GetFromTime().AsTime().UTC().Format(time.RFC3339))
	params.Add("till", request.GetToTime().AsTime().UTC().Format(time.RFC3339))

	requestURL, err := url.Parse(c.baseURL + "/unit_data/ignitions.json")
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

	var responseBody jsonIgnitionResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	var units []*maponv1.UnitIgnitions
	for _, u := range responseBody.Data.Units {
		ui := &maponv1.UnitIgnitions{}
		ui.SetUnitId(u.UnitID)

		var events []*maponv1.IgnitionEvent
		for _, evt := range u.Ignitions {
			protoEvt := &maponv1.IgnitionEvent{}
			if t, err := time.Parse("2006-01-02 15:04:05", evt.On); err == nil {
				protoEvt.SetOnTime(timestamppb.New(t))
			}
			if evt.Off != "" {
				if t, err := time.Parse("2006-01-02 15:04:05", evt.Off); err == nil {
					protoEvt.SetOffTime(timestamppb.New(t))
				}
			}
			events = append(events, protoEvt)
		}
		ui.SetIgnitions(events)
		units = append(units, ui)
	}

	resp := &maponv1.ListIgnitionsResponse{}
	resp.SetUnits(units)
	return resp, nil
}

type jsonIgnitionResponse struct {
	Data struct {
		Units []struct {
			UnitID    int64 `json:"unit_id"`
			Ignitions []struct {
				On  string `json:"on"`
				Off string `json:"off"`
			} `json:"ignitions"`
		} `json:"units"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
