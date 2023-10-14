package consumers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_virtual_repo "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestUpcomingLiveLessonNotificationHandler(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ctx = auth.InjectFakeJwtToken(ctx, "1")
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonMemberRepo := &mock_virtual_repo.MockLessonMemberRepo{}
	virtualLessonRepo := &mock_virtual_repo.MockVirtualLessonRepo{}
	liveLessonSentNotificationRepo := &mock_virtual_repo.MockLiveLessonSentNotificationRepo{}
	studentParentRepo := &mock_virtual_repo.MockStudentParentRepo{}
	userRepo := &mock_virtual_repo.MockUserRepo{}
	interval := H24
	lessonIDs := []string{"lesson-id-1"}
	mockMemberUserIds := []string{"user-1"}
	mockVirtualLessons := []*domain.VirtualLesson{
		{
			LessonID:  lessonIDs[0],
			StartTime: time.Now().Add(23 * time.Hour),
		},
	}
	mockLessonMemberUsers := map[string]*domain.User{
		mockMemberUserIds[0]: {
			ID:      mockMemberUserIds[0],
			Country: "COUNTRY_JP",
		},
	}
	mockStudentParent := []domain.StudentParent{
		{
			StudentID: mockMemberUserIds[0],
			ParentID:  "parent-0",
		},
	}

	tcs := []struct {
		name     string
		data     []byte
		setup    func()
		hasError bool
	}{
		{
			name: "happy case",
			data: func() []byte {
				r := &vpb.UpcomingLiveLessonNotificationRequest{
					LessonIds: lessonIDs,
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			setup: func() {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil).Once()

				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				virtualLessonRepo.On("GetVirtualLessonsByLessonIDs", mock.Anything, db, lessonIDs).Return(mockVirtualLessons, nil).Once()
				liveLessonSentNotificationRepo.On("GetLiveLessonSentNotificationCount", mock.Anything, db, lessonIDs[0], interval).Return(int32(0), nil).Once()
				liveLessonSentNotificationRepo.On("CreateLiveLessonSentNotificationRecord", mock.Anything, tx, lessonIDs[0], interval, mock.Anything).Return(nil).Once()
				lessonMemberRepo.On("GetLearnerIDsByLessonID", mock.Anything, tx, lessonIDs[0]).Return(mockMemberUserIds, nil).Once()
				userRepo.On("GetUsersByIDs", mock.Anything, db, mockMemberUserIds).Return(mockLessonMemberUsers, nil).Once()
				studentParentRepo.On("GetStudentParents", mock.Anything, tx, mockMemberUserIds).Return(mockStudentParent, nil).Once()
				jsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "happy case with lessonmgmt db",
			data: func() []byte {
				r := &vpb.UpcomingLiveLessonNotificationRequest{
					LessonIds: lessonIDs,
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			setup: func() {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil).Once()

				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				virtualLessonRepo.On("GetVirtualLessonsByLessonIDs", mock.Anything, db, lessonIDs).Return(mockVirtualLessons, nil).Once()
				liveLessonSentNotificationRepo.On("GetLiveLessonSentNotificationCount", mock.Anything, db, lessonIDs[0], interval).Return(int32(0), nil).Once()
				liveLessonSentNotificationRepo.On("CreateLiveLessonSentNotificationRecord", mock.Anything, tx, lessonIDs[0], interval, mock.Anything).Return(nil).Once()
				lessonMemberRepo.On("GetLearnerIDsByLessonID", mock.Anything, tx, lessonIDs[0]).Return(mockMemberUserIds, nil).Once()
				userRepo.On("GetUsersByIDs", mock.Anything, db, mockMemberUserIds).Return(mockLessonMemberUsers, nil).Once()
				studentParentRepo.On("GetStudentParents", mock.Anything, tx, mockMemberUserIds).Return(mockStudentParent, nil).Once()
				jsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name: "happy case with no learners found",
			data: func() []byte {
				r := &vpb.UpcomingLiveLessonNotificationRequest{
					LessonIds: lessonIDs,
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			setup: func() {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil).Once()
				tx.On("Commit", mock.Anything).Return(nil).Once()

				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				virtualLessonRepo.On("GetVirtualLessonsByLessonIDs", mock.Anything, db, lessonIDs).Return(mockVirtualLessons, nil).Once()
				liveLessonSentNotificationRepo.On("GetLiveLessonSentNotificationCount", mock.Anything, db, lessonIDs[0], interval).Return(int32(0), nil).Once()
				liveLessonSentNotificationRepo.On("CreateLiveLessonSentNotificationRecord", mock.Anything, tx, lessonIDs[0], interval, mock.Anything).Return(nil).Once()
				lessonMemberRepo.On("GetLearnerIDsByLessonID", mock.Anything, tx, lessonIDs[0]).Return([]string{}, nil).Once()
			},
		},
		{
			name: "error failed GetVirtualLessonsByLessonIDs",
			data: func() []byte {
				r := &vpb.UpcomingLiveLessonNotificationRequest{
					LessonIds: lessonIDs,
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			hasError: true,
			setup: func() {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil).Once()
				tx.On("Rollback", mock.Anything).Return(nil).Once()
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				virtualLessonRepo.On("GetVirtualLessonsByLessonIDs", mock.Anything, db, lessonIDs).Return(nil, errors.New("fail")).Once()
			},
		},
		{
			name: "error failed GetLiveLessonSentNotificationCount",
			data: func() []byte {
				r := &vpb.UpcomingLiveLessonNotificationRequest{
					LessonIds: lessonIDs,
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			hasError: true,
			setup: func() {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil).Once()
				tx.On("Rollback", mock.Anything).Return(nil).Once()

				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				virtualLessonRepo.On("GetVirtualLessonsByLessonIDs", mock.Anything, db, lessonIDs).Return(mockVirtualLessons, nil).Once()
				liveLessonSentNotificationRepo.On("GetLiveLessonSentNotificationCount", mock.Anything, db, lessonIDs[0], interval).Return(int32(0), errors.New("fail")).Once()
			},
		},
		{
			name: "error failed CreateLiveLessonSentNotificationRecord",
			data: func() []byte {
				r := &vpb.UpcomingLiveLessonNotificationRequest{
					LessonIds: lessonIDs,
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			hasError: true,
			setup: func() {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil).Once()
				tx.On("Rollback", mock.Anything).Return(nil).Once()

				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				virtualLessonRepo.On("GetVirtualLessonsByLessonIDs", mock.Anything, db, lessonIDs).Return(mockVirtualLessons, nil).Once()
				liveLessonSentNotificationRepo.On("GetLiveLessonSentNotificationCount", mock.Anything, db, lessonIDs[0], interval).Return(int32(0), nil).Once()
				liveLessonSentNotificationRepo.On("CreateLiveLessonSentNotificationRecord", mock.Anything, tx, lessonIDs[0], interval, mock.Anything).Return(errors.New("fail")).Once()
			},
		},
		{
			name: "error failed GetLearnerIDsByLessonID",
			data: func() []byte {
				r := &vpb.UpcomingLiveLessonNotificationRequest{
					LessonIds: lessonIDs,
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			hasError: true,
			setup: func() {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil).Once()
				tx.On("Rollback", mock.Anything).Return(nil).Once()

				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				virtualLessonRepo.On("GetVirtualLessonsByLessonIDs", mock.Anything, db, lessonIDs).Return(mockVirtualLessons, nil).Once()
				liveLessonSentNotificationRepo.On("GetLiveLessonSentNotificationCount", mock.Anything, db, lessonIDs[0], interval).Return(int32(0), nil).Once()
				liveLessonSentNotificationRepo.On("CreateLiveLessonSentNotificationRecord", mock.Anything, tx, lessonIDs[0], interval, mock.Anything).Return(nil).Once()
				lessonMemberRepo.On("GetLearnerIDsByLessonID", mock.Anything, tx, lessonIDs[0]).Return(nil, errors.New("fail")).Once()
			},
		},
		{
			name: "error failed GetUsersByIDs",
			data: func() []byte {
				r := &vpb.UpcomingLiveLessonNotificationRequest{
					LessonIds: lessonIDs,
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			hasError: true,
			setup: func() {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil).Once()
				tx.On("Rollback", mock.Anything).Return(nil).Once()

				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				virtualLessonRepo.On("GetVirtualLessonsByLessonIDs", mock.Anything, db, lessonIDs).Return(mockVirtualLessons, nil).Once()
				liveLessonSentNotificationRepo.On("GetLiveLessonSentNotificationCount", mock.Anything, db, lessonIDs[0], interval).Return(int32(0), nil).Once()
				liveLessonSentNotificationRepo.On("CreateLiveLessonSentNotificationRecord", mock.Anything, tx, lessonIDs[0], interval, mock.Anything).Return(nil).Once()
				lessonMemberRepo.On("GetLearnerIDsByLessonID", mock.Anything, tx, lessonIDs[0]).Return(mockMemberUserIds, nil).Once()
				userRepo.On("GetUsersByIDs", mock.Anything, db, mockMemberUserIds).Return(nil, errors.New("fail")).Once()
			},
		},
		{
			name: "error failed GetStudentParents",
			data: func() []byte {
				r := &vpb.UpcomingLiveLessonNotificationRequest{
					LessonIds: lessonIDs,
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			hasError: true,
			setup: func() {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil).Once()
				tx.On("Rollback", mock.Anything).Return(nil).Once()

				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				virtualLessonRepo.On("GetVirtualLessonsByLessonIDs", mock.Anything, db, lessonIDs).Return(mockVirtualLessons, nil).Once()
				liveLessonSentNotificationRepo.On("GetLiveLessonSentNotificationCount", mock.Anything, db, lessonIDs[0], interval).Return(int32(0), nil).Once()
				liveLessonSentNotificationRepo.On("CreateLiveLessonSentNotificationRecord", mock.Anything, tx, lessonIDs[0], interval, mock.Anything).Return(nil).Once()
				lessonMemberRepo.On("GetLearnerIDsByLessonID", mock.Anything, tx, lessonIDs[0]).Return(mockMemberUserIds, nil).Once()
				userRepo.On("GetUsersByIDs", mock.Anything, db, mockMemberUserIds).Return(mockLessonMemberUsers, nil).Once()
				studentParentRepo.On("GetStudentParents", mock.Anything, tx, mockMemberUserIds).Return(nil, errors.New("fail")).Once()
			},
		},
		{
			name: "error failed PublishContext",
			data: func() []byte {
				r := &vpb.UpcomingLiveLessonNotificationRequest{
					LessonIds: lessonIDs,
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			hasError: true,
			setup: func() {
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil).Once()
				tx.On("Rollback", mock.Anything).Return(nil).Once()

				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				virtualLessonRepo.On("GetVirtualLessonsByLessonIDs", mock.Anything, db, lessonIDs).Return(mockVirtualLessons, nil).Once()
				liveLessonSentNotificationRepo.On("GetLiveLessonSentNotificationCount", mock.Anything, db, lessonIDs[0], interval).Return(int32(0), nil).Once()
				liveLessonSentNotificationRepo.On("CreateLiveLessonSentNotificationRecord", mock.Anything, tx, lessonIDs[0], interval, mock.Anything).Return(nil).Once()
				lessonMemberRepo.On("GetLearnerIDsByLessonID", mock.Anything, tx, lessonIDs[0]).Return(mockMemberUserIds, nil).Once()
				userRepo.On("GetUsersByIDs", mock.Anything, db, mockMemberUserIds).Return(mockLessonMemberUsers, nil).Once()
				studentParentRepo.On("GetStudentParents", mock.Anything, tx, mockMemberUserIds).Return(mockStudentParent, nil).Once()
				jsm.On("PublishContext", mock.Anything, "Notification.Created", mock.Anything, mock.Anything).Return(nil, errors.New("fail"))
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup()
			l := &UpcomingLiveLessonNotificationHandler{
				Logger:                         ctxzap.Extract(ctx),
				BobDB:                          db,
				WrapperConnection:              wrapperConnection,
				JSM:                            jsm,
				VirtualLessonRepo:              virtualLessonRepo,
				LiveLessonSentNotificationRepo: liveLessonSentNotificationRepo,
				LessonMemberRepo:               lessonMemberRepo,
				StudentParentRepo:              studentParentRepo,
				UserRepo:                       userRepo,
			}

			success, err := l.Handle(ctx, tc.data)
			if tc.hasError {
				assert.Error(t, err)
				assert.Equal(t, success, false)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, success, true)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}

func TestCreateCountryParticipantMap(t *testing.T) {
	t.Parallel()

	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tcs := []struct {
		name         string
		participants []*LiveLessonParticipant
		expectedMap  map[string][]string
	}{
		{
			name: "map with multiple countries",
			participants: []*LiveLessonParticipant{
				{
					StudentID: "student-id1",
					ParentIDs: []string{
						"parent-id1",
					},
					Country: "COUNTRY_JP",
				},
				{
					StudentID: "student-id2",
					ParentIDs: []string{
						"parent-id2",
						"parent-id3",
					},
					Country: "COUNTRY_JP",
				},
				{
					StudentID: "student-id3",
					ParentIDs: []string{
						"parent-id4",
						"parent-id5",
						"parent-id6",
					},
					Country: "COUNTRY_VN",
				},
				{
					StudentID: "student-id4",
					ParentIDs: []string{},
					Country:   "COUNTRY_VN",
				},
				{
					StudentID: "student-id5",
					ParentIDs: []string{
						"parent-id7",
						"parent-id8",
					},
					Country: "COUNTRY_PH",
				},
				{
					StudentID: "student-id6",
					ParentIDs: []string{
						"parent-id9",
					},
					Country: "COUNTRY_PH",
				},
			},
			expectedMap: map[string][]string{
				"COUNTRY_JP": {
					"student-id1",
					"parent-id1",
					"student-id2",
					"parent-id2",
					"parent-id3",
				},
				"COUNTRY_VN": {
					"student-id3",
					"parent-id4",
					"parent-id5",
					"parent-id6",
					"student-id4",
				},
				"COUNTRY_PH": {
					"student-id5",
					"parent-id7",
					"parent-id8",
					"student-id6",
					"parent-id9",
				},
			},
		},
		{
			name: "map with one country",
			participants: []*LiveLessonParticipant{
				{
					StudentID: "student-id1",
					ParentIDs: []string{
						"parent-id1",
					},
					Country: "COUNTRY_JP",
				},
				{
					StudentID: "student-id2",
					ParentIDs: []string{
						"parent-id2",
						"parent-id3",
					},
					Country: "COUNTRY_JP",
				},
				{
					StudentID: "student-id3",
					ParentIDs: []string{
						"parent-id4",
						"parent-id5",
						"parent-id6",
					},
					Country: "COUNTRY_JP",
				},
				{
					StudentID: "student-id4",
					ParentIDs: []string{},
					Country:   "COUNTRY_JP",
				},
			},
			expectedMap: map[string][]string{
				"COUNTRY_JP": {
					"student-id1",
					"parent-id1",
					"student-id2",
					"parent-id2",
					"parent-id3",
					"student-id3",
					"parent-id4",
					"parent-id5",
					"parent-id6",
					"student-id4",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			l := &UpcomingLiveLessonNotificationHandler{}
			actualMap := l.createCountryParticipantMap(tc.participants)
			assert.Equal(t, actualMap, tc.expectedMap)
		})
	}
}
