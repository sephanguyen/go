package commands

import "github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"

type ImportAcademicCalendarPayload struct {
	AcademicWeeks      []*domain.AcademicWeek
	AcademicClosedDays []*domain.AcademicClosedDay
}
