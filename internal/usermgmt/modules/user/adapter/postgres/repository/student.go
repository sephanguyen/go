package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type StudentRepo struct {
}

const (
	StudentTrialTime = 0
)

func (r *StudentRepo) Create(ctx context.Context, db database.QueryExecer, s *entity.LegacyStudent) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		s.UpdatedAt.Set(now),
		s.CreatedAt.Set(now),
		s.Group.Set(entity.UserGroupStudent),
		s.OnTrial.Set(false),
		s.BillingDate.Set(now.Add(StudentTrialTime)),

		s.LegacyUser.ID.Set(s.ID.String),
		s.LegacyUser.UpdatedAt.Set(now),
		s.LegacyUser.CreatedAt.Set(now),
		s.LegacyUser.DeviceToken.Set(nil),
		s.LegacyUser.AllowNotification.Set(true),
	); err != nil {
		return fmt.Errorf("err set entity: %w", err)
	}

	if s.LegacyUser.ResourcePath.Status == pgtype.Null {
		resourcePath := golibs.ResourcePathFromCtx(ctx)
		if err := s.LegacyUser.ResourcePath.Set(resourcePath); err != nil {
			return err
		}
	}

	cmdTag, err := database.Insert(ctx, &s.LegacyUser, db.Exec)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("insert user %d RowsAffected", cmdTag.RowsAffected())
	}

	// This is to prevent breaking legacy code
	if s.EnrollmentStatus.Get() == nil {
		if err := s.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"); err != nil {
			return err
		}
	}
	if s.StudentNote.Get() == nil {
		if err := s.StudentNote.Set(""); err != nil {
			return err
		}
	}

	switch {
	case s.School == nil:
		if s.SchoolID.Status == pgtype.Undefined {
			if err := s.SchoolID.Set(nil); err != nil {
				return err
			}
		}
	case s.School.ID.Int != 0: // student selects existed school
		s.SchoolID = s.School.ID
	}

	cmdTag, err = database.Insert(ctx, s, db.Exec)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("insert student: %d RowsAffected", cmdTag.RowsAffected())
	}

	group := &entity.UserGroup{}
	if err := multierr.Combine(
		group.UserID.Set(s.ID.String),
		group.GroupID.Set(entity.UserGroupStudent),
		group.IsOrigin.Set(true),
		group.Status.Set(entity.UserGroupStatusActive),
		group.CreatedAt.Set(now),
		group.UpdatedAt.Set(now),
		group.ResourcePath.Set(s.LegacyUser.ResourcePath),
	); err != nil {
		return fmt.Errorf("err set UserGroup: %w", err)
	}

	cmdTag, err = database.Insert(ctx, group, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert UserGroup: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("insert users_groups: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *StudentRepo) FindStudentProfilesByIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*entity.LegacyStudent, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.FindByIDs")
	defer span.End()

	e := &entity.LegacyStudent{}
	userFields := database.GetFieldNames(&e.LegacyUser)
	studentFields := database.GetFieldNames(e)
	queryStmt := `SELECT u.%s, s.%s
	FROM students s JOIN users  u ON s.student_id = u.user_id
	WHERE student_id = ANY($1)`
	query := fmt.Sprintf(queryStmt, strings.Join(userFields, ", u."), strings.Join(studentFields, ", s. "))

	rows, err := db.Query(ctx, query, &studentIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make([]*entity.LegacyStudent, 0, len(studentIDs.Elements))
	for rows.Next() {
		student := &entity.LegacyStudent{}
		scanFields := database.GetScanFields(&student.LegacyUser, userFields)
		scanFields = append(scanFields, database.GetScanFields(student, studentFields)...)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, err
		}

		students = append(students, student)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return students, nil
}

func (r *StudentRepo) Find(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) (*entity.LegacyStudent, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.Find")
	defer span.End()

	e := &entity.LegacyStudent{}
	fields := database.GetFieldNames(e)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1", strings.Join(fields, ","), e.TableName())
	row := db.QueryRow(ctx, query, &studentID)
	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return e, nil
}

// Update will update both user and student entity
func (r *StudentRepo) Update(ctx context.Context, db database.QueryExecer, student *entity.LegacyStudent) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.Update")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		student.UpdatedAt.Set(now),
		student.LegacyUser.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("err set entity: %w", err)
	}

	// update user
	cmdTag, err := database.Update(ctx, &student.LegacyUser, db.Exec, "user_id")
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot update user")
	}

	cmdTag, err = database.Update(ctx, student, db.Exec, "student_id")
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("cannot update student")
	}

	return nil
}

func (r *StudentRepo) GetStudentsByParentID(ctx context.Context, db database.QueryExecer, parentID pgtype.Text) ([]*entity.LegacyUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.GetStudentsByParentID")
	defer span.End()

	user := &entity.LegacyUser{}
	userFields := database.GetFieldNames(user)

	appleUser := &entity.AppleUser{}
	appleUserFields := database.GetFieldNames(appleUser)

	selectFields := make([]string, 0, len(userFields)+len(appleUserFields))

	for _, userField := range userFields {
		selectFields = append(selectFields, user.TableName()+"."+userField)
	}

	for _, appleUserField := range appleUserFields {
		selectFields = append(selectFields, appleUser.TableName()+"."+appleUserField)
	}

	query := `
		SELECT
			%s
		FROM users

		JOIN students ON
			users.user_id = students.student_id
				AND
			students.deleted_at IS NULL

		JOIN student_parents ON
			student_parents.student_id = students.student_id
				AND
			student_parents.parent_id = $1
				AND
			student_parents.deleted_at IS NULL

		JOIN parents ON
			parents.parent_id = student_parents.parent_id
				AND
			parents.deleted_at IS NULL

		LEFT OUTER JOIN apple_users ON
			apple_users.user_id = users.user_id

		WHERE
			users.deleted_at IS NULL

		ORDER BY
			student_parents.updated_at ASC
	`
	selectStmt := fmt.Sprintf(query, strings.Join(selectFields, ","))

	rows, err := db.Query(ctx, selectStmt, &parentID)
	if err != nil {
		return nil, fmt.Errorf("err query: %w", err)
	}
	defer rows.Close()

	users := make([]*entity.LegacyUser, 0)
	for rows.Next() {
		user := new(entity.LegacyUser)
		scanFields := database.GetScanFields(user, database.GetFieldNames(user))
		scanFields = append(scanFields, database.GetScanFields(&entity.AppleUser{}, database.GetFieldNames(&entity.AppleUser{}))...)

		if err := rows.Scan(scanFields...); err != nil {
			return nil, fmt.Errorf("StudentRepo.GetStudentsByParentID: cannot scan value: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("StudentRepo.GetStudentsByParentID: err: %w", err)
	}

	return users, nil
}

func (r *StudentRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, students []*entity.LegacyStudent) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, u *entity.LegacyStudent) {
		fields, values := u.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			u.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		b.Queue(stmt, values...)
	}

	b := &pgx.Batch{}
	now := time.Now()

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	for _, student := range students {
		if err := multierr.Combine(
			student.UpdatedAt.Set(now),
			student.CreatedAt.Set(now),
			student.Group.Set(entity.UserGroupStudent),
			student.OnTrial.Set(false),
			student.BillingDate.Set(now.Add(StudentTrialTime)),

			student.LegacyUser.ID.Set(student.ID.String),
			student.LegacyUser.UpdatedAt.Set(now),
			student.LegacyUser.CreatedAt.Set(now),
			student.LegacyUser.DeviceToken.Set(nil),
			student.LegacyUser.AllowNotification.Set(true),
			student.LegacyUser.UserRole.Set(database.Text(string(constant.UserRoleStudent))),
		); err != nil {
			return fmt.Errorf("err set entity: %w", err)
		}
		if student.ResourcePath.Status == pgtype.Null {
			if err := student.ResourcePath.Set(resourcePath); err != nil {
				return err
			}
		}
		queueFn(b, student)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(students); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("student is not inserted")
		}
	}

	return nil
}

func (r *StudentRepo) SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.SoftDelete")
	defer span.End()

	sql := `UPDATE students SET deleted_at = NOW(), updated_at = NOW() WHERE student_id = ANY($1)`
	_, err := db.Exec(ctx, sql, &studentIDs)
	if err != nil {
		return err
	}

	return nil
}

// StudentProfile including school
type StudentProfile struct {
	Student entity.LegacyStudent
	School  entity.School
	Grade   GradeEntity
}

// Retrieve pull student from db by list of uuid
func (r *StudentRepo) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]StudentProfile, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.Retrieve")
	defer span.End()

	student := &entity.LegacyStudent{}
	studentFields := database.GetFieldNames(student)

	user := &entity.LegacyUser{}
	userFields := database.GetFieldNames(user)

	school := &entity.School{}
	schoolFields := database.GetFieldNames(school)

	city := &entity.City{}
	cityFields := database.GetFieldNames(city)

	district := &entity.District{}
	districtFields := database.GetFieldNames(district)

	grade := &GradeEntity{}
	gradeFields := database.GetFieldNames(grade)

	selectFields := make([]string, 0, len(studentFields)+len(userFields)+len(schoolFields)+len(cityFields)+len(districtFields))
	for _, field := range studentFields {
		selectFields = append(selectFields, student.TableName()+"."+field)
	}
	for _, field := range userFields {
		selectFields = append(selectFields, user.TableName()+"."+field)
	}
	for _, field := range schoolFields {
		selectFields = append(selectFields, school.TableName()+"."+field)
	}
	for _, field := range cityFields {
		selectFields = append(selectFields, city.TableName()+"."+field)
	}
	for _, field := range districtFields {
		selectFields = append(selectFields, district.TableName()+"."+field)
	}
	for _, field := range gradeFields {
		selectFields = append(selectFields, grade.TableName()+"."+field)
	}

	query := `SELECT %s
		FROM students JOIN users ON student_id = user_id
			LEFT JOIN schools ON schools.school_id = students.school_id
			LEFT JOIN cities ON schools.city_id = cities.city_id
			LEFT JOIN districts ON schools.district_id = districts.district_id
			LEFT JOIN grade ON grade.grade_id = students.grade_id
		WHERE students.student_id = ANY($1) AND students.deleted_at IS NULL`
	selectStmt := fmt.Sprintf(query, strings.Join(selectFields, ","))

	rows, err := db.Query(ctx, selectStmt, &ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	listStudentProfile := make([]StudentProfile, 0, len(ids.Elements))
	for rows.Next() {
		school := entity.School{
			City:     new(entity.City),
			District: new(entity.District),
		}

		var profile StudentProfile
		profile.School = school

		scanFields := append(database.GetScanFields(&profile.Student, studentFields), database.GetScanFields(&profile.Student.LegacyUser, userFields)...)
		scanFields = append(scanFields, database.GetScanFields(&profile.School, schoolFields)...)
		scanFields = append(scanFields, database.GetScanFields(profile.School.City, cityFields)...)
		scanFields = append(scanFields, database.GetScanFields(profile.School.District, districtFields)...)
		scanFields = append(scanFields, database.GetScanFields(&profile.Grade, gradeFields)...)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, err
		}

		listStudentProfile = append(listStudentProfile, profile)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return listStudentProfile, nil
}
