package app

import (
	"github.com/gin-gonic/gin"
	"softbaer.dev/ass/internal/app/staticfiles"
	"softbaer.dev/ass/internal/dbdir"
)

func RegisterRoutes(router *gin.Engine, dbDirectory *dbdir.DbDirectory) {
	router.GET("/health", HealthHandler())

	router.Static("/static", "./static")

	router.GET("/favicon.png", staticfiles.FaviconHandler)
	router.GET("/style.css", staticfiles.StyleCssHandler)
	router.GET("/index.js", staticfiles.IndexJsHandler)

	router.GET("/courses/new", CoursesNew())
	router.POST("/courses", CoursesCreate())
	router.DELETE("/courses/:id", CoursesDelete())
	router.GET("/courses/button-new", CoursesButtonNew)

	router.GET("/participants/new", ParticipantsNew)
	router.POST("/participants", ParticipantsCreate)
	router.DELETE("/participants/:id", ParticipantsDelete)
	router.GET("/participants/button-new", ParticipantsButtonNew)

	router.GET("/scenario", ScenarioIndex)

	router.POST("/participants/:id/assignments/:course-id", AssignmentsCreate)
	router.PUT("/participants/:id/assignments/:course-id", AssignmentsUpdate)
	router.DELETE("/participants/:id/assignments", AssignmentsDelete)

	router.PUT("/assignments", SolveAssignments)

	router.GET("/save", Save)
	router.GET("/load", LoadDialog)
	router.POST("/load", Load)

	router.GET("/sessions/new", SessionNew)
	router.POST("sessions", SessionCreate(dbDirectory))
}
