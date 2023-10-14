package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"

	"github.com/jackc/pgx/v4"
)

type StudentCourseRepo struct{}

func (l *StudentCourseRepo) GetByStudentCourseID(ctx context.Context, db database.QueryExecer, studentID, courseID, locationID, studentPackageID string) (*domain.StudentCourse, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentCourseRepo.GetByStudentCourseID")
	defer span.End()

	sc := &StudentCourse{}
	fields, values := sc.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1 and course_id = $2 and location_id = $3 and student_package_id = $4 and deleted_at is null", strings.Join(fields, ","), sc.TableName())
	err := db.QueryRow(ctx, query, studentID, courseID, locationID, studentPackageID).Scan(values...)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrorNotFound
	} else if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}
	return &domain.StudentCourse{
		StudentID:         sc.StudentID.String,
		CourseID:          sc.CourseID.String,
		LocationID:        sc.LocationID.String,
		CourseSlot:        sc.CourseSlot.Int,
		CourseSlotPerWeek: sc.CourseSlotPerWeek.Int,
	}, nil
}
