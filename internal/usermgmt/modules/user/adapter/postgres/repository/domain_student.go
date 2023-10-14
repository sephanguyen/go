package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	db_usermgmt "github.com/manabie-com/backend/internal/usermgmt/pkg/database"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type DomainStudentRepo struct {
	UserRepo            userRepo
	LegacyUserGroupRepo legacyUserGroupRepo
	UserAccessPathRepo  userAccessPathRepo
	UserGroupMemberRepo userGroupMemberRepo
}

type userAccessPathRepo interface {
	UpsertMultiple(ctx context.Context, db database.QueryExecer, userAccessPaths ...entity.DomainUserAccessPath) error
	SoftDeleteByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) error
}

type userGroupMemberRepo interface {
	CreateMultiple(ctx context.Context, db database.QueryExecer, userGroupMembers ...entity.DomainUserGroupMember) error
}

type StudentAttribute struct {
	ID                field.String
	ExternalStudentID field.String
	EnrollmentStatus  field.String
	StudentNote       field.String
	GradeID           field.String
	SchoolID          field.Int32
	OrganizationID    field.String
	CurrentGrade      field.Int16
	ContactPreference field.String
	BillingDate       field.Time // This field is deprecated
}

type Student struct {
	entity.EmptyUser
	User entity.UserProfile

	StudentAttribute

	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Time
}

func NewStudent(student entity.DomainStudent) *Student {
	now := field.NewTime(time.Now())
	externalStudentID := student.ExternalStudentID()
	if field.IsPresent(student.ExternalUserID()) {
		externalStudentID = student.ExternalUserID()
	}
	return &Student{
		StudentAttribute: StudentAttribute{
			ID:                student.UserID(),
			ExternalStudentID: externalStudentID,
			CurrentGrade:      student.CurrentGrade(),
			EnrollmentStatus:  student.EnrollmentStatus(),
			StudentNote:       student.StudentNote(),
			GradeID:           student.GradeID(),
			OrganizationID:    student.OrganizationID(),
			SchoolID:          student.SchoolID(),
			ContactPreference: student.ContactPreference(),
			BillingDate:       field.NewTime(time.Now().Add(0)),
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
}

func (student *Student) UserID() field.String {
	return student.StudentAttribute.ID
}
func (student *Student) CurrentGrade() field.Int16 {
	return student.StudentAttribute.CurrentGrade
}
func (student *Student) EnrollmentStatus() field.String {
	return student.StudentAttribute.EnrollmentStatus
}
func (student *Student) StudentNote() field.String {
	return student.StudentAttribute.StudentNote
}
func (student *Student) GradeID() field.String {
	return student.StudentAttribute.GradeID
}
func (student *Student) SchoolID() field.Int32 {
	return student.StudentAttribute.SchoolID
}
func (student *Student) ContactPreference() field.String {
	return student.StudentAttribute.ContactPreference
}
func (student *Student) OrganizationID() field.String {
	return student.StudentAttribute.OrganizationID
}
func (student *Student) ExternalStudentID() field.String {
	return student.StudentAttribute.ExternalStudentID
}

func (student *Student) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id",
			"student_external_id",
			"current_grade",
			"grade_id",
			"enrollment_status",
			"student_note",
			"school_id",
			"contact_preference",
			"billing_date",
			"updated_at",
			"created_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&student.StudentAttribute.ID,
			&student.StudentAttribute.ExternalStudentID,
			&student.StudentAttribute.CurrentGrade,
			&student.StudentAttribute.GradeID,
			&student.StudentAttribute.EnrollmentStatus,
			&student.StudentAttribute.StudentNote,
			&student.StudentAttribute.SchoolID,
			&student.StudentAttribute.ContactPreference,
			&student.StudentAttribute.BillingDate,
			&student.UpdatedAt,
			&student.CreatedAt,
			&student.DeletedAt,
			&student.StudentAttribute.OrganizationID,
		}
}

func (student *Student) TableName() string {
	return "students"
}

func (repo *DomainStudentRepo) GetByEmails(ctx context.Context, db database.QueryExecer, emails []string) ([]entity.DomainStudent, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainStudentRepo.GetByEmails")
	defer span.End()

	user := &User{}
	userFields, _ := user.FieldMap()

	student := &Student{}
	studentFields, _ := student.FieldMap()

	query := fmt.Sprintf("SELECT u.%s, s.%s FROM public.%s s JOIN public.%s u ON s.student_id = u.user_id "+
		"WHERE (u.email = ANY($1))",
		strings.Join(userFields, ",u."), strings.Join(studentFields, ",s."), student.TableName(), user.TableName())

	rows, err := db.Query(
		ctx,
		query,
		database.TextArray(emails),
	)
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}

	defer rows.Close()

	students := make([]entity.DomainStudent, 0, len(emails))
	for rows.Next() {
		user := &User{}
		_, userFieldValues := user.FieldMap()

		student := &Student{}
		_, studentFieldValues := student.FieldMap()

		fieldValues := []interface{}{}
		fieldValues = append(fieldValues, userFieldValues...)
		fieldValues = append(fieldValues, studentFieldValues...)

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}
		student.User = user
		students = append(students, student)
	}

	return students, nil
}

func (repo *DomainStudentRepo) UpsertMultiple(ctx context.Context, db database.QueryExecer, isEnableUsername bool, studentsToCreate ...aggregate.DomainStudent) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainStudentRepo.UpsertMultiple")
	defer span.End()

	zapLogger := ctxzap.Extract(ctx)

	usersToCreate := entity.Users{}
	legacyUserGroups := entity.LegacyUserGroups{}
	userGroupMembers := entity.DomainUserGroupMembers{}
	userAccessPaths := entity.DomainUserAccessPaths{}
	userAccessPathsWillBeRemoved := entity.DomainUserAccessPaths{}

	for _, studentToCreate := range studentsToCreate {
		usersToCreate = append(usersToCreate, studentToCreate)
		legacyUserGroups = append(legacyUserGroups, studentToCreate.LegacyUserGroups...)
		userGroupMembers = append(userGroupMembers, studentToCreate.UserGroupMembers...)
		userAccessPaths = append(userAccessPaths, studentToCreate.UserAccessPaths...)

		// this is just for backward compatible with the old locations
		if len(studentToCreate.EnrollmentStatusHistories) == 0 {
			userAccessPathsWillBeRemoved = append(userAccessPathsWillBeRemoved, studentToCreate.UserAccessPaths...)
		}
	}
	now := time.Now()
	err := repo.UserAccessPathRepo.SoftDeleteByUserIDs(ctx, db, userAccessPathsWillBeRemoved.UserIDs())
	if err != nil {
		return InternalError{RawError: errors.Wrap(err, "repo.UserAccessPathRepo.SoftDeleteByUserIDs")}
	}

	zapLogger.Debug(
		"--end repo.UserAccessPathRepo.SoftDeleteByUserIDs--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)

	now = time.Now()
	err = repo.UserAccessPathRepo.UpsertMultiple(ctx, db, userAccessPaths...)
	if err != nil {
		return InternalError{RawError: errors.Wrap(err, "repo.UserAccessPathRepo.upsertMultiple")}
	}
	zapLogger.Debug(
		"--end repo.UserAccessPathRepo.upsertMultiple--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)
	now = time.Now()
	err = repo.UserRepo.UpsertMultiple(ctx, db, isEnableUsername, usersToCreate...)
	if err != nil {
		return InternalError{RawError: errors.Wrap(err, "repo.UserRepo.UpsertMultiple")}
	}
	zapLogger.Debug(
		"--end repo.UserRepo.upsertMultiple--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)
	now = time.Now()
	err = repo.LegacyUserGroupRepo.createMultiple(ctx, db, legacyUserGroups...)
	if err != nil {
		return InternalError{RawError: errors.Wrap(err, "repo.LegacyUserGroupRepo.createMultiple")}
	}
	zapLogger.Debug(
		"--end repo.LegacyUserGroupRepo.createMultiple--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)
	now = time.Now()
	err = repo.UserGroupMemberRepo.CreateMultiple(ctx, db, userGroupMembers...)
	if err != nil {
		return InternalError{RawError: errors.Wrap(err, "repo.UserGroupMemberRepo.CreateMultiple")}
	}
	zapLogger.Debug(
		"--end repo.UserGroupMemberRepo.createMultiple--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)

	now = time.Now()
	batch := &pgx.Batch{}
	queueFn := func(b *pgx.Batch, student *Student) {
		fields, values := student.FieldMap()

		// TODO: need check this case in query in this task: https://manabie.atlassian.net/browse/LT-38084
		if student.EnrollmentStatus().String() == "" {
			for index, field := range fields {
				if field == "enrollment_status" {
					fields = append(fields[:index], fields[index+1:]...)
					values = append(values[:index], values[index+1:]...)
					break
				}
			}
		}

		insertPlaceHolders := database.GeneratePlaceholders(len(fields))
		updatePlaceHolders := db_usermgmt.GenerateUpdatePlaceholders(fields)
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT students_pk DO UPDATE SET %s",
			student.TableName(),
			strings.Join(fields, ","),
			insertPlaceHolders,
			updatePlaceHolders,
		)

		b.Queue(stmt, values...)
	}

	for _, studentToCreate := range studentsToCreate {
		databaseStudentToCreate := NewStudent(studentToCreate)

		queueFn(batch, databaseStudentToCreate)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer func() {
		_ = batchResults.Close()
	}()

	for i := 0; i < len(studentsToCreate); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return InternalError{RawError: errors.Wrap(err, "batchResults.Exec")}
		}

		if cmdTag.RowsAffected() != 1 {
			return InternalError{RawError: errors.Errorf("student was not inserted")}
		}
	}
	zapLogger.Debug(
		"--end Student.UpsertMultiple--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)
	return nil
}

func (repo *DomainStudentRepo) GetByIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]entity.DomainStudent, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainStudentRepo.GetByIDs")
	defer span.End()

	student := &Student{}
	fields, _ := student.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s  "+
		"WHERE (student_id = ANY($1))",
		strings.Join(fields, ","), student.TableName())

	rows, err := db.Query(
		ctx,
		query,
		database.TextArray(studentIDs),
	)
	if err != nil {
		return nil, InternalError{RawError: errors.Wrap(err, "db.Query")}
	}

	defer rows.Close()

	var result []entity.DomainStudent
	for rows.Next() {
		item := &Student{}

		_, fieldValues := item.FieldMap()

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{RawError: errors.Wrap(err, "rows.Scan")}
		}

		result = append(result, item)
	}

	return result, nil
}

func (repo *DomainStudentRepo) GetUsersByExternalUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (entity.Users, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainStudentRepo.GetUserByIDs")
	defer span.End()

	stmt := `
		SELECT users.%s
		FROM %s
		JOIN students ON
		        students.student_id = users.user_id
		    AND students.deleted_at IS NULL
		WHERE
			    users.user_external_id = ANY($1)
			AND users.deleted_at IS NULL
	`

	user, err := NewUser(entity.EmptyUser{})
	if err != nil {
		return nil, InternalError{RawError: err}
	}

	fieldNames, _ := user.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ", users."),
		user.TableName(),
	)

	rows, err := db.Query(ctx, stmt, database.TextArray(userIDs))
	if err != nil {
		return nil, InternalError{
			RawError: errors.Wrap(err, "db.Query"),
		}
	}

	defer rows.Close()

	result := make(entity.Users, 0, len(userIDs))
	for rows.Next() {
		item, err := NewUser(entity.EmptyUser{})
		if err != nil {
			return nil, InternalError{RawError: err}
		}

		_, fieldValues := item.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{
				RawError: fmt.Errorf("rows.Scan: %w", err),
			}
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, InternalError{RawError: err}
	}

	return result, nil
}
