package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LessonMemberRepo struct{}

func (l *LessonMemberRepo) ListStudentsByLessonArgs(ctx context.Context, db database.QueryExecer, args *domain.ListStudentsByLessonArgs) ([]*domain.User, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.ListStudentsByLessonArgs")
	defer span.End()
	fields, _ := (&User{}).FieldMap()

	sql := fmt.Sprintf(
		`SELECT u.%s
        FROM lesson_members AS lm
        INNER JOIN users AS u ON lm.user_id = u.user_id
        WHERE lm.lesson_id = $1
        AND u.user_group = 'USER_GROUP_STUDENT'
        AND (($2::text = '' AND $3::text = '') OR (concat(u.given_name || ' ', u.name), u.user_id) > ($2::text COLLATE "C", $3::text)) 
        AND lm.deleted_at IS NULL
        ORDER BY concat(u.given_name || ' ', u.name) COLLATE "C" ASC, u.user_id ASC
        LIMIT $4`, strings.Join(fields, ", u."))

	users := Users{}
	err := database.Select(ctx, db, sql, &args.LessonID, &args.UserName, &args.UserID, &args.Limit).ScanAll(&users)
	if err != nil {
		return nil, fmt.Errorf("db.Select: %w", err)
	}
	domainUsers := make([]*domain.User, 0, len(users))
	for _, e := range users {
		domainUsers = append(domainUsers, e.ToUserEntity())
	}
	return domainUsers, nil
}

func (l *LessonMemberRepo) SoftDelete(ctx context.Context, db database.QueryExecer, studentID string, lessonIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.SoftDelete")
	defer span.End()
	sql := `UPDATE lesson_members
		SET deleted_at = NOW()
		WHERE user_id = $1 AND lesson_id = ANY($2) AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, &studentID, &lessonIDs)
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (l *LessonMemberRepo) GetLessonIDsByStudentCourseRemovedLocation(ctx context.Context, db database.QueryExecer, courseID, userID string, locationIDs []string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLessonIDsByStudentCourseRemovedLocation")
	defer span.End()
	query := `SELECT lm.lesson_id FROM lesson_members lm 
	JOIN lessons l ON l.lesson_id = lm.lesson_id 
	WHERE lm.course_id = $1 AND lm.user_id = $2
	AND NOT (l.center_id = ANY ($3))
	AND l.deleted_at is null AND lm.deleted_at is null 
	AND l.start_time > NOW()`
	rows, err := db.Query(ctx, query, &courseID, &userID, &locationIDs)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lessonIDs []string
	for rows.Next() {
		var lessonID string
		if err := rows.Scan(&lessonID); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonIDs = append(lessonIDs, lessonID)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return lessonIDs, nil
}

func (l *LessonMemberRepo) FindByID(ctx context.Context, db database.QueryExecer, lessonID, userID string) (*domain.LessonMember, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.FindByID")
	defer span.End()

	lessonMember := &LessonMember{}
	fields, _ := lessonMember.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_id = $1 AND user_id= $2 AND deleted_at is null", strings.Join(fields, ","), lessonMember.TableName())
	err := db.QueryRow(ctx, query, lessonID, userID).Scan(database.GetScanFields(lessonMember, fields)...)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return lessonMember.ToLessonMemberEntity(), nil
}

func (l *LessonMemberRepo) GetLessonMembersInLessons(ctx context.Context, db database.QueryExecer, lessonIDs []string) ([]*domain.LessonMember, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLessonMembersInLessons")
	defer span.End()

	lessonMember := &LessonMember{}
	fields, _ := lessonMember.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE lesson_id = ANY($1) AND deleted_at is null", strings.Join(fields, ","), lessonMember.TableName())
	rows, err := db.Query(ctx, query, lessonIDs)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	lessonMembers := []*domain.LessonMember{}
	for rows.Next() {
		lm := &LessonMember{}
		if err := rows.Scan(database.GetScanFields(lm, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonMembers = append(lessonMembers, lm.ToLessonMemberEntity())
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return lessonMembers, nil
}

func (l *LessonMemberRepo) GetLessonLearnersWithCourseAndNamesByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string, _ bool) (map[string]domain.LessonLearners, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLessonLearnersWithCourseAndNamesByLessonIDs")
	defer span.End()

	lessonMember := &LessonMember{}
	baseQuery := `SELECT lm.lesson_id, lm.user_id, lm.course_id, lm.attendance_status, 
				  lm.attendance_notice, lm.attendance_reason, lm.attendance_note, c.name as "course_name" `
	whereClause := ` WHERE lm.lesson_id = ANY($1) AND lm.deleted_at is null AND c.deleted_at IS NULL `

	baseQuery += ` ,ubi.name FROM lesson_members lm  
			LEFT JOIN courses c ON lm.course_id = c.course_id
			JOIN user_basic_info ubi ON ubi.user_id = lm.user_id `
	whereClause += ` AND ubi.deleted_at is null `

	fields := []string{
		"lesson_id",
		"user_id",
		"course_id",
		"attendance_status",
		"attendance_notice",
		"attendance_reason",
		"attendance_note",
	}
	query := baseQuery + whereClause

	rows, err := db.Query(ctx, query, lessonIDs)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	// fetch results of query
	lessonLearnersMap := make(map[string]domain.LessonLearners, len(lessonIDs))
	var courseName, name pgtype.Text
	scanFields := append(database.GetScanFields(lessonMember, fields), &courseName, &name)

	for rows.Next() {
		if err := rows.Scan(scanFields...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonLearner := lessonMember.ToLessonLearnerEntity()
		lessonLearner.AddCourseName(courseName.String)
		lessonLearner.AddLearnerName(name.String)
		lessonLearnersMap[lessonMember.LessonID.String] = append(lessonLearnersMap[lessonMember.LessonID.String], lessonLearner)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return lessonLearnersMap, nil
}

func (l *LessonMemberRepo) InsertLessonMembers(ctx context.Context, db database.QueryExecer, lessonMembers []*domain.LessonMember) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.InsertLessonMembers")
	defer span.End()
	b := &pgx.Batch{}
	strQuery := `INSERT INTO lesson_members (%s) 
	VALUES (%s) ON CONFLICT ON CONSTRAINT pk__lesson_members DO UPDATE SET updated_at = $3, deleted_at = NULL`
	for _, lessonMember := range lessonMembers {
		l, err := NewLessonMembersFromLessonMemberEntity(lessonMember)
		if err != nil {
			return err
		}
		fieldsToCreate, valuesToCreate := l.FieldMap()

		query := fmt.Sprintf(
			strQuery,
			strings.Join(fieldsToCreate, ","),
			database.GeneratePlaceholders(len(fieldsToCreate)),
		)
		b.Queue(query, valuesToCreate...)
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

func (l *LessonMemberRepo) DeleteLessonMembers(ctx context.Context, db database.QueryExecer, lessonMembers []*domain.LessonMember) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.DeleteLessonMembers")
	defer span.End()
	b := &pgx.Batch{}
	strQuery := `UPDATE lesson_members SET deleted_at = NOW()
	 WHERE lesson_id = $1 AND user_id = $2 `
	for _, lessonMember := range lessonMembers {
		b.Queue(strQuery, lessonMember.LessonID, lessonMember.StudentID)
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()
	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

func (l *LessonMemberRepo) GetLessonsOutOfStudentCourse(ctx context.Context, db database.QueryExecer, sc *user_domain.StudentSubscription) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetLessonsOutOfStudentCourse")
	defer span.End()
	query := `SELECT lm.lesson_id 
			FROM lesson_members lm JOIN lessons l 
				ON l.lesson_id = lm.lesson_id 
			WHERE (
				(l.class_id IS NOT NULL 
					AND NOT EXISTS(SELECT(1) FROM class_member cm 
						JOIN class c ON  c.class_id = cm.class_id 
							WHERE c.course_id = lm.course_id
							AND cm.user_id = lm.user_id 
							AND c.deleted_at is NULL 
							AND cm.deleted_at is NULL)) 
			OR l.class_id IS NULL)
			AND lm.course_id = $1
			AND lm.user_id = $2
			AND l.deleted_at is NULL
			AND lm.deleted_at is NULL
			AND (l.start_time < $3 OR l.start_time > $4)
	`
	rows, err := db.Query(ctx, query,
		sc.CourseID,
		sc.StudentID,
		sc.StartAt,
		sc.EndAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lessonIDs []string
	for rows.Next() {
		var lessonID string
		if err := rows.Scan(&lessonID); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonIDs = append(lessonIDs, lessonID)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return lessonIDs, nil
}

func (l *LessonMemberRepo) DeleteLessonMembersByStartDate(ctx context.Context, db database.QueryExecer, studentID string, classID string, startTime time.Time) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.DeleteLessonMembersByStartDate")
	defer span.End()
	strQuery := `UPDATE lesson_members lm SET deleted_at = NOW()
	 FROM lessons l
	 WHERE l.lesson_id = lm.lesson_id AND lm.user_id = $1 AND l.class_id = $2 AND l.start_time > $3::timestamptz and l.deleted_at is null and lm.deleted_at is null
	 RETURNING lm.lesson_id;`

	rows, err := db.Query(ctx, strQuery, &studentID, &classID, &startTime)
	if err != nil {
		return nil, fmt.Errorf("err db.Exec: %w", err)
	}
	defer rows.Close()
	lessonID := make([]string, 0)
	for rows.Next() {
		var id pgtype.Text
		err = rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lesson id item: %w", err)
		}
		lessonID = append(lessonID, id.String)
	}
	return lessonID, nil
}

func (l *LessonMemberRepo) UpdateLessonMembers(ctx context.Context, db database.QueryExecer, lessonMembers []*domain.UpdateLessonMemberReport) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.UpdateLessonMembers")
	defer span.End()
	b := &pgx.Batch{}
	strQuery := `UPDATE lesson_members SET updated_at = NOW(), 
				attendance_status = $3, attendance_notice = $4,
				attendance_reason = $5, attendance_note = $6
	 WHERE lesson_id = $1 AND user_id = $2 `
	for _, lessonMember := range lessonMembers {
		b.Queue(strQuery, lessonMember.LessonID, lessonMember.StudentID,
			lessonMember.AttendanceStatus, lessonMember.AttendanceNotice,
			lessonMember.AttendanceReason, lessonMember.AttendanceNote)
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()
	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

func (l *LessonMemberRepo) FindByResourcePath(ctx context.Context, db database.QueryExecer, resourcePath string, limit int, offSet int) (*domain.LessonMembers, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.FindByResourcePath")
	defer span.End()
	values := LessonMembers{}

	query := fmt.Sprintf(`
		SELECT l.lesson_id, l.user_id, l.updated_at, l.created_at,  l.deleted_at,
			l.attendance_status, l.attendance_remark, l.course_id, l.attendance_notice, 
			l.attendance_reason, l.attendance_note
		FROM lesson_members l
		WHERE l.deleted_at IS NULL
		AND l.resource_path = $1
		ORDER BY l.lesson_id, l.user_id
		LIMIT $2 OFFSET $3`,
	)
	err := database.Select(ctx, db, query, &resourcePath, &limit, &offSet).ScanAll(&values)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}
	lessonMembers := make(domain.LessonMembers, 0, len(values))
	for _, v := range values {
		lessonMember := v.ToLessonMemberEntity()
		lessonMembers = append(lessonMembers, lessonMember)
	}
	return &lessonMembers, nil
}

func (l *LessonMemberRepo) UpdateLessonMemberNames(ctx context.Context, db database.QueryExecer, lessonMembers []*domain.UpdateLessonMemberName) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.UpdateLessonMemberNames")
	defer span.End()
	b := &pgx.Batch{}
	strQuery := `UPDATE lesson_members SET updated_at = NOW(), 
		user_first_name = $3, user_last_name = $4
		WHERE lesson_id = $1 AND user_id = $2 `

	for _, lessonMember := range lessonMembers {
		b.Queue(strQuery, lessonMember.LessonID, lessonMember.StudentID,
			lessonMember.UserFirstName, lessonMember.UserLastName)
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()
	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}
	return nil
}

func (l *LessonMemberRepo) UpdateLessonMembersFields(ctx context.Context, db database.QueryExecer, lessonMemberDomainList []*domain.LessonMember, updateFields UpdateLessonMemberFields) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.UpdateLessonMembersFields")
	defer span.End()

	b := &pgx.Batch{}
	for _, lessonMember := range lessonMemberDomainList {
		err := l.UpdateFieldsQueue(b, lessonMember, updateFields)
		if err != nil {
			return fmt.Errorf("UpdateLessonMembersFields.UpdateFieldsQueue:%w", err)
		}
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()
	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}

	return nil
}

func (l *LessonMemberRepo) UpdateFieldsQueue(b *pgx.Batch, e *domain.LessonMember, updateFields UpdateLessonMemberFields) error {
	placeHolders := database.GenerateUpdatePlaceholders(updateFields.StringArray(), 3)
	lessonMemberDTO, err := NewLessonMembersFromLessonMemberEntity(e)
	if err != nil {
		return err
	}
	stmt := fmt.Sprintf("UPDATE %s SET updated_at = now(), %s WHERE lesson_id = $1 and user_id = $2 and deleted_at is NULL", lessonMemberDTO.TableName(), placeHolders)

	args := database.GetScanFields(lessonMemberDTO, updateFields.StringArray())
	args = append([]interface{}{database.Text(e.LessonID), database.Text(e.StudentID)}, args...)
	b.Queue(stmt, args...)
	return nil
}
