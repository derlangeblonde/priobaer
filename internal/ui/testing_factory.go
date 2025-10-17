package ui

import "softbaer.dev/ass/internal/seededuuid"

func RandomParticipant() Participant {
	prename := seededuuid.SeededUUID()
	surname := seededuuid.SeededUUID()

	p := Participant{Prename: prename.String(), Surname: surname.String()}

	return p
}

func RandomParticipants(n int) (result []Participant) {
	for i := 0; i < n; i++ {
		result = append(result, RandomParticipant())
	}

	return
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
