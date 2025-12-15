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
// docs/api/methods/08-method-unit.html

// ListUnitsRequest is the request for [Client.ListUnits].
type ListUnitsRequest struct {
	// UnitIDs is a list of unit IDs to filter by.
	UnitIDs []int64
	// Include is a list of additional data to include in the response.
	// Possible values: "fuel", "drivers", "location", "routes".
	Include []string
}

// ListUnitsResponse is the response for [Client.ListUnits].
type ListUnitsResponse struct {
	// Units is the list of units returned by the API.
	Units []*maponv1.Unit
}

// ListUnits lists the units available for the current API key.
func (c *Client) ListUnits(ctx context.Context, request *ListUnitsRequest, opts ...ClientOption) (_ *ListUnitsResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("mapon: list units: %w", err)
		}
	}()
	cfg := c.config.with(opts...)

	params := url.Values{}
	for _, id := range request.UnitIDs {
		params.Add("unit_id[]", strconv.FormatInt(id, 10))
	}
	for _, inc := range request.Include {
		params.Add("include[]", inc)
	}

	requestURL, err := url.Parse(c.baseURL + "/unit/list.json")
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

	var responseBody jsonUnitResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	units := make([]*maponv1.Unit, 0, len(responseBody.Data.Units))
	for _, u := range responseBody.Data.Units {
		units = append(units, mapJSONUnitToProto(u))
	}

	return &ListUnitsResponse{
		Units: units,
	}, nil
}

// Internal JSON structs

type jsonUnitResponse struct {
	Data struct {
		Units []jsonUnit `json:"units"`
	} `json:"data"`
	Error *jsonError `json:"error"`
}

type jsonError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type jsonUnit struct {
	UnitID            int64       `json:"unit_id"`
	BoxID             int64       `json:"box_id"`
	CompanyID         int64       `json:"company_id"`
	CountryCode       *string     `json:"country_code"`
	Label             *string     `json:"label"`
	Number            *string     `json:"number"`
	VehicleTitle      *string     `json:"vehicle_title"`
	CarRegCertificate *string     `json:"car_reg_certificate"`
	RegCountry        *string     `json:"reg_country"`
	VIN               *string     `json:"vin"`
	Type              *string     `json:"type"` // "car", "truck", etc.
	Icon              *string     `json:"icon"`
	Mileage           float64     `json:"mileage"` // Meters
	Speed             *int32      `json:"speed"`
	Direction         int32       `json:"direction"`
	FuelType          *string     `json:"fuel_type"`
	CreatedAt         string      `json:"created_at"`

	LastUpdate string `json:"last_update"` // "2017-05-22T12:23:46Z"

	State struct {
		Name     string `json:"name"` // "standing", "driving"
		Start    string `json:"start"`
		Duration int64  `json:"duration"`
	} `json:"state"`

	// Flattened from API example structure or separate fields
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`

	IgnitionTotalTime int64 `json:"ignition_total_time"`

	Fuel []struct {
		Type    string  `json:"type"`
		Metrics string  `json:"metrics"` // "L", "KG"
		Value   float64 `json:"value"`
	} `json:"fuel"`

	SupplyVoltage *struct {
		Value float64 `json:"value"`
	} `json:"supply_voltage"`

	BatteryVoltage *struct {
		Value float64 `json:"value"`
	} `json:"battery_voltage"`

	Device *struct {
		ID           int64       `json:"id"`
		SerialNumber string      `json:"serial_number"`
		IMEI         interface{} `json:"imei"` // Can be int or string in JSON
		Sim          string      `json:"sim"`
	} `json:"device"`
}

func mapJSONUnitToProto(j jsonUnit) *maponv1.Unit {
	u := &maponv1.Unit{}
	u.SetUnitId(j.UnitID)
		u.SetCompanyId(j.CompanyID)
		u.SetBoxId(j.BoxID)
		if j.Label != nil && *j.Label != "" {
			u.SetLabel(*j.Label)
		}
		if j.Number != nil && *j.Number != "" {
			u.SetNumber(*j.Number)
		}
		if j.CountryCode != nil && *j.CountryCode != "" {
			u.SetCountryCode(*j.CountryCode)
		}
		if j.VehicleTitle != nil && *j.VehicleTitle != "" {
			u.SetVehicleTitle(*j.VehicleTitle)
		}
		if j.CarRegCertificate != nil && *j.CarRegCertificate != "" {
			u.SetCarRegCertificate(*j.CarRegCertificate)
		}
		if j.RegCountry != nil && *j.RegCountry != "" {
			u.SetRegCountry(*j.RegCountry)
		}
		if j.VIN != nil && *j.VIN != "" {
			u.SetVin(*j.VIN)
		}
		
	typeStr := ""
		if j.Type != nil {
			typeStr = *j.Type
			if typeStr != "" {
				u.SetType(mapUnitType(typeStr))
			}
		}
	
			if j.Icon != nil && *j.Icon != "" {
	
				u.SetIcon(*j.Icon)
	
			}
	
			
	
			if j.FuelType != nil && *j.FuelType != "" {
	
				fuelType := mapFuelType(*j.FuelType)
	
				u.SetFuelType(fuelType)
	
				if fuelType == maponv1.FuelType_FUEL_TYPE_UNRECOGNIZED {
	
					u.SetUnrecognizedFuelType(*j.FuelType)
	
				}
	
			}
	
			
	
			if t, err := time.Parse(time.RFC3339, j.CreatedAt); err == nil {
	
				u.SetCreatedAt(timestamppb.New(t))
	
			}
	
		
	
			if typeStr != "" && u.GetType() == maponv1.UnitType_UNIT_TYPE_UNRECOGNIZED {
	
				u.SetUnrecognizedType(typeStr)
	
			}
	
		
	
			// Device
	
			if j.Device != nil {
	
				imeiStr := fmt.Sprintf("%v", j.Device.IMEI)
	
				dev := &maponv1.Unit_Device{}
	
				dev.SetDeviceId(j.Device.ID)
	
				dev.SetSerialNumber(j.Device.SerialNumber)
	
				if imeiStr != "" {
	
					dev.SetImei(imeiStr)
	
				}
	
				u.SetDevice(dev)
	
			}
	
		
	
			// State
	
			state := &maponv1.UnitState{}
	
			
	
			loc := &maponv1.Location{}
	
			loc.SetLatitude(j.Lat)
	
			loc.SetLongitude(j.Lng)
	
			state.SetLocation(loc)
	
		
	
			if j.Speed != nil {
	
				state.SetSpeedKmh(*j.Speed)
	
			}
	
			state.SetDirectionDeg(j.Direction)
	
			state.SetOdometerM(int64(j.Mileage))
	
			state.SetIgnitionTotalDurationS(j.IgnitionTotalTime)
	
			
	
			if j.State.Name != "" {
	
				moveStatus := mapMovementStatus(j.State.Name)
	
				state.SetMovementStatus(moveStatus)
	
				if moveStatus == maponv1.MovementStatus_MOVEMENT_STATUS_UNRECOGNIZED {
	
					state.SetUnrecognizedMovementStatus(j.State.Name)
	
				}
	
			}
	
			
	
			state.SetDurationS(j.State.Duration)
	
		
	
			if t, err := time.Parse(time.RFC3339, j.State.Start); err == nil {
	
				state.SetStartTime(timestamppb.New(t))
	
			}
	
		
	
			// Time
	
			if t, err := time.Parse(time.RFC3339, j.LastUpdate); err == nil {
	
				state.SetTime(timestamppb.New(t))
	
			}
	
		
	
			// Fuel
	
			for _, f := range j.Fuel {
	
				// Assuming we want the first valid liquid fuel or just the first one
	
				if f.Metrics == "L" {
	
					state.SetFuelLevelL(f.Value)
	
					break
	
				}
	
			}
	
		
	
			// Voltages
	
			if j.SupplyVoltage != nil {
	
				state.SetSupplyVoltageV(j.SupplyVoltage.Value)
	
			}
	
			if j.BatteryVoltage != nil {
	
				state.SetBatteryVoltageV(j.BatteryVoltage.Value)
	
			}
	
		
	
			u.SetState(state)
	
			return u
	
		}
	
		
	
		func mapUnitType(t string) maponv1.UnitType {
	
			switch strings.ToLower(t) {
	
			case "car":
	
				return maponv1.UnitType_CAR
	
			case "truck":
	
				return maponv1.UnitType_TRUCK
	
			case "trailer":
	
				return maponv1.UnitType_TRAILER
	
			case "van":
	
				return maponv1.UnitType_VAN
	
			case "bus":
	
				return maponv1.UnitType_BUS
	
			case "tractor":
	
				return maponv1.UnitType_TRACTOR
	
			default:
	
				return maponv1.UnitType_UNIT_TYPE_UNRECOGNIZED
	
			}
	
		}
	
		
	
		func mapFuelType(t string) maponv1.FuelType {
	
			switch strings.ToUpper(t) {
	
			case "P":
	
				return maponv1.FuelType_PETROL
	
			case "D":
	
				return maponv1.FuelType_DIESEL
	
			case "G":
	
				return maponv1.FuelType_LPG
	
			case "E":
	
				return maponv1.FuelType_ELECTRIC
	
			case "PROPANE":
	
				return maponv1.FuelType_PROPANE
	
			case "LNG":
	
				return maponv1.FuelType_LNG
	
			case "CNG":
	
				return maponv1.FuelType_CNG
	
			case "ETHANOL":
	
				return maponv1.FuelType_ETHANOL
	
			case "HYDROGEN":
	
				return maponv1.FuelType_HYDROGEN
	
			case "HYBRID":
	
				return maponv1.FuelType_HYBRID
	
			case "L":
	
				return maponv1.FuelType_AGRICULTURE_FUEL
	
			default:
	
				return maponv1.FuelType_FUEL_TYPE_UNRECOGNIZED
	
			}
	
		}
	
		
	
		func mapMovementStatus(s string) maponv1.MovementStatus {
	
			switch strings.ToLower(s) {
	
			case "driving":
	
				return maponv1.MovementStatus_DRIVING
	
			case "standing":
	
				return maponv1.MovementStatus_STANDING
	
			case "nodata":
	
				return maponv1.MovementStatus_NODATA
	
			case "nogps":
	
				return maponv1.MovementStatus_NOGPS
	
			case "service":
	
				return maponv1.MovementStatus_SERVICE
	
			default:
	
				return maponv1.MovementStatus_MOVEMENT_STATUS_UNRECOGNIZED
	
			}
	
		}