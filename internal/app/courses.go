package app

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model"
)

func CoursesNew() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "courses/new", gin.H{"Errors": make(map[string]string)})
	}
}

func CoursesCreate() gin.HandlerFunc {
	type request struct {
		Name        string `form:"name"`
		MaxCapacity *int   `form:"max-capacity" binding:"required"`
		MinCapacity *int   `form:"min-capacity" binding:"required"`
	}

	return func(c *gin.Context) {
		db := GetDB(c)

		var req request
		err := c.Bind(&req)

		if err != nil {
			slog.Error("Could not bind request to when creating course", "err", err)
			return
		}

		course := model.Course{Name: req.Name, MaxCapacity: *req.MaxCapacity, MinCapacity: *req.MinCapacity}
		validationErrors := course.Valid()

		if len(validationErrors) > 0 {
			c.HTML(422, "courses/new", gin.H{"Errors": validationErrors, "Value": course})

			return
		}

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

		viewCourse := toViewCourse(course, sql.NullInt64{Valid: false}, false)

		if c.GetHeader("HX-Request") == "true" {
			c.HTML(http.StatusOK, "courses/_show-with-new-button", viewCourse)
		} else {
			c.Redirect(http.StatusSeeOther, "/scenario")
		}
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

		course := model.Course{ID: req.ID}
		err = db.Transaction(func(tx *gorm.DB) error {
			err := db.Model(model.Participant{}).Where("course_id = ?", req.ID).Update("course_id", nil).Error

			if err != nil {
				return err
			}

			return db.Unscoped().Delete(&course).Error
		})

		if err != nil {
			slog.Error("Delete of course failed on db level", "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)

			return
		}

		c.Data(http.StatusOK, "text/html", []byte(""))
	}
}

func CoursesButtonNew(c *gin.Context) {
	c.HTML(http.StatusOK, "courses/_new-button", nil)
}
