package controller

import (
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
	router.GET("/courses/button-new", CoursesButtonNew)

	router.GET("/participants/new", ParticipantsNew)
	router.GET("/participants", ParticipantsIndex)
	router.POST("/participants", ParticipantsCreate)
	router.DELETE("/participants/:id", ParticipantsDelete)
	router.GET("/participants/button-new", ParticipantsButtonNew)

	router.PUT("/assignments", AssignmentsUpdate)
	router.GET("/assignments", AssignmentsIndex)

	router.GET("/save", Save)
}
