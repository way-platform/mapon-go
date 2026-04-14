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
// docs/api/methods/16-method-object.html

// ListObjects lists the geofence objects.
func (c *Client) ListObjects(
	ctx context.Context,
	request *maponv1.ListObjectsRequest,
) (_ *maponv1.ListObjectsResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list objects: %w", err)
		}
	}()

	requestURL, err := url.Parse(c.baseURL + "/object/list.json")
	if err != nil {
		return nil, fmt.Errorf("invalid request URL: %w", err)
	}

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

	var responseBody jsonObjectResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	objects := make([]*maponv1.Object, 0, len(responseBody.Data.Objects))
	for _, o := range responseBody.Data.Objects {
		objects = append(objects, mapJSONObjectToProto(o))
	}

	resp := &maponv1.ListObjectsResponse{}
	resp.SetObjects(objects)
	return resp, nil
}

type jsonObjectResponse struct {
	Data struct {
		Objects []jsonObject `json:"objects"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}

type jsonObject struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	WKT     string `json:"wkt"`
	UserID  string `json:"user_id"`  // API returns string "1"
	GroupID string `json:"group_id"` // API returns string "0"
	Private string `json:"private"`  // "N" or "Y"
	Color   string `json:"color"`    // Hex like "FF0000"
}

func mapJSONObjectToProto(j jsonObject) *maponv1.Object {
	uid, _ := strconv.ParseInt(j.UserID, 10, 64)
	gid, _ := strconv.ParseInt(j.GroupID, 10, 64)

	o := &maponv1.Object{}
	o.SetObjectId(j.ID)
	o.SetName(j.Name)
	o.SetWkt(j.WKT)
	o.SetGroupId(gid)
	o.SetUserId(uid)
	o.SetIsPrivate(j.Private == "Y")
	o.SetColorHex(j.Color)

	return o
}
