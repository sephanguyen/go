package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type AssignedStudentRepo struct{}

const (
	StudentCourseSlotTable      string = "student_course_slot_info_fn($1, $2, $3, $4, $5, $6, $7, NULL, $8) at "
	StudentCourseRecurringTable string = "student_course_recurring_slot_info_fn($1, $2, $3, $4, $5, $6, $7, NULL, $8) at "
)

const getStudentCourseSlot = `
	SELECT at.student_id, at.course_id, at.location_id, at.student_start_date as start_date, at.student_end_date as end_date,
		(at.student_start_date || ' - ' || at.student_end_date) as duration, 
		at.purchased_slot, at.assigned_slot, at.slot_gap, at.status, at.unique_id
	FROM %s
	`

const getStudentCourseRecurringSlot = `
	SELECT at.student_id, at.course_id, at.location_id, at.week_start as start_date, at.week_end as end_date,
		(at.week_start || ' - ' || at.week_end) as duration, 
		at.purchased_slot, at.assigned_slot, at.slot_gap, at.status, at.unique_id
	FROM %s
	`

const totalStudentCourseRecurringSlot = `
	SELECT COUNT((at.unique_id || '_' || at.week_start))
	FROM %s
	`

const totalStudentCourseSlot = ` 
	SELECT COUNT(at.unique_id)
	FROM %s
	`
const orderbyStudentCourseSlotASC = ` ORDER BY at.student_start_date ASC, at.course_id ASC, at.student_id ASC, at.unique_id ASC `
const orderbyStudentRecurringSlotASC = ` ORDER BY at.week_start ASC, at.course_id ASC, at.student_id ASC, (at.unique_id || '_' || at.week_start) ASC `
const orderbyStudentCourseSlotDESC = ` ORDER BY at.student_start_date DESC, at.course_id DESC, at.student_id DESC, at.unique_id DESC `
const orderbyStudentRecurringSlotDESC = ` ORDER BY at.week_start DESC, at.course_id DESC, at.student_id DESC, (at.unique_id || '_' || at.week_start) DESC `

func (ar *AssignedStudentRepo) GetAssignedStudentList(ctx context.Context, db database.QueryExecer, params *payloads.GetAssignedStudentListArg) ([]*domain.AssignedStudent, uint32, string, uint32, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignedStudentRepo.GetAssignedStudentList")
	defer span.End()

	var baseTable, orderbyASC, where, query, queryTotal, table, preQuery string
	var total pgtype.Int8
	purchaseMethod := params.PurchaseMethod
	paramDtos := ToListAsgStudentArgsDto(params)

	if purchaseMethod == string(domain.PurchaseMethodSlot) {
		baseTable = fmt.Sprintf(getStudentCourseSlot, StudentCourseSlotTable)
		queryTotal = fmt.Sprintf(totalStudentCourseSlot, StudentCourseSlotTable)
		where = fmt.Sprintf(" WHERE (current_timestamp at time zone '%s')::DATE < at.student_end_date ", paramDtos.Timezone.String)
		orderbyASC = orderbyStudentCourseSlotASC
	} else if purchaseMethod == string(domain.PurchaseMethodRecurring) {
		baseTable = fmt.Sprintf(getStudentCourseRecurringSlot, StudentCourseRecurringTable)
		queryTotal = fmt.Sprintf(totalStudentCourseRecurringSlot, StudentCourseRecurringTable)
		where = fmt.Sprintf(" WHERE (current_timestamp at time zone '%s')::DATE < at.week_end ", paramDtos.Timezone.String)
		orderbyASC = orderbyStudentRecurringSlotASC
	}

	args := []interface{}{
		&paramDtos.KeyWord,
		&paramDtos.Students,
		&paramDtos.Courses,
		&paramDtos.LocationIDs,
		&paramDtos.FromDate,
		&paramDtos.ToDate,
		&paramDtos.Timezone,
	}

	paramsNum := len(args)

	if paramDtos.AssignedStudentStatus.Status == pgtype.Present {
		paramsNum++
		where += fmt.Sprintf(` AND at.status = ANY($%d)`, paramsNum)
		args = append(args, &paramDtos.AssignedStudentStatus)
	}

	// get total
	queryTotal += where
	if err := db.QueryRow(ctx, queryTotal, args...).Scan(&total); err != nil {
		return nil, 0, "", 0, errors.Wrap(err, "get total err")
	}

	// get list
	if len(paramDtos.StudentSubscriptionID.String) > 0 {
		// get ids for paging
		var queryIDsPaging string
		var startDate pgtype.Date
		var courseID, studentID, uniqueID pgtype.Text
		pagingParamsCount := 4
		pagingArgs := make([]interface{}, 0, pagingParamsCount)

		if purchaseMethod == string(domain.PurchaseMethodRecurring) {
			queryIDsPaging = ` SELECT week_start, course_id, student_id 
					FROM student_course_recurring_slot_info_fn(null, null, null, null, null, null, $1, $2, $3) 
					WHERE (unique_id || '_' || week_start ) = $4
					LIMIT 1`
			uniqueID = database.Text(strings.Split(paramDtos.StudentSubscriptionID.String, "_")[0])
			pagingArgs = append(pagingArgs, &paramDtos.Timezone, &uniqueID, true, &paramDtos.StudentSubscriptionID)

			where += ` AND ((week_start, course_id, student_id, (unique_id || '_' || week_start )) > ( `
		} else {
			queryIDsPaging = ` SELECT student_start_date, course_id, student_id  
				FROM student_course_slot_info_fn(null, null, null, null, null, null, $1, $2, $3)
				WHERE unique_id = $2
				LIMIT 1`
			pagingArgs = append(pagingArgs, &paramDtos.Timezone, &paramDtos.StudentSubscriptionID, true)

			where += ` AND ((student_start_date, course_id, student_id, unique_id) > ( `
		}
		if err := db.QueryRow(ctx, queryIDsPaging, pagingArgs...).Scan(&startDate, &courseID, &studentID); err != nil {
			return nil, 0, "", 0, errors.Wrap(err, "get ids for paging")
		}

		for i := 0; i < pagingParamsCount; i++ {
			paramsNum++
			if i == (pagingParamsCount - 1) {
				where += fmt.Sprintf(` $%v `, paramsNum)
			} else {
				where += fmt.Sprintf(` $%v, `, paramsNum)
			}
		}
		where += ` )) `
		args = append(args, &startDate, &courseID, &studentID, &paramDtos.StudentSubscriptionID)
	}

	args = append(args, &paramDtos.Limit)
	limit := paramsNum + 1
	query = baseTable + where + orderbyASC + fmt.Sprintf(" LIMIT $%d ", limit)

	asgStudents := AsgStudents{}
	var rows pgx.Rows
	var err error
	rows, err = db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, "", 0, err
	}
	defer rows.Close()

	for rows.Next() {
		student := asgStudents.Add()
		fields := database.GetFieldNames(student)
		scanFields := database.GetScanFields(student, fields)
		if err := rows.Scan(scanFields...); err != nil {
			return nil, 0, "", 0, errors.Wrap(err, "rows.Scan")
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, "", 0, errors.Wrap(err, "rows.Err")
	}

	var preTotal = pgtype.Int8{Int: 0, Status: pgtype.Present}
	var preOffset = pgtype.Text{String: "", Status: pgtype.Present}

	if len(paramDtos.StudentSubscriptionID.String) > 0 {
		if purchaseMethod == string(domain.PurchaseMethodRecurring) {
			table = StudentCourseRecurringTable
			preQuery = fmt.Sprintf(`WITH prequery as (
				select (unique_id || '_' || week_start) as pre_offset_id, 
					count(*) OVER() AS total, 
					row_number() over() as row_num,
					at.week_start, 
					at.course_id, 
					at.student_id, 
					at.unique_id
				FROM %s`, table) + where + orderbyStudentRecurringSlotDESC + fmt.Sprintf(" LIMIT $%d ", limit) +
				" ) SELECT pre_offset_id, total FROM prequery ORDER BY week_start ASC, course_id ASC, student_id ASC, unique_id ASC LIMIT 1 "
			preQuery = strings.Replace(preQuery, "(week_start, course_id, student_id, (unique_id || '_' || week_start )) > ", "(week_start, course_id, student_id, (unique_id || '_' || week_start )) < ", 1)
		} else {
			table = StudentCourseSlotTable
			preQuery = fmt.Sprintf(`WITH prequery as (
				select at.unique_id as pre_offset_id, 
					count(*) OVER() AS total, 
					at.student_start_date, 
					at.course_id, 
					at.student_id, 
					at.unique_id
				FROM %s`, table) + where + orderbyStudentCourseSlotDESC + fmt.Sprintf(" LIMIT $%d ", limit) +
				" ) SELECT pre_offset_id, total FROM prequery ORDER BY student_start_date ASC, course_id ASC, student_id ASC, unique_id ASC LIMIT 1 "
			preQuery = strings.Replace(preQuery, "(student_start_date, course_id, student_id, unique_id) > ", "(student_start_date, course_id, student_id, unique_id) < ", 1)
		}

		if err := db.QueryRow(ctx, preQuery, args...).Scan(&preOffset, &preTotal); err != nil {
			if err != pgx.ErrNoRows {
				return nil, 0, "", 0, errors.Wrap(err, "get previous err")
			}
		}
	}

	res := []*domain.AssignedStudent{}
	for _, student := range asgStudents {
		studentEntity := domain.NewAssignedStudent().
			WithID(student.StudentID.String).
			WithCourseID(student.CourseID.String).
			WithLocationID(student.LocationID.String).
			WithDuration(student.Duration.String).
			WithPurchaseSlot(student.PurchaseSlot.Int).
			WithAssignedSlot(student.AssignedSlot.Int).
			WithSlotGap(student.SlotGap.Int).
			WithAssignedStatus(domain.AssignedStudentStatus(student.Status.String)).
			WithStudentSubscriptionID(student.StudentSubscriptionID.String).
			BuildDraft()
		res = append(res, studentEntity)
	}

	return res, uint32(total.Int), preOffset.String, uint32(preTotal.Int), nil
}

func (l *AssignedStudentRepo) GetStudentAttendance(ctx context.Context, db database.QueryExecer, filter domain.GetStudentAttendanceParams) ([]*domain.StudentAttendance, uint32, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonMemberRepo.GetByFilter")
	defer span.End()

	startDate := filter.StartDate
	endDate := filter.EndDate
	yearStartDate := filter.YearStartDate
	yearEndDate := filter.YearEndDate
	tz := filter.Timezone

	selectQ := "SELECT count(*) over() as total, lm.lesson_id,lm.user_id,lm.attendance_status, lm.course_id, lssap.location_id "
	baseTable := `FROM lessons l join lesson_members lm on l.lesson_id = lm.lesson_id
				  join lesson_student_subscriptions lss on lss.student_id = lm.user_id and lss.course_id = lm.course_id
		          join lesson_student_subscription_access_path lssap on lssap.student_subscription_id = lss.student_subscription_id  `

	whereQ := "WHERE l.deleted_at is null and lm.deleted_at is null and lss.deleted_at is null and lssap.deleted_at is null "
	orderBy := "ORDER BY l.start_time::date desc, l.center_id desc, l.lesson_id, lm.user_id "
	args := []interface{}{}

	if filter.IsFilterByCurrentYear {
		if yearStartDate.After(startDate) {
			startDate = yearStartDate
		}

		if endDate.Unix() <= support.UnixToEnd || endDate.After(yearEndDate) {
			endDate = yearEndDate
		}
		args = append(args, yearStartDate, yearEndDate)
		whereQ += fmt.Sprintf(`AND (lss.start_at at time zone '%s')::date >= ($%d::timestamptz at time zone '%s')::date 
							   AND (lss.end_at at time zone '%s')::date <= ($%d::timestamptz at time zone '%s')::date 
		`, tz, len(args)-1, tz, tz, len(args), tz)
	}

	if startDate.Unix() > 0 {
		args = append(args, startDate)
		whereQ += fmt.Sprintf("AND (l.start_time at time zone '%s')::date >= ($%d::timestamptz at time zone '%s')::date ", tz, len(args), tz)
	}

	if endDate.Unix() > support.UnixToEnd {
		args = append(args, endDate)
		whereQ += fmt.Sprintf("AND (l.end_time at time zone '%s')::date <= ($%d::timestamptz at time zone '%s')::date ", tz, len(args), tz)
	}

	if len(filter.StudentID) > 0 {
		args = append(args, filter.StudentID)
		whereQ += fmt.Sprintf("and lm.user_id = ANY($%d) ", len(args))
	}

	if len(filter.CourseID) > 0 {
		args = append(args, filter.CourseID)
		whereQ += fmt.Sprintf("and lm.course_id = ANY($%d) ", len(args))
	}

	if len(filter.LocationID) > 0 {
		args = append(args, filter.LocationID)
		whereQ += fmt.Sprintf("and lssap.location_id = ANY($%d) ", len(args))
	}

	if len(filter.AttendStatus) > 0 {
		args = append(args, filter.AttendStatus)
		whereQ += fmt.Sprintf("and lm.attendance_status = ANY($%d) ", len(args))
	}
	if len(filter.SearchKey) > 0 {
		args = append(args, &filter.SearchKey)
		baseTable += "join user_basic_info ubi on ubi.user_id = lm.user_id "
		whereQ += fmt.Sprintf(` AND nospace(ubi."name") ILIKE nospace(CONCAT('%%',$%d::text,'%%')) `, len(args))
	}

	limitHolder := len(args) + 1
	pagingQ := fmt.Sprintf("limit $%d offset $%d", limitHolder, limitHolder+1)
	args = append(args, filter.Limit, filter.Offset)
	query := selectQ + baseTable + whereQ + orderBy + pagingQ
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	studentAttendance := []*domain.StudentAttendance{}
	var total pgtype.Int8
	for rows.Next() {
		var (
			studentId    pgtype.Text
			lessonId     pgtype.Text
			attendStatus pgtype.Text
			courseId     pgtype.Text
			locationId   pgtype.Text
		)
		if err = rows.Scan(
			&total,
			&lessonId,
			&studentId,
			&attendStatus,
			&courseId,
			&locationId,
		); err != nil {
			return nil, 0, err
		}
		sa := &domain.StudentAttendance{
			StudentID:    studentId.String,
			CourseID:     courseId.String,
			LocationID:   locationId.String,
			LessonID:     lessonId.String,
			AttendStatus: attendStatus.String,
		}
		studentAttendance = append(studentAttendance, sa)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, errors.Wrap(err, "rows.Err")
	}
	return studentAttendance, uint32(total.Int), nil
}
