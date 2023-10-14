package services

import (
	"context"
	"errors"

	"github.com/manabie-com/backend/internal/golibs/database"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
)

const (
	LOEventType    = "learning_objective"
	CompletedEvent = "completed"
)

type StudentService struct {
	DB                  database.Ext
	LoStudyPlanItemRepo interface {
		UpdateCompleted(ctx context.Context, db database.QueryExecer, studyPlanItemID pgtype.Text, loID pgtype.Text) error
	}
}

func (s *StudentService) HandleStudentEvent(ctx context.Context, req *epb.CreateStudentEventLogsRequest) error {
	for _, log := range req.StudentEventLogs {
		if log.EventType == LOEventType {
			if log.Payload == nil {
				continue
			}

			if log.Payload.Event == CompletedEvent {
				loID := log.Payload.LoId
				if loID == "" {
					return errors.New("missing lo_id in req")
				}
				studyPlanItemID := log.Payload.StudyPlanItemId
				if studyPlanItemID == "" {
					return errors.New("missing study_plan_item_id in req")
				}
				err := s.LoStudyPlanItemRepo.UpdateCompleted(ctx, s.DB, database.Text(studyPlanItemID), database.Text(loID))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func StructToMap(s *types.Struct) map[string]interface{} {
	if s == nil {
		return nil
	}
	m := map[string]interface{}{}
	for k, v := range s.Fields {
		m[k] = decodeValue(v)
	}
	return m
}

func decodeValue(v *types.Value) interface{} {
	switch k := v.Kind.(type) {
	case *types.Value_NullValue:
		return nil
	case *types.Value_NumberValue:
		return k.NumberValue
	case *types.Value_StringValue:
		return k.StringValue
	case *types.Value_BoolValue:
		return k.BoolValue
	case *types.Value_StructValue:
		return StructToMap(k.StructValue)
	case *types.Value_ListValue:
		s := make([]interface{}, len(k.ListValue.Values))
		for i, e := range k.ListValue.Values {
			s[i] = decodeValue(e)
		}
		return s
	default:
		panic("protostruct: unknown kind")
	}
}
