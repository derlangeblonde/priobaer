package app

import "github.com/gin-gonic/gin"

func Help(c *gin.Context) {
	c.HTML(200, "general/help", gin.H{})
}

func DataProtection(c *gin.Context) {
	c.HTML(200, "general/data-protection", gin.H{})
}
