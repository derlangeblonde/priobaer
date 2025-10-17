package ui

import "softbaer.dev/ass/internal/seededuuid"

func RandomParticipant() Participant {
	prename := seededuuid.SeededUUID()
	surname := seededuuid.SeededUUID()

	p := Participant{Prename: prename.String(), Surname: surname.String()}

	return p
}

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
