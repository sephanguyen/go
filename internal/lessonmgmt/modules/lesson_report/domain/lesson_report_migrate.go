package domain

import (
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
)

type LessonReportMigrateData struct {
	LessonMemberUpdate *lesson_domain.UpdateLessonMemberReport
	Fields             PartnerDynamicFormFieldValues
}
