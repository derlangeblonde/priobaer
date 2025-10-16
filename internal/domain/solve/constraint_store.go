package solve

import (
	"slices"

	"gorm.io/gorm"
	"softbaer.dev/ass/internal/domain"
	"softbaer.dev/ass/internal/model"
)

func queryPriorityConstraints(db *gorm.DB) ([]priorityConstraint, error) {
	var assignableParticipants []model.Participant
	if err := db.Find(&assignableParticipants, "course_id is null").Error; err != nil {
		return nil, err
	}

	var assignableParticipantIds []int
	for _, p := range assignableParticipants {
		assignableParticipantIds = append(assignableParticipantIds, p.ID)
	}

	var satisfiablePrios []model.Priority
	if err := db.Find(&satisfiablePrios, "participant_id in ?", assignableParticipantIds).Error; err != nil {
		return nil, err
	}

	var relevantCourseIds []int
	for _, prio := range satisfiablePrios {
		if !slices.Contains(relevantCourseIds, prio.CourseID) {
			relevantCourseIds = append(relevantCourseIds, prio.CourseID)
		}
	}

	var relevantCourses []model.Course
	if err := db.Preload("Participants").Find(&relevantCourses, "id in ?", relevantCourseIds).Error; err != nil {
		return nil, err
	}

	courseConstraintsById := make(map[int]courseConstraint)
	for _, c := range relevantCourses {
		courseConstraintsById[c.ID] = newCourseConstraint(domain.CourseID(c.ID), c.GapToMinCapacity(), c.RemainingCapacity())
	}

	var result []priorityConstraint
	for _, prio := range satisfiablePrios {
		courseConstraint := courseConstraintsById[prio.CourseID]
		result = append(result, newPriorityConstraint(domain.PriorityLevel(prio.Level), courseConstraint, domain.ParticipantID(prio.ParticipantID)))
	}

	return result, nil
}
