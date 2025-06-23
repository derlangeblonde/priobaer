package domain

type PriorityLevel uint8

type Priority struct {
	Level       PriorityLevel
	Participant Participant
	Course      Course
}
