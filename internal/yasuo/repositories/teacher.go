package repositories

import (
	"context"
	"fmt"
	"strings"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

//TeacherRepo struct
type TeacherRepo struct{}

//Get get teacher that has all school ids
func (r *TeacherRepo) GetTeacherHasSchoolIDs(ctx context.Context, db database.QueryExecer, teacherID string, schoolIds []int32) (*entities_bob.Teacher, error) {
	ctx, span := interceptors.StartSpan(ctx, "TeacherRepo.GetTeacherHasSchoolIDs")
	defer span.End()

	e := &entities_bob.Teacher{}
	fields, values := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE deleted_at IS NULL AND teacher_id = $1 AND school_ids @> $2", strings.Join(fields, ","), e.TableName())

	err := db.QueryRow(ctx, query, &teacherID, &schoolIds).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return e, nil
}

func (r *TeacherRepo) IsInSchool(ctx context.Context, db database.QueryExecer, teacherID string, schoolID int32) (bool, error) {
	totalSchools := 0
	query := "SELECT COUNT(teacher_id) FROM teacher_by_school_id WHERE school_id = $1 AND teacher_id = $2 AND deleted_at IS NULL"

	err := database.Select(ctx, db, query, &schoolID, &teacherID).ScanFields(&totalSchools)
	if err != nil {
		return false, err
	}
	return totalSchools != 0, nil
}
func (r *TeacherRepo) ManyTeacherIsInSchool(ctx context.Context, db database.QueryExecer, teacherIDs pgtype.TextArray, schoolID pgtype.Int4) (bool, error) {
	totalSchools := 0
	query := fmt.Sprintf("SELECT COUNT(teacher_id) FROM teacher_by_school_id WHERE school_id = $1 AND teacher_id = ANY($2)")

	err := database.Select(ctx, db, query, &schoolID, &teacherIDs).ScanFields(&totalSchools)
	if err != nil {
		return false, err
	}

	IDs := map[string]string{}
	for _, id := range teacherIDs.Elements {
		IDs[id.String] = id.String
	}

	return totalSchools == len(IDs), nil
}

func (r *TeacherRepo) JoinSchool(ctx context.Context, db database.QueryExecer, teacherID string, schoolID int32) error {
	query := `UPDATE public.teachers SET
		school_ids = (SELECT array_agg(DISTINCT e) FROM unnest(school_ids || $1) e)
		WHERE teacher_id=$2 AND NOT school_ids @> $1`
	_, err := db.Exec(ctx, query, database.Int4Array([]int32{schoolID}), database.Text(teacherID))
	return err
}

func (r *TeacherRepo) LeaveSchool(ctx context.Context, db database.QueryExecer, teacherID string, schoolID int32) error {
	query := `UPDATE public.teachers SET school_ids = array_remove(school_ids, $1) WHERE teacher_id = $2 AND school_ids @> $3`
	_, err := db.Exec(ctx, query, schoolID, teacherID, database.Int4Array([]int32{schoolID}))
	return err
}
