package loadsave

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

type sheetReader struct {
	file       *excelize.File
	sheetName  string
	currentRow int
}

func newSheetReader(file *excelize.File, sheetName string) (*sheetReader, error) {
	index, err := file.GetSheetIndex(sheetName)
	if err != nil {
		return nil, err
	}
	if index == -1 {
		if _, err = file.NewSheet(sheetName); err != nil {
			return nil, err
		}
	}

	return &sheetReader{
		file:       file,
		sheetName:  sheetName,
		currentRow: 1,
	}, nil
}

func (sr *sheetReader) read() ([]string, error) {
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
