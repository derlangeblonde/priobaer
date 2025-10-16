package solve

import (
	"errors"

	"gorm.io/gorm"
)

var NotSolvable = errors.New("problem instance is not solvable")

func ApplyOptimalAssignments(tx *gorm.DB) error {
	priorityConstraints, err := queryPriorityConstraints(tx)
	if err != nil {
		return err
	}

	optimalAssignments, err := computeOptimalAssignments(priorityConstraints)
	if err != nil {
		return err
	}

	return applyAssignments(tx, optimalAssignments)
}
