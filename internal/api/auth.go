package api

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const authPrefix = "Basic "

func (a *API) auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		if authorization == "" {
			a.abort(c, http.StatusBadRequest, errors.New("missing Authorization header"))
			return
		}
		token64, found := strings.CutPrefix(authorization, authPrefix)
		if !found {
			a.abort(c, http.StatusBadRequest, errors.New("malformed Authorization header. Basic is required"))
			return
		}

		tokenBin, err := base64.StdEncoding.DecodeString(token64)
		if err != nil {
			a.abort(c, http.StatusBadRequest, err)
			return
		}

		if string(tokenBin) != a.secret {
			a.abort(c, http.StatusUnauthorized, errors.New("invalid authorization"))
			return
		}
		c.Next()
	}
}
