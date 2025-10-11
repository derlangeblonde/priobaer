package staticfiles

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

// //go:embed favicon.ico
// var faviconBytes []byte
//
//	func FaviconHandler(c *gin.Context) {
//		c.Data(http.StatusOK, "image/x-icon", faviconBytes)
//	}
//
//go:embed style.css
var styleCss []byte

//go:embed index.js
var indexJs []byte

func StyleCssHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/css", styleCss)
}

func IndexJsHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/javascript", indexJs)
}
