package consumers

import (
	"context"
	"fmt"
	"testing"
	"time"

	domain_lesson "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_lesson_repo "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_repositories_report "github.com/manabie-com/backend/mock/lessonmgmt/lesson_report/repositories"
	mock_class_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/class/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestStudentChangeClassHandler_Handle(t *testing.T) {
	t.Parallel()

	jsm := new(mock_nats.JetStreamManagement)
	lessonRepo := new(mock_lesson_repo.MockLessonRepo)
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}

	lessonMemberRepo := new(mock_lesson_repo.MockLessonMemberRepo)
	classMemberRepo := new(mock_class_repo.MockClassMemberRepo)
	unleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, unleashClient, "local")
	lessonReportRepo := new(mock_repositories_report.MockLessonReportRepo)
	classID := "class-id"
	studentID := "student-id"
	mockLessons := []*domain_lesson.Lesson{
		{LessonID: "lesson-1", CourseID: "course-1"},
		{LessonID: "lesson-2", CourseID: "course-2"},
	}

	getEvent := func(eventChange *mpb.EvtClass) []byte {
		msg, _ := proto.Marshal(eventChange)
		return msg
	}

	getEventLeaveClass := getEvent(&mpb.EvtClass{
		Message: &mpb.EvtClass_LeaveClass_{
			LeaveClass: &mpb.EvtClass_LeaveClass{
				ClassId: classID,
				UserId:  studentID,
			},
		},
	})

	getEventJoinClass := getEvent(&mpb.EvtClass{
		Message: &mpb.EvtClass_JoinClass_{
			JoinClass: &mpb.EvtClass_JoinClass{
				ClassId: classID,
				UserId:  studentID,
			},
		},
	})

	type TestCase struct {
		name        string
		getMessage  []byte
		expectedErr error
		setup       func(ctx context.Context)
	}

	testCases := []TestCase{
		{
			name:       "should success when receive event join class",
			getMessage: getEventJoinClass,
			setup: func(ctx context.Context) {
				mapClassMember := make(map[string]*domain.ClassMember)
				startDateOfStudentInClass := time.Now()
				endDateOfStudentInClass := time.Now().AddDate(0, 3, 0)
				mapClassMember[studentID] = &domain.ClassMember{ClassMemberID: "class-member-id",
					ClassID: classID, UserID: studentID, StartDate: startDateOfStudentInClass, EndDate: endDateOfStudentInClass}

				unleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, db, classID, []string{studentID}).
					Once().Return(mapClassMember, nil)

				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, db, mock.Anything).
					Once().Return(mockLessons, nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				lessonMemberRepo.On("DeleteLessonMembersByStartDate", ctx, tx, studentID, classID, mock.Anything).
					Once().Return([]string{"lesson-3"}, nil)
				lessonMemberRepo.On("InsertLessonMembers", ctx, tx, mock.Anything).
					Once().Return(nil)
				lessonMemberRepo.On("GetLessonMembersInLessons", ctx, tx, mock.Anything).
					Once().Return([]*domain_lesson.LessonMember{
					{
						LessonID: "lesson-1",
					},
					{
						LessonID: "lesson-2",
					},
				}, nil)
				lessonReportRepo.On("DeleteLessonReportWithoutStudent", ctx, tx, []string{"lesson-3"}).Once().Return(nil)
			},
		},
		{
			name:        "should success when receive event leave class",
			expectedErr: nil,
			getMessage:  getEventLeaveClass,
			setup: func(ctx context.Context) {
				mapClassMember := make(map[string]*domain.ClassMember)
				startDateOfStudentInClass := time.Now()
				endDateOfStudentInClass := time.Now().AddDate(0, 3, 0)
				mapClassMember[studentID] = &domain.ClassMember{ClassMemberID: "class-member-id",
					ClassID: classID, UserID: studentID, StartDate: startDateOfStudentInClass, EndDate: endDateOfStudentInClass}

				unleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, db, mock.Anything).
					Once().Return(mockLessons, nil)

				lessonMemberRepo.On("DeleteLessonMembers", ctx, db, mock.Anything).
					Once().Return(nil)
			},
		},
		{
			name:        "should throw error when not found student in class",
			expectedErr: fmt.Errorf("Not found student %s in class %s", studentID, classID),
			getMessage:  getEventJoinClass,
			setup: func(ctx context.Context) {
				unleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				unleashClient.On("WaitForUnleashReady").Once()
				unleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, db, classID, []string{studentID}).
					Once().Return(nil, fmt.Errorf("Not found"))

			},
		},
		{
			name:        "should not throw error when not found lesson match condition",
			expectedErr: nil,
			getMessage:  getEventJoinClass,
			setup: func(ctx context.Context) {

				mapClassMember := make(map[string]*domain.ClassMember)
				startDateOfStudentInClass := time.Now()
				endDateOfStudentInClass := time.Now().AddDate(0, 3, 0)

				mapClassMember[studentID] = &domain.ClassMember{ClassMemberID: "class-member-id",
					ClassID: classID, UserID: studentID, StartDate: startDateOfStudentInClass, EndDate: endDateOfStudentInClass}

				unleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, db, classID, []string{studentID}).
					Once().Return(mapClassMember, nil)

				lessonMemberRepo.On("DeleteLessonMembersByStartDate", ctx, tx, mock.Anything, mock.Anything, mock.Anything).
					Once().Return([]string{}, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, db, mock.Anything).
					Once().Return(make([]*domain_lesson.Lesson, 0, 2), nil)

			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			tc.setup(ctx)
			studentChangeClassHandler := &StudentChangeClassHandler{
				Logger:            ctxzap.Extract(ctx),
				DB:                db,
				WrapperConnection: wrapperConnection,
				JSM:               jsm,
				LessonRepo:        lessonRepo,
				LessonMemberRepo:  lessonMemberRepo,
				ClassMemberRepo:   classMemberRepo,
				LessonReportRepo:  lessonReportRepo,
			}
			_, err := studentChangeClassHandler.Handle(ctx, tc.getMessage)
			if tc.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
