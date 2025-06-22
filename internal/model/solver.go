package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"softbaer.dev/ass/internal/z3"
)

const separator = "[in]"

var NotSolvable = errors.New("problem instance is not solvable")

func SolveAssignment(priorities []Priority) (assignments []Assignment, err error) {
	optimizationProblem := newOptimizationProblem(priorities)
	defer optimizationProblem.Close()

	return optimizationProblem.Solve()
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
	p.optimize.Close()
	p.ctx.Close()
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
		c.optimize.Assert(zero.Add(allVariablesForOneParticipant...).Gt(zero))
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
		// JS 21.06.2025 - Both maps share the same keys. Therefore, this value always exists.
		remainingCapacity, _ := c.remainingCapacityByCourseId[courseId]
		c.optimize.Assert(zero.Add(variablesForCourse...).Le(c.ctx.Int(remainingCapacity, c.ctx.IntSort())))
	}
}

type minimumCapacityConstraint struct {
	ctx                        *z3.Context
	optimize                   *z3.Optimize
	variablesByCourseId        map[int][]*z3.AST
	gapToMinCapacityByCourseId map[int]int
}

func newMinimumCapacityConstraint(s *optimizationProblem) *minimumCapacityConstraint {
	return &minimumCapacityConstraint{ctx: s.ctx, optimize: s.optimize, variablesByCourseId: make(map[int][]*z3.AST), gapToMinCapacityByCourseId: make(map[int]int, 0)}
}

func (c *minimumCapacityConstraint) add(prio Priority, variable *z3.AST) {
	c.variablesByCourseId[prio.CourseID] = append(c.variablesByCourseId[prio.CourseID], variable)
	c.gapToMinCapacityByCourseId[prio.CourseID] = prio.Course.GapToMinCapacity()
}

func (c *minimumCapacityConstraint) build() {
	zero := c.ctx.Int(0, c.ctx.IntSort())

	for courseId, variablesForCourse := range c.variablesByCourseId {
		// JS 21.06.2025 - Both maps share the same keys. Therefore, this value always exists.
		gapToMinCapacity, _ := c.gapToMinCapacityByCourseId[courseId]
		c.optimize.Assert(zero.Add(variablesForCourse...).Ge(c.ctx.Int(gapToMinCapacity, c.ctx.IntSort())).Or(zero.Add(variablesForCourse...).Eq(zero)))
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

func (o *maximizeHighPrioritiesObjective) add(prio Priority, variable *z3.AST) {
	o.variablesWithPriorityLevels = append(o.variablesWithPriorityLevels, varWithPriorityLevel{variable, prio.Level})

	if prio.Level > o.maximumPrioLevel {
		o.maximumPrioLevel = prio.Level
	}
}

func (o *maximizeHighPrioritiesObjective) build() {
	objective := o.ctx.Int(0, o.ctx.IntSort())

	for _, varWithPriorityLevel := range o.variablesWithPriorityLevels {
		objective = objective.Add(o.weightedTerm(varWithPriorityLevel))
	}

	o.optimize.Maximize(objective)
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

func (p *optimizationProblem) Solve() (assignments []Assignment, err error) {
	constrainBuilders := []constraintBuilder{
		newExactlyOneCoursePerParticipantConstraint(p),
		newMaximumCapacityConstraint(p),
		newMinimumCapacityConstraint(p),
		newPreferHighPrioritiesObjective(p),
	}
	solutionParser := newSolutionParser()

	for _, prio := range p.priorities {
		if prio.Course.RemainingCapacity() <= 0 {
			continue
		}

		variable := p.priorityVariable(prio)
		solutionParser.registerLookups(prio.Participant, prio.Course)

		for _, constraint := range constrainBuilders {
			constraint.add(prio, variable)
		}
	}

	for _, constraint := range constrainBuilders {
		constraint.build()
	}

	if v := p.optimize.Check(); v != z3.True {
		return assignments, NotSolvable
	}

	m := p.optimize.Model()
	solution := m.Assignments()

	return solutionParser.parse(solution)
}

func (p *optimizationProblem) priorityVariable(prio Priority) *z3.AST {
	varName := fmt.Sprintf("%d%s%d", prio.ParticipantID, separator, prio.CourseID)
	variable := p.ctx.Const(p.ctx.Symbol(varName), p.ctx.IntSort())

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
		return assignmentId, fmt.Errorf("splitting of varName did not give exactly two ids. VarName: %s", varName)
	}

	participantId, err := strconv.Atoi(idsAsStr[0])

	if err != nil {
		return assignmentId, fmt.Errorf("could not parse participantId: %d, err: %s", participantId, err)
	}

	courseId, err := strconv.Atoi(idsAsStr[1])

	if err != nil {
		return assignmentId, fmt.Errorf("could not parse courseId: %d, err: %s", courseId, err)
	}

	return AssignmentID{ParticipantId: participantId, CourseId: courseId}, err
}
