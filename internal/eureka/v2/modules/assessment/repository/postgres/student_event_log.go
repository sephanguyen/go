package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"

	"github.com/jackc/pgx/v4"
)

type StudentEventLogRepo struct{}

func (s *StudentEventLogRepo) GetManyByEventTypesAndLMs(ctx context.Context, db database.Ext, courseID, userID string, eventTypes, lmIDs []string) ([]domain.StudentEventLog, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEventLogRepo.GetManyByEventTypesAndLMs")
	defer span.End()

	var entity dto.StudentEventLog
	fields, _ := entity.FieldMap()

	buffer := 2
	eventTypePlaceHolders := sliceutils.Map(eventTypes, func(t string) string {
		buffer++
		return fmt.Sprintf("$%d", buffer)
	})
	lmIDPlaceHolders := sliceutils.Map(lmIDs, func(t string) string {
		buffer++
		return fmt.Sprintf("$%d", buffer)
	})
	eventTypeCond := strings.Join(eventTypePlaceHolders, ",")
	lmCond := strings.Join(lmIDPlaceHolders, ",")

	query := fmt.Sprintf(`SELECT %s from %s
		 WHERE student_id = $1
		 AND payload->>'course_id' = $2
		 AND event_type in (%s)
		 AND learning_material_id in (%s);
	`, strings.Join(fields, ", "), entity.TableName(), eventTypeCond, lmCond)

	values := make([]any, 2, len(eventTypes)+len(lmIDs)+2)
	values[0] = userID
	values[1] = courseID
	for _, v := range eventTypes {
		values = append(values, v)
	}
	for _, v := range lmIDs {
		values = append(values, v)
	}

	rows, err := db.Query(ctx, query, values...)
	if err != nil {
		return nil, errors.NewDBError("StudentEventLogRepo.GetManyByEventTypesAndLMs", err)
	}

	return scanStudentEventLogs(rows)
}

func scanStudentEventLogs(rows pgx.Rows) ([]domain.StudentEventLog, error) {
	var eventLogs []domain.StudentEventLog
	dtoHolder := &dto.StudentEventLog{}
	fields, _ := dtoHolder.FieldMap()

	defer rows.Close()
	for rows.Next() {
		e := new(dto.StudentEventLog)
		if err := rows.Scan(database.GetScanFields(e, fields)...); err != nil {
			return nil, errors.NewConversionError("StudentEventLogRepo.scanStudentEventLogs", err)
		}
		a, err := e.ToEntity()
		if err != nil {
			return nil, errors.New("StudentEventLog.ToEntity", err)
		}
		eventLogs = append(eventLogs, a)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewConversionError("StudentEventLogRepo.scanStudentEventLogs", err)
	}

	return eventLogs, nil
}
