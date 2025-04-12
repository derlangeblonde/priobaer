package app

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DbError logs the error and renders a generic error dialog.
// Variadic arguments are key-value pairs that are logged.
// Therefore args must be an even number of strings.
func DbError(c *gin.Context, err error, function string, args ...string) {
    logArgs := []any{"function", function, "err", err}
    
    for i := 0; i < len(args); i += 2 {
        if i+1 < len(args) {
            logArgs = append(logArgs, args[i], args[i+1])
        } else {
            logArgs = append(logArgs, args[i], "(missing value)")
        }
    }

    slog.Error("DB-Operation Failed", logArgs...)
    c.HTML(http.StatusInternalServerError, "dialogs/generic-error", err)
}

