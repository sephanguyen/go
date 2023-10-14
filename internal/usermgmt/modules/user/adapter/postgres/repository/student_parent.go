package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StudentParentRepo provides method to work with student_parent entity
type StudentParentRepo struct{}

func (r *StudentParentRepo) queueUpsert(b *pgx.Batch, studentParents []*entity.StudentParent) {
	queueFn := func(b *pgx.Batch, u *entity.StudentParent) {
		fields, values := u.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf(`
				INSERT INTO student_parents (%s) VALUES (%s)
				ON CONFLICT ON CONSTRAINT student_parents_pk
					DO UPDATE SET updated_at = $3, deleted_at = NULL, relationship = $6`,
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

// Upsert also deletes all rows belonging to student and parent before upserting.
func (r *StudentParentRepo) Upsert(ctx context.Context, db database.QueryExecer, studentParents []*entity.StudentParent) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentParentRepo.Upsert")
	defer span.End()

	var studentIDs pgtype.TextArray
	for _, v := range studentParents {
		studentIDs = database.AppendText(studentIDs, v.StudentID)
	}

	var parentIDs []string
	for _, v := range studentParents {
		parentIDs = append(parentIDs, v.ParentID.String)
	}
	parentIDsString := strings.Join(parentIDs, ",")

	b := &pgx.Batch{}
	now := time.Now()
	b.Queue(`UPDATE student_parents SET deleted_at = $1 WHERE student_id = ANY($2) AND parent_id IN ($3)`, now, studentIDs, parentIDsString)
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

func (r *StudentParentRepo) FindParentIDsFromStudentID(ctx context.Context, db database.QueryExecer, studentID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentParentRepo.FindParentIDsFromStudentID")
	defer span.End()

	parentIDs := []string{}

	query := fmt.Sprintf(`
			SELECT sp.parent_id FROM %s sp 
			WHERE sp.student_id = $1 
				AND sp.deleted_at is null 
		`,
		(&entity.StudentParent{}).TableName())

	rows, err := db.Query(ctx, query, &studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		parentID := ""
		if err := rows.Scan(&parentID); err != nil {
			return nil, err
		}

		parentIDs = append(parentIDs, parentID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return parentIDs, nil
}

// RemoveParentFromStudent Remove parent by parent id and student id
func (r *StudentParentRepo) RemoveParentFromStudent(ctx context.Context, db database.QueryExecer, parentID pgtype.Text, studentID pgtype.Text) error {
	sp := &entity.StudentParent{}
	query := fmt.Sprintf(` UPDATE %s SET deleted_at = $1 WHERE parent_id = $2 AND student_id = $3;
	`, sp.TableName())
	now := time.Now()
	cmdTag, err := db.Exec(ctx, query, now, parentID, studentID)
	if err != nil {
		return errorx.ToStatusError(status.Error(codes.Internal, err.Error()))
	}
	if cmdTag.RowsAffected() != 1 {
		return errorx.ToStatusError(status.Error(codes.InvalidArgument, fmt.Sprintf("student with id %s don't have relationship with parent with id %s", studentID.String, parentID.String)))
	}
	return nil
}

func (r *StudentParentRepo) UpsertParentAccessPathByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentParentRepo.UpsertParentAccessPathByStudentIDs")
	defer span.End()

	batch := &pgx.Batch{}
	// soft delete parent_locations
	stmtSoftDelete := `
	UPDATE user_access_paths SET deleted_at = now()
	WHERE user_id IN (
		select parent_id from student_parents 
		where student_id = ANY($1))`
	batch.Queue(stmtSoftDelete, studentIDs)

	// upsert parent_locations base on their children locations
	stmtInsert := `WITH siblings AS (
SELECT student_id FROM student_parents sp
INNER JOIN (
	SELECT parent_id from student_parents
	WHERE student_id = ANY($1) AND deleted_at IS null) AS tem
ON sp.parent_id = tem.parent_id
)

INSERT INTO user_access_paths 
	(user_id, location_id, created_at, updated_at, resource_path)
SELECT DISTINCT (parent_id), location_id, now(), now(), sp.resource_path FROM student_parents sp 
	INNER JOIN user_access_paths uap 
		ON sp.student_id = uap.user_id AND uap.deleted_at is null
	INNER JOIN (
		SELECT DISTINCT(student_id) FROM siblings) AS tem
	ON uap.user_id = tem.student_id
ON CONFLICT ON CONSTRAINT user_access_paths_pk DO UPDATE SET updated_at = now(), deleted_at = null
`
	batch.Queue(stmtInsert, studentIDs)

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
	}

	return nil
}

func (r *StudentParentRepo) UpsertParentAccessPathByID(ctx context.Context, db database.QueryExecer, parentIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentParentRepo.UpsertParentAccessPathByID")
	defer span.End()

	batch := &pgx.Batch{}
	// soft delete parent_locations
	stmtSoftDelete := `
	UPDATE user_access_paths SET deleted_at = now()
	WHERE user_id = ANY($1)`
	batch.Queue(stmtSoftDelete, parentIDs)

	// upsert parent_locations base on their children locations
	stmtInsert := `
INSERT INTO user_access_paths 
	(user_id, location_id, created_at, updated_at, resource_path)
SELECT DISTINCT (parent_id), location_id, now(), now(), sp.resource_path FROM student_parents sp 
	INNER JOIN user_access_paths uap 
		ON sp.student_id = uap.user_id AND uap.deleted_at is null
	WHERE sp.parent_id = ANY($1) AND sp.deleted_at is null
ON CONFLICT ON CONSTRAINT user_access_paths_pk DO UPDATE SET updated_at = now(), deleted_at = null
`
	batch.Queue(stmtInsert, parentIDs)

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
	}

	return nil
}

func (r *StudentParentRepo) FindStudentParentsByParentID(ctx context.Context, db database.QueryExecer, parentID string) ([]*entity.StudentParent, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentParentRepo.FindStudentParentsByParentID")
	defer span.End()
	studentParentEntity := &entity.StudentParent{}
	studentParentFields := database.GetFieldNames(studentParentEntity)
	query := fmt.Sprintf(`
			SELECT sp.%s FROM %s sp 
			WHERE sp.parent_id = $1 
				AND sp.deleted_at is null 
		`, strings.Join(studentParentFields, ", sp."),
		studentParentEntity.TableName())
	rows, err := db.Query(ctx, query, &parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	studentParents := make([]*entity.StudentParent, 0)
	for rows.Next() {
		studentParent := &entity.StudentParent{}
		scanFields := database.GetScanFields(studentParent, studentParentFields)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, err
		}
		studentParents = append(studentParents, studentParent)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return studentParents, nil
}

func (r *StudentParentRepo) InsertParentAccessPathByStudentID(ctx context.Context, db database.QueryExecer, parentID, studentID string) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentParentRepo.InsertParentAccessPathByStudentID")
	defer span.End()

	stmt := `
INSERT INTO user_access_paths 
	(user_id, location_id, created_at, updated_at, resource_path)
SELECT $1, location_id, now(), now(), uap.resource_path 
FROM user_access_paths uap where uap.user_id = $2
ON CONFLICT ON CONSTRAINT user_access_paths_pk DO UPDATE SET updated_at = now(), deleted_at = null`

	_, err := db.Exec(ctx, stmt, parentID, studentID)
	if err != nil {
		return err
	}

	return nil
}
