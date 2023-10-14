package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
)

type LessonReportApprovalRecordRepo struct{}

type ListLessonReportApprovalRecordArgs struct {
	Limit          uint32
	LessonReportID pgtype.Text
}

// Currently unused
// func (l *LessonReportApprovalRecordRepo) FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.LessonReport, error) {
// 	ctx, span := interceptors.StartSpan(ctx, "LessonReportApprovalRecordRepo.FindByID")
// 	defer span.End()

// 	lessonReportApprovalRecordRepo := &entities.LessonReport{}
// 	fields, values := lessonReportApprovalRecordRepo.FieldMap()
// 	query := fmt.Sprintf(`
// 		SELECT %s FROM lesson_report_approval_records
// 		WHERE lesson_report_id = $1
// 			AND deleted_at IS NULL`,
// 		strings.Join(fields, ","),
// 	)

// 	err := db.QueryRow(ctx, query, &id).Scan(values...)
// 	if err != nil {
// 		return nil, fmt.Errorf("db.QueryRow: %w", err)
// 	}

// 	return lessonReportApprovalRecordRepo, nil
// }

func (r *LessonReportApprovalRecordRepo) Create(ctx context.Context, db database.QueryExecer, l *entities.LessonReportApprovalRecord) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonReportApprovalRecordRepo.Create")
	defer span.End()

	script := "INSERT INTO public.lesson_report_approval_records " +
		"(record_id, lesson_report_id, description, approved_by, created_at, updated_at, deleted_at) " +
		"VALUES($1, $2, $3, $4, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL);"
	cmd, err := db.Exec(ctx, script, l.RecordID.String, l.LessonReportID.String, l.Description.String, l.ApprovedBy.String)
	if err != nil {
		return fmt.Errorf("LessonReportApprovalRecordRepo.Create: %w", err)
	}
	if cmd.RowsAffected() != 1 {
		return fmt.Errorf("cannot create approval record")
	}
	return nil
}
