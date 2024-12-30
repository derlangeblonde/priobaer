package model

import (
	"strconv"

	"github.com/google/uuid"
	"golang.org/x/exp/rand"
)

var namespace uuid.UUID = uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
var intSeed uint64 = 69420
var SeededRand *rand.Rand = rand.New(rand.NewSource(intSeed))

func SeededUUID() uuid.UUID {
	oneTimeSeedStr := strconv.Itoa(SeededRand.Int())
	return uuid.NewMD5(namespace, []byte(oneTimeSeedStr))
}

func RandomCourse() Course {
	name := SeededUUID()

	minCap := SeededRand.Intn(30)
	maxCap := minCap + SeededRand.Intn(30)

	return Course{Name: name.String(), MinCapacity: minCap, MaxCapacity: maxCap}
}

func RandomNameCourse(minCap, maxCap int) Course {
	name := SeededUUID()

	return Course{Name: name.String(), MinCapacity: minCap, MaxCapacity: maxCap}
}

func RandomParticipant() Participant {
	prename := SeededUUID()
	surname := SeededUUID()

	return Participant{Prename: prename.String(), Surname: surname.String()}
}

func RandomCourses(n int) (result []Course) {
	for i := 0; i < n; i++ {
		result = append(result, RandomCourse())
	}

	return
}

func RandomParticipants(n int) (result []Participant) {
	for i := 0; i < n; i++ {
		result = append(result, RandomParticipant())
	}

	return
}
