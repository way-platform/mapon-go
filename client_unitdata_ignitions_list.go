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

// ListIgnitionsRequest is the request for [Client.ListIgnitions].
type ListIgnitionsRequest struct {
	UnitIDs []int64
	From    time.Time
	To      time.Time
}

// ListIgnitionsResponse is the response for [Client.ListIgnitions].
type ListIgnitionsResponse struct {
	Units []*maponv1.UnitIgnitions
}

// ListIgnitions returns ignition events for the specified units and period.
func (c *Client) ListIgnitions(ctx context.Context, request *ListIgnitionsRequest, opts ...ClientOption) (_ *ListIgnitionsResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list ignitions: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	for _, id := range request.UnitIDs {
		params.Add("unit_id[]", strconv.FormatInt(id, 10))
	}
	params.Add("from", request.From.UTC().Format(time.RFC3339))
	params.Add("till", request.To.UTC().Format(time.RFC3339))

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

	var responseBody jsonIgnitionResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	res := &ListIgnitionsResponse{}

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
		res.Units = append(res.Units, ui)
	}

	return res, nil
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
