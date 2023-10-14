package virtualclassroom

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	vc_repo "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/jackc/pgtype"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) anExistingSetOfLessonsWithWait(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// sleep to make sure NATS sync data successfully from bob to lessonmgmt data
	time.Sleep(5 * time.Second)

	now := time.Now()
	lessonStatus := int32(0)
	for i := -12; i <= 12; i++ {
		lessonTime := now.Add(time.Duration(i) * 24 * time.Hour)

		req := s.CommonSuite.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(StepStateToContext(ctx, stepState), cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE)
		req.StartTime = timestamppb.New(lessonTime)
		req.EndTime = timestamppb.New(lessonTime.Add(30 * time.Minute))
		req.SchedulingStatus = lpb.LessonStatus(lpb.LessonStatus_value[lpb.LessonStatus_name[lessonStatus]])

		ctx, err := s.CommonSuite.UserCreateALessonWithRequestInLessonmgmt(StepStateToContext(ctx, stepState), req)
		stepState = StepStateFromContext(ctx)
		if err != nil || stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create lesson, err: %w | response err: %w", err, stepState.ResponseErr)
		}

		// update lesson end at
		if i%4 == 0 {
			if err := s.modifyLessonAddEndAt(ctx, stepState.CurrentLessonID); err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}

		// increment lesson status selection
		if lessonStatus == 3 {
			lessonStatus = 0
		} else {
			lessonStatus++
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) modifyLessonAddEndAt(ctx context.Context, lessonID string) error {
	time.Sleep(3 * time.Second)

	query := `UPDATE lessons 
		SET end_at = now(), updated_at = now()
		WHERE lesson_id = $1`

	cmdTag, err := s.LessonmgmtDBTrace.Exec(ctx, query, &lessonID)
	if err != nil {
		return fmt.Errorf("failed to update end_at of lesson %s db.Exec: %w", lessonID, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("lesson %s end_at was not updated", lessonID)
	}

	return err
}

func (s *suite) getArrangedLessonsUsingCurrentTime(ctx context.Context, currentTime time.Time) ([]string, error) {
	query := `SELECT lesson_id FROM lessons 
		WHERE end_time >= $1::timestamptz
		AND deleted_at IS NULL
		AND resource_path = $2
		ORDER BY start_time ASC, end_time ASC, lesson_id ASC
		LIMIT 40`

	rows, err := s.LessonmgmtDB.Query(ctx, query, database.Timestamptz(currentTime), database.Text("-2147483648"))
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var lessonID pgtype.Text
	var lessonIDs []string
	for rows.Next() {
		if err := rows.Scan(&lessonID); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		lessonIDs = append(lessonIDs, lessonID.String)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows.Err: %w", rows.Err())
	}

	return lessonIDs, nil
}

func (s *suite) userGetsListOfLessonsOnPageWithLimit(ctx context.Context, page, limit string) (context.Context, error) {
	time.Sleep(3 * time.Second)

	stepState := StepStateFromContext(ctx)
	now := time.Now().Add(-5 * 24 * time.Hour).UTC()

	convLimit, err := strconv.Atoi(limit)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to convert limit to a number: %w", err)
	}

	convPage, err := strconv.Atoi(page)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to convert page to a number: %w", err)
	}

	stepState.ArrangedLessonIDs, err = s.getArrangedLessonsUsingCurrentTime(ctx, now)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get arranged lesson IDs: %w", err)
	}

	// get the last item of the previous page
	// if 0, first page
	offsetIndex := convLimit * (convPage - 1)
	offsetString := ""
	if offsetIndex > 0 {
		if len(stepState.ArrangedLessonIDs) < offsetIndex {
			return StepStateToContext(ctx, stepState), fmt.Errorf("attempted to get offset string but lesson IDs is less than the offset index")
		}
		offsetString = stepState.ArrangedLessonIDs[offsetIndex-1]
	}

	req := &vpb.GetLessonsRequest{
		Paging: &cpb.Paging{
			Limit: uint32(convLimit),
			Offset: &cpb.Paging_OffsetString{
				OffsetString: offsetString,
			},
		},
		CurrentTime:       timestamppb.New(now),
		LessonTimeCompare: vpb.LessonTimeCompare_LESSON_TIME_COMPARE_FUTURE_AND_EQUAL,
		TimeLookup:        vpb.TimeLookup_TIME_LOOKUP_END_TIME,
		SortAsc:           true,
	}
	stepState.Response, stepState.ResponseErr = vpb.NewVirtualLessonReaderServiceClient(s.VirtualClassroomConn).
		GetLessons(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req
	stepState.OffsetIndex = offsetIndex

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsListOfLessonsUsingTimeCompareAndLookup(ctx context.Context, compare, lookup string) (context.Context, error) {
	time.Sleep(3 * time.Second)

	stepState := StepStateFromContext(ctx)
	now := time.Now().Add(-5 * 24 * time.Hour).UTC()

	vpbCompare, ok := vpb.LessonTimeCompare_value[compare]
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson time compare value is not valid: %s", compare)
	}

	vpbLookup, ok := vpb.TimeLookup_value[lookup]
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("time lookup value is not valid %s", lookup)
	}

	req := &vpb.GetLessonsRequest{
		Paging: &cpb.Paging{
			Limit: 15,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: "",
			},
		},
		CurrentTime:       timestamppb.New(now),
		LessonTimeCompare: vpb.LessonTimeCompare(vpbCompare),
		TimeLookup:        vpb.TimeLookup(vpbLookup),
		SortAsc:           true,
	}
	stepState.Response, stepState.ResponseErr = vpb.NewVirtualLessonReaderServiceClient(s.VirtualClassroomConn).
		GetLessons(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsListOfLessonsUsingFilter(ctx context.Context, filter, filterValue string) (context.Context, error) {
	time.Sleep(3 * time.Second)

	stepState := StepStateFromContext(ctx)
	now := time.Now().Add(-5 * 24 * time.Hour).UTC()

	req := &vpb.GetLessonsRequest{
		Paging: &cpb.Paging{
			Limit: 15,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: "",
			},
		},
		CurrentTime:       timestamppb.New(now),
		LessonTimeCompare: vpb.LessonTimeCompare_LESSON_TIME_COMPARE_FUTURE_AND_EQUAL,
		TimeLookup:        vpb.TimeLookup_TIME_LOOKUP_END_TIME,
		SortAsc:           true,
	}
	req.Filter = &vpb.GetLessonsFilter{}

	switch filter {
	case "LOCATION_IDS":
		req.LocationIds = stepState.LocationIDs
	case "TEACHER_IDS":
		req.Filter.TeacherIds = stepState.TeacherIDs
	case "STUDENT_IDS":
		req.Filter.StudentIds = stepState.StudentIds
	case "COURSE_IDS":
		req.Filter.CourseIds = stepState.CourseIDs
	case "LESSON_STATUS":
		cpbLessonStatus, ok := cpb.LessonSchedulingStatus_value[filterValue]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson scheduling status is not valid %s", filterValue)
		}
		req.Filter.SchedulingStatus = append(req.Filter.SchedulingStatus, cpb.LessonSchedulingStatus(cpbLessonStatus))
	case "LIVE_LESSON_STATUS":
		vpbLiveLessonStatus, ok := vpb.LiveLessonStatus_value[filterValue]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("live lesson status is not valid %s", filterValue)
		}
		req.Filter.LiveLessonStatus = vpb.LiveLessonStatus(vpbLiveLessonStatus)
	case "FROM_DATE":
		req.Filter.FromDate = timestamppb.New(time.Now().Add(-2 * 24 * time.Hour))
	case "TO_DATE":
		req.Filter.ToDate = timestamppb.New(time.Now().Add(2 * 24 * time.Hour))
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualLessonReaderServiceClient(s.VirtualClassroomConn).
		GetLessons(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req
	stepState.GetLessonsSelectedFilter = filter

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsAListOfLessonsWithTheCorrectPageInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request := stepState.Request.(*vpb.GetLessonsRequest)
	response := stepState.Response.(*vpb.GetLessonsResponse)

	lessonItems := response.GetItems()
	totalLessons := response.GetTotalLesson()
	totalItems := response.GetTotalItems()
	offsetIndex := stepState.OffsetIndex
	limit := int(request.GetPaging().GetLimit())

	if len(lessonItems) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting lessons but got 0")
	}

	if totalLessons != totalItems {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting total lessons %d be equal to total items %d", totalLessons, totalItems)
	}

	// items validate if are the expected lessons [X::] where X is the index in the previous step
	expectedEndIndex := offsetIndex + limit
	expectedLessonIDs := stepState.ArrangedLessonIDs[offsetIndex:expectedEndIndex]
	for i, lesson := range lessonItems {
		if lesson.Id != expectedLessonIDs[i] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson ID %s is not the same with the actual lesson ID %s", expectedLessonIDs[i], lesson.Id)
		}
	}

	// check previous page offset
	previousPageIndex := offsetIndex - limit
	expectedPrePageOffset := ""
	if previousPageIndex >= limit {
		if len(stepState.ArrangedLessonIDs) < previousPageIndex {
			return StepStateToContext(ctx, stepState), fmt.Errorf("attempted to get offset previous string but lesson IDs is less than the previous page offset index")
		}

		expectedPrePageOffset = stepState.ArrangedLessonIDs[previousPageIndex-1]
	}
	prevPageOffset := response.PreviousPage.GetOffsetString()
	if expectedPrePageOffset != prevPageOffset {
		return StepStateToContext(ctx, stepState), fmt.Errorf("prev page offset %s is not the same with the expected ID %s, expected offset index: %d, offset index used in request: %d, limit: %d", prevPageOffset, expectedPrePageOffset, previousPageIndex, offsetIndex, limit)
	}

	// check next page offset
	lastItemID := lessonItems[len(lessonItems)-1].GetId()
	nextPageOffset := response.NextPage.GetOffsetString()
	if nextPageOffset != lastItemID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("next page offset %s is not the same with the last item %s", nextPageOffset, lastItemID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsAListOfLessonsWithTheCorrectTime(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request := stepState.Request.(*vpb.GetLessonsRequest)
	response := stepState.Response.(*vpb.GetLessonsResponse)
	lessonItems := response.GetItems()

	if len(lessonItems) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting lessons but got 0")
	}

	compareCondition := func(currentTime, timeCompare time.Time) bool { return false }
	// conditions below use failing conditions on the lesson time compare
	// ex. Future is start_time > current_time
	// 	   failing conditions are start time before the current time
	//        and start time equal to current time
	switch request.GetLessonTimeCompare() {
	case vpb.LessonTimeCompare_LESSON_TIME_COMPARE_FUTURE:
		compareCondition = func(currentTime, timeCompare time.Time) bool {
			return timeCompare.Before(currentTime) && currentTime != timeCompare
		}
	case vpb.LessonTimeCompare_LESSON_TIME_COMPARE_FUTURE_AND_EQUAL:
		compareCondition = func(currentTime, timeCompare time.Time) bool {
			return timeCompare.Before(currentTime)
		}
	case vpb.LessonTimeCompare_LESSON_TIME_COMPARE_PAST:
		compareCondition = func(currentTime, timeCompare time.Time) bool {
			return timeCompare.After(currentTime) && currentTime != timeCompare
		}
	case vpb.LessonTimeCompare_LESSON_TIME_COMPARE_PAST_AND_EQUAL:
		compareCondition = func(currentTime, timeCompare time.Time) bool {
			return timeCompare.After(currentTime)
		}
	}

	useStartTime := true
	comapreEnded := func(endAt *timestamppb.Timestamp) bool { return false }
	// also use failing condition to check if lesson has ended or not
	switch request.GetTimeLookup() {
	case vpb.TimeLookup_TIME_LOOKUP_END_TIME:
		useStartTime = false
	case vpb.TimeLookup_TIME_LOOKUP_END_TIME_INCLUDE_WITHOUT_END_AT:
		comapreEnded = func(endAt *timestamppb.Timestamp) bool { return endAt != nil }
		useStartTime = false
	case vpb.TimeLookup_TIME_LOOKUP_END_TIME_INCLUDE_WITH_END_AT:
		comapreEnded = func(endAt *timestamppb.Timestamp) bool { return endAt == nil }
		useStartTime = false
	}

	currentTime := request.GetCurrentTime().AsTime()
	for _, lesson := range lessonItems {
		var lessonTime time.Time
		if useStartTime {
			lessonTime = lesson.GetStartTime().AsTime()
		} else {
			lessonTime = lesson.GetEndTime().AsTime()
		}

		if compareCondition(currentTime, lessonTime) {
			if comapreEnded(lesson.GetEndAt()) {
				return StepStateToContext(ctx, stepState), fmt.Errorf(`lesson %s did not pass time checking condition: 
					lesson time compare: %s
					time lookup: %s
					current time: %s
					compared time: %s
					end time: %s`,
					lesson.Id, request.GetLessonTimeCompare(), request.GetTimeLookup(), currentTime.String(), lessonTime.String(), lesson.GetEndAt().AsTime().String())
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsAListOfCorrectLessonsBasedFromTheFilter(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request := stepState.Request.(*vpb.GetLessonsRequest)
	response := stepState.Response.(*vpb.GetLessonsResponse)

	lessonItems := response.GetItems()
	if len(lessonItems) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting lessons but got 0")
	}

	switch stepState.GetLessonsSelectedFilter {
	case "LOCATION_IDS":
		if err := s.checkLocationIDs(request.GetLocationIds(), lessonItems); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "TEACHER_IDS":
		if err := s.checkLessonTeacherIDs(request.Filter.GetTeacherIds(), lessonItems); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "STUDENT_IDS":
		if err := s.checkLessonStudentIDs(ctx, request.Filter.GetStudentIds(), lessonItems); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "COURSE_IDS":
		if err := s.checkCourseIDs(request.Filter.GetCourseIds(), lessonItems); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "LESSON_STATUS":
		if err := s.checkLessonSchedulingStatus(request.Filter.GetSchedulingStatus(), lessonItems); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "LIVE_LESSON_STATUS":
		if err := s.checkLiveLessonStatus(request.Filter.GetLiveLessonStatus(), lessonItems); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "FROM_DATE":
		if err := s.checkLessonFromDateFilter(request.Filter.GetFromDate(), lessonItems); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	case "TO_DATE":
		if err := s.checkLessonToDateFilter(request.Filter.GetToDate(), lessonItems); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("selected filter %s not supported", stepState.GetLessonsSelectedFilter)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkLocationIDs(locationIDs []string, lessonItems []*vpb.GetLessonsResponse_Lesson) error {
	if len(locationIDs) == 0 {
		return fmt.Errorf("expecting filter location IDs but got empty")
	}

	for _, lesson := range lessonItems {
		if !sliceutils.Contains(locationIDs, lesson.CenterId) {
			return fmt.Errorf("lesson %s with location %s is not in location filter %s", lesson.Id, lesson.CenterId, locationIDs)
		}
	}
	return nil
}

func (s *suite) checkLessonTeacherIDs(teacherIDs []string, lessonItems []*vpb.GetLessonsResponse_Lesson) error {
	if len(teacherIDs) == 0 {
		return fmt.Errorf("expecting filter teacher IDs but got empty")
	}

	for _, lesson := range lessonItems {
		isValid := false
		lessonTeacherIDs := lesson.GetTeacherIds()
		for _, teacherID := range lessonTeacherIDs {
			if slices.Contains(teacherIDs, teacherID) {
				isValid = true
			}
		}
		if !isValid {
			return fmt.Errorf("lesson %s found invalid teacher IDs %s not in the teacher filter %s", lesson.Id, lessonTeacherIDs, teacherIDs)
		}
	}

	return nil
}

func (s *suite) checkLessonStudentIDs(ctx context.Context, studentIDs []string, lessonItems []*vpb.GetLessonsResponse_Lesson) error {
	if len(studentIDs) == 0 {
		return fmt.Errorf("expecting filter student IDs but got empty")
	}

	lessonIDs := make([]string, len(lessonItems))
	for _, lesson := range lessonItems {
		lessonIDs = append(lessonIDs, lesson.Id)
	}

	lessonMembersMap, err := (&vc_repo.LessonMemberRepo{}).GetLessonLearnersByLessonIDs(ctx, s.LessonmgmtDBTrace, lessonIDs)
	if err != nil {
		return fmt.Errorf("error when fetching learners for lessons %s: %w", lessonIDs, err)
	}

	for _, lesson := range lessonItems {
		learners, ok := lessonMembersMap[lesson.Id]
		if !ok {
			return fmt.Errorf("expected learners in lesson %s but found nothing", lesson.Id)
		}
		lessonStudentIDs := learners.GetLearnerIDs()

		isValid := false
		for _, studentID := range lessonStudentIDs {
			if slices.Contains(studentIDs, studentID) {
				isValid = true
			}
		}
		if !isValid {
			return fmt.Errorf("lesson %s found invalid student IDs %s not in the teacher filter %s", lesson.Id, lessonStudentIDs, studentIDs)
		}
	}

	return nil
}

func (s *suite) checkCourseIDs(courseIDs []string, lessonItems []*vpb.GetLessonsResponse_Lesson) error {
	if len(courseIDs) == 0 {
		return fmt.Errorf("expecting filter course IDs but got empty")
	}

	for _, lesson := range lessonItems {
		if !sliceutils.Contains(courseIDs, lesson.CourseId) {
			return fmt.Errorf("lesson %s with course %s is not in course filter %s", lesson.Id, lesson.CourseId, courseIDs)
		}
	}
	return nil
}

func (s *suite) checkLessonSchedulingStatus(statuses []cpb.LessonSchedulingStatus, lessonItems []*vpb.GetLessonsResponse_Lesson) error {
	if len(statuses) == 0 {
		return fmt.Errorf("expecting lesson status filter but got empty")
	}
	lessonStatus := statuses[0]

	for _, lesson := range lessonItems {
		if lesson.SchedulingStatus != lessonStatus {
			return fmt.Errorf("lesson %s status %s does not match with lesson status filter %s", lesson.Id, lesson.SchedulingStatus.String(), lessonStatus.String())
		}
	}
	return nil
}

func (s *suite) checkLiveLessonStatus(liveLessonStatus vpb.LiveLessonStatus, lessonItems []*vpb.GetLessonsResponse_Lesson) error {
	switch liveLessonStatus {
	case vpb.LiveLessonStatus_LIVE_LESSON_STATUS_ENDED:
		for _, lesson := range lessonItems {
			if endAt := lesson.GetEndAt(); endAt == nil {
				return fmt.Errorf("lesson %s expected to be ended but got no end time", lesson.Id)
			}
		}
	case vpb.LiveLessonStatus_LIVE_LESSON_STATUS_NOT_ENDED:
		for _, lesson := range lessonItems {
			if endAt := lesson.GetEndAt(); endAt != nil {
				return fmt.Errorf("lesson %s expected to be not ended but got end time", lesson.Id)
			}
		}
	default:
		return fmt.Errorf("expecting live lesson status filter but got none")
	}
	return nil
}

func (s *suite) checkLessonFromDateFilter(fromDateFilter *timestamppb.Timestamp, lessonItems []*vpb.GetLessonsResponse_Lesson) error {
	if fromDateFilter == nil {
		return fmt.Errorf("expecting from date filter but got none")
	}

	for _, lesson := range lessonItems {
		lessonEndTime := lesson.GetEndTime()
		if fromDateFilter.AsTime().After(lessonEndTime.AsTime()) {
			return fmt.Errorf("lesson %s expected to be before the from date filter but is after", lesson.Id)
		}
	}
	return nil
}

func (s *suite) checkLessonToDateFilter(toDateFilter *timestamppb.Timestamp, lessonItems []*vpb.GetLessonsResponse_Lesson) error {
	if toDateFilter == nil {
		return fmt.Errorf("expecting to date filter but got none")
	}

	for _, lesson := range lessonItems {
		lessonStartTime := lesson.GetStartTime()
		if toDateFilter.AsTime().Before(lessonStartTime.AsTime()) {
			return fmt.Errorf("lesson %s expected to be after the to date filter but is before", lesson.Id)
		}
	}
	return nil
}
