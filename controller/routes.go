package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	router.GET("/health", HealthHandler())
	router.GET("/index", LandingPage)

	router.Static("/static", "./static")

	router.GET("/favicon.png", FaviconHandler)

	router.GET("/courses/new", CoursesNew())
	router.GET("/courses", CoursesIndex())
	router.POST("/courses", CoursesCreate())
	router.DELETE("/courses/:id", CoursesDelete())

	router.GET("/participants/new", ParticipantsNew)
	router.GET("/participants", ParticipantsIndex)
	router.POST("/participants", ParticipantsCreate)
	router.DELETE("/participants/:id", ParticipantsDelete)

	router.GET("/assignments", func(c *gin.Context) {
		fmt.Fprint(c.Writer, "OK")
	})
}
