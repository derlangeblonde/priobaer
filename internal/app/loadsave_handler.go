package app

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"softbaer.dev/ass/internal/crypt"
	"softbaer.dev/ass/internal/domain"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/model/loadsave"
)

func LoadDialog(c *gin.Context) {
	c.HTML(200, "dialogs/load", nil)
}

func Load(c *gin.Context) {
	db := GetDB(c)
	secret := crypt.GetSecret(c)

	formFile, err := c.FormFile("file")

	if err != nil {
		slog.Error("Could not get uploaded form file", "err", err)
		c.AbortWithError(500, err)

		return
	}

	file, err := formFile.Open()

	if err != nil {
		slog.Error("Could not open formFile", "err", err)
		c.AbortWithError(500, err)
		return
	}

	scenario, err := loadsave.LoadScenarioFromExcelFile(file)

	if err != nil {
		slog.Error("Could not unmarshal models from excel-file", "err", err)
		c.Header("HX-Retarget", "body")
		c.Header("HX-Reswap", "beforeend")
		err := fmt.Errorf("Excel-Datei konnte nicht geladen werden.\n%w", err)

		stackedErrs := strings.Split(err.Error(), "\n")
		c.HTML(422, "dialogs/validation-error", stackedErrs)
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		return domain.OverwriteScenario(tx, scenario, secret)
	})

	if err != nil {
		slog.Error("Error while inserting unmarshalled structs into db", "err", err)
		c.AbortWithError(500, err)

		return
	}

	c.Redirect(http.StatusSeeOther, "/scenario")
}

func Save(c *gin.Context) {
	db := GetDB(c)
	secret := crypt.GetSecret(c)

	scenario, err := domain.LoadScenario(db, secret)

	if err != nil {
		slog.Error("Error while loading scenario", "err", err)
		c.AbortWithError(500, err)

		return
	}

	excelBytes, err := loadsave.SaveScenarioToExcelFile(scenario)

	if err != nil {
		slog.Error("Error while exporting models to ExcelBytes", "err", err)
		c.AbortWithError(500, err)

		return
	}

	extraHeaders := map[string]string{
		"Content-Disposition": `attachment; filename="export.xlsx"`,
	}

	c.DataFromReader(200, int64(len(excelBytes)), "application/octet-stream", bytes.NewReader(excelBytes), extraHeaders)
}
