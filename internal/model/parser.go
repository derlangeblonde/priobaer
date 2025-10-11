package model

import (
	"fmt"
	"strconv"

	"softbaer.dev/ass/internal/z3"
)

type solutionParser struct {
	coursesById      map[int]Course
	participantsById map[int]Participant
}

func newSolutionParser() *solutionParser {
	return &solutionParser{coursesById: make(map[int]Course), participantsById: make(map[int]Participant)}
}

func (s *solutionParser) registerLookups(p Participant, c Course) {
	s.participantsById[p.ID] = p
	s.coursesById[c.ID] = c
}

func (s *solutionParser) parse(solution map[string]*z3.AST) (assignments []Assignment, err error) {
	for varName, solutionStr := range solution {
		solution, err := strconv.Atoi(solutionStr.String())

		if err != nil {
			return assignments, fmt.Errorf("could not parse assigned solution. varName: %s, solution: %s", varName, solutionStr)
		}

		if solution != 1 {
			continue
		}

		assignmentId, err := parseAssignmentId(varName)

		if err != nil {
			return assignments, err
		}

		course, ok := s.coursesById[assignmentId.CourseId]

		if !ok {
			return assignments, fmt.Errorf("did not find course with id: %d", assignmentId.CourseId)
		}

		participant, ok := s.participantsById[assignmentId.ParticipantId]

		if !ok {
			return assignments, fmt.Errorf("did not find participant with id: %d", assignmentId.ParticipantId)
		}

		assignment := Assignment{Course: course, Participant: participant}
		assignments = append(assignments, assignment)
	}

	return
}
