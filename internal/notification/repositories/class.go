package repositories

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type ClassRepo struct{}

func (r *ClassRepo) FindCourseIDByClassID(ctx context.Context, db database.QueryExecer, classID string) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.FindByID")
	defer span.End()

	query := `
		SELECT course_id
		FROM class
		WHERE class_id = $1
	`
	courseID := ""
	err := db.QueryRow(ctx, query, database.Text(classID)).Scan(&courseID)
	if err != nil {
		return "", err
	}

	return courseID, nil
}
