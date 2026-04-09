package mapon

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	maponv1 "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// rawPushEvent covers the flat JSON fields across all Mapon push pack types.
type rawPushEvent struct {
	ID          int     `json:"id"`
	CarID       int64   `json:"car_id"`
	DeviceID    int64   `json:"device_id"`
	CompanyID   int64   `json:"company_id"`
	PackID      int32   `json:"pack_id"`
	GMT         string  `json:"gmt"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Speed       float64 `json:"speed"`
	Direction   float64 `json:"direction"`
	Altitude    float64 `json:"altitude"`
	State       bool    `json:"state"`
	Liters      float64 `json:"liters"`
	Odometer    float64 `json:"odometer"`
	SensorID    int32   `json:"sensor_id"`
	Temperature float64 `json:"temperature"`
}

// gmtLayout is the datetime format used in Mapon push payloads ("YYYY-MM-DD HH:MM:SS" in UTC).
const gmtLayout = "2006-01-02 15:04:05"

// packIDToType maps Mapon wire pack IDs to PushMessage_Type enum values,
// derived from the mapon_pack_id annotations on the PushMessage.Type enum.
var packIDToType = func() map[int32]maponv1.PushMessage_Type {
	m := make(map[int32]maponv1.PushMessage_Type)
	enumDesc := maponv1.PushMessage_TYPE_UNSPECIFIED.Descriptor()
	for i := range enumDesc.Values().Len() {
		v := enumDesc.Values().Get(i)
		packID, ok := proto.GetExtension(v.Options(), maponv1.E_MaponPackId).(int32)
		if ok && packID != 0 {
			m[packID] = maponv1.PushMessage_Type(v.Number())
		}
	}
	return m
}()

// ParsePushMessage parses a raw Mapon push JSON payload into a maponv1.PushMessage proto.
func ParsePushMessage(data []byte) (*maponv1.PushMessage, error) {
	var event rawPushEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("unmarshal push event: %w", err)
	}
	t, err := time.ParseInLocation(gmtLayout, event.GMT, time.UTC)
	if err != nil {
		return nil, fmt.Errorf("parse gmt %q: %w", event.GMT, err)
	}
	var msg maponv1.PushMessage
	msg.SetId(int64(event.ID))
	msg.SetCarId(event.CarID)
	msg.SetDeviceId(event.DeviceID)
	msg.SetCompanyId(event.CompanyID)
	msg.SetPackId(event.PackID)
	msg.SetVehicleTime(timestamppb.New(t))
	msgType, ok := packIDToType[event.PackID]
	if !ok {
		msgType = maponv1.PushMessage_TYPE_UNRECOGNIZED
		slog.Warn("unrecognized pack_id", "packId", event.PackID)
	}
	msg.SetType(msgType)
	switch msgType {
	case maponv1.PushMessage_TYPE_POSITION:
		var pos maponv1.PushMessage_Position
		pos.SetLatitude(event.Lat)
		pos.SetLongitude(event.Lng)
		pos.SetSpeedKmh(event.Speed)
		pos.SetHeadingDeg(event.Direction)
		pos.SetAltitudeM(event.Altitude)
		msg.SetPosition(&pos)
	case maponv1.PushMessage_TYPE_IGNITION:
		var ign maponv1.PushMessage_Ignition
		ign.SetState(event.State)
		msg.SetIgnition(&ign)
	case maponv1.PushMessage_TYPE_FUEL:
		var fuel maponv1.PushMessage_Fuel
		fuel.SetLevelL(event.Liters)
		msg.SetFuel(&fuel)
	case maponv1.PushMessage_TYPE_ODOMETER:
		var odo maponv1.PushMessage_Odometer
		odo.SetValueM(event.Odometer)
		msg.SetOdometer(&odo)
	case maponv1.PushMessage_TYPE_TEMPERATURE:
		var temp maponv1.PushMessage_Temperature
		temp.SetSensorId(event.SensorID)
		temp.SetValueC(event.Temperature)
		msg.SetTemperature(&temp)
	}
	return &msg, nil
}
