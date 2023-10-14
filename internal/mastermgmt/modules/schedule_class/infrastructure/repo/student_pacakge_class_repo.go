package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/pkg/errors"
)

type StudentPackageClassRepo struct {
}

func (s *StudentPackageClassRepo) GetManyByStudentPackageIDAndStudentIDAndCourseID(ctx context.Context, db database.QueryExecer, queryString string) ([]*StudentPackageClassDTO, map[string]*StudentPackageClassDTO, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentPackageClassRepo.GetManyByStudentPackageIdAndStudentIdAndCourseId")
	defer span.End()
	spc := &StudentPackageClassDTO{}
	fields := database.GetFieldNames(spc)
	query := fmt.Sprintf(
		`SELECT %s 
		FROM %s
		WHERE (student_package_id, student_id, course_id) in (%s)
		and deleted_at is null`,
		strings.Join(fields, ","),
		spc.TableName(),
		queryString,
	)
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	mapSpc := make(map[string]*StudentPackageClassDTO, 0)
	itemList := make([]*StudentPackageClassDTO, 0)
	for rows.Next() {
		item := new(StudentPackageClassDTO)
		if err := rows.Scan(database.GetScanFields(item, fields)...); err != nil {
			return nil, nil, errors.Wrap(err, "rows.Scan")
		}
		mapSpc[s.GetStudentPackageClassID(item.StudentPackageID.String, item.StudentID.String, item.CourseID.String)] = item
		itemList = append(itemList, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, errors.Wrap(err, "rows.Err")
	}
	return itemList, mapSpc, nil
}

func (s *StudentPackageClassRepo) GetStudentPackageClassID(studentPackageID, studentID, courseID string) string {
	return fmt.Sprintf("%s-%s-%s", studentPackageID, studentID, courseID)
}
