package domain

import (
	"database/sql"

	"gorm.io/gorm"
	"softbaer.dev/ass/internal/crypt"
	"softbaer.dev/ass/internal/domain/store"
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
	encryptedName, err := pc.ParticipantName.encrypt(secret)
	if err != nil {
		return Participant{}, err
	}

	dbModel := model.Participant{
		Prename:  encryptedName.Prename,
		Surname:  encryptedName.Surname,
		CourseID: sql.NullInt64{Valid: pc.isAssigned, Int64: int64(pc.assignedCourseId)},
	}

	if err := db.Create(&dbModel).Error; err != nil {
		return Participant{}, err
	}

	// TODO: When SetPriories is part of domain, it can accpect the type CourseID instead of int
	courseIds := make([]int, len(pc.prioritizedCourseIds))
	for i, courseId := range pc.prioritizedCourseIds {
		courseIds[i] = int(courseId)
	}

	if err := store.SetPriorities(db, dbModel.ID, courseIds); err != nil {
		return Participant{}, err
	}

	savedData, err := participantDataFromDbModel(dbModel, secret)
	if err != nil {
		return Participant{}, err
	}

	result := Participant{
		ParticipantData: savedData,
	}

	courseRows, err := store.GetPriorities(db, dbModel.ID)

	if err != nil {
		return Participant{}, err
	}

	result.PrioritizedCourses = CoursesFromDbModels(courseRows)

	if pc.isAssigned {
		var assignedCourseRow model.Course
		if err := db.First(&assignedCourseRow, "id = ?", int(pc.assignedCourseId)).Error; err != nil {
			return Participant{}, err
		}

		result.assignedCourse = CourseFromDbModel(assignedCourseRow)
	}

	return result, nil
}
