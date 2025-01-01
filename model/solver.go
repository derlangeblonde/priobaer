package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/go-z3"
)

const separator = "[in]"

type Assignment struct {
	Participant Participant
	Course      Course
}

type AssignmentID struct {
	ParticipantId int
	CourseId      int
}

func SolveAssignment(availableCourses []Course, unassignedParticipants []Participant) (assignments []Assignment, err error) {
	ctx, o := NewZ3Optimizer()
	defer ctx.Close()
	defer o.Close()

	coursesById := make(map[int]Course, 0)
	participantsById := make(map[int]Participant, 0)

	zero := ctx.Int(0, ctx.IntSort())
	one := ctx.Int(1, ctx.IntSort())

	variablesByAssignmentId := make(map[AssignmentID]*z3.AST, 0)
	variablesByParticipantId := make(map[int][]*z3.AST, 0)
	variablesByCourseId := make(map[int][]*z3.AST, 0)
	var allVariables []*z3.AST

	for _, course := range availableCourses {
		if course.RemainingCapacity() <= 0 {
			continue
		}

		coursesById[course.ID] = course

		for _, participant := range unassignedParticipants {
			participantsById[participant.ID] = participant

			assignmentId := AssignmentID{ParticipantId: participant.ID, CourseId: course.ID}
			varName := fmt.Sprintf("%d%s%d", participant.ID, separator, course.ID)
			variable := ctx.Const(ctx.Symbol(varName), ctx.IntSort())

			allVariables = append(allVariables, variable)
			variablesByAssignmentId[assignmentId] = variable
			variablesByParticipantId[participant.ID] = append(variablesByParticipantId[participant.ID], variable)
			variablesByCourseId[course.ID] = append(variablesByCourseId[course.ID], variable)
		}
	}

	// optimize for most participants assigned
	o.Maximize(zero.Add(allVariables...))

	// Exactly one particpant in one course
	for _, variableForOneParticipant := range variablesByParticipantId {
		o.Assert(zero.Add(variableForOneParticipant...).Le(one))

		for _, variable := range variableForOneParticipant {
			o.Assert(variable.Ge(zero))
			o.Assert(variable.Le(one))
		}
	}

	// respect maxCap for Course
	for courseId, variableForOneCourse := range variablesByCourseId {
		course, ok := coursesById[courseId]

		if !ok {
			return assignments, fmt.Errorf("Did not find course with id: %d", courseId)
		}

		o.Assert(zero.Add(variableForOneCourse...).Le(ctx.Int(course.RemainingCapacity(), ctx.IntSort())))
	}

	if v := o.Check(); v != z3.True {
		return assignments, err
	}

	m := o.Model()
	varsSolved := m.Assignments()

	for varName, solutionStr := range varsSolved {
		solution, err := strconv.Atoi(solutionStr.String())

		if err != nil {
			return assignments, fmt.Errorf("Could not parse assigned solution. varName: %s, solution: %s", varName, solutionStr)
		}

		if solution != 1 {
			continue
		}

		idTuple, err := ParseAssignmentTuple(varName)

		if err != nil {
			return assignments, err
		}

		course, ok := coursesById[idTuple.CourseId]

		if !ok {
			return assignments, fmt.Errorf("Did not find course with id: %d", idTuple.CourseId)
		}

		participant, ok := participantsById[idTuple.ParticipantId]

		if !ok {
			return assignments, fmt.Errorf("Did not find participant with id: %d", idTuple.ParticipantId)
		}

		assignment := Assignment{Course: course, Participant: participant}
		assignments = append(assignments, assignment)
	}

	return assignments, err
}

func NewZ3Optimizer() (*z3.Context, *z3.Optimize) {
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	config.Close()
	o := ctx.NewOptimizer()

	return ctx, o
}

func ParseAssignmentTuple(varName string) (tuple AssignmentID, err error) {
	idsAsStr := strings.Split(varName, separator)

	if len(idsAsStr) != 2 {
		return tuple, fmt.Errorf("Splitting of varName did not give exactly two ids. VarName: %s", varName)
	}

	participantId, err := strconv.Atoi(idsAsStr[0])

	if err != nil {
		return tuple, fmt.Errorf("Could not parse participantId: %d, err: %s", participantId, err)
	}

	courseId, err := strconv.Atoi(idsAsStr[1])

	if err != nil {
		return tuple, fmt.Errorf("Could not parse courseId: %d, err: %s", courseId, err)
	}

	return AssignmentID{ParticipantId: participantId, CourseId: courseId}, err
}
