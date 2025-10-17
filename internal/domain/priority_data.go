package domain

type PriorityData struct {
	Level         PriorityLevel
	ParticipantID ParticipantID
	CourseID      CourseID
}

func NewPriorityData(participant ParticipantID, course CourseID, level PriorityLevel) PriorityData {
	return PriorityData{ParticipantID: participant, CourseID: course, Level: level}
}

type PriorityLevel uint8
