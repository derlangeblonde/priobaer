package controller

import (
	"fmt"
	"net/http"
	_ "embed"

	"github.com/gin-gonic/gin"
)

//go:embed favicon.ico
var faviconBytes []byte

func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Fprintf(c.Writer, "OK")
	}
}

func FaviconHandler(c *gin.Context) {
	c.Data(http.StatusOK, "image/x-icon", faviconBytes)
}


func LandingPage(c *gin.Context) {
	fmt.Fprintf(c.Writer, "This is the landing page!")
}
