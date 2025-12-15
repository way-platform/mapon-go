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
// docs/api/methods/35-method-tell_tale.html

// ListTellTaleValuesRequest is the request for [Client.ListTellTaleValues].
type ListTellTaleValuesRequest struct {
	UnitID int64
	From   time.Time
	To     time.Time
}

// ListTellTaleValuesResponse is the response for [Client.ListTellTaleValues].
type ListTellTaleValuesResponse struct {
	Data *maponv1.UnitTellTaleData
}

// ListTellTaleValues retrieves FMS tell tale values for a specified unit within a date range.
func (c *Client) ListTellTaleValues(ctx context.Context, request *ListTellTaleValuesRequest, opts ...ClientOption) (_ *ListTellTaleValuesResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list tell tale values: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	params.Add("unit_id", strconv.FormatInt(request.UnitID, 10))
	params.Add("from", request.From.UTC().Format(time.RFC3339))
	params.Add("till", request.To.UTC().Format(time.RFC3339))

	requestURL, err := url.Parse(c.baseURL + "/tell_tale/values.json")
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

	var responseBody jsonTellTaleResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	// The API returns data keyed by unit ID string.
	// We expect data for the requested UnitID.
	unitIDStr := strconv.FormatInt(request.UnitID, 10)
	var values []*maponv1.TellTaleValue

	if list, ok := responseBody.Data[unitIDStr]; ok {
		for _, v := range list {
			ttv := &maponv1.TellTaleValue{}
			ttv.SetName(v.Name)
			ttv.SetTelltaleId(int32(v.TelltaleID))
			ttv.SetValue(int32(v.Value))
			ttv.SetValueTitle(v.ValueTitle)

			if t, err := time.Parse(time.RFC3339, v.Datetime); err == nil {
				ttv.SetTime(timestamppb.New(t))
			}
			values = append(values, ttv)
		}
	}

	responseData := &maponv1.UnitTellTaleData{}
	responseData.SetUnitId(request.UnitID)
	responseData.SetValues(values)

	return &ListTellTaleValuesResponse{
		Data: responseData,
	}, nil
}

type jsonTellTaleResponse struct {
	Data  map[string][]jsonTellTaleValue `json:"data"`
	Error *jsonError                     `json:"error"`
}

type jsonTellTaleValue struct {
	Name       string `json:"name"`
	TelltaleID int    `json:"telltale_id"`
	Value      int    `json:"value"`
	ValueTitle string `json:"value_title"`
	Datetime   string `json:"datetime"`
}
