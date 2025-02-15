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

type CourseOption func(*Course)
type ParticipantOption func(*Participant)

func RandomCourse(options ...CourseOption) Course {
	name := SeededUUID()

	minCap := SeededRand.Intn(30)
	maxCap := minCap + SeededRand.Intn(30)

	c := Course{Name: name.String(), MinCapacity: minCap, MaxCapacity: maxCap}

	for _, option := range options {
		option(&c)
	}

	return c
}

func WithCourseId(id int) CourseOption {
	return func(c *Course) {
		c.ID = id
	}
}

func WithCapacity(min, max int) CourseOption {
	return func(c *Course) {
		c.MinCapacity = min
		c.MaxCapacity = max
	}
}

func RandomParticipant(options ...ParticipantOption) Participant {
	prename := SeededUUID()
	surname := SeededUUID()

	p := Participant{Prename: prename.String(), Surname: surname.String()}

	for _, option := range options {
		option(&p)
	}

	return p
}

func WithParticipantId(id int) ParticipantOption {
	return func(p *Participant) {
		p.ID = id
	}
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
