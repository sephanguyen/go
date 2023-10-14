package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/bob/services/log"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	virDomain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_lessonmgmt_repo "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestLessonReaderService_GetStreamingLearners(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	lessonRepo := new(mock_repositories.MockLessonRepo)
	s := &LessonReaderServices{
		LessonRepo: lessonRepo,
		DB:         db,
	}

	testCases := []TestCase{
		{
			name:         "happy case",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &bpb.GetStreamingLearnersRequest{LessonId: "0"},
			expectedErr:  nil,
			expectedResp: &bpb.GetStreamingLearnersResponse{LearnerIds: []string{"0", "1"}},
			setup: func(ctx context.Context) {
				lessonRepo.On("GetStreamingLearners", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{"0", "1"}, nil)
			},
		},
		{
			name:        "error query",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &bpb.GetStreamingLearnersRequest{LessonId: "0"},
			expectedErr: fmt.Errorf("s.LessonStreamRepo.GetStreamingLearners: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				lessonRepo.On("GetStreamingLearners", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return([]string{}, pgx.ErrNoRows)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*bpb.GetStreamingLearnersRequest)
			resp, err := s.GetStreamingLearners(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestRetrieveLesson(t *testing.T) {
	t.Parallel()
	t.Run("empty lesson", func(t *testing.T) {
		t.Parallel()
		lessonRepo := &mock_repositories.MockLessonRepo{}
		userRepo := &mock_repositories.MockUserRepo{}

		userRepo.On("UserGroup", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.UserGroupAdmin, nil)
		lessonRepo.On("Retrieve", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, uint32(0), "", uint32(0), nil)

		svc := &LessonReaderServices{
			LessonRepo: lessonRepo,
			UserRepo:   userRepo,
		}

		resp, err := svc.RetrieveLessons(context.Background(), &bpb.RetrieveLessonsRequest{
			Paging: &cpb.Paging{Limit: 5},
		})
		assert.Nil(t, err)
		assert.Equal(t, &bpb.RetrieveLessonsResponse{}, resp)
	})

	t.Run("missing paging info", func(t *testing.T) {
		t.Parallel()
		lessonRepo := &mock_repositories.MockLessonRepo{}
		lessonRepo.On("Retrieve", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, uint32(0), "", uint32(0), nil)

		svc := &LessonReaderServices{
			LessonRepo: lessonRepo,
		}

		resp, err := svc.RetrieveLessons(context.Background(), &bpb.RetrieveLessonsRequest{})
		assert.Nil(t, resp)
		assert.Equal(t, status.Error(codes.Internal, "missing paging info"), err)
	})

	t.Run("admin success in query", func(t *testing.T) {
		t.Parallel()
		items := []*entities.Lesson{
			{
				LessonID: database.Text("id1"),
				Name:     database.Text("sid"),
			},
			{
				LessonID: database.Text("id2"),
				Name:     database.Text("sid"),
			},
		}

		userRepo := &mock_repositories.MockUserRepo{}
		lessonRepo := &mock_repositories.MockLessonRepo{}
		userRepo.On("UserGroup", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.UserGroupAdmin, nil)
		lessonRepo.On("Retrieve", mock.Anything, mock.Anything, &repositories.ListLessonArgs{
			Limit:            2,
			LessonID:         database.Text("id"),
			SchoolID:         pgtype.Int4{Status: pgtype.Null},
			Courses:          pgtype.TextArray{Status: pgtype.Null},
			StartTime:        pgtype.Timestamptz{Status: pgtype.Null},
			EndTime:          pgtype.Timestamptz{Status: pgtype.Null},
			StatusNotStarted: pgtype.Text{Status: pgtype.Null},
			StatusInProcess:  pgtype.Text{Status: pgtype.Null},
			StatusCompleted:  pgtype.Text{Status: pgtype.Null},
			KeyWord:          pgtype.Text{Status: pgtype.Null},
		}).Once().Return(items, uint32(2), "pre_id", uint32(3), nil)
		lessonRepo.On("GetTeacherIDsOfLesson", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.TextArray{}, nil)
		lessonRepo.On("GetCourseIDsOfLesson", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.TextArray{}, nil)

		svc := &LessonReaderServices{
			LessonRepo: lessonRepo,
			UserRepo:   userRepo,
		}

		resp, err := svc.RetrieveLessons(context.Background(), &bpb.RetrieveLessonsRequest{
			Paging: &cpb.Paging{
				Limit:  2,
				Offset: &cpb.Paging_OffsetString{OffsetString: "id"},
			},
		})

		assert.Nil(t, err)
		assert.Equal(t, len(items), len(resp.Items))
		assert.Equal(t, uint32(2), resp.NextPage.Limit)
		assert.Equal(t, "id2", resp.NextPage.GetOffsetString())
		assert.Equal(t, "pre_id", resp.PreviousPage.GetOffsetString())
	})

	t.Run("school admin success in query", func(t *testing.T) {
		t.Parallel()
		schoolAdmin := entities.SchoolAdmin{
			User: entities.User{
				ID: database.Text("user_id_1"),
			},
			SchoolID: database.Int4(5),
		}
		items := []*entities.Lesson{
			{
				LessonID: database.Text("id1"),
				Name:     database.Text("sid"),
			},
			{
				LessonID: database.Text("id2"),
				Name:     database.Text("sid"),
			},
		}

		userRepo := &mock_repositories.MockUserRepo{}
		schoolAdminRepo := &mock_repositories.MockSchoolAdminRepo{}
		lessonRepo := &mock_repositories.MockLessonRepo{}
		userRepo.On("UserGroup", mock.Anything, mock.Anything, mock.Anything).Once().Return(constant.UserGroupSchoolAdmin, nil)
		schoolAdminRepo.On("Get", mock.Anything, mock.Anything, mock.Anything).Once().Return(&schoolAdmin, nil)
		lessonRepo.On("Retrieve", mock.Anything, mock.Anything, &repositories.ListLessonArgs{
			Limit:            2,
			LessonID:         database.Text("id"),
			SchoolID:         database.Int4(5),
			Courses:          pgtype.TextArray{Status: pgtype.Null},
			StartTime:        pgtype.Timestamptz{Status: pgtype.Null},
			EndTime:          pgtype.Timestamptz{Status: pgtype.Null},
			StatusNotStarted: pgtype.Text{Status: pgtype.Null},
			StatusInProcess:  pgtype.Text{Status: pgtype.Null},
			StatusCompleted:  pgtype.Text{Status: pgtype.Null},
			KeyWord:          pgtype.Text{Status: pgtype.Null},
		}).Once().Return(items, uint32(2), "pre_id", uint32(3), nil)
		lessonRepo.On("GetTeacherIDsOfLesson", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.TextArray{}, nil)
		lessonRepo.On("GetCourseIDsOfLesson", mock.Anything, mock.Anything, mock.Anything).Return(pgtype.TextArray{}, nil)

		svc := &LessonReaderServices{
			LessonRepo:      lessonRepo,
			UserRepo:        userRepo,
			SchoolAdminRepo: schoolAdminRepo,
		}

		resp, err := svc.RetrieveLessons(context.Background(), &bpb.RetrieveLessonsRequest{
			Paging: &cpb.Paging{
				Limit:  2,
				Offset: &cpb.Paging_OffsetString{OffsetString: "id"},
			},
		})

		assert.Nil(t, err)
		assert.Equal(t, len(items), len(resp.Items))
		assert.Equal(t, uint32(2), resp.NextPage.Limit)
		assert.Equal(t, "id2", resp.NextPage.GetOffsetString())
		assert.Equal(t, "pre_id", resp.PreviousPage.GetOffsetString())
	})
}

func TestLiveLessonStateResponseFromLiveLessonState(t *testing.T) {
	t.Parallel()
	now := time.Now()
	creator := "creator"
	tcs := []struct {
		name  string
		state *LiveLessonState
		media *entities.Media
		res   *bpb.LiveLessonStateResponse
	}{
		{
			name: "convert with full live lesson states and media",
			state: &LiveLessonState{
				LessonID: "lesson-1",
				RoomState: &LessonRoomState{
					CurrentMaterial: &CurrentMaterial{
						MediaID:   "media-1",
						UpdatedAt: now,
						VideoState: &VideoState{
							CurrentTime: Duration(23 * time.Minute),
							PlayerState: PlayerStatePlaying,
						},
					},
					CurrentPolling: &CurrentPolling{
						Options: []*PollingOption{
							{
								Answer:    "A",
								IsCorrect: true,
							},
							{
								Answer:    "B",
								IsCorrect: false,
							},
							{
								Answer:    "C",
								IsCorrect: false,
							},
						},
						Status:    PollingStateStarted,
						CreatedAt: now,
						StoppedAt: now.Add(-2 * time.Minute),
					},
					Recording: &RecordingState{
						IsRecording: true,
						Creator:     &creator,
					},
				},
				UserStates: &UserStates{
					LearnersState: []*LearnerState{
						{
							UserID: "user-1",
							HandsUp: &UserHandsUp{
								Value:     false,
								UpdatedAt: now.Add(-2 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     false,
								UpdatedAt: now.Add(-2 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        now.Add(-2 * time.Minute),
							},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-2",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        now.Add(-20 * time.Minute),
							},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-3",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-2 * time.Hour),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-2 * time.Hour),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        now.Add(-2 * time.Hour),
							},
							Chat: &UserChat{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
					},
				},
			},
			media: &entities.Media{
				MediaID:   database.Text("media-1"),
				Name:      database.Text("media-1-name"),
				Resource:  database.Text("https://example.com/video.mp4"),
				Type:      database.Text(string(entities.MediaTypeVideo)),
				CreatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
				UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
				Comments: database.JSONB(`
								[
									{
										"comment": "hello",
										"duration": 200
									},
									{
										"comment": "hi",
										"duration": 500
									}
								]`),
			},
			res: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				CurrentMaterial: &bpb.LiveLessonState_CurrentMaterial{
					MediaId:   "media-1",
					UpdatedAt: timestamppb.New(now),
					State: &bpb.LiveLessonState_CurrentMaterial_VideoState_{
						VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
							CurrentTime: durationpb.New(23 * time.Minute),
							PlayerState: bpb.PlayerState_PLAYER_STATE_PLAYING,
						},
					},
					Data: &bpb.Media{
						MediaId:   "media-1",
						Name:      "media-1-name",
						Resource:  "https://example.com/video.mp4",
						Type:      bpb.MediaType_MEDIA_TYPE_VIDEO,
						CreatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						Comments: []*bpb.Comment{
							{
								Comment:  "hello",
								Duration: durationpb.New(200 * time.Second),
							},
							{
								Comment:  "hi",
								Duration: durationpb.New(500 * time.Second),
							},
						},
					},
				},
				CurrentPolling: &bpb.LiveLessonState_CurrentPolling{
					Options: []*bpb.LiveLessonState_PollingOption{
						{
							Answer:    "A",
							IsCorrect: true,
						},
						{
							Answer:    "B",
							IsCorrect: false,
						},
						{
							Answer:    "C",
							IsCorrect: false,
						},
					},
					Status:    bpb.PollingState_POLLING_STATE_STARTED,
					CreatedAt: timestamppb.New(now),
					StoppedAt: timestamppb.New(now.Add(-2 * time.Minute)),
				},
				UsersState: &bpb.LiveLessonStateResponse_UsersState{
					Learners: []*bpb.LiveLessonStateResponse_UsersState_LearnerState{
						{
							UserId: "user-1",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-2",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-3",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Hour)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Hour)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        timestamppb.New(now.Add(-2 * time.Hour)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
					},
				},
				Recording: &bpb.LiveLessonState_Recording{
					IsRecording: true,
					Creator:     creator,
				},
			},
		},
		{
			name: "convert without current material",
			state: &LiveLessonState{
				LessonID: "lesson-1",
				RoomState: &LessonRoomState{
					CurrentPolling: &CurrentPolling{
						Options: []*PollingOption{
							{
								Answer:    "A",
								IsCorrect: true,
							},
							{
								Answer:    "B",
								IsCorrect: false,
							},
							{
								Answer:    "C",
								IsCorrect: false,
							},
						},
						Status:    PollingStateStarted,
						CreatedAt: now,
						StoppedAt: now.Add(-2 * time.Minute),
					},
					Recording: &RecordingState{
						IsRecording: true,
						Creator:     &creator,
					},
				},
				UserStates: &UserStates{
					LearnersState: []*LearnerState{
						{
							UserID: "user-1",
							HandsUp: &UserHandsUp{
								Value:     false,
								UpdatedAt: now.Add(-2 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     false,
								UpdatedAt: now.Add(-2 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        now,
							},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-2",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        now,
							},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-3",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-2 * time.Hour),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-2 * time.Hour),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        now,
							},
							Chat: &UserChat{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
					},
				},
			},
			res: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				CurrentPolling: &bpb.LiveLessonState_CurrentPolling{
					Options: []*bpb.LiveLessonState_PollingOption{
						{
							Answer:    "A",
							IsCorrect: true,
						},
						{
							Answer:    "B",
							IsCorrect: false,
						},
						{
							Answer:    "C",
							IsCorrect: false,
						},
					},
					Status:    bpb.PollingState_POLLING_STATE_STARTED,
					CreatedAt: timestamppb.New(now),
					StoppedAt: timestamppb.New(now.Add(-2 * time.Minute)),
				},
				UsersState: &bpb.LiveLessonStateResponse_UsersState{
					Learners: []*bpb.LiveLessonStateResponse_UsersState_LearnerState{
						{
							UserId: "user-1",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        timestamppb.New(now),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-2",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        timestamppb.New(now),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-3",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Hour)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Hour)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        timestamppb.New(now),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
					},
				},
				Recording: &bpb.LiveLessonState_Recording{
					IsRecording: true,
					Creator:     creator,
				},
			},
		},
		{
			name: "convert without current polling",
			state: &LiveLessonState{
				LessonID: "lesson-1",
				RoomState: &LessonRoomState{
					CurrentMaterial: &CurrentMaterial{
						MediaID:   "media-1",
						UpdatedAt: now,
						VideoState: &VideoState{
							CurrentTime: Duration(23 * time.Minute),
							PlayerState: PlayerStatePlaying,
						},
					},
					Recording: &RecordingState{
						IsRecording: true,
						Creator:     &creator,
					},
				},
				UserStates: &UserStates{
					LearnersState: []*LearnerState{
						{
							UserID: "user-1",
							HandsUp: &UserHandsUp{
								Value:     false,
								UpdatedAt: now.Add(-2 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     false,
								UpdatedAt: now.Add(-2 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        now,
							},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-2",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        now,
							},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-3",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-2 * time.Hour),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-2 * time.Hour),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        now,
							},
							Chat: &UserChat{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
					},
				},
			},
			media: &entities.Media{
				MediaID:   database.Text("media-1"),
				Name:      database.Text("media-1-name"),
				Resource:  database.Text("https://example.com/video.mp4"),
				Type:      database.Text(string(entities.MediaTypeVideo)),
				CreatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
				UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
				Comments: database.JSONB(`
								[
									{
										"comment": "hello",
										"duration": 200
									},
									{
										"comment": "hi",
										"duration": 500
									}
								]`),
			},
			res: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				CurrentMaterial: &bpb.LiveLessonState_CurrentMaterial{
					MediaId:   "media-1",
					UpdatedAt: timestamppb.New(now),
					State: &bpb.LiveLessonState_CurrentMaterial_VideoState_{
						VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
							CurrentTime: durationpb.New(23 * time.Minute),
							PlayerState: bpb.PlayerState_PLAYER_STATE_PLAYING,
						},
					},
					Data: &bpb.Media{
						MediaId:   "media-1",
						Name:      "media-1-name",
						Resource:  "https://example.com/video.mp4",
						Type:      bpb.MediaType_MEDIA_TYPE_VIDEO,
						CreatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						Comments: []*bpb.Comment{
							{
								Comment:  "hello",
								Duration: durationpb.New(200 * time.Second),
							},
							{
								Comment:  "hi",
								Duration: durationpb.New(500 * time.Second),
							},
						},
					},
				},
				UsersState: &bpb.LiveLessonStateResponse_UsersState{
					Learners: []*bpb.LiveLessonStateResponse_UsersState_LearnerState{
						{
							UserId: "user-1",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        timestamppb.New(now),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-2",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        timestamppb.New(now),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-3",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Hour)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Hour)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        timestamppb.New(now),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
					},
				},
				Recording: &bpb.LiveLessonState_Recording{
					IsRecording: true,
					Creator:     creator,
				},
			},
		},
		{
			name: "convert without recording",
			state: &LiveLessonState{
				LessonID: "lesson-1",
				RoomState: &LessonRoomState{
					CurrentMaterial: &CurrentMaterial{
						MediaID:   "media-1",
						UpdatedAt: now,
						VideoState: &VideoState{
							CurrentTime: Duration(23 * time.Minute),
							PlayerState: PlayerStatePlaying,
						},
					},
					CurrentPolling: &CurrentPolling{
						Options: []*PollingOption{
							{
								Answer:    "A",
								IsCorrect: true,
							},
							{
								Answer:    "B",
								IsCorrect: false,
							},
							{
								Answer:    "C",
								IsCorrect: false,
							},
						},
						Status:    PollingStateStarted,
						CreatedAt: now,
						StoppedAt: now.Add(-2 * time.Minute),
					},
				},
				UserStates: &UserStates{
					LearnersState: []*LearnerState{
						{
							UserID: "user-1",
							HandsUp: &UserHandsUp{
								Value:     false,
								UpdatedAt: now.Add(-2 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     false,
								UpdatedAt: now.Add(-2 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        now.Add(-2 * time.Minute),
							},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-2",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        now.Add(-20 * time.Minute),
							},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-3",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-2 * time.Hour),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-2 * time.Hour),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        now.Add(-2 * time.Hour),
							},
							Chat: &UserChat{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
					},
				},
			},
			media: &entities.Media{
				MediaID:   database.Text("media-1"),
				Name:      database.Text("media-1-name"),
				Resource:  database.Text("https://example.com/video.mp4"),
				Type:      database.Text(string(entities.MediaTypeVideo)),
				CreatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
				UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
				Comments: database.JSONB(`
								[
									{
										"comment": "hello",
										"duration": 200
									},
									{
										"comment": "hi",
										"duration": 500
									}
								]`),
			},
			res: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				CurrentMaterial: &bpb.LiveLessonState_CurrentMaterial{
					MediaId:   "media-1",
					UpdatedAt: timestamppb.New(now),
					State: &bpb.LiveLessonState_CurrentMaterial_VideoState_{
						VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
							CurrentTime: durationpb.New(23 * time.Minute),
							PlayerState: bpb.PlayerState_PLAYER_STATE_PLAYING,
						},
					},
					Data: &bpb.Media{
						MediaId:   "media-1",
						Name:      "media-1-name",
						Resource:  "https://example.com/video.mp4",
						Type:      bpb.MediaType_MEDIA_TYPE_VIDEO,
						CreatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						Comments: []*bpb.Comment{
							{
								Comment:  "hello",
								Duration: durationpb.New(200 * time.Second),
							},
							{
								Comment:  "hi",
								Duration: durationpb.New(500 * time.Second),
							},
						},
					},
				},
				CurrentPolling: &bpb.LiveLessonState_CurrentPolling{
					Options: []*bpb.LiveLessonState_PollingOption{
						{
							Answer:    "A",
							IsCorrect: true,
						},
						{
							Answer:    "B",
							IsCorrect: false,
						},
						{
							Answer:    "C",
							IsCorrect: false,
						},
					},
					Status:    bpb.PollingState_POLLING_STATE_STARTED,
					CreatedAt: timestamppb.New(now),
					StoppedAt: timestamppb.New(now.Add(-2 * time.Minute)),
				},
				UsersState: &bpb.LiveLessonStateResponse_UsersState{
					Learners: []*bpb.LiveLessonStateResponse_UsersState_LearnerState{
						{
							UserId: "user-1",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-2",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-3",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Hour)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Hour)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        timestamppb.New(now.Add(-2 * time.Hour)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
					},
				},
			},
		},
		{
			name: "convert without any room's states",
			state: &LiveLessonState{
				LessonID: "lesson-1",
				UserStates: &UserStates{
					LearnersState: []*LearnerState{
						{
							UserID: "user-1",
							HandsUp: &UserHandsUp{
								Value:     false,
								UpdatedAt: now.Add(-2 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     false,
								UpdatedAt: now.Add(-2 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        now,
							},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-2",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        now,
							},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-3",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-2 * time.Hour),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-2 * time.Hour),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        now,
							},
							Chat: &UserChat{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
					},
				},
			},
			res: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				UsersState: &bpb.LiveLessonStateResponse_UsersState{
					Learners: []*bpb.LiveLessonStateResponse_UsersState_LearnerState{
						{
							UserId: "user-1",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        timestamppb.New(now),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-2",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        timestamppb.New(now),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-3",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Hour)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Hour)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        timestamppb.New(now),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
					},
				},
			},
		},
		{
			name: "convert without learner's states",
			state: &LiveLessonState{
				LessonID: "lesson-1",
				RoomState: &LessonRoomState{
					CurrentMaterial: &CurrentMaterial{
						MediaID:   "media-1",
						UpdatedAt: now,
						VideoState: &VideoState{
							CurrentTime: Duration(23 * time.Minute),
							PlayerState: PlayerStatePlaying,
						},
					},
				},
				UserStates: &UserStates{},
			},
			media: &entities.Media{
				MediaID:   database.Text("media-1"),
				Name:      database.Text("media-1-name"),
				Resource:  database.Text("https://example.com/video.mp4"),
				Type:      database.Text(string(entities.MediaTypeVideo)),
				CreatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
				UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
			},
			res: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				CurrentMaterial: &bpb.LiveLessonState_CurrentMaterial{
					MediaId:   "media-1",
					UpdatedAt: timestamppb.New(now),
					State: &bpb.LiveLessonState_CurrentMaterial_VideoState_{
						VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
							CurrentTime: durationpb.New(23 * time.Minute),
							PlayerState: bpb.PlayerState_PLAYER_STATE_PLAYING,
						},
					},
					Data: &bpb.Media{
						MediaId:   "media-1",
						Name:      "media-1-name",
						Resource:  "https://example.com/video.mp4",
						Type:      bpb.MediaType_MEDIA_TYPE_VIDEO,
						CreatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
					},
				},
				UsersState: &bpb.LiveLessonStateResponse_UsersState{},
			},
		},
		{
			name: "convert with current material is pdf",
			state: &LiveLessonState{
				LessonID: "lesson-1",
				RoomState: &LessonRoomState{
					CurrentMaterial: &CurrentMaterial{
						MediaID:   "media-1",
						UpdatedAt: now,
					},
				},
				UserStates: &UserStates{
					LearnersState: []*LearnerState{
						{
							UserID: "user-1",
							HandsUp: &UserHandsUp{
								Value:     false,
								UpdatedAt: now.Add(-2 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     false,
								UpdatedAt: now.Add(-2 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        now,
							},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-2",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        now,
							},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-3",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-2 * time.Hour),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-2 * time.Hour),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        now,
							},
							Chat: &UserChat{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
					},
				},
			},
			media: &entities.Media{
				MediaID:   database.Text("media-1"),
				Name:      database.Text("media-1-name"),
				Resource:  database.Text("https://example.com/slide.pdf"),
				Type:      database.Text(string(entities.MediaTypePDF)),
				CreatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
				UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
			},
			res: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				CurrentMaterial: &bpb.LiveLessonState_CurrentMaterial{
					MediaId:   "media-1",
					UpdatedAt: timestamppb.New(now),
					Data: &bpb.Media{
						MediaId:   "media-1",
						Name:      "media-1-name",
						Resource:  "https://example.com/slide.pdf",
						Type:      bpb.MediaType_MEDIA_TYPE_PDF,
						CreatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
					},
				},
				UsersState: &bpb.LiveLessonStateResponse_UsersState{
					Learners: []*bpb.LiveLessonStateResponse_UsersState_LearnerState{
						{
							UserId: "user-1",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        timestamppb.New(now),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-2",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        timestamppb.New(now),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-3",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Hour)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Hour)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{},
								UpdatedAt:        timestamppb.New(now),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := liveLessonStateResponseFromLiveLessonState(tc.state, tc.media)
			require.NoError(t, err)
			assert.EqualValues(t, tc.res, actual)
		})
	}
}

func TestLessonReaderServices_GetLiveLessonState(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now().UTC()

	// mock structs
	db := &mock_database.Ext{}
	lessonRepo := &mock_repositories.MockLessonRepo{}
	lessonMemberRepo := &mock_repositories.MockLessonMemberRepo{}
	userRepo := &mock_repositories.MockUserRepo{}
	mediaRepo := &mock_repositories.MockMediaRepo{}
	logRepo := new(mock_repositories.MockVirtualClassroomLogRepo)
	lessonRoomStateRepo := &mock_lessonmgmt_repo.MockLessonRoomStateRepo{}

	wDefault := new(virDomain.WhiteboardZoomState).SetDefault()
	whiteboardZoomStateDefaultRes := toWhiteboardZoomStateBp(wDefault)

	tcs := []struct {
		name        string
		reqUserID   string
		req         *bpb.LiveLessonStateRequest
		setup       func(context.Context)
		expectedRes *bpb.LiveLessonStateResponse
		hasError    bool
	}{
		{
			name:      "teacher get live lesson state with full data",
			reqUserID: "user-5",
			req: &bpb.LiveLessonStateRequest{
				Id: "lesson-1",
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
					}, nil).
					Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-1",
							"user-2",
							"user-3",
						}),
						nil,
					).
					Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-5",
							"user-6",
						}),
						nil,
					).
					Once()
				lessonMemberRepo.
					On("GetLessonMemberStates", ctx, db, database.Text("lesson-1")).
					Return(
						entities.LessonMemberStates{
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue: database.Bool(false),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue: database.Bool(false),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-1"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt:        database.Timestamptz(now.Add(-2 * time.Minute)),
								StringArrayValue: database.TextArray([]string{"A"}),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-2"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								StringArrayValue: database.TextArray([]string{"B"}),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-3"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
								StringArrayValue: database.TextArray([]string{"B", "C"}),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(false),
							},
						},
						nil,
					).
					Once()
				mediaRepo.
					On("RetrieveByIDs", ctx, db, database.TextArray([]string{"media-1"})).
					Return(
						[]*entities.Media{
							{
								MediaID:   database.Text("media-1"),
								Name:      database.Text("media-1-name"),
								Resource:  database.Text("https://example.com/video.mp4"),
								Type:      database.Text(string(entities.MediaTypeVideo)),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								Comments: database.JSONB(`
								[
									{
										"comment": "hello",
										"duration": 200
									},
									{
										"comment": "hi",
										"duration": 500
									}
								]`),
							},
						},
						nil,
					).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesGettingRoomState,
				).
					Return(nil).
					Once()
				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, database.Text("lesson-1")).
					Return(&domain.LessonRoomState{
						LessonID:        "lesson-1",
						SpotlightedUser: "user-1",
						WhiteboardZoomState: &virDomain.WhiteboardZoomState{
							PdfScaleRatio: 23.32,
							CenterX:       243.5,
							CenterY:       -432.034,
							PdfWidth:      234.43,
							PdfHeight:     -0.33424,
						},
						Recording: &virDomain.CompositeRecordingState{
							ResourceID:  "resource-id",
							SID:         "s-id",
							UID:         123342,
							IsRecording: true,
							Creator:     "user-id-1",
						},
						CurrentMaterial: &virDomain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
							VideoState: &virDomain.VideoState{
								CurrentTime: virDomain.Duration(23 * time.Minute),
								PlayerState: virDomain.PlayerStatePlaying,
							},
						},
					}, nil).Once()
			},
			expectedRes: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				CurrentMaterial: &bpb.LiveLessonState_CurrentMaterial{
					MediaId:   "media-1",
					UpdatedAt: timestamppb.New(now),
					State: &bpb.LiveLessonState_CurrentMaterial_VideoState_{
						VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
							CurrentTime: durationpb.New(23 * time.Minute),
							PlayerState: bpb.PlayerState_PLAYER_STATE_PLAYING,
						},
					},
					Data: &bpb.Media{
						MediaId:   "media-1",
						Name:      "media-1-name",
						Resource:  "https://example.com/video.mp4",
						Type:      bpb.MediaType_MEDIA_TYPE_VIDEO,
						CreatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						Comments: []*bpb.Comment{
							{
								Comment:  "hello",
								Duration: durationpb.New(200 * time.Second),
							},
							{
								Comment:  "hi",
								Duration: durationpb.New(500 * time.Second),
							},
						},
					},
				},
				UsersState: &bpb.LiveLessonStateResponse_UsersState{
					Learners: []*bpb.LiveLessonStateResponse_UsersState_LearnerState{
						{
							UserId: "user-1",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-2",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"B"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-3",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Hour)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Hour)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"B", "C"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Hour)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
					},
				},
				Recording: &bpb.LiveLessonState_Recording{
					IsRecording: true,
					Creator:     "user-id-1",
				},
				Spotlight: &bpb.LiveLessonState_Spotlight{
					IsSpotlight: true,
					UserId:      "user-1",
				},
				WhiteboardZoomState: &bpb.LiveLessonState_WhiteboardZoomState{
					PdfScaleRatio: 23.32,
					CenterX:       243.5,
					CenterY:       -432.034,
					PdfWidth:      234.43,
					PdfHeight:     -0.33424,
				},
			},
		},
		{
			name:      "teacher get live lesson state with full data and audio state",
			reqUserID: "user-5",
			req: &bpb.LiveLessonStateRequest{
				Id: "lesson-1",
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
					}, nil).
					Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-1",
							"user-2",
							"user-3",
						}),
						nil,
					).
					Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-5",
							"user-6",
						}),
						nil,
					).
					Once()
				lessonMemberRepo.
					On("GetLessonMemberStates", ctx, db, database.Text("lesson-1")).
					Return(
						entities.LessonMemberStates{
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue: database.Bool(false),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue: database.Bool(false),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-1"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt:        database.Timestamptz(now.Add(-2 * time.Minute)),
								StringArrayValue: database.TextArray([]string{"A"}),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-2"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								StringArrayValue: database.TextArray([]string{"B"}),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-3"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
								StringArrayValue: database.TextArray([]string{"B", "C"}),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(false),
							},
						},
						nil,
					).
					Once()
				mediaRepo.
					On("RetrieveByIDs", ctx, db, database.TextArray([]string{"media-1"})).
					Return(
						[]*entities.Media{
							{
								MediaID:   database.Text("media-1"),
								Name:      database.Text("media-1-name"),
								Resource:  database.Text("https://example.com/video.mp4"),
								Type:      database.Text(string(entities.MediaTypeVideo)),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								Comments: database.JSONB(`
								[
									{
										"comment": "hello",
										"duration": 200
									},
									{
										"comment": "hi",
										"duration": 500
									}
								]`),
							},
						},
						nil,
					).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesGettingRoomState,
				).
					Return(nil).
					Once()
				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, database.Text("lesson-1")).
					Return(&domain.LessonRoomState{
						LessonID:        "lesson-1",
						SpotlightedUser: "user-1",
						WhiteboardZoomState: &virDomain.WhiteboardZoomState{
							PdfScaleRatio: 23.32,
							CenterX:       243.5,
							CenterY:       -432.034,
							PdfWidth:      234.43,
							PdfHeight:     -0.33424,
						},
						Recording: &virDomain.CompositeRecordingState{
							ResourceID:  "resource-id",
							SID:         "s-id",
							UID:         123342,
							IsRecording: true,
							Creator:     "user-id-1",
						},
						CurrentMaterial: &virDomain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
							AudioState: &virDomain.AudioState{
								CurrentTime: virDomain.Duration(23 * time.Minute),
								PlayerState: virDomain.PlayerStatePlaying,
							},
						},
					}, nil).Once()
			},
			expectedRes: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				CurrentMaterial: &bpb.LiveLessonState_CurrentMaterial{
					MediaId:   "media-1",
					UpdatedAt: timestamppb.New(now),
					State: &bpb.LiveLessonState_CurrentMaterial_AudioState_{
						AudioState: &bpb.LiveLessonState_CurrentMaterial_AudioState{
							CurrentTime: durationpb.New(23 * time.Minute),
							PlayerState: bpb.PlayerState_PLAYER_STATE_PLAYING,
						},
					},
					Data: &bpb.Media{
						MediaId:   "media-1",
						Name:      "media-1-name",
						Resource:  "https://example.com/video.mp4",
						Type:      bpb.MediaType_MEDIA_TYPE_VIDEO,
						CreatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						Comments: []*bpb.Comment{
							{
								Comment:  "hello",
								Duration: durationpb.New(200 * time.Second),
							},
							{
								Comment:  "hi",
								Duration: durationpb.New(500 * time.Second),
							},
						},
					},
				},
				UsersState: &bpb.LiveLessonStateResponse_UsersState{
					Learners: []*bpb.LiveLessonStateResponse_UsersState_LearnerState{
						{
							UserId: "user-1",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-2",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"B"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-3",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Hour)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Hour)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"B", "C"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Hour)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
					},
				},
				Recording: &bpb.LiveLessonState_Recording{
					IsRecording: true,
					Creator:     "user-id-1",
				},
				Spotlight: &bpb.LiveLessonState_Spotlight{
					IsSpotlight: true,
					UserId:      "user-1",
				},
				WhiteboardZoomState: &bpb.LiveLessonState_WhiteboardZoomState{
					PdfScaleRatio: 23.32,
					CenterX:       243.5,
					CenterY:       -432.034,
					PdfWidth:      234.43,
					PdfHeight:     -0.33424,
				},
			},
		},
		{
			name:      "teacher get live lesson state with current material is pdf",
			reqUserID: "user-5",
			req: &bpb.LiveLessonStateRequest{
				Id: "lesson-1",
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
					}, nil).
					Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-1",
							"user-2",
							"user-3",
						}),
						nil,
					).
					Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-5",
							"user-6",
						}),
						nil,
					).
					Once()
				lessonMemberRepo.
					On("GetLessonMemberStates", ctx, db, database.Text("lesson-1")).
					Return(
						entities.LessonMemberStates{
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue: database.Bool(false),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue: database.Bool(false),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-1"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt:        database.Timestamptz(now.Add(-2 * time.Minute)),
								StringArrayValue: database.TextArray([]string{"A"}),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-2"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								StringArrayValue: database.TextArray([]string{"B"}),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-3"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
								StringArrayValue: database.TextArray([]string{"B", "C"}),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(false),
							},
						},
						nil,
					).
					Once()
				mediaRepo.
					On("RetrieveByIDs", ctx, db, database.TextArray([]string{"media-1"})).
					Return(
						[]*entities.Media{
							{
								MediaID:   database.Text("media-1"),
								Name:      database.Text("media-1-name"),
								Resource:  database.Text("https://example.com/slide.pdf"),
								Type:      database.Text(string(entities.MediaTypeVideo)),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
							},
						},
						nil,
					).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesGettingRoomState,
				).
					Return(nil).
					Once()
				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, database.Text("lesson-1")).
					Return(&domain.LessonRoomState{
						LessonID:            "lesson-1",
						SpotlightedUser:     "user-1",
						WhiteboardZoomState: wDefault,
						Recording: &virDomain.CompositeRecordingState{
							ResourceID:  "resource-id",
							SID:         "s-id",
							UID:         123342,
							IsRecording: false,
							Creator:     "user-id-1",
						},
						CurrentMaterial: &virDomain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
						},
					}, nil).Once()
			},
			expectedRes: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				CurrentMaterial: &bpb.LiveLessonState_CurrentMaterial{
					MediaId:   "media-1",
					UpdatedAt: timestamppb.New(now),
					Data: &bpb.Media{
						MediaId:   "media-1",
						Name:      "media-1-name",
						Resource:  "https://example.com/slide.pdf",
						Type:      bpb.MediaType_MEDIA_TYPE_VIDEO,
						CreatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
					},
				},
				UsersState: &bpb.LiveLessonStateResponse_UsersState{
					Learners: []*bpb.LiveLessonStateResponse_UsersState_LearnerState{
						{
							UserId: "user-1",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-2",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"B"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-3",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Hour)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Hour)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"B", "C"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Hour)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
					},
				},
				Spotlight: &bpb.LiveLessonState_Spotlight{
					IsSpotlight: true,
					UserId:      "user-1",
				},
				WhiteboardZoomState: whiteboardZoomStateDefaultRes,
				Recording: &bpb.LiveLessonState_Recording{
					IsRecording: false,
					Creator:     "user-id-1",
				},
			},
		},
		{
			name:      "student get live lesson state with full data",
			reqUserID: "user-2",
			req: &bpb.LiveLessonStateRequest{
				Id: "lesson-1",
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
					}, nil).
					Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-1",
							"user-2",
							"user-3",
						}),
						nil,
					).
					Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-5",
							"user-6",
						}),
						nil,
					).
					Once()
				lessonMemberRepo.
					On("GetLessonMemberStates", ctx, db, database.Text("lesson-1")).
					Return(
						entities.LessonMemberStates{
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue: database.Bool(false),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue: database.Bool(false),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-1"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt:        database.Timestamptz(now.Add(-2 * time.Minute)),
								StringArrayValue: database.TextArray([]string{"A"}),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-2"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								StringArrayValue: database.TextArray([]string{"B"}),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-3"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
								StringArrayValue: database.TextArray([]string{"B", "C"}),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(false),
							},
						},
						nil,
					).
					Once()
				mediaRepo.
					On("RetrieveByIDs", ctx, db, database.TextArray([]string{"media-1"})).
					Return(
						[]*entities.Media{
							{
								MediaID:   database.Text("media-1"),
								Name:      database.Text("media-1-name"),
								Resource:  database.Text("https://example.com/video.mp4"),
								Type:      database.Text(string(entities.MediaTypeVideo)),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								Comments: database.JSONB(`
								[
									{
										"comment": "hello",
										"duration": 200
									},
									{
										"comment": "hi",
										"duration": 500
									}
								]`),
							},
						},
						nil,
					).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesGettingRoomState,
				).
					Return(nil).
					Once()
				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, database.Text("lesson-1")).
					Return(&domain.LessonRoomState{
						LessonID:        "lesson-1",
						SpotlightedUser: "user-1",
						WhiteboardZoomState: &virDomain.WhiteboardZoomState{
							PdfScaleRatio: 23.32,
							CenterX:       243.5,
							CenterY:       -432.034,
							PdfWidth:      234.43,
							PdfHeight:     -0.33424,
						},
						Recording: &virDomain.CompositeRecordingState{},
						CurrentMaterial: &virDomain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
							VideoState: &virDomain.VideoState{
								CurrentTime: virDomain.Duration(23 * time.Minute),
								PlayerState: virDomain.PlayerStatePlaying,
							},
						},
					}, nil).Once()
			},
			expectedRes: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				CurrentMaterial: &bpb.LiveLessonState_CurrentMaterial{
					MediaId:   "media-1",
					UpdatedAt: timestamppb.New(now),
					State: &bpb.LiveLessonState_CurrentMaterial_VideoState_{
						VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
							CurrentTime: durationpb.New(23 * time.Minute),
							PlayerState: bpb.PlayerState_PLAYER_STATE_PLAYING,
						},
					},
					Data: &bpb.Media{
						MediaId:   "media-1",
						Name:      "media-1-name",
						Resource:  "https://example.com/video.mp4",
						Type:      bpb.MediaType_MEDIA_TYPE_VIDEO,
						CreatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						Comments: []*bpb.Comment{
							{
								Comment:  "hello",
								Duration: durationpb.New(200 * time.Second),
							},
							{
								Comment:  "hi",
								Duration: durationpb.New(500 * time.Second),
							},
						},
					},
				},
				UsersState: &bpb.LiveLessonStateResponse_UsersState{
					Learners: []*bpb.LiveLessonStateResponse_UsersState_LearnerState{
						{
							UserId: "user-1",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-2",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"B"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-3",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Hour)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Hour)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"B", "C"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Hour)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
					},
				},
				Spotlight: &bpb.LiveLessonState_Spotlight{
					IsSpotlight: true,
					UserId:      "user-1",
				},
				WhiteboardZoomState: &bpb.LiveLessonState_WhiteboardZoomState{
					PdfScaleRatio: 23.32,
					CenterX:       243.5,
					CenterY:       -432.034,
					PdfWidth:      234.43,
					PdfHeight:     -0.33424,
				},
				Recording: &bpb.LiveLessonState_Recording{},
			},
		},
		{
			name:      "teacher who not belong to lesson get live lesson state with full data",
			reqUserID: "user-7",
			req: &bpb.LiveLessonStateRequest{
				Id: "lesson-1",
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
					}, nil).
					Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-1",
							"user-2",
							"user-3",
						}),
						nil,
					).
					Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-5",
							"user-6",
						}),
						nil,
					).
					Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("user-7")).
					Return(entities.UserGroupTeacher, nil).
					Once()
				lessonMemberRepo.
					On("GetLessonMemberStates", ctx, db, database.Text("lesson-1")).
					Return(
						entities.LessonMemberStates{
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue: database.Bool(false),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue: database.Bool(false),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-1"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt:        database.Timestamptz(now.Add(-2 * time.Minute)),
								StringArrayValue: database.TextArray([]string{"A"}),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-2"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								StringArrayValue: database.TextArray([]string{"B"}),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-3"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
								StringArrayValue: database.TextArray([]string{"B", "C"}),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(false),
							},
						},
						nil,
					).
					Once()
				mediaRepo.
					On("RetrieveByIDs", ctx, db, database.TextArray([]string{"media-1"})).
					Return(
						[]*entities.Media{
							{
								MediaID:   database.Text("media-1"),
								Name:      database.Text("media-1-name"),
								Resource:  database.Text("https://example.com/video.mp4"),
								Type:      database.Text(string(entities.MediaTypeVideo)),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								Comments: database.JSONB(`
								[
									{
										"comment": "hello",
										"duration": 200
									},
									{
										"comment": "hi",
										"duration": 500
									}
								]`),
							},
						},
						nil,
					).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesGettingRoomState,
				).
					Return(nil).
					Once()
				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, database.Text("lesson-1")).
					Return(&domain.LessonRoomState{
						LessonID:        "lesson-1",
						SpotlightedUser: "user-1",
						WhiteboardZoomState: &virDomain.WhiteboardZoomState{
							PdfScaleRatio: 23.32,
							CenterX:       243.5,
							CenterY:       -432.034,
							PdfWidth:      234.43,
							PdfHeight:     -0.33424,
						},
						Recording: &virDomain.CompositeRecordingState{},
						CurrentMaterial: &virDomain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
							VideoState: &virDomain.VideoState{
								CurrentTime: virDomain.Duration(23 * time.Minute),
								PlayerState: virDomain.PlayerStatePlaying,
							},
						},
					}, nil).Once()
			},
			expectedRes: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				CurrentMaterial: &bpb.LiveLessonState_CurrentMaterial{
					MediaId:   "media-1",
					UpdatedAt: timestamppb.New(now),
					State: &bpb.LiveLessonState_CurrentMaterial_VideoState_{
						VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
							CurrentTime: durationpb.New(23 * time.Minute),
							PlayerState: bpb.PlayerState_PLAYER_STATE_PLAYING,
						},
					},
					Data: &bpb.Media{
						MediaId:   "media-1",
						Name:      "media-1-name",
						Resource:  "https://example.com/video.mp4",
						Type:      bpb.MediaType_MEDIA_TYPE_VIDEO,
						CreatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						Comments: []*bpb.Comment{
							{
								Comment:  "hello",
								Duration: durationpb.New(200 * time.Second),
							},
							{
								Comment:  "hi",
								Duration: durationpb.New(500 * time.Second),
							},
						},
					},
				},
				UsersState: &bpb.LiveLessonStateResponse_UsersState{
					Learners: []*bpb.LiveLessonStateResponse_UsersState_LearnerState{
						{
							UserId: "user-1",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-2",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"B"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-3",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Hour)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Hour)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"B", "C"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Hour)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
					},
				},
				Spotlight: &bpb.LiveLessonState_Spotlight{
					IsSpotlight: true,
					UserId:      "user-1",
				},
				WhiteboardZoomState: &bpb.LiveLessonState_WhiteboardZoomState{
					PdfScaleRatio: 23.32,
					CenterX:       243.5,
					CenterY:       -432.034,
					PdfWidth:      234.43,
					PdfHeight:     -0.33424,
				},
				Recording: &bpb.LiveLessonState_Recording{},
			},
		},
		{
			name:      "school admin get live lesson state with full data",
			reqUserID: "user-7",
			req: &bpb.LiveLessonStateRequest{
				Id: "lesson-1",
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
					}, nil).
					Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-1",
							"user-2",
							"user-3",
						}),
						nil,
					).
					Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-5",
							"user-6",
						}),
						nil,
					).
					Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("user-7")).
					Return(entities.UserGroupSchoolAdmin, nil).
					Once()
				lessonMemberRepo.
					On("GetLessonMemberStates", ctx, db, database.Text("lesson-1")).
					Return(
						entities.LessonMemberStates{
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue: database.Bool(false),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue: database.Bool(false),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-1"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt:        database.Timestamptz(now.Add(-2 * time.Minute)),
								StringArrayValue: database.TextArray([]string{"A"}),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-2"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								StringArrayValue: database.TextArray([]string{"B"}),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-3"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
								StringArrayValue: database.TextArray([]string{"B", "C"}),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(false),
							},
						},
						nil,
					).
					Once()
				mediaRepo.
					On("RetrieveByIDs", ctx, db, database.TextArray([]string{"media-1"})).
					Return(
						[]*entities.Media{
							{
								MediaID:   database.Text("media-1"),
								Name:      database.Text("media-1-name"),
								Resource:  database.Text("https://example.com/video.mp4"),
								Type:      database.Text(string(entities.MediaTypeVideo)),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								Comments: database.JSONB(`
								[
									{
										"comment": "hello",
										"duration": 200
									},
									{
										"comment": "hi",
										"duration": 500
									}
								]`),
							},
						},
						nil,
					).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesGettingRoomState,
				).
					Return(nil).
					Once()
				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, database.Text("lesson-1")).
					Return(&domain.LessonRoomState{
						LessonID:        "lesson-1",
						SpotlightedUser: "",
						WhiteboardZoomState: &virDomain.WhiteboardZoomState{
							PdfScaleRatio: 23.32,
							CenterX:       243.5,
							CenterY:       -432.034,
							PdfWidth:      234.43,
							PdfHeight:     -0.33424,
						},
						Recording: &virDomain.CompositeRecordingState{},
						CurrentMaterial: &virDomain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
							VideoState: &virDomain.VideoState{
								CurrentTime: virDomain.Duration(23 * time.Minute),
								PlayerState: virDomain.PlayerStatePlaying,
							},
						},
					}, nil).Once()
			},
			expectedRes: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				CurrentMaterial: &bpb.LiveLessonState_CurrentMaterial{
					MediaId:   "media-1",
					UpdatedAt: timestamppb.New(now),
					State: &bpb.LiveLessonState_CurrentMaterial_VideoState_{
						VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
							CurrentTime: durationpb.New(23 * time.Minute),
							PlayerState: bpb.PlayerState_PLAYER_STATE_PLAYING,
						},
					},
					Data: &bpb.Media{
						MediaId:   "media-1",
						Name:      "media-1-name",
						Resource:  "https://example.com/video.mp4",
						Type:      bpb.MediaType_MEDIA_TYPE_VIDEO,
						CreatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						Comments: []*bpb.Comment{
							{
								Comment:  "hello",
								Duration: durationpb.New(200 * time.Second),
							},
							{
								Comment:  "hi",
								Duration: durationpb.New(500 * time.Second),
							},
						},
					},
				},
				UsersState: &bpb.LiveLessonStateResponse_UsersState{
					Learners: []*bpb.LiveLessonStateResponse_UsersState_LearnerState{
						{
							UserId: "user-1",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-2",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"B"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-3",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Hour)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Hour)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"B", "C"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Hour)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
					},
				},
				Spotlight: &bpb.LiveLessonState_Spotlight{},
				WhiteboardZoomState: &bpb.LiveLessonState_WhiteboardZoomState{
					PdfScaleRatio: 23.32,
					CenterX:       243.5,
					CenterY:       -432.034,
					PdfWidth:      234.43,
					PdfHeight:     -0.33424,
				},
				Recording: &bpb.LiveLessonState_Recording{},
			},
		},
		{
			name:      "learner who not belong to lesson get live lesson state with full data",
			reqUserID: "user-7",
			req: &bpb.LiveLessonStateRequest{
				Id: "lesson-1",
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
					}, nil).
					Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-1",
							"user-2",
							"user-3",
						}),
						nil,
					).
					Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-5",
							"user-6",
						}),
						nil,
					).
					Once()
				userRepo.
					On("UserGroup", ctx, db, database.Text("user-7")).
					Return(entities.UserGroupStudent, nil).
					Once()
			},
			hasError: true,
		},
		{
			name:      "teacher get live lesson state without room's state",
			reqUserID: "user-5",
			req: &bpb.LiveLessonStateRequest{
				Id: "lesson-1",
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{LessonID: database.Text("lesson-1")}, nil).
					Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-1",
							"user-2",
							"user-3",
						}),
						nil,
					).
					Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-5",
							"user-6",
						}),
						nil,
					).
					Once()
				lessonMemberRepo.
					On("GetLessonMemberStates", ctx, db, database.Text("lesson-1")).
					Return(
						entities.LessonMemberStates{
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue: database.Bool(false),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								BoolValue: database.Bool(false),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-1"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt:        database.Timestamptz(now.Add(-2 * time.Minute)),
								StringArrayValue: database.TextArray([]string{"A"}),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-2"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-2 * time.Hour)),
								UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Minute)),
								StringArrayValue: database.TextArray([]string{"B"}),
							},
							{
								LessonID:         database.Text("lesson-1"),
								UserID:           database.Text("user-3"),
								StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
								CreatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
								UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
								StringArrayValue: database.TextArray([]string{"B", "C"}),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-1"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-2"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(true),
							},
							{
								LessonID:  database.Text("lesson-1"),
								UserID:    database.Text("user-3"),
								StateType: database.Text(string(LearnerStateTypeChat)),
								CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
								BoolValue: database.Bool(false),
							},
						},
						nil,
					).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesGettingRoomState,
				).
					Return(nil).
					Once()
				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, database.Text("lesson-1")).
					Return(&domain.LessonRoomState{
						LessonID:            "lesson-1",
						SpotlightedUser:     "",
						WhiteboardZoomState: wDefault,
					}, nil).Once()
			},
			expectedRes: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				UsersState: &bpb.LiveLessonStateResponse_UsersState{
					Learners: []*bpb.LiveLessonStateResponse_UsersState_LearnerState{
						{
							UserId: "user-1",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        timestamppb.New(now.Add(-2 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-2",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"B"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Minute)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
						{
							UserId: "user-3",
							HandsUp: &bpb.LiveLessonState_HandsUp{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Hour)),
							},
							Annotation: &bpb.LiveLessonState_Annotation{
								Value:     true,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Hour)),
							},
							PollingAnswer: &bpb.LiveLessonState_PollingAnswer{
								StringArrayValue: []string{"B", "C"},
								UpdatedAt:        timestamppb.New(now.Add(-20 * time.Hour)),
							},
							Chat: &bpb.LiveLessonState_Chat{
								Value:     false,
								UpdatedAt: timestamppb.New(now.Add(-20 * time.Minute)),
							},
						},
					},
				},
				Spotlight:           &bpb.LiveLessonState_Spotlight{},
				WhiteboardZoomState: whiteboardZoomStateDefaultRes,
			},
		},
		{
			name:      "teacher get live lesson state without any learner's state",
			reqUserID: "user-5",
			req: &bpb.LiveLessonStateRequest{
				Id: "lesson-1",
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
					}, nil).
					Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-1",
							"user-2",
							"user-3",
						}),
						nil,
					).
					Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-5",
							"user-6",
						}),
						nil,
					).
					Once()
				lessonMemberRepo.
					On("GetLessonMemberStates", ctx, db, database.Text("lesson-1")).
					Return(
						entities.LessonMemberStates{},
						nil,
					).
					Once()
				mediaRepo.
					On("RetrieveByIDs", ctx, db, database.TextArray([]string{"media-1"})).
					Return(
						[]*entities.Media{
							{
								MediaID:   database.Text("media-1"),
								Name:      database.Text("media-1-name"),
								Resource:  database.Text("https://example.com/video.mp4"),
								Type:      database.Text(string(entities.MediaTypeVideo)),
								CreatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								UpdatedAt: database.Timestamptz(now.Add(-2 * time.Minute)),
								Comments: database.JSONB(`
								[
									{
										"comment": "hello",
										"duration": 200
									},
									{
										"comment": "hi",
										"duration": 500
									}
								]`),
							},
						},
						nil,
					).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesGettingRoomState,
				).
					Return(nil).
					Once()
				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, database.Text("lesson-1")).
					Return(&domain.LessonRoomState{
						LessonID:            "lesson-1",
						SpotlightedUser:     "",
						WhiteboardZoomState: wDefault,
						CurrentMaterial: &virDomain.CurrentMaterial{
							MediaID:   "media-1",
							UpdatedAt: now,
							VideoState: &virDomain.VideoState{
								CurrentTime: virDomain.Duration(23 * time.Minute),
								PlayerState: virDomain.PlayerStatePlaying,
							},
						},
					}, nil).Once()
			},
			expectedRes: &bpb.LiveLessonStateResponse{
				Id: "lesson-1",
				CurrentMaterial: &bpb.LiveLessonState_CurrentMaterial{
					MediaId:   "media-1",
					UpdatedAt: timestamppb.New(now),
					State: &bpb.LiveLessonState_CurrentMaterial_VideoState_{
						VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
							CurrentTime: durationpb.New(23 * time.Minute),
							PlayerState: bpb.PlayerState_PLAYER_STATE_PLAYING,
						},
					},
					Data: &bpb.Media{
						MediaId:   "media-1",
						Name:      "media-1-name",
						Resource:  "https://example.com/video.mp4",
						Type:      bpb.MediaType_MEDIA_TYPE_VIDEO,
						CreatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						UpdatedAt: timestamppb.New(now.Add(-2 * time.Minute)),
						Comments: []*bpb.Comment{
							{
								Comment:  "hello",
								Duration: durationpb.New(200 * time.Second),
							},
							{
								Comment:  "hi",
								Duration: durationpb.New(500 * time.Second),
							},
						},
					},
				},
				UsersState:          &bpb.LiveLessonStateResponse_UsersState{},
				Spotlight:           &bpb.LiveLessonState_Spotlight{},
				WhiteboardZoomState: whiteboardZoomStateDefaultRes,
			},
		},
		{
			name:      "teacher get live lesson state without any state",
			reqUserID: "user-5",
			req: &bpb.LiveLessonStateRequest{
				Id: "lesson-1",
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(&entities.Lesson{
						LessonID: database.Text("lesson-1"),
					}, nil).
					Once()
				lessonRepo.
					On("GetLearnerIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-1",
							"user-2",
							"user-3",
						}),
						nil,
					).
					Once()
				lessonRepo.
					On("GetTeacherIDsOfLesson", ctx, db, database.Text("lesson-1")).
					Return(
						database.TextArray([]string{
							"user-5",
							"user-6",
						}),
						nil,
					).
					Once()
				lessonMemberRepo.
					On("GetLessonMemberStates", ctx, db, database.Text("lesson-1")).
					Return(
						entities.LessonMemberStates{},
						nil,
					).
					Once()
				logRepo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-1"),
					entities.TotalTimesGettingRoomState,
				).
					Return(nil).
					Once()
				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, database.Text("lesson-1")).
					Return(&domain.LessonRoomState{
						LessonID:            "lesson-1",
						SpotlightedUser:     "",
						WhiteboardZoomState: wDefault,
					}, nil).Once()
			},
			expectedRes: &bpb.LiveLessonStateResponse{
				Id:                  "lesson-1",
				UsersState:          &bpb.LiveLessonStateResponse_UsersState{},
				Spotlight:           &bpb.LiveLessonState_Spotlight{},
				WhiteboardZoomState: whiteboardZoomStateDefaultRes,
			},
		},
		{
			name:      "teacher get live lesson not exist",
			reqUserID: "user-5",
			req: &bpb.LiveLessonStateRequest{
				Id: "lesson-1",
			},
			setup: func(ctx context.Context) {
				lessonRepo.
					On("FindByID", ctx, db, database.Text("lesson-1")).
					Return(nil, pgx.ErrNoRows).
					Once()
			},
			hasError: true,
		},
		{
			name:      "teacher get live lesson with empty params",
			reqUserID: "user-5",
			req: &bpb.LiveLessonStateRequest{
				Id: "",
			},
			setup: func(ctx context.Context) {
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctxT := interceptors.ContextWithUserID(ctx, tc.reqUserID)
			tc.setup(ctxT)

			srv := LessonReaderServices{
				DB:                         db,
				LessonRepo:                 lessonRepo,
				LessonMemberRepo:           lessonMemberRepo,
				MediaRepo:                  mediaRepo,
				UserRepo:                   userRepo,
				VirtualClassRoomLogService: &log.VirtualClassRoomLogService{DB: db, Repo: logRepo},
				LessonRoomStateRepo:        lessonRoomStateRepo,
			}
			actualRes, err := srv.GetLiveLessonState(ctxT, tc.req)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tc.expectedRes.UsersState != nil && tc.expectedRes.UsersState.Learners != nil {
					lsExpected := tc.expectedRes.UsersState.Learners
					tc.expectedRes.UsersState.Learners = nil
					lsActual := actualRes.UsersState.Learners
					actualRes.UsersState.Learners = nil
					assert.ElementsMatch(t, lsExpected, lsActual)
				}
				assert.False(t, actualRes.CurrentTime.AsTime().IsZero())
				tc.expectedRes.CurrentTime = actualRes.CurrentTime
				assert.EqualValues(t, tc.expectedRes, actualRes)
				mock.AssertExpectationsForObjects(t, db, lessonRepo, lessonMemberRepo, mediaRepo, userRepo, logRepo)
			}
		})
	}
}

func TestRetrieveLiveLessonByLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	lessonRepo := new(mock_repositories.MockLessonRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	classMemberRepo := new(mock_repositories.MockClassMemberRepo)
	courseClassRepo := new(mock_repositories.MockCourseClassRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	classRepo := new(mock_repositories.MockClassRepo)
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	lessonReaderServices := &LessonReaderServices{
		UserRepo:         userRepo,
		LessonRepo:       lessonRepo,
		UnleashClientIns: mockUnleashClient,
		ClassRepo:        classRepo,
		CourseClassRepo:  courseClassRepo,
		Env:              "local",
		LessonMemberRepo: lessonMemberRepo,
	}
	var pgtypeTime pgtype.Timestamptz
	pgtypeTime.Set(time.Now())
	lessons := []*repositories.LessonWithTime{
		{
			Lesson: entities.Lesson{
				LessonID:  database.Text("lessonID"),
				TeacherID: database.Text("teacherID"),
				CourseID:  database.Text("courseID"),
				EndAt:     pgtypeTime,
			},
			PresetStudyPlanWeekly: entities.PresetStudyPlanWeekly{
				PresetStudyPlanID: database.Text("presetStudyPlanID"),
				TopicID:           database.Text("topicID"),
				StartDate:         pgtypeTime,
				EndDate:           pgtypeTime,
			},
		},
	}
	teachers := []*entities.User{
		{
			ID: database.Text("teacherID"),
		},
	}

	locationIds := []string{"location-1", "location-2"}

	mapClassIDByCourseID := make(map[pgtype.Text]pgtype.Int4Array)
	mapClassIDByCourseID[lessons[0].Lesson.CourseID] = database.Int4Array([]int32{0, 1, 2})
	emptymapClassIDByCourseID := make(map[pgtype.Text]pgtype.Int4Array)
	emptymapClassID := make(map[pgtype.Int4]pgtype.TextArray)

	var userClassIDs []int32
	var total pgtype.Int8
	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	// expectedCode := codes.Unknown
	pgSchedulingStatus := pgtype.Text{String: string(entities.LessonSchedulingStatusPublished), Status: pgtype.Present}
	pgSchedulingStatusNull := pgtype.Text{String: "", Status: pgtype.Null}
	courses := []string{"course1", "course2"}
	emptyClass := []*entities.Class{}
	testCases := []TestCase{
		{
			name:         "err find lesson",
			ctx:          ctx,
			req:          &bpb.RetrieveLiveLessonByLocationsRequest{},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_TEACHER.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()

				lessonRepo.On("FindLessonWithTime", ctx, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatus).Once().Return(lessons, total, pgx.ErrNoRows)
			},
		},
		{
			name:         "err find teacher",
			ctx:          ctx,
			req:          &bpb.RetrieveLiveLessonByLocationsRequest{},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_TEACHER.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()

				lessonRepo.On("FindLessonWithTime", ctx, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatus).Once().Return(lessons, total, nil)

				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(teachers, pgx.ErrNoRows)
			},
		},
		{
			name:         "teacher retrieve live lesson happy case",
			ctx:          ctx,
			req:          &bpb.RetrieveLiveLessonByLocationsRequest{},
			expectedResp: &bpb.RetrieveLiveLessonByLocationsResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_TEACHER.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()

				lessonRepo.On("FindLessonWithTime", ctx, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatus).Once().Return(lessons, total, nil)

				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(teachers, nil)
			},
		},
		{
			name:         "student retrieve live lesson happy case",
			ctx:          ctx,
			req:          &bpb.RetrieveLiveLessonByLocationsRequest{},
			expectedResp: &bpb.RetrieveLiveLessonByLocationsResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_STUDENT.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(emptyClass, nil)
				courseClassRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(emptymapClassID, nil)
				lessonMemberRepo.On("CourseAccessible", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)
				lessonRepo.On("FindLessonJoined", ctx, mock.Anything, database.Text(userID), mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatus).Once().Return(lessons, total, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(teachers, nil)

			},
		},
		{
			name:         "happy case with empty class map",
			ctx:          ctx,
			req:          &bpb.RetrieveLiveLessonByLocationsRequest{},
			expectedResp: &bpb.RetrieveLiveLessonByLocationsResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_TEACHER.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
				classMemberRepo.On("FindUsersClass", ctx, mock.Anything, mock.Anything).Once().Return(userClassIDs, nil)
				courseClassRepo.On("FindClassInCourse", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(emptymapClassIDByCourseID, nil)

				lessonRepo.On("FindLessonWithTime", ctx, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatus).Once().Return(lessons, total, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(teachers, nil)
			},
		},
		{
			name:         "teacher retrieve live lesson happy case with locationIDs",
			ctx:          ctx,
			req:          &bpb.RetrieveLiveLessonByLocationsRequest{LocationIds: locationIds},
			expectedResp: &bpb.RetrieveLiveLessonByLocationsResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_TEACHER.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()

				lessonRepo.On("FindLessonWithTimeAndLocations", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatus).Once().Return(lessons, total, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(teachers, nil)
			},
		},
		{
			name:         "student retrieve live lesson happy case with locationIDs",
			ctx:          ctx,
			req:          &bpb.RetrieveLiveLessonByLocationsRequest{LocationIds: locationIds},
			expectedResp: &bpb.RetrieveLiveLessonByLocationsResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_STUDENT.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(true, nil).Once()
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(emptyClass, nil)
				courseClassRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(emptymapClassID, nil)
				lessonMemberRepo.On("CourseAccessible", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)
				lessonRepo.On("FindLessonJoinedWithLocations", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatus).Once().Return(lessons, total, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(teachers, nil)
			},
		},
		{
			name:         "student retrieve live lesson happy case with locationIDs",
			ctx:          ctx,
			req:          &bpb.RetrieveLiveLessonByLocationsRequest{LocationIds: locationIds},
			expectedResp: &bpb.RetrieveLiveLessonByLocationsResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				userRepo.On("UserGroup", ctx, mock.Anything, database.Text(userID)).Once().Return(pb.USER_GROUP_STUDENT.String(), nil)
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				classRepo.On("FindJoined", ctx, mock.Anything, mock.Anything).Once().Return(emptyClass, nil)
				courseClassRepo.On("Find", ctx, mock.Anything, mock.Anything).Once().Return(emptymapClassID, nil)
				lessonMemberRepo.On("CourseAccessible", ctx, mock.Anything, mock.Anything).Once().Return(courses, nil)

				lessonRepo.On("FindLessonJoinedWithLocations", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, pgSchedulingStatusNull).Once().Return(lessons, total, nil)
				userRepo.On("Retrieve", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(teachers, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)
			rsp, err := lessonReaderServices.RetrieveLiveLessonByLocations(ctx, testCase.req.(*bpb.RetrieveLiveLessonByLocationsRequest))
			assert.Equal(t, status.Code(testCase.expectedErr), status.Code(err))
			if testCase.expectedErr != nil {
				assert.Nil(t, rsp, "expecting nil response")
			}
		})
	}
}
