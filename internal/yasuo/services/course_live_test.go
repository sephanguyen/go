package services

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/types"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	"github.com/pkg/errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	bobRepo "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/yasuo/repositories"
	pb_bob "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	"github.com/nats-io/nats.go"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestValidControlSettingLiveLesson_Error(t *testing.T) {
	t.Parallel()
	lessonRepo := &mock_repositories.MockLessonRepo{}
	courseRepo := &bobRepo.MockCourseRepo{}
	teacherRepo := &mock_repositories.MockTeacherRepo{}
	courseService := &CourseService{
		CourseRepo:  courseRepo,
		LessonRepo:  lessonRepo,
		TeacherRepo: teacherRepo,
	}

	testCases := map[string]TestCase{
		"missing teacher teach": {
			req: &pb.ControlSettingLiveLesson{
				TeacherObversers:          []string{"teacher-obversers"},
				Lectures:                  []string{},
				DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
				PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
				UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing teacher teach"),
		},
		"missing unmute student audio": {
			req: &pb.ControlSettingLiveLesson{
				TeacherObversers:          []string{"teacher-obversers"},
				Lectures:                  []string{"teacher-teach"},
				DefaultView:               pb_bob.LIVE_LESSON_VIEW_NONE,
				PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
				UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_NONE,
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing unmute student audio"),
		},
		"missing publish student video": {
			req: &pb.ControlSettingLiveLesson{
				TeacherObversers:          []string{"teacher-obversers"},
				Lectures:                  []string{"teacher-teach"},
				DefaultView:               pb_bob.LIVE_LESSON_VIEW_NONE,
				PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_NONE,
				UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing publish student video"),
		},
		"missing live lesson view": {
			req: &pb.ControlSettingLiveLesson{
				TeacherObversers:          []string{"teacher-obversers"},
				Lectures:                  []string{"teacher-teach"},
				DefaultView:               pb_bob.LIVE_LESSON_VIEW_NONE,
				PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
				UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing live lesson view"),
		},
	}

	for caseName, testCase := range testCases {
		caseName := caseName
		testCase := testCase
		t.Run(caseName, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			err := courseService.handleControlSettingLiveLesson(ctx, testCase.req.(*pb.ControlSettingLiveLesson), database.Int4(1))
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
		})
	}
}

func TestHandleControlSettingLiveLesson_Error(t *testing.T) {
	t.Parallel()
	lessonRepo := &mock_repositories.MockLessonRepo{}
	courseRepo := &bobRepo.MockCourseRepo{}
	teacherRepo := &mock_repositories.MockTeacherRepo{}
	courseService := &CourseService{
		CourseRepo:  courseRepo,
		LessonRepo:  lessonRepo,
		TeacherRepo: teacherRepo,
	}

	testCases := map[string]TestCase{
		"not found teacher": {
			req: &pb.ControlSettingLiveLesson{
				TeacherObversers:          []string{"teacher-obversers"},
				Lectures:                  []string{"teacher-teach"},
				DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
				PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
				UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
			},
			expectedErr: status.Error(codes.NotFound, "cannot find teacher"),
			setup: func(ctx context.Context) {
				teacherRepo.On("ManyTeacherIsInSchool", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(false, pgx.ErrNoRows)
			},
		},
		"teacher doesnot in school": {
			req: &pb.ControlSettingLiveLesson{
				TeacherObversers:          []string{"teacher-obversers"},
				Lectures:                  []string{"teacher-teach"},
				DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
				PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
				UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
			},
			expectedErr: status.Error(codes.PermissionDenied, "teacher doesnot in school"),
			setup: func(ctx context.Context) {
				teacherRepo.On("ManyTeacherIsInSchool", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(false, nil)
			},
		},
		"happy case": {
			req: &pb.ControlSettingLiveLesson{
				TeacherObversers:          []string{"teacher-obversers"},
				Lectures:                  []string{"teacher-teach"},
				DefaultView:               pb_bob.LIVE_LESSON_VIEW_GALLERY,
				PublishStudentVideoStatus: pb_bob.PUBLISH_STUDENT_VIDEO_STATUS_ON,
				UnmuteStudentAudioStatus:  pb_bob.UNMUTE_STUDENT_AUDIO_STATUS_ON,
			},
			setup: func(ctx context.Context) {
				teacherRepo.On("ManyTeacherIsInSchool", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(true, nil)
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			err := courseService.handleControlSettingLiveLesson(ctx, testCase.req.(*pb.ControlSettingLiveLesson), database.Int4(1))
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
		})
	}
}

func TestUpdateTimeLiveCourse_Error(t *testing.T) {
	t.Parallel()
	lessonRepo := &mock_repositories.MockLessonRepo{}
	courseRepo := &bobRepo.MockCourseRepo{}
	courseService := &CourseService{
		CourseRepo: courseRepo,
		LessonRepo: lessonRepo,
	}
	mockTxer := &mock_database.Tx{}

	type UpdateTimeCourse struct {
		Course *entities_bob.Course
	}
	course := &entities_bob.Course{}
	database.AllNullEntity(course)
	_ = course.TeacherIDs.Set([]string{"teacher-id"})
	_ = course.ID.Set("course-id")

	var start, end *time.Time
	testCases := map[string]TestCase{
		"missing course": {
			req: &UpdateTimeCourse{
				Course: &entities_bob.Course{
					ID: pgtype.Text{String: ""},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing course"),
			setup: func(ctx context.Context) {
			},
		},
		"happy case": {
			req: &UpdateTimeCourse{
				Course: course,
			},

			setup: func(ctx context.Context) {
				lessonRepo.On("FindEarlierAndLatestTimeLesson", ctx, mockTxer, course.ID).Once().Return(start, end, nil)
				courseRepo.On("Upsert", ctx, mockTxer, []*entities_bob.Course{course}).Once().Return(nil)
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			err := courseService.updateTimeLiveCourse(ctx, mockTxer, testCase.req.(*UpdateTimeCourse).Course)
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
		})
	}
}

func TestUpdateTimeLiveCourseLessonmgmt_Error(t *testing.T) {
	t.Parallel()
	lessonRepo := &mock_repositories.MockLessonRepo{}
	courseRepo := &bobRepo.MockCourseRepo{}
	courseService := &CourseService{
		CourseRepo: courseRepo,
		LessonRepo: lessonRepo,
	}
	mockTxer := &mock_database.Tx{}
	mockLessonTxer := &mock_database.Tx{}

	type UpdateTimeCourse struct {
		Course *entities_bob.Course
	}
	course := &entities_bob.Course{}
	database.AllNullEntity(course)
	_ = course.TeacherIDs.Set([]string{"teacher-id"})
	_ = course.ID.Set("course-id")

	var start, end *time.Time
	testCases := map[string]TestCase{
		"missing course": {
			req: &UpdateTimeCourse{
				Course: &entities_bob.Course{
					ID: pgtype.Text{String: ""},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing course"),
			setup: func(ctx context.Context) {
			},
		},
		"happy case": {
			req: &UpdateTimeCourse{
				Course: course,
			},

			setup: func(ctx context.Context) {
				lessonRepo.On("FindEarlierAndLatestTimeLesson", ctx, mockLessonTxer, course.ID).Once().Return(start, end, nil)
				courseRepo.On("Upsert", ctx, mockTxer, []*entities_bob.Course{course}).Once().Return(nil)
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			err := courseService.updateTimeLiveCourse(ctx, mockTxer, testCase.req.(*UpdateTimeCourse).Course)
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
		})
	}
}

func TestCourseService_SyncStudentLesson(t *testing.T) {
	t.Parallel()
	lessonRepo := &mock_repositories.MockLessonRepo{}
	lessonMemberRepo := &bobRepo.MockLessonMemberRepo{}

	jsm := new(mock_nats.JetStreamManagement)

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	batch := &mock_database.BatchResults{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	batch.On("Close").Return(nil)
	db.On("Begin").Return(tx, nil)
	tx.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Commit", mock.Anything).Return(nil)
	tx.On("SendBatch", mock.Anything, mock.Anything).Return(batch)
	courseService := &CourseService{
		DBTrace:          tx,
		UnleashClientIns: mockUnleashClient,
		LessonRepo:       lessonRepo,
		LessonMemberRepo: lessonMemberRepo,
		JSM:              jsm,
		Env:              "local",
	}
	insertedStudent := idutil.ULIDNow()
	deletedStudent := idutil.ULIDNow()
	existJoinedLesson := idutil.ULIDNow()
	existButToLeaveLesson := idutil.ULIDNow()
	nonExistLesson := idutil.ULIDNow()

	insertedStudentMembers := []*entities.LessonMember{
		{
			LessonID: database.Text(existJoinedLesson),
			UserID:   database.Text(insertedStudent),
		},
		{
			LessonID: database.Text(existButToLeaveLesson),
			UserID:   database.Text(insertedStudent),
		},
	}
	deletedStudentMembers := []*entities.LessonMember{
		{
			LessonID: database.Text(existButToLeaveLesson),
			UserID:   database.Text(deletedStudent),
		},
		{
			LessonID: database.Text(existJoinedLesson),
			UserID:   database.Text(deletedStudent),
		},
	}
	testCases := map[string]TestCase{
		"joining non exist lesson": {
			req: []*npb.EventSyncUserCourse_StudentLesson{
				{
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					StudentId:  insertedStudent,
					LessonIds:  []string{existJoinedLesson, nonExistLesson},
				},
			},
			expectedErr: fmt.Errorf("studentID %s can not join not existed lessons %v", insertedStudent, []string{nonExistLesson}),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.On("Find", mock.Anything, mock.Anything, database.Text(insertedStudent)).Once().Return(insertedStudentMembers, nil)
				lessonMemberRepo.On("UpsertQueue", mock.Anything, mock.Anything)
				lessonMemberRepo.On("SoftDelete", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				lessonRepo.On("CheckExisted", mock.Anything, mock.Anything, database.TextArray([]string{nonExistLesson})).Once().
					Return([]string{existJoinedLesson}, []string{nonExistLesson}, nil)
			},
		},
		"upsert removing one lesson": {
			req: []*npb.EventSyncUserCourse_StudentLesson{
				{
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					StudentId:  insertedStudent,
					LessonIds:  []string{existJoinedLesson},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.On("Find", mock.Anything, mock.Anything, database.Text(insertedStudent)).Once().Return(deletedStudentMembers, nil)
				lessonMemberRepo.On("UpsertQueue", mock.Anything, mock.Anything)

				lessonMemberRepo.On("SoftDelete", mock.Anything, mock.Anything, mock.Anything,
					database.TextArray([]string{existButToLeaveLesson})).Return(nil)

				lessonRepo.On("CheckExisted", mock.Anything, mock.Anything, database.TextArray(nil)).Once().
					Return([]string{}, []string{}, nil)
				lessonRepo.On("GetLiveLessons", mock.Anything, mock.Anything, database.TextArray([]string{existButToLeaveLesson})).Once().
					Return([]string{existJoinedLesson}, nil)
				jsm.On("PublishContext", ctx, constants.SubjectSyncStudentLessons, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		"delete success": {
			req: []*npb.EventSyncUserCourse_StudentLesson{
				{
					ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
					StudentId:  deletedStudent,
					LessonIds:  []string{existButToLeaveLesson},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.On("SoftDelete", mock.Anything, mock.Anything,
					database.Text(deletedStudent), database.TextArray([]string{existButToLeaveLesson})).
					Return(nil)
				lessonRepo.On("GetLiveLessons", mock.Anything, mock.Anything, database.TextArray([]string{existButToLeaveLesson})).Once().
					Return([]string{existJoinedLesson}, nil)
				jsm.On("PublishContext", ctx, constants.SubjectSyncStudentLessons, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			err := courseService.SyncStudentLesson(ctx, testCase.req.([]*npb.EventSyncUserCourse_StudentLesson))
			if testCase.expectedErr != nil {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestCourseService_SyncStudentLessonmgmt(t *testing.T) {
	t.Parallel()
	lessonRepo := &mock_repositories.MockLessonRepo{}
	lessonMemberRepo := &bobRepo.MockLessonMemberRepo{}

	jsm := new(mock_nats.JetStreamManagement)

	db := &mock_database.Ext{}
	lessonDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	batch := &mock_database.BatchResults{}
	batch.On("Close").Return(nil)
	lessonDB.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Commit", mock.Anything).Return(nil)
	tx.On("SendBatch", mock.Anything, mock.Anything).Return(batch)
	courseService := &CourseService{
		DBTrace:          db,
		LessonDBTrace:    lessonDB,
		LessonRepo:       lessonRepo,
		LessonMemberRepo: lessonMemberRepo,
		JSM:              jsm,
		Env:              "local",
	}
	insertedStudent := idutil.ULIDNow()
	deletedStudent := idutil.ULIDNow()
	existJoinedLesson := idutil.ULIDNow()
	existButToLeaveLesson := idutil.ULIDNow()
	nonExistLesson := idutil.ULIDNow()

	insertedStudentMembers := []*entities.LessonMember{
		{
			LessonID: database.Text(existJoinedLesson),
			UserID:   database.Text(insertedStudent),
		},
		{
			LessonID: database.Text(existButToLeaveLesson),
			UserID:   database.Text(insertedStudent),
		},
	}
	deletedStudentMembers := []*entities.LessonMember{
		{
			LessonID: database.Text(existButToLeaveLesson),
			UserID:   database.Text(deletedStudent),
		},
		{
			LessonID: database.Text(existJoinedLesson),
			UserID:   database.Text(deletedStudent),
		},
	}
	testCases := map[string]TestCase{
		"joining non exist lesson": {
			req: []*npb.EventSyncUserCourse_StudentLesson{
				{
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					StudentId:  insertedStudent,
					LessonIds:  []string{existJoinedLesson, nonExistLesson},
				},
			},
			expectedErr: fmt.Errorf("studentID %s can not join not existed lessons %v", insertedStudent, []string{nonExistLesson}),
			setup: func(ctx context.Context) {
				lessonMemberRepo.On("Find", ctx, lessonDB, database.Text(insertedStudent)).Once().Return(insertedStudentMembers, nil)
				lessonMemberRepo.On("UpsertQueue", mock.Anything, mock.Anything)
				lessonMemberRepo.On("SoftDelete", ctx, lessonDB, mock.Anything, mock.Anything).Return(nil)

				lessonRepo.On("CheckExisted", ctx, lessonDB, database.TextArray([]string{nonExistLesson})).Once().
					Return([]string{existJoinedLesson}, []string{nonExistLesson}, nil)
			},
		},
		"upsert removing one lesson": {
			req: []*npb.EventSyncUserCourse_StudentLesson{
				{
					ActionKind: npb.ActionKind_ACTION_KIND_UPSERTED,
					StudentId:  insertedStudent,
					LessonIds:  []string{existJoinedLesson},
				},
			},
			setup: func(ctx context.Context) {
				lessonMemberRepo.On("Find", ctx, lessonDB, database.Text(insertedStudent)).Once().Return(deletedStudentMembers, nil)
				lessonMemberRepo.On("UpsertQueue", mock.Anything, mock.Anything)

				lessonMemberRepo.On("SoftDelete", ctx, lessonDB, mock.Anything,
					database.TextArray([]string{existButToLeaveLesson})).Return(nil)

				lessonRepo.On("CheckExisted", ctx, lessonDB, database.TextArray(nil)).Once().
					Return([]string{}, []string{}, nil)
				lessonRepo.On("GetLiveLessons", ctx, lessonDB, database.TextArray([]string{existButToLeaveLesson})).Once().
					Return([]string{existJoinedLesson}, nil)
				jsm.On("PublishContext", ctx, constants.SubjectSyncStudentLessons, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		"delete success": {
			req: []*npb.EventSyncUserCourse_StudentLesson{
				{
					ActionKind: npb.ActionKind_ACTION_KIND_DELETED,
					StudentId:  deletedStudent,
					LessonIds:  []string{existButToLeaveLesson},
				},
			},
			setup: func(ctx context.Context) {
				lessonMemberRepo.On("SoftDelete", ctx, lessonDB,
					database.Text(deletedStudent), database.TextArray([]string{existButToLeaveLesson})).
					Return(nil)
				lessonRepo.On("GetLiveLessons", ctx, lessonDB, database.TextArray([]string{existButToLeaveLesson})).Once().
					Return([]string{existJoinedLesson}, nil)
				jsm.On("PublishContext", ctx, constants.SubjectSyncStudentLessons, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			err := courseService.SyncStudentLessonLessonmgmt(ctx, testCase.req.([]*npb.EventSyncUserCourse_StudentLesson))
			if testCase.expectedErr != nil {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestCourseService_CreateLiveLesson(t *testing.T) {
	t.Parallel()
	courseRepo := &mock_repositories.MockCourseRepo{}
	lessonRepo := &mock_repositories.MockLessonRepo{}
	jsm := new(mock_nats.JetStreamManagement)

	db := &mock_database.Ext{}
	lessonDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	batch := &mock_database.BatchResults{}
	batch.On("Close").Return(nil)
	lessonDB.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Commit", mock.Anything).Return(nil)
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("SendBatch", mock.Anything, mock.Anything).Return(batch)
	courseService := &CourseService{
		DBTrace:       db,
		LessonDBTrace: lessonDB,
		CourseRepo:    courseRepo,
		LessonRepo:    lessonRepo,
		JSM:           jsm,
		Env:           "local",
	}

	startTime := time.Now()
	endTime := startTime.Add(time.Duration(100))
	err := fmt.Errorf("sample error")
	course := &entities.Course{
		ID:        pgtype.Text{String: "courseID"},
		StartDate: pgtype.Timestamptz{Time: startTime},
		EndDate:   pgtype.Timestamptz{Time: endTime},
	}

	testCases := map[string]TestCase{
		"Find course by ID false": {
			req: &CreateLiveLessonOpt{
				CourseID: course.ID.String,
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Sprintf("c.CourseRepo.FindByID id: %v: %v", course.ID, err)),
			setup: func(ctx context.Context) {
				courseRepo.On("FindByID", ctx, db, database.Text(course.ID.String)).
					Return(nil, err).Once()
			},
		},
		"Create lesson false": {
			req: &CreateLiveLessonOpt{
				CourseID: course.ID.String,
				Lessons: []*LiveLessonOpt{
					{
						CreateLiveLessonRequest_Lesson: &pb.CreateLiveLessonRequest_Lesson{
							Name: "lessonName",
							StartDate: &types.Timestamp{
								Seconds: int64(startTime.Second()),
								Nanos:   int32(startTime.Nanosecond()),
							},
							EndDate: &types.Timestamp{
								Seconds: int64(endTime.Second()),
								Nanos:   int32(endTime.Nanosecond()),
							},
							LessonGroup: "lessonGroupID",
						},
					},
				},
			},
			expectedErr: fmt.Errorf("LessonRepo.Create: %w", err),
			setup: func(ctx context.Context) {
				courseRepo.On("FindByID", ctx, db, database.Text(course.ID.String)).
					Return(&entities.Course{}, nil).Once()
				lessonRepo.On("Create", ctx, tx, mock.Anything).
					Return(err).Once()
			},
		},
		"Update Live Course Time false": {
			req: &CreateLiveLessonOpt{
				CourseID: course.ID.String,
				Lessons: []*LiveLessonOpt{
					{
						CreateLiveLessonRequest_Lesson: &pb.CreateLiveLessonRequest_Lesson{
							Name: "lessonName",
							StartDate: &types.Timestamp{
								Seconds: int64(startTime.Second()),
								Nanos:   int32(startTime.Nanosecond()),
							},
							EndDate: &types.Timestamp{
								Seconds: int64(endTime.Second()),
								Nanos:   int32(endTime.Nanosecond()),
							},
							LessonGroup: "lessonGroupID",
						},
					},
				},
			},
			expectedErr: err,
			setup: func(ctx context.Context) {
				courseRepo.On("FindByID", ctx, db, database.Text(course.ID.String)).
					Return(course, nil).Once()
				lessonRepo.On("Create", ctx, tx, mock.Anything).
					Return(nil).Once()
				lessonRepo.On("FindEarlierAndLatestTimeLesson", ctx, tx, course.ID).Return(&startTime, &endTime, nil).Once()
				courseRepo.On("Upsert", ctx, db, []*entities_bob.Course{course}).Return(err).Once()
			},
		},
		"Publish event lesson false": {
			req: &CreateLiveLessonOpt{
				CourseID: course.ID.String,
				Lessons: []*LiveLessonOpt{
					{
						CreateLiveLessonRequest_Lesson: &pb.CreateLiveLessonRequest_Lesson{
							Name: "lessonName",
							StartDate: &types.Timestamp{
								Seconds: int64(startTime.Second()),
								Nanos:   int32(startTime.Nanosecond()),
							},
							EndDate: &types.Timestamp{
								Seconds: int64(endTime.Second()),
								Nanos:   int32(endTime.Nanosecond()),
							},
							LessonGroup: "lessonGroupID",
						},
						LessonType: cpb.LessonType_LESSON_TYPE_ONLINE,
					},
				},
			},
			expectedErr: errors.Wrap(fmt.Errorf("PublishLessonEvt rcv.JSM.PublishAsyncContext Lesson.Created failed, msgID: %s, %w", "msgID", err), "rcv.PublishLessonEvt"),
			setup: func(ctx context.Context) {
				courseRepo.On("FindByID", ctx, db, database.Text(course.ID.String)).
					Return(course, nil).Once()
				lessonRepo.On("Create", ctx, tx, mock.Anything).
					Return(nil).Once()
				lessonRepo.On("FindEarlierAndLatestTimeLesson", ctx, tx, course.ID).Return(&startTime, &endTime, nil).Once()
				courseRepo.On("Upsert", ctx, db, []*entities_bob.Course{course}).Return(nil).Once()
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).
					Return("mgsID", err).Once()
			},
		},
		"Success": {
			req: &CreateLiveLessonOpt{
				CourseID: course.ID.String,
				Lessons: []*LiveLessonOpt{
					{
						CreateLiveLessonRequest_Lesson: &pb.CreateLiveLessonRequest_Lesson{
							Name: "lessonName",
							StartDate: &types.Timestamp{
								Seconds: int64(startTime.Second()),
								Nanos:   int32(startTime.Nanosecond()),
							},
							EndDate: &types.Timestamp{
								Seconds: int64(endTime.Second()),
								Nanos:   int32(endTime.Nanosecond()),
							},
							LessonGroup: "lessonGroupID",
						},
						LessonType: cpb.LessonType_LESSON_TYPE_ONLINE,
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				courseRepo.On("FindByID", ctx, db, database.Text(course.ID.String)).
					Return(course, nil).Once()
				lessonRepo.On("Create", ctx, tx, mock.Anything).
					Return(nil).Once()
				lessonRepo.On("FindEarlierAndLatestTimeLesson", ctx, tx, course.ID).Return(&startTime, &endTime, nil).Once()
				courseRepo.On("Upsert", ctx, db, []*entities_bob.Course{course}).Return(nil).Once()
				jsm.On("PublishAsyncContext", ctx, mock.Anything, mock.Anything).
					Return("mgsID", nil).Once()
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			err := courseService.createLiveLessonLessonmgmt(ctx, testCase.req.(*CreateLiveLessonOpt), false)
			if testCase.expectedErr != nil {
				assert.NotNil(t, err)
			}

			mock.AssertExpectationsForObjects(t, courseRepo, lessonRepo, jsm)
		})
	}
}

func TestCourseService_UpdateLiveLesson(t *testing.T) {
	t.Parallel()
	courseRepo := &mock_repositories.MockCourseRepo{}
	lessonRepo := &mock_repositories.MockLessonRepo{}
	jsm := new(mock_nats.JetStreamManagement)

	db := &mock_database.Ext{}
	lessonDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	batch := &mock_database.BatchResults{}
	batch.On("Close").Return(nil)
	lessonDB.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Commit", mock.Anything).Return(nil)
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("SendBatch", mock.Anything, mock.Anything).Return(batch)
	courseService := &CourseService{
		DBTrace:       db,
		LessonDBTrace: lessonDB,
		CourseRepo:    courseRepo,
		LessonRepo:    lessonRepo,
		JSM:           jsm,
		Env:           "local",
	}

	startTime := time.Now()
	endTime := startTime.Add(time.Duration(100))
	err := fmt.Errorf("sample error")
	course := &entities.Course{
		ID:        pgtype.Text{String: "courseID"},
		StartDate: pgtype.Timestamptz{Time: startTime},
		EndDate:   pgtype.Timestamptz{Time: endTime},
	}
	lesson := &entities.Lesson{
		LessonID:   pgtype.Text{String: "lessonID"},
		Name:       pgtype.Text{String: ""},
		LessonType: pgtype.Text{String: "LESSON_TYPE_ONLINE"},
		CourseID:   pgtype.Text{String: course.ID.String},
		StartTime:  pgtype.Timestamptz{Time: startTime},
		EndTime:    pgtype.Timestamptz{Time: endTime},
	}
	req := &updateLiveLessonV2Request{
		LessonType: cpb.LessonType_LESSON_TYPE_ONLINE,
		UpdateLiveLessonRequest: &pb.UpdateLiveLessonRequest{
			LessonId: lesson.LessonID.String,
			Name:     lesson.Name.String,
			StartDate: &types.Timestamp{
				Seconds: int64(startTime.Second()),
				Nanos:   int32(startTime.Nanosecond()),
			},
			EndDate: &types.Timestamp{
				Seconds: int64(endTime.Second()),
				Nanos:   int32(endTime.Nanosecond()),
			},
			CourseId: lesson.CourseID.String,
		},
	}

	testCases := map[string]TestCase{
		"Find lesson by ID false": {
			req:         req,
			expectedErr: fmt.Errorf("LessonRepo.FindByID: %w", err),
			setup: func(ctx context.Context) {
				lessonRepo.On("FindByID", ctx, lessonDB, database.Text(req.LessonId)).
					Return(nil, err).Once()
			},
		},
		"Find course by ID false": {
			req:         req,
			expectedErr: fmt.Errorf("CourseRepo.FindByID: %w", err),
			setup: func(ctx context.Context) {
				lessonRepo.On("FindByID", ctx, lessonDB, database.Text(req.LessonId)).
					Return(lesson, nil).Once()
				courseRepo.On("FindByID", ctx, db, database.Text(req.CourseId)).
					Return(nil, err).Once()
			},
		},
		"Update lesson false": {
			req:         req,
			expectedErr: fmt.Errorf("LessonRepo.Update: %w", err),
			setup: func(ctx context.Context) {
				lessonRepo.On("FindByID", ctx, lessonDB, database.Text(req.LessonId)).
					Return(lesson, nil).Once()
				courseRepo.On("FindByID", ctx, db, database.Text(req.CourseId)).
					Return(course, nil).Once()
				lessonRepo.On("Update", ctx, tx, mock.MatchedBy(func(l *entities.Lesson) bool {
					return lesson.LessonID == l.LessonID
				})).
					Return(err).Once()
			},
		},
		"Update course time false": {
			req:         req,
			expectedErr: err,
			setup: func(ctx context.Context) {
				lessonRepo.On("FindByID", ctx, lessonDB, database.Text(req.LessonId)).
					Return(lesson, nil).Once()
				courseRepo.On("FindByID", ctx, db, database.Text(req.CourseId)).
					Return(course, nil).Once()
				lessonRepo.On("Update", ctx, tx, mock.MatchedBy(func(l *entities.Lesson) bool {
					return lesson.LessonID == l.LessonID
				})).
					Return(nil).Once()
				lessonRepo.On("FindEarlierAndLatestTimeLesson", ctx, tx, course.ID).Return(&startTime, &endTime, nil).Once()
				courseRepo.On("Upsert", ctx, db, []*entities_bob.Course{course}).Return(err).Once()
			},
		},
		"Success": {
			req:         req,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				lessonRepo.On("FindByID", ctx, lessonDB, database.Text(req.LessonId)).
					Return(lesson, nil).Once()
				courseRepo.On("FindByID", ctx, db, database.Text(req.CourseId)).
					Return(course, nil).Once()
				lessonRepo.On("Update", ctx, tx, mock.MatchedBy(func(l *entities.Lesson) bool {
					return lesson.LessonID == l.LessonID
				})).
					Return(nil).Once()
				lessonRepo.On("FindEarlierAndLatestTimeLesson", ctx, tx, course.ID).Return(&startTime, &endTime, nil).Once()
				courseRepo.On("Upsert", ctx, db, []*entities_bob.Course{course}).Return(nil).Once()
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			err := courseService.updateLiveLessonV2Lessonmgmt(ctx, testCase.req.(*updateLiveLessonV2Request))
			if testCase.expectedErr != nil {
				assert.NotNil(t, err)
			}

			mock.AssertExpectationsForObjects(t, courseRepo, lessonRepo, jsm)
		})
	}
}

func TestCourseService_DeleteLiveCourse(t *testing.T) {
	t.Parallel()
	courseRepo := &mock_repositories.MockCourseRepo{}
	courseClassRepo := &mock_repositories.MockCourseClassRepo{}
	lessonRepo := &mock_repositories.MockLessonRepo{}

	jsm := new(mock_nats.JetStreamManagement)

	db := &mock_database.Ext{}
	lessonDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	batch := &mock_database.BatchResults{}
	batch.On("Close").Return(nil)
	db.On("Begin").Return(tx, nil)
	tx.On("Begin", mock.Anything).Return(tx, nil)
	tx.On("Commit", mock.Anything).Return(nil)
	tx.On("Rollback", mock.Anything).Return(nil)
	tx.On("SendBatch", mock.Anything, mock.Anything).Return(batch)
	courseService := &CourseService{
		DBTrace:         tx,
		LessonDBTrace:   lessonDB,
		CourseRepo:      courseRepo,
		CourseClassRepo: courseClassRepo,
		LessonRepo:      lessonRepo,
		JSM:             jsm,
		Env:             "local",
	}

	err := fmt.Errorf("sample error")

	courseIDs := []string{"course-id-1", "course-id-2"}

	testCases := map[string]TestCase{
		"Soft delete course false": {
			req:         &pb.DeleteLiveCourseRequest{CourseIds: courseIDs},
			expectedErr: fmt.Errorf("c.CourseRepo.SoftDelete: %w", err),
			setup: func(ctx context.Context) {
				courseRepo.On("SoftDelete", ctx, tx, database.TextArray(courseIDs)).
					Return(err).Once()
			},
		},
		"Find By Course IDs false": {
			req:         &pb.DeleteLiveCourseRequest{CourseIds: courseIDs},
			expectedErr: errors.Wrap(err, "c.CourseRepo.FindByIDs"),
			setup: func(ctx context.Context) {
				courseRepo.On("SoftDelete", ctx, tx, database.TextArray(courseIDs)).
					Return(nil).Once()
				courseClassRepo.On("FindByCourseIDs", ctx, tx, database.TextArray(courseIDs), false).
					Return(nil, err).Once()
			},
		},
		"Soft delete course class false": {
			req:         &pb.DeleteLiveCourseRequest{CourseIds: courseIDs},
			expectedErr: fmt.Errorf("c.CourseClassRepo.SoftDelete: %w", err),
			setup: func(ctx context.Context) {
				courseRepo.On("SoftDelete", ctx, tx, database.TextArray(courseIDs)).
					Return(nil).Once()
				courseClassRepo.On("FindByCourseIDs", ctx, tx, database.TextArray(courseIDs), false).
					Return([]*entities.CourseClass{{}}, nil).Once()
				courseClassRepo.On("SoftDelete", ctx, tx, database.TextArray(courseIDs)).
					Return(err).Once()
			},
		},
		"Find lesson by course ids false": {
			req:         &pb.DeleteLiveCourseRequest{CourseIds: courseIDs},
			expectedErr: fmt.Errorf("c.LessonRepo.SoftDeleteByCourseIDs: %w", err),
			setup: func(ctx context.Context) {
				courseRepo.On("SoftDelete", ctx, tx, database.TextArray(courseIDs)).
					Return(nil).Once()
				courseClassRepo.On("FindByCourseIDs", ctx, tx, database.TextArray(courseIDs), false).
					Return([]*entities.CourseClass{{}}, nil).Once()
				courseClassRepo.On("SoftDelete", ctx, tx, database.TextArray(courseIDs)).
					Return(nil).Once()
				lessonRepo.On("FindByCourseIDs", ctx, lessonDB, database.TextArray(courseIDs), false).
					Return(nil, err).Once()
			},
		},
		"Soft delete lesson false": {
			req:         &pb.DeleteLiveCourseRequest{CourseIds: courseIDs},
			expectedErr: fmt.Errorf("c.LessonRepo.SoftDeleteByCourseIDs: %w", err),
			setup: func(ctx context.Context) {
				courseRepo.On("SoftDelete", ctx, tx, database.TextArray(courseIDs)).
					Return(nil).Once()
				courseClassRepo.On("FindByCourseIDs", ctx, tx, database.TextArray(courseIDs), false).
					Return([]*entities.CourseClass{{}}, nil).Once()
				courseClassRepo.On("SoftDelete", ctx, tx, database.TextArray(courseIDs)).
					Return(nil).Once()
				lessonRepo.On("FindByCourseIDs", ctx, lessonDB, database.TextArray(courseIDs), false).
					Return([]*entities.Lesson{{}}, nil).Once()
				lessonRepo.On("SoftDeleteByCourseIDs", ctx, lessonDB, database.TextArray(courseIDs)).
					Return(err).Once()
			},
		},
		"Success": {
			req:         &pb.DeleteLiveCourseRequest{CourseIds: courseIDs},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				courseRepo.On("SoftDelete", ctx, tx, database.TextArray(courseIDs)).
					Return(nil).Once()
				courseClassRepo.On("FindByCourseIDs", ctx, tx, database.TextArray(courseIDs), false).
					Return([]*entities.CourseClass{{}}, nil).Once()
				courseClassRepo.On("SoftDelete", ctx, tx, database.TextArray(courseIDs)).
					Return(nil).Once()
				lessonRepo.On("FindByCourseIDs", ctx, lessonDB, database.TextArray(courseIDs), false).
					Return([]*entities.Lesson{{}}, nil).Once()
				lessonRepo.On("SoftDeleteByCourseIDs", ctx, lessonDB, database.TextArray(courseIDs)).
					Return(nil).Once()
			},
		},
	}

	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			err := courseService.deleteLiveCourseLessonmgmt(ctx, testCase.req.(*pb.DeleteLiveCourseRequest))
			if testCase.expectedErr != nil {
				assert.NotNil(t, err)
			}

			mock.AssertExpectationsForObjects(t, courseRepo, courseClassRepo, lessonRepo)
		})
	}
}
