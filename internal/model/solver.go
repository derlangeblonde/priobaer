package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"softbaer.dev/ass/internal/z3"
)

const separator = "[in]"
var notSolvable = errors.New("Problem instance is not solvable")


func SolveAssignment(availableCourses []Course, unassignedParticipants []Participant, priorities []Priority) (assignments []Assignment, err error) {
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
	objective := ctx.Int(0, ctx.IntSort())
	for _, prio := range priorities {
		variable, ok := variablesByAssignmentId[prio.AssignmentID()]		

		if !ok {
			continue
		}

		objective = objective.Add(priorityLevelToZ3Const(ctx, prio.Level).Mul(variable))
	}
	o.Maximize(objective)

	// Exactly one particpant in one course
	for _, variablesForOneParticipant := range variablesByParticipantId {
		o.Assert(zero.Add(variablesForOneParticipant...).Le(one))

		for _, variable := range variablesForOneParticipant {
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
		return assignments, notSolvable 
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


		assignmentId, err := ParseAssignmentId(varName)

		if err != nil {
			return assignments, err
		}

		course, ok := coursesById[assignmentId.CourseId]

		if !ok {
			return assignments, fmt.Errorf("Did not find course with id: %d", assignmentId.CourseId)
		}

		participant, ok := participantsById[assignmentId.ParticipantId]

		if !ok {
			return assignments, fmt.Errorf("Did not find participant with id: %d", assignmentId.ParticipantId)
		}

		assignment := Assignment{Course: course, Participant: participant}
		assignments = append(assignments, assignment)
	}

	return assignments, nil 
}

func NewZ3Optimizer() (*z3.Context, *z3.Optimize) {
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	config.Close()
	o := ctx.NewOptimizer()

	return ctx, o
}

func ParseAssignmentId(varName string) (assignmentId AssignmentID, err error) {
	idsAsStr := strings.Split(varName, separator)

	if len(idsAsStr) != 2 {
		return assignmentId, fmt.Errorf("Splitting of varName did not give exactly two ids. VarName: %s", varName)
	}

	participantId, err := strconv.Atoi(idsAsStr[0])

	if err != nil {
		return assignmentId, fmt.Errorf("Could not parse participantId: %d, err: %s", participantId, err)
	}

	courseId, err := strconv.Atoi(idsAsStr[1])

	if err != nil {
		return assignmentId, fmt.Errorf("Could not parse courseId: %d, err: %s", courseId, err)
	}

	return AssignmentID{ParticipantId: participantId, CourseId: courseId}, err
}

func priorityLevelToZ3Const(ctx *z3.Context, prioLevel PriorityLevel) *z3.AST {
	// TODO: 4 - x is a hack
	return ctx.Int(4 - int(prioLevel), ctx.IntSort()) 
}


