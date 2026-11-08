package api

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"mesh.leinadium.dev/internal/store"
)

type Server interface {
	Nodes(ctx context.Context, from time.Time) ([]*store.Node, error)
	Node(context.Context, store.NodeID) (*store.Node, error)
	Stream() chan store.Event
	CancelStream(chan store.Event)
}

type Params struct {
	Server       Server
	AllowOrigins []string
}

type API struct {
	e *gin.Engine
	s Server
}

func New(p Params) (*API, error) {
	if p.Server == nil {
		return nil, fmt.Errorf("server cannot be nil")
	}

	engine := gin.Default()
	engine.RedirectTrailingSlash = false
	engine.RemoveExtraSlash = false
	_ = engine.SetTrustedProxies(nil)

	if len(p.AllowOrigins) > 0 {
		engine.Use(cors.New(cors.Config{
			AllowOrigins:     p.AllowOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowHeaders:     []string{"Content-Type"},
			AllowCredentials: true,
		}))
	}

	api := &API{
		e: engine,
		s: p.Server,
	}

	// websocket
	engine.GET("/ws", api.handleWebsocket)

	// v1 endpoints
	v1 := engine.Group("/v1")
	v1.GET("/node", api.getNodes)
	v1.GET("/node/:node", api.getNode)

	return api, nil
}

func (a *API) Run(port int) error {
	return a.e.Run(fmt.Sprintf(":%d", port))
}
