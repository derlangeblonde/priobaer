package model

import (
	"testing"

	"github.com/matryer/is"
)

// var byteFixture []byte = []byte(`ID,Name,MinCapacity,MaxCapacity
// 1,foo,5,25`)
var byteFixture []byte = []byte(`1,foo,5,25
2,bar,0,9000
`)

func TestToCsvBytes_WritesCoursesCorrectly(t *testing.T) {
	is := is.New(t)

	courses := []Course{
		{ID: 1, Name: "foo", MinCapacity: 5, MaxCapacity: 25},
		{ID: 2, Name: "bar", MinCapacity: 0, MaxCapacity: 9000},
	}
	participants := []Participant{}

	actualBytes, err := toCsvBytes(courses, participants)
	is.NoErr(err)

	is.Equal(len(actualBytes), len(byteFixture))
	is.Equal(actualBytes, byteFixture)
}

func TestFromCsvBytes_ReadsCoursesCorrectly(t *testing.T) {
	is := is.New(t)

	courses, _, err := fromCsvBytes(byteFixture)
	is.NoErr(err)

	is.Equal(len(courses), 2)
}
