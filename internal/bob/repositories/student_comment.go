package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

// QuestionRepo stores
type StudentCommentRepo struct{}

func (repo *StudentCommentRepo) Upsert(ctx context.Context, db database.QueryExecer, comment *entities.StudentComment) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentCommentRepo.Upsert")
	defer span.End()

	fieldNames := []string{"comment_id", "student_id", "coach_id", "comment_content", "updated_at", "created_at"}
	placeHolder := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT student_comments_pk DO UPDATE SET comment_content = $4, updated_at = $5;", comment.TableName(), strings.Join(fieldNames, ","), placeHolder)
	cmdTag, err := db.Exec(ctx, query, &comment.CommentID, &comment.StudentID, &comment.CoachID, &comment.CommentContent, &comment.UpdatedAt, &comment.CreatedAt)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot insert new " + comment.TableName())
	}
	return nil
}

// RetrieveByStudentID returns comment list by studentId. If not specific fields, return all fields
func (repo *StudentCommentRepo) RetrieveByStudentID(ctx context.Context, db database.QueryExecer, studentId pgtype.Text, fields ...string) ([]entities.StudentComment, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentCommentRepo.RetrieveByStudentID")
	defer span.End()

	c := &entities.StudentComment{}

	if len(fields) == 0 {
		fields = database.GetFieldNames(c)
	}

	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1 AND deleted_at IS NULL ORDER BY created_at ASC",
		strings.Join(fields, ","),
		c.TableName(),
	)

	rows, err := db.Query(ctx, selectStmt, &studentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studentComment []entities.StudentComment

	for rows.Next() {
		c := entities.StudentComment{}
		if err := rows.Scan(database.GetScanFields(&c, fields)...); err != nil {
			return nil, err
		}
		studentComment = append(studentComment, c)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return studentComment, nil
}

const deleteStmtPlt = `UPDATE student_comments SET deleted_at = $1, updated_at = $1 WHERE comment_id = ANY($2::_text)`

func (repo *StudentCommentRepo) DeleteStudentComments(ctx context.Context, db database.QueryExecer, cmtIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentCommentRepo.DeleteStudentComments")
	defer span.End()

	var deletedAt pgtype.Timestamptz
	_ = deletedAt.Set(timeutil.Now())

	cmdTag, err := db.Exec(ctx, deleteStmtPlt, deletedAt, cmtIDs)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != int64(len(cmtIDs.Elements)) {
		return ErrUnAffected
	}
	return nil
}
