package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mesh.leinadium.dev/internal/logger"
	"mesh.leinadium.dev/internal/store"
)

const (
	timeFormat  = "2006-01-02T15:04:05"
	defaultFrom = 3 * time.Hour
)

func (a *API) getNodes(c *gin.Context) {
	from, ok := timeFromStr(c.Query("from"))
	if !ok {
		from = time.Now().Add(-defaultFrom)
	}

	nodes, err := a.s.Nodes(c.Request.Context(), from)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}

func (a *API) getNode(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		a.error(c, http.StatusBadRequest, err)
		return
	}

	node, err := a.s.Node(c.Request.Context(), store.NodeID(id))
	if err != nil {
		a.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"node": node})
}

func (a *API) error(c *gin.Context, status int, err error) {
	logger.Error("api error", zap.Error(err))
	c.JSON(status, gin.H{"error": err.Error()})
}

func (a *API) abort(c *gin.Context, status int, err error) {
	a.error(c, status, err)
	c.Abort()
}

func timeFromStr(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	ret, err := time.Parse(timeFormat, s)
	return ret, err == nil
}
