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
// docs/api/methods/19-method-driver.html

// ListDrivers lists the drivers available for the current API key.
func (c *Client) ListDrivers(
	ctx context.Context,
	request *maponv1.ListDriversRequest,
) (_ *maponv1.ListDriversResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list drivers: %w", err)
		}
	}()

	params := url.Values{}
	if request.GetId() != 0 {
		params.Add("id", strconv.FormatInt(request.GetId(), 10))
	}

	requestURL, err := url.Parse(c.baseURL + "/driver/list.json")
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

	var responseBody jsonDriverResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	drivers := make([]*maponv1.Driver, 0, len(responseBody.Data.Drivers))
	for _, d := range responseBody.Data.Drivers {
		drivers = append(drivers, mapJSONDriverToProto(d))
	}

	resp := &maponv1.ListDriversResponse{}
	resp.SetDrivers(drivers)
	return resp, nil
}

type jsonDriverResponse struct {
	Data struct {
		Drivers []jsonDriver `json:"drivers"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}

type jsonDriver struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	IButton string `json:"ibutton"`
	Tacho   string `json:"tacho"`
	Blocked bool   `json:"blocked"`
	Created string `json:"created"` // "2016-08-10 12:50:56"
}

func mapJSONDriverToProto(j jsonDriver) *maponv1.Driver {
	d := &maponv1.Driver{}
	d.SetDriverId(j.ID)
	d.SetName(j.Name)
	d.SetSurname(j.Surname)
	d.SetEmail(j.Email)
	d.SetPhone(j.Phone)
	d.SetIbuttonValue(j.IButton)
	d.SetTachographId(j.Tacho)
	d.SetBlocked(j.Blocked)

	// Time format "2006-01-02 15:04:05"
	if t, err := time.Parse("2006-01-02 15:04:05", j.Created); err == nil {
		d.SetCreatedAt(timestamppb.New(t))
	}

	return d
}
