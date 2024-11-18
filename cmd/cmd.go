package cmd

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/view"
)

//go:embed favicon.ico
var faviconBytes []byte

func Run(ctx context.Context, getenv func(string) string) error {
	router := gin.Default()

	templates, err := view.LoadTemplate()

	if err != nil {
		panic(err)
	}

	config, err := ParseConfig(getenv)

	if err != nil {
		panic(fmt.Sprintf("Could not parse config from env, Err: %v. Panic...", err))
	}

	sessionMaxAgeSeconds := int(config.SessionMaxAge.Seconds())

	if sessionMaxAgeSeconds == 0 {
		sessionMaxAgeSeconds = 1
	}

	// TODO: (Prod) read secret from file
	cookieStore := cookie.NewStore([]byte("secret"))
	cookieStore.Options(
		sessions.Options{
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   sessionMaxAgeSeconds,
		},
	)

	dbManager := NewDbManager(config.DbRootDir, config.SessionMaxAge)
	err = dbManager.ReadExistingSessions()

	if err != nil {
		panic(err)
	}

	router.Use(sessions.Sessions("session", cookieStore))
	router.Use(InjectDB(dbManager))

	router.SetHTMLTemplate(templates)

	RegisterRoutes(router)

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: router.Handler(),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServed errored", "err", err)
		}
	}()

	<-ctx.Done()
	server.Shutdown(ctx)
	errs := dbManager.TryCloseAllDbs()

	if len(errs) != 0 {
		slog.Error("Could not close all dbs", "errs", errs)
	}

	return nil
}

func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Fprintf(c.Writer, "OK")
	}
}

func FaviconHandler(c *gin.Context) {
	c.Data(http.StatusOK, "image/x-icon", faviconBytes)
}

type Course struct {
	gorm.Model
	ID int
	// TODO: unique constraint does not go well with soft delete
	Name        string `gorm:"unique"`
	MaxCapacity int
	MinCapacity int
}

func LandingPage(c *gin.Context) {
	fmt.Fprintf(c.Writer, "This is the landing page!")
}

func CoursesIndex() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := GetDB(c)

		courses := make([]Course, 0)
		result := db.Find(&courses)

		if result.Error != nil {
			slog.Error("Unexpected error while showing course index", "err", result.Error)
			c.AbortWithStatus(http.StatusInternalServerError)

			return
		}

		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusOK, "courses/index", gin.H{"fullPage": false, "courses": courses})
		} else {
			c.HTML(http.StatusOK, "courses/index", gin.H{"fullPage": true, "courses": courses})
		}
	}
}

func CoursesNew() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "courses/new", nil)
	}
}

func CoursesCreate() gin.HandlerFunc {
	type request struct {
		Name        string `form:"name" binding:"required"`
		MaxCapacity int    `form:"max-capacity" binding:"required"`
		MinCapacity int    `form:"min-capacity" binding:"required"`
	}

	return func(c *gin.Context) {
		db := GetDB(c)

		var req request
		err := c.Bind(&req)

		if err != nil {
			return
		}

		course := Course{Name: req.Name, MaxCapacity: req.MaxCapacity, MinCapacity: req.MinCapacity}
		result := db.Create(&course)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrCheckConstraintViolated) {
				slog.Error("Constraint violated while creating Course", "err", err, "course.Name", course.Name)
				c.AbortWithStatus(http.StatusConflict)

				return
			}

			slog.Error("Unexpected error while creating Course", "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)

			return
		}

		c.Redirect(http.StatusSeeOther, "/courses")
	}
}

func CoursesDelete() gin.HandlerFunc {
	type request struct {
		ID int `uri:"id" binding:"required"`
	}
	return func(c *gin.Context) {
		db := GetDB(c)

		var req request
		err := c.BindUri(&req)

		if err != nil {
			slog.Error("Could not parse id from uri in CoursesDelete", "err", err)
			c.AbortWithStatus(http.StatusNotFound)

			return
		}

		course := Course{ID: req.ID}
		result := db.Delete(&course)

		if result.Error != nil {
			slog.Error("Delete of course failed on db level", "err", result.Error)
			c.AbortWithStatus(http.StatusInternalServerError)

			return
		}

		c.Data(http.StatusOK, "text/html", []byte(""))
	}
}
