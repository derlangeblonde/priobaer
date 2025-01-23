package model

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

type SheetReader struct {
	file      *excelize.File
	sheetName string
	currentRow int
}

func NewSheetReader(file *excelize.File, sheetName string) (*SheetReader, error) {
	index, err := file.GetSheetIndex(sheetName)
	if err != nil {
		return nil, err
	}
	if index == -1 {
		file.NewSheet(sheetName)
	}

	return &SheetReader{
		file:      file,
		sheetName: sheetName,
		currentRow: 1,
	}, nil
}

// TODO: it might make sense to always trim?
// Then this should also be done on inserting data via REST-endpoints 
func (sr *SheetReader) Read() ([]string, error) {
	row, err := sr.file.GetRows(sr.sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from sheet: %w", err)
	}
	if sr.currentRow > len(row) {
		return nil, io.EOF 
	}
	result := row[sr.currentRow-1] 
	sr.currentRow++
	return result, nil
}
