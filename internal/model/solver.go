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

func SolveAssignment(priorities []Priority) (assignments []Assignment, err error) {
	problem := NewProblem(priorities)
	// TODO: verify that might not be necessary and also causes a double free
	// defer problem.Close()

	return problem.Solve()
}

type Problem struct {
	ctx *z3.Context	
	optimize *z3.Optimize
	priorities []Priority
}

func NewProblem(priorities []Priority) *Problem {
	ctx, o := NewZ3Optimizer()
	
	return &Problem{ctx: ctx, optimize: o, priorities: priorities}
}

func (p *Problem) Close() {
	p.ctx.Close()
	p.optimize.Close()
}

func (p *Problem) Solve() (assignments []Assignment, err error) {
	coursesById := make(map[int]Course, 0)
	participantsById := make(map[int]Participant, 0)

	zero := p.ctx.Int(0, p.ctx.IntSort())
	one := p.ctx.Int(1, p.ctx.IntSort())

	variablesByAssignmentId := make(map[AssignmentID]*z3.AST, 0)
	variablesByParticipantId := make(map[int][]*z3.AST, 0)
	variablesByCourseId := make(map[int][]*z3.AST, 0)
	var allVariables []*z3.AST
	objective := p.ctx.Int(0, p.ctx.IntSort())

	for _, prio := range p.priorities {
		if prio.Course.RemainingCapacity() <= 0 {
			continue
		}

		coursesById[prio.CourseID] = prio.Course 
		participantsById[prio.ParticipantID] = prio.Participant
		assignmentId := AssignmentID{ParticipantId: prio.ParticipantID, CourseId: prio.CourseID}
		varName := fmt.Sprintf("%d%s%d", prio.ParticipantID, separator, prio.CourseID)
		variable := p.ctx.Const(p.ctx.Symbol(varName), p.ctx.IntSort())

		allVariables = append(allVariables, variable)
		variablesByAssignmentId[assignmentId] = variable
		variablesByParticipantId[prio.ParticipantID] = append(variablesByParticipantId[prio.ParticipantID], variable)
		variablesByCourseId[prio.CourseID] = append(variablesByCourseId[prio.CourseID], variable)

		objective = objective.Add(priorityLevelToZ3Const(p.ctx, prio.Level).Mul(variable))
	}

	p.optimize.Maximize(objective)

	// Exactly one particpant in one course
	for _, variablesForOneParticipant := range variablesByParticipantId {
		p.optimize.Assert(zero.Add(variablesForOneParticipant...).Le(one))

		for _, variable := range variablesForOneParticipant {
			p.optimize.Assert(variable.Ge(zero))
			p.optimize.Assert(variable.Le(one))
		}
	}

	// respect maxCap for Course
	for courseId, variableForOneCourse := range variablesByCourseId {
		course, ok := coursesById[courseId]

		if !ok {
			return assignments, fmt.Errorf("Did not find course with id: %d", courseId)
		}

		p.optimize.Assert(zero.Add(variableForOneCourse...).Le(p.ctx.Int(course.RemainingCapacity(), p.ctx.IntSort())))
	}

	if v := p.optimize.Check(); v != z3.True {
		return assignments, notSolvable 
	}

	m := p.optimize.Model()
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


