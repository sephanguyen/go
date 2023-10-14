package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type StudentStatRepo struct{}

func (r *StudentStatRepo) Stat(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) (*entities.StudentStat, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentStatRepo.Stat")
	defer span.End()

	fields := database.GetFieldNames(&entities.StudentStat{})
	query := fmt.Sprintf("SELECT %s FROM student_statistics WHERE student_id = $1", strings.Join(fields, ","))
	row := db.QueryRow(ctx, query, &studentID)
	s := new(entities.StudentStat)
	if err := row.Scan(database.GetScanFields(s, fields)...); err != nil {
		return nil, errors.Wrap(err, "rows.Scan")
	}
	return s, nil
}

func (r *StudentStatRepo) Upsert(ctx context.Context, db database.QueryExecer, s *entities.StudentStat) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentStatRepo.Upsert")
	defer span.End()

	now := time.Now()
	s.UpdatedAt.Set(now)
	s.CreatedAt.Set(now)

	fieldNames := []string{"student_id", "created_at", "updated_at"}
	if s.TotalLOFinished.Status != pgtype.Undefined {
		fieldNames = append(fieldNames, "total_lo_finished")
	}
	if s.AdditionalData.Status != pgtype.Undefined {
		fieldNames = append(fieldNames, "additional_data")
	}
	if s.TotalLearningTime.Status != pgtype.Undefined {
		fieldNames = append(fieldNames, "total_learning_time", "last_time_completed_lo")
	}
	placeHolders := database.GeneratePlaceholders(len(fieldNames))

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", s.TableName(), strings.Join(fieldNames, ","), placeHolders)

	var upsertClause string
	indexField := 3
	if s.TotalLOFinished.Status != pgtype.Undefined {
		indexField++
		upsertClause += fmt.Sprintf(", total_lo_finished = $%d", indexField)
	}
	if s.AdditionalData.Status != pgtype.Undefined {
		indexField++
		upsertClause += fmt.Sprintf(", additional_data = $%d", indexField)
	}
	if s.TotalLearningTime.Status != pgtype.Undefined {
		upsertClause += fmt.Sprintf(", total_learning_time = student_statistics.total_learning_time + $%d, last_time_completed_lo = $%d", len(fieldNames)-1, len(fieldNames))
	}

	if upsertClause != "" {
		upsertClause = " ON CONFLICT ON CONSTRAINT statistics_student_un DO UPDATE SET updated_at = $3" + upsertClause
	}

	if _, err := db.Exec(ctx, query+upsertClause, database.GetScanFields(s, fieldNames)...); err != nil {
		return errors.Wrap(err, "r.DB.ExecEx")
	}
	return nil
}

func (r *StudentStatRepo) RetrieveLastTimeCompletedLO(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) (*pgtype.Timestamptz, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentStatRepo.RetrieveLastTimeCompletedLO")
	defer span.End()

	t := new(pgtype.Timestamptz)
	row := db.QueryRow(ctx, "SELECT last_time_completed_lo FROM student_statistics WHERE student_id = $1", &studentID)
	if err := row.Scan(t); err != nil && err != pgx.ErrNoRows {
		return nil, errors.Wrap(err, "row.Scan")
	}
	if t.Status != pgtype.Present {
		return nil, nil
	}
	return t, nil
}
