package model

import (
	"softbaer.dev/ass/internal/seededuuid"
)

type CourseOption func(*Course)

func RandomCourse(options ...CourseOption) Course {
	name := seededuuid.SeededUUID()

	minCap := seededuuid.SeededRand.Intn(30)
	maxCap := minCap + seededuuid.SeededRand.Intn(30)

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

func WithCourseName(name string) CourseOption {
	return func(c *Course) {
		c.Name = name
	}
}

func WithCapacity(min, max int) CourseOption {
	return func(c *Course) {
		c.MinCapacity = min
		c.MaxCapacity = max
	}
}

func RandomParticipant(options ...ParticipantOption) Participant {
	prename := seededuuid.SeededUUID()
	surname := seededuuid.SeededUUID()

	p := Participant{EncryptedPrename: prename.String(), EncryptedSurname: surname.String()}

	for _, option := range options {
		option(&p)
	}

	return p
}

func RandomParticipants(n int) (result []Participant) {
	for i := 0; i < n; i++ {
		result = append(result, RandomParticipant())
	}

	return
}

func RandomCourses(n int) (result []Course) {
	for i := 0; i < n; i++ {
		result = append(result, RandomCourse())
	}

	return
}
