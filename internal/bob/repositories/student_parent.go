package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

// TeacherRepo provides method to work with teachers entity
type StudentParentRepo struct{}

func (r *StudentParentRepo) queueUpsert(b *pgx.Batch, studentParents []*entities.StudentParent) {
	queueFn := func(b *pgx.Batch, u *entities.StudentParent) {
		fields, values := u.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf(`
				INSERT INTO student_parents (%s) VALUES (%s)
				ON CONFLICT ON CONSTRAINT student_parents_pk
					DO UPDATE SET updated_at = $3, created_at = $4, deleted_at = NULL, relationship = $6`,
			strings.Join(fields, ","),
			placeHolders,
		)

		b.Queue(stmt, values...)
	}

	now := time.Now()
	for _, u := range studentParents {
		if u.ParentID.Status != pgtype.Present {
			continue
		}
		_ = u.UpdatedAt.Set(now)
		_ = u.CreatedAt.Set(now)
		queueFn(b, u)
	}
}

// Upsert also deletes all rows belonging to student before upserting.
func (r *StudentParentRepo) Upsert(ctx context.Context, db database.QueryExecer, studentParents []*entities.StudentParent) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentParentRepo.Upsert")
	defer span.End()

	// get lesson ids
	var studentIDs pgtype.TextArray
	for _, v := range studentParents {
		studentIDs = database.AppendText(studentIDs, v.StudentID)
	}

	b := &pgx.Batch{}
	now := time.Now()
	b.Queue(`UPDATE student_parents SET deleted_at = $1 WHERE student_id = ANY($2)`, now, studentIDs)
	r.queueUpsert(b, studentParents)

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
	}

	return nil
}

func (r *StudentParentRepo) FindParentByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*entities.Parent, error) {
	sp := &entities.StudentParent{}
	p := &entities.Parent{}
	query := fmt.Sprintf(` SELECT p.%s FROM %s p INNER JOIN (
		SELECT parent_id FROM %s sp
		WHERE sp.student_id = ANY($1) AND sp.deleted_at IS NULL 
		GROUP BY sp.parent_id
		) AS sp ON p.parent_id = sp.parent_id
		WHERE p.deleted_at IS NULL;
	`, strings.Join(database.GetFieldNames(p), ", p."), p.TableName(), sp.TableName())

	rows, err := db.Query(ctx, query, studentIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	parents := make([]*entities.Parent, 0)
	for rows.Next() {
		e := &entities.Parent{}
		err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...)
		if err != nil {
			return nil, err
		}
		parents = append(parents, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return parents, nil
}

func (r *StudentParentRepo) GetStudentParents(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*entities.StudentParent, error) {
	sp := &entities.StudentParent{}
	query := fmt.Sprintf(`
		SELECT %s FROM %s WHERE student_id = ANY($1) AND deleted_at IS NULL`, strings.Join(database.GetFieldNames(sp), ","), sp.TableName())

	rows, err := db.Query(ctx, query, studentIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	studentParents := make([]*entities.StudentParent, 0)
	for rows.Next() {
		e := &entities.StudentParent{}
		err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...)
		if err != nil {
			return nil, err
		}
		studentParents = append(studentParents, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return studentParents, nil
}
