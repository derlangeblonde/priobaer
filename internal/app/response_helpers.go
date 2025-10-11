package app

import "github.com/gin-gonic/gin"

func emptyBadRequestResponse(c *gin.Context) {
	c.Data(400, "text/html", []byte(""))
}

func internalServerErrorResponse(c *gin.Context) {
	c.HTML(500, "general/500", gin.H{})
}
