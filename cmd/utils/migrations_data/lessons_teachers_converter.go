package migrationsdata

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/scanner"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
)

type LessonTeacherConverter struct {
}

func (s *LessonTeacherConverter) GetHeader() []string {
	return []string{lessonIDCol, teacherIDCol, teacherNameCol, resourcePathCol}
}

func (s *LessonTeacherConverter) GetLineConverted(sc scanner.CSVScanner, orgID string) []string {
	teacherID := sc.Text(teacherIDCol)
	teacherName := sc.Text(teacherNameCol)
	lessonID := sc.Text(lessonIDCol)
	if lessonID == "" {
		reference := sc.Text("Reference1")
		ulidID := support.GenerateULIDFromString(reference)
		lessonID = ulidID.String()
	}
	line := []string{
		lessonID, teacherID, teacherName, orgID,
	}

	return line
}

func (s *LessonTeacherConverter) ValidationData(sc scanner.CSVScanner) []string {
	teacherID := sc.Text(teacherIDCol)
	errs := make([]string, 0, len(sc.Head))
	if teacherID == "" {
		err := fmt.Sprintf("%d teacherID not be valid", sc.GetCurRow())
		errs = append(errs, err)
	}
	return errs
}
