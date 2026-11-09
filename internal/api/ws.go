package api

import (
	"errors"
	"net/http"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"mesh.leinadium.dev/internal/logger"
)

func (a *API) handleWebsocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("websocket upgrade", zap.Error(err))
		return
	}
	defer conn.Close()

	ch := a.s.Stream()
	defer a.s.CancelStream(ch)

	for {
		select {
		case <-c.Done():
			return
		case event := <-ch:
			if err := conn.WriteJSON(event); err != nil {
				if !errors.Is(err, syscall.EPIPE) {
					logger.Warn("websocket write", zap.Error(err))
				}
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
