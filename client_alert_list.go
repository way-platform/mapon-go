package mapon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	maponv1 "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This API endpoint is documented in:
// docs/api/methods/01-method-alert.html

// ListAlerts returns triggered alerts.
func (c *Client) ListAlerts(
	ctx context.Context,
	request *maponv1.ListAlertsRequest,
) (_ *maponv1.ListAlertsResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list alerts: %w", err)
		}
	}()

	params := url.Values{}
	params.Add("from", request.GetFromTime().AsTime().UTC().Format(time.RFC3339))
	params.Add("till", request.GetToTime().AsTime().UTC().Format(time.RFC3339))

	for _, id := range request.GetUnitIds() {
		params.Add("unit_id[]", strconv.FormatInt(id, 10))
	}
	if request.GetDriver() != 0 {
		params.Add("driver", strconv.FormatInt(request.GetDriver(), 10))
	}
	// Always include details
	params.Add("include[]", "location")
	params.Add("include[]", "address")
	params.Add("include[]", "driver")

	requestURL, err := url.Parse(c.baseURL + "/alert/list.json")
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

	var responseBody jsonAlertResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	alerts := make([]*maponv1.Alert, 0, len(responseBody.Data))
	for _, a := range responseBody.Data {
		alerts = append(alerts, mapJSONAlertToProto(a))
	}

	resp := &maponv1.ListAlertsResponse{}
	resp.SetAlerts(alerts)
	return resp, nil
}

type jsonAlertResponse struct {
	Data  []jsonAlert `json:"data"`
	Error *jsonError  `json:"error"`
}

type jsonAlert struct {
	ID       int64  `json:"id"`
	UnitID   int64  `json:"unit_id"`
	DriverID int64  `json:"driver"`
	Type     string `json:"alert_type"`
	Value    string `json:"alert_val"`
	Msg      string `json:"msg"`
	Time     string `json:"time"`
	Location string `json:"location"` // "lat,lng"
	Address  string `json:"address"`
}

func mapJSONAlertToProto(j jsonAlert) *maponv1.Alert {
	a := &maponv1.Alert{}
	a.SetAlertId(j.ID)
	a.SetUnitId(j.UnitID)
	a.SetDriverId(j.DriverID)
	a.SetType(j.Type)
	a.SetMessage(j.Msg)
	a.SetValueRaw(j.Value)

	if t, err := time.Parse(time.RFC3339, j.Time); err == nil {
		a.SetTime(timestamppb.New(t))
	}

	// Location parse "lat,lng"
	if j.Location != "" {
		parts := strings.Split(j.Location, ",")
		if len(parts) == 2 {
			lat, _ := strconv.ParseFloat(parts[0], 64)
			lng, _ := strconv.ParseFloat(parts[1], 64)

			loc := &maponv1.Location{}
			loc.SetLatitude(lat)
			loc.SetLongitude(lng)
			loc.SetAddress(j.Address)

			a.SetLocation(loc)
		}
	}

	return a
}
