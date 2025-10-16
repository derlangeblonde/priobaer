package solve

import (
	"gorm.io/gorm"
)

// ComputeAndApplyOptimalAssignments reads current scenario from the DB, computes which assignments would be optimal
// to satisfy the prioritization of the still unassigned participants and writes these assignments to the DB.
// Prefer passing a transaction as DB-handle so that partial updates will be rolled back in case of an error.
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
