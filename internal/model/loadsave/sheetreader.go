package loadsave

import (
	"io"

	"github.com/xuri/excelize/v2"
)

type sheetReader struct {
	file       *excelize.File
	sheetName  string
	currentRow int
	cells      [][]string
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

	cells, err := file.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	return &sheetReader{
		file:       file,
		sheetName:  sheetName,
		currentRow: 1,
		cells:      cells,
	}, nil
}

func (sr *sheetReader) read() ([]string, error) {
	if sr.currentRow > len(sr.cells) {
		return nil, io.EOF
	}
	result := sr.cells[sr.currentRow-1]
	sr.currentRow++
	return result, nil
}
