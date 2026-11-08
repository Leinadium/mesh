package stream

import (
	"context"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
	"mesh.leinadium.dev/internal/event"
	"mesh.leinadium.dev/internal/logger"
	"mesh.leinadium.dev/internal/store"
)

const (
	loopReconnect = 10 * time.Second
	endDisconnect = 250 // milisseconds
)

type Store interface {
	NewEvent(ctx context.Context, event store.Event) error
}

type Params struct {
	BrokerURL string
	Topic     string
	ClientID  string

	Store Store
}

type Handler struct {
	o *mqtt.ClientOptions
	c mqtt.Client
	p Params
}

func New(p Params) *Handler {
	if p.Store == nil {
		panic("nil store in stream handler")
	}
	opts := mqtt.NewClientOptions()

	h := &Handler{p: p}

	opts.AddBroker(p.BrokerURL)
	opts.SetClientID(p.ClientID)
	opts.SetDefaultPublishHandler(h.handleGen())
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.AutoReconnect = true
	opts.ConnectRetry = true
	opts.ConnectRetryInterval = 1 * time.Minute

	h.o = opts
	go h.connect()
	return h
}

func (h *Handler) connect() {
	h.c = mqtt.NewClient(h.o)
	for {
		logger.Info("stream: trying to connect")
		token := h.c.Connect()
		token.Wait()
		if token.Error() != nil {
			logger.Error("stream: could not connect")
			time.Sleep(loopReconnect)
		} else {
			logger.Info("stream: connected")
			break
		}
	}

	token := h.c.Subscribe(h.p.Topic, 1, nil)
	token.Wait()
	logger.Info("stream: subscribed", zap.String("topic", h.p.Topic))
}

func (h *Handler) Stop() {
	if h.c != nil {
		h.c.Unsubscribe(h.p.Topic)
		h.c.Disconnect(endDisconnect)
	}
}

func (h *Handler) handleGen() func(mqtt.Client, mqtt.Message) {
	return func(client mqtt.Client, msg mqtt.Message) {
		e, err := event.FromBytes(msg.Payload())
		if err != nil {
			logger.Warn("stream: could not decode event", zap.String("type", string(e.Type())), zap.Error(err))
			return
		}
		logger.Info("stream: new event", zap.String("type", string(e.Type())))
		if err := h.p.Store.NewEvent(context.Background(), e); err != nil {
			logger.Error("stream: could not store event", zap.Error(err))
		}
	}
}

var connectHandler mqtt.OnConnectHandler = func(_ mqtt.Client) {
	logger.Info("stream: connected to mqtt broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(_ mqtt.Client, err error) {
	logger.Error("stream: connection lost", zap.Error(err))
}
