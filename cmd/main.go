package main

import (
	"go.uber.org/zap"
	"mesh.leinadium.dev/internal"
	"mesh.leinadium.dev/internal/api"
	"mesh.leinadium.dev/internal/logger"
	"mesh.leinadium.dev/internal/server"
	"mesh.leinadium.dev/internal/store"
	"mesh.leinadium.dev/internal/stream"
)

func main() {
	p, err := internal.NewParams()
	if err != nil {
		panic(err)
	}

	storeSvc := store.NewPGStore(store.PGParams{
		DSN:         p.DatabaseDSN,
		AutoMigrate: true,
	})

	streamSvc := stream.New(stream.Params{
		BrokerURL: p.StreamURL,
		Topic:     p.StreamTopic,
		ClientID:  p.StreamClientID,
		Store:     storeSvc,
	})

	serverSvc, err := server.New(server.Params{
		Store:    storeSvc,
		Streamer: storeSvc,
	})
	if err != nil {
		panic(err.Error())
	}

	apiSvc, err := api.New(api.Params{
		Server:       serverSvc,
		AllowOrigins: p.AllowOrigins,
	})
	if err != nil {
		panic(err.Error())
	}

	err = apiSvc.Run(p.Port)

	storeSvc.Stop()
	streamSvc.Stop()

	logger.Info("server closed", zap.Error(err))
}
