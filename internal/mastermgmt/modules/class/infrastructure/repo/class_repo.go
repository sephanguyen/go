package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type ClassRepo struct{}

func (c *ClassRepo) GetByID(ctx context.Context, db database.QueryExecer, id string) (*domain.Class, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.GetByID")
	defer span.End()
	class := &Class{}
	fields, values := class.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE class_id = $1 AND deleted_at is null",
		strings.Join(fields, ", "), class.TableName())
	err := db.QueryRow(ctx, query, id).Scan(values...)
	if err != nil {
		return nil, err
	}
	return class.ToClassEntity(), nil
}

func (c *ClassRepo) GetAll(ctx context.Context, db database.QueryExecer) (classes []*domain.ExportingClass, err error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.GetAll")
	defer span.End()

	query := `SELECT class_id, cl."name" , cl."course_id", cl."location_id", cl.school_id, cl.created_at, cl.updated_at, cl.deleted_at, c."name" as course_name, l."name" as location_name
	FROM public."class" cl
	join locations l on l.location_id = cl.location_id 
	join courses c on cl.course_id  = c.course_id 
	where cl.deleted_at is null and l.deleted_at is null and c.deleted_at is null
	`

	rows, err := db.Query(
		ctx,
		query,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*domain.ExportingClass
	for rows.Next() {
		var item domain.ExportingClass
		_, fieldValues := item.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		result = append(result, &item)
	}
	return result, nil
}

func (c *ClassRepo) Insert(ctx context.Context, db database.QueryExecer, classes []*domain.Class) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.Insert")
	defer span.End()
	b := &pgx.Batch{}
	for _, c := range classes {
		class, err := NewClassFromEntity(c)
		if class.ClassID.String == "" {
			err = multierr.Append(err, class.ClassID.Set(idutil.ULIDNow()))
		}
		if err != nil {
			return err
		}
		fields, args := class.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			class.TableName(),
			strings.Join(fields, ","),
			placeHolders)
		b.Queue(query, args...)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < b.Len(); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("class is not inserted")
		}
	}
	return nil
}

func (c *ClassRepo) UpsertClasses(ctx context.Context, db database.Ext, classes []*domain.Class) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.UpsertClasses")
	defer span.End()
	b := &pgx.Batch{}
	for _, c := range classes {
		class, err := NewClassFromEntity(c)
		if err != nil {
			return err
		}
		fields, args := class.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT pk__class DO UPDATE
		SET name = $2, course_id = $3, location_id = $4, updated_at = $7, deleted_at = $8`, class.TableName(), strings.Join(fields, ","), placeHolders)
		b.Queue(query, args...)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < b.Len(); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("class is not upserted")
		}
	}
	return nil
}

func (c *ClassRepo) UpdateClassNameByID(ctx context.Context, db database.QueryExecer, id, name string) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.UpdateClassNameByID")
	defer span.End()
	query := fmt.Sprintf("UPDATE class SET name = $1, updated_at = now() WHERE class_id = $2 AND deleted_at IS NULL")
	cmd, err := db.Exec(ctx, query, name, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (c *ClassRepo) Delete(ctx context.Context, db database.QueryExecer, id string) error {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.Delete")
	defer span.End()
	query := fmt.Sprintf("UPDATE class SET deleted_at = now(), updated_at = now() WHERE class_id = $1")
	cmd, err := db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (c *ClassRepo) RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids []string) (classes []*domain.Class, err error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.RetrieveByIDs")
	defer span.End()

	query := `select class_id, name, location_id from class
	WHERE deleted_at is NULL AND class_id = ANY($1)`

	classesDto := Classes{}
	err = database.Select(ctx, db, query, &ids).ScanAll(&classesDto)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	for _, c := range classesDto {
		cDomain := &domain.Class{
			ClassID:    c.ClassID.String,
			Name:       c.Name.String,
			LocationID: c.LocationID.String,
		}
		classes = append(classes, cDomain)
	}

	return classes, nil
}

func (c *ClassRepo) FindByCourseIDsAndStudentIDs(ctx context.Context, db database.QueryExecer, cs []*domain.ClassWithCourseStudent) ([]*domain.ClassWithCourseStudent, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.FindByCourseIDsAndStudentIDs")
	defer span.End()

	query := `SELECT c.class_id, c.course_id, m.user_id as student_id FROM class c INNER JOIN class_member m on c.class_id = m.class_id
	WHERE c.deleted_at is NULL AND m.deleted_at is NULL AND (c.course_id, m.user_id ) IN (:PlaceHolderVar)`
	inputLen := len(cs)
	studentIDCourseID := make([]string, 0, inputLen) // will like ["($1, $2)", "($3, $4)", ...]
	args := make([]interface{}, 0, inputLen*2)
	numArgs := 0
	for i := 0; i < inputLen; i++ {
		args = append(args, &cs[i].CourseID, &cs[i].StudentID)
		studentIDCourseID = append(studentIDCourseID, fmt.Sprintf("($%d, $%d)", numArgs+1, numArgs+2))
		numArgs += 2
	}
	// placeHolderVar will like ($1, $2), ($3, $4), ($5, $6), ....3
	placeHolderVar := strings.Join(studentIDCourseID, ", ")
	query = strings.ReplaceAll(query, ":PlaceHolderVar", placeHolderVar)
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var result []*domain.ClassWithCourseStudent
	for rows.Next() {
		e := &ClassWithCourseStudent{}
		_, values := e.FieldMap()
		if err := rows.Scan(values...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		result = append(result, NewCourseStudentFromDto(e))
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return result, nil
}
