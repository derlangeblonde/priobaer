package solve

import (
	"fmt"
	"strconv"
	"strings"

	"softbaer.dev/ass/internal/domain"
	"softbaer.dev/ass/internal/z3"
)

func parseSolution(solution map[string]*z3.AST) (assignments []computedAssignment, err error) {
	for varName, solutionStr := range solution {
		solution, err := strconv.Atoi(solutionStr.String())

		if err != nil {
			return assignments, fmt.Errorf("could not parse assigned solution. varName: %s, solution: %s", varName, solutionStr)
		}

		if solution != 1 {
			continue
		}

		assignment, err := parseAssignment(varName)

		if err != nil {
			return assignments, err
		}

		assignments = append(assignments, assignment)
	}

	return
}

func parseAssignment(varName string) (assignment computedAssignment, err error) {
	idsAsStr := strings.Split(varName, separator)

	if len(idsAsStr) != 2 {
		return assignment, fmt.Errorf("splitting of varName did not give exactly two ids. VarName: %s", varName)
	}

	participantId, err := strconv.Atoi(idsAsStr[0])

	if err != nil {
		return assignment, fmt.Errorf("could not parse participantId: %d, err: %s", participantId, err)
	}

	courseId, err := strconv.Atoi(idsAsStr[1])

	if err != nil {
		return assignment, fmt.Errorf("could not parse courseId: %d, err: %s", courseId, err)
	}

	return newComputedAssignment(domain.ParticipantID(participantId), domain.CourseID(courseId)), nil
}
