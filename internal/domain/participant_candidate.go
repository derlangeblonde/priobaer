package domain

import (
	"gorm.io/gorm"
	"softbaer.dev/ass/internal/crypt"
	"softbaer.dev/ass/internal/model"
)

type ParticipantCandidate struct {
	ParticipantName
	prioritizedCourseIds []CourseID
	assignedCourseId     CourseID
	isAssigned           bool
}

func NewParticipantCandidate(prename, surname string) *ParticipantCandidate {
	return &ParticipantCandidate{
		ParticipantName: ParticipantName{
			Prename: prename,
			Surname: surname,
		},
	}
}

func (pc *ParticipantCandidate) Prioritize(descendingPriorities []int) {
	pc.prioritizedCourseIds = make([]CourseID, len(descendingPriorities))

	for i := range len(descendingPriorities) {
		pc.prioritizedCourseIds[i] = CourseID(descendingPriorities[i])
	}
}

func (pc *ParticipantCandidate) Assign(maybeCourseID *int) {
	if maybeCourseID != nil {
		pc.assignedCourseId = CourseID(*maybeCourseID)
		pc.isAssigned = true
	}
}

func (pc *ParticipantCandidate) Valid() map[string]string {
	return pc.ParticipantNameValid()
}

func (pc *ParticipantCandidate) Save(db *gorm.DB, secret crypt.Secret) (Participant, error) {
	courseId := model.WithNoCourseId()
	if pc.isAssigned {
		courseId = model.WithSomeCourseId(int64(pc.assignedCourseId))
	}
	dbModel, err := model.NewParticipant(
		pc.Prename,
		pc.Surname,
		secret,
		courseId,
	)
	if err != nil {
		return Participant{}, err
	}

	if err := db.Create(&dbModel).Error; err != nil {
		return Participant{}, err
	}

	priorities := make([]PriorityData, len(pc.prioritizedCourseIds))
	for i, courseId := range pc.prioritizedCourseIds {
		priorities[i] = NewPriorityData(ParticipantID(dbModel.ID), courseId, PriorityLevel(i+1))
	}

	if err := savePriorities(db, priorities); err != nil {
		return Participant{}, err
	}

	savedData, err := participantDataFromDbModel(dbModel, secret)
	if err != nil {
		return Participant{}, err
	}

	result := Participant{
		ParticipantData: savedData,
	}

	result.PrioritizedCourses, err = findCourseDataById(db, pc.prioritizedCourseIds)
	if err != nil {
		return Participant{}, err
	}

	if pc.isAssigned {
		var assignedCourseRow model.Course
		if err := db.First(&assignedCourseRow, "id = ?", int(pc.assignedCourseId)).Error; err != nil {
			return Participant{}, err
		}

		result.assignedCourse = courseFromDbModel(assignedCourseRow)
	}

	return result, nil
}
