package postgres

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type CourseRepo struct {
	DB database.Ext
}

func (repo *CourseRepo) RetrieveByIDs(ctx context.Context, ids []string) ([]*dto.CourseDto, error) {
	ctx, span := interceptors.StartSpan(ctx, "CourseRepo.RetrieveByIDs")
	defer span.End()

	e := &dto.CourseDto{}
	query := fmt.Sprintf(`SELECT c.course_id,c.name,c.icon,c.display_order,c.school_id,c.course_partner_id, c.remarks,c.teaching_method,c.course_type_id,c.is_archived,c.start_date, c.end_date, c.created_at,c.updated_at, c.deleted_at,cb.book_id FROM %s c LEFT JOIN courses_books cb on c.course_id = cb.course_id WHERE c.course_id = ANY($1) AND c.deleted_at IS null ORDER BY c.created_at DESC, c.name ASC, c.course_id DESC`, e.TableName())

	rows, err := repo.DB.Query(ctx, query, &ids)
	if err != nil {
		return nil, errors.NewDBError("eureka_db.Query", err)
	}
	defer rows.Close()

	recordMap := make(map[string]*dto.CourseDto)
	var records []*dto.CourseDto
	for rows.Next() {
		c := &dto.CourseDto{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &c.DisplayOrder, &c.SchoolID, &c.PartnerID, &c.Remarks, &c.TeachingMethod, &c.CourseTypeID, &c.IsArchived, &c.StartDate, &c.EndDate, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &c.BookID); err != nil {
			return nil, errors.NewDBError("rows.Scan", err)
		}

		if _, found := recordMap[c.ID.String]; !found {
			recordMap[c.ID.String] = c
			records = append(records, c)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewDBError("rows.Err", err)
	}

	return records, nil
}
