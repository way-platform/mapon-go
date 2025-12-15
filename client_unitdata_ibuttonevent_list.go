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

type ListIbuttonsRequest struct {
	UnitIDs []int64
	From    time.Time
	To      time.Time
}

type ListIbuttonsResponse struct {
	Units []*maponv1.UnitIbuttons
}

// ListIbuttons returns ibuttons for selected period.
func (c *Client) ListIbuttons(ctx context.Context, request *ListIbuttonsRequest, opts ...ClientOption) (_ *ListIbuttonsResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list ibuttons: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	for _, id := range request.UnitIDs {
		params.Add("unit_id[]", strconv.FormatInt(id, 10))
	}
	params.Add("from", request.From.UTC().Format(time.RFC3339))
	params.Add("till", request.To.UTC().Format(time.RFC3339))

	requestURL, err := url.Parse(c.baseURL + "/unit_data/ibuttons.json")
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

	var responseBody jsonIbuttonsResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	res := &ListIbuttonsResponse{}
	for _, u := range responseBody.Data.Units {
		ui := &maponv1.UnitIbuttons{}
		ui.SetUnitId(u.UnitID)

		var events []*maponv1.IbuttonEvent
		for _, ib := range u.Ibuttons {
			evt := &maponv1.IbuttonEvent{}
			evt.SetValue(ib.Value)
			if t, err := time.Parse("2006-01-02 15:04:05", ib.GMT); err == nil {
				evt.SetTime(timestamppb.New(t))
			}
			events = append(events, evt)
		}
		ui.SetIbuttons(events)
		res.Units = append(res.Units, ui)
	}

	return res, nil
}

type jsonIbuttonsResponse struct {
	Data struct {
		Units []struct {
			UnitID   int64 `json:"unit_id"`
			Ibuttons []struct {
				GMT   string `json:"gmt"`
				Value string `json:"value"`
			} `json:"ibuttons"`
		} `json:"units"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}
