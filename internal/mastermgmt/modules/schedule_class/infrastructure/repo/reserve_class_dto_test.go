package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestNewReserveClassFromEntity(t *testing.T) {
	now := time.Now()
	t.Run("success", func(t *testing.T) {
		reserveClassEntity := &domain.ReserveClass{
			ReserveClassID:   "reserve_class_id_01",
			StudentID:        "student_id_01",
			StudentPackageID: "student_package_id_01",
			CourseID:         "course_id_01",
			ClassID:          "class_id_01",
			EffectiveDate:    now,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		expectedReserveClass := &ReserveClassDTO{
			ReserveClassID:   database.Text("reserve_class_id_01"),
			StudentID:        database.Text("student_id_01"),
			StudentPackageID: database.Text("student_package_id_01"),
			CourseID:         database.Text("course_id_01"),
			ClassID:          database.Text("class_id_01"),
			EffectiveDate:    pgtype.Date{Time: now, Status: pgtype.Present},
			CreatedAt:        database.Timestamptz(now),
			UpdatedAt:        database.Timestamptz(now),
			DeletedAt:        pgtype.Timestamptz{Status: pgtype.Null},
		}
		gotReserveClass, err := NewReserveClassFromEntity(reserveClassEntity)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedReserveClass, gotReserveClass)
	})

}
