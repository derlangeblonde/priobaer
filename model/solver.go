package model

import "slices"

type Assignment struct {
	Participant Participant
	Course      Course
}

func SolveAssignment(availableCourses []Course, unassignedParticipants []Participant) (assignments []Assignment) {

	for len(availableCourses) > 0 && len(unassignedParticipants) > 0 {
		courseCandidate := Head(availableCourses)

		if courseCandidate.RemainingCapacity() <= 0 {
			availableCourses = RemoveHead(availableCourses)	

			continue
		}

		participantCandidate := Head(unassignedParticipants)

		courseCandidate.Participants = append(courseCandidate.Participants, participantCandidate)
		assignments =append(assignments, Assignment{Participant: participantCandidate, Course: courseCandidate})

		unassignedParticipants = RemoveHead(unassignedParticipants)
	}

	return assignments
}

func Head[T any](s []T) T {
	return s[0]
}

func RemoveHead[T any](s []T) []T {
	return slices.Delete(s, 0, 1)
}
