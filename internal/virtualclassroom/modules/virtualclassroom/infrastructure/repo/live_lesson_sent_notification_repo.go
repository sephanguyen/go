package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
)

type LiveLessonSentNotificationRepo struct{}

func (l *LiveLessonSentNotificationRepo) GetLiveLessonSentNotificationCount(ctx context.Context, db database.QueryExecer, lessonID, interval string) (int32, error) {
	ctx, span := interceptors.StartSpan(ctx, "LiveLessonSentNotificationRepo.GetLiveLessonSentNotificationCount")
	defer span.End()

	liveLessonSentNotificationRecord := &LiveLessonSentNotification{}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s 
				WHERE lesson_id = $1 
				AND sent_at_interval = $2 
				AND deleted_at IS NULL`,
		liveLessonSentNotificationRecord.TableName())

	var total pgtype.Int8
	err := db.QueryRow(ctx, query, lessonID, interval).Scan(&total)
	if err != nil {
		return int32(0), fmt.Errorf("db.Query: %w", err)
	}

	return int32(total.Int), nil
}

func (l *LiveLessonSentNotificationRepo) CreateLiveLessonSentNotificationRecord(ctx context.Context, db database.QueryExecer, lessonID, interval string, sentAt time.Time) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveLessonSentNotificationRepo.CreateLiveLessonSentNotificationRecord")
	defer span.End()

	stmt := "INSERT INTO %s (%s) VALUES (%s);"
	now := time.Now()
	resourcePath := golibs.ResourcePathFromCtx(ctx)
	sentNotificationRecord, err := NewLiveLessonSentNotificationEntity(&domain.LiveLessonSentNotification{
		SentNotificationID: idutil.ULIDNow(),
		LessonID:           lessonID,
		SentAt:             sentAt,
		SentAtInterval:     interval,
		UpdatedAt:          now,
		CreatedAt:          now,
	})
	if err != nil {
		return fmt.Errorf("error creating live lesson notification entity %v", err)
	}

	fields, _ := sentNotificationRecord.FieldMap()
	fields = append(fields, "resource_path")
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertRecordStatement := fmt.Sprintf(stmt, sentNotificationRecord.TableName(), strings.Join(fields, ","), placeHolder)

	values := database.GetScanFields(sentNotificationRecord, fields)
	values = append(values, database.Text(resourcePath))

	_, err = db.Exec(ctx, insertRecordStatement, values...)

	if err != nil {
		return fmt.Errorf("error inserting sent live lesson notification record %v", err)
	}

	return nil
}

func (l *LiveLessonSentNotificationRepo) SoftDeleteLiveLessonSentNotificationRecord(ctx context.Context, db database.QueryExecer, lessonID string) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveLessonSentNotificationRepo.SoftDeleteLiveLessonSentNotificationRecord")
	defer span.End()

	stmt := "UPDATE live_lesson_sent_notifications SET deleted_at = now(), updated_at = now() WHERE lesson_id = $1 AND deleted_at IS NULL;"
	_, err := db.Exec(ctx, stmt, lessonID)
	if err != nil {
		return fmt.Errorf("error deleting live lesson sent notification record %v", err)
	}
	return nil
}
