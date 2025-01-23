package model

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

type SheetWriter struct {
	file       *excelize.File
	sheetName  string
	currentRow int
}

func NewSheetWriter(file *excelize.File, sheetName string) (*SheetWriter, error) {
	index, err := file.GetSheetIndex(sheetName)
	if err != nil {
		return nil, err
	}
	if index == -1 {
		file.NewSheet(sheetName)
	}

	return &SheetWriter{
		file:       file,
		sheetName:  sheetName,
		currentRow: 1,
	}, nil
}

func (sw *SheetWriter) Write(row []string) error {
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

