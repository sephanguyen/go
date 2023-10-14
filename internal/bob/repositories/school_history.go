package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

// SchoolHistoryRepo repository
type SchoolHistoryRepo struct{}

type StudentSchoolInfo struct {
	StudentID  pgtype.Text
	SchoolID   pgtype.Text
	SchoolName pgtype.Text
}

func (r *SchoolHistoryRepo) GetCurrentSchoolInfoByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*StudentSchoolInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolHistoryRepo.GetCurrentSchoolInfoByStudentIDs")
	defer span.End()

	e := new(entities.SchoolInfo)
	stmt := fmt.Sprintf(`
		SELECT si.school_id, si.school_name, sh.student_id
		FROM school_history sh
		JOIN %s si USING(school_id)
		WHERE sh.student_id = ANY($1::_TEXT)
			AND sh.is_current = TRUE
			AND si.is_archived = FALSE
			AND sh.deleted_at IS NULL
			AND si.deleted_at IS NULL
	`,
		e.TableName(),
	)

	rows, err := db.Query(ctx, stmt, studentIDs)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	studentSchools := make([]*StudentSchoolInfo, 0)

	for rows.Next() {
		var schoolID, schoolName, studentID pgtype.Text
		if err := rows.Scan(&schoolID, &schoolName, &studentID); err != nil {
			return nil, fmt.Errorf("rows.Scan %w", err)
		}

		studentSchools = append(studentSchools, &StudentSchoolInfo{
			StudentID:  studentID,
			SchoolID:   schoolID,
			SchoolName: schoolName,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err %w", err)
	}

	return studentSchools, nil
}

func (r *SchoolHistoryRepo) FindBySchoolAndStudentIDs(ctx context.Context, db database.QueryExecer, schoolIDs, studentIDs pgtype.TextArray) ([]*entities.SchoolHistory, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolHistoryRepo.FindBySchoolAndStudentIDs")
	defer span.End()

	listSchoolHistory := &entities.SchoolHistories{}
	schoolHistory := &entities.SchoolHistory{}
	fields, _ := schoolHistory.FieldMap()

	stmt := fmt.Sprintf(`
	SELECT %s FROM %s
	WHERE deleted_at is null
	AND is_current = TRUE
	AND school_id = ANY($1::TEXT[])
	AND student_id = ANY($2::TEXT[])
	`, strings.Join(fields, ", "), schoolHistory.TableName())

	if err := database.Select(ctx, db, stmt, schoolIDs, studentIDs).ScanAll(listSchoolHistory); err != nil {
		return nil, err
	}

	return *listSchoolHistory, nil
}

func (r *SchoolHistoryRepo) FindByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*entities.SchoolHistory, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolHistoryRepo.FindByStudentIDs")
	defer span.End()

	listSchoolHistory := &entities.SchoolHistories{}
	schoolHistory := &entities.SchoolHistory{}
	fields, _ := schoolHistory.FieldMap()

	stmt := fmt.Sprintf(`
	SELECT sh.%s FROM %s sh
	JOIN school_info si on sh.school_id = si.school_id 
	AND sh.deleted_at is null
	AND sh.is_current = TRUE
	AND si.deleted_at is null
	AND sh.student_id = ANY($1::TEXT[]);
	`, strings.Join(fields, ", sh."), schoolHistory.TableName())

	if err := database.Select(ctx, db, stmt, studentIDs).ScanAll(listSchoolHistory); err != nil {
		return nil, err
	}

	return *listSchoolHistory, nil
}
