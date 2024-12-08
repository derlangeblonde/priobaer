package cmdtest

import (
	"math/rand/v2"

	"github.com/google/uuid"
	"softbaer.dev/ass/model"
)

func RandomCourse() model.Course {
	name := uuid.New()

	minCap := rand.IntN(30)
	maxCap := minCap + rand.IntN(30)

	return model.Course{Name: name.String(), MinCapacity: minCap, MaxCapacity: maxCap}
}
