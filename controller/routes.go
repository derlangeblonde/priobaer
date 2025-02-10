package controller

import (
	"github.com/gin-gonic/gin"
	"softbaer.dev/ass/dbdir"
)

func RegisterRoutes(router *gin.Engine, dbDirectory *dbdir.DbDirectory) {
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
	router.GET("/load", LoadDialog)
	router.POST("/load", Load)

	router.GET("/sessions/new", SessionNew)
	router.POST("sessions", SessionCreate(dbDirectory))
}
