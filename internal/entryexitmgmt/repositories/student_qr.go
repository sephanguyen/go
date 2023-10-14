package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"go.uber.org/multierr"
)

type StudentQRRepo struct {
}

// Create student_qr entity
func (r *StudentQRRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.StudentQR) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentQRRepo.Create")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("err set studentQR: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"qr_id", "resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert StudentQR: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert StudentQR: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

// Upsert insert student QR entity if student_id in rows still not exists.
// If student_id already exists, it will update the row with the given entity value.
func (r *StudentQRRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.StudentQR) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentQRRepo.Update")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("err set studentQR: %w", err)
	}

	stmt := "INSERT INTO student_qr(student_id, qr_url, version, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (student_id) DO UPDATE SET qr_url = $2, version = $3, updated_at = $5;"

	cmd, err := db.Exec(ctx, stmt, []interface{}{&e.StudentID, &e.QRURL, &e.Version, &e.CreatedAt, &e.UpdatedAt}...)
	if err != nil {
		return fmt.Errorf("err upsert StudentQR: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("err upsert StudentQR: %d RowsAffected", cmd.RowsAffected())
	}

	return nil
}

// Find student by id
func (r *StudentQRRepo) FindByID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.StudentQR, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentQRRepo.FindByID")
	defer span.End()

	query := `SELECT student_id, qr_url, version FROM student_qr s WHERE s.student_id = $1`

	studentQr := &entities.StudentQR{}
	if err := database.Select(ctx, db, query, studentID).ScanOne(studentQr); err != nil {
		return nil, fmt.Errorf("err FindByID StudentQR: %w", err)
	}

	return studentQr, nil
}

// Delete qrcode by student id
func (r *StudentQRRepo) DeleteByStudentID(ctx context.Context, db database.QueryExecer, studentID string) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentQRRepo.DeleteByStudentID")
	defer span.End()

	query := `DELETE FROM student_qr s WHERE s.student_id = $1`
	_, err := db.Exec(ctx, query, studentID)
	if err != nil {
		return fmt.Errorf("err DeleteByStudentID StudentQR: %w", err)
	}

	return nil
}
