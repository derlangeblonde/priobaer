package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jonboulle/clockwork"
	"softbaer.dev/ass/controller"
	"softbaer.dev/ass/dbdir"
	"softbaer.dev/ass/model"
	"softbaer.dev/ass/view"
)

func Run(ctx context.Context, getenv func(string) string, clock clockwork.Clock) error {
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

	cookieStore := cookie.NewStore([]byte(config.Secret))
	cookieStore.Options(
		sessions.Options{
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   sessionMaxAgeSeconds,
		},
	)

	dbDirectory, err := dbdir.New(config.DbRootDir, config.SessionMaxAge, clock, []any{&model.Course{}, &model.Participant{}, &model.Priority{}})

	if err != nil {
		panic(err)
	}

	router.Use(sessions.Sessions("session", cookieStore))
	router.Use(controller.InjectDB(dbDirectory))

	router.SetHTMLTemplate(templates)

	controller.RegisterRoutes(router, dbDirectory)

	server := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", config.Port),
		Handler: router.Handler(),
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe returned an errore", "err", err)
		}
	}()

	<-ctx.Done()
	server.Shutdown(ctx)
	errs := dbDirectory.Close()

	if len(errs) != 0 {
		slog.Error("Could not close all dbs", "errs", errs)
	}

	return nil
}
