package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _ = fmt.Fprintf(c.Writer, "OK")
	}
}
