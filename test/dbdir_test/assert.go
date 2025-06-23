package dbdirtest

import (
	"github.com/matryer/is"
	"gorm.io/gorm"
)

func connHasNoRows(conn *gorm.DB, is *is.I) {
	var datas []testData
	result := conn.Find(&datas)
	is.NoErr(result.Error)

	is.Equal(len(datas), 0)
}
