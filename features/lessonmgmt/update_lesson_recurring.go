package lessonmgmt

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

/*
updateLessonRequestDefault: update lesson request default
We had have lesson chain: A -> B -> C -> D
Default: change C
This request hasn't changed from what was created before
*/
func (s *Suite) updateLessonRequestDefault(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	createLessonReq, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected *lpb.CreateLessonRequest, but got %T type", stepState.Request.(*lpb.CreateLessonRequest))
	}
	lessonChain, err := s.retrieveLessonChainByLessonID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(lessonChain) < 2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of lesson in chain is not correct")
	}
	lessonC := lessonChain[2]
	stepState.OldEndDate = lessonChain[1].EndTime
	updateLessonReq := &bpb.UpdateLessonRequest{
		LessonId:        lessonC.LessonID,
		StartTime:       timestamppb.New(lessonC.StartTime),
		EndTime:         timestamppb.New(lessonC.EndTime),
		TeachingMedium:  createLessonReq.TeachingMedium,
		TeachingMethod:  createLessonReq.TeachingMethod,
		TeacherIds:      createLessonReq.TeacherIds,
		CenterId:        createLessonReq.LocationId,
		StudentInfoList: []*bpb.UpdateLessonRequest_StudentInfo{},
		Materials:       []*bpb.Material{},
		SavingOption: &bpb.UpdateLessonRequest_SavingOption{
			Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
			Recurrence: &bpb.Recurrence{
				EndDate: createLessonReq.SavingOption.Recurrence.EndDate,
			},
		},
	}
	for _, v := range createLessonReq.StudentInfoList {
		updateLessonReq.StudentInfoList = append(updateLessonReq.StudentInfoList, &bpb.UpdateLessonRequest_StudentInfo{
			StudentId:        v.StudentId,
			CourseId:         v.CourseId,
			LocationId:       v.LocationId,
			AttendanceStatus: bpb.StudentAttendStatus(v.AttendanceStatus),
		})
	}
	materials := make([]*bpb.Material, 0, len(createLessonReq.Materials))
	for _, v := range createLessonReq.Materials {
		switch resource := v.Resource.(type) {
		case *lpb.Material_BrightcoveVideo_:
			material := &bpb.Material{
				Resource: &bpb.Material_BrightcoveVideo_{
					BrightcoveVideo: &bpb.Material_BrightcoveVideo{
						Name: resource.BrightcoveVideo.Name,
						Url:  resource.BrightcoveVideo.Url,
					}}}
			materials = append(materials, material)
		case *lpb.Material_MediaId:
			material := &bpb.Material{
				Resource: &bpb.Material_MediaId{
					MediaId: resource.MediaId,
				}}
			materials = append(materials, material)
		default:
			return nil, status.Error(codes.Internal, fmt.Errorf(`unexpected material's type %T`, resource).Error())
		}
	}
	updateLessonReq.Materials = materials
	stepState.Request = updateLessonReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userChangedLocationTo(ctx context.Context, _location, _endDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	endDate, _ := time.Parse(time.RFC3339, _endDate)
	indexLocation, _ := strconv.Atoi(_location)
	locationID := stepState.CenterIDs[indexLocation]

	ctx, err := s.updateLessonRequestDefault(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	updateLessonReq, ok := stepState.Request.(*bpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	updateLessonReq.CenterId = locationID
	updateLessonReq.SavingOption = &bpb.UpdateLessonRequest_SavingOption{
		Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
		Recurrence: &bpb.Recurrence{
			EndDate: timestamppb.New(endDate),
		},
	}
	for i := range updateLessonReq.StudentInfoList {
		updateLessonReq.StudentInfoList[i].LocationId = locationID
	}
	stepState.Request = updateLessonReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userChangedStudentInfoTo(ctx context.Context, startAt, endAt string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.updateLessonRequestDefault(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	updateLessonReq, ok := stepState.Request.(*bpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong type %T", updateLessonReq)
	}
	loc := LoadLocalLocation()
	updateLessonReq.StartTime = timestamppb.New(time.Date(2022, 7, 23, 9, 0, 0, 0, loc))
	updateLessonReq.EndTime = timestamppb.New(time.Date(2022, 7, 23, 10, 0, 0, 0, loc))
	// create new student course
	ctx, err = s.CommonSuite.CreateStudentAccounts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	ctx, err = s.CommonSuite.SomeStudentSubscriptionsWithParams(ctx, startAt, endAt)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	updateLessonReq.StudentInfoList = []*bpb.UpdateLessonRequest_StudentInfo{}
	studentIDWithCourseID := stepState.StudentIDWithCourseID
	addedStudentIDs := make(map[string]bool)
	for i := 0; i < len(studentIDWithCourseID); i += 2 {
		studentID := studentIDWithCourseID[i]
		courseID := studentIDWithCourseID[i+1]
		if _, ok := addedStudentIDs[studentID]; ok {
			continue
		}
		addedStudentIDs[studentID] = true
		updateLessonReq.StudentInfoList = append(updateLessonReq.StudentInfoList, &bpb.UpdateLessonRequest_StudentInfo{
			StudentId:        studentID,
			CourseId:         courseID,
			AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
			LocationId:       updateLessonReq.CenterId,
		})
	}
	stepState.Request = updateLessonReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userChangedLessonInfoTo(ctx context.Context, _endDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	endDate, _ := time.Parse(time.RFC3339, _endDate)
	ctx, err := s.updateLessonRequestDefault(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	updateLessonReq, ok := stepState.Request.(*bpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong type %T", updateLessonReq)
	}
	loc := LoadLocalLocation()
	updateLessonReq.StartTime = timestamppb.New(time.Date(2022, 7, 23, 9, 0, 0, 0, loc))
	updateLessonReq.EndTime = timestamppb.New(time.Date(2022, 7, 23, 10, 0, 0, 0, loc))
	updateLessonReq.TeachingMedium = cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE
	// create new teachers
	ctx, err = s.CreateTeacherAccounts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when create teacher account %w", err)
	}
	updatedTeachers := stepState.TeacherIDs
	updateLessonReq.TeacherIds = updatedTeachers
	// create new student course
	ctx, err = s.CommonSuite.CreateStudentAccounts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when create student account %w", err)
	}
	ctx, err = s.CommonSuite.SomeStudentSubscriptions(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when create student subscription %w", err)
	}
	updateLessonReq.StudentInfoList = []*bpb.UpdateLessonRequest_StudentInfo{}
	studentIDWithCourseID := stepState.StudentIDWithCourseID
	addedStudentIDs := make(map[string]bool)
	for i := 0; i < len(studentIDWithCourseID); i += 2 {
		studentID := studentIDWithCourseID[i]
		courseID := studentIDWithCourseID[i+1]
		if _, ok := addedStudentIDs[studentID]; ok {
			continue
		}
		addedStudentIDs[studentID] = true
		updateLessonReq.StudentInfoList = append(updateLessonReq.StudentInfoList, &bpb.UpdateLessonRequest_StudentInfo{
			StudentId:        studentID,
			CourseId:         courseID,
			AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
			LocationId:       updateLessonReq.CenterId,
		})
	}
	updateLessonReq.SavingOption = &bpb.UpdateLessonRequest_SavingOption{
		Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
		Recurrence: &bpb.Recurrence{
			EndDate: timestamppb.New(endDate),
		},
	}
	stepState.Request = updateLessonReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) selectedAndFollowingLessonUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	updateLessonReq, ok := stepState.Request.(*bpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong type %T", updateLessonReq)
	}
	lessonChain, err := s.retrieveLessonChainByLessonID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to retrieve lesson chain :%w", err)
	}
	ctx, err = s.checkThisAndFollowingLesson(ctx, lessonChain)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) selectedLessonUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	updatedReq := stepState.Request.(*bpb.UpdateLessonRequest)
	lesson, err := (&repo.LessonRepo{}).GetLessonByID(ctx, s.CommonSuite.BobDB, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if lesson.LocationID != updatedReq.CenterId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected CenterId %s but got %s", updatedReq.CenterId, lesson.LocationID)
	}
	if !lesson.StartTime.Equal(updatedReq.StartTime.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected startTime %s but got %s", updatedReq.StartTime, lesson.StartTime)
	}
	if !lesson.EndTime.Equal(updatedReq.EndTime.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected endTime %s but got %s", updatedReq.EndTime, lesson.EndTime)
	}
	stepState.Lesson = lesson
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) selectedLessonLeaveChain(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Lesson.SchedulerID == stepState.OldSchedulerID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson should be leave chain")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) selectedLessonKeepChain(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Lesson.SchedulerID != stepState.OldSchedulerID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson should be keep chain")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) haveUpdatedEndDate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	query := "select end_date from scheduler where scheduler_id = $1"
	row := s.BobDB.QueryRow(ctx, query, stepState.OldSchedulerID)
	var endDate pgtype.Timestamptz
	if err := row.Scan(&endDate); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	expectedDate := stepState.OldEndDate
	if !endDate.Time.Equal(expectedDate) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected end_date %s but got %s", expectedDate, endDate.Time)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkThisAndFollowingLesson(ctx context.Context, lessonChain []*domain.Lesson) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*bpb.UpdateLessonRequest)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", stepState.ResponseErr)
	}
	lessons, err := s.retrieveLessonChainByLessonID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	stepState.LessonDomains = lessons
	thisAndFollowingLesson := []*domain.Lesson{}
	for _, ls := range lessons {
		if (ls.StartTime.After(req.StartTime.AsTime()) ||
			ls.StartTime.Equal(req.StartTime.AsTime())) && !ls.IsLocked {
			thisAndFollowingLesson = append(thisAndFollowingLesson, ls)
		}
	}
	startTime := golibs.TimestamppbToTime(req.StartTime)
	endTime := golibs.TimestamppbToTime(req.EndTime)
	for _, lesson := range thisAndFollowingLesson {
		if err := s.equalTime(ctx, lesson, startTime, endTime); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
		if lesson.LocationID != req.CenterId {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected CenterId %s but got %s", req.CenterId, lesson.LocationID)
		}
		if req.TeachingMedium.String() != string(lesson.TeachingMedium) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected TeachingMedium %s but got %s", req.TeachingMedium.String(), lesson.TeachingMedium)
		}
		actualTeacherIDs := lesson.GetTeacherIDs()
		if !stringutil.SliceElementsMatch(actualTeacherIDs, req.TeacherIds) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for teacher IDs, got %s", req.TeacherIds, actualTeacherIDs)
		}
		if lesson.StartTime.Format(domain.Ymd) <= stepState.EndDate.Format(domain.Ymd) {
			learnerIds := make([]string, 0, len(req.StudentInfoList))
			for _, studentInfo := range req.StudentInfoList {
				learnerIds = append(learnerIds, studentInfo.StudentId)
			}
			actualLearnerIDs := lesson.GetLearnersIDs()
			if !stringutil.SliceElementsMatch(actualLearnerIDs, learnerIds) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for learner IDs, got %s", learnerIds, actualLearnerIDs)
			}
		} else if len(lesson.Learners) > 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson `%s` have student info not correct", lesson.LessonID)
		}

		startTime = startTime.AddDate(0, 0, 7)
		endTime = endTime.AddDate(0, 0, 7)
	}
	return StepStateToContext(ctx, stepState), nil
}

// lessons is locked are not updated and deleted
func (s *Suite) checkLockedLessonInThisAndFollowingLesson(ctx context.Context, _startTime, _endTime, location, generalInfor string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*bpb.UpdateLessonRequest)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", stepState.ResponseErr)
	}
	lessons := stepState.LessonDomains
	thisAndFollowingLesson := []*domain.Lesson{}
	for _, ls := range lessons {
		if (ls.StartTime.After(req.StartTime.AsTime()) ||
			ls.StartTime.Equal(req.StartTime.AsTime())) && ls.IsLocked {
			if ls.DeletedAt != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("the lesson is locked with lessonId %s should not deleted", ls.LessonID)
			}
			thisAndFollowingLesson = append(thisAndFollowingLesson, ls)
		}
	}

	startTime := golibs.TimestamppbToTime(req.StartTime)
	endTime := golibs.TimestamppbToTime(req.EndTime)
	for _, lesson := range thisAndFollowingLesson {
		if _startTime != "none_start_time" {
			if err := s.equalTime(ctx, lesson, startTime, endTime); err == nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
			}
		}

		if location != "none_location" {
			if lesson.LocationID != req.CenterId {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected CenterId %s but got %s", req.CenterId, lesson.LocationID)
			}
		}
		if generalInfor != "none_general_infor" {
			if req.TeachingMedium.String() != string(lesson.TeachingMedium) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected TeachingMedium %s but got %s", req.TeachingMedium.String(), lesson.TeachingMedium)
			}
			actualTeacherIDs := lesson.GetTeacherIDs()
			if !stringutil.SliceElementsMatch(actualTeacherIDs, req.TeacherIds) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for teacher IDs, got %s", req.TeacherIds, actualTeacherIDs)
			}

			learnerIds := make([]string, 0, len(req.StudentInfoList))
			for _, studentInfo := range req.StudentInfoList {
				learnerIds = append(learnerIds, studentInfo.StudentId)
			}
			actualLearnerIDs := lesson.GetLearnersIDs()
			if !stringutil.SliceElementsMatch(actualLearnerIDs, learnerIds) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for learner IDs, got %s", learnerIds, actualLearnerIDs)
			}
		}

		startTime = startTime.AddDate(0, 0, 7)
		endTime = endTime.AddDate(0, 0, 7)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) lockLessonAt(ctx context.Context, _lockAt string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if _lockAt != "" {
		lockAt, _ := time.Parse(time.RFC3339, _lockAt)
		lessons, err := s.retrieveLessonChainByLessonID(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
		for _, ls := range lessons {
			if ls.StartTime.Equal(lockAt) {
				stepState.CurrentLessonID = ls.LessonID
				ctx, err := s.lockLesson(ctx, "true")
				if err != nil {
					return StepStateToContext(ctx, stepState), err
				}
				break
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userChangedLessonTimeTo(ctx context.Context, _startTime, _endTime, _endDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startTime, _ := time.Parse(time.RFC3339, _startTime)
	endTime, _ := time.Parse(time.RFC3339, _endTime)
	endDate, _ := time.Parse(time.RFC3339, _endDate)
	ctx, err := s.updateLessonRequestDefault(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	updateLessonReq, ok := stepState.Request.(*bpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	updateLessonReq.StartTime = timestamppb.New(startTime)
	updateLessonReq.EndTime = timestamppb.New(endTime)
	updateLessonReq.SavingOption = &bpb.UpdateLessonRequest_SavingOption{
		Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
		Recurrence: &bpb.Recurrence{
			EndDate: timestamppb.New(endDate),
		},
	}
	stepState.Request = updateLessonReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) updateLessonBySaving(ctx context.Context, savingOption string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	updatedReq := stepState.Request.(*bpb.UpdateLessonRequest)
	if savingOption == "one-time" {
		updatedReq.SavingOption = &bpb.UpdateLessonRequest_SavingOption{
			Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
		}
	} else if savingOption == "weekly recurrence" {
		updatedReq.SavingOption = &bpb.UpdateLessonRequest_SavingOption{
			Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
			Recurrence: &bpb.Recurrence{
				EndDate: timestamppb.New(updatedReq.EndTime.AsTime().AddDate(0, 1, 0)),
			},
		}
	}
	stepState.Response, stepState.ResponseErr = bpb.NewLessonManagementServiceClient(s.BobConn).UpdateLesson(s.CommonSuite.SignedCtx(ctx), updatedReq)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) updateLessonBySavingWeekly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	updatedReq := stepState.Request.(*bpb.UpdateLessonRequest)
	ctx, err := s.createDeletedLessonSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	stepState.Response, stepState.ResponseErr = bpb.NewLessonManagementServiceClient(s.BobConn).UpdateLesson(s.CommonSuite.SignedCtx(ctx), updatedReq)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	stepState.CurrentLessonID = updatedReq.LessonId
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) saveLessonByStatus(ctx context.Context, lessonStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	updatedReq := stepState.Request.(*bpb.UpdateLessonRequest)
	updatedReq.SchedulingStatus = bpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED
	if lessonStatus == "draft" {
		updatedReq.SchedulingStatus = bpb.LessonStatus_LESSON_SCHEDULING_STATUS_DRAFT
	} else {
		if len(updatedReq.TeacherIds) == 0 {
			updatedReq.TeacherIds = stepState.TeacherIDs
		}
		if len(updatedReq.StudentInfoList) == 0 {
			addedStudentIDs := make(map[string]bool)
			for i := 0; i < len(stepState.StudentIDWithCourseID); i += 2 {
				studentID := stepState.StudentIDWithCourseID[i]
				courseID := stepState.StudentIDWithCourseID[i+1]
				if _, ok := addedStudentIDs[studentID]; ok {
					continue
				}
				addedStudentIDs[studentID] = true
				updatedReq.StudentInfoList = append(updatedReq.StudentInfoList, &bpb.UpdateLessonRequest_StudentInfo{
					StudentId:        studentID,
					CourseId:         courseID,
					AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
				})
			}
		}
	}
	stepState.Response, stepState.ResponseErr = bpb.NewLessonManagementServiceClient(s.BobConn).UpdateLesson(s.CommonSuite.SignedCtx(ctx), updatedReq)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	stepState.CurrentLessonID = updatedReq.LessonId
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) updateLessonBySavingOnlyThis(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	updatedReq := stepState.Request.(*bpb.UpdateLessonRequest)
	updatedReq.SavingOption = &bpb.UpdateLessonRequest_SavingOption{
		Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
	}
	stepState.Response, stepState.ResponseErr = bpb.NewLessonManagementServiceClient(s.BobConn).UpdateLesson(s.CommonSuite.SignedCtx(ctx), updatedReq)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	stepState.CurrentLessonID = updatedReq.LessonId
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userChangedLessonTimeAndLocation(ctx context.Context, _startTime, _endTime, _endDate, _location string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startTime, _ := time.Parse(time.RFC3339, _startTime)
	endTime, _ := time.Parse(time.RFC3339, _endTime)
	endDate, _ := time.Parse(time.RFC3339, _endDate)
	ctx, err := s.updateLessonRequestDefault(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}

	updateLessonReq, ok := stepState.Request.(*bpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong type %T", updateLessonReq)
	}

	indexLocation, _ := strconv.Atoi(_location)
	locationID := stepState.CenterIDs[indexLocation]
	updateLessonReq.StartTime = timestamppb.New(startTime)
	updateLessonReq.EndTime = timestamppb.New(endTime)
	updateLessonReq.CenterId = locationID
	updateLessonReq.TeachingMedium = cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE
	// create new teachers
	ctx, err = s.CommonSuite.CreateTeacherAccounts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	updatedTeachers := stepState.TeacherIDs
	updateLessonReq.TeacherIds = updatedTeachers
	// create new student course
	ctx, err = s.CommonSuite.CreateStudentAccounts(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}

	ctx, err = s.CommonSuite.SomeStudentSubscriptions(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	updateLessonReq.StudentInfoList = []*bpb.UpdateLessonRequest_StudentInfo{}
	addedStudentIDs := make(map[string]bool)
	for i := 0; i < len(stepState.StudentIDWithCourseID); i += 2 {
		studentID := stepState.StudentIDWithCourseID[i]
		courseID := stepState.StudentIDWithCourseID[i+1]
		if _, ok := addedStudentIDs[studentID]; ok {
			continue
		}
		addedStudentIDs[studentID] = true
		updateLessonReq.StudentInfoList = append(updateLessonReq.StudentInfoList, &bpb.UpdateLessonRequest_StudentInfo{
			StudentId:        studentID,
			CourseId:         courseID,
			AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
			LocationId:       locationID,
		})
	}

	updateLessonReq.SavingOption = &bpb.UpdateLessonRequest_SavingOption{
		Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
		Recurrence: &bpb.Recurrence{
			EndDate: timestamppb.New(endDate),
		},
	}

	stepState.Request = updateLessonReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnSomeLessonDates(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	schedulerID, err := s.getSchedulerIDByLessonID(ctx, stepState.CurrentLessonID)
	loc := LoadLocalLocation()
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.OldSchedulerID = schedulerID

	lessons, err := s.retrieveLessonChainByLessonID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}

	lessonIDs := make([]string, 0, len(lessons))
	lessonDates := make(map[string]string)
	for _, v := range lessons {
		lessonIDs = append(lessonIDs, v.LessonID)
		lessonDate := v.StartTime.In(loc).Format(timeLayout)
		lessonDates[lessonDate] = v.LessonID
	}
	stepState.LessonIDs = lessonIDs
	stepState.LessonDates = lessonDates
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) AnExistingRecurringLessonWithClassroom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return s.CreateRecurringLessonWithClassrooms(StepStateToContext(ctx, stepState), "existing")
}

func (s *Suite) UserChangedLessonWithClassroom(ctx context.Context, recordState string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.updateLessonRequestDefault(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}

	updateLessonReq, ok := stepState.Request.(*bpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong type %T", updateLessonReq)
	}

	if recordState == "existing" {
		updateLessonReq.ClassroomIds = append(updateLessonReq.ClassroomIds, stepState.ClassroomIDs[0], stepState.ClassroomIDs[1])
	} else {
		updateLessonReq.ClassroomIds = append(updateLessonReq.ClassroomIds, stepState.ClassroomIDs[0], idutil.ULIDNow())
	}

	stepState.Request = updateLessonReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) TheSelectedLessonClassroomIsUpdated(ctx context.Context, updateState string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonRepo := repo.LessonRepo{}
	lesson, err := lessonRepo.GetLessonByID(ctx, s.CommonSuite.BobDB, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}
	actualClassroomIDs := lesson.Classrooms.GetIDs()

	updatedRequest, ok := stepState.Request.(*bpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLessonRequest, got %T", updatedRequest)
	}

	if updateState == "updated" && !stringutil.SliceElementsMatch(actualClassroomIDs, updatedRequest.ClassroomIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for updated lesson classroom IDs, got %s", updatedRequest.ClassroomIds, actualClassroomIDs)
	} else if updateState == "not updated" && stringutil.SliceElementsMatch(actualClassroomIDs, updatedRequest.ClassroomIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson classroom IDs are not updated, but got %s", actualClassroomIDs)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) TheOtherLessonsClassroomAre(ctx context.Context, updateState string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	currentLessonID := stepState.CurrentLessonID

	updatedRequest, ok := stepState.Request.(*bpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", updatedRequest)
	}

	schedulerID, err := s.getSchedulerIDByLessonID(ctx, currentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.OldSchedulerID = schedulerID

	lessons, err := s.retrieveLessonChainByLessonID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}

	for i, lesson := range lessons {
		// added condition i <= 2 because of the scenario in s.updateLessonRequestDefault(ctx)
		// we only modify the 3rd lesson and beyond
		if i <= 2 {
			continue
		}

		actualClassroomIDs := lesson.Classrooms.GetIDs()
		if updateState == "updated" && !stringutil.SliceElementsMatch(actualClassroomIDs, updatedRequest.ClassroomIds) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for updated lesson ID %s classroom IDs, got %s", updatedRequest.ClassroomIds, lesson.LessonID, actualClassroomIDs)
		} else if updateState == "not updated" && stringutil.SliceElementsMatch(actualClassroomIDs, updatedRequest.ClassroomIds) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson ID %s classroom IDs are not updated, but got %s", lesson.LessonID, actualClassroomIDs)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
