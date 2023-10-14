package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/constants"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type StudentSubscriptionRepo struct{}

func (s *StudentSubscriptionRepo) GetStudentCourseSubscriptions(ctx context.Context, db database.QueryExecer, locationID []string, studentIDWithCourseID ...string) (domain.StudentSubscriptions, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionRepo.GetStudentCourseSubscriptions")
	defer span.End()

	if len(studentIDWithCourseID)%2 != 0 {
		return nil, fmt.Errorf("missing course id of student %s", studentIDWithCourseID[len(studentIDWithCourseID)-1])
	}

	query := "SELECT :ColumnNames FROM lesson_student_subscriptions ls " +
		"LEFT JOIN lesson_student_subscription_access_path ap ON ls.student_subscription_id = ap.student_subscription_id " +
		"LEFT JOIN user_basic_info ubi ON ls.student_id = ubi.user_id AND ubi.deleted_at IS NULL " +
		"WHERE ls.deleted_at IS NULL AND ap.deleted_at IS NULL AND (ls.student_id,ls.course_id) IN (:PlaceHolderVar) " +
		":ExtendedCondition"

	// get all columns of lesson_student_subscriptions table and location_id of lesson_student_subscription_access_path table
	columnNames := fmt.Sprintf("ls.%s, ap.location_id, ubi.grade_id ", strings.Join(database.GetFieldNames(&StudentSubscription{}), ", ls."))
	studentIDCourseID := make([]string, 0, len(studentIDWithCourseID)/2) // will like ["($1, $2)", "($3, $4)", ...]
	args := make([]interface{}, 0, len(studentIDWithCourseID))
	for i := 0; i < len(studentIDWithCourseID); i += 2 {
		studentID := studentIDWithCourseID[i]
		courseID := studentIDWithCourseID[i+1]
		args = append(args, &studentID, &courseID)
		studentIDCourseID = append(studentIDCourseID, fmt.Sprintf("($%d, $%d)", i+1, i+2))
	}
	// placeHolderVar will like ($1, $2), ($3, $4), ($5, $6), ....
	placeHolderVar := strings.Join(studentIDCourseID, ", ")
	// add condition: location_id of records equal locationID
	var extendedCondition string
	if len(locationID) != 0 {
		args = append(args, &locationID)
		extendedCondition = fmt.Sprintf("AND ap.location_id = ANY($%d) ", len(args))
	}

	query = strings.ReplaceAll(query, ":ColumnNames", columnNames)
	query = strings.ReplaceAll(query, ":PlaceHolderVar", placeHolderVar)
	query = strings.ReplaceAll(query, ":ExtendedCondition", extendedCondition)
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	locationsBySubscriptionID := make(map[string][]string)
	gradesBySubscriptionID := make(map[string]string)
	studentSubscriptionsMap := make(map[string]*StudentSubscription)
	for rows.Next() {
		record := StudentSubscription{}
		var locationID pgtype.Text
		var gradeID pgtype.Text

		scanFields := database.GetScanFields(&record, database.GetFieldNames(&StudentSubscription{}))
		scanFields = append(scanFields, &locationID, &gradeID)
		if err = rows.Scan(scanFields...); err != nil {
			return nil, err
		}
		subID := record.StudentSubscriptionID.String
		studentSubscriptionsMap[subID] = &record

		if locationID.Status == pgtype.Present {
			locationsBySubscriptionID[subID] = append(locationsBySubscriptionID[subID], locationID.String)
		}
		if gradeID.Status == pgtype.Present {
			gradesBySubscriptionID[subID] = gradeID.String
		}
	}

	studentSubscriptions := make(StudentSubscriptions, 0, len(studentSubscriptionsMap))
	for k := range studentSubscriptionsMap {
		studentSubscriptions = append(studentSubscriptions, studentSubscriptionsMap[k])
	}

	return studentSubscriptions.ToListStudentSubscriptionEntities(locationsBySubscriptionID, gradesBySubscriptionID), nil
}

func (s *StudentSubscriptionRepo) RetrieveStudentSubscription(ctx context.Context, db database.QueryExecer, params *payloads.ListStudentSubScriptionsArgs) ([]*domain.StudentSubscription, uint32, string, uint32, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonStudentSubScriptionRepo.RetrieveStudentSubscription")
	defer span.End()
	// query for offset
	flrSubQuery := `WITH filter_student_subscriptions AS(select distinct lss.student_subscription_id , lss.course_id , lss.student_id , lss.created_at, lss.start_at, lss.end_at :baseTable :flrCondition :studentSubscriptionOneCondition :flrOrder ) `
	// query for selection
	creationQuerySelection := fmt.Sprintf(`:flrSubQuery :selectQuery FROM filter_student_subscriptions flr JOIN user_basic_info ubi ON flr.student_id = ubi.user_id and ubi.user_role = '%s' :whereAfterJoinStudents `, constants.UserRoleStudent)
	totalQuerySelection := `:flrSubQuery :selectQuery FROM filter_student_subscriptions flr :joinStudents :whereAfterJoinStudents `

	detailQuerySelectSection := "select flr.student_subscription_id, flr.course_id , flr.student_id, ubi.grade_id, flr.start_at, flr.end_at"
	flrOrder := "order by lss.created_at DESC, lss.student_subscription_id DESC"
	totalCountSelectSection := "select count(flr.student_subscription_id)"

	baseTable := "FROM lesson_student_subscriptions lss JOIN student_enrollment_status_history sesh ON sesh.student_id = lss.student_id "
	flrCondition := "WHERE lss.deleted_at IS NULL AND sesh.enrollment_status = ANY('{STUDENT_ENROLLMENT_STATUS_POTENTIAL, STUDENT_ENROLLMENT_STATUS_ENROLLED}') and sesh.deleted_at IS null and ( sesh.end_date is null or sesh.end_date > now() ) "
	whereAfterJoinStudents := ""
	joinStudents := ""

	args := []interface{}{}
	paramsNum := len(args)
	whereCondition := ""

	if !params.LessonDate.IsZero() {
		paramsNum++
		whereCondition += fmt.Sprintf(` AND lss.start_at::date <= ($%d::timestamptz)::date AND lss.end_at::date >= ($%d::timestamptz)::date`, paramsNum, paramsNum)
		args = append(args, params.LessonDate)
	}

	if len(params.CourseIDs) > 0 {
		paramsNum++
		whereCondition += fmt.Sprintf(` AND lss.course_id = ANY($%d)`, paramsNum)
		args = append(args, &params.CourseIDs)
	}

	if len(params.StudentIDWithCourseIDs) > 0 {
		query := " AND (lss.student_id,lss.course_id) IN (:PlaceHolderVar) "

		// will like ["($1, $2)", "($3, $4)", ...]
		studentIDCourseIDQuery := make([]string, 0, len(params.StudentIDWithCourseIDs)/2)
		for i := 0; i < len(params.StudentIDWithCourseIDs); i += 2 {
			studentID := params.StudentIDWithCourseIDs[i]
			courseID := params.StudentIDWithCourseIDs[i+1]
			args = append(args, &studentID, &courseID)
			studentIDCourseIDQuery = append(studentIDCourseIDQuery, fmt.Sprintf("($%d, $%d)", paramsNum+1, paramsNum+2))
			paramsNum += 2
		}

		// placeHolderVar will like ($1, $2), ($3, $4), ($5, $6), ....
		placeHolderVar := strings.Join(studentIDCourseIDQuery, ", ")
		query = strings.ReplaceAll(query, ":PlaceHolderVar", placeHolderVar)
		whereCondition += query
	}

	if len(params.StudentSubscriptionIDs) > 0 {
		paramsNum++
		whereCondition += fmt.Sprintf(` AND lss.student_subscription_id = ANY($%d)`, paramsNum)
		args = append(args, &params.StudentSubscriptionIDs)
	}

	// just using GradesV2 now
	if len(params.GradesV2) > 0 {
		paramsNum++
		whereAfterJoinStudents += fmt.Sprintf(` AND ubi.grade_id = ANY($%d)`, paramsNum)
		joinStudents = fmt.Sprintf("JOIN user_basic_info ubi ON flr.student_id = ubi.user_id and ubi.user_role = '%s' ", constants.UserRoleStudent)
		args = append(args, &params.GradesV2)
	}

	if params.KeyWord != "" {
		paramsNum++
		baseTable += ` left join user_basic_info ubi on ubi.user_id = lss.student_id `
		whereCondition += fmt.Sprintf(` AND (nospace(ubi."name") ILIKE nospace(CONCAT('%%',$%d::text,'%%'))
				OR nospace(ubi."full_name_phonetic") ILIKE nospace(CONCAT('%%',$%d::text,'%%'))
				)`, paramsNum, paramsNum)
		args = append(args, &params.KeyWord)
	}
	flrCondition += whereCondition
	countArgs := args
	// build whereAfterJoinStudents for placeholder creationQuerySelection
	creationQuerySelection = strings.Replace(creationQuerySelection, ":whereAfterJoinStudents", whereAfterJoinStudents, 1)

	// enough data to build query totalCountQuery

	// get list
	paramNumLimit := paramsNum + 1
	paramNumSchoolID := paramsNum + 2

	args = append(args, &params.Limit)

	studentSubscriptionOneCondition := ""

	if params.StudentSubscriptionID != "" {
		studentSubscriptionOneCondition += fmt.Sprintf(` AND (lss.created_at, lss.student_subscription_id) < ((SELECT created_at FROM lesson_student_subscriptions WHERE student_subscription_id = $%d LIMIT 1), $%d)`, paramNumSchoolID, paramNumSchoolID)
		args = append(args, &params.StudentSubscriptionID)
	}

	// build flrCondition and baseTable for placeholder flrSubQuery
	flrSubQuery = strings.Replace(flrSubQuery, ":flrCondition", flrCondition, 1)
	flrSubQuery = strings.Replace(flrSubQuery, ":baseTable", baseTable, 1)
	// flrSubQuery has flrOrder has been not built yet

	// build flrSubQuery for placeholder creationQuerySelection
	creationQuerySelection = strings.Replace(creationQuerySelection, ":flrSubQuery", flrSubQuery, 1)
	// creationQuerySelection has selectQuery, studentSubscriptionOneCondition and flrOrder have been not built yet

	// build placeholder listDetailQuery for get list detail from creationQuerySelection
	listDetailQuery := creationQuerySelection

	// build selectQuery and flrOrder for placeholder listDetailQuery
	listDetailQuery = strings.Replace(listDetailQuery, ":selectQuery", detailQuerySelectSection, 1)
	listDetailQuery = strings.Replace(listDetailQuery, ":flrOrder", flrOrder, 1)
	listDetailQuery = strings.Replace(listDetailQuery, ":studentSubscriptionOneCondition", studentSubscriptionOneCondition, 1)

	listDetailQuery += fmt.Sprintf(`LIMIT $%d`, paramNumLimit)

	var ss domain.StudentSubscriptions
	var rows pgx.Rows
	var err error
	rows, err = db.Query(ctx, listDetailQuery, args...)

	if err != nil {
		return nil, 0, "", 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var s StudentSubscription
		var gradeV2 pgtype.Text
		if err := rows.Scan(&s.StudentSubscriptionID, &s.CourseID, &s.StudentID, &gradeV2, &s.StartAt, &s.EndAt); err != nil {
			return nil, 0, "", 0, errors.Wrap(err, "rows.Scan")
		}
		ss = append(ss, s.ToStudentSubscriptionEntity().WithGrade(gradeV2.String))
	}

	if err := rows.Err(); err != nil {
		return nil, 0, "", 0, errors.Wrap(err, "rows.Err")
	}
	var preOffset pgtype.Text
	var preTotal pgtype.Int8
	if params.StudentSubscriptionID != "" {
		// build new sub query filter_student_subscriptions_ac with join filter_student_subscriptions and students
		temporaryTableValue := `with filter_student_subscriptions_ac as (:flrSubQuery :detailQuerySelectSection 
			FROM filter_student_subscriptions flr JOIN user_basic_info ON flr.student_id = ubi.user_id and ubi.user_role = 'student' :whereAfterJoinStudents )`
		previousDetailQuerySelectSection := detailQuerySelectSection + ", flr.created_at "
		temporaryTableValue = strings.Replace(temporaryTableValue, ":detailQuerySelectSection", previousDetailQuerySelectSection, 1)
		temporaryTableValue = strings.Replace(temporaryTableValue, ":flrSubQuery", flrSubQuery, 1)
		temporaryTableValue = strings.Replace(temporaryTableValue, ":whereAfterJoinStudents", whereAfterJoinStudents, 1)
		temporaryTableValue = strings.Replace(temporaryTableValue, ":studentSubscriptionOneCondition", "", 1)

		// flrOrder is contained in flrSubQuery
		temporaryTableValue = strings.Replace(temporaryTableValue, ":flrOrder", "", 1)

		query := temporaryTableValue + fmt.Sprintf(`
		, previous_sort as(
			select fsa.student_subscription_id, fsa.created_at,
					COUNT(*) OVER() AS total
			from filter_student_subscriptions_ac fsa
			where $%d::text is not NULL
					and (fsa.created_at, fsa.student_subscription_id) > ((SELECT created_at FROM lesson_student_subscriptions WHERE student_subscription_id = $%d LIMIT 1), $%d)
			order by fsa.created_at ASC, fsa.student_subscription_id ASC
			LIMIT $%d
		)
			select ps.student_subscription_id AS pre_offset, ps.total AS pre_total
			from previous_sort ps
			order by ps.created_at desc 
			limit 1
			`, paramNumSchoolID, paramNumSchoolID, paramNumSchoolID, paramNumLimit)
		if err := db.QueryRow(ctx, query, args...).Scan(&preOffset, &preTotal); err != nil {
			if err.Error() != pgx.ErrNoRows.Error() {
				return nil, 0, "", 0, errors.Wrap(err, "get previous err")
			}
		}
	}

	// get total

	var total pgtype.Int8

	totalCountQuery := totalQuerySelection
	totalCountQuery = strings.Replace(totalCountQuery, ":flrSubQuery", flrSubQuery, 1)
	totalCountQuery = strings.Replace(totalCountQuery, ":selectQuery", totalCountSelectSection, 1)
	totalCountQuery = strings.Replace(totalCountQuery, ":joinStudents", joinStudents, 1)
	totalCountQuery = strings.Replace(totalCountQuery, ":whereAfterJoinStudents", whereAfterJoinStudents, 1)
	totalCountQuery = strings.Replace(totalCountQuery, ":flrOrder", "", 1)
	totalCountQuery = strings.Replace(totalCountQuery, ":studentSubscriptionOneCondition", "", 1)

	if err := db.QueryRow(ctx, totalCountQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, "", 0, errors.Wrap(err, "get total err")
	}

	return ss, uint32(total.Int), preOffset.String, uint32(preTotal.Int), nil
}

func (s *StudentSubscriptionRepo) BulkUpsertStudentSubscription(ctx context.Context, db database.QueryExecer, subList domain.StudentSubscriptions) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionRepo.BulkUpsertStudentSubscription")
	defer span.End()

	excludedFields := []string{"deleted_at"}
	studentSubList, err := NewStudentSubscriptionListFromDomainList(subList)
	if err != nil {
		return err
	}

	queueFn := func(b *pgx.Batch, studentSub *StudentSubscription) {
		fieldNames := database.GetFieldNamesExcepts(studentSub, excludedFields)
		args := database.GetScanFields(studentSub, fieldNames)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT lesson_student_subscriptions_pkey
			DO UPDATE SET 
				start_at = $5::Date, 
				end_at = $6::Date, 
				course_slot = $7, 
				course_slot_per_week = $8, 
				student_first_name = $9, 
				student_last_name = $10,
				package_type = $11,
				updated_at = $13`,
			studentSub.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		b.Queue(query, args...)
	}

	b := &pgx.Batch{}
	for _, studentSubInfo := range studentSubList {
		if err := studentSubInfo.PreUpsert(); err != nil {
			return fmt.Errorf("got error on PreUpsert student subscription: %w", err)
		}

		queueFn(b, studentSubInfo)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(studentSubList); i++ {
		commandTag, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("failed to bulk upsert student subscription batchResults.Exec: %w", err)
		}
		if commandTag.RowsAffected() != 1 {
			return fmt.Errorf("student subscription not inserted/updated")
		}
	}

	return nil
}

func (s *StudentSubscriptionRepo) GetStudentSubscriptionIDByUniqueIDs(ctx context.Context, db database.QueryExecer, subscriptionID, studentID, courseID string) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionRepo.GetStudentSubscriptionIDByUniqueIDs")
	defer span.End()

	studentSub := &StudentSubscription{}
	var studentSubID pgtype.Text

	query := fmt.Sprintf(`
		SELECT student_subscription_id
		FROM %s 
		WHERE subscription_id = $1
		AND student_id = $2
		AND course_id = $3
		AND deleted_at IS NULL`,
		studentSub.TableName(),
	)

	if err := db.QueryRow(ctx, query, &subscriptionID, &studentID, &courseID).Scan(&studentSubID); err != nil && err != pgx.ErrNoRows {
		return "", fmt.Errorf("failed to query student subscription: %w", err)
	}

	return studentSubID.String, nil
}

func (s *StudentSubscriptionRepo) UpdateMultiStudentNameByStudents(ctx context.Context, db database.QueryExecer, users domain.Users) error {
	ctx, span := interceptors.StartSpan(ctx, "LessonTeacherRepo.UpdateMultiStudentNameByStudents")
	defer span.End()
	b := &pgx.Batch{}
	studentSubscriptionDTO := StudentSubscription{}
	strQuery := fmt.Sprintf(`
	UPDATE %s 
	SET updated_at = NOW(), 
	student_first_name = $2, student_last_name = $3
	WHERE student_id = $1 `, studentSubscriptionDTO.TableName())

	for _, user := range users {
		b.Queue(strQuery, user.ID, user.FirstName, user.LastName)
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

func (s *StudentSubscriptionRepo) RetrieveStudentPendingReallocate(ctx context.Context, db database.QueryExecer, params domain.RetrieveStudentPendingReallocateDto) ([]*domain.ReallocateStudent, uint32, error) {
	ctx, span := interceptors.StartSpan(ctx, "ReallocationRepo.RetrieveStudentPendingReallocate")
	defer span.End()

	selectQ := "select count(*) over() as total,rf.student_id ,rf.original_lesson_id ,rf.course_id,lss.start_at ,lss.end_at, ubi.grade_id, cl.class_id, lssap.location_id "

	baseTable := fmt.Sprintf(`from (select r.student_id, r.course_id, r.original_lesson_id
			from reallocation r where r.new_lesson_id is null and r.deleted_at is null) rf
	join lesson_student_subscriptions lss on lss.student_id = rf.student_id and lss.course_id = rf.course_id
	join lesson_student_subscription_access_path lssap on lssap.student_subscription_id = lss.student_subscription_id 
	join user_basic_info ubi on ubi.user_id = rf.student_id and ubi.user_role = '%s'
	left join (select cm.user_id ,c.course_id,c.class_id  
				from class_member cm join class c on c.class_id = cm.class_id where cm.deleted_at is null) cl on (cl.user_id = rf.student_id and cl.course_id = rf.course_id)
	join lessons l on l.lesson_id = rf.original_lesson_id `, constants.UserRoleStudent)

	orderByQ := "order by l.start_time desc "
	var args []interface{}
	paramsNum := len(args)
	whereQ := "where lss.deleted_at is null and lssap.deleted_at is null "
	if params.LessonDate.Unix() > 0 {
		paramsNum++
		whereQ += fmt.Sprintf("and (lss.start_at at time zone '%s')::date <= ($%d::timestamptz at time zone '%s')::date AND (lss.end_at at time zone '%s')::date >= ($%d::timestamptz at time zone '%s')::date ", params.Timezone, paramsNum, params.Timezone, params.Timezone, paramsNum, params.Timezone)
		args = append(args, params.LessonDate)
	}

	if len(params.CourseID) > 0 {
		paramsNum++
		whereQ += fmt.Sprintf("AND rf.course_id = ANY($%d) ", paramsNum)
		args = append(args, params.CourseID)
	}

	if len(params.LocationID) > 0 {
		paramsNum++
		whereQ += fmt.Sprintf("AND lssap.location_id = ANY($%d) ", paramsNum)
		args = append(args, params.LocationID)
	}

	if len(params.GradeID) > 0 {
		paramsNum++
		whereQ += fmt.Sprintf("AND ubi.grade_id = ANY($%d) ", paramsNum)
		args = append(args, params.GradeID)
	}

	if len(params.ClassID) > 0 {
		paramsNum++
		whereQ += fmt.Sprintf("AND cl.class_id = ANY($%d) ", paramsNum)
		args = append(args, params.ClassID)
	}

	if params.StartDate.Unix() > 0 {
		paramsNum++
		whereQ += fmt.Sprintf("AND (l.start_time at time zone '%s')::date >= ($%d::timestamptz at time zone '%s')::date ", params.Timezone, paramsNum, params.Timezone)
		args = append(args, params.StartDate)
	}

	if params.EndDate.Unix() > support.UnixToEnd {
		paramsNum++
		whereQ += fmt.Sprintf("AND (l.end_time at time zone '%s')::date <= ($%d::timestamptz at time zone '%s')::date ", params.Timezone, paramsNum, params.Timezone)
		args = append(args, params.EndDate)
	}

	if len(params.SearchKey) > 0 {
		paramsNum++
		baseTable += "left join user_basic_info ubi on ubi.user_id = rf.student_id "
		whereQ += fmt.Sprintf(`AND ( nospace(ubi."name") ILIKE nospace(concat('%%',$%d::text,'%%')) OR  nospace(ubi."full_name_phonetic") ILIKE nospace(concat('%%',$%d::text,'%%'))) `, paramsNum, paramsNum)
		args = append(args, params.SearchKey)
	}
	paramsNum++
	pagingQ := fmt.Sprintf("limit $%d offset $%d", paramsNum, paramsNum+1)
	args = append(args, params.Limit, params.Offset)

	retrieveStudentQ := selectQ + baseTable + whereQ + orderByQ + pagingQ
	rows, err := db.Query(ctx, retrieveStudentQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	results := []*domain.ReallocateStudent{}
	var total pgtype.Int8
	for rows.Next() {
		var (
			studentId        pgtype.Text
			originalLessonId pgtype.Text
			courseId         pgtype.Text
			startAt          pgtype.Timestamptz
			endAt            pgtype.Timestamptz
			gradeId          pgtype.Text
			classId          pgtype.Text
			locationId       pgtype.Text
		)
		if err = rows.Scan(
			&total,
			&studentId,
			&originalLessonId,
			&courseId,
			&startAt,
			&endAt,
			&gradeId,
			&classId,
			&locationId,
		); err != nil {
			return nil, 0, err
		}
		rs := &domain.ReallocateStudent{
			StudentId:        studentId.String,
			OriginalLessonID: originalLessonId.String,
			CourseID:         courseId.String,
			LocationID:       locationId.String,
			ClassID:          classId.String,
			GradeID:          gradeId.String,
			StartAt:          startAt.Time,
			EndAt:            endAt.Time,
		}
		results = append(results, rs)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	return results, uint32(total.Int), nil
}

func (s *StudentSubscriptionRepo) GetStudentCoursesAndClasses(ctx context.Context, db database.QueryExecer, studentID string) (*domain.StudentCoursesAndClasses, error) {
	query := `
	WITH class_temp AS (
		SELECT c.class_id, c.course_id , c."name"  
		FROM class_member cm 
		JOIN class c on c.class_id = cm.class_id 
		WHERE user_id = $1 AND cm.deleted_at is null 
		GROUP BY c.class_id, c.course_id 
	)
	SELECT ct.class_id, ct.name AS "class_name", c.course_id AS "course_id", c."name" AS "course_name"  
	FROM lesson_student_subscriptions lss
	LEFT OUTER JOIN courses c ON lss.course_id = c.course_id 
	LEFT OUTER JOIN class_temp ct ON lss.course_id = ct.course_id 
	WHERE lss.student_id = $1 AND lss.deleted_at is null`

	rows, err := db.Query(ctx, query, studentID)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	res := &domain.StudentCoursesAndClasses{
		StudentID: studentID,
		Courses:   nil,
		Classes:   nil,
	}
	courseIDsMap := make(map[string]bool)
	for rows.Next() {
		var classID pgtype.Text
		var className pgtype.Text
		var courseID pgtype.Text
		var courseName pgtype.Text

		if err = rows.Scan(&classID, &className, &courseID, &courseName); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		if _, ok := courseIDsMap[courseID.String]; !ok {
			res.Courses = append(res.Courses, &domain.StudentCoursesAndClassesCourses{
				CourseID: courseID.String,
				Name:     courseName.String,
			})
			courseIDsMap[courseID.String] = true
		}
		if classID.Status == pgtype.Present {
			res.Classes = append(res.Classes, &domain.StudentCoursesAndClassesClasses{
				ClassID:  classID.String,
				Name:     className.String,
				CourseID: courseID.String,
			})
		}
	}

	return res, nil
}

func (s *StudentSubscriptionRepo) GetAll(ctx context.Context, db database.QueryExecer) ([]*domain.EnrolledStudent, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonStudentSubScriptionRepo.GetAll")
	defer span.End()

	query := `select ls.student_id,ls.course_id,ls.start_at,ls.end_at,ap.location_id, s.enrollment_status from lesson_student_subscriptions ls 
	          join lesson_student_subscription_access_path ap ON ls.student_subscription_id = ap.student_subscription_id 
			  join student_enrollment_status_history s on s.student_id = ls.student_id and s.location_id = ap.location_id `

	whereClause := ` where ls.deleted_at is null and s.deleted_at is null and ap.deleted_at is null 
			  and s.enrollment_status = ANY('{STUDENT_ENROLLMENT_STATUS_POTENTIAL, STUDENT_ENROLLMENT_STATUS_ENROLLED}') `

	query += whereClause
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	enrolledStudent := []*domain.EnrolledStudent{}
	for rows.Next() {
		var (
			studentID, courseID, enrolledStatus, locationID pgtype.Text
			startAt, endAt                                  pgtype.Timestamptz
		)
		if err := rows.Scan(&studentID, &courseID, &startAt, &endAt, &locationID, &enrolledStatus); err != nil {
			return nil, err
		}
		enrolledStudent = append(enrolledStudent, &domain.EnrolledStudent{
			StudentID:        studentID.String,
			CourseID:         courseID.String,
			StartAt:          startAt.Time,
			EndAt:            endAt.Time,
			EnrollmentStatus: enrolledStatus.String,
			LocationID:       locationID.String,
		})
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return enrolledStudent, nil
}

func (s *StudentSubscriptionRepo) GetByStudentSubscriptionID(ctx context.Context, db database.QueryExecer, studentSubscriptionID string) (*domain.StudentSubscription, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionRepo.GetByStudentSubscriptionID")
	defer span.End()

	ss := &StudentSubscription{}
	fields, values := ss.FieldMap()

	query := fmt.Sprintf(`
		SELECT %s FROM %s WHERE student_subscription_id = $1 AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		ss.TableName(),
	)

	err := db.QueryRow(ctx, query, &studentSubscriptionID).Scan(values...)
	if err != nil {
		return nil, err
	}
	return ss.ToStudentSubscriptionEntity(), nil
}

func (s *StudentSubscriptionRepo) GetByStudentSubscriptionIDs(ctx context.Context, db database.QueryExecer, studentSubscriptionID []string) ([]*domain.StudentSubscription, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentSubscriptionRepo.GetByStudentSubscriptionIDs")
	defer span.End()
	ss := &StudentSubscription{}
	fields, _ := ss.FieldMap()

	query := fmt.Sprintf(`
		SELECT %s FROM %s WHERE student_subscription_id = ANY($1) AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		ss.TableName(),
	)
	rows, err := db.Query(ctx, query, &studentSubscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := []*domain.StudentSubscription{}
	for rows.Next() {
		ss := &StudentSubscription{}
		_, values := ss.FieldMap()
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		res = append(res, ss.ToStudentSubscriptionEntity())
	}
	return res, nil
}
