package app

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"softbaer.dev/ass/internal/app/respond"
	"softbaer.dev/ass/internal/crypt"
	"softbaer.dev/ass/internal/domain"
	"softbaer.dev/ass/internal/ui"
)

func ScenarioIndex(c *gin.Context) {
	type request struct {
		CourseIdSelected *int `form:"selected-course"`
	}

	db := GetDB(c)
	secret := crypt.GetSecret(c)

	var req request
	err := c.Bind(&req)

	if err != nil {
		slog.Error("Bad request on AssignmentsIndex", "err", err)
		return
	}

	scenario, err := domain.LoadScenario(db, secret)
	if err != nil {
		respond.InternalServerError(c, "Error while loading scenario", err)
		return
	}

	var selectedParticipants []domain.ParticipantData
	if req.CourseIdSelected == nil {
		selectedParticipants = scenario.Unassigned()
	} else {
		selectedParticipants = scenario.ParticipantsAssignedTo(domain.CourseID(*req.CourseIdSelected))
	}

	uiParticipants := toViewParticipants(selectedParticipants, scenario.AllPrioLists())
	uiCourses := ui.NewInBandCourseListUpdate().
		SetUnassignedCount(len(scenario.Unassigned()))

	if req.CourseIdSelected == nil {
		uiCourses.SelectUnassignedEntry()
	}

	for course := range scenario.AllCourses() {
		allocation := scenario.AllocationOf(course.ID)

		var uiCourse ui.Course
		if req.CourseIdSelected != nil && *req.CourseIdSelected == int(course.ID) {
			uiCourse = newSelectedUiCourse(course, allocation)
		} else {
			uiCourse = newUiCourse(course, allocation)
		}

		uiCourses.AppendCourse(uiCourse)
	}

	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "scenario/index", gin.H{"fullPage": false, "participants": uiParticipants, "courseList": uiCourses})

		return
	}

	c.HTML(http.StatusOK, "scenario/index", gin.H{"fullPage": true, "participants": uiParticipants, "courseList": uiCourses})
}
