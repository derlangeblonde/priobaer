package infra

type Assignment struct {
	Participant Participant
	Course      Course
}

type AssignmentID struct {
	ParticipantId int
	CourseId      int
}
