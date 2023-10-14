package classes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_whiteboard "github.com/manabie-com/backend/mock/golibs/whiteboard"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStudentRetrieveStreamToken(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	classService := &ClassService{
		Cfg: &configurations.Config{},
	}
	lessonID := "teacher-id"
	teacherID := "teacher-id"

	ctx = interceptors.ContextWithUserID(ctx, teacherID)
	testCases := map[string]TestCase{
		"valid subscribe token": {
			ctx: ctx,
			req: &pb.StudentRetrieveStreamTokenRequest{
				LessonId: lessonID,
			},
			expectedResp: &pb.StudentRetrieveStreamTokenResponse{
				StreamToken: "stream token",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(ctx)
			req := testCase.req.(*pb.StudentRetrieveStreamTokenRequest)
			rsp, err := classService.StudentRetrieveStreamToken(ctx, req)
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
			}
		})
	}
}

type LessonModifierServicesMock struct {
	resetAllLiveLessonStatesInternal func(ctx context.Context, lessonID string) error
}

func (l *LessonModifierServicesMock) ResetAllLiveLessonStatesInternal(ctx context.Context, lessonID string) error {
	return l.resetAllLiveLessonStatesInternal(ctx, lessonID)
}

func TestEndLiveLesson_Error(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := new(mock_repositories.MockLessonRepo)
	classRepo := new(mock_repositories.MockClassRepo)
	courseClassRepo := new(mock_repositories.MockCourseClassRepo)

	jsm := new(mock_nats.JetStreamManagement)

	classService := &ClassService{
		Cfg:             &configurations.Config{},
		LessonRepo:      lessonRepo,
		CourseClassRepo: courseClassRepo,
		ClassRepo:       classRepo,
		JSM:             jsm,
	}
	lessonID := "teacher-id"
	teacherID := "teacher-id"

	lesson := &entities.Lesson{}
	_ = lesson.TeacherID.Set(teacherID)
	_ = lesson.LessonID.Set(lessonID)

	lessons := []*entities.Lesson{
		lesson,
	}
	ctx = interceptors.ContextWithUserID(ctx, teacherID)

	testCases := map[string]TestCase{
		"error update time": {
			ctx: ctx,
			req: &pb.EndLiveLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				filter := &repositories.LessonFilter{}
				_ = filter.LessonID.Set([]string{lessonID})
				_ = filter.TeacherID.Set([]string{teacherID})
				_ = filter.CourseID.Set(nil)

				lessonRepo.On("Find", ctx, mock.Anything, filter).Once().Return(lessons, nil)
				lessonRepo.On("EndLiveLesson", ctx, mock.Anything, lesson.LessonID, mock.Anything).Once().Return(pgx.ErrNoRows)

			},
		},
		"valid end live lesson token": {
			ctx: ctx,
			req: &pb.EndLiveLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: &pb.EndLiveLessonResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				filter := &repositories.LessonFilter{}
				_ = filter.LessonID.Set([]string{lessonID})
				_ = filter.TeacherID.Set([]string{teacherID})
				_ = filter.CourseID.Set(nil)

				lessonRepo.On("Find", ctx, mock.Anything, filter).Once().Return(lessons, nil)
				lessonRepo.On("EndLiveLesson", ctx, mock.Anything, lesson.LessonID, mock.Anything).Once().Return(nil)

				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonUpdated, mock.Anything).Once().Return("", nil)
			},
			lessonModSrvMock: &LessonModifierServicesMock{
				resetAllLiveLessonStatesInternal: func(ctx context.Context, actualLessonID string) error {
					assert.Equal(t, lessonID, actualLessonID)
					userID := interceptors.UserIDFromContext(ctx)
					assert.Equal(t, teacherID, userID)
					return nil
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(ctx)
			classService.LessonModifierServices = testCase.lessonModSrvMock
			req := testCase.req.(*pb.EndLiveLessonRequest)
			rsp, err := classService.EndLiveLesson(ctx, req)
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
			}
		})
	}
}

func TestTeacherRetrieveStreamToken_Error(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := new(mock_repositories.MockLessonRepo)
	classRepo := new(mock_repositories.MockClassRepo)
	courseClassRepo := new(mock_repositories.MockCourseClassRepo)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	schoolRepo := new(mock_repositories.MockSchoolRepo)
	courseRepo := new(mock_repositories.MockCourseRepo)
	whiteboardSvc := new(mock_whiteboard.MockService)
	classService := &ClassService{
		Cfg:             &configurations.Config{},
		LessonRepo:      lessonRepo,
		CourseClassRepo: courseClassRepo,
		ClassRepo:       classRepo,
		TeacherRepo:     teacherRepo,
		SchoolRepo:      schoolRepo,
		CourseRepo:      courseRepo,
		WhiteboardSvc:   whiteboardSvc,
	}
	lessonID := "teacher-id"
	teacherID := "teacher-id"

	lesson := &entities.Lesson{}
	_ = lesson.TeacherID.Set(teacherID)
	_ = lesson.LessonID.Set(lessonID)
	_ = lesson.CourseID.Set("course-2")

	lessons := []*entities.Lesson{
		lesson,
	}

	teacher := &entities.Teacher{}
	_ = teacher.ID.Set(teacherID)
	_ = teacher.SchoolIDs.Set([]int{1})

	course := &entities.Course{}
	_ = course.ID.Set("course-2")
	courses := entities.Courses{course}

	invalidCourse := &entities.Course{}
	_ = invalidCourse.ID.Set("course-1")
	invalidCourses := entities.Courses{invalidCourse}
	// emptyLessons := []*entities.Lesson{}
	ctx = interceptors.ContextWithUserID(ctx, teacherID)
	testCases := map[string]TestCase{
		"invalid RetrieveCourses": {
			ctx: ctx,
			req: &pb.TeacherRetrieveStreamTokenRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				teacherRepo.On("FindByID", ctx, mock.Anything, lesson.TeacherID).Once().Return(teacher, nil)
				schoolRepo.On("RetrieveCountries", ctx, mock.Anything, mock.Anything).Once().Return([]string{"COUNTRY_VN"}, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, mock.Anything).Once().Return(invalidCourses, pgx.ErrNoRows)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return("token", nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
			},
		},
		"valid token": {
			ctx: ctx,
			req: &pb.TeacherRetrieveStreamTokenRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				teacherRepo.On("FindByID", ctx, mock.Anything, lesson.TeacherID).Once().Return(teacher, nil)
				schoolRepo.On("RetrieveCountries", ctx, mock.Anything, mock.Anything).Once().Return([]string{"COUNTRY_VN"}, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return("token", nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(ctx)
			req := testCase.req.(*pb.TeacherRetrieveStreamTokenRequest)
			rsp, err := classService.TeacherRetrieveStreamToken(ctx, req)
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
			}
		})
	}
}

func TestLeaveLiveLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	jsm := new(mock_nats.JetStreamManagement)

	classService := &ClassService{
		Cfg: &configurations.Config{},
		JSM: jsm,
	}
	lessonID := "lesson-id"
	userID := "user-id"

	ctx = interceptors.ContextWithUserID(ctx, userID)
	testCases := map[string]TestCase{
		"error publish event": {
			ctx: ctx,
			req: &pb.LeaveLessonRequest{
				LessonId: lessonID,
				UserId:   userID,
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("error publish"),
			setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonUpdated, mock.Anything).Once().Return("", fmt.Errorf("error publish"))
			},
		},

		"success": {
			ctx: ctx,
			req: &pb.LeaveLessonRequest{
				LessonId: lessonID,
				UserId:   userID,
			},
			expectedResp: &pb.LeaveLessonResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonUpdated, mock.Anything).Once().Return("", nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(ctx)
			classService.LessonModifierServices = testCase.lessonModSrvMock
			req := testCase.req.(*pb.LeaveLessonRequest)
			rsp, err := classService.LeaveLesson(ctx, req)
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
			}
		})
	}
}

func TestJoinLiveLesson(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lessonRepo := new(mock_repositories.MockLessonRepo)
	classRepo := new(mock_repositories.MockClassRepo)
	courseClassRepo := new(mock_repositories.MockCourseClassRepo)
	teacherRepo := new(mock_repositories.MockTeacherRepo)
	schoolRepo := new(mock_repositories.MockSchoolRepo)
	courseRepo := new(mock_repositories.MockCourseRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	jsm := new(mock_nats.JetStreamManagement)
	whiteboardSvc := new(mock_whiteboard.MockService)

	var courseIDs pgtype.TextArray
	_ = courseIDs.Set([]string{"course-1", "course-2"})

	var classIDs pgtype.Int4Array
	_ = classIDs.Set([]int{1})

	// emptyLessons := []*entities.Lesson{}
	classService := &ClassService{
		Cfg:             &configurations.Config{},
		UserRepo:        userRepo,
		LessonRepo:      lessonRepo,
		CourseClassRepo: courseClassRepo,
		ClassRepo:       classRepo,
		TeacherRepo:     teacherRepo,
		SchoolRepo:      schoolRepo,
		CourseRepo:      courseRepo,
		WhiteboardSvc:   whiteboardSvc,
		JSM:             jsm,
	}
	lessonID := "teacher-id"
	teacherID := "teacher-id"
	roomID := "room-id"

	lesson := &entities.Lesson{}
	_ = lesson.TeacherID.Set(teacherID)
	_ = lesson.LessonID.Set(lessonID)
	_ = lesson.CourseID.Set("course-2")
	_ = lesson.RoomID.Set(roomID)

	lessons := []*entities.Lesson{
		lesson,
	}

	teacher := &entities.Teacher{}
	_ = teacher.ID.Set(teacherID)
	_ = teacher.SchoolIDs.Set([]int{1})

	course := &entities.Course{}
	_ = course.ID.Set("course-2")
	courses := entities.Courses{course}

	invalidCourse := &entities.Course{}
	_ = invalidCourse.ID.Set("course-1")
	invalidCourses := entities.Courses{invalidCourse}

	ctx = interceptors.ContextWithUserID(ctx, teacherID)
	testCases := map[string]TestCase{
		"student subscribe token": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: &pb.JoinLessonResponse{
				StreamToken:     "",
				WhiteboardToken: "",
				RoomId:          roomID,
				VideoToken:      "token",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				filter := &repositories.LessonFilter{}
				_ = filter.LessonID.Set([]string{lessonID})
				_ = filter.TeacherID.Set(nil)
				_ = filter.CourseID.Set(nil)

				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupStudent, nil)
				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonUpdated, mock.Anything).Once().Return("", nil)
				lessonRepo.On("Find", ctx, mock.Anything, filter).Once().Return(lessons, nil)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return("token", nil)
			},
		},
		"lesson have no room id": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: &pb.JoinLessonResponse{
				StreamToken:     "",
				WhiteboardToken: "",
				RoomId:          "room-id-2",
				VideoToken:      "token",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				filter := &repositories.LessonFilter{}
				_ = filter.LessonID.Set([]string{lessonID})
				_ = filter.TeacherID.Set(nil)
				_ = filter.CourseID.Set(nil)

				lessonNoRoom := &entities.Lesson{}
				_ = lessonNoRoom.TeacherID.Set(teacherID)
				_ = lessonNoRoom.LessonID.Set(lessonID)
				_ = lessonNoRoom.CourseID.Set("course-2")
				userRepo.On("UserGroup", ctx, mock.Anything, lessonNoRoom.TeacherID).Once().Return(entities.UserGroupStudent, nil)
				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonUpdated, mock.Anything).Once().Return("", nil)
				lessonRepo.On("Find", ctx, mock.Anything, filter).Once().Return([]*entities.Lesson{
					lessonNoRoom,
				}, nil)
				whiteboardSvc.On("CreateRoom", ctx, &whiteboard.CreateRoomRequest{
					Name:     lessonNoRoom.LessonID.String,
					IsRecord: false,
				}).Once().Return(&whiteboard.CreateRoomResponse{
					UUID: "room-id-2",
				}, nil)
				lessonRepo.On("UpdateRoomID", ctx, mock.Anything, lessonNoRoom.LessonID, database.Text("room-id-2")).
					Once().Return(nil)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return("token", nil)
			},
		},
		"error finding lesson": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				filter := &repositories.LessonFilter{}
				_ = filter.LessonID.Set([]string{lessonID})
				_ = filter.TeacherID.Set(nil)
				_ = filter.CourseID.Set(nil)

				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupTeacher, nil)
				lessonRepo.On("Find", ctx, mock.Anything, filter).Once().Return(lessons, pgx.ErrNoRows)
			},
		},
		"invalid RetrieveCourses": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupTeacher, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, mock.Anything).Once().Return(invalidCourses, pgx.ErrNoRows)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return("token", nil)
			},
		},
		"invalid subscribe token": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.PermissionDenied, ""),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupTeacher, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, mock.Anything).Once().Return(entities.Courses{}, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Twice().Return(lessons, nil)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return("token", nil)
			},
		},
		"valid subscribe token": {
			ctx: ctx,
			req: &pb.JoinLessonRequest{
				LessonId: lessonID,
			},
			expectedResp: &pb.JoinLessonResponse{
				StreamToken:     "",
				WhiteboardToken: "",
				RoomId:          roomID,
				VideoToken:      "token",
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, lesson.TeacherID).Once().Return(entities.UserGroupTeacher, nil)
				courseRepo.On("RetrieveCourses", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				lessonRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(lessons, nil)
				whiteboardSvc.On("FetchRoomToken", ctx, mock.Anything).Once().Return("token", nil)
				jsm.On("PublishAsyncContext", ctx, constants.SubjectLessonUpdated, mock.Anything).Once().Return("", nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(ctx)
			req := testCase.req.(*pb.JoinLessonRequest)
			rsp, err := classService.JoinLesson(ctx, req)
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
			} else {
				expectedResp := testCase.expectedResp.(*pb.JoinLessonResponse)
				assert.Equal(t, expectedResp.RoomId, rsp.RoomId)
			}
			mock.AssertExpectationsForObjects(t, userRepo, teacherRepo, schoolRepo, courseRepo, lessonRepo, whiteboardSvc, jsm)
		})
	}
}
