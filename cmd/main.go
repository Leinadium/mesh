package main

import (
	"fmt"

	"mesh.leinadium.dev/internal"
	"mesh.leinadium.dev/internal/api"
	"mesh.leinadium.dev/internal/server"
	"mesh.leinadium.dev/internal/store"
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

	fmt.Printf("server closed with err=%v\n", err)
}
