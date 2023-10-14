package dto

import (
	"encoding/json"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type StudentEventLog struct {
	ID                 pgtype.Int4
	StudentID          pgtype.Text
	StudyPlanID        pgtype.Text
	LearningMaterialID pgtype.Text
	EventID            pgtype.Varchar
	EventType          pgtype.Varchar
	Payload            pgtype.JSONB
	CreatedAt          pgtype.Timestamptz
}

func (s *StudentEventLog) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_event_log_id",
			"student_id",
			"study_plan_id",
			"learning_material_id",
			"event_id",
			"event_type",
			"payload",
			"created_at",
		}, []interface{}{
			&s.ID,
			&s.StudentID,
			&s.StudyPlanID,
			&s.LearningMaterialID,
			&s.EventID,
			&s.EventType,
			&s.Payload,
			&s.CreatedAt,
		}
}

// TableName returns "student_event_logs"
func (s *StudentEventLog) TableName() string {
	return "student_event_logs"
}

func (s *StudentEventLog) ToEntity() (e domain.StudentEventLog, err error) {
	bytes, err := s.Payload.MarshalJSON()
	if err != nil {
		return e, errors.NewConversionError("StudentEventLog.ToEntity", err)
	}
	var payload map[string]any
	err = json.Unmarshal(bytes, &payload)
	if err != nil {
		return e, errors.NewConversionError("StudentEventLog.ToEntity", err)
	}
	e = domain.StudentEventLog{
		EventID:            s.EventID.String,
		EventType:          s.EventType.String,
		LearningMaterialID: s.LearningMaterialID.String,
		StudentID:          s.StudentID.String,
		Payload:            payload,
		CreatedAt:          s.CreatedAt.Time,
	}

	return e, nil
}

func FromStudentEventLogEntity(now time.Time, d domain.StudentEventLog) (StudentEventLog, error) {
	var s = StudentEventLog{}
	database.AllNullEntity(&s)

	if err := multierr.Combine(
		s.EventID.Set(d.EventID),
		s.EventType.Set(d.EventType),
		s.StudentID.Set(d.StudentID),
		s.LearningMaterialID.Set(d.LearningMaterialID),
		s.Payload.Set(d.Payload),
		s.CreatedAt.Set(now),
	); err != nil {
		return s, errors.NewConversionError("multierr.Combine", err)
	}

	return s, nil
}
