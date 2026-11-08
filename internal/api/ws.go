package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (a *API) handleWebsocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v\n", err)
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
				log.Printf("websocket write error: %v\n", err)
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
