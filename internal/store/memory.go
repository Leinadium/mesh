package store

import (
	"context"
	"fmt"
	"sync"
)

type log struct {
	Type  EventType
	Event Event
}

type MemoryStore struct {
	logs []log
	m    map[NodeID]*Node
	ch   chan Event

	lock sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		logs: make([]log, 0),
		m:    make(map[NodeID]*Node),
		lock: sync.RWMutex{},
		ch:   make(chan Event),
	}
}

func (s *MemoryStore) NewEvent(ctx context.Context, event Event) error {
	return nil
}

func (s *MemoryStore) Nodes(ctx context.Context) ([]*Node, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var ret []*Node
	for _, node := range s.m {
		ret = append(ret, node)
	}
	return ret, nil
}

func (s *MemoryStore) Node(ctx context.Context, id NodeID) (*Node, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	node, ok := s.m[id]
	if !ok {
		return nil, fmt.Errorf("unknown node")
	}
	return node, nil
}

func (s *MemoryStore) Stream(context.Context) chan<- Event {
	return nil
}

type positions struct {
	p    []*Position
	size int
}
