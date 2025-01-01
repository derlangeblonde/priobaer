package model

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/mitchellh/go-z3"
)

const separator = "[in]"

type Assignment struct {
	Participant Participant
	Course      Course
}

type AssignmentIdTuple struct {
	ParticipantId int
	CourseId      int
}

func SolveAssignment(availableCourses []Course, unassignedParticipants []Participant) (assignments []Assignment) {
	ctx, o := NewZ3Optimizer()
	defer ctx.Close()
	defer o.Close()

	idToCourses := make(map[int]Course, 0)
	idToParticipants:= make(map[int]Participant, 0)

	zero := ctx.Int(0, ctx.IntSort())
	one := ctx.Int(1, ctx.IntSort())

	idTupleToVariable := make(map[AssignmentIdTuple]*z3.AST, 0)
	participantIdToVariables := make(map[int][]*z3.AST, 0)
	courseIdToVariables := make(map[int][]*z3.AST, 0)
	var allVariables []*z3.AST

	for _, course := range availableCourses {
		if course.RemainingCapacity() <= 0 {
			continue
		}

		idToCourses[course.ID] = course

		for _, participant := range unassignedParticipants {
			idToParticipants[participant.ID] = participant

			idTuple := AssignmentIdTuple{ParticipantId: participant.ID, CourseId: course.ID}
			varName := fmt.Sprintf("%d%s%d", participant.ID, separator, course.ID)
			variable := ctx.Const(ctx.Symbol(varName), ctx.IntSort())

			allVariables = append(allVariables, variable)
			idTupleToVariable[idTuple] = variable
			participantIdToVariables[participant.ID] = append(participantIdToVariables[participant.ID], variable)
			courseIdToVariables[course.ID] = append(courseIdToVariables[course.ID], variable)
		}
	}

	// optimize for most participants assigned
	o.Maximize(zero.Add(allVariables...))

	// Exactly one particpant in one course
	for _, variableForOneParticipant := range participantIdToVariables {
		o.Assert(zero.Add(variableForOneParticipant...).Le(one))

		for _, variable := range variableForOneParticipant {
			o.Assert(variable.Ge(zero))
			o.Assert(variable.Le(one))
		}
	}

	// respect maxCap for Course
	for courseId, variableForOneCourse := range courseIdToVariables {
		// TODO: handle not ok? 
		course, _ := idToCourses[courseId] 
		o.Assert(zero.Add(variableForOneCourse...).Le(ctx.Int(course.RemainingCapacity(), ctx.IntSort())))
	}

	if v := o.Check(); v != z3.True {
		return assignments
	}

	m := o.Model()
	varsSolved := m.Assignments()

	for varName, solutionStr := range varsSolved {
		// TODO: handle err
		solution, _ := strconv.Atoi(solutionStr.String())

		if solution == 1 {
			idTuple := ParseAssignmentTuple(varName)
			course, _ := idToCourses[idTuple.CourseId]
			participant, _ := idToParticipants[idTuple.ParticipantId]
			assignment := Assignment{Course: course, Participant: participant}
			assignments = append(assignments, assignment)
		}
	}

	return assignments
}

func NewZ3Optimizer() (*z3.Context, *z3.Optimize) {
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	config.Close()
	o := ctx.NewOptimizer()

	return ctx, o
}

func ParseAssignmentTuple(varName string) AssignmentIdTuple {
	idsAsStr := strings.Split(varName, separator)
	// TODO: assert that we get exactly two results

	// TODO: check errs
	participantId, _ := strconv.Atoi(idsAsStr[0])
	courseId, _ := strconv.Atoi(idsAsStr[1])

	return AssignmentIdTuple{ParticipantId: participantId, CourseId: courseId}
}

func Head[T any](s []T) T {
	return s[0]
}

func RemoveHead[T any](s []T) []T {
	return slices.Delete(s, 0, 1)
}

