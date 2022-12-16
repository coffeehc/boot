package manage

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ping() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	}
}
