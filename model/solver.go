package model

import (
	"log/slog"
	"slices"
	"github.com/mitchellh/go-z3"
)

type Assignment struct {
	Participant Participant
	Course      Course
}

func SolveAssignment(availableCourses []Course, unassignedParticipants []Participant) (assignments []Assignment) {
	z3.NewConfig()

	for len(availableCourses) > 0 && len(unassignedParticipants) > 0 {
		slog.Error("New iteration", "lenc", len(availableCourses), "lenp", len(unassignedParticipants))
		courseCandidate := Head(availableCourses)

		if courseCandidate.RemainingCapacity() <= 0 {
			slog.Error("Removing full course")
			availableCourses = RemoveHead(availableCourses)	

			continue
		}

		slog.Error("Choose to assign", "Remaining Cap in candC", courseCandidate.RemainingCapacity(), "Allocation", courseCandidate.Allocation())

		participantCandidate := Head(unassignedParticipants)

		courseCandidate.Participants = append(courseCandidate.Participants, participantCandidate)
		assignments = append(assignments, Assignment{Participant: participantCandidate, Course: courseCandidate})

		unassignedParticipants = RemoveHead(unassignedParticipants)
		slog.Error("Do Assign")
	}

	return assignments
}

func Head[T any](s []T) T {
	return s[0]
}

func RemoveHead[T any](s []T) []T {
	return slices.Delete(s, 0, 1)
}
