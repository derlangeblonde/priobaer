package solve

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"golang.org/x/sync/semaphore"
	"softbaer.dev/ass/internal/domain"
	"softbaer.dev/ass/internal/z3"
)

var NotSolvable = errors.New("problem instance is not solvable")

// rateLimit limits the number of assignment problems that can be solved in parallel.
// Solving can be rather comput intensive. We limit parallelization to prevent CPU from being overbooked.
var rateLimit = semaphore.NewWeighted(1)

func computeOptimalAssignments(priorities []priorityConstraint) (assignments []computedAssignment, err error) {
	if err = rateLimit.Acquire(context.Background(), 1); err != nil {
		return nil, err
	}
	defer rateLimit.Release(1)

	optimizationProblem := newOptimizationProblem(priorities)
	defer optimizationProblem.Close()

	return optimizationProblem.solve()
}

type optimizationProblem struct {
	ctx        *z3.Context
	optimize   *z3.Optimize
	priorities []priorityConstraint
}

func newOptimizationProblem(priorities []priorityConstraint) *optimizationProblem {
	ctx, o := newZ3Optimizer()

	return &optimizationProblem{ctx: ctx, optimize: o, priorities: priorities}
}

func (p *optimizationProblem) Close() {
	if err := p.optimize.Close(); err != nil {
		slog.Error("Could not close z3.Optimize", "err", err)
	}
	if err := p.ctx.Close(); err != nil {
		slog.Error("Could not close ctx", "err", err)
	}
}

type constraintBuilder interface {
	add(prio priorityConstraint, variable *z3.AST)
	build()
}

func newExactlyOneCoursePerParticipantConstraint(s *optimizationProblem) *exactlyOneCoursePerParticipantConstraint {
	return &exactlyOneCoursePerParticipantConstraint{ctx: s.ctx, optimize: s.optimize, variablesByParticipantId: make(map[domain.ParticipantID][]*z3.AST)}
}

type exactlyOneCoursePerParticipantConstraint struct {
	ctx                      *z3.Context
	optimize                 *z3.Optimize
	variablesByParticipantId map[domain.ParticipantID][]*z3.AST
}

func (c *exactlyOneCoursePerParticipantConstraint) add(prio priorityConstraint, variable *z3.AST) {
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
	variablesByCourseId         map[domain.CourseID][]*z3.AST
	remainingCapacityByCourseId map[domain.CourseID]int
}

func newMaximumCapacityConstraint(s *optimizationProblem) *maximumCapacityConstraint {
	return &maximumCapacityConstraint{ctx: s.ctx, optimize: s.optimize, variablesByCourseId: make(map[domain.CourseID][]*z3.AST), remainingCapacityByCourseId: make(map[domain.CourseID]int)}
}

func (c *maximumCapacityConstraint) add(prio priorityConstraint, variable *z3.AST) {
	c.variablesByCourseId[prio.CourseConstraint.CourseID] = append(c.variablesByCourseId[prio.CourseConstraint.CourseID], variable)
	c.remainingCapacityByCourseId[prio.CourseConstraint.CourseID] = prio.CourseConstraint.RemainingCapacity
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
	variablesByCourseId        map[domain.CourseID][]*z3.AST
	gapToMinCapacityByCourseId map[domain.CourseID]int
}

func newMinimumCapacityConstraint(s *optimizationProblem) *minimumCapacityConstraint {
	return &minimumCapacityConstraint{ctx: s.ctx, optimize: s.optimize, variablesByCourseId: make(map[domain.CourseID][]*z3.AST), gapToMinCapacityByCourseId: make(map[domain.CourseID]int)}
}

func (c *minimumCapacityConstraint) add(prio priorityConstraint, variable *z3.AST) {
	c.variablesByCourseId[prio.CourseConstraint.CourseID] = append(c.variablesByCourseId[prio.CourseConstraint.CourseID], variable)
	c.gapToMinCapacityByCourseId[prio.CourseConstraint.CourseID] = prio.CourseConstraint.GapToMinCapacity
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
	prioLevel domain.PriorityLevel
}

type maximizeHighPrioritiesObjective struct {
	ctx                         *z3.Context
	optimize                    *z3.Optimize
	variablesWithPriorityLevels []varWithPriorityLevel
	maximumPrioLevel            domain.PriorityLevel
}

func newPreferHighPrioritiesObjective(s *optimizationProblem) *maximizeHighPrioritiesObjective {
	return &maximizeHighPrioritiesObjective{ctx: s.ctx, optimize: s.optimize}
}

func (o *maximizeHighPrioritiesObjective) add(prio priorityConstraint, variable *z3.AST) {
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
func (o *maximizeHighPrioritiesObjective) invertPriorityLevel(level domain.PriorityLevel) *z3.AST {
	coeff := (int(o.maximumPrioLevel) + 1) - int(level)
	return o.ctx.Int(coeff, o.ctx.IntSort())
}

func (o *maximizeHighPrioritiesObjective) weightedTerm(varWithPriorityLevel varWithPriorityLevel) *z3.AST {
	return o.invertPriorityLevel(varWithPriorityLevel.prioLevel).Mul(varWithPriorityLevel.variable)
}

func (p *optimizationProblem) solve() (assignments []computedAssignment, err error) {
	constrainBuilders := []constraintBuilder{
		newExactlyOneCoursePerParticipantConstraint(p),
		newMaximumCapacityConstraint(p),
		newMinimumCapacityConstraint(p),
		newPreferHighPrioritiesObjective(p),
	}

	for _, prio := range p.priorities {
		if prio.CourseConstraint.RemainingCapacity <= 0 {
			continue
		}

		variable := p.priorityVariable(prio)

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

	return parseSolution(solution)
}

func (p *optimizationProblem) priorityVariable(prio priorityConstraint) *z3.AST {
	varName := fmt.Sprintf("%d%s%d", prio.ParticipantID, separator, prio.CourseConstraint.CourseID)
	variable := p.ctx.Const(p.ctx.Symbol(varName), p.ctx.IntSort())

	return variable
}

func newZ3Optimizer() (*z3.Context, *z3.Optimize) {
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	if err := config.Close(); err != nil {
		slog.Error("Failed to close config", "err", err)
	}
	o := ctx.NewOptimizer()

	return ctx, o
}
