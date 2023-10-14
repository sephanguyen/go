package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// StudentRepo stores
type StudentRepo struct {
	CreateSchoolFn func(context.Context, database.QueryExecer, *entities.School) error
}

const (
	StudentTrialTime = 0
)

// Create insert student object to db
// Note that to prevent breaking legacy code from calling this function,
//
//	some fields are set to default value if they're nil to avoid not-null
//	constraint error, the list of fields and default values below:
//	  - enrollment_status: "STUDENT_ENROLLMENT_STATUS_ENROLLED"
//	  - student_note: ""
//
//	so if you want nil validation on these fields,
//	make sure to validate it from the callers
func (r *StudentRepo) Create(ctx context.Context, db database.QueryExecer, s *entities.Student) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.Create")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		s.UpdatedAt.Set(now),
		s.CreatedAt.Set(now),
		s.Group.Set(entities.UserGroupStudent),
		s.OnTrial.Set(false),
		s.BillingDate.Set(now.Add(StudentTrialTime)),

		s.User.ID.Set(s.ID.String),
		s.User.UpdatedAt.Set(now),
		s.User.CreatedAt.Set(now),
		s.User.DeviceToken.Set(nil),
		s.User.AllowNotification.Set(true),
	)
	if s.User.ResourcePath.Status != pgtype.Present {
		resourcePath := golibs.ResourcePathFromCtx(ctx)
		s.User.ResourcePath.Set(resourcePath)
	}
	if err != nil {
		return fmt.Errorf("err set entity: %w", err)
	}

	var userID pgtype.Text
	if err := database.InsertReturning(ctx, &s.User, db, "user_id", &userID); err != nil {
		return err
	}

	// This is to prevent breaking legacy code
	if s.EnrollmentStatus.Get() == nil {
		s.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED")
	}
	if s.StudentNote.Get() == nil {
		s.StudentNote.Set("")
	}

	switch {
	case s.School == nil:
		if s.SchoolID.Status == pgtype.Undefined {
			s.SchoolID.Set(nil)
		}
	case s.School.ID.Int != 0: // student selects existed school
		s.SchoolID = s.School.ID
	default:
		// creates new school then assigns created school id to student
		if err := r.CreateSchoolFn(ctx, db, s.School); err != nil {
			return errors.Wrap(err, "r.CreateSchoolFn")
		}
		s.SchoolID = s.School.ID
	}

	var studentID pgtype.Text
	if err := database.InsertReturning(ctx, s, db, "student_id", &studentID); err != nil {
		return err
	}

	group := &entities.UserGroup{}
	err = multierr.Combine(
		group.UserID.Set(s.ID.String),
		group.GroupID.Set(entities.UserGroupStudent),
		group.IsOrigin.Set(true),
		group.Status.Set(entities.UserGroupStatusActive),
		group.CreatedAt.Set(now),
		group.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("err set UserGroup: %w", err)
	}

	cmdTag, err := database.Insert(ctx, group, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert UserGroup: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("%d RowsAffected: %w", cmdTag.RowsAffected(), ErrUnAffected)
	}

	return nil
}

func (r *StudentRepo) CreateEn(ctx context.Context, db database.QueryExecer, s *entities.Student) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.CreateEn")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		s.UpdatedAt.Set(now),
		s.CreatedAt.Set(now),
		s.Group.Set(entities.UserGroupStudent),
		s.OnTrial.Set(false),
		s.BillingDate.Set(now.Add(StudentTrialTime)),
	)
	if s.User.ResourcePath.Status == pgtype.Null {
		resourcePath := golibs.ResourcePathFromCtx(ctx)
		s.User.ResourcePath.Set(resourcePath)
	}
	if err != nil {
		return fmt.Errorf("err set entity: %w", err)
	}

	if s.EnrollmentStatus.Get() == nil {
		s.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED")
	}
	if s.StudentNote.Get() == nil {
		s.StudentNote.Set("")
	}

	switch {
	case s.School == nil:
		if s.SchoolID.Status == pgtype.Undefined {
			s.SchoolID.Set(nil)
		}
	case s.School.ID.Int != 0: // student selects existed school
		s.SchoolID = s.School.ID
	default:
		// creates new school then assigns created school id to student
		if err := r.CreateSchoolFn(ctx, db, s.School); err != nil {
			return errors.Wrap(err, "r.CreateSchoolFn")
		}
		s.SchoolID = s.School.ID
	}

	_, err = database.Insert(ctx, s, db.Exec)
	if err != nil {
		return err
	}

	return nil
}

// StudentProfile including school
type StudentProfile struct {
	Student entities.Student
	School  entities.School
	CoachID pgtype.Text
}

// Retrieve pull student from db by list of uuid
func (r *StudentRepo) Retrieve(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]StudentProfile, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.Retrieve")
	defer span.End()

	s := &entities.Student{}
	studentFields := database.GetFieldNames(s)

	u := &entities.User{}
	userFields := database.GetFieldNames(u)

	school := &entities.School{}
	schoolFields := database.GetFieldNames(school)

	city := &entities.City{}
	cityFields := database.GetFieldNames(city)

	district := &entities.District{}
	districtFields := database.GetFieldNames(district)

	selectFields := make([]string, 0, len(studentFields)+len(userFields)+len(schoolFields)+len(cityFields)+len(districtFields))
	for _, f := range studentFields {
		selectFields = append(selectFields, s.TableName()+"."+f)
	}
	for _, f := range userFields {
		selectFields = append(selectFields, u.TableName()+"."+f)
	}
	for _, f := range schoolFields {
		selectFields = append(selectFields, school.TableName()+"."+f)
	}
	for _, f := range cityFields {
		selectFields = append(selectFields, city.TableName()+"."+f)
	}
	for _, f := range districtFields {
		selectFields = append(selectFields, district.TableName()+"."+f)
	}

	query := `SELECT %s
		FROM students JOIN users ON student_id = user_id
			LEFT JOIN schools ON schools.school_id = students.school_id
			LEFT JOIN cities ON schools.city_id = cities.city_id
			LEFT JOIN districts ON schools.district_id = districts.district_id
		WHERE students.student_id = ANY($1) AND students.deleted_at IS NULL`
	selectStmt := fmt.Sprintf(query, strings.Join(selectFields, ","))

	rows, err := db.Query(ctx, selectStmt, &ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sc := make([]StudentProfile, 0, len(ids.Elements))
	for rows.Next() {
		school := entities.School{
			City:     new(entities.City),
			District: new(entities.District),
		}

		var sp StudentProfile
		sp.School = school

		scanFields := append(database.GetScanFields(&sp.Student, studentFields), database.GetScanFields(&sp.Student.User, userFields)...)
		scanFields = append(scanFields, database.GetScanFields(&sp.School, schoolFields)...)
		scanFields = append(scanFields, database.GetScanFields(sp.School.City, cityFields)...)
		scanFields = append(scanFields, database.GetScanFields(sp.School.District, districtFields)...)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, err
		}

		sc = append(sc, sp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sc, nil
}

func (r *StudentRepo) UpdateStudentProfile(ctx context.Context, db database.QueryExecer, s *entities.Student) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.UpdateStudentProfile")
	defer span.End()

	now := time.Now()
	s.UpdatedAt.Set(now)
	s.User.UpdatedAt = s.UpdatedAt

	updateUserQuery := "UPDATE users SET given_name = $1, name = $2, avatar = $3, updated_at = $4 WHERE user_id = $5"
	_, err := db.Exec(ctx, updateUserQuery, &s.User.GivenName, &s.User.LastName, &s.User.Avatar, &s.User.UpdatedAt, &s.User.ID)
	if err != nil {
		return err
	}

	switch {
	case s.School == nil:
		s.SchoolID.Set(nil)
	case s.School.ID.Int != 0: // student selects existed school
		s.SchoolID = s.School.ID
	default:
		// creates new school then assigns created school id to student
		if err := r.CreateSchoolFn(ctx, db, s.School); err != nil {
			return errors.Wrap(err, "r.CreateSchoolFn")
		}
		s.SchoolID = s.School.ID
	}

	updateStudentQuery := "UPDATE students SET current_grade = $1, target_university = $2, birthday = $3, biography = $4, school_id = $5, updated_at = $6 WHERE student_id = $7"
	_, err = db.Exec(ctx, updateStudentQuery, &s.CurrentGrade, &s.TargetUniversity, &s.Birthday, &s.Biography, &s.SchoolID, &s.UpdatedAt, &s.ID)
	return err
}

func (r *StudentRepo) Find(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) (*entities.Student, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.Find")
	defer span.End()

	e := &entities.Student{}
	fields := database.GetFieldNames(e)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1", strings.Join(fields, ","), e.TableName())
	row := db.QueryRow(ctx, query, &studentID)
	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return e, nil
}

func (r *StudentRepo) FindByPhone(ctx context.Context, db database.QueryExecer, phone pgtype.Text) (*entities.Student, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.FindByPhone")
	defer span.End()

	s := &entities.Student{}
	studentFields := database.GetFieldNames(s)
	userFields := database.GetFieldNames(&s.User)

	selectFields := make([]string, 0, len(studentFields)+len(userFields))
	for _, f := range studentFields {
		selectFields = append(selectFields, s.TableName()+"."+f)
	}

	for _, f := range userFields {
		selectFields = append(selectFields, s.User.TableName()+"."+f)
	}

	selectStmt := fmt.Sprintf("SELECT %s FROM students JOIN users ON student_id = user_id WHERE users.phone_number = $1",
		strings.Join(selectFields, ","))

	scanFields := append(database.GetScanFields(s, studentFields), database.GetScanFields(&s.User, userFields)...)

	err := db.QueryRow(ctx, selectStmt, &phone).Scan(scanFields...)
	if err != nil {
		return nil, errors.Wrap(err, "Scan")
	}

	return s, nil
}

func (r *StudentRepo) Update(ctx context.Context, db database.QueryExecer, s *entities.Student) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.Update")
	defer span.End()

	now := time.Now()
	s.UpdatedAt.Set(now)

	cmdTag, err := database.Update(ctx, s, db.Exec, "student_id")
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update student")
	}

	return nil
}

// UpdateV2 will update both user and student entity
func (r *StudentRepo) UpdateV2(ctx context.Context, db database.QueryExecer, s *entities.Student) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.UpdateV2")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		s.UpdatedAt.Set(now),
		s.User.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("err set entity: %w", err)
	}

	// update user
	cmdTag, err := database.Update(ctx, &s.User, db.Exec, "user_id")
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update user")
	}

	cmdTag, err = database.Update(ctx, s, db.Exec, "student_id")
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return errors.New("cannot update student")
	}

	return nil
}

func (r *StudentRepo) TotalQuestionLimit(ctx context.Context, db database.QueryExecer, id pgtype.Text) (uint, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.TotalQuestionLimit")
	defer span.End()

	var totalQuestionLimit uint
	query := fmt.Sprintf("SELECT total_question_limit FROM %s WHERE student_id = $1", new(entities.Student).TableName())
	if err := db.QueryRow(ctx, query, &id).Scan(&totalQuestionLimit); err != nil {
		return 0, errors.Wrap(err, "db.QueryRowEx")
	}

	return totalQuestionLimit, nil
}

func (r *StudentRepo) GetCountryByStudent(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) (string, error) {
	var country string
	query := "SELECT country FROM users WHERE user_id = $1"
	if err := db.QueryRow(ctx, query, &studentID).Scan(&country); err != nil {
		return "", errors.Wrap(err, "db.QueryRowEx")
	}
	return country, nil
}

func (r *StudentRepo) SoftDelete(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) error {
	sql := `UPDATE students SET deleted_at = NOW(), updated_at = NOW() WHERE student_id = ANY($1)`
	_, err := db.Exec(ctx, sql, &studentIDs)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	return nil
}

type FindStudentFilter struct {
	StudentIDs pgtype.TextArray
	GradeIDs   pgtype.Int2Array
	SchoolID   pgtype.Int4
}

func (r *StudentRepo) FindStudents(ctx context.Context, db database.QueryExecer, filter FindStudentFilter) ([]*entities.Student, error) {
	e := &entities.Student{}
	query := fmt.Sprintf(`SELECT %s FROM %s 
	WHERE ($1::TEXT[] IS NULL OR student_id = ANY($1))  
	AND ($2::INT2[] IS NULL OR current_grade = ANY($2))
	AND school_id = $3::INT4
	AND deleted_at IS NULL`, strings.Join(database.GetFieldNames(e), ", "), e.TableName())

	rows, err := db.Query(ctx, query, filter.StudentIDs, filter.GradeIDs, filter.SchoolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make([]*entities.Student, 0)
	for rows.Next() {
		e := &entities.Student{}
		err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...)
		if err != nil {
			return nil, err
		}
		students = append(students, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return students, nil
}

const findFullStudentProfiles = `SELECT u.%s, s.%s 
	FROM students s JOIN users  u ON s.student_id = u.user_id
	WHERE student_id = ANY($1)`

func (r *StudentRepo) FindStudentProfilesByIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) ([]*entities.Student, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.FindByIDs")
	defer span.End()

	e := &entities.Student{}
	userFields := database.GetFieldNames(&e.User)
	studentFields := database.GetFieldNames(e)
	query := fmt.Sprintf(findFullStudentProfiles, strings.Join(userFields, ", u."), strings.Join(studentFields, ", s. "))

	rows, err := db.Query(ctx, query, &studentIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make([]*entities.Student, 0, len(studentIDs.Elements))
	for rows.Next() {
		student := &entities.Student{}
		scanFields := database.GetScanFields(&student.User, userFields)
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

func (r *StudentRepo) GetStudentsByParentID(ctx context.Context, db database.QueryExecer, parentID pgtype.Text) ([]*entities.User, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.GetStudentsByParentID")
	defer span.End()

	ue := &entities.User{}
	userFields := database.GetFieldNames(ue)

	ae := &entities.AppleUser{}
	appleUserFields := database.GetFieldNames(ae)

	selectFields := make([]string, 0, len(userFields)+len(appleUserFields))

	for _, f := range userFields {
		selectFields = append(selectFields, ue.TableName()+"."+f)
	}
	for _, f := range appleUserFields {
		selectFields = append(selectFields, ae.TableName()+"."+f)
	}

	beautyQuery := `
		SELECT
			%s
		FROM users 		
		JOIN students ON users.user_id = students.student_id AND students.deleted_at IS NULL
		JOIN student_parents  ON student_parents.student_id = students.student_id AND student_parents.parent_id = $1 AND student_parents.deleted_at IS NULL
		JOIN parents ON parents.parent_id = student_parents.parent_id AND parents.deleted_at IS NULL
		LEFT OUTER JOIN apple_users  ON apple_users.user_id = users.user_id
		WHERE users.deleted_at IS NULL
		ORDER BY student_parents.updated_at ASC
	`
	selectStmt := fmt.Sprintf(beautyQuery, strings.Join(selectFields, ","))

	rows, err := db.Query(ctx, selectStmt, &parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*entities.User, 0)
	for rows.Next() {
		user := new(entities.User)
		scanFields := database.GetScanFields(user, database.GetFieldNames(user))
		scanFields = append(scanFields, database.GetScanFields(&entities.AppleUser{}, database.GetFieldNames(&entities.AppleUser{}))...)

		if err := rows.Scan(scanFields...); err != nil {
			return nil, fmt.Errorf("StudentRepo.GetStudentsByParentID: cannot scan value: %w", err)
		}

		users = append(users, user)
	}

	return users, nil
}
