package mapon

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	maponv1 "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ParseUnitsResponse parses a raw JSON response from the units endpoint
// and converts it to a slice of protobuf Unit messages.
func ParseUnitsResponse(data []byte) ([]*maponv1.Unit, error) {
	var responseBody jsonUnitResponse
	if err := json.Unmarshal(data, &responseBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("api error %d: %s", responseBody.Error.Code, responseBody.Error.Msg)
	}

	units := make([]*maponv1.Unit, 0, len(responseBody.Data.Units))
	for _, u := range responseBody.Data.Units {
		units = append(units, mapJSONUnitToProto(u))
	}

	return units, nil
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
	UnitID            int64   `json:"unit_id"`
	BoxID             int64   `json:"box_id"`
	CompanyID         int64   `json:"company_id"`
	CountryCode       *string `json:"country_code"`
	Label             *string `json:"label"`
	Number            *string `json:"number"`
	Shortcut          *string `json:"shortcut"`
	VehicleTitle      *string `json:"vehicle_title"`
	CarRegCertificate *string `json:"car_reg_certificate"`
	RegCountry        *string `json:"reg_country"`
	VIN               *string `json:"vin"`
	Type              *string `json:"type"` // "car", "truck", etc.
	Icon              *string `json:"icon"`
	Mileage           float64 `json:"mileage"` // Meters
	Speed             *int32  `json:"speed"`
	Direction         int32   `json:"direction"`
	FuelType          *string `json:"fuel_type"`
	CreatedAt         string  `json:"created_at"`

	LastUpdate string `json:"last_update"` // "2017-05-22T12:23:46Z"

	State struct {
		Name      string `json:"name"` // "standing", "driving"
		Start     string `json:"start"`
		Duration  int64  `json:"duration"`
		DebugInfo *struct {
			Msg        string                 `json:"msg"`
			Data       map[string]interface{} `json:"data"`
			LastValues []interface{}          `json:"lastValues"`
		} `json:"debug_info"`
	} `json:"state"`

	MovementState *struct {
		Name     string `json:"name"`
		Start    string `json:"start"`
		Duration int64  `json:"duration"`
	} `json:"movement_state"`

	// Flattened from API example structure or separate fields
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`

	IgnitionTotalTime int64 `json:"ignition_total_time"`

	AvgFuelConsumption *struct {
		Norm        float64 `json:"norm"`
		Measurement string  `json:"measurement"`
	} `json:"avg_fuel_consumption"`

	Fuel []struct {
		Type       string  `json:"type"`
		Metrics    string  `json:"metrics"` // "L", "KG", "pct"
		Value      float64 `json:"value"`
		LastUpdate *string `json:"last_update"`
	} `json:"fuel"`

	FuelTank map[string]interface{} `json:"fuel_tank"` // Dynamic keys: total_vol, fuel_tank_vol_0, etc.

	SupplyVoltage *struct {
		GMT   string  `json:"gmt"`
		Value float64 `json:"value"`
	} `json:"supply_voltage"`

	BatteryVoltage *struct {
		GMT   string  `json:"gmt"`
		Value float64 `json:"value"`
	} `json:"battery_voltage"`

	Ignition *struct {
		GMT   string `json:"gmt"`
		Value string `json:"value"` // "on" or "off"
	} `json:"ignition"`

	AmbientTemp *struct {
		GMT   string  `json:"gmt"`
		Value float64 `json:"value"`
	} `json:"ambienttemp"`

	Device *struct {
		ID           int64       `json:"id"`
		SerialNumber string      `json:"serial_number"`
		IMEI         interface{} `json:"imei"` // Can be int or string in JSON
		Sim          string      `json:"sim"`
	} `json:"device"`

	Can *struct {
		Odom *struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"odom"`
		FuelTotal *struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"fuel_total"`
		EngineRPM *struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"engine_rpm_avg"`
		CanFuel *struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"can_fuel"`
		EngineHours *struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"engine_hours"`
		ServiceBrakeSwitch *struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"service_brake_switch"`
		ParkingBrakeSwitch *struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"parking_brake_switch"`
		EngineLoad *struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"engine_load"`
	} `json:"can"`

	Weights *struct {
		Axis map[string]struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"axis"`
		CombinationWeight *struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"combination_weight"`
		PoweredWeight *struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"powered_weight"`
	} `json:"weights"`

	EvValues *struct {
		CanEvBatteryRel *struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"can_ev_battery_rel"`
		CanEvBatteryAbs *struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"can_ev_battery_abs"`
		EvCharging *struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"ev_charging"`
		EvChargerConnected *struct {
			GMT   string      `json:"gmt"`
			Value interface{} `json:"value"`
		} `json:"ev_charger_connected"`
	} `json:"ev_values"`

	Altitude *struct {
		GMT   string      `json:"gmt"`
		Value interface{} `json:"value"`
	} `json:"altitude"`

	AdBlueLevelFraction interface{} `json:"adblue_level_fraction"`

	TechnicalDetails *struct {
		StageClassification *string  `json:"stage_classification"`
		EmissionClass       *string  `json:"emission_class"`
		GrossWeight         *int64   `json:"gross_weight"`
		MakeYear            *string  `json:"make_year"`
		MakeMonth           *string  `json:"make_month"`
		PowerPS             *int32   `json:"power_ps"`
		PowerKW             *int32   `json:"power_kw"`
		CubicCapacity       *float64 `json:"cubic_capacity"`
		CO2Emissions        *struct {
			Value   interface{} `json:"value"` // Can be number or string
			Metrics string      `json:"metrics"`
		} `json:"co2_emissions"`
	} `json:"technical_details"`

	Connected *struct {
		UnitID   string `json:"unit_id"`
		Type     string `json:"type"`
		Location *struct {
			Lat string `json:"lat"`
			Lng string `json:"lng"`
		} `json:"location"`
	} `json:"connected"`

	InObjects *struct {
		Objects []struct {
			ObjectID string `json:"object_id"`
			Name     string `json:"name"`
		} `json:"objects"`
	} `json:"in_objects"`

	SavedValues []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
		GMT   string `json:"gmt"`
	} `json:"saved_values"`

	// Complex objects that may need separate handling
	IODin       interface{}      `json:"io_din"`
	Drivers     []jsonDriverUnit `json:"drivers"`
	Relays      []jsonRelay      `json:"relays"`
	Reefer      *jsonReefer      `json:"reefer"`
	Temperature interface{}      `json:"temperature"`
	Humidity    interface{}      `json:"humidity"`
	Tachograph  interface{}      `json:"tachograph"`
	AppFields   interface{}      `json:"application_fields"`
}

type jsonRelay struct {
	RelayID            int32  `json:"relay_id"`
	RelayState         int32  `json:"relay_state"`
	Type               string `json:"type"`
	Title              string `json:"title"`
	Inverted           int32  `json:"inverted"`
	ControlWhileMoving int32  `json:"control_while_moving"`
	Enabled            int32  `json:"enabled"`
}

type jsonDriverUnit struct {
	DriverID     int64  `json:"driver_id"`
	Name         string `json:"name"`
	Surname      string `json:"surname"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	IButtonValue string `json:"ibutton_value"`
	TachographID string `json:"tachograph_id"`
	Blocked      int    `json:"blocked"` // Assuming 0/1 or bool, but using int to be safe or interface
	CreatedAt    string `json:"created_at"`
}

type jsonReefer struct {
	RefrigeratorType              *string `json:"refrigerator_type"`
	RefrigeratorCompartmentCount  *int32  `json:"refrigerator_compartment_count"`
	RefrigeratorCommunicationType *string `json:"refrigerator_communication_type"`
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
	if j.Shortcut != nil && *j.Shortcut != "" {
		u.SetShortcut(*j.Shortcut)
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
		if j.Device.Sim != "" {
			dev.SetSim(j.Device.Sim)
		}
		u.SetDevice(dev)
	}

	// Avg fuel consumption
	if j.AvgFuelConsumption != nil {
		fc := &maponv1.Unit_FuelConsumption{}
		fc.SetNorm(j.AvgFuelConsumption.Norm)
		if j.AvgFuelConsumption.Measurement != "" {
			fc.SetMeasurement(j.AvgFuelConsumption.Measurement)
		}
		u.SetAvgFuelConsumption(fc)
	}

	// Fuel tank - parse dynamic keys
	if j.FuelTank != nil {
		ft := &maponv1.Unit_FuelTank{}
		if totalVol, ok := j.FuelTank["total_vol"].(float64); ok {
			ft.SetTotalVolL(totalVol)
		}
		tankVolumes := make(map[int32]float64)
		for k, v := range j.FuelTank {
			if strings.HasPrefix(k, "fuel_tank_vol_") {
				if axisStr := strings.TrimPrefix(k, "fuel_tank_vol_"); axisStr != "" {
					if axisNum, err := strconv.ParseInt(axisStr, 10, 32); err == nil {
						if vol, ok := v.(float64); ok {
							tankVolumes[int32(axisNum)] = vol
						}
					}
				}
			}
		}
		if len(tankVolumes) > 0 {
			ft.SetTankVolumesL(tankVolumes)
		}
		u.SetFuelTank(ft)
	}

	// Technical details
	if j.TechnicalDetails != nil {
		td := &maponv1.Unit_TechnicalDetails{}
		if j.TechnicalDetails.StageClassification != nil {
			td.SetStageClassification(*j.TechnicalDetails.StageClassification)
		}
		if j.TechnicalDetails.EmissionClass != nil {
			td.SetEmissionClass(*j.TechnicalDetails.EmissionClass)
		}
		if j.TechnicalDetails.GrossWeight != nil {
			td.SetGrossWeightKg(*j.TechnicalDetails.GrossWeight)
		}
		if j.TechnicalDetails.MakeYear != nil {
			td.SetMakeYear(*j.TechnicalDetails.MakeYear)
		}
		if j.TechnicalDetails.MakeMonth != nil {
			td.SetMakeMonth(*j.TechnicalDetails.MakeMonth)
		}
		if j.TechnicalDetails.PowerPS != nil {
			td.SetPowerPs(*j.TechnicalDetails.PowerPS)
		}
		if j.TechnicalDetails.PowerKW != nil {
			td.SetPowerKw(*j.TechnicalDetails.PowerKW)
		}
		if j.TechnicalDetails.CubicCapacity != nil {
			td.SetCubicCapacityL(*j.TechnicalDetails.CubicCapacity)
		}
		if j.TechnicalDetails.CO2Emissions != nil {
			co2 := &maponv1.Unit_CO2Emissions{}
			// Convert value to string (can be number or string in JSON)
			valueStr := fmt.Sprintf("%v", j.TechnicalDetails.CO2Emissions.Value)
			co2.SetValue(valueStr)
			co2.SetMetrics(j.TechnicalDetails.CO2Emissions.Metrics)
			td.SetCo2Emissions(co2)
		}
		u.SetTechnicalDetails(td)
	}

	// Movement state
	if j.MovementState != nil {
		ms := &maponv1.Unit_MovementState{}
		if j.MovementState.Name != "" {
			ms.SetName(j.MovementState.Name)
		}
		if t, err := time.Parse(time.RFC3339, j.MovementState.Start); err == nil {
			ms.SetStart(timestamppb.New(t))
		}
		ms.SetDurationS(j.MovementState.Duration)
		u.SetMovementState(ms)
	}

	// Connected trailer
	if j.Connected != nil {
		ct := &maponv1.Unit_ConnectedTrailer{}
		if j.Connected.UnitID != "" {
			ct.SetUnitId(j.Connected.UnitID)
		}
		if j.Connected.Type != "" {
			ct.SetType(j.Connected.Type)
		}
		if j.Connected.Location != nil {
			loc := &maponv1.Location{}
			if lat, err := strconv.ParseFloat(j.Connected.Location.Lat, 64); err == nil {
				loc.SetLatitude(lat)
			}
			if lng, err := strconv.ParseFloat(j.Connected.Location.Lng, 64); err == nil {
				loc.SetLongitude(lng)
			}
			ct.SetLocation(loc)
		}
		u.SetConnected(ct)
	}

	// In objects
	if j.InObjects != nil && len(j.InObjects.Objects) > 0 {
		objects := make([]*maponv1.Unit_ObjectLocation, 0, len(j.InObjects.Objects))
		for _, obj := range j.InObjects.Objects {
			ol := &maponv1.Unit_ObjectLocation{}
			ol.SetObjectId(obj.ObjectID)
			ol.SetName(obj.Name)
			objects = append(objects, ol)
		}
		u.SetInObjects(objects)
	}

	// Saved values
	if len(j.SavedValues) > 0 {
		savedVals := make([]*maponv1.Unit_SavedValue, 0, len(j.SavedValues))
		for _, sv := range j.SavedValues {
			svProto := &maponv1.Unit_SavedValue{}
			svProto.SetKey(sv.Key)
			svProto.SetValue(sv.Value)
			if t, err := time.Parse("2006-01-02 15:04:05", sv.GMT); err == nil {
				svProto.SetGmt(timestamppb.New(t))
			}
			savedVals = append(savedVals, svProto)
		}
		u.SetSavedValues(savedVals)
	}

	// Drivers
	if len(j.Drivers) > 0 {
		drivers := make([]*maponv1.Driver, 0, len(j.Drivers))
		for _, d := range j.Drivers {
			driver := &maponv1.Driver{}
			driver.SetDriverId(d.DriverID)
			driver.SetName(d.Name)
			driver.SetSurname(d.Surname)
			driver.SetEmail(d.Email)
			driver.SetPhone(d.Phone)
			driver.SetIbuttonValue(d.IButtonValue)
			driver.SetTachographId(d.TachographID)
			driver.SetBlocked(d.Blocked > 0)
			if t, err := time.Parse(time.RFC3339, d.CreatedAt); err == nil {
				driver.SetCreatedAt(timestamppb.New(t))
			}
			drivers = append(drivers, driver)
		}
		u.SetDrivers(drivers)
	}

	// Relays
	if len(j.Relays) > 0 {
		relays := make([]*maponv1.Unit_Relay, 0, len(j.Relays))
		for _, r := range j.Relays {
			relay := &maponv1.Unit_Relay{}
			relay.SetRelayId(r.RelayID)
			relay.SetRelayState(r.RelayState)
			relay.SetType(r.Type)
			relay.SetTitle(r.Title)
			relay.SetInverted(r.Inverted)
			relay.SetControlWhileMoving(r.ControlWhileMoving)
			relay.SetEnabled(r.Enabled)
			relays = append(relays, relay)
		}
		u.SetRelays(relays)
	}

	// Reefer
	if j.Reefer != nil {
		reefer := &maponv1.Reefer{}
		if j.Reefer.RefrigeratorType != nil {
			reefer.SetRefrigeratorType(*j.Reefer.RefrigeratorType)
		}
		if j.Reefer.RefrigeratorCompartmentCount != nil {
			reefer.SetRefrigeratorCompartmentCount(*j.Reefer.RefrigeratorCompartmentCount)
		}
		if j.Reefer.RefrigeratorCommunicationType != nil {
			reefer.SetRefrigeratorCommunicationType(*j.Reefer.RefrigeratorCommunicationType)
		}
		u.SetReefer(reefer)
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

	// Fuel entries
	if len(j.Fuel) > 0 {
		fuelEntries := make([]*maponv1.UnitState_FuelEntry, 0, len(j.Fuel))
		for _, f := range j.Fuel {
			fe := &maponv1.UnitState_FuelEntry{}
			if f.Type != "" {
				fe.SetType(f.Type)
			}
			if f.Metrics != "" {
				fe.SetMetrics(f.Metrics)
			}
			fe.SetValue(f.Value)
			if f.LastUpdate != nil && *f.LastUpdate != "" {
				if t, err := time.Parse(time.RFC3339, *f.LastUpdate); err == nil {
					fe.SetLastUpdate(timestamppb.New(t))
				}
			}
			fuelEntries = append(fuelEntries, fe)
			// Also set fuel_level_l for backward compatibility (first L metric)
			if f.Metrics == "L" && state.GetFuelLevelL() == 0 {
				state.SetFuelLevelL(f.Value)
			}
		}
		state.SetFuelEntries(fuelEntries)
	}

	// Voltages
	if j.SupplyVoltage != nil {
		state.SetSupplyVoltageV(j.SupplyVoltage.Value)
		if j.SupplyVoltage.GMT != "" {
			if t, err := time.Parse(time.RFC3339, j.SupplyVoltage.GMT); err == nil {
				state.SetSupplyVoltageTime(timestamppb.New(t))
			}
		}
	}
	if j.BatteryVoltage != nil {
		state.SetBatteryVoltageV(j.BatteryVoltage.Value)
		if j.BatteryVoltage.GMT != "" {
			if t, err := time.Parse(time.RFC3339, j.BatteryVoltage.GMT); err == nil {
				state.SetBatteryVoltageTime(timestamppb.New(t))
			}
		}
	}

	// Ignition
	if j.Ignition != nil {
		state.SetIgnitionState(j.Ignition.Value == "on")
		if j.Ignition.GMT != "" {
			if t, err := time.Parse(time.RFC3339, j.Ignition.GMT); err == nil {
				state.SetIgnitionTime(timestamppb.New(t))
			}
		}
	}

	// Ambient temperature
	if j.AmbientTemp != nil {
		state.SetAmbientTemperatureC(j.AmbientTemp.Value)
		if j.AmbientTemp.GMT != "" {
			if t, err := time.Parse(time.RFC3339, j.AmbientTemp.GMT); err == nil {
				state.SetAmbientTemperatureTime(timestamppb.New(t))
			}
		}
	}

	// Debug info
	if j.State.DebugInfo != nil && j.State.DebugInfo.Msg != "" {
		state.SetDebugMessage(j.State.DebugInfo.Msg)
	}

	// CAN fields
	if j.Can != nil {
		if j.Can.Odom != nil && j.Can.Odom.Value != nil {
			if j.Can.Odom.GMT != "" {
				if t, err := time.Parse(time.RFC3339, j.Can.Odom.GMT); err == nil {
					state.SetCanOdometerTime(timestamppb.New(t))
				}
			}
		}
		if j.Can.FuelTotal != nil && j.Can.FuelTotal.Value != nil {
			v, _ := parseFloat(j.Can.FuelTotal.Value)
			state.SetTotalFuelUsedLifetimeL(v)
			if j.Can.FuelTotal.GMT != "" {
				if t, err := time.Parse(time.RFC3339, j.Can.FuelTotal.GMT); err == nil {
					state.SetCanFuelTotalTime(timestamppb.New(t))
				}
			}
		}
		if j.Can.EngineRPM != nil && j.Can.EngineRPM.Value != nil {
			v, _ := parseFloat(j.Can.EngineRPM.Value)
			state.SetCanEngineRpm(v)
			if j.Can.EngineRPM.GMT != "" {
				if t, err := time.Parse(time.RFC3339, j.Can.EngineRPM.GMT); err == nil {
					state.SetCanEngineRpmTime(timestamppb.New(t))
				}
			}
		}
		if j.Can.CanFuel != nil && j.Can.CanFuel.Value != nil {
			v, _ := parseFloat(j.Can.CanFuel.Value)
			state.SetCanFuelLevelL(v)
			if j.Can.CanFuel.GMT != "" {
				if t, err := time.Parse(time.RFC3339, j.Can.CanFuel.GMT); err == nil {
					state.SetCanFuelLevelTime(timestamppb.New(t))
				}
			}
		}
		if j.Can.EngineHours != nil && j.Can.EngineHours.Value != nil {
			v, _ := parseFloat(j.Can.EngineHours.Value)
			state.SetCanEngineHoursH(v)
			if j.Can.EngineHours.GMT != "" {
				if t, err := time.Parse(time.RFC3339, j.Can.EngineHours.GMT); err == nil {
					state.SetCanEngineHoursTime(timestamppb.New(t))
				}
			}
		}
		// Extended CAN
		if j.Can.ServiceBrakeSwitch != nil && j.Can.ServiceBrakeSwitch.Value != nil {
			v, _ := parseFloat(j.Can.ServiceBrakeSwitch.Value)
			state.SetCanServiceBrakeSwitch(v > 0.5)
			if j.Can.ServiceBrakeSwitch.GMT != "" {
				if t, err := time.Parse(time.RFC3339, j.Can.ServiceBrakeSwitch.GMT); err == nil {
					state.SetCanServiceBrakeSwitchTime(timestamppb.New(t))
				}
			}
		}
		if j.Can.ParkingBrakeSwitch != nil && j.Can.ParkingBrakeSwitch.Value != nil {
			v, _ := parseFloat(j.Can.ParkingBrakeSwitch.Value)
			state.SetCanParkingBrakeSwitch(v > 0.5)
			if j.Can.ParkingBrakeSwitch.GMT != "" {
				if t, err := time.Parse(time.RFC3339, j.Can.ParkingBrakeSwitch.GMT); err == nil {
					state.SetCanParkingBrakeSwitchTime(timestamppb.New(t))
				}
			}
		}
		if j.Can.EngineLoad != nil && j.Can.EngineLoad.Value != nil {
			v, _ := parseFloat(j.Can.EngineLoad.Value)
			state.SetCanEngineLoadPercent(v)
			if j.Can.EngineLoad.GMT != "" {
				if t, err := time.Parse(time.RFC3339, j.Can.EngineLoad.GMT); err == nil {
					state.SetCanEngineLoadTime(timestamppb.New(t))
				}
			}
		}
	}

	// Weights
	if j.Weights != nil {
		if j.Weights.CombinationWeight != nil && j.Weights.CombinationWeight.Value != nil {
			v, _ := parseFloat(j.Weights.CombinationWeight.Value)
			state.SetGrossCombinationWeightKg(v)
			if j.Weights.CombinationWeight.GMT != "" {
				if t, err := time.Parse(time.RFC3339, j.Weights.CombinationWeight.GMT); err == nil {
					state.SetCombinationWeightTime(timestamppb.New(t))
				}
			}
		}
		if j.Weights.PoweredWeight != nil && j.Weights.PoweredWeight.Value != nil {
			v, _ := parseFloat(j.Weights.PoweredWeight.Value)
			state.SetPoweredWeightKg(v)
			if j.Weights.PoweredWeight.GMT != "" {
				if t, err := time.Parse(time.RFC3339, j.Weights.PoweredWeight.GMT); err == nil {
					state.SetPoweredWeightTime(timestamppb.New(t))
				}
			}
		}
		if len(j.Weights.Axis) > 0 {
			axisWeights := make(map[int32]*maponv1.UnitState_AxisWeight)
			for axisStr, axisData := range j.Weights.Axis {
				if axisNum, err := strconv.ParseInt(axisStr, 10, 32); err == nil {
					aw := &maponv1.UnitState_AxisWeight{}
					if axisData.Value != nil {
						v, _ := parseFloat(axisData.Value)
						aw.SetWeightKg(v)
					}
					if axisData.GMT != "" {
						if t, err := time.Parse(time.RFC3339, axisData.GMT); err == nil {
							aw.SetTime(timestamppb.New(t))
						}
					}
					axisWeights[int32(axisNum)] = aw
				}
			}
			if len(axisWeights) > 0 {
				state.SetAxisWeights(axisWeights)
			}
		}
	}

	// EV Values
	if j.EvValues != nil {
		if j.EvValues.CanEvBatteryRel != nil && j.EvValues.CanEvBatteryRel.Value != nil {
			v, _ := parseFloat(j.EvValues.CanEvBatteryRel.Value)
			state.SetBatterySocPercent(v)
			if j.EvValues.CanEvBatteryRel.GMT != "" {
				if t, err := time.Parse(time.RFC3339, j.EvValues.CanEvBatteryRel.GMT); err == nil {
					state.SetBatterySocPercentTime(timestamppb.New(t))
				}
			}
		}
		if j.EvValues.CanEvBatteryAbs != nil && j.EvValues.CanEvBatteryAbs.Value != nil {
			v, _ := parseFloat(j.EvValues.CanEvBatteryAbs.Value)
			state.SetBatterySocKwh(v)
			if j.EvValues.CanEvBatteryAbs.GMT != "" {
				if t, err := time.Parse(time.RFC3339, j.EvValues.CanEvBatteryAbs.GMT); err == nil {
					state.SetBatterySocKwhTime(timestamppb.New(t))
				}
			}
		}
		if j.EvValues.EvCharging != nil && j.EvValues.EvCharging.Value != nil {
			v, _ := parseFloat(j.EvValues.EvCharging.Value)
			state.SetChargingState(v > 0)
			if j.EvValues.EvCharging.GMT != "" {
				if t, err := time.Parse(time.RFC3339, j.EvValues.EvCharging.GMT); err == nil {
					state.SetEvChargingTime(timestamppb.New(t))
				}
			}
		}
		if j.EvValues.EvChargerConnected != nil && j.EvValues.EvChargerConnected.Value != nil {
			v, _ := parseFloat(j.EvValues.EvChargerConnected.Value)
			state.SetEvChargerConnected(v > 0)
			if j.EvValues.EvChargerConnected.GMT != "" {
				if t, err := time.Parse(time.RFC3339, j.EvValues.EvChargerConnected.GMT); err == nil {
					state.SetEvChargerConnectedTime(timestamppb.New(t))
				}
			}
		}
	}

	// Altitude
	if j.Altitude != nil {
		if j.Altitude.Value != nil {
			v, _ := parseFloat(j.Altitude.Value)
			state.SetAltitudeM(v)
		}
		if j.Altitude.GMT != "" {
			if t, err := time.Parse(time.RFC3339, j.Altitude.GMT); err == nil {
				state.SetAltitudeTime(timestamppb.New(t))
			}
		}
	}

	// AdBlue
	if j.AdBlueLevelFraction != nil {
		v, _ := parseFloat(j.AdBlueLevelFraction)
		state.SetAdblueLevelFraction(v)
	}

	u.SetState(state)
	return u
}

func parseFloat(v interface{}) (float64, error) {
	if v == nil {
		return 0, nil
	}
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("unknown type")
	}
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
