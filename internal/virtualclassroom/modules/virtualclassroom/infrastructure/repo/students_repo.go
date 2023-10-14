package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type StudentsRepo struct{}

func (s *StudentsRepo) GetStudentByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (domain.Student, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentsRepo.GetStudentByStudentID")
	defer span.End()

	dto := &Student{}
	fields, values := dto.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s 
		WHERE student_id = $1 
		AND deleted_at IS NULL `,
		strings.Join(fields, ", "),
		dto.TableName(),
	)
	err := db.QueryRow(ctx, query, &studentID).Scan(values...)
	if err == pgx.ErrNoRows {
		return domain.Student{}, domain.ErrStudentNotFound
	} else if err != nil {
		return domain.Student{}, errors.Wrap(err, "db.QueryRow: %w")
	}

	return dto.ToStudentDomain(), nil
}

func (s *StudentsRepo) IsUserIDAStudent(ctx context.Context, db database.QueryExecer, userID string) (bool, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentsRepo.IsUserIDAStudent")
	defer span.End()

	_, err := s.GetStudentByStudentID(ctx, db, userID)
	if err == domain.ErrStudentNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
