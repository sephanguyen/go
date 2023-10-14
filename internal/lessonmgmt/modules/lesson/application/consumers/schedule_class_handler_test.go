package consumers

import (
	"context"
	"fmt"
	"testing"
	"time"

	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	class_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
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
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setupMock(ctx context.Context) (*ScheduleClassHandler, *mock_database.Ext, *mock_database.Ext, *mock_database.Tx, *mock_unleash_client.UnleashClientInstance, *mock_class_repo.MockClassMemberRepo, *mock_lesson_repo.MockLessonRepo, *mock_lesson_repo.MockLessonMemberRepo, *mock_repositories_report.MockLessonReportRepo) {
	mockBobDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(mockBobDB, mockBobDB, mockUnleashClient, "local")
	lessonRepo := new(mock_lesson_repo.MockLessonRepo)
	lessonMemberRepo := new(mock_lesson_repo.MockLessonMemberRepo)
	classMemberRepo := new(mock_class_repo.MockClassMemberRepo)
	lessonReportRepo := new(mock_repositories_report.MockLessonReportRepo)

	s := &ScheduleClassHandler{
		Logger:            ctxzap.Extract(ctx),
		BobDB:             mockBobDB,
		WrapperConnection: wrapperConnection,
		JSM:               jsm,
		LessonRepo:        lessonRepo,
		LessonMemberRepo:  lessonMemberRepo,
		ClassMemberRepo:   classMemberRepo,
		LessonReportRepo:  lessonReportRepo,
	}

	return s, mockBobDB, mockBobDB, tx, mockUnleashClient, classMemberRepo, lessonRepo, lessonMemberRepo, lessonReportRepo
}

func TestScheduleClassHandler_HandleScheduleClassEvent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s, mockBobDB, mockWrapperDB, tx, mockUnleashClient, classMemberRepo, lessonRepo, lessonMemberRepo, lessonReportRepo := setupMock(ctx)

	now := time.Now()
	classID := "class_id"
	studentID := "student_id"
	currentClassID := "current_class_id"
	effectiveDate := &timestamppb.Timestamp{Seconds: now.Unix()}
	oldClassID := "old_class_id"
	oldEffectiveDate := &timestamppb.Timestamp{Seconds: now.Unix()}

	mapClassMember := make(map[string]*class_domain.ClassMember)

	mapClassMember[studentID] = &class_domain.ClassMember{
		ClassMemberID: "class_member_id",
		ClassID:       currentClassID,
		UserID:        studentID,
		StartDate:     now.Add(-10 * 24 * time.Hour),
		EndDate:       now.Add(-30 * 24 * time.Hour),
	}

	eventData := &mpb.EvtScheduleClass{
		Message: &mpb.EvtScheduleClass_ScheduleClass_{
			ScheduleClass: &mpb.EvtScheduleClass_ScheduleClass{
				ScheduleClassId:           classID,
				UserId:                    studentID,
				CurrentClassId:            currentClassID,
				EffectiveDate:             effectiveDate,
				OldScheduledClassId:       oldClassID,
				OldScheduledEffectiveDate: oldEffectiveDate,
			},
		},
	}

	tempLessons := []*lesson_domain.Lesson{
		{
			LessonID: "lesson_id",
			CourseID: "course_id",
		},
	}

	testCases := []struct {
		name         string
		expectedErr  error
		expectedResp bool
		setup        func(ctx context.Context)
	}{
		{
			name:         "GetByClassIDAndUserIDs fail",
			expectedErr:  fmt.Errorf("GetByClassIDAndUserIDs fail: error"),
			expectedResp: true,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, mockBobDB, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name:         "GetByClassIDAndUserIDs return no resp",
			expectedErr:  fmt.Errorf("not found student %s in class %s", studentID, currentClassID),
			expectedResp: false,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, mockBobDB, mock.Anything, mock.Anything).Once().Return(map[string]*class_domain.ClassMember{}, nil)
			},
		},
		{
			name:         "get lessons of student on current class fail",
			expectedErr:  fmt.Errorf("get lessons of student %s on current class %s: error", studentID, currentClassID),
			expectedResp: false,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, mockBobDB, mock.Anything, mock.Anything).Once().Return(mapClassMember, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, mockWrapperDB, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name:         "get lessons of student on old scheduled class fail",
			expectedErr:  fmt.Errorf("get lessons of student %s on old scheduled class %s: error", studentID, oldClassID),
			expectedResp: false,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, mockBobDB, mock.Anything, mock.Anything).Once().Return(mapClassMember, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, mockWrapperDB, mock.Anything).Once().Return(tempLessons, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, mockWrapperDB, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name:         "get lessons of student on schedule class fail",
			expectedErr:  fmt.Errorf("get lessons of student %s on schedule class %s: error", studentID, classID),
			expectedResp: false,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, mockBobDB, mock.Anything, mock.Anything).Once().Return(mapClassMember, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, mockWrapperDB, mock.Anything).Twice().Return(tempLessons, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, mockWrapperDB, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name:         "upsert lesson members fail",
			expectedErr:  fmt.Errorf("upsert lesson members fail: remove lesson members fail: error"),
			expectedResp: true,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, mockBobDB, mock.Anything, mock.Anything).Once().Return(mapClassMember, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, mockWrapperDB, mock.Anything).Twice().Return(tempLessons, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, mockWrapperDB, mock.Anything).Once().Return(tempLessons, nil)
				mockWrapperDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonMemberRepo.On("DeleteLessonMembers", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name:         "success",
			expectedErr:  nil,
			expectedResp: true,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, mockBobDB, mock.Anything, mock.Anything).Once().Return(mapClassMember, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, mockWrapperDB, mock.Anything).Twice().Return(tempLessons, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, mockWrapperDB, mock.Anything).Once().Return(tempLessons, nil)
				mockWrapperDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				lessonMemberRepo.On("DeleteLessonMembers", ctx, tx, mock.Anything).Once().Return(nil)
				lessonReportRepo.On("DeleteLessonReportWithoutStudent", ctx, tx, mock.Anything).Once().Return(nil)
				lessonMemberRepo.On("InsertLessonMembers", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)

			resp, err := s.handleScheduleClassEvent(ctx, eventData)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedResp, resp)

			mock.AssertExpectationsForObjects(t, mockBobDB, mockWrapperDB, lessonRepo, classMemberRepo, lessonMemberRepo, lessonReportRepo)
		})
	}
}

func TestScheduleClassHandler_CancelScheduledClassEvent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s, mockBobDB, mockWrapperDB, tx, mockUnleashClient, classMemberRepo, lessonRepo, lessonMemberRepo, lessonReportRepo := setupMock(ctx)

	now := time.Now()
	classID := "class_id"
	studentID := "student_id"
	currentClassID := "current_class_id"
	effectiveDate := &timestamppb.Timestamp{Seconds: now.Unix()}

	mapClassMember := make(map[string]*class_domain.ClassMember)

	mapClassMember[studentID] = &class_domain.ClassMember{
		ClassMemberID: "class_member_id",
		ClassID:       currentClassID,
		UserID:        studentID,
		StartDate:     now.Add(-10 * 24 * time.Hour),
		EndDate:       now.Add(-30 * 24 * time.Hour),
	}

	eventData := &mpb.EvtScheduleClass{
		Message: &mpb.EvtScheduleClass_CancelScheduledClass_{
			CancelScheduledClass: &mpb.EvtScheduleClass_CancelScheduledClass{
				ScheduledClassId: classID,
				UserId:           studentID,
				CurrentClassId:   currentClassID,
				EffectiveDate:    effectiveDate,
			},
		},
	}

	tempLessons := []*lesson_domain.Lesson{
		{
			LessonID: "lesson_id",
			CourseID: "course_id",
		},
	}

	testCases := []struct {
		name         string
		expectedErr  error
		expectedResp bool
		setup        func(ctx context.Context)
	}{
		{
			name:         "GetByClassIDAndUserIDs fail",
			expectedErr:  fmt.Errorf("GetByClassIDAndUserIDs fail: error"),
			expectedResp: true,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, mockBobDB, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name:         "GetByClassIDAndUserIDs return no resp",
			expectedErr:  fmt.Errorf("not found student %s in class %s", studentID, currentClassID),
			expectedResp: false,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, mockBobDB, mock.Anything, mock.Anything).Once().Return(map[string]*class_domain.ClassMember{}, nil)
			},
		},
		{
			name:         "get lessons of student on current class fail",
			expectedErr:  fmt.Errorf("get lessons of student %s on current active class %s: error", studentID, currentClassID),
			expectedResp: false,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, mockBobDB, mock.Anything, mock.Anything).Once().Return(mapClassMember, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, mockBobDB, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name:         "get lessons of student on old scheduled class fail",
			expectedErr:  fmt.Errorf("get lessons of student %s on scheduled class %s: error", studentID, classID),
			expectedResp: false,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, mockBobDB, mock.Anything, mock.Anything).Once().Return(mapClassMember, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, mockBobDB, mock.Anything).Once().Return(tempLessons, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, mockBobDB, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name:         "upsert lesson members fail",
			expectedErr:  fmt.Errorf("upsert lesson members fail: remove lesson members fail: error"),
			expectedResp: true,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, mockBobDB, mock.Anything, mock.Anything).Once().Return(mapClassMember, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, mockWrapperDB, mock.Anything).Twice().Return(tempLessons, nil)
				mockWrapperDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonMemberRepo.On("DeleteLessonMembers", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name:         "success",
			expectedErr:  nil,
			expectedResp: true,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classMemberRepo.On("GetByClassIDAndUserIDs", ctx, mockBobDB, mock.Anything, mock.Anything).Once().Return(mapClassMember, nil)
				lessonRepo.On("GetLessonsTeachingModelGroupByClassIdWithDuration", ctx, mockWrapperDB, mock.Anything).Twice().Return(tempLessons, nil)
				mockWrapperDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				lessonMemberRepo.On("DeleteLessonMembers", ctx, tx, mock.Anything).Once().Return(nil)
				lessonReportRepo.On("DeleteLessonReportWithoutStudent", ctx, tx, mock.Anything).Once().Return(nil)
				lessonMemberRepo.On("InsertLessonMembers", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)

			resp, err := s.cancelScheduledClassEvent(ctx, eventData)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedResp, resp)

			mock.AssertExpectationsForObjects(t, mockBobDB, mockWrapperDB, lessonRepo, classMemberRepo, lessonMemberRepo, lessonReportRepo)
		})
	}
}

func TestScheduleClassHandler_UpsertLessonMembers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s, mockBobDB, mockWrapperDB, tx, mockUnleashClient, _, _, lessonMemberRepo, lessonReportRepo := setupMock(ctx)

	tempLessonMembers := []*lesson_domain.LessonMember{
		{
			LessonID:  "lesson_id",
			StudentID: "student_id",
			CourseID:  "course_id",
		},
	}

	testCases := []struct {
		name         string
		expectedErr  error
		expectedResp bool
		setup        func(ctx context.Context)
	}{
		{
			name:        "DeleteLessonReportWithoutStudent fail",
			expectedErr: fmt.Errorf("clean lesson reports fail: error"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockWrapperDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonMemberRepo.On("DeleteLessonMembers", ctx, tx, mock.Anything).Once().Return(nil)
				lessonReportRepo.On("DeleteLessonReportWithoutStudent", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name:        "InsertLessonMembers fail",
			expectedErr: fmt.Errorf("insert lesson members fail: error"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockWrapperDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				tx.On("Rollback", ctx).Return(nil).Once()
				lessonMemberRepo.On("DeleteLessonMembers", ctx, tx, mock.Anything).Once().Return(nil)
				lessonReportRepo.On("DeleteLessonReportWithoutStudent", ctx, tx, mock.Anything).Once().Return(nil)
				lessonMemberRepo.On("InsertLessonMembers", ctx, tx, mock.Anything).Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name:        "success",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mockWrapperDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				lessonMemberRepo.On("DeleteLessonMembers", ctx, tx, mock.Anything).Once().Return(nil)
				lessonReportRepo.On("DeleteLessonReportWithoutStudent", ctx, tx, mock.Anything).Once().Return(nil)
				lessonMemberRepo.On("InsertLessonMembers", ctx, tx, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)

			err := s.upsertLessonMembers(ctx, mockWrapperDB, tempLessonMembers, tempLessonMembers)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, mockBobDB, mockWrapperDB, lessonMemberRepo, lessonReportRepo)
		})
	}

	// test for case lessonMembersAdded and lessonMembersAdded just be empty
	t.Run("pass empty array lesson members", func(t *testing.T) {
		mockWrapperDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Commit", mock.Anything).Return(nil)
		lessonMemberRepo.On("DeleteLessonMembers", ctx, tx, mock.Anything).Once().Return(nil)

		err := s.upsertLessonMembers(ctx, mockWrapperDB, []*lesson_domain.LessonMember{}, []*lesson_domain.LessonMember{})
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockWrapperDB, lessonMemberRepo)
	})
}
