package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type StudentSubscriptionRepo struct{}

type ListStudentSubscriptionsArgs struct {
	Limit          uint32
	LessonReportID pgtype.Text
}

func (r *StudentSubscriptionRepo) RetrieveStudentSubscriptions(ctx context.Context, db database.QueryExecer, q *CourseQuery) (entities.StudentSubscriptions, error) {
	ss := &entities.StudentSubscription{}

	query := fmt.Sprintf(`SELECT %s FROM student_subscriptions WHERE start_at IS NOT NULL AND end_at IS NOT NULL AND deleted_at IS NULL`, strings.Join(database.GetFieldNames(ss), ","))

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*entities.StudentSubscription, 0)

	for rows.Next() {
		ss := &entities.StudentSubscription{}
		err := rows.Scan(database.GetScanFields(ss, database.GetFieldNames(ss))...)
		if err != nil {
			return nil, err
		}
		result = append(result, ss)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *StudentSubscriptionRepo) RetrieveStudentSubscriptionID(ctx context.Context, db database.QueryExecer, courseID, studentID, subscriptionID pgtype.Text) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionRepo.RetrieveStudentSubscriptionID")
	defer span.End()

	ss := &entities.StudentSubscription{}
	fields, values := ss.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM lesson_student_subscriptions
		WHERE course_id = $1 AND student_id = $2 AND subscription_id = $3`,
		strings.Join(fields, ","),
	)

	err := db.QueryRow(ctx, query, &courseID, &studentID, &subscriptionID).Scan(values...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("db.QueryRow: %w", err)
	}

	return ss.StudentSubscriptionID.String, nil
}

func (r *StudentSubscriptionRepo) QueueUpsertStudentSubscription(b *pgx.Batch, item *entities.StudentSubscription) {
	fieldNames := database.GetFieldNames(item)
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf(`
	INSERT INTO %s (%s) VALUES (%s) 
	ON CONFLICT ON CONSTRAINT lesson_student_subscriptions_uniq 
	DO UPDATE SET start_at = $5, end_at = $6, updated_at = $8,purchased_slot_total = $10, deleted_at = NULL`, item.TableName(), strings.Join(fieldNames, ","), placeHolders)
	scanFields := database.GetScanFields(item, fieldNames)
	b.Queue(query, scanFields...)
}

func (r *StudentSubscriptionRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, studentSubscriptionItems []*entities.StudentSubscription) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionRepo.BulkUpsert")
	defer span.End()
	b := &pgx.Batch{}
	for _, item := range studentSubscriptionItems {
		r.QueueUpsertStudentSubscription(b, item)
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

// subscriptionID is to check whether the data is new or needed to be replaced
func (r *StudentSubscriptionRepo) DeleteByCourseIDAndStudentID(ctx context.Context, db database.QueryExecer, courseID, studentID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionRepo.DeleteByCourseIDAndStudentID")
	defer span.End()
	sql := `UPDATE lesson_student_subscriptions 
		SET deleted_at = NOW() 
		WHERE course_id = $1 AND student_id = $2 AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, &courseID, &studentID)
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}
