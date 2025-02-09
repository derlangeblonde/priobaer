package controller

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"softbaer.dev/ass/model"
	"softbaer.dev/ass/model/loadsave"
)

func Load(c *gin.Context) {
	const batchSize = 10
	db := GetDB(c)

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

	courses, participants, err := loadsave.FromExcelBytes(file)

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
		if err := db.CreateInBatches(&courses, batchSize).Error; err != nil {
			return err
		}

		if err := db.CreateInBatches(&participants, batchSize).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		slog.Error("Error while inserting unmarshalled structs into db", "err", err)
		c.AbortWithError(500, err)

		return
	}

	c.Redirect(http.StatusSeeOther, "/assignments")
}

func Save(c *gin.Context) {
	db := GetDB(c)

	var participants []model.Participant
	var courses []model.Course

	if err := db.Find(&participants).Error; err != nil {
		slog.Error("Error while fetching participants from db", "err", err)
		c.AbortWithError(500, err)

		return
	}

	if err := db.Find(&courses).Error; err != nil {
		slog.Error("Error while fetching course from db", "err", err)
		c.AbortWithError(500, err)

		return
	}

	excelBytes, err := loadsave.ToExcelBytes(courses, participants)

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
