package event

import (
	"errors"
	"fmt"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"github.com/guregu/null/v6"
	"google.golang.org/protobuf/proto"
	"mesh.leinadium.dev/internal/store"
)

type decoder = func(b []byte) (store.Event, error)

var ErrCouldNotDecode = errors.New("could not decode payload")

func FromBytes(b []byte) (store.Event, error) {
	for _, p := range decoders {
		event, err := p(b)
		if err != nil {
			if errors.Is(err, ErrCouldNotDecode) {
				continue
			}
			return nil, err
		}
		return event, nil
	}
	return nil, ErrCouldNotDecode
}

var decoders []decoder = []decoder{
	asServiceEnvelope,
}

func asServiceEnvelope(b []byte) (store.Event, error) {
	var data meshtastic.ServiceEnvelope
	if err := proto.Unmarshal(b, &data); err != nil {
		return nil, ErrCouldNotDecode
	}

	if data.Packet != nil {
		nodeID := store.NodeID(data.Packet.From)
		decoded := data.Packet.GetDecoded()
		if decoded != nil {
			payload := decoded.GetPayload()
			if payload != nil {
				return decodeServiceEnvelopeData(nodeID, decoded.Portnum, payload)
			}
			return nil, errors.New("serviceEnvelope: payload is empty")
		}
		return nil, errors.New("serviceEnvelope: packet is encrypted")
	}
	return nil, errors.New("serviceEnvelope: packet is nil")
}

func decodeServiceEnvelopeData(id store.NodeID, portNum meshtastic.PortNum, payload []byte) (store.Event, error) {
	switch portNum {
	case meshtastic.PortNum_POSITION_APP:
		var pos meshtastic.Position
		if err := proto.Unmarshal(payload, &pos); err != nil {
			return nil, errors.New("serviceEnvelope: could not decode position payload")
		}
		return &store.Position{
			NodeID:     id,
			Latitude:   parseCoord(pos.LatitudeI),
			Longitude:  parseCoord(pos.LongitudeI),
			Altitude:   null.Int32FromPtr(pos.Altitude),
			Accuracy:   int32(pos.GpsAccuracy),
			Satellites: int32(pos.SatsInView),
			Timestamp:  time.Unix(int64(pos.Timestamp), 0),
		}, nil

	case meshtastic.PortNum_TELEMETRY_APP:
		var tel meshtastic.Telemetry
		if err := proto.Unmarshal(payload, &tel); err != nil {
			return nil, errors.New("serviceEnvelope: could not decode telemetry payload")
		}

		device := tel.GetDeviceMetrics()
		if device == nil {
			return nil, errors.New("serviceEnvelope: device metris is nil in telemetry payload")
		}

		return &store.Telemetry{
			NodeID:  id,
			Battery: null.Int32From(int32(device.GetBatteryLevel())),
		}, nil

	default:
		return nil, fmt.Errorf("serviceEnvelope: unknown type %s", portNum.String())
	}
}

func parseCoord(x *int32) null.Float {
	if x == nil {
		return null.FloatFromPtr(nil)
	}
	return null.FloatFrom(float64(*x) * 1e-7)
}
