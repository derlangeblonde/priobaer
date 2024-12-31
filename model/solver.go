package model

import (
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/mitchellh/go-z3"
)

const separator = "[in]"

type Assignment struct {
	Participant Participant
	Course      Course
}

type AssignmentIdTuple struct {
	ParticipantId int
	CourseId      int
}

func SolveAssignment(availableCourses []Course, unassignedParticipants []Participant) (assignments []Assignment) {
	config := z3.NewConfig()
	ctx := z3.NewContext(config)
	config.Close()
	defer ctx.Close()
	s := ctx.NewSolver()
	defer s.Close()

	zero := ctx.Int(0, ctx.IntSort())
	one := ctx.Int(1, ctx.IntSort())

	idTupleToVariable := make(map[AssignmentIdTuple]*z3.AST, 0)
	participantIdToVariables := make(map[int][]*z3.AST, 0)
	courseIdToVariables := make(map[int][]*z3.AST, 0)

	for _, course := range availableCourses {
		if course.RemainingCapacity() <= 0 {
			continue
		}

		for _, participant := range unassignedParticipants {
			idTuple := AssignmentIdTuple{ParticipantId: participant.ID, CourseId: course.ID} 
			varName := fmt.Sprintf("%d%s%d", participant.ID, separator, course.ID)
			variable := ctx.Const(ctx.Symbol(varName), ctx.IntSort())

			idTupleToVariable[idTuple] = variable
			participantIdToVariables[participant.ID] = append(participantIdToVariables[participant.ID], variable)
			courseIdToVariables[course.ID] = append(courseIdToVariables[course.ID], variable)
		}
	}
	slog.Error("varmaps", "all", idTupleToVariable)

	// Exactly one particpant in one course
	for _, variableForOneParticipant := range participantIdToVariables {
		s.Assert(zero.Add(variableForOneParticipant...).Eq(one))	

		for _, variable := range variableForOneParticipant {
			s.Assert(variable.Ge(zero))
			s.Assert(variable.Le(one))
		}
	}

	// respect maxCap for Course
	for courseId, variableForOneCourse := range courseIdToVariables {
		course := FindCourse(availableCourses, courseId)
		s.Assert(zero.Add(variableForOneCourse...).Le(ctx.Int(course.RemainingCapacity(), ctx.IntSort())))
	}

	
	slog.Error("sabelsabelsabel")
	if v := s.Check(); v != z3.True {
		slog.Error("Unsolveable")
		return assignments
	}

	m := s.Model()
	varsSolved := m.Assignments()

	slog.Error("blablabl")
	for varName, solutionStr := range varsSolved {
		slog.Error("Var", "Name", varName, "solution", solutionStr)
		// TODO: handle err
		solution, _  := strconv.Atoi(solutionStr.String())

		if solution == 1 {
			idTuple := ParseAssignmentTuple(varName)
			assignments = append(assignments, FullAssignmentFromTuple(idTuple, availableCourses, unassignedParticipants))
		}
	}

	// for len(availableCourses) > 0 && len(unassignedParticipants) > 0 {
	// 	slog.Error("New iteration", "lenc", len(availableCourses), "lenp", len(unassignedParticipants))
	// 	courseCandidate := Head(availableCourses)
	//
	// 	if courseCandidate.RemainingCapacity() <= 0 {
	// 		slog.Error("Removing full course")
	// 		availableCourses = RemoveHead(availableCourses)
	//
	// 		continue
	// 	}
	//
	// 	slog.Error("Choose to assign", "Remaining Cap in candC", courseCandidate.RemainingCapacity(), "Allocation", courseCandidate.Allocation())
	//
	// 	participantCandidate := Head(unassignedParticipants)
	//
	// 	courseCandidate.Participants = append(courseCandidate.Participants, participantCandidate)
	// 	assignments = append(assignments, Assignment{Participant: participantCandidate, Course: courseCandidate})
	//
	// 	unassignedParticipants = RemoveHead(unassignedParticipants)
	// 	slog.Error("Do Assign")
	// }

	return assignments
}

func ParseAssignmentTuple(varName string) AssignmentIdTuple {
	idsAsStr := strings.Split(varName, separator)
	// TODO: assert that we get exactly two results

	// TODO: check errs
	participantId, _ := strconv.Atoi(idsAsStr[0])
	courseId, _ := strconv.Atoi(idsAsStr[1])


	return AssignmentIdTuple{ParticipantId: participantId, CourseId: courseId}
}

func FullAssignmentFromTuple(tuple AssignmentIdTuple, courses []Course, participants []Participant) Assignment {
	return Assignment{Course: FindCourse(courses, tuple.CourseId), Participant: FindParticipant(participants, tuple.ParticipantId)}
}

func Head[T any](s []T) T {
	return s[0]
}

func RemoveHead[T any](s []T) []T {
	return slices.Delete(s, 0, 1)
}

func FindCourse(courses []Course, id int) Course {
	for _, c := range courses {
		if c.ID == id {
			return c 
		}
	}

	// TODO: is this a good idea???
	panic(fmt.Sprintf("No Course found with id: %d", id))
}

func FindParticipant(participants []Participant, id int) Participant{
	for _, p := range participants {
		if p.ID == id {
			return p 
		}
	}

	// TODO: is this a good idea???
	panic(fmt.Sprintf("No Participant found with id: %d", id))
}
