package server

import (
	"context"
	"fmt"
	"time"

	"mesh.leinadium.dev/internal/store"
)

type Store interface {
	Nodes(ctx context.Context, from time.Time) ([]*store.Node, error)
	Node(context.Context, store.NodeID) (*store.Node, error)
}

type Streamer interface {
	Stream() chan store.Event
	CancelStream(chan store.Event)
}

type Params struct {
	Store    Store
	Streamer Streamer
}

type Server struct {
	store    Store
	streamer Streamer
}

func New(p Params) (*Server, error) {
	if p.Store == nil {
		return nil, fmt.Errorf("store cannot be nil")
	}
	if p.Streamer == nil {
		return nil, fmt.Errorf("streamer cannot be nil")
	}

	server := &Server{
		store:    p.Store,
		streamer: p.Streamer,
	}

	return server, nil
}

func (s *Server) Nodes(ctx context.Context, from time.Time) ([]*store.Node, error) {
	return s.store.Nodes(ctx, from)
}
func (s *Server) Node(ctx context.Context, id store.NodeID) (*store.Node, error) {
	return s.store.Node(ctx, id)

}
func (s *Server) Stream() chan store.Event {
	return s.streamer.Stream()
}

func (s *Server) CancelStream(ch chan store.Event) {
	s.streamer.CancelStream(ch)
}
