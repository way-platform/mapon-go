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

// ListRoutesRequest is the request for [Client.ListRoutes].
type ListRoutesRequest struct {
	From    time.Time
	Till    time.Time
	UnitIDs []int64
	// Include additional data. E.g., "polyline".
	Include []string
}

// ListRoutesResponse is the response for [Client.ListRoutes].
type ListRoutesResponse struct {
	Routes []*maponv1.Route
}

// ListRoutes lists the routes for units in the specified period.
func (c *Client) ListRoutes(ctx context.Context, request *ListRoutesRequest, opts ...ClientOption) (_ *ListRoutesResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list routes: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	// API expects Y-m-dTH:i:sZ
	params.Add("from", request.From.UTC().Format(time.RFC3339))
	params.Add("till", request.Till.UTC().Format(time.RFC3339))

	for _, id := range request.UnitIDs {
		params.Add("unit_id[]", strconv.FormatInt(id, 10))
	}
	for _, inc := range request.Include {
		params.Add("include[]", inc)
	}

	requestURL, err := url.Parse(c.baseURL + "/route/list.json")
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

	var responseBody jsonRouteResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	var routes []*maponv1.Route
	for _, unitRoute := range responseBody.Data.Units {
		for _, r := range unitRoute.Routes {
			routes = append(routes, mapJSONRouteToProto(unitRoute.UnitID, r))
		}
	}

	return &ListRoutesResponse{
		Routes: routes,
	}, nil
}

type jsonRouteResponse struct {
	Data struct {
		Units []jsonUnitRoutes `json:"units"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}

type jsonUnitRoutes struct {
	UnitID int64       `json:"unit_id"`
	Routes []jsonRoute `json:"routes"`
}

type jsonRoute struct {
	RouteID  int64   `json:"route_id"`
	Type     string  `json:"type"` // "route" or "stop"
	Distance int64   `json:"distance"`
	AvgSpeed float64 `json:"avg_speed"`
	MaxSpeed float64 `json:"max_speed"`
	Polyline string  `json:"polyline"`
	DriverID int64   `json:"driver_id"`

	Start jsonRoutePoint `json:"start"`
	End   jsonRoutePoint `json:"end"`
}

type jsonRoutePoint struct {
	Time    string  `json:"time"`
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Can     *struct {
		FuelLevelL float64 `json:"fuel_level_liters"`
		OdometerM  int64   `json:"total_distance"`
	} `json:"can"`
}

func mapJSONRouteToProto(unitID int64, j jsonRoute) *maponv1.Route {
	r := &maponv1.Route{}
	r.SetRouteId(j.RouteID)
	r.SetUnitId(unitID)
	r.SetDriverId(j.DriverID)
	r.SetType(mapRouteType(j.Type))
	r.SetDistanceM(j.Distance)
	r.SetAvgSpeedKmh(j.AvgSpeed)
	r.SetMaxSpeedKmh(j.MaxSpeed)
	r.SetPolyline(j.Polyline)

	if j.Type != "" && r.GetType() == maponv1.RouteType_ROUTE_TYPE_UNRECOGNIZED {
		r.SetUnrecognizedType(j.Type)
	}

	r.SetStart(mapJSONPointToState(j.Start))
	r.SetEnd(mapJSONPointToState(j.End))

	return r
}

func mapJSONPointToState(p jsonRoutePoint) *maponv1.UnitState {
	s := &maponv1.UnitState{}
	
	loc := &maponv1.Location{}
	loc.SetLatitude(p.Lat)
	loc.SetLongitude(p.Lng)
	loc.SetAddress(p.Address)
	s.SetLocation(loc)

	if t, err := time.Parse(time.RFC3339, p.Time); err == nil {
		s.SetTime(timestamppb.New(t))
	}

	if p.Can != nil {
		s.SetFuelLevelL(p.Can.FuelLevelL)
		s.SetOdometerM(p.Can.OdometerM * 1000) // Converting km to meters if CAN uses km, which is typical.
	}

	return s
}

func mapRouteType(t string) maponv1.RouteType {
	switch strings.ToLower(t) {
	case "route":
		return maponv1.RouteType_ROUTE
	case "stop":
		return maponv1.RouteType_STOP
	default:
		return maponv1.RouteType_ROUTE_TYPE_UNRECOGNIZED
	}
}