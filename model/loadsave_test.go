package model

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/matryer/is"
)

func TestMarshalCourseIsRoundTripConsistent(t *testing.T) {
	is := is.New(t)

	testcases := []struct {
		coursesInput      []Course
		participantsInput []Participant
	}{
		{
			[]Course{
				{ID: 1, Name: "foo", MinCapacity: 5, MaxCapacity: 25},
				{ID: 2, Name: "bar", MinCapacity: 0, MaxCapacity: 9000},
			},
			[]Participant{
				{ID: 1, Prename: "mady", Surname: "morison"},
				{ID: 2, Prename: "Breathe", Surname: "Flow"},
			},
		},
		{
			[]Course{},
			[]Participant{{ID: 143920, Prename: "we have", Surname: "no courses"}},
		},
		{
			[]Course{{ID: 42904, Name: "I am" , MinCapacity: 741, MaxCapacity: 4920}},
			[]Participant{},
		},
		{
			[]Course{},
			[]Participant{},
		},
	}

	for _, tc := range testcases {
		excelBytes, err := ToExcelBytes(tc.coursesInput, tc.participantsInput)
		is.NoErr(err) // err while marshall to excel

		coursesOutput, participantsOutput, err := FromExcelBytes(bytes.NewReader(excelBytes))
		is.NoErr(err) // err while unmarshalling from excel

		is.Equal(len(tc.coursesInput), len(coursesOutput)) // count of courses same after marshal-roundtrip

		for i := 0; i < len(tc.coursesInput); i++ {
			is.True(reflect.DeepEqual(tc.coursesInput[i], coursesOutput[i]))
		}

		is.Equal(len(tc.participantsInput), len(participantsOutput)) // count of participants same after marshal-roundtrip

		for i := 0; i < len(tc.participantsInput); i++ {
			if !reflect.DeepEqual(tc.participantsInput[i], participantsOutput[i]) {
				t.Fatalf("Participant not equal. Got=%v, Want=%v", participantsOutput[i], tc.participantsInput[i])
			}
		}
	}
}
