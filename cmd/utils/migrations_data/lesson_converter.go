package migrationsdata

import (
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
)

type LessonConverter struct {
}

func (s *LessonConverter) GetHeader() []string {
	return []string{lessonIDCol, startTimeCol, endTimeCol, centerIDCol, courseIDCol, classIDCol, teachingMethodCol, teachingMediumCol, schedulingStatusCol, isLockedCol, resourcePathCol}
}

func (s *LessonConverter) GetLineConverted(sc scanner.CSVScanner, orgID string) []string {
	centerID := sc.Text(centerIDCol)
	startTime := sc.Text(startTimeCol)
	endTime := sc.Text(endTimeCol)

	teachingMethod := sc.Text(teachingMethodCol)
	teachingMedium := sc.Text(teachingMediumCol)
	ID := sc.Text(lessonIDCol)
	if ID == "" {
		reference := sc.Text("Reference")
		ulidID := support.GenerateULIDFromString(reference)
		ID = ulidID.String()
	}
	courseID := sc.Text(courseIDCol)
	classID := sc.Text(classIDCol)
	schedulingStatus := sc.Text(schedulingStatusCol)
	isLocked := sc.Text(isLockedCol)

	line := []string{
		ID, startTime, endTime, centerID, courseID, classID, teachingMethod, teachingMedium, schedulingStatus, isLocked, orgID,
	}

	return line
}

func (s *LessonConverter) ValidationData(sc scanner.CSVScanner) []string {
	isLocked := sc.Text(isLockedCol)
	errs := make([]string, 0, len(sc.Head))
	_, err := strconv.ParseBool(isLocked)
	if err != nil {
		err := fmt.Sprintf("%d locked value not be valid", sc.GetCurRow())
		errs = append(errs, err)
	}
	return errs
}
