package store

import (
	"time"

	"github.com/guregu/null/v6"
)

type EventType string

const (
	EventTypePosition  EventType = "position"
	EventTypeTelemetry EventType = "telemetry"
)

type Event interface {
	Type() EventType
}

type NodeID int

type Position struct {
	NodeID     NodeID     `gorm:"primary_key" json:"id"`
	Latitude   null.Float `json:"lat"`
	Longitude  null.Float `json:"lon"`
	Altitude   null.Int32 `json:"alt"`
	Accuracy   int32      `json:"acc"`
	Satellites int32      `json:"sat"`

	Timestamp time.Time `json:"time"`
}

func (p Position) Type() EventType {
	return EventTypePosition
}

func (Position) TableName() string {
	return "positions"
}

type Telemetry struct {
	NodeID  NodeID     `gorm:"primary_key" json:"id"`
	Battery null.Int32 `json:"bat"`

	Timestamp time.Time `json:"time"`
}

func (p Telemetry) Type() EventType {
	return EventTypeTelemetry
}

func (Telemetry) TableName() string {
	return "telemtries"
}

type Node struct {
	ID               NodeID      `json:"id"`
	CurrentTelemetry *Telemetry  `json:"current_telemetry"`
	LatestPositions  []*Position `json:"latest_positions"`
}
