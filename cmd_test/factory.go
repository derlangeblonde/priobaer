package cmdtest

import (
	"math/rand/v2"

	"github.com/google/uuid"
	"softbaer.dev/ass/cmd"
)

func RandomCourse() cmd.Course {
	name := uuid.New()

	minCap := rand.IntN(30)
	maxCap := minCap + rand.IntN(30)

	return cmd.Course{Name: name.String(), MinCapacity: minCap, MaxCapacity: maxCap}
}
