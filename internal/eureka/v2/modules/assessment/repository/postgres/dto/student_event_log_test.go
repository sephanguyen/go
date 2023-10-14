package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestFromStudentEventLogEntity(t *testing.T) {
	t.Parallel()

	t.Run("happy case", func(t *testing.T) {
		// arrange
		now := time.Now()
		entity := domain.StudentEventLog{
			EventID:            "id_1",
			EventType:          "type",
			StudentID:          "id_1",
			LearningMaterialID: "id_1",
			Payload: map[string]any{
				"course_id": "abc",
			},
			CreatedAt: now,
		}
		expected := StudentEventLog{
			StudentID:          database.Text("id_1"),
			LearningMaterialID: database.Text("id_1"),
			EventID:            database.Varchar("id_1"),
			EventType:          database.Varchar("type"),
			Payload:            pgtype.JSONB{Bytes: json.RawMessage(`{"course_id":"abc"}`), Status: pgtype.Present},
			CreatedAt:          database.Timestamptz(now),

			// skipped values
			ID: pgtype.Int4{
				Int:    0,
				Status: 1,
			},
			StudyPlanID: pgtype.Text{
				String: "",
				Status: 1,
			},
		}
		// act
		actual, err := FromStudentEventLogEntity(now, entity)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestStudentEventLog_ToEntity(t *testing.T) {
	t.Parallel()

	t.Run("happy case", func(t *testing.T) {
		// arrange
		now := time.Now()
		dto := StudentEventLog{
			StudentID:          database.Text("id_1"),
			LearningMaterialID: database.Text("id_1"),
			EventID:            database.Varchar("id_1"),
			EventType:          database.Varchar("type"),
			Payload:            pgtype.JSONB{Bytes: json.RawMessage(`{"course_id":"abc"}`), Status: pgtype.Present},
			CreatedAt:          database.Timestamptz(now),

			// skipped values
			ID: pgtype.Int4{
				Int:    0,
				Status: 1,
			},
			StudyPlanID: pgtype.Text{
				String: "",
				Status: 1,
			},
		}

		expected := domain.StudentEventLog{
			EventID:            "id_1",
			EventType:          "type",
			StudentID:          "id_1",
			LearningMaterialID: "id_1",
			Payload: map[string]any{
				"course_id": "abc",
			},
			CreatedAt: now,
		}

		// act
		actual, err := dto.ToEntity()

		// assert
		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})
}
