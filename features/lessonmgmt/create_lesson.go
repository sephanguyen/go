package lessonmgmt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
)

func (s *Suite) aClassWithIDPrefixAndACourseWithIDPrefix(ctx context.Context, classIDPrefix string, courseIDPrefix string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	idNow := idutil.ULIDNow()
	classID := classIDPrefix + idNow
	courseID := courseIDPrefix + idNow
	className := "bdd-test-class-name"
	courseName := "bdd-test-course-name"
	locationID := constants.ManabieOrgLocation
	schoolID := golibs.ResourcePathFromCtx(ctx)
	classFields := []string{"class_id", "name", "course_id", "location_id", "school_id", "created_at", "updated_at"}
	courseFields := []string{"course_id", "name", "school_id", "created_at", "updated_at"}
	insertClassQuery := fmt.Sprintf("INSERT INTO class (%s) VALUES ($1,$2,$3,$4,$5,$6,$7)",
		strings.Join(classFields, ","))
	insertCourseQuery := fmt.Sprintf("INSERT INTO courses (%s) VALUES ($1,$2,$3,$4,$5)",
		strings.Join(courseFields, ","))
	now := time.Now()
	stepState.RequestSentAt = now
	_, err := s.BobDBTrace.Exec(ctx, insertCourseQuery, courseID, courseName, schoolID, now, now)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err insert course %w", err)
	}
	stepState.CurrentCourseID = courseID
	_, err = s.BobDBTrace.Exec(ctx, insertClassQuery, classID, className, courseID, locationID, schoolID, now, now)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err insert class %w", err)
	}
	stepState.CurrentClassId = classID
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserCreateALessonWithAllRequiredFields(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentTeachingMethod = "individual"
	return s.CommonSuite.UserCreateALessonWithMissingFields(ctx)
}
func (s *Suite) UserCreateALessonWithTeachingMethodAndAllRequiredFields(ctx context.Context, teachingMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentTeachingMethod = teachingMethod
	return s.CommonSuite.UserCreateALessonWithMissingFields(ctx)
}

func (s *Suite) UserCreateALiveLessonWithAllRequiredFields(ctx context.Context) (context.Context, error) {
	return s.CommonSuite.UserCreateALiveLessonWithMissingFields(ctx)
}

func (s *Suite) UserCreateALessonWithAllRequiredFieldsWithSub(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.ValidateLessonCreatedSubscription(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.CreateLessonCreatedSubscription: %w", err)
	}

	return s.UserCreateALessonWithAllRequiredFields(ctx)
}

func (s *Suite) TheLessonWasCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonRepo := repo.LessonRepo{}
	lesson, err := lessonRepo.GetLessonByID(ctx, s.CommonSuite.BobDB, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}

	createdRequest, ok := stepState.Request.(*bpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", createdRequest)
	}
	if ctx, err = s.ValidateLessonForCreatedRequestMGMT(ctx, lesson, createdRequest); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation failed for create Lesson: %s", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) StudentTeacherNameMustBeCorrect(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// student
	studentNames := s.CommonSuite.StepState.StudentNames
	lessonMemberDTO := &repo.LessonMember{}
	studentFields, _ := lessonMemberDTO.FieldMap()
	studentQuery := fmt.Sprintf("SELECT %s FROM lesson_members WHERE lesson_id = $1 AND user_last_name = ANY($2) ", strings.Join(studentFields, ", "))
	studentRows, err := s.BobDB.Query(ctx, studentQuery, &s.CommonSuite.StepState.CurrentLessonID, &studentNames)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.DB.QueryRow: %w", err)
	}

	for studentRows.Next() {
		studentRow := &repo.LessonMember{}
		_, studentValues := studentRow.FieldMap()
		if err := studentRows.Scan(studentValues...); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if valid := golibs.InArrayString(studentRow.UserLastName.String, s.CommonSuite.StepState.StudentNames); !valid {
			return StepStateToContext(ctx, stepState), fmt.Errorf("student name not found")
		}
	}
	// teacher
	teacherNames := s.CommonSuite.StepState.TeacherNames
	lessonTeacherDTO := &repo.LessonTeacher{}
	teacherFields, _ := lessonTeacherDTO.FieldMap()
	teacherQuery := fmt.Sprintf("SELECT %s FROM lessons_teachers WHERE lesson_id = $1 AND teacher_name = ANY($2) ", strings.Join(teacherFields, ", "))
	teacherRows, err := s.BobDB.Query(ctx, teacherQuery, &s.CommonSuite.StepState.CurrentLessonID, &teacherNames)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.DB.QueryRow: %w", err)
	}
	for teacherRows.Next() {
		teacherRowDTO := &repo.LessonTeacher{}
		_, teacherValues := teacherRowDTO.FieldMap()
		if err := teacherRows.Scan(teacherValues...); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if valid := golibs.InArrayString(teacherRowDTO.TeacherName.String, s.CommonSuite.StepState.TeacherNames); !valid {
			return StepStateToContext(ctx, stepState), fmt.Errorf("teacher name not found")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) TheLessonWasUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonRepo := repo.LessonRepo{}
	lesson, err := lessonRepo.GetLessonByID(ctx, s.CommonSuite.BobDB, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}
	updatedRequest, ok := stepState.Request.(*bpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", updatedRequest)
	}

	if ctx, err = s.validateLessonForUpdateRequestMGMT(ctx, lesson, updatedRequest); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation failed for update lesson: %s", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ValidateLessonForCreatedRequestMGMT(ctx context.Context, e *domain.Lesson, req *bpb.CreateLessonRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if !e.StartTime.Equal(req.StartTime.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for start time, got %s", req.StartTime.AsTime(), e.StartTime)
	}
	if !e.EndTime.Equal(req.EndTime.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for end time, got %s", req.EndTime.AsTime(), e.EndTime)
	}
	if req.Materials != nil && len(req.GetMaterials()) > 0 {
		actualMediaIDs := make(map[string]bool)
		if e.Material != nil {
			for _, mediaID := range e.Material.MediaIDs {
				actualMediaIDs[mediaID] = true
			}
		}
		for _, expectedMediaID := range stepState.MediaIDs {
			if _, ok := actualMediaIDs[expectedMediaID]; !ok {
				return StepStateToContext(ctx, stepState), fmt.Errorf("could not find media ID %s", expectedMediaID)
			}
		}
	}

	if e.LocationID != req.CenterId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected CenterId %s but got %s", req.CenterId, e.LocationID)
	}
	if req.TeachingMedium.String() != string(e.TeachingMedium) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected TeachingMedium %s but got %s", req.TeachingMedium.String(), e.TeachingMedium)
	}
	if req.TeachingMethod.String() != string(e.TeachingMethod) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected TeachingMethod %s but got %s", req.TeachingMethod.String(), e.TeachingMethod)
	}
	if e.SchedulingStatus != domain.LessonSchedulingStatus(req.SchedulingStatus.String()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected SchedulingStatus %s but got %s", domain.LessonSchedulingStatus(req.SchedulingStatus.String()), e.SchedulingStatus)
	}

	actualTeacherIDs := e.GetTeacherIDs()
	if !stringutil.SliceElementsMatch(actualTeacherIDs, req.TeacherIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for teacher IDs, got %s", req.TeacherIds, actualTeacherIDs)
	}

	learnerIds := make([]string, 0, len(req.StudentInfoList))
	for _, studentInfo := range req.StudentInfoList {
		learnerIds = append(learnerIds, studentInfo.StudentId)
	}
	actualLearnerIDs := e.GetLearnersIDs()
	if !stringutil.SliceElementsMatch(actualLearnerIDs, learnerIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for learner IDs, got %s", learnerIds, actualLearnerIDs)
	}
	// TODO: check course and location of lesson members
	// Validate lesson group
	if req.TeachingMethod == cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP {
		if req.CourseId != string(e.CourseID) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected CourseID %s but got %s", req.CourseId, e.CourseID)
		}
		if req.ClassId != string(e.ClassID) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected ClassID %s but got %s", req.ClassId, e.ClassID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) validateLessonForUpdateRequestMGMT(ctx context.Context, e *domain.Lesson, req *bpb.UpdateLessonRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if e.LessonID != req.LessonId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.LessonID mismatched")
	}

	studentInfoList := make([]*bpb.CreateLessonRequest_StudentInfo, 0, len(req.StudentInfoList))
	for _, v := range req.StudentInfoList {
		studentInfoList = append(studentInfoList, &bpb.CreateLessonRequest_StudentInfo{
			StudentId:        v.StudentId,
			CourseId:         v.CourseId,
			AttendanceStatus: v.AttendanceStatus,
			LocationId:       v.LocationId,
		})
	}

	return s.ValidateLessonForCreatedRequestMGMT(ctx, e, &bpb.CreateLessonRequest{
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		TeachingMedium:  req.TeachingMedium,
		TeachingMethod:  req.TeachingMethod,
		TeacherIds:      req.TeacherIds,
		CenterId:        req.CenterId,
		StudentInfoList: studentInfoList,
		Materials:       req.Materials,
		SavingOption: &bpb.CreateLessonRequest_SavingOption{
			Method: req.SavingOption.Method,
		},
		ClassId:          req.ClassId,
		CourseId:         req.CourseId,
		SchedulingStatus: req.SchedulingStatus,
	})
}

func (s *Suite) ValidateLessonCreatedSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	handlerLessonCreatedSubscription := func(ctx context.Context, data []byte) (bool, error) {
		r := &bpb.EvtLesson{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return false, err
		}
		switch r.Message.(type) {
		case *bpb.EvtLesson_CreateLessons_:
			timer := time.NewTimer(time.Minute * 1)
			defer timer.Stop()
			for {
				switch req := stepState.Request.(type) {
				case *bpb.CreateLessonRequest:
					if stepState.CurrentLessonID != "" {
						learnerIDs := make([]string, 0, len(req.StudentInfoList))
						for _, studentInfo := range req.StudentInfoList {
							learnerIDs = append(learnerIDs, studentInfo.StudentId)
						}
						lesson := r.GetCreateLessons().Lessons[0]
						if lesson.LessonId == stepState.CurrentLessonID && cmp.Equal(learnerIDs, lesson.LearnerIds) {
							stepState.FoundChanForJetStream <- r.Message
							return true, nil
						}
						return false, nil
					}
				}
			}
		}
		return false, fmt.Errorf("StudentID not equal leanerID")
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonCreated, opts, handlerLessonCreatedSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ValidateLessonCreatedSubscriptionInLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	handlerLessonCreatedSubscription := func(ctx context.Context, data []byte) (bool, error) {
		r := &bpb.EvtLesson{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return false, err
		}
		switch r.Message.(type) {
		case *bpb.EvtLesson_CreateLessons_:
			timer := time.NewTimer(time.Minute * 1)
			defer timer.Stop()
			for {
				switch req := stepState.Request.(type) {
				case *lpb.CreateLessonRequest:
					if stepState.CurrentLessonID != "" {
						learnerIDs := make([]string, 0, len(req.StudentInfoList))
						for _, studentInfo := range req.StudentInfoList {
							learnerIDs = append(learnerIDs, studentInfo.StudentId)
						}
						lesson := r.GetCreateLessons().Lessons[0]
						if lesson.LessonId == stepState.CurrentLessonID && cmp.Equal(learnerIDs, lesson.LearnerIds) {
							stepState.FoundChanForJetStream <- r.Message
							return true, nil
						}
						return false, nil
					}
				}
			}
		}
		return false, fmt.Errorf("StudentID not equal leanerID")
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonCreated, opts, handlerLessonCreatedSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) CreateLessons(ctx context.Context, num string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numOfLesson, _ := strconv.Atoi(num)
	for i := 0; i < numOfLesson; i++ {
		req := DefaultCreateLessonRequest(ctx, "")
		stepState.Requests = append(stepState.Requests, req.CreateLessonRequest)
		res, err := lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).CreateLesson(contextWithToken(s, ctx), req.CreateLessonRequest)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Responses = append(stepState.Responses, res)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) CreateLessonsWithAttendanceStatus(ctx context.Context, num string, attendanceStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numOfLesson, _ := strconv.Atoi(num)
	for i := 0; i < numOfLesson; i++ {
		req := DefaultCreateLessonRequest(ctx, attendanceStatus)
		stepState.Requests = append(stepState.Requests, req.CreateLessonRequest)
		res, err := lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).CreateLesson(contextWithToken(s, ctx), req.CreateLessonRequest)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Responses = append(stepState.Responses, res)
	}
	return StepStateToContext(ctx, stepState), nil
}
