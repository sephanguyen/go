package repo

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/exporter"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/lessonmgmt/constants"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type LessonRepo struct{}

func (l *LessonRepo) getLessonByID(ctx context.Context, db database.QueryExecer, id string) (*Lesson, error) {
	lesson := &Lesson{}
	fields, values := lesson.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM lessons
		WHERE lesson_id = $1 AND deleted_at IS NULL `,
		strings.Join(fields, ","),
	)
	if err := db.QueryRow(ctx, query, &id).Scan(values...); err != nil {
		return nil, err
	}

	return lesson, nil
}

func (l *LessonRepo) GetLessonByID(ctx context.Context, db database.QueryExecer, id string) (*domain.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.GetLessonByID")
	defer span.End()

	lesson, err := l.getLessonByID(ctx, db, id)
	if err != nil {
		return nil, err
	}
	gr := &LessonGroup{}
	if lesson.LessonGroupID.Status == pgtype.Present && lesson.CourseID.Status == pgtype.Present {
		gr, err = (&LessonGroupRepo{}).getByIDAndCourseID(ctx, db, lesson.LessonGroupID.String, lesson.CourseID.String)
		if err != nil {
			return nil, fmt.Errorf("LessonGroupRepo.getByIDAndCourseID: %w", err)
		}
	}

	lessonTeacher, err := (&LessonTeacherRepo{}).GetTeacherIDsByLessonID(ctx, db, lesson.LessonID.String)

	if err != nil {
		return nil, fmt.Errorf("LessonTeacherRepo.GetTeacherIDsByLessonID: %w", err)
	}
	lessonClassroom, err := (&LessonClassroomRepo{}).GetClassroomIDsByLessonID(ctx, db, lesson.LessonID.String)
	if err != nil {
		return nil, fmt.Errorf("LessonClassroomRepo.GetClassroomIDsByLessonID: %w", err)
	}
	lessonMembers, err := (&bob_repo.LessonMemberRepo{}).GetLessonMembersInLesson(ctx, db, lesson.LessonID)
	if err != nil {
		return nil, fmt.Errorf("LessonTeacherRepo.GetTeacherIDsByLessonID: %w", err)
	}

	lessonWithStudent := make([]string, 0, len(lessonMembers)*2)
	for _, lm := range lessonMembers {
		lessonWithStudent = append(lessonWithStudent, lm.LessonID.String, lm.UserID.String)
	}
	reallocation, err := (&ReallocationRepo{}).GetByNewLessonIDAndStudentID(ctx, db, lessonWithStudent)
	if err != nil {
		return nil, fmt.Errorf("ReallocationRepo.GetByNewLessonIDAndStudentID: %w", err)
	}
	reallocationMap := make(map[string]string, len(reallocation))
	for _, r := range reallocation {
		reallocationMap[r.NewLessonID+"-"+r.StudentID] = r.OriginalLessonID
	}

	lessonLearners := domain.LessonLearners{}
	for _, lm := range lessonMembers {
		lessonLearners = append(lessonLearners, &domain.LessonLearner{
			LearnerID:        lm.UserID.String,
			CourseID:         lm.CourseID.String,
			AttendStatus:     domain.StudentAttendStatus(lm.AttendanceStatus.String),
			AttendanceNote:   lm.AttendanceNote.String,
			AttendanceNotice: domain.StudentAttendanceNotice(lm.AttendanceNotice.String),
			AttendanceReason: domain.StudentAttendanceReason(lm.AttendanceReason.String),
			Reallocate: func() *domain.Reallocate {
				uniqueKey := lm.LessonID.String + "-" + lm.UserID.String
				if originalLessonID, exists := reallocationMap[uniqueKey]; exists {
					return &domain.Reallocate{
						OriginalLessonID: originalLessonID,
					}
				}
				return nil
			}(),
		})
	}
	res := domain.NewLesson().
		WithID(lesson.LessonID.String).
		WithName(lesson.Name.String).
		WithLocationID(lesson.CenterID.String).
		WithModificationTime(lesson.CreatedAt.Time, lesson.UpdatedAt.Time).
		WithTimeRange(lesson.StartTime.Time, lesson.EndTime.Time).
		WithSchedulingStatus(domain.LessonSchedulingStatus(lesson.SchedulingStatus.String)).
		WithTeachingMedium(domain.LessonTeachingMedium(lesson.TeachingMedium.String)).
		WithTeachingMethod(domain.LessonTeachingMethod(lesson.TeachingMethod.String)).
		WithLearners(lessonLearners).
		WithTeacherIDs(lessonTeacher).
		WithClassroomIDs(lessonClassroom).
		WithMaterials(database.FromTextArray(gr.MediaIDs)).
		WithCourseID(lesson.CourseID.String).
		WithClassID(lesson.ClassID.String).
		WithSchedulerID(lesson.SchedulerID.String).
		WithIsLocked(lesson.IsLocked.Bool).
		WithZoomLink(lesson.ZoomLink.String).
		WithZoomID(lesson.ZoomID.String).
		WithZoomAccountID(lesson.ZoomOwnerID.String).
		WithClassDoOwnerID(lesson.ClassDoOwnerID.String).
		WithClassDoLink(lesson.ClassDoLink.String).
		WithClassDoRoomID(lesson.ClassDoRoomID.String).
		WithLessonCapacity(lesson.LessonCapacity.Int).
		BuildDraft()

	return res, nil
}

func (l *LessonRepo) fillLessonEntities(ctx context.Context, db database.QueryExecer, lessons []*Lesson) ([]*domain.Lesson, error) {
	lessonIDs := make([]string, 0, len(lessons))
	for _, ls := range lessons {
		lessonIDs = append(lessonIDs, ls.LessonID.String)
	}
	lessonTeachers, err := (&LessonTeacherRepo{}).GetTeacherIDsByLessonIDs(ctx, db, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("LessonTeacherRepo.GetTeacherIDsByLessonID: %w", err)
	}
	lessonTeacherMap := make(map[string][]string)
	for _, lt := range lessonTeachers {
		lessonTeacherMap[lt.LessonID.String] = append(lessonTeacherMap[lt.LessonID.String], lt.TeacherID.String)
	}
	lessonMembers, err := (&bob_repo.LessonMemberRepo{}).GetLessonMembersInLessons(ctx, db, database.TextArray(lessonIDs))
	if err != nil {
		return nil, fmt.Errorf("LessonMemberRepo.GetLessonMembersInLesson: %w", err)
	}
	lessonLearnersMap := make(map[string][]*domain.LessonLearner)
	for _, lm := range lessonMembers {
		lessonLearner := &domain.LessonLearner{
			LearnerID:        lm.UserID.String,
			CourseID:         lm.CourseID.String,
			AttendStatus:     domain.StudentAttendStatus(lm.AttendanceStatus.String),
			AttendanceNote:   lm.AttendanceNote.String,
			AttendanceNotice: domain.StudentAttendanceNotice(lm.AttendanceNotice.String),
			AttendanceReason: domain.StudentAttendanceReason(lm.AttendanceReason.String),
		}
		lessonLearnersMap[lm.LessonID.String] = append(lessonLearnersMap[lm.LessonID.String], lessonLearner)
	}
	res := []*domain.Lesson{}
	for _, lesson := range lessons {
		lessonEntity := domain.NewLesson().
			WithID(lesson.LessonID.String).
			WithName(lesson.Name.String).
			WithLocationID(lesson.CenterID.String).
			WithCourseID(lesson.CourseID.String).
			WithClassID(lesson.ClassID.String).
			WithDeletedTime(database.FromTimestamptz(lesson.DeletedAt)).
			WithSchedulingStatus(domain.LessonSchedulingStatus(lesson.SchedulingStatus.String)).
			WithTeachingMedium(domain.LessonTeachingMedium(lesson.TeachingMedium.String)).
			WithTeachingMethod(domain.LessonTeachingMethod(lesson.TeachingMethod.String)).
			WithLearners(lessonLearnersMap[lesson.LessonID.String]).
			WithTeacherIDs(lessonTeacherMap[lesson.LessonID.String]).
			WithSchedulerID(lesson.SchedulerID.String).
			WithModificationTime(lesson.CreatedAt.Time, lesson.UpdatedAt.Time).
			WithTimeRange(lesson.StartTime.Time, lesson.EndTime.Time).
			WithLessonCapacity(lesson.LessonCapacity.Int).
			WithPreparationTime(lesson.PreparationTime.Int).
			WithBreakTime(lesson.BreakTime.Int).
			BuildDraft()
		res = append(res, lessonEntity)
	}
	return res, nil
}

func (l *LessonRepo) getLessonByIDs(ctx context.Context, db database.QueryExecer, ids []string) ([]*Lesson, error) {
	lesson := &Lesson{}
	fields, _ := lesson.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM lessons WHERE lesson_id = ANY($1)", strings.Join(fields, ","))
	rows, err := db.Query(ctx, query, ids)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	lessons := []*Lesson{}
	for rows.Next() {
		lesson := &Lesson{}
		if err := rows.Scan(database.GetScanFields(lesson, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessons = append(lessons, lesson)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return lessons, nil
}

func (l *LessonRepo) GetLessonByIDs(ctx context.Context, db database.QueryExecer, ids []string) ([]*domain.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.GetLessonByIDs")
	defer span.End()
	lessons, err := l.getLessonByIDs(ctx, db, ids)
	if err != nil {
		return nil, err
	}
	return l.fillLessonEntities(ctx, db, lessons)
}

func (l *LessonRepo) UpsertLessons(ctx context.Context, db database.QueryExecer, recurringLesson *domain.RecurringLesson) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.UpsertLessons")
	defer span.End()

	lessonsDto := make([]*Lesson, 0, len(recurringLesson.Lessons))
	for _, lesson := range recurringLesson.Lessons {
		lessonDto, err := NewLessonFromEntity(lesson)
		if err != nil {
			return nil, err
		}
		if err = lessonDto.Normalize(); err != nil {
			return nil, fmt.Errorf("got error when normalize lesson dto: %w", err)
		}
		if lessonDto.CourseID.Status == pgtype.Present {
			lg := NewLessonGroupFromLessonEntity(lesson, lessonDto.CourseID)
			if err := (&LessonGroupRepo{}).Upsert(ctx, db, lg); err != nil {
				return nil, fmt.Errorf("LessonGroupRepo.insert: %w", err)
			}
			lessonDto.LessonGroupID = lg.LessonGroupID
		}
		lessonsDto = append(lessonsDto, lessonDto)
	}
	// upsert lessons
	if err := l.upsertLessons(ctx, db, lessonsDto); err != nil {
		return nil, fmt.Errorf("failed to upsert lesson: %w", err)
	}
	baseLesson := recurringLesson.GetBaseLesson()
	recurLessonIDs := recurringLesson.GetIDs()
	// upsert teacher
	if err := l.upsertRecurringLessonTeachers(
		ctx,
		db,
		baseLesson.Teachers,
		database.TextArray(recurLessonIDs),
	); err != nil {
		return nil, fmt.Errorf("failed to upsert lesson teachers: %w", err)
	}
	// upsert lesson classrooms
	if err := l.upsertLessonClassrooms(ctx, db, recurLessonIDs, baseLesson.Classrooms); err != nil {
		return nil, fmt.Errorf("got error when upsert lesson classrooms: %w", err)
	}
	// upsert lesson member
	fields := []string{
		"lesson_id",
		"user_id",
		"updated_at",
		"created_at",
		"attendance_status",
		"attendance_remark",
		"course_id",
		"attendance_reason",
		"attendance_notice",
		"attendance_note",
		"user_first_name",
		"user_last_name",
	}
	lessonCourseIDs := recurringLesson.GetLessonCourses()
	for _, ls := range recurringLesson.Lessons {
		members, err := NewLessonMembersFromLessonEntity(ls)
		if err != nil {
			return nil, err
		}
		if err = l.upsertLessonMembers(
			ctx,
			db,
			database.Text(ls.LessonID),
			members,
			fields,
		); err != nil {
			return nil, fmt.Errorf("failed to upsert lesson members: %w", err)
		}
		if len(lessonCourseIDs) > 0 {
			if err = l.upsertLessonCourses(
				ctx,
				db,
				database.Text(ls.LessonID),
				database.TextArray(lessonCourseIDs)); err != nil {
				return nil, fmt.Errorf("failed to upsert lesson courses: %w", err)
			}
		}
	}

	if err := l.upsertReallocateStudents(ctx, db, baseLesson); err != nil {
		return nil, fmt.Errorf("got error when upsert reallocate student: %w", err)
	}

	return recurringLesson.GetIDs(), nil
}

func (l *LessonRepo) upsertLessons(ctx context.Context, db database.QueryExecer, lessons []*Lesson) error {
	b := &pgx.Batch{}
	for _, lesson := range lessons {
		fields, args := lesson.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT lessons_pk DO 
		UPDATE SET deleted_at = NULL, teacher_id = $2, course_id = $3, updated_at = $6, lesson_group_id = $9,
		            lesson_type = $11, status = $12, start_time = $16, end_time = $17, teaching_model = $19,
		            class_id = $20, center_id = $21, teaching_medium = $22, teaching_method = $23, scheduling_status = $24, scheduler_id = $25,
					zoom_link = $27, zoom_owner_id = $28, zoom_id = $29, zoom_occurrence_id = $30, lesson_capacity = $31, preparation_time = $32, break_time = $33,
					classdo_owner_id = $34, classdo_link = $35, classdo_room_id =$36
					`,
			lesson.TableName(),
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
			return fmt.Errorf("lessons is not upserted")
		}
	}
	return nil
}

func (l *LessonRepo) InsertLesson(ctx context.Context, db database.QueryExecer, lesson *domain.Lesson) (*domain.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.InsertLesson")
	defer span.End()

	lessonDto, err := NewLessonFromEntity(lesson)
	if err != nil {
		return nil, err
	}
	if err = lessonDto.Normalize(); err != nil {
		return nil, fmt.Errorf("got error when normalize lesson dto: %w", err)
	}

	// insert lesson group record
	if lessonDto.CourseID.Status == pgtype.Present {
		lg := NewLessonGroupFromLessonEntity(lesson, lessonDto.CourseID)
		if err = (&LessonGroupRepo{}).Insert(ctx, db, lg); err != nil {
			return nil, fmt.Errorf("LessonGroupRepo.insert: %w", err)
		}
		lessonDto.LessonGroupID = lg.LessonGroupID
	}
	// insert lesson record
	if err = l.insertLesson(ctx, db, lessonDto); err != nil {
		return nil, fmt.Errorf("got error when insert lesson record: %w", err)
	}
	// upsert lesson teachers record
	if err = l.upsertLessonTeachers(ctx, db, lessonDto.LessonID, lesson.Teachers); err != nil {
		return nil, fmt.Errorf("got error when upsert lesson teachers: %w", err)
	}

	// upsert lesson members record
	members, err := NewLessonMembersFromLessonEntity(lesson)
	if err != nil {
		return nil, err
	}
	fields := []string{
		"lesson_id",
		"user_id",
		"updated_at",
		"created_at",
		"attendance_status",
		"attendance_remark",
		"course_id",
		"attendance_reason",
		"attendance_notice",
		"attendance_note",
		"user_first_name",
		"user_last_name",
	}
	if err = l.upsertLessonMembers(
		ctx,
		db,
		lessonDto.LessonID,
		members,
		fields,
	); err != nil {
		return nil, fmt.Errorf("got error when upsert lesson members: %w", err)
	}

	// upsert lesson classrooms
	if err = l.upsertLessonClassrooms(ctx, db, []string{lessonDto.LessonID.String}, lesson.Classrooms); err != nil {
		return nil, fmt.Errorf("got error when upsert lesson classrooms: %w", err)
	}

	// upsert reallocate students
	if err = l.upsertReallocateStudents(ctx, db, lesson); err != nil {
		return nil, fmt.Errorf("got error when upsert reallocate student: %w", err)
	}

	// legacy: upsert lesson courses record
	coursesMap := make(map[string]bool)
	var courseIDs []string
	if lesson.TeachingMethod == domain.LessonTeachingMethodGroup {
		if lesson.CourseID != "" {
			coursesMap[lesson.CourseID] = true
			courseIDs = append(courseIDs, lesson.CourseID)
		}
	} else {
		for _, member := range members {
			if _, ok := coursesMap[member.CourseID.String]; !ok {
				courseIDs = append(courseIDs, member.CourseID.String)
				coursesMap[member.CourseID.String] = true
			}
		}
	}

	if len(courseIDs) > 0 {
		if err = l.upsertLessonCourses(ctx, db, lessonDto.LessonID, database.TextArray(courseIDs)); err != nil {
			return nil, fmt.Errorf("got error when upsert lesson courses: %w", err)
		}
	}

	lesson.CreatedAt = lessonDto.CreatedAt.Time
	lesson.UpdatedAt = lessonDto.UpdatedAt.Time
	return lesson, nil
}

func (l *LessonRepo) insertLesson(ctx context.Context, db database.QueryExecer, lesson *Lesson) error {
	if err := lesson.PreInsert(); err != nil {
		return fmt.Errorf("got error when preInsert lesson dto: %w", err)
	}
	fieldNames, args := lesson.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	query := fmt.Sprintf("INSERT INTO lessons (%s) VALUES (%s)",
		strings.Join(fieldNames, ","),
		placeHolders,
	)

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (l *LessonRepo) updateLesson(ctx context.Context, db database.QueryExecer, lesson *Lesson, updatedFields []string) error {
	if err := lesson.PreUpdate(); err != nil {
		return fmt.Errorf("got error when preUpdate lesson dto: %w", err)
	}
	_, err := database.UpdateFields(ctx, lesson, db.Exec, "lesson_id", updatedFields)
	if err != nil {
		return err
	}
	return nil
}

// upsertLessonTeachers also deletes all rows belonging to lessonID before upserting.
func (l *LessonRepo) upsertLessonTeachers(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, teachers domain.LessonTeachers) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.upsertLessonTeachers")
	defer span.End()

	var now pgtype.Timestamptz
	if err := now.Set(time.Now()); err != nil {
		return fmt.Errorf("now.Set(time.Now()): %s", err)
	}

	b := &pgx.Batch{}
	// deletes all rows belonging to lessonID
	b.Queue(`UPDATE lessons_teachers SET deleted_at = $2 WHERE lesson_id = $1`, lessonID, now)
	l.queueUpsertLessonTeacher(b, lessonID, teachers, now)
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

func (l *LessonRepo) upsertRecurringLessonTeachers(ctx context.Context, db database.QueryExecer, teachers domain.LessonTeachers, lessonIDs pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.upsertRecurringLessonTeachers")
	defer span.End()

	var now pgtype.Timestamptz
	if err := now.Set(time.Now()); err != nil {
		return fmt.Errorf("now.Set(time.Now()): %s", err)
	}

	b := &pgx.Batch{}
	// deletes all rows belonging to lessonID
	for _, lessonID := range lessonIDs.Elements {
		b.Queue(`UPDATE lessons_teachers SET deleted_at = $2 WHERE lesson_id = $1`, lessonID, now)
		l.queueUpsertLessonTeacher(b, lessonID, teachers, now)
	}
	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}
	return nil
}

func (l *LessonRepo) queueUpsertLessonTeacher(b *pgx.Batch, lessonID pgtype.Text, teachers domain.LessonTeachers, t pgtype.Timestamptz) {
	queueFn := func(b *pgx.Batch, teacherID pgtype.Text, teacherName pgtype.Text) {
		query := `
			INSERT INTO lessons_teachers (lesson_id, teacher_id, teacher_name) VALUES ($1, $2, $3)
			ON CONFLICT ON CONSTRAINT lessons_teachers_pk 
				DO UPDATE SET created_at = $4, deleted_at = NULL`
		b.Queue(query, lessonID, teacherID, teacherName, t)
	}

	for _, teacher := range teachers {
		queueFn(b, database.Text(teacher.TeacherID), database.Text(teacher.Name))
	}
}

// upsertLessonMembers also deletes all rows belonging to lessonID before upserting.
func (l *LessonRepo) upsertLessonMembers(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, members LessonMembers, upsertFields []string) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.upsertLessonMembers")
	defer span.End()

	var now pgtype.Timestamptz
	err := now.Set(time.Now())
	if err != nil {
		return fmt.Errorf("now.Set(time.Now()): %w", err)
	}

	b := &pgx.Batch{}
	// deletes all rows belonging to lessonID
	b.Queue(
		fmt.Sprintf(`UPDATE lesson_members SET deleted_at = $2 WHERE lesson_id = $1`),
		lessonID,
		now,
	)

	for i := range members {
		if err := members[i].PreUpsert(); err != nil {
			return fmt.Errorf("could not preupsert lesson member %s", members[i].UserID.String)
		}
		l.queueUpsertLessonMember(b, members[i], upsertFields)
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

func (l *LessonRepo) queueUpsertLessonMember(b *pgx.Batch, e *LessonMember, upsertFields []string) {
	args := database.GetScanFields(e, upsertFields)
	placeHolders := database.GeneratePlaceholders(len(upsertFields))
	updatePlaceHolders := database.GenerateUpdatePlaceholders(upsertFields, 1)
	sql := fmt.Sprintf("INSERT INTO lesson_members (%s) "+
		"VALUES (%s) ON CONFLICT ON CONSTRAINT pk__lesson_members DO "+
		"UPDATE SET deleted_at = NULL, %s ",
		strings.Join(upsertFields, ", "),
		placeHolders,
		updatePlaceHolders,
	)

	b.Queue(sql, args...)
}

// upsertLessonCourses also deletes all rows belonging to lessonID before upserting.
func (l *LessonRepo) upsertLessonCourses(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, courseIDs pgtype.TextArray) error {
	b := &pgx.Batch{}
	b.Queue(`UPDATE lessons_courses SET deleted_at = now() WHERE lesson_id = $1`, lessonID)
	queueFn := func(b *pgx.Batch, courseID pgtype.Text) {
		query := `
			INSERT INTO lessons_courses (lesson_id, course_id) VALUES ($1, $2)
			ON CONFLICT ON CONSTRAINT lessons_courses_pk
				DO UPDATE SET deleted_at = NULL`
		b.Queue(query, lessonID, courseID)
	}
	for _, courseID := range courseIDs.Elements {
		queueFn(b, courseID)
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

func (l *LessonRepo) UpdateLesson(ctx context.Context, db database.QueryExecer, lesson *domain.Lesson) (*domain.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.UpdateLesson")
	defer span.End()

	currentLesson, err := l.getLessonByID(ctx, db, lesson.LessonID)
	if err != nil {
		return nil, fmt.Errorf("could not get current lesson: %w", err)
	}

	lessonDto, err := NewLessonFromEntity(lesson)
	if err != nil {
		return nil, err
	}
	if err = lessonDto.Normalize(); err != nil {
		return nil, fmt.Errorf("got error when normalize lesson dto: %w", err)
	}

	// upsert media ids
	// now we'll still keep lesson group if course not present
	if lessonDto.CourseID.Status == pgtype.Present {
		lg := NewLessonGroupFromLessonEntity(lesson, lessonDto.CourseID)
		if currentLesson.CourseID.String == lg.CourseID.String {
			lg.LessonGroupID = currentLesson.LessonGroupID // just update media for current lesson group
			if err = (&LessonGroupRepo{}).updateMedia(ctx, db, lg); err != nil {
				return nil, fmt.Errorf("LessonGroupRepo.updateMedia: %w", err)
			}
		} else { // create new a lesson gr
			if err = (&LessonGroupRepo{}).Insert(ctx, db, lg); err != nil {
				return nil, fmt.Errorf("LessonGroupRepo.insert: %w", err)
			}
		}
		lessonDto.LessonGroupID = lg.LessonGroupID
	}

	updatedFields := []string{
		"teacher_id",
		"course_id",
		"updated_at",
		"lesson_group_id",
		"lesson_type",
		"status",
		"start_time",
		"end_time",
		"teaching_model",
		"center_id",
		"teaching_medium",
		"teaching_method",
		"scheduling_status",
		"class_id",
		"scheduler_id",
		"zoom_link",
		"zoom_owner_id",
		"zoom_id",
		"lesson_capacity",
		"classdo_owner_id",
		"classdo_link",
		"classdo_room_id",
		"preparation_time",
		"break_time",
	}

	// update lesson
	if err = l.updateLesson(ctx, db, lessonDto, updatedFields); err != nil {
		return nil, fmt.Errorf("got error when update lesson record: %w", err)
	}

	// upsert lesson teachers
	if err = l.upsertLessonTeachers(ctx, db, lessonDto.LessonID, lesson.Teachers); err != nil {
		return nil, fmt.Errorf("got error when upsert lesson teachers: %w", err)
	}

	// upsert lesson classrooms
	if err = l.upsertLessonClassrooms(ctx, db, []string{lessonDto.LessonID.String}, lesson.Classrooms); err != nil {
		return nil, fmt.Errorf("got error when upsert lesson classrooms: %w", err)
	}

	// upsert lesson members
	members, err := NewLessonMembersFromLessonEntity(lesson)
	if err != nil {
		return nil, err
	}
	fields := []string{
		"lesson_id",
		"user_id",
		"updated_at",
		"created_at",
		"attendance_status",
		"attendance_remark",
		"course_id",
		"attendance_reason",
		"attendance_notice",
		"attendance_note",
		"user_first_name",
		"user_last_name",
	}
	if err = l.upsertLessonMembers(
		ctx,
		db,
		lessonDto.LessonID,
		members,
		fields,
	); err != nil {
		return nil, fmt.Errorf("got error when upsert lesson members: %w", err)
	}

	// upsert reallocate students
	if err = l.upsertReallocateStudents(ctx, db, lesson); err != nil {
		return nil, fmt.Errorf("got error when upsert reallocate student: %w", err)
	}

	// legacy: upsert lesson courses record
	coursesMap := make(map[string]bool)
	var courseIDs []string
	if lesson.TeachingMethod == domain.LessonTeachingMethodGroup {
		if lesson.CourseID != "" {
			coursesMap[lesson.CourseID] = true
			courseIDs = append(courseIDs, lesson.CourseID)
		}
	} else {
		for _, member := range members {
			if _, ok := coursesMap[member.CourseID.String]; !ok {
				courseIDs = append(courseIDs, member.CourseID.String)
				coursesMap[member.CourseID.String] = true
			}
		}
	}
	if len(courseIDs) > 0 {
		if err = l.upsertLessonCourses(ctx, db, lessonDto.LessonID, database.TextArray(courseIDs)); err != nil {
			return nil, fmt.Errorf("got error when upsert lesson courses: %w", err)
		}
	}

	lesson.UpdatedAt = lessonDto.UpdatedAt.Time
	return lesson, nil
}

func (l *LessonRepo) upsertReallocateStudents(ctx context.Context, db database.QueryExecer, lesson *domain.Lesson) error {
	reallocateStudents, err := NewReallocateStudentFromLessonEntity(lesson)
	if err != nil {
		return err
	}

	if len(reallocateStudents) > 0 {
		if err = (&ReallocationRepo{}).upsertReallocation(
			ctx,
			db,
			lesson.LessonID,
			reallocateStudents,
		); err != nil {
			return fmt.Errorf("got error when upsert reallocate students: %w", err)
		}
	}
	return nil
}

func (l *LessonRepo) UpdateLessonSchedulingStatus(ctx context.Context, db database.Ext, lesson *domain.Lesson) (*domain.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.UpdateLessonSchedulingStatus")
	defer span.End()

	lessonDto, err := NewLessonFromEntity(lesson)
	if err != nil {
		return nil, err
	}
	if err = lessonDto.Normalize(); err != nil {
		return nil, fmt.Errorf("got error when normalize lesson dto: %w", err)
	}

	// update lesson
	if err = l.updateLesson(ctx, db, lessonDto,
		[]string{
			"scheduling_status",
			"updated_at",
		},
	); err != nil {
		return nil, fmt.Errorf("got error when update lesson status: %w", err)
	}

	lesson.UpdatedAt = lessonDto.UpdatedAt.Time
	return lesson, nil
}

func (l *LessonRepo) UpdateSchedulingStatus(ctx context.Context, db database.QueryExecer, lessonStatus map[string]domain.LessonSchedulingStatus) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.UpdateSchedulingStatus")
	defer span.End()
	b := &pgx.Batch{}
	now := time.Now()
	for lessonID, schedulingStatus := range lessonStatus {
		query := "UPDATE lessons SET scheduling_status = $1, updated_at = $2 WHERE lesson_id = $3"
		b.Queue(query, schedulingStatus, now, lessonID)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < b.Len(); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("cannot update lesson status")
		}
	}
	return nil
}

func (l *LessonRepo) UpdateSchedulerID(ctx context.Context, db database.Ext, lessonIDs []string, schedulerID string) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.UpdateSchedulerID")
	defer span.End()

	lessonDto := &Lesson{}
	b := &pgx.Batch{}

	for _, v := range lessonIDs {
		if err := lessonDto.PreUpdateSchedulingID(v, schedulerID); err != nil {
			return fmt.Errorf("got error when PreUpdateSchedulerID lesson dto: %w", err)
		}

		if err := lessonDto.PreUpdate(); err != nil {
			return fmt.Errorf("got error when PreUpdate lesson dto: %w", err)
		}

		query := fmt.Sprintf(`UPDATE %s SET scheduler_id = $1, updated_at = $2 WHERE lesson_id = $3`,
			lessonDto.TableName())

		b.Queue(query, lessonDto.SchedulerID, lessonDto.UpdatedAt, lessonDto.LessonID)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < b.Len(); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("lessons is not update schedulerId")
		}
	}

	return nil
}

func (l *LessonRepo) LockLesson(ctx context.Context, db database.Ext, lessonIds []string) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.LockLesson")
	defer span.End()

	for _, lessonId := range lessonIds {
		dto := &Lesson{}
		if err = dto.PreLock(lessonId); err != nil {
			return err
		}
		// update lesson
		if err = l.updateLesson(ctx, db, dto,
			[]string{
				"is_locked",
				"updated_at",
			},
		); err != nil {
			return fmt.Errorf("got error when update is_locked in lessonId %s: %w", lessonId, err)
		}
	}

	return nil
}

func (l *LessonRepo) Delete(ctx context.Context, db database.QueryExecer, lessonIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.Delete")
	defer span.End()

	query := "UPDATE lessons SET deleted_at = now(), updated_at = now() WHERE lesson_id = ANY($1) AND deleted_at IS NULL"
	_, err := db.Exec(ctx, query, &lessonIDs)

	return err
}

func (l *LessonRepo) getLessonBySchedulerID(ctx context.Context, db database.QueryExecer, schedulerID string) ([]*Lesson, error) {
	fields, _ := (&Lesson{}).FieldMap()
	query := fmt.Sprintf("SELECT %s FROM lessons WHERE deleted_at is null and scheduler_id = $1 order by start_time asc", strings.Join(fields, ","))
	rows, err := db.Query(ctx, query, schedulerID)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	lessons := []*Lesson{}
	for rows.Next() {
		lesson := &Lesson{}
		if err := rows.Scan(database.GetScanFields(lesson, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessons = append(lessons, lesson)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return lessons, nil
}

func (l *LessonRepo) GetLessonBySchedulerID(ctx context.Context, db database.QueryExecer, schedulerID string) ([]*domain.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.GetLessonBySchedulerID")
	defer span.End()
	lessons, err := l.getLessonBySchedulerID(ctx, db, schedulerID)
	if err != nil {
		return nil, fmt.Errorf("l.getLessonBySchedulerID: %w", err)
	}
	lessonIDs := make([]string, 0, len(lessons))
	lessonGroupMap := make(map[string]string)
	var lessonGroupIDs []string
	for _, ls := range lessons {
		lessonIDs = append(lessonIDs, ls.LessonID.String)
		if ls.LessonGroupID.Status == pgtype.Present {
			groupID := ls.LessonGroupID.String
			lessonGroupMap[groupID] = ls.LessonID.String
			lessonGroupIDs = append(lessonGroupIDs, groupID)
		}
	}
	lessonTeachers, err := (&LessonTeacherRepo{}).GetTeacherIDsByLessonIDs(ctx, db, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("LessonTeacherRepo.GetTeacherIDsByLessonID: %w", err)
	}
	lessonTeacherMap := make(map[string][]string)
	for _, lt := range lessonTeachers {
		lessonTeacherMap[lt.LessonID.String] = append(lessonTeacherMap[lt.LessonID.String], lt.TeacherID.String)
	}
	lessonMembers, err := (&bob_repo.LessonMemberRepo{}).GetLessonMembersInLessons(ctx, db, database.TextArray(lessonIDs))
	if err != nil {
		return nil, fmt.Errorf("LessonMemberRepo.GetLessonMembersInLesson: %w", err)
	}
	lessonWithStudent := make([]string, 0, len(lessonMembers)*2)
	for _, lm := range lessonMembers {
		lessonWithStudent = append(lessonWithStudent, lm.LessonID.String, lm.UserID.String)
	}
	reallocation, err := (&ReallocationRepo{}).GetByNewLessonIDAndStudentID(ctx, db, lessonWithStudent)
	if err != nil {
		return nil, fmt.Errorf("ReallocationRepo.GetByNewLessonIDAndStudentID: %w", err)
	}
	reallocationMap := make(map[string]string, len(reallocation))
	for _, r := range reallocation {
		reallocationMap[r.NewLessonID+"-"+r.StudentID] = r.OriginalLessonID
	}

	lessonLearnersMap := make(map[string][]*domain.LessonLearner)
	for _, lm := range lessonMembers {
		lessonLearner := &domain.LessonLearner{
			LearnerID:        lm.UserID.String,
			CourseID:         lm.CourseID.String,
			AttendStatus:     domain.StudentAttendStatus(lm.AttendanceStatus.String),
			AttendanceNotice: domain.StudentAttendanceNotice(lm.AttendanceNotice.String),
			AttendanceReason: domain.StudentAttendanceReason(lm.AttendanceReason.String),
			AttendanceNote:   lm.AttendanceNote.String,
			Reallocate: func() *domain.Reallocate {
				uniqueKey := lm.LessonID.String + "-" + lm.UserID.String
				if originalLessonID, exists := reallocationMap[uniqueKey]; exists {
					return &domain.Reallocate{
						OriginalLessonID: originalLessonID,
					}
				}
				return nil
			}(),
		}
		lessonLearnersMap[lm.LessonID.String] = append(lessonLearnersMap[lm.LessonID.String], lessonLearner)
	}

	lessonGroups, err := (&LessonGroupRepo{}).getByIDs(ctx, db, lessonGroupIDs)
	if err != nil {
		return nil, fmt.Errorf("LessonGroupRepo.getByIDs: %w", err)
	}
	lessonGroupMaterial := make(map[string][]string)
	for _, lg := range lessonGroups {
		lessonID := lessonGroupMap[lg.LessonGroupID.String]
		lessonGroupMaterial[lessonID] = database.FromTextArray(lg.MediaIDs)
	}
	lessonClassrooms, err := (&LessonClassroomRepo{}).GetClassroomIDsByLessonIDs(ctx, db, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("LessonClassroomRepo.GetClassroomIDsByLessonIDs: %w", err)
	}
	lessonClassroomMap := make(map[string][]string)
	for _, lc := range lessonClassrooms {
		lessonClassroomMap[lc.LessonID.String] = append(lessonClassroomMap[lc.LessonID.String], lc.ClassroomID.String)
	}
	res := []*domain.Lesson{}
	for _, lesson := range lessons {
		lessonEntity := domain.NewLesson().
			WithID(lesson.LessonID.String).
			WithLocationID(lesson.CenterID.String).
			WithTimeRange(lesson.StartTime.Time, lesson.EndTime.Time).
			WithSchedulingStatus(domain.LessonSchedulingStatus(lesson.SchedulingStatus.String)).
			WithTeachingMedium(domain.LessonTeachingMedium(lesson.TeachingMedium.String)).
			WithTeachingMethod(domain.LessonTeachingMethod(lesson.TeachingMethod.String)).
			WithLearners(lessonLearnersMap[lesson.LessonID.String]).
			WithTeacherIDs(lessonTeacherMap[lesson.LessonID.String]).
			WithMaterials(lessonGroupMaterial[lesson.LessonID.String]).
			WithClassroomIDs(lessonClassroomMap[lesson.LessonID.String]).
			WithCourseID(lesson.CourseID.String).
			WithClassID(lesson.ClassID.String).
			WithIsLocked(lesson.IsLocked.Bool).
			WithLessonCapacity(lesson.LessonCapacity.Int).
			WithZoomAccountID(lesson.ZoomOwnerID.String).
			WithZoomID(lesson.ZoomID.String).
			WithZoomLink(lesson.ZoomLink.String).
			WithZoomOccurrenceID(lesson.ZoomOccurrenceID.String).
			WithClassDoOwnerID(lesson.ClassDoOwnerID.String).
			WithClassDoLink(lesson.ClassDoLink.String).
			WithClassDoRoomID(lesson.ClassDoRoomID.String).
			BuildDraft()
		res = append(res, lessonEntity)
	}
	return res, nil
}

func (l *LessonRepo) GetFutureRecurringLessonIDs(ctx context.Context, db database.QueryExecer, lessonID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.GetFutureRecurringLessonIDs")
	defer span.End()

	currentLesson, err := l.getLessonByID(ctx, db, lessonID)
	if err != nil {
		return nil, fmt.Errorf("could not get current lesson: %s", lessonID)
	}

	query := `SELECT lesson_id FROM lessons
			WHERE scheduler_id = $1 AND deleted_at IS NULL 
			AND start_time >= $2::timestamptz AND is_locked = false`

	rows, err := db.Query(ctx, query, &currentLesson.SchedulerID, &currentLesson.StartTime)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	lessonIDs := []string{}
	for rows.Next() {
		var lessonId pgtype.Text
		if err := rows.Scan(&lessonId); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonIDs = append(lessonIDs, lessonId.String)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return lessonIDs, nil
}

const DistinctKeyword string = "distinct"

const retrieveLessonQueryFuture = `
	select fl.lesson_id, fl."name", fl.start_time, fl.end_time, fl.teaching_method, 
	fl.teaching_medium, fl.center_id, fl.course_id, fl.class_id, fl.scheduling_status, fl.lesson_capacity, 
	fl.end_at, fl.zoom_link, fl.classdo_link
	from filter_lesson fl
	where $%d::text IS NULL
				OR (fl.start_time, fl.end_time, fl.lesson_id) > ((SELECT start_time FROM lessons WHERE lesson_id = $%d LIMIT 1), (SELECT end_time FROM lessons WHERE lesson_id = $%d LIMIT 1), $%d)
	order by fl.start_time ASC, fl.end_time ASC, fl.lesson_id ASC
	LIMIT $%d
	`

const retrieveLessonQueryPast = `
	select fl.lesson_id, fl."name", fl.start_time, fl.end_time, fl.teaching_method, 
	fl.teaching_medium, fl.center_id, fl.course_id, fl.class_id, fl.scheduling_status, fl.lesson_capacity, 
	fl.end_at, fl.zoom_link, fl.classdo_link
	from filter_lesson fl
	where $%d::text IS NULL
				OR (fl.start_time, fl.end_time, fl.lesson_id) < ((SELECT start_time FROM lessons WHERE lesson_id = $%d LIMIT 1), (SELECT end_time FROM lessons WHERE lesson_id = $%d LIMIT 1), $%d)
	order by fl.start_time DESC, fl.end_time DESC, fl.lesson_id DESC
	LIMIT $%d
	`

const previousLessonQueryFuture = `
	, previous_sort as (select fl.lesson_id, count(*) OVER() AS total, fl.start_time, fl.end_time
	from filter_lesson fl
	where $%d::text IS NULL
				OR (fl.start_time, fl.end_time, fl.lesson_id) < ((SELECT start_time FROM lessons WHERE lesson_id = $%d LIMIT 1), (SELECT end_time FROM lessons WHERE lesson_id = $%d LIMIT 1), $%d)
	order by fl.start_time DESC, fl.end_time DESC, fl.lesson_id DESC
	LIMIT $%d) select ps.lesson_id, ps.total
		FROM previous_sort ps
		order by ps.start_time ASC, ps.end_time ASC, ps.lesson_id ASC
		LIMIT 1
	`

const previousLessonQueryPast = `
	, previous_sort as (select fl.lesson_id, count(*) OVER() AS total, fl.start_time, fl.end_time
	from filter_lesson fl
	where $%d::text IS NULL
				OR (fl.start_time, fl.end_time, fl.lesson_id) > ((SELECT start_time FROM lessons WHERE lesson_id = $%d LIMIT 1), (SELECT end_time FROM lessons WHERE lesson_id = $%d LIMIT 1), $%d)
	order by fl.start_time ASC, fl.end_time ASC, fl.lesson_id ASC
	LIMIT $%d) select ps.lesson_id, ps.total
	FROM previous_sort ps
	order by ps.start_time DESC, ps.end_time DESC, ps.lesson_id DESC
	LIMIT 1
	`

func (l *LessonRepo) Retrieve(ctx context.Context, db database.QueryExecer, params *payloads.GetLessonListArg) ([]*domain.Lesson, uint32, string, uint32, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.Retrieve")
	defer span.End()

	paramDtos := ToListLessonArgsDto(params)

	distinct := ""
	baseTable := `
	WITH filter_lesson AS (
		SELECT %s l.lesson_id, l."name", l.start_time , l.end_time, l.teaching_method, 
		l.teaching_medium, l.center_id, l.course_id, l.class_id, l.scheduling_status, l.lesson_capacity, 
		l.end_at, l.zoom_link, l.classdo_link
		FROM lessons l 
	`
	where := fmt.Sprintf(`WHERE l.deleted_at IS NULL
		AND l.start_time %s $1::timestamptz
		AND ($2::timestamptz IS NULL OR l.end_time >= $2::timestamptz)
		AND ($3::timestamptz IS NULL OR l.start_time <= $3::timestamptz)
		AND l.resource_path = $4
	`, paramDtos.Compare)

	args := []interface{}{
		&paramDtos.CurrentTime,
		&paramDtos.FromDate,
		&paramDtos.ToDate,
		&paramDtos.SchoolID,
	}
	paramsNum := len(args)
	if len(paramDtos.KeyWord.String) > 0 {
		distinct = DistinctKeyword
		paramsNum++
		baseTable += fmt.Sprintf(` join lesson_members lm on l.lesson_id = lm.lesson_id join user_basic_info ubi on ubi.user_id = lm.user_id AND ubi.user_role = '%s' AND ubi.deleted_at is null `, constants.UserRoleStudent)
		where += fmt.Sprintf(` AND lm.deleted_at IS NULL 
				AND ( nospace(ubi."name") ILIKE nospace(CONCAT('%%',$%d::text,'%%'))
					OR nospace(ubi."full_name_phonetic") ILIKE nospace(CONCAT('%%',$%d::text,'%%'))
				)`, paramsNum, paramsNum)
		args = append(args, &paramDtos.KeyWord)
	}
	if len(paramDtos.Dow) > 0 {
		paramsNum++
		where += fmt.Sprintf(` AND EXTRACT(DOW from l.start_time at time zone '%s') = ANY($%d)`, paramDtos.TimeZone, paramsNum)
		args = append(args, &paramDtos.Dow)
	}
	if paramDtos.FromTime != "" {
		where += fmt.Sprintf(` AND cast((l.end_time AT time zone '%s') as time) >= '%s'`, paramDtos.TimeZone, paramDtos.FromTime)
	}
	if paramDtos.ToTime != "" {
		where += fmt.Sprintf(` AND cast((l.start_time AT time zone '%s') as time) <= '%s'`, paramDtos.TimeZone, paramDtos.ToTime)
	}
	if paramDtos.LocationIDs.Status == pgtype.Present {
		paramsNum++
		where += fmt.Sprintf(` AND l.center_id = ANY($%d)`, paramsNum)
		args = append(args, &paramDtos.LocationIDs)
	}
	if paramDtos.ClassIDs.Status == pgtype.Present {
		paramsNum++
		where += fmt.Sprintf(` AND l.class_id = ANY($%d)`, paramsNum)
		args = append(args, &paramDtos.ClassIDs)
	}
	if paramDtos.Teachers.Status == pgtype.Present {
		paramsNum++
		distinct = DistinctKeyword
		baseTable += ` join lessons_teachers lt on l.lesson_id = lt.lesson_id `
		where += fmt.Sprintf(` AND lt.teacher_id = ANY($%d) and lt.deleted_at IS NULL`, paramsNum)
		args = append(args, &paramDtos.Teachers)
	}
	if paramDtos.Students.Status == pgtype.Present {
		paramsNum++
		distinct = DistinctKeyword
		if len(paramDtos.KeyWord.String) == 0 {
			baseTable += ` join lesson_members lm  on l.lesson_id = lm.lesson_id `
			where += " AND lm.deleted_at IS NULL "
		}
		where += fmt.Sprintf(` AND lm.user_id = ANY($%d) `, paramsNum)
		args = append(args, &paramDtos.Students)
	}

	isFilterGrade := paramDtos.ExistsConditionGrades()
	if isFilterGrade {
		paramsNum++
		distinct = DistinctKeyword
		if paramDtos.Students.Status == pgtype.Null && len(paramDtos.KeyWord.String) == 0 {
			baseTable += ` join lesson_members lm on l.lesson_id = lm.lesson_id `
		}

		if len(paramDtos.KeyWord.String) == 0 {
			baseTable += fmt.Sprintf(` join user_basic_info ubi on ubi.user_id = lm.user_id AND ubi.user_role = '%s' AND ubi.deleted_at is null `, constants.UserRoleStudent)
		}
		where += fmt.Sprintf(` AND ubi.grade_id = ANY($%d) AND lm.deleted_at IS NULL `, paramsNum)
		args = append(args, paramDtos.GetParamGrades())
	}
	if paramDtos.Courses.Status == pgtype.Present {
		paramsNum++
		distinct = DistinctKeyword
		if paramDtos.Students.Status == pgtype.Null && len(paramDtos.Grades) == 0 && len(paramDtos.KeyWord.String) == 0 && !isFilterGrade {
			baseTable += ` left join lesson_members lm on l.lesson_id = lm.lesson_id `
		}
		where += fmt.Sprintf(` AND (( lm.course_id = ANY($%d) AND lm.deleted_at IS NULL ) OR l.course_id = ANY($%d) ) `, paramsNum, paramsNum)
		args = append(args, &paramDtos.Courses)
	}

	if paramDtos.LessonSchedulingStatus.Status == pgtype.Present {
		paramsNum++
		where += fmt.Sprintf(` AND l.scheduling_status = ANY($%d)`, paramsNum)
		args = append(args, &paramDtos.LessonSchedulingStatus)
	}
	// filter by course types
	if paramDtos.CourseTypeIDs.Status == pgtype.Present {
		paramsNum++
		// group teaching method: get course_type by course assigned to lesson
		where += fmt.Sprintf(` AND ( exists ( 
			 select lesson_id from lessons l2 
			 left join courses c2  on c2.course_id = l2.course_id 
			 where l.lesson_id = l2.lesson_id 
			 and c2.course_type_id = ANY($%d) 
			 and l.teaching_method ='%s')  `,
			paramsNum, domain.LessonTeachingMethodGroup)
		// individual teaching method: get course_type by course assigned to each lesson_member
		where += fmt.Sprintf(` OR exists ( 
			 select lesson_id from lesson_members lm 
			 left join courses c on c.course_id = lm.course_id 
			 where l.lesson_id = lm.lesson_id
			 and lm.deleted_at IS NULL
			 and c.course_type_id = ANY($%d) 
			 and l.teaching_method ='%s') )  `,
			paramsNum, domain.LessonTeachingMethodIndividual)
		args = append(args, &paramDtos.CourseTypeIDs)
	}
	// filter by report status
	if paramDtos.IsPresentReportStatus() {
		paramsNum++
		baseTable += " left join lesson_reports lr on lr.lesson_id = l.lesson_id "
		if paramDtos.IsHaveNoneReportStatus() {
			where += fmt.Sprintf(" AND ( lr.report_submitting_status = ANY($%d) OR lr.report_submitting_status is null ) AND lr.deleted_at is null ", paramsNum)
		} else {
			where += fmt.Sprintf(" AND lr.report_submitting_status = ANY($%d) AND lr.deleted_at is null ", paramsNum)
		}
		args = append(args, &paramDtos.ReportStatus)
	}
	var query, queryTotal string
	var total pgtype.Int8

	// get total
	baseTable = fmt.Sprintf(baseTable, distinct)
	queryTotal = strings.Replace(baseTable, fmt.Sprintf(`WITH filter_lesson AS (
		SELECT %s l.lesson_id, l."name", l.start_time , l.end_time, l.teaching_method, 
		l.teaching_medium, l.center_id, l.course_id, l.class_id, l.scheduling_status, l.lesson_capacity, 
		l.end_at, l.zoom_link, l.classdo_link`, distinct), fmt.Sprintf("select count(%s l.lesson_id)", distinct), 1) + where
	if err := db.QueryRow(ctx, queryTotal, args...).Scan(&total); err != nil {
		return nil, 0, "", 0, errors.Wrap(err, "get total err")
	}
	args = append(args, &paramDtos.LessonID)
	args = append(args, &paramDtos.Limit)
	lessonID := paramsNum + 1
	limit := paramsNum + 2
	// get list
	query = baseTable + where + ")" + fmt.Sprintf(retrieveLessonQueryFuture, lessonID, lessonID, lessonID, lessonID, limit)
	if paramDtos.LessonTime == "past" {
		query = baseTable + where + ")" + fmt.Sprintf(retrieveLessonQueryPast, lessonID, lessonID, lessonID, lessonID, limit)
	}
	lessons := Lessons{}
	fields := []string{"lesson_id", "name", "start_time", "end_time", "teaching_method", "teaching_medium", "center_id", "course_id", "class_id", "scheduling_status", "lesson_capacity", "end_at", "zoom_link", "classdo_link"}
	var rows pgx.Rows
	var err error
	rows, err = db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, "", 0, err
	}
	defer rows.Close()

	for rows.Next() {
		lesson := lessons.Add()
		scanFields := database.GetScanFields(lesson, fields)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, 0, "", 0, errors.Wrap(err, "rows.Scan")
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, "", 0, errors.Wrap(err, "rows.Err")
	}
	var preTotal = pgtype.Int8{Int: 0, Status: pgtype.Present}
	var preOffset = pgtype.Text{String: "", Status: pgtype.Present}
	if len(paramDtos.LessonID.String) > 0 {
		query = baseTable + where + ")" + fmt.Sprintf(previousLessonQueryFuture, lessonID, lessonID, lessonID, lessonID, limit)
		if paramDtos.LessonTime == "past" {
			query = baseTable + where + ")" + fmt.Sprintf(previousLessonQueryPast, lessonID, lessonID, lessonID, lessonID, limit)
		}
		if err := db.QueryRow(ctx, query, args...).Scan(&preOffset, &preTotal); err != nil {
			if err != pgx.ErrNoRows {
				return nil, 0, "", 0, errors.Wrap(err, "get previous err")
			}
		}
	}

	res := []*domain.Lesson{}
	for _, lesson := range lessons {
		lessonEntity := domain.NewLesson().
			WithID(lesson.LessonID.String).
			WithName(lesson.Name.String).
			WithLocationID(lesson.CenterID.String).
			WithModificationTime(lesson.CreatedAt.Time, lesson.UpdatedAt.Time).
			WithTimeRange(lesson.StartTime.Time, lesson.EndTime.Time).
			WithSchedulingStatus(domain.LessonSchedulingStatus(lesson.SchedulingStatus.String)).
			WithTeachingMedium(domain.LessonTeachingMedium(lesson.TeachingMedium.String)).
			WithTeachingMethod(domain.LessonTeachingMethod(lesson.TeachingMethod.String)).
			WithCourseID(lesson.CourseID.String).
			WithClassID(lesson.ClassID.String).
			WithLessonCapacity(lesson.LessonCapacity.Int).
			WithEndAt(database.FromTimestamptz(lesson.EndAt)).
			WithZoomLink(lesson.ZoomLink.String).
			WithClassDoLink(lesson.ClassDoLink.String).
			BuildDraft()

		res = append(res, lessonEntity)
	}

	return res, uint32(total.Int), preOffset.String, uint32(preTotal.Int), nil
}

func (l *LessonRepo) GetLessonsTeachingModelGroupByClassIdWithDuration(ctx context.Context, db database.Ext, queryLesson *domain.QueryLesson) ([]*domain.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.GetLessonByClassIdWithStartAndEndDate")
	defer span.End()

	lesson := &Lesson{}
	fields, _ := lesson.FieldMap()
	var queryBuilder strings.Builder
	params := make([]interface{}, 0, reflect.TypeOf(domain.QueryLesson{}).NumField())
	params = append(params, queryLesson.ClassID)
	queryBuilder.WriteString(`SELECT %s FROM lessons WHERE teaching_model = 'LESSON_TEACHING_MODEL_GROUP'
	 AND class_id = $1 AND deleted_at IS NULL `)
	if queryLesson.StartTime != nil {
		queryBuilder.WriteString(`AND start_time >= $2::timestamptz `)
		params = append(params, queryLesson.StartTime)
	}
	if queryLesson.EndTime != nil {
		queryBuilder.WriteString(`AND start_time <= $3::timestamptz `)
		params = append(params, queryLesson.EndTime)
	}
	query := fmt.Sprintf(queryBuilder.String(), strings.Join(fields, ","))
	rows, err := db.Query(ctx, query, params...)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	lessons := []*Lesson{}
	for rows.Next() {
		lesson := &Lesson{}
		if err := rows.Scan(database.GetScanFields(lesson, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessons = append(lessons, lesson)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	res := make([]*domain.Lesson, 0, len(lessons))
	for _, lesson := range lessons {
		lessonEntity := domain.NewLesson().
			WithID(lesson.LessonID.String).
			WithName(lesson.Name.String).
			WithLocationID(lesson.CenterID.String).
			WithModificationTime(lesson.CreatedAt.Time, lesson.UpdatedAt.Time).
			WithTimeRange(lesson.StartTime.Time, lesson.EndTime.Time).
			WithSchedulingStatus(domain.LessonSchedulingStatus(lesson.SchedulingStatus.String)).
			WithTeachingMedium(domain.LessonTeachingMedium(lesson.TeachingMedium.String)).
			WithTeachingMethod(domain.LessonTeachingMethod(lesson.TeachingMethod.String)).
			WithCourseID(lesson.CourseID.String).
			WithClassID(lesson.ClassID.String).
			WithLessonCapacity(lesson.LessonCapacity.Int).
			BuildDraft()

		res = append(res, lessonEntity)
	}
	return res, nil
}

func (l *LessonRepo) GetLessonsOnCalendar(ctx context.Context, db database.QueryExecer, params *payloads.GetLessonListOnCalendarArgs) ([]*domain.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.GetLessonsOnCalendar")
	defer span.End()

	paramsDTO, err := ToListLessonOnCalendarArgsDto(params)
	if err != nil {
		return nil, err
	}
	distinct := ""

	baseQuery := `SELECT %s l.lesson_id, l.name, l.start_time, l.end_time, l.teaching_method, l.teaching_medium, l.center_id,
		l.course_id, l.class_id, l.scheduling_status, l.scheduler_id, l.lesson_capacity, c.name as "course_name", cl.name as "class_name"
		FROM lessons l
		LEFT JOIN class cl ON l.class_id = cl.class_id 
		LEFT JOIN courses c ON l.course_id = c.course_id `
	whereClause := fmt.Sprintf(`WHERE ( (l.start_time at time zone '%s')::date 
				BETWEEN ($1 at time zone '%s')::date
				AND ($2 at time zone '%s')::date )
				AND l.center_id = $3
				AND l.deleted_at IS NULL
				AND cl.deleted_at IS NULL
				AND c.deleted_at IS NULL `, paramsDTO.Timezone, paramsDTO.Timezone, paramsDTO.Timezone)
	noneAssignedTeacherLessonQuery := ``

	orderByClause := `ORDER BY l.start_time ASC, l.end_time ASC, l.lesson_id ASC `

	args := []interface{}{
		&paramsDTO.FromDate,
		&paramsDTO.ToDate,
		&paramsDTO.LocationID,
	}
	paramsCount := len(args)

	if paramsDTO.StudentIDs.Status == pgtype.Present {
		paramsCount++
		distinct = DistinctKeyword
		baseQuery += `LEFT JOIN lesson_members lm ON l.lesson_id = lm.lesson_id `
		whereClause += fmt.Sprintf(`AND lm.user_id = ANY($%d) and lm.deleted_at IS NULL `, paramsCount)
		args = append(args, &paramsDTO.StudentIDs)
	}

	if paramsDTO.ClassIDs.Status == pgtype.Present {
		paramsCount++
		whereClause += fmt.Sprintf(`AND l.class_id = ANY($%d) `, paramsCount)
		args = append(args, &paramsDTO.ClassIDs)
	}

	if paramsDTO.CourseIDs.Status == pgtype.Present {
		paramsCount++
		distinct = DistinctKeyword
		if paramsDTO.StudentIDs.Status == pgtype.Null {
			baseQuery += `LEFT JOIN lesson_members lm ON l.lesson_id = lm.lesson_id `
		}

		whereClause += fmt.Sprintf(`AND (lm.course_id = ANY($%d) OR l.course_id = ANY($%d))
										AND lm.deleted_at IS NULL `, paramsCount, paramsCount)

		args = append(args, &paramsDTO.CourseIDs)
	}

	if paramsDTO.TeacherIDs.Status == pgtype.Present {
		paramsCount++
		distinct = DistinctKeyword
		baseQuery += `JOIN lessons_teachers lt ON l.lesson_id = lt.lesson_id `
		whereClause += `AND lt.deleted_at IS NULL `
		if paramsDTO.IsIncludeNoneAssignedTeacherLessons.Bool {
			noneAssignedTeacherLessonQuery = fmt.Sprintf(baseQuery, distinct) + whereClause + "AND lt.lesson_id is NULL " + orderByClause
			noneAssignedTeacherLessonQuery = strings.Replace(noneAssignedTeacherLessonQuery, "JOIN lessons_teachers lt ON l.lesson_id = lt.lesson_id", "LEFT JOIN lessons_teachers lt ON l.lesson_id = lt.lesson_id", 1)
		}
		whereClause += fmt.Sprintf(`AND lt.teacher_id = ANY($%d) `, paramsCount)
		args = append(args, &paramsDTO.TeacherIDs)
	}

	query := fmt.Sprintf(baseQuery, distinct) + whereClause + orderByClause

	if paramsDTO.TeacherIDs.Status == pgtype.Present && paramsDTO.IsIncludeNoneAssignedTeacherLessons.Bool {
		query = fmt.Sprintf("(%s) union (%s)", noneAssignedTeacherLessonQuery, query)
	}

	fields := []string{
		"lesson_id",
		"name",
		"start_time",
		"end_time",
		"teaching_method",
		"teaching_medium",
		"center_id",
		"course_id",
		"class_id",
		"scheduling_status",
		"scheduler_id",
		"lesson_capacity",
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// fetch results of query
	var (
		courseName pgtype.Text
		className  pgtype.Text
	)
	lessonResult := &Lesson{}
	scanFields := append(database.GetScanFields(lessonResult, fields), &courseName, &className)

	lessonList := []*domain.Lesson{}
	for rows.Next() {
		if err := rows.Scan(scanFields...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		lessonDomain := domain.NewLesson().
			WithID(lessonResult.LessonID.String).
			WithName(lessonResult.Name.String).
			WithTimeRange(lessonResult.StartTime.Time, lessonResult.EndTime.Time).
			WithTeachingMedium(domain.LessonTeachingMedium(lessonResult.TeachingMedium.String)).
			WithTeachingMethod(domain.LessonTeachingMethod(lessonResult.TeachingMethod.String)).
			WithLocationID(lessonResult.CenterID.String).
			WithCourseID(lessonResult.CourseID.String).
			WithCourseName(courseName.String).
			WithClassID(lessonResult.ClassID.String).
			WithClassName(className.String).
			WithSchedulingStatus(domain.LessonSchedulingStatus(lessonResult.SchedulingStatus.String)).
			WithSchedulerID(lessonResult.SchedulerID.String).
			WithLessonCapacity(lessonResult.LessonCapacity.Int).
			BuildDraft()
		lessonList = append(lessonList, lessonDomain)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return lessonList, nil
}

func (l *LessonRepo) upsertLessonClassrooms(ctx context.Context, db database.QueryExecer, lessonIDs []string, classrooms domain.LessonClassrooms) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.upsertLessonClassrooms")
	defer span.End()

	now := time.Now()
	b := &pgx.Batch{}
	for _, lessonID := range lessonIDs {
		// deletes all rows belonging to lessonID before upserting
		b.Queue(`UPDATE lesson_classrooms SET deleted_at = $2 WHERE lesson_id = $1`, lessonID, now)
		l.queueUpsertLessonClassroom(b, lessonID, classrooms, now)
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

func (l *LessonRepo) queueUpsertLessonClassroom(b *pgx.Batch, lessonID string, classrooms domain.LessonClassrooms, now time.Time) {
	queueFn := func(b *pgx.Batch, classroomID string) {
		query := `
			INSERT INTO lesson_classrooms (lesson_id, classroom_id) VALUES ($1, $2)
			ON CONFLICT ON CONSTRAINT pk__lesson_classrooms 
				DO UPDATE SET updated_at = $3, deleted_at = NULL`
		b.Queue(query, lessonID, classroomID, now)
	}

	for _, classroom := range classrooms {
		queueFn(b, classroom.ClassroomID)
	}
}

func (l *LessonRepo) GenerateLessonTemplate(ctx context.Context, db database.QueryExecer) ([]byte, error) {
	lessons := []*LessonToExport{InitLessonTemplate()}

	exportCols := []exporter.ExportColumnMap{
		{
			DBColumn:  "partner_internal_id",
			CSVColumn: "partner_internal_id",
		},
		{
			DBColumn:  "start_time",
			CSVColumn: "start_date_time",
		},
		{
			DBColumn:  "end_time",
			CSVColumn: "end_date_time",
		},
		{
			DBColumn:  "teaching_method",
			CSVColumn: "teaching_method",
		},
	}
	exportable := sliceutils.Map(lessons, func(d *LessonToExport) database.Entity {
		return d
	})

	str, err := exporter.ExportBatch(exportable, exportCols)
	if err != nil {
		return nil, fmt.Errorf("ExportBatch: %w", err)
	}
	return exporter.ToCSV(str), nil
}

func (l *LessonRepo) GenerateLessonTemplateV2(ctx context.Context, db database.QueryExecer) ([]byte, error) {
	lessons := []*LessonToExport{InitLessonTemplateV2()}

	exportCols := []exporter.ExportColumnMap{
		{
			DBColumn:  "partner_internal_id",
			CSVColumn: "partner_internal_id",
		},
		{
			DBColumn:  "start_time",
			CSVColumn: "start_date_time",
		},
		{
			DBColumn:  "end_time",
			CSVColumn: "end_date_time",
		},
		{
			DBColumn:  "teaching_method",
			CSVColumn: "teaching_method",
		},
		{
			DBColumn:  "teaching_medium",
			CSVColumn: "teaching_medium",
		},
		{
			DBColumn:  "teacher_ids",
			CSVColumn: "teacher_ids",
		},
		{
			DBColumn:  "student_course_ids",
			CSVColumn: "student_course_ids",
		},
	}
	exportable := sliceutils.Map(lessons, func(d *LessonToExport) database.Entity {
		return d
	})

	str, err := exporter.ExportBatch(exportable, exportCols)
	if err != nil {
		return nil, fmt.Errorf("ExportBatch: %w", err)
	}
	return exporter.ToCSV(str), nil
}

func (l *LessonRepo) GetLessonWithNamesByID(ctx context.Context, db database.QueryExecer, lessonID string) (*domain.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.GetLessonWithNamesByID")
	defer span.End()

	query := `SELECT l.lesson_id, l.name, l.start_time, l.end_time, l.teaching_method, l.teaching_medium, l.center_id
			,l.course_id, l.class_id, l.scheduling_status, l.scheduler_id, l.is_locked
			,l.zoom_id, l.zoom_link, l.zoom_owner_id, l.classdo_owner_id, l.classdo_link, l.classdo_room_id, l.lesson_capacity
			,c.name as "course_name", cl.name as "class_name", loc.name as "location_name"
		FROM lessons l
		LEFT JOIN class cl ON l.class_id = cl.class_id 
		LEFT JOIN courses c ON l.course_id = c.course_id 
		LEFT JOIN locations loc ON l.center_id = loc.location_id 
		WHERE l.lesson_id = $1
		AND l.deleted_at IS NULL
		AND cl.deleted_at IS NULL
		AND c.deleted_at IS NULL 
		AND loc.deleted_at IS NULL `

	fields := []string{
		"lesson_id",
		"name",
		"start_time",
		"end_time",
		"teaching_method",
		"teaching_medium",
		"center_id",
		"course_id",
		"class_id",
		"scheduling_status",
		"scheduler_id",
		"is_locked",
		"zoom_id",
		"zoom_link",
		"zoom_owner_id",
		"classdo_owner_id",
		"classdo_link",
		"classdo_room_id",
		"lesson_capacity",
	}

	lessonResult := &Lesson{}
	var (
		courseName   pgtype.Text
		className    pgtype.Text
		locationName pgtype.Text
	)
	scanFields := append(database.GetScanFields(lessonResult, fields), &courseName, &className, &locationName)
	if err := db.QueryRow(ctx, query, &lessonID).Scan(scanFields...); err != nil {
		return nil, err
	}

	lesson := domain.NewLesson().
		WithIsLocked(lessonResult.IsLocked.Bool).
		WithID(lessonResult.LessonID.String).
		WithName(lessonResult.Name.String).
		WithTimeRange(lessonResult.StartTime.Time, lessonResult.EndTime.Time).
		WithTeachingMedium(domain.LessonTeachingMedium(lessonResult.TeachingMedium.String)).
		WithTeachingMethod(domain.LessonTeachingMethod(lessonResult.TeachingMethod.String)).
		WithSchedulingStatus(domain.LessonSchedulingStatus(lessonResult.SchedulingStatus.String)).
		WithCourseID(lessonResult.CourseID.String).
		WithCourseName(courseName.String).
		WithLocationID(lessonResult.CenterID.String).
		WithLocationName(locationName.String).
		WithClassID(lessonResult.ClassID.String).
		WithClassName(className.String).
		WithSchedulerID(lessonResult.SchedulerID.String).
		WithZoomID(lessonResult.ZoomID.String).
		WithZoomLink(lessonResult.ZoomLink.String).
		WithZoomAccountID(lessonResult.ZoomOwnerID.String).
		WithClassDoOwnerID(lessonResult.ClassDoOwnerID.String).
		WithClassDoLink(lessonResult.ClassDoLink.String).
		WithClassDoRoomID(lessonResult.ClassDoRoomID.String).
		WithLessonCapacity(lessonResult.LessonCapacity.Int).
		BuildDraft()

	return lesson, nil
}

func (l *LessonRepo) RemoveZoomLinkOfLesson(ctx context.Context, db database.QueryExecer, zoomOwnerIds []string) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.RemoveZoomLinkOfLesson")
	defer span.End()
	sql := `UPDATE lessons
		SET zoom_link = NULL, zoom_owner_id = NULL, zoom_id = NULL, updated_at = $2
		WHERE  zoom_owner_id = ANY($1) AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, &zoomOwnerIds, time.Now())
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}
	return nil
}

func (l *LessonRepo) RemoveClassDoLinkOfLesson(ctx context.Context, db database.QueryExecer, classDoOwnerIds []string) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.RemoveZoomLinkOfLesson")
	defer span.End()
	sql := `UPDATE lessons
		SET classdo_link = NULL, classdo_owner_id = NULL, classdo_room_id = NULL, updated_at = $2
		WHERE  classdo_owner_id = ANY($1) AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, &classDoOwnerIds, time.Now())
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}
	return nil
}

func (l *LessonRepo) RemoveZoomLinkByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.RemoveZoomLinkByLessonID")
	defer span.End()
	sql := `UPDATE lessons
		SET zoom_link = NULL, zoom_owner_id = NULL, zoom_id = NULL, updated_at = $2
		WHERE  lesson_id = $1 AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, &lessonID, time.Now())
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}
	return nil
}
func (l *LessonRepo) RemoveClassDoLinkByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.RemoveClassDoLinkByLessonID")
	defer span.End()
	sql := `UPDATE lessons
		SET classdo_link = NULL, classdo_owner_id = NULL, classdo_room_id = NULL, updated_at = $2
		WHERE  lesson_id = $1 AND deleted_at IS NULL`
	_, err := db.Exec(ctx, sql, &lessonID, time.Now())
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}
	return nil
}

func (l *LessonRepo) GetLessonsByLocationStatusAndDateTimeRange(ctx context.Context, db database.QueryExecer, params *payloads.GetLessonsByLocationStatusAndDateTimeRangeArgs) ([]*domain.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.GetLessonsByLocationStatusAndDateTimeRange")
	defer span.End()

	paramsDTO := ToListLessonByLocationStatusAndDateTimeRangeArgsDto(params)
	lesson := &Lesson{}
	fields, values := lesson.FieldMap()

	baseQuery := fmt.Sprintf(`SELECT %s FROM %s `, strings.Join(fields, ","), lesson.TableName())
	whereClause := ` WHERE deleted_at IS NULL
		AND center_id = $1
		AND scheduling_status = $2
		AND (start_time at time zone $3)::date >= ($4 at time zone $3)::date
		AND (end_time at time zone $3)::date <= ($5 at time zone $3)::date 
	`
	args := []interface{}{
		&paramsDTO.LocationID,
		&paramsDTO.LessonStatus,
		&paramsDTO.Timezone,
		&paramsDTO.StartDate,
		&paramsDTO.EndDate,
	}
	paramsCount := len(args)

	if params.LessonStatus == domain.LessonSchedulingStatusCompleted {
		whereClause += ` AND is_locked = false `
	}

	if paramsDTO.StartTime.Status == pgtype.Present {
		paramsCount++
		whereClause += fmt.Sprintf(` AND (end_time at time zone $3)::time >= ($%v at time zone $3)::time `, paramsCount)
		args = append(args, &paramsDTO.StartTime)
	}

	if paramsDTO.EndTime.Status == pgtype.Present {
		paramsCount++
		whereClause += fmt.Sprintf(` AND (start_time at time zone $3)::time <= ($%v at time zone $3)::time `, paramsCount)
		args = append(args, &paramsDTO.EndTime)
	}

	query := baseQuery + whereClause
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	lessonList := []*domain.Lesson{}
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		lessonDomain := domain.NewLesson().
			WithID(lesson.LessonID.String).
			WithName(lesson.Name.String).
			WithTimeRange(lesson.StartTime.Time, lesson.EndTime.Time).
			WithTeachingMedium(domain.LessonTeachingMedium(lesson.TeachingMedium.String)).
			WithTeachingMethod(domain.LessonTeachingMethod(lesson.TeachingMethod.String)).
			WithLocationID(lesson.CenterID.String).
			WithCourseID(lesson.CourseID.String).
			WithClassID(lesson.ClassID.String).
			WithSchedulingStatus(domain.LessonSchedulingStatus(lesson.SchedulingStatus.String)).
			WithModificationTime(lesson.CreatedAt.Time, lesson.UpdatedAt.Time).
			WithLessonCapacity(lesson.LessonCapacity.Int).
			BuildDraft()
		lessonList = append(lessonList, lessonDomain)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return lessonList, nil
}

func (l *LessonRepo) GetLessonWithSchedulerInfoByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) (*domain.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchedulerRepo.GetByLessonID")
	defer span.End()

	query := `SELECT l.lesson_id, l.name, l.start_time, l.end_time, l.teaching_method, l.teaching_medium, l.center_id
								,l.course_id, l.class_id, l.scheduling_status, l.scheduler_id, l.is_locked
								,l.zoom_id, l.zoom_link, l.zoom_owner_id, l.classdo_owner_id, l.classdo_link, l.classdo_room_id, l.lesson_capacity, s.freq 
						FROM lessons l 
						JOIN  scheduler s on s.scheduler_id = l.scheduler_id
						WHERE l.lesson_id = $1 AND s.deleted_at is null`
	fields := []string{
		"lesson_id",
		"name",
		"start_time",
		"end_time",
		"teaching_method",
		"teaching_medium",
		"center_id",
		"course_id",
		"class_id",
		"scheduling_status",
		"scheduler_id",
		"is_locked",
		"zoom_id",
		"zoom_link",
		"zoom_owner_id",
		"classdo_owner_id",
		"classdo_link",
		"classdo_room_id",
		"lesson_capacity",
	}

	lessonResult := &Lesson{}
	var (
		freq pgtype.Text
	)
	scanFields := append(database.GetScanFields(lessonResult, fields), &freq)
	if err := db.QueryRow(ctx, query, &lessonID).Scan(scanFields...); err != nil {
		return nil, err
	}
	lesson := domain.NewLesson().
		WithIsLocked(lessonResult.IsLocked.Bool).
		WithID(lessonResult.LessonID.String).
		WithName(lessonResult.Name.String).
		WithTimeRange(lessonResult.StartTime.Time, lessonResult.EndTime.Time).
		WithTeachingMedium(domain.LessonTeachingMedium(lessonResult.TeachingMedium.String)).
		WithTeachingMethod(domain.LessonTeachingMethod(lessonResult.TeachingMethod.String)).
		WithSchedulingStatus(domain.LessonSchedulingStatus(lessonResult.SchedulingStatus.String)).
		WithCourseID(lessonResult.CourseID.String).
		WithLocationID(lessonResult.CenterID.String).
		WithClassID(lessonResult.ClassID.String).
		WithSchedulerID(lessonResult.SchedulerID.String).
		WithZoomID(lessonResult.ZoomID.String).
		WithZoomLink(lessonResult.ZoomLink.String).
		WithZoomAccountID(lessonResult.ZoomOwnerID.String).
		WithClassDoOwnerID(lessonResult.ClassDoOwnerID.String).
		WithClassDoLink(lessonResult.ClassDoLink.String).
		WithClassDoRoomID(lessonResult.ClassDoRoomID.String).
		WithLessonCapacity(lessonResult.LessonCapacity.Int).
		WithSchedulerInfo(&domain.SchedulerInfo{SchedulerID: lessonResult.SchedulerID.String, Freq: freq.String}).
		BuildDraft()

	return lesson, nil
}

func (l *LessonRepo) UpdateLessonsTeachingTime(ctx context.Context, db database.Ext, lessons []*domain.Lesson) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.UpdateLessonsTeachingTime")
	defer span.End()

	lessonsDto := make([]*Lesson, 0, len(lessons))
	for _, lesson := range lessons {
		lessonDto, err := NewLessonFromEntity(lesson)
		if err != nil {
			return err
		}
		if err = lessonDto.Normalize(); err != nil {
			return fmt.Errorf("got error when normalize lesson dto: %w", err)
		}
		if lessonDto.CourseID.Status == pgtype.Present {
			lg := NewLessonGroupFromLessonEntity(lesson, lessonDto.CourseID)
			if err = (&LessonGroupRepo{}).Upsert(ctx, db, lg); err != nil {
				return fmt.Errorf("LessonGroupRepo.insert: %w", err)
			}
			lessonDto.LessonGroupID = lg.LessonGroupID
		}
		lessonsDto = append(lessonsDto, lessonDto)
	}

	// update lesson
	if err := l.upsertLessons(ctx, db, lessonsDto); err != nil {
		return fmt.Errorf("failed to upsert lesson: %w", err)
	}

	return nil
}

func (l *LessonRepo) GetFutureLessonsByCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs []string, timezone string) ([]*domain.Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.GetFutureLessonsByCourseIDs")
	defer span.End()
	lesson := &Lesson{}
	fields, _ := lesson.FieldMap()
	query := fmt.Sprintf(`SELECT l.%s 
			FROM %s l
			WHERE l.lesson_id IN (select lc.lesson_id from lessons_courses lc where lc.course_id = ANY($1) AND lc.deleted_at IS NULL)
			AND l.start_time::timestamptz AT TIME ZONE $2 > now()::timestamptz AT TIME ZONE $2
			AND l.is_locked = false AND l.deleted_at IS NULL`,
		strings.Join(fields, ",l."), lesson.TableName())

	rows, err := db.Query(ctx, query, courseIDs, timezone)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	lessonList := []*Lesson{}
	for rows.Next() {
		lEntity := &Lesson{}
		if err = rows.Scan(database.GetScanFields(lEntity, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonList = append(lessonList, lEntity)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return l.fillLessonEntities(ctx, db, lessonList)
}

func (l *LessonRepo) GetLessonsWithSchedulerNull(ctx context.Context, db database.QueryExecer, limit int, offset int) ([]*Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.GetLessonsWithSchedulerNull")
	defer span.End()

	lesson := &Lesson{}
	fields, _ := lesson.FieldMap()
	query := fmt.Sprintf(`SELECT l.%s 
			FROM %s l WHERE l.scheduler_id IS NULL AND l.deleted_at IS NULL 
			LIMIT $1 OFFSET $2`,
		strings.Join(fields, ",l."), lesson.TableName())

	rows, err := db.Query(ctx, query, &limit, &offset)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	lessonList := []*Lesson{}
	for rows.Next() {
		lesson := &Lesson{}
		if err := rows.Scan(database.GetScanFields(lesson, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonList = append(lessonList, lesson)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return lessonList, nil
}

func (l *LessonRepo) FillSchedulerToLessons(ctx context.Context, db database.QueryExecer, schedulerMap map[string]string) error {
	ctx, span := interceptors.StartSpan(ctx, "AcademicWeekRepo.Insert")
	defer span.End()
	b := &pgx.Batch{}
	for lessonID, schedulerID := range schedulerMap {
		query := "UPDATE lessons SET scheduler_id = $1 WHERE lesson_id = $2"
		b.Queue(query, schedulerID, lessonID)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < b.Len(); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("lessons is not updated")
		}
	}
	return nil
}

func (l *LessonRepo) GetLessonsWithInvalidSchedulerID(ctx context.Context, db database.QueryExecer) ([]*Lesson, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonRepo.GetLessonsWithSchedulerNull")
	defer span.End()

	lesson := &Lesson{}
	fields, _ := lesson.FieldMap()
	query := fmt.Sprintf(`SELECT l.%s 
			FROM %s l WHERE l.scheduler_id not in (select scheduler_id from scheduler s)
			ORDER BY l.scheduler_id, l.start_time`,
		strings.Join(fields, ",l."), lesson.TableName())

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	lessonList := []*Lesson{}
	for rows.Next() {
		lesson := &Lesson{}
		if err := rows.Scan(database.GetScanFields(lesson, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonList = append(lessonList, lesson)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return lessonList, nil
}
