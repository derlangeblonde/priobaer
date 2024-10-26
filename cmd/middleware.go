package cmd

import (
	"errors"
	"net/http"
	"slices"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const UserIdSessionKey = "userId"

func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	whitelist := []string{
		"/favicon.png",
		"/favicon.ico",
		"/health",
		"/login",
		"/register",
	}

	return func(c *gin.Context) {
		if slices.Contains(whitelist, c.FullPath()) {
			c.Next()

			return
		}

		userId, err := getUserIdFromSession(c)

		if err != nil {
			c.Redirect(http.StatusSeeOther, "/login")

			return
		}

		user := User{ID: userId}
		result := db.First(&user)

		if result.Error != nil {
			c.Redirect(http.StatusSeeOther, "/login")

			return
		}

		c.Next()
	}

}

func getUserIdFromSession(c *gin.Context) (int, error) {
	session := sessions.Default(c)

	userIdRaw := session.Get(UserIdSessionKey)

	if userIdRaw == nil {
		return -1, errors.New("UserId not set")
	}

	userId, err := strconv.Atoi(userIdRaw.(string))

	return userId, err
}
