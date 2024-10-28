package cmd

import (
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"softbaer.dev/ass/view"
)

//go:embed favicon.ico
var faviconBytes []byte

func Run(path string) error {
	router := gin.Default()

	templates, err := view.LoadTemplate()

	if err != nil {
		panic(err)
	}

	// TODO: (Prod) read secret from file
	cookieStore := cookie.NewStore([]byte("secret"))
	cookieStore.Options(
		sessions.Options{
			Secure: true,
		},
	)

	dbManager := NewSessionDBMapper()

	router.Use(sessions.Sessions("session", cookieStore))
	router.Use(dbManager.InjectDB())

	router.SetHTMLTemplate(templates)

	router.GET("/health", HealthHandler())

	router.Static("/static", "./static")

	router.GET("/favicon.png", FaviconHandler)
	router.GET("/favicon.ico", FaviconHandler)

	router.GET("/register", UsersNew())
	// router.POST("/register", UsersCreate(db))

	router.GET("/login", SessionsNew())
	// router.POST("/login", SessionsCreate(db))

	router.GET("/index", IndexHandler())

	router.Run(":8080")

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

type User struct {
	gorm.Model
	ID           int
	Email        string `gorm:"unique"`
	PasswordHash string
}

func NewUser(email, password string) (User, error) {
	passwordHash, err := HashPassword(password)

	if err != nil {
		return User{}, err
	}

	return User{Email: email, PasswordHash: passwordHash}, nil
}

func UsersNew() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "users/new", nil)
	}
}

func UsersCreate(db *gorm.DB) gin.HandlerFunc {
	type request struct {
		Email    string `form:"email" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	return func(c *gin.Context) {
		var req request
		err := c.Bind(&req)

		if err != nil {
			return
		}

		user, err := NewUser(req.Email, req.Password)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)

			return
		}

		result := db.Create(&user)

		if result.Error != nil {
			c.AbortWithError(http.StatusConflict, result.Error)
		}

		c.Redirect(http.StatusFound, "/login")
	}
}

func SessionsNew() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "sessions/new", nil)
	}
}

func SessionsCreate(db *gorm.DB) gin.HandlerFunc {
	type request struct {
		Email    string `form:"email" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	return func(c *gin.Context) {
		var req request
		err := c.Bind(&req)

		if err != nil {
			return
		}

		var user User
		result := db.Where("email = ?", req.Email).Find(&user)

		if result.Error != nil {
			c.AbortWithError(http.StatusUnauthorized, errors.New("Password or Email wrong"))

			return
		}

		ok := CheckPasswordHash(req.Password, user.PasswordHash)

		if !ok {
			c.AbortWithError(http.StatusUnauthorized, errors.New("Password or Email wrong"))

			return
		}

		session := sessions.Default(c)

		session.Set("userId", strconv.Itoa(user.ID))
		err = session.Save()

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, errors.New("Something went wrong internally with the login process"))
		}

		c.Redirect(http.StatusFound, "/index")
	}
}

func IndexHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, ok := GetDb(c)
		if !ok {
			fmt.Fprintln(c.Writer, "not ok")

			return
		}

		var s Session
		conn.First(&s)

		fmt.Fprintln(c.Writer, s.SessionId)
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
