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
	systemOfEquations := newSystenOfEquations(priorities)
	// TODO: verify that might not be necessary and also causes a double free
	// defer problem.Close()

	return systemOfEquations.Solve()
}

type systemOfEquations struct {
	ctx *z3.Context	
	optimize *z3.Optimize
	priorities []Priority
}

func newSystenOfEquations(priorities []Priority) *systemOfEquations {
	ctx, o := newZ3Optimizer()
	
	return &systemOfEquations{ctx: ctx, optimize: o, priorities: priorities}
}

func (p *systemOfEquations) Close() {
	p.ctx.Close()
	p.optimize.Close()
}

type constraint interface {
	respect(prio Priority)
	finalize()
}

func newParticipantCanBeOnlyInOneCourseConstraint(s *systemOfEquations) *participantCanBeOnlyInOneCourseConstraint {
	return &participantCanBeOnlyInOneCourseConstraint{ctx: s.ctx, optimize: s.optimize, variablesByParticipantId: make(map[int][]*z3.AST)}
}
type participantCanBeOnlyInOneCourseConstraint struct {
	ctx *z3.Context
	optimize *z3.Optimize
	variablesByParticipantId map[int][]*z3.AST
}


func (c *participantCanBeOnlyInOneCourseConstraint) respect(prio Priority, variable *z3.AST) {
	zero := c.ctx.Int(0, c.ctx.IntSort())
	c.optimize.Assert(variable.Ge(zero))

	c.variablesByParticipantId[prio.ParticipantID] = append(c.variablesByParticipantId[prio.ParticipantID], variable)
}

func (c *participantCanBeOnlyInOneCourseConstraint) finalize() {
	zero := c.ctx.Int(0, c.ctx.IntSort())
	one := c.ctx.Int(1, c.ctx.IntSort())
	for _, allVariablesForOneParticipant := range c.variablesByParticipantId {
		c.optimize.Assert(zero.Add(allVariablesForOneParticipant...).Le(one))
	}
}

type courseCannotBeOverbookedConstraint struct {
	ctx *z3.Context
	optimize *z3.Optimize
	variablesByCourseId map[int][]*z3.AST
	remainingCapacityByCourseId map[int]int
} 

func newCourseCannotBeOverbookedConstraint(s *systemOfEquations) *courseCannotBeOverbookedConstraint {
	return &courseCannotBeOverbookedConstraint{ctx: s.ctx, optimize: s.optimize, variablesByCourseId: make(map[int][]*z3.AST), remainingCapacityByCourseId: make(map[int]int, 0)}
}

func (c *courseCannotBeOverbookedConstraint) respect(prio Priority, variable *z3.AST) {
	c.variablesByCourseId[prio.CourseID] = append(c.variablesByCourseId[prio.CourseID], variable)
	// TODO: unnecessary to set it new everytime
	c.remainingCapacityByCourseId[prio.CourseID] = prio.Course.RemainingCapacity() 
}

func (c *courseCannotBeOverbookedConstraint) finalize() {
	zero := c.ctx.Int(0, c.ctx.IntSort())

	for courseId, variablesForCourse := range c.variablesByCourseId {
		remainingCapacity, _ := c.remainingCapacityByCourseId[courseId]
		c.optimize.Assert(zero.Add(variablesForCourse...).Le(c.ctx.Int(remainingCapacity, c.ctx.IntSort())))
	}
}

type varWithPriorityLevel struct{
	variable *z3.AST
	prioLevel PriorityLevel
}

type maximizeHighPrioritiesObjective struct {
	ctx *z3.Context
	optimize *z3.Optimize
	variablesWithPriorityLevels []varWithPriorityLevel
	maximumPrioLevel PriorityLevel
} 

func newMaximizeHighPrioritiesObjective(s *systemOfEquations) *maximizeHighPrioritiesObjective {
	return &maximizeHighPrioritiesObjective{ctx: s.ctx, optimize: s.optimize}
}

func (c *maximizeHighPrioritiesObjective) respect(prio Priority, variable *z3.AST) {
	c.variablesWithPriorityLevels = append(c.variablesWithPriorityLevels, varWithPriorityLevel{variable, prio.Level})

	if prio.Level > c.maximumPrioLevel {
		c.maximumPrioLevel = prio.Level
	}
}

func (c *maximizeHighPrioritiesObjective) finalize() {
	objective := c.ctx.Int(0, c.ctx.IntSort())

	for _, varWithPriorityLevel := range c.variablesWithPriorityLevels {
		objective = objective.Add(c.weightedTerm(varWithPriorityLevel))
	}

	c.optimize.Maximize(objective)
}

func (s *systemOfEquations) Solve() (assignments []Assignment, err error) {
	constraint := newParticipantCanBeOnlyInOneCourseConstraint(s)
	constraint2 := newCourseCannotBeOverbookedConstraint(s)
	objective := newMaximizeHighPrioritiesObjective(s)

	coursesById := make(map[int]Course, 0)
	participantsById := make(map[int]Participant, 0)

	for _, prio := range s.priorities {
		if prio.Course.RemainingCapacity() <= 0 {
			continue
		}
		varName := fmt.Sprintf("%d%s%d", prio.ParticipantID, separator, prio.CourseID)
		variable := s.ctx.Const(s.ctx.Symbol(varName), s.ctx.IntSort())

		constraint.respect(prio, variable)
		constraint2.respect(prio, variable)
		objective.respect(prio, variable)

		coursesById[prio.CourseID] = prio.Course 
		participantsById[prio.ParticipantID] = prio.Participant

	}

	constraint.finalize()
	constraint2.finalize()
	objective.finalize()

	if v := s.optimize.Check(); v != z3.True {
		return assignments, notSolvable 
	}

	m := s.optimize.Model()
	varsSolved := m.Assignments()

	for varName, solutionStr := range varsSolved {
		solution, err := strconv.Atoi(solutionStr.String())

		if err != nil {
			return assignments, fmt.Errorf("Could not parse assigned solution. varName: %s, solution: %s", varName, solutionStr)
		}

		if solution != 1 {
			continue
		}


		assignmentId, err := parseAssignmentId(varName)

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

func newZ3Optimizer() (*z3.Context, *z3.Optimize) {
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	config.Close()
	o := ctx.NewOptimizer()

	return ctx, o
}

func parseAssignmentId(varName string) (assignmentId AssignmentID, err error) {
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

func (o *maximizeHighPrioritiesObjective) invert(prioLevel PriorityLevel) *z3.AST {
	return o.ctx.Int((int(o.maximumPrioLevel) + 1) - int(prioLevel), o.ctx.IntSort()) 
}

// invertPriorityLevel turns a raw PriorityLevel into a Z3 coefficient,
// so that numerically low levels map to high coefficients.
func (o *maximizeHighPrioritiesObjective) invertPriorityLevel(level PriorityLevel) *z3.AST {
    coeff := (int(o.maximumPrioLevel) + 1) - int(level)
    return o.ctx.Int(coeff, o.ctx.IntSort())
}

func (o *maximizeHighPrioritiesObjective) weightedTerm(varWithPriorityLevel varWithPriorityLevel) *z3.AST {
    return o.invertPriorityLevel(varWithPriorityLevel.prioLevel).Mul(varWithPriorityLevel.variable)
}



