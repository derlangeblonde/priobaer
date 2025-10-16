package solve

import (
	"gorm.io/gorm"
)

func ComputeAndApplyOptimalAssignments(tx *gorm.DB) error {
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
