package controller

import (
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"softbaer.dev/ass/dbdir"
)

func RegisterRoutes(router *gin.Engine, dbDirectory *dbdir.DbDirectory) {
	router.GET("/health", HealthHandler())
	router.GET("/index", LandingPage)

	router.Static("/static", "./static")

	router.GET("/favicon.png", FaviconHandler)

	router.GET("/courses/new", CoursesNew())
	router.POST("/courses", CoursesCreate())
	router.DELETE("/courses/:id", CoursesDelete())
	router.GET("/courses/button-new", CoursesButtonNew)

	router.GET("/participants/new", ParticipantsNew)
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

	router.POST("/prio", func (c *gin.Context){ 
	body, _ := io.ReadAll(c.Request.Body)
		fmt.Println(string(body))
	})
}
