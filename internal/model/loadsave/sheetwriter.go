package loadsave

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

type sheetWriter struct {
	file       *excelize.File
	sheetName  string
	currentRow int
}

func newSheetWriter(file *excelize.File, sheetName string) (*sheetWriter, error) {
	index, err := file.GetSheetIndex(sheetName)
	if err != nil {
		return nil, err
	}
	if index == -1 {
		if _, err = file.NewSheet(sheetName); err != nil {
			return nil, err
		}
	}

	return &sheetWriter{
		file:       file,
		sheetName:  sheetName,
		currentRow: 1,
	}, nil
}

func (sw *sheetWriter) write(row []string) error {
	for colIndex, value := range row {
		cell, err := excelize.CoordinatesToCellName(colIndex+1, sw.currentRow)
		if err != nil {
			return fmt.Errorf("failed to get cell name: %w", err)
		}
		if err := sw.file.SetCellValue(sw.sheetName, cell, value); err != nil {
			return fmt.Errorf("failed to set cell value: %w", err)
		}
	}
	sw.currentRow++
	return nil
}
