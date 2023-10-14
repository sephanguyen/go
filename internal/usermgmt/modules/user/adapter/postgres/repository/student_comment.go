package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

// StudentCommentRepo stores
type StudentCommentRepo struct{}

func (repo *StudentCommentRepo) Upsert(ctx context.Context, db database.QueryExecer, comment *entity.StudentComment) error {
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

func (repo *StudentCommentRepo) DeleteStudentComments(ctx context.Context, db database.QueryExecer, cmtIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentCommentRepo.DeleteStudentComments")
	defer span.End()

	var deletedAt pgtype.Timestamptz
	err := deletedAt.Set(timeutil.Now())
	if err != nil {
		return err
	}

	cmtIDsPgTextArray := database.TextArray(cmtIDs)
	query := "UPDATE student_comments SET deleted_at = $1, updated_at = $1 WHERE comment_id = ANY($2::_text)"
	commandTag, err := db.Exec(ctx, query, deletedAt, cmtIDsPgTextArray)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() != int64(len(cmtIDsPgTextArray.Elements)) {
		return ErrUnAffected
	}
	return nil
}

func (repo *StudentCommentRepo) RetrieveByStudentID(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, fields ...string) ([]entity.StudentComment, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentCommentRepo.RetrieveByStudentID")
	defer span.End()

	sComment := &entity.StudentComment{}

	if len(fields) == 0 {
		fields = database.GetFieldNames(sComment)
	}

	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1 AND deleted_at IS NULL ORDER BY created_at ASC",
		strings.Join(fields, ","),
		sComment.TableName(),
	)

	rows, err := db.Query(ctx, selectStmt, &studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studentComments []entity.StudentComment

	for rows.Next() {
		studentComment := entity.StudentComment{}
		if err := rows.Scan(database.GetScanFields(&studentComment, fields)...); err != nil {
			return nil, err
		}
		studentComments = append(studentComments, studentComment)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return studentComments, nil
}
