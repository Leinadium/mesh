package event

import (
	"context"
	"errors"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
	"mesh.leinadium.dev/internal/logger"
	"mesh.leinadium.dev/internal/store"
)

type Store interface {
	NewEvent(context.Context, store.Event) error
}

type Handler struct {
	p Params

	client  mqtt.Client
	decoder *decoder
	alive   bool
	cancel  context.CancelFunc
}

type Params struct {
	ClientID string
	Host     string
	Topic    string

	Store Store
}

func New(p Params) *Handler {
	return &Handler{
		p: p,
	}
}

func (h *Handler) Start(ctx context.Context) error {
	if h.p.Host == "" {
		return errors.New("event host is not configured")
	}

	if h.p.Topic == "" {
		h.p.Topic = "#"
	}

	if h.p.ClientID == "" {
		h.p.ClientID = generateClientID()
	}

	opts := mqtt.NewClientOptions().
		AddBroker(h.p.Host).
		SetClientID(h.p.ClientID).
		SetDefaultPublishHandler(h.handle).
		SetConnectRetry(true).
		SetOnConnectHandler(h.onConnect).
		SetConnectionLostHandler(h.onConnectionLost)

	h.client = mqtt.NewClient(opts)

	// connecting
	connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := wait(connectCtx, h.client.Connect()); err != nil {
		return err
	}

	// subscribe
	subscribeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := wait(subscribeCtx, h.client.Subscribe(h.p.Topic, 1, nil)); err != nil {
		return err
	}

	return nil
}

func (h *Handler) handle(c mqtt.Client, msg mqtt.Message) {
	ctx := context.Background()

	event, err := FromBytes(msg.Payload())
	if err != nil {
		logger.Error("could not handle event", zap.Error(err))
		return
	}
	if err := h.p.Store.NewEvent(ctx, event); err != nil {
		logger.Error("could not store event", zap.Error(err))
	}
}

func (h *Handler) Alive() bool {
	return h.alive
}

func (h *Handler) onConnect(c mqtt.Client) {
	logger.Info("event handler connected")
	h.alive = true
}

func (h *Handler) onConnectionLost(c mqtt.Client, err error) {
	logger.Info("event handler disconnected")
	h.alive = false
}

func (h *Handler) Stop(ctx context.Context) error {
	if h.client != nil {
		logger.Info("event handler stopping")
		h.client.Unsubscribe(h.p.Topic)
		h.client.Disconnect(100)
	}
	return nil
}

func wait(ctx context.Context, token mqtt.Token) error {
	select {
	case <-token.Done():
		if err := token.Error(); err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func generateClientID() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return "mesh-event-reader-" + string(b)
}
