package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type LessonAllocationRepo struct {
}

func (l *LessonAllocationRepo) GetLessonAllocation(ctx context.Context, db database.QueryExecer, filter domain.LessonAllocationFilter) ([]*domain.AllocatedStudent, map[string]uint32, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonAllocation.GetLessonAllocation")
	defer span.End()

	withQuery := `WITH lss AS (
		select lss.student_subscription_id, lss.student_id,lss.course_id, lssap.location_id, lss.start_at, lss.end_at, cls.product_type_schedule,
		case WHEN cls.product_type_schedule = 'PACKAGE_TYPE_FREQUENCY'  THEN calculate_purchased_slot_total_v2(sc.course_slot_per_week::smallint,lss.start_at::date,lss.end_at::date,lss.course_id ,lssap.location_id,lss.student_id)
			 WHEN cls.product_type_schedule = 'PACKAGE_TYPE_SCHEDULED'  THEN calculate_purchased_slot_total_v2(cls.frequency::smallint,lss.start_at::date,lss.end_at::date,lss.course_id ,lssap.location_id,lss.student_id)
			 WHEN cls.product_type_schedule = 'PACKAGE_TYPE_SLOT_BASED' THEN sc.course_slot 		 				
			 WHEN cls.product_type_schedule = 'PACKAGE_TYPE_ONE_TIME'   THEN cls.total_no_lessons
			 ELSE 0 END as purchased_slot_total
		from lesson_student_subscriptions lss 
		join lesson_student_subscription_access_path lssap on lss.student_subscription_id = lssap.student_subscription_id
		left join student_course sc on sc.student_id = lss.student_id and sc.course_id = lss.course_id and sc.location_id = lssap.location_id 
		left join course_location_schedule cls on lss.course_id = cls.course_id and lssap.location_id = cls.location_id `

	whereWithQuery := "where lss.deleted_at is null and lssap.deleted_at is null and sc.deleted_at is null "

	selectBaseQuery := "select lss.student_subscription_id, student_id , lss.course_id , lss.location_id ,start_at ,end_at, lss.product_type_schedule, coalesce(assigned_slot_tmp.assigned_slot, 0) assigned_slot, purchased_slot_total "

	baseQuery := `from lss
				  left join (select user_id,lm.course_id,count(case WHEN attendance_status != 'STUDENT_ATTEND_STATUS_REALLOCATE' THEN 1
				  													WHEN attendance_status = 'STUDENT_ATTEND_STATUS_REALLOCATE' and r.new_lesson_id is null THEN 1
				   													ELSE null END )::int as "assigned_slot"
							from lesson_members lm
							left join reallocation r on r.original_lesson_id = lm.lesson_id and r.student_id = lm.user_id
							where lm.deleted_at is null and lm.course_id is not null :whereAssignedSlotQuery 
							and not exists (select 1 from lessons l where l.scheduling_status = 'LESSON_SCHEDULING_STATUS_CANCELED' and l.deleted_at is null and lm.lesson_id = l.lesson_id)
							group by user_id,lm.course_id ) assigned_slot_tmp on (assigned_slot_tmp.user_id = lss.student_id and assigned_slot_tmp.course_id = lss.course_id) `
	args := []interface{}{}
	params := len(args)

	if len(filter.KeySearch) > 0 {
		params++
		withQuery += "left join user_basic_info ubi on ubi.user_id = lss.student_id "
		args = append(args, filter.KeySearch)
		whereWithQuery += fmt.Sprintf(`and (lower(ubi."name") like lower(concat('%%',$%d::text,'%%')) OR  lower(ubi."full_name_phonetic") like lower(concat('%%',$%d::text,'%%'))) `, params, params)
	}

	if len(filter.LocationID) > 0 {
		params++
		args = append(args, filter.LocationID)
		whereWithQuery += fmt.Sprintf("and lssap.location_id = any($%d) ", params)
	}

	whereAssignedSlotQuery := ""
	if len(filter.CourseID) > 0 {
		params++
		args = append(args, filter.CourseID)
		whereWithQuery += fmt.Sprintf("and lss.course_id = any($%d) ", params)
		whereAssignedSlotQuery += fmt.Sprintf(" and lm.course_id = any($%d) ", params)
	}

	if len(filter.CourseTypeID) > 0 {
		params++
		args = append(args, filter.CourseTypeID)
		withQuery += "join courses c on c.course_id = lss.course_id "
		whereWithQuery += fmt.Sprintf("and c.course_type_id = any($%d) ", params)
	}

	if len(filter.TeachingMethod) > 0 {
		params++
		teachingMethod := []string{}
		for _, tm := range filter.TeachingMethod {
			teachingMethod = append(teachingMethod, string(tm))
		}
		args = append(args, teachingMethod)
		if len(filter.CourseTypeID) == 0 {
			withQuery += "join courses c on c.course_id = lss.course_id "
		}
		whereWithQuery += fmt.Sprintf("and c.teaching_method = any($%d) ", params)
	}

	if len(filter.ProductID) > 0 {
		params++
		args = append(args, filter.ProductID)
		withQuery += "join student_packages sp on sp.student_package_id = lss.subscription_id "
		whereWithQuery += fmt.Sprintf("and sp.package_id = any($%d) ", params)
	}

	atTimezone := ""
	if filter.StartDate.Unix() > 0 || filter.EndDate.Unix() > support.UnixToEnd {
		params++
		atTimezone += fmt.Sprintf("at time zone $%d", params)
		args = append(args, filter.TimeZone)
	}

	if filter.StartDate.Unix() > support.UnixToEnd {
		params++
		args = append(args, filter.StartDate)
		whereWithQuery += fmt.Sprintf("and (lss.end_at %s)::date >= ($%d %s)::date ", atTimezone, params, atTimezone)
	}

	if filter.EndDate.Unix() > support.UnixToEnd {
		params++
		args = append(args, filter.EndDate)
		whereWithQuery += fmt.Sprintf("and (lss.start_at %s)::date <= ($%d %s)::date ", atTimezone, params, atTimezone)
	}

	if filter.IsOnlyReallocation {
		whereWithQuery += "and (lss.student_id,lss.course_id) in (select student_id,course_id from reallocation where new_lesson_id is null and deleted_at is null) "
	}

	if filter.IsClassUnassigned {
		withQuery += `left join (select cm.user_id ,c.course_id  ,c.class_id
			from class_member cm join class c on c.class_id = cm.class_id where cm.deleted_at is null) cl 
			on (cl.user_id = lss.student_id and cl.course_id = lss.course_id )
	    left join (
	        select student_id , course_id,class_id from reserve_class where deleted_at is null 
	    ) rc on rc.student_id = lss.student_id  and rc.course_id = lss.course_id `
		if len(filter.CourseTypeID) == 0 || len(filter.TeachingMethod) == 0 {
			withQuery += "join courses c on c.course_id = lss.course_id "
		}
		whereWithQuery += "and cl.class_id is null and rc.class_id is null and c.teaching_method = 'COURSE_TEACHING_METHOD_GROUP' "
	}

	whereBaseQuery := ""
	switch filter.LessonAllocationStatus {
	case domain.NoneAssigned:
		whereBaseQuery += "where coalesce(assigned_slot_tmp.assigned_slot, 0) = 0 "
	case domain.PartiallyAssigned:
		whereBaseQuery += "where assigned_slot_tmp.assigned_slot is not null and (coalesce(assigned_slot_tmp.assigned_slot, 0) - purchased_slot_total )::int < 0 "
	case domain.OverAssigned:
		whereBaseQuery += "where (coalesce(assigned_slot_tmp.assigned_slot, 0) - purchased_slot_total )::int > 0 "
	case domain.FullyAssigned:
		whereBaseQuery += "where assigned_slot_tmp.assigned_slot is not null and (coalesce(assigned_slot_tmp.assigned_slot, 0) - purchased_slot_total )::int = 0 "
	}

	orderQuery := "order by start_at asc, location_id asc, lss.course_id asc "
	pagingQuery := fmt.Sprintf("LIMIT %d OFFSET %d", filter.Limit, filter.Offset)
	baseQuery = strings.Replace(baseQuery, ":whereAssignedSlotQuery", whereAssignedSlotQuery, 1)

	query := strings.Builder{}
	query.WriteString(withQuery)
	query.WriteString(whereWithQuery + ")")
	query.WriteString(selectBaseQuery)
	query.WriteString(baseQuery)
	query.WriteString(whereBaseQuery)
	query.WriteString(orderQuery)
	query.WriteString(pagingQuery)
	rows, err := db.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	allocatedStudent := []*domain.AllocatedStudent{}
	for rows.Next() {
		var (
			studentSubscriptionID, studentID, courseID, locationID, productTypeSchedule pgtype.Text
			startDate, endDate                                                          pgtype.Timestamptz
			assignedSlot, purchasedSlot                                                 pgtype.Int4
		)
		if err := rows.Scan(
			&studentSubscriptionID,
			&studentID,
			&courseID,
			&locationID,
			&startDate,
			&endDate,
			&productTypeSchedule,
			&assignedSlot,
			&purchasedSlot,
		); err != nil {
			return nil, nil, errors.Wrap(err, "rows.Scan")
		}
		allocatedStudent = append(allocatedStudent, &domain.AllocatedStudent{
			StudentSubscriptionID: studentSubscriptionID.String,
			StudentID:             studentID.String,
			CourseID:              courseID.String,
			LocationID:            locationID.String,
			StartTime:             startDate.Time,
			EndTime:               endDate.Time,
			ProductTypeSchedule:   productTypeSchedule.String,
			AssignedSlot:          assignedSlot.Int,
			PurchasedSlot:         purchasedSlot.Int,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, errors.Wrap(err, "rows.Err")
	}

	countQuery := ` select case WHEN coalesce(assigned_slot_tmp.assigned_slot, 0) = 0 THEN 'NONE_ASSIGNED'
								WHEN (coalesce(assigned_slot_tmp.assigned_slot, 0) - purchased_slot_total ) < 0 THEN 'PARTIALLY_ASSIGNED'
  								WHEN (coalesce(assigned_slot_tmp.assigned_slot, 0) - purchased_slot_total ) > 0 THEN 'OVER_ASSIGNED'
  								WHEN (coalesce(assigned_slot_tmp.assigned_slot, 0) - purchased_slot_total ) = 0 THEN 'FULLY_ASSIGNED'
   								END as lesson_allocation_status , count(*) `
	countTotalQuery := strings.Builder{}
	countTotalQuery.WriteString(withQuery)
	countTotalQuery.WriteString(whereWithQuery + ")")
	countTotalQuery.WriteString(countQuery)
	countTotalQuery.WriteString(baseQuery)
	countTotalQuery.WriteString("group by lesson_allocation_status")
	allocationStatusMap, err := l.countTotal(ctx, db, countTotalQuery.String(), args)
	return allocatedStudent, allocationStatusMap, err
}

func (l *LessonAllocationRepo) countTotal(ctx context.Context, db database.QueryExecer, query string, args []interface{}) (map[string]uint32, error) {
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	lessonAllocationStatus := make(map[string]uint32, 0)
	for rows.Next() {
		var (
			allocationStatus pgtype.Text
			count            pgtype.Int8
		)
		if err := rows.Scan(&allocationStatus, &count); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonAllocationStatus[allocationStatus.String] = uint32(count.Int)
	}
	return lessonAllocationStatus, nil
}

func (l *LessonAllocationRepo) GetByStudentSubscriptionAndWeek(ctx context.Context, db database.QueryExecer, studentID, courseID string, academicWeekID []string) (map[string][]*domain.LessonAllocationInfo, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonAllocation.GetByStudentSubscriptionAndWeek")
	defer span.End()

	query := `with lesson_member as (select * from lesson_members lm where user_id = $1 and course_id = $2 and deleted_at is null )
	select academic_week_id , l.lesson_id,l.start_time ,l.end_time,lm.attendance_status, l.scheduling_status, l.center_id , l.teaching_method, lr.lesson_report_id, l.is_locked
	from lesson_member lm
	join lessons l on l.lesson_id = lm.lesson_id 
	left join lesson_reports lr on l.lesson_id = lr.lesson_id 
	join ( SELECT academic_week_id,start_date,end_date from academic_week where academic_week_id = any($3)) weeks
	on (l.start_time at time zone 'Asia/Ho_Chi_Minh')::date >= weeks.start_date and (l.end_time at time zone 'Asia/Ho_Chi_Minh')::date <= weeks.end_date
	where l.deleted_at is null and lr.deleted_at is null `

	rows, err := db.Query(ctx, query, studentID, courseID, academicWeekID)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()
	lessonAllocationByWeeks := LessonAllocationByWeeks{}
	for rows.Next() {
		lessonAllocationByWeek := &LessonAllocationByWeek{}
		_, values := lessonAllocationByWeek.FieldMap()
		if err := rows.Scan(values...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		lessonAllocationByWeeks = append(lessonAllocationByWeeks, lessonAllocationByWeek)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return lessonAllocationByWeeks.ToLessonAllocation(), nil
}

func (l *LessonAllocationRepo) CountAssignedSlotPerStudentCourse(ctx context.Context, db database.QueryExecer, studentID, courseID string) (uint32, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonAllocationRepo.CountAssignedSlotPerStudentCourse")
	defer span.End()

	query := `with lesson_member as (
		select lm.lesson_id,user_id ,lm.course_id,attendance_status  
		from lesson_members lm 
		join lessons l on l.lesson_id = lm.lesson_id 
		where user_id = $1 and lm.course_id = $2 and l.scheduling_status != 'LESSON_SCHEDULING_STATUS_CANCELLED' and l.deleted_at is null and lm.deleted_at is null )
	select count(case WHEN attendance_status != 'STUDENT_ATTEND_STATUS_REALLOCATE' THEN 1
					  WHEN attendance_status = 'STUDENT_ATTEND_STATUS_REALLOCATE' and r.new_lesson_id is null THEN 1
					  ELSE null END )::int as "assigned_slot" 
	from lesson_member lm
	join lessons l on l.lesson_id = lm.lesson_id
	left join reallocation r on r.original_lesson_id = lm.lesson_id and r.student_id = lm.user_id 
	where l.deleted_at is null and l.scheduling_status  != 'LESSON_SCHEDULING_STATUS_CANCELED'
	group by user_id,lm.course_id`
	var assignedSlot pgtype.Int4

	err := db.QueryRow(ctx, query, studentID, courseID).Scan(&assignedSlot)
	if err == pgx.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, errors.Wrap(err, "db.QueryRow")
	}
	return uint32(assignedSlot.Int), nil
}

func (l *LessonAllocationRepo) CountPurchasedSlotPerStudentSubscription(ctx context.Context, db database.QueryExecer, freq uint8, startTime, endTime time.Time, courseID, locationID, studentID string) (uint32, error) {
	ctx, span := interceptors.StartSpan(ctx, "LessonAllocationRepo.CountPurchasedSlotPerStudentSubscription")
	defer span.End()
	args := []interface{}{freq, startTime, endTime, courseID, locationID, studentID}
	query := "select calculate_purchased_slot_total_v2($1::smallint,$2::date,$3::date,$4,$5,$6)"
	var total pgtype.Int2
	err := db.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, errors.Wrap(err, "db.QueryRow")
	}
	return uint32(total.Int), nil
}
