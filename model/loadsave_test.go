package model

import (
	"reflect"
	"testing"

	"github.com/matryer/is"
)

// var byteFixture []byte = []byte(`ID,Name,MinCapacity,MaxCapacity
// 1,foo,5,25`)
// var byteFixture []byte = []byte(`1,foo,5,25
// 2,bar,0,9000
// `)

func TestMarshalCourseIsRoundTripConsistent(t *testing.T) {
	is := is.New(t)

	coursesInput := []Course{
		{ID: 1, Name: "foo", MinCapacity: 5, MaxCapacity: 25},
		{ID: 2, Name: "bar", MinCapacity: 0, MaxCapacity: 9000},
	}
	participants := []Participant{}

	bytes, err := toExcelBytes(coursesInput, participants)
	is.NoErr(err)

	coursesOutput, _, err := fromExcelBytes(bytes)

	is.Equal(len(coursesInput), len(coursesOutput))

	for i := 0; i < len(coursesInput); i++ {
		is.True(reflect.DeepEqual(coursesInput[i], coursesOutput[i]))
	}
}

