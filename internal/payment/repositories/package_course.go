package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type PackageCourseRepo struct{}

func (r *PackageCourseRepo) GetByPackageIDForUpdate(ctx context.Context, db database.QueryExecer, packageID string) ([]entities.PackageCourse, error) {
	var packageCourses []entities.PackageCourse
	packageCourse := &entities.PackageCourse{}
	fieldNames, fieldValues := packageCourse.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			package_id = $1
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		packageCourse.TableName(),
	)
	rows, err := db.Query(ctx, stmt, packageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		packageCourses = append(packageCourses, *packageCourse)
	}
	return packageCourses, nil
}

func (r *PackageCourseRepo) GetByPackageIDAndCourseID(ctx context.Context, db database.QueryExecer, packageID string, courseID string) (entities.PackageCourse, error) {
	packageCourse := &entities.PackageCourse{}
	packageCourseFieldNames, packageCourseFieldValues := packageCourse.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			package_id = $1
		AND
			course_id = $2
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(packageCourseFieldNames, ","),
		packageCourse.TableName(),
	)
	row := db.QueryRow(ctx, stmt, packageID, courseID)
	err := row.Scan(packageCourseFieldValues...)
	if err != nil {
		return entities.PackageCourse{}, err
	}
	return *packageCourse, nil
}

func (r *PackageCourseRepo) queueUpsert(b *pgx.Batch, productCourses []*entities.PackageCourse) {
	queueFn := func(b *pgx.Batch, u *entities.PackageCourse) {
		fields, values := u.FieldMap()
		fieldsExceptResourcePath := fields[0 : len(fields)-1]
		valuesExceptResourcePath := values[0 : len(values)-1]
		placeHolders := database.GeneratePlaceholders(len(fieldsExceptResourcePath))
		stmt := "INSERT INTO " + u.TableName() + " (" + strings.Join(fieldsExceptResourcePath, ",") + ") VALUES (" + placeHolders + ");"

		b.Queue(stmt, valuesExceptResourcePath...)
	}

	now := time.Now()
	for _, u := range productCourses {
		_ = u.CreatedAt.Set(now)
		queueFn(b, u)
	}
}

func (r *PackageCourseRepo) Upsert(ctx context.Context, db database.QueryExecer, packageID pgtype.Text, productCourses []*entities.PackageCourse) error {
	ctx, span := interceptors.StartSpan(ctx, "PackageCourseRepo.Upsert")
	defer span.End()

	b := &pgx.Batch{}
	b.Queue(`DELETE FROM package_course WHERE package_id = $1;`, packageID)
	r.queueUpsert(b, productCourses)

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
