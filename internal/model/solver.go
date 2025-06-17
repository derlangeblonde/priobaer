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
	systemOfEquations := newOptimizationProblem(priorities)
	// TODO: verify that might not be necessary and also causes a double free
	// defer problem.Close()

	return systemOfEquations.Solve()
}

type optimizationProblem struct {
	ctx        *z3.Context
	optimize   *z3.Optimize
	priorities []Priority
}

func newOptimizationProblem(priorities []Priority) *optimizationProblem {
	ctx, o := newZ3Optimizer()

	return &optimizationProblem{ctx: ctx, optimize: o, priorities: priorities}
}

func (p *optimizationProblem) Close() {
	p.ctx.Close()
	p.optimize.Close()
}

type constraintBuilder interface {
	add(prio Priority, variable *z3.AST)
	build()
}

func newExactlyOneCoursePerParticipantConstraint(s *optimizationProblem) *exactlyOneCoursePerParticipantConstraint {
	return &exactlyOneCoursePerParticipantConstraint{ctx: s.ctx, optimize: s.optimize, variablesByParticipantId: make(map[int][]*z3.AST)}
}

type exactlyOneCoursePerParticipantConstraint struct {
	ctx                      *z3.Context
	optimize                 *z3.Optimize
	variablesByParticipantId map[int][]*z3.AST
}

func (c *exactlyOneCoursePerParticipantConstraint) add(prio Priority, variable *z3.AST) {
	zero := c.ctx.Int(0, c.ctx.IntSort())
	c.optimize.Assert(variable.Ge(zero))

	c.variablesByParticipantId[prio.ParticipantID] = append(c.variablesByParticipantId[prio.ParticipantID], variable)
}

func (c *exactlyOneCoursePerParticipantConstraint) build() {
	zero := c.ctx.Int(0, c.ctx.IntSort())
	one := c.ctx.Int(1, c.ctx.IntSort())
	for _, allVariablesForOneParticipant := range c.variablesByParticipantId {
		c.optimize.Assert(zero.Add(allVariablesForOneParticipant...).Le(one))
	}
}

type maximumCapacityConstraint struct {
	ctx                         *z3.Context
	optimize                    *z3.Optimize
	variablesByCourseId         map[int][]*z3.AST
	remainingCapacityByCourseId map[int]int
}

func newMaximumCapacityConstraint(s *optimizationProblem) *maximumCapacityConstraint {
	return &maximumCapacityConstraint{ctx: s.ctx, optimize: s.optimize, variablesByCourseId: make(map[int][]*z3.AST), remainingCapacityByCourseId: make(map[int]int, 0)}
}

func (c *maximumCapacityConstraint) add(prio Priority, variable *z3.AST) {
	c.variablesByCourseId[prio.CourseID] = append(c.variablesByCourseId[prio.CourseID], variable)
	c.remainingCapacityByCourseId[prio.CourseID] = prio.Course.RemainingCapacity()
}

func (c *maximumCapacityConstraint) build() {
	zero := c.ctx.Int(0, c.ctx.IntSort())

	for courseId, variablesForCourse := range c.variablesByCourseId {
		// Both maps share the same keys. Therefore, this value always exists.
		remainingCapacity, _ := c.remainingCapacityByCourseId[courseId]
		c.optimize.Assert(zero.Add(variablesForCourse...).Le(c.ctx.Int(remainingCapacity, c.ctx.IntSort())))
	}
}

type varWithPriorityLevel struct {
	variable  *z3.AST
	prioLevel PriorityLevel
}

type maximizeHighPrioritiesObjective struct {
	ctx                         *z3.Context
	optimize                    *z3.Optimize
	variablesWithPriorityLevels []varWithPriorityLevel
	maximumPrioLevel            PriorityLevel
}

func newPreferHighPrioritiesObjective(s *optimizationProblem) *maximizeHighPrioritiesObjective {
	return &maximizeHighPrioritiesObjective{ctx: s.ctx, optimize: s.optimize}
}

func (c *maximizeHighPrioritiesObjective) add(prio Priority, variable *z3.AST) {
	c.variablesWithPriorityLevels = append(c.variablesWithPriorityLevels, varWithPriorityLevel{variable, prio.Level})

	if prio.Level > c.maximumPrioLevel {
		c.maximumPrioLevel = prio.Level
	}
}

func (c *maximizeHighPrioritiesObjective) build() {
	objective := c.ctx.Int(0, c.ctx.IntSort())

	for _, varWithPriorityLevel := range c.variablesWithPriorityLevels {
		objective = objective.Add(c.weightedTerm(varWithPriorityLevel))
	}

	c.optimize.Maximize(objective)
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

func (o *optimizationProblem) Solve() (assignments []Assignment, err error) {
	constrainBuilders := []constraintBuilder{
		newExactlyOneCoursePerParticipantConstraint(o),
		newMaximumCapacityConstraint(o),
		newPreferHighPrioritiesObjective(o),
	}
	solutionParser := newSolutionParser()

	for _, prio := range o.priorities {
		if prio.Course.RemainingCapacity() <= 0 {
			continue
		}

		variable := o.priorityVariable(prio)
		solutionParser.registerLookups(prio.Participant, prio.Course)

		for _, constraint := range constrainBuilders {
			constraint.add(prio, variable)
		}
	}

	for _, constraint := range constrainBuilders {
		constraint.build()
	}

	if v := o.optimize.Check(); v != z3.True {
		return assignments, notSolvable
	}

	m := o.optimize.Model()
	solution := m.Assignments()

	return solutionParser.parse(solution)
}

func (o *optimizationProblem) priorityVariable(prio Priority) *z3.AST {
	varName := fmt.Sprintf("%d%s%d", prio.ParticipantID, separator, prio.CourseID)
	variable := o.ctx.Const(o.ctx.Symbol(varName), o.ctx.IntSort())

	return variable
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
