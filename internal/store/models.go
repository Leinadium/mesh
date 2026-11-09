package store

import (
	"time"

	"github.com/guregu/null/v6"
)

type EventType string

const (
	EventTypePosition  EventType = "position"
	EventTypeTelemetry EventType = "telemetry"
	EventTypeUser      EventType = "user"
)

type Event interface {
	Type() EventType
}

type NodeID int

type User struct {
	NodeID    NodeID `gorm:"primary_key" json:"id"`
	ShortName string `json:"short_name"`
}

func (u User) Type() EventType {
	return EventTypeUser
}

func (User) TableName() string {
	return "users"
}

type Position struct {
	NodeID     NodeID     `gorm:"primary_key" json:"id"`
	Latitude   null.Float `json:"lat"`
	Longitude  null.Float `json:"lon"`
	Altitude   null.Int32 `json:"alt"`
	Accuracy   int32      `json:"acc"`
	Satellites int32      `json:"sat"`

	ReceivedAt time.Time `gorm:"primary_key" json:"received_at"`
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

	ReceivedAt time.Time `gorm:"primary_key" json:"received_at"`
}

func (p Telemetry) Type() EventType {
	return EventTypeTelemetry
}

func (Telemetry) TableName() string {
	return "telemtries"
}

type Node struct {
	ID               NodeID      `json:"id"`
	Name             string      `json:"name"`
	CurrentTelemetry *Telemetry  `json:"current_telemetry"`
	LatestPositions  []*Position `json:"latest_positions"`
}
