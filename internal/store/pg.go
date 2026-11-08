package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"mesh.leinadium.dev/internal/logger"
)

type PGParams struct {
	DSN         string
	AutoMigrate bool
}

type PGStore struct {
	db       *gorm.DB
	streamer *streamer[Event]
}

func NewPGStore(p PGParams) *PGStore {
	dsn := p.DSN
	if dsn == "" {
		panic("dsn is required")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("newpgstore", zap.Error(err))
		panic("failed to connect to database")
	}

	if p.AutoMigrate {
		_ = db.AutoMigrate(&Position{})
		_ = db.AutoMigrate(&Telemetry{})
	}

	streamer := Streamer[Event]()
	streamer.Start()

	return &PGStore{
		db:       db,
		streamer: streamer,
	}
}

func (s *PGStore) NewEvent(ctx context.Context, event Event) error {
	s.streamer.Publish(event)

	switch event.Type() {
	case EventTypePosition:
		pos, ok := event.(*Position)
		if !ok {
			return errors.New("pgstore new event: position is unknown")
		} else {
			return gorm.G[Position](s.db).Create(ctx, pos)
		}
	case EventTypeTelemetry:
		tel, ok := event.(*Telemetry)
		if !ok {
			return errors.New("pgstore new event: position is unknown")
		} else {
			return gorm.G[Telemetry](s.db).Create(ctx, tel)
		}
	}
	return fmt.Errorf("pgstore new event: unknown event %s", event.Type())
}

func (s *PGStore) Nodes(ctx context.Context, from time.Time) ([]*Node, error) {
	// getting positions
	positions, err := gorm.G[Position](s.db).Where("received_at > ?", from).Find(ctx)
	if err != nil {
		return nil, err
	}

	// getting telemetries
	subQuery := s.db.Table("telemtries").
		Select("*, ROW_NUMBER() OVER (PARTITION BY node_id ORDER BY received_at DESC) as rn")

	var telemetries []Telemetry
	err = s.db.Table("(?) as ranked", subQuery).
		Where("rn = ?", 1).
		Find(&telemetries).Error
	if err != nil {
		return nil, err
	}

	// merging
	m := make(map[NodeID]*Node)
	for _, pos := range positions {
		node, ok := m[pos.NodeID]
		if !ok {
			node = &Node{ID: pos.NodeID}
		}
		node.LatestPositions = append(node.LatestPositions, &pos)
		m[pos.NodeID] = node
	}
	for _, telemetry := range telemetries {
		node, ok := m[telemetry.NodeID]
		if !ok {
			node = &Node{ID: telemetry.NodeID}
		}
		node.CurrentTelemetry = &telemetry
		m[telemetry.NodeID] = node
	}

	var ret []*Node
	for _, node := range m {
		ret = append(ret, node)
	}

	return ret, nil
}

func (s *PGStore) Node(ctx context.Context, id NodeID) (*Node, error) {
	// getting positions
	positions, err := gorm.G[*Position](s.db).Find(ctx)
	if err != nil {
		return nil, err
	}

	var telemetry Telemetry
	err = s.db.Where("id = ?", id).
		Order("time DESC").
		First(&telemetry).Error
	if err != nil {
		return nil, err
	}

	return &Node{
		ID:               id,
		CurrentTelemetry: &telemetry,
		LatestPositions:  positions,
	}, nil
}

func (s *PGStore) Stream() chan Event {
	return s.streamer.Subscribe()
}

func (s *PGStore) CancelStream(ch chan Event) {
	s.streamer.Unsubscribe(ch)
}

func (s *PGStore) Stop() {
	s.streamer.Stop()
}
