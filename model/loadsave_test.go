package model

import (
	"fmt"
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

	bytes, err := toCsvBytes(coursesInput, participants)
	is.NoErr(err)

	coursesOutput, _, err := fromCsvBytes(bytes)

	is.Equal(len(coursesInput), len(coursesOutput))

	for i := 0; i < len(coursesInput); i++ {
		err := coursesInput[i].ShallowEqual(coursesOutput[i])
		is.NoErr(err)
	}
}

func (c *Course) ShallowEqual(other Course) error {
	if c.ID != other.ID {
		return fmt.Errorf("IDs: %d != %d", c.ID, other.ID)
	}

	if c.MinCapacity!= other.MinCapacity {
		return fmt.Errorf("MinCapacities: %d != %d", c.MinCapacity, other.MinCapacity)
	}

	if c.MaxCapacity!= other.MaxCapacity {
		return fmt.Errorf("MaxCapacities: %d != %d", c.MaxCapacity, other.MaxCapacity)
	}

	if c.Name != other.Name {
		return fmt.Errorf("Names: %s != %s", c.Name, other.Name)
	}

	return nil
}
