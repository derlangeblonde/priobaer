package model

import (
	"reflect"
	"testing"

	"github.com/matryer/is"
)

func TestMarshalCourseIsRoundTripConsistent(t *testing.T) {
	is := is.New(t)

	coursesInput := []Course{
		{ID: 1, Name: "foo", MinCapacity: 5, MaxCapacity: 25},
		{ID: 2, Name: "bar", MinCapacity: 0, MaxCapacity: 9000},
	}
	participantsInput := []Participant{
		{ID: 1, Prename: "mady", Surname: "morison"},
		{ID: 2, Prename: "Breathe", Surname: "Flow"},
	}

	bytes, err := ToExcelBytes(coursesInput, participantsInput)
	is.NoErr(err)

	coursesOutput, participantsOutput, err := FromExcelBytes(bytes)

	is.Equal(len(coursesInput), len(coursesOutput)) // count of courses same after marshal-roundtrip

	for i := 0; i < len(coursesInput); i++ {
		is.True(reflect.DeepEqual(coursesInput[i], coursesOutput[i]))
	}

	is.Equal(len(participantsInput), len(participantsOutput)) // count of participants same after marshal-roundtrip

	for i := 0; i < len(participantsInput); i++ {
		is.True(reflect.DeepEqual(participantsInput[i], participantsOutput[i]))
	}
}
