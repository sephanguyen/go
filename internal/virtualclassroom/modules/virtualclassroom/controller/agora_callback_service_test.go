package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	lr_queries "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/queries"
	lr_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/queries"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/middlewares"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_liveroom_repo "github.com/manabie-com/backend/mock/virtualclassroom/liveroom/repositories"
	mock_virtual_repo "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name        string
	payload     *vc_domain.AgoraCallbackPayload
	ctx         context.Context
	expectedErr error
	setup       func(ctx context.Context)
}

func TestCallBackService(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRoomStateRepo := &mock_virtual_repo.MockLessonRoomStateRepo{}
	liveRoomStateRepo := &mock_liveroom_repo.MockLiveRoomStateRepo{}
	orgRepo := &mock_virtual_repo.MockOrganizationRepo{}
	w := httptest.NewRecorder()

	ctx, _ := gin.CreateTestContext(w)
	state := &vc_domain.LessonRoomState{
		LessonID:        "lesson-id",
		SpotlightedUser: "user-1",
		WhiteboardZoomState: &vc_domain.WhiteboardZoomState{
			PdfScaleRatio: 23.32,
			CenterX:       243.5,
			CenterY:       -432.034,
			PdfWidth:      234.43,
			PdfHeight:     -0.33424,
		},
		Recording: &vc_domain.CompositeRecordingState{
			ResourceID:  "resource-id",
			SID:         "s-id",
			UID:         123001,
			IsRecording: true,
			Creator:     "user-id-1",
		},
	}

	liveRoomState := &lr_domain.LiveRoomState{
		ChannelID: "channel-id",
		Recording: &vc_domain.CompositeRecordingState{
			ResourceID:  "resource-id",
			SID:         "s-id",
			UID:         123001,
			IsRecording: true,
			Creator:     "user-id-1",
		},
	}

	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: "-2147483647",
		},
	}
	orgIDs := []string{
		"-2147483642",
		"-2147483647",
	}

	ctxWithResourcePath := interceptors.ContextWithJWTClaims(ctx, claim)
	testCases := []TestCase{
		{
			name: "happy case with EventType is not supported",
			payload: &vc_domain.AgoraCallbackPayload{
				NoticeID:  "notice-id",
				ProductID: 34,
				EventType: 12,
				NotifyMs:  32,
				Payload: vc_domain.CloudRecordingPayload{
					ChannelName: "lesson-id",
					UID:         "234343001",
					SID:         "sid",
					Sequence:    2,
					SendTS:      234235,
					ServiceType: 2,
				},
			},
			ctx: ctxWithResourcePath,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				orgRepo.On("GetIDs", mock.Anything, db).Once().Return(orgIDs, nil)

				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, "lesson-id").
					Return(state, nil).Once()
			},
		},
		{
			name: "case with EventType is CloudRecordingServiceExited",
			payload: &vc_domain.AgoraCallbackPayload{
				NoticeID:  "notice-id",
				ProductID: 34,
				EventType: vc_domain.CloudRecordingServiceExited,
				NotifyMs:  32,
				Payload: vc_domain.CloudRecordingPayload{
					ChannelName: "lesson-id",
					UID:         "000123001",
					SID:         "s-id",
					Sequence:    2,
					SendTS:      234235,
					ServiceType: 2,
					Details: &vc_domain.SessionExitDetail{
						MsgName:    "abc",
						ExitStatus: 2,
					},
				},
			},
			ctx: ctxWithResourcePath,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(3)
				orgRepo.On("GetIDs", mock.Anything, db).Once().Return(orgIDs, nil)

				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, "lesson-id").
					Return(state, nil).Once()

				lessonRoomStateRepo.On("UpsertRecordingState", ctx, db, "lesson-id", mock.MatchedBy(func(req *vc_domain.CompositeRecordingState) bool {
					var r *vc_domain.CompositeRecordingState
					r = nil
					assert.Equal(t, r, req)
					return true
				})).
					Return(nil).Once()
			},
		},
		{
			name: "case with EventType is CloudRecordingServiceError",
			payload: &vc_domain.AgoraCallbackPayload{
				NoticeID:  "notice-id",
				ProductID: 34,
				EventType: vc_domain.CloudRecordingServiceError,
				NotifyMs:  32,
				Payload: vc_domain.CloudRecordingPayload{
					ChannelName: "lesson-id",
					UID:         "000123001",
					SID:         "s-id",
					Sequence:    2,
					SendTS:      234235,
					ServiceType: 2,
					Details: &vc_domain.CloudRecordingErrorDetail{
						MsgName:    "abc",
						Module:     0,
						ErrorLevel: 1,
						ErrorCode:  2,
						Stat:       1,
						ErrorMsg:   "sample-msg",
					},
				},
			},
			ctx: ctxWithResourcePath,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				orgRepo.On("GetIDs", mock.Anything, db).Once().Return(orgIDs, nil)

				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, "lesson-id").
					Return(state, nil).Once()
			},
		},
		{
			name: "case with EventType is CloudRecordingServiceExited and with live room",
			payload: &vc_domain.AgoraCallbackPayload{
				NoticeID:  "notice-id",
				ProductID: 34,
				EventType: vc_domain.CloudRecordingServiceExited,
				NotifyMs:  32,
				Payload: vc_domain.CloudRecordingPayload{
					ChannelName: "channel-id",
					UID:         "000123001",
					SID:         "s-id",
					Sequence:    2,
					SendTS:      234235,
					ServiceType: 2,
					Details: &vc_domain.SessionExitDetail{
						MsgName:    "abc",
						ExitStatus: 2,
					},
				},
			},
			ctx: ctxWithResourcePath,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Times(3)
				orgRepo.On("GetIDs", mock.Anything, db).Once().Return(orgIDs, nil)

				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, "channel-id").
					Return(nil, vc_domain.ErrLessonRoomStateNotFound).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, db, "channel-id").
					Return(liveRoomState, nil)

				liveRoomStateRepo.On("UpsertRecordingState", ctx, db, "channel-id", mock.MatchedBy(func(req *vc_domain.CompositeRecordingState) bool {
					var r *vc_domain.CompositeRecordingState
					r = nil
					assert.Equal(t, r, req)
					return true
				})).Return(nil).Once()
			},
		},
		{
			name: "case with EventType is CloudRecordingServiceError and with live room",
			payload: &vc_domain.AgoraCallbackPayload{
				NoticeID:  "notice-id",
				ProductID: 34,
				EventType: vc_domain.CloudRecordingServiceError,
				NotifyMs:  32,
				Payload: vc_domain.CloudRecordingPayload{
					ChannelName: "channel-id",
					UID:         "000123001",
					SID:         "s-id",
					Sequence:    2,
					SendTS:      234235,
					ServiceType: 2,
					Details: &vc_domain.CloudRecordingErrorDetail{
						MsgName:    "abc",
						Module:     0,
						ErrorLevel: 1,
						ErrorCode:  2,
						Stat:       1,
						ErrorMsg:   "sample-msg",
					},
				},
			},
			ctx: ctxWithResourcePath,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				orgRepo.On("GetIDs", mock.Anything, db).Once().Return(orgIDs, nil)

				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, "channel-id").
					Return(nil, vc_domain.ErrLessonRoomStateNotFound).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, db, "channel-id").
					Return(liveRoomState, nil)
			},
		},
		{
			name: "case with EventType is CloudRecordingServiceExited but lesson and live room states are empty",
			payload: &vc_domain.AgoraCallbackPayload{
				NoticeID:  "notice-id",
				ProductID: 34,
				EventType: vc_domain.CloudRecordingServiceExited,
				NotifyMs:  32,
				Payload: vc_domain.CloudRecordingPayload{
					ChannelName: "channel-id1",
					UID:         "000123001",
					SID:         "s-id",
					Sequence:    2,
					SendTS:      234235,
					ServiceType: 2,
					Details: &vc_domain.SessionExitDetail{
						MsgName:    "abc",
						ExitStatus: 2,
					},
				},
			},
			ctx: ctxWithResourcePath,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Twice()
				orgRepo.On("GetIDs", mock.Anything, db).Once().Return(orgIDs, nil)

				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, "channel-id1").
					Return(nil, vc_domain.ErrLessonRoomStateNotFound).Once()

				liveRoomStateRepo.On("GetLiveRoomStateByChannelID", ctx, db, "channel-id1").
					Return(nil, lr_domain.ErrChannelNotFound)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			payload, err := json.Marshal(testCase.payload)
			if err != nil {
				assert.Error(t, err)
			}

			ctx.Keys = map[string]interface{}{
				"payload": payload,
			}
			ctx.Request = &http.Request{
				Header: http.Header{
					middlewares.AgoraHeaderKey: []string{
						"signature",
					},
				},
			}

			a := &AgoraCallbackService{
				Logger: ctxzap.Extract(ctx),
				RecordingCommand: commands.RecordingCommand{
					WrapperDBConnection: wrapperConnection,
					LessonmgmtDB:        db,
					LessonRoomStateRepo: lessonRoomStateRepo,
					LiveRoomStateRepo:   liveRoomStateRepo,
				},
				LessonRoomStateQuery: queries.LessonRoomStateQuery{
					WrapperDBConnection: wrapperConnection,
					LessonRoomStateRepo: lessonRoomStateRepo,
				},
				LiveRoomStateQuery: lr_queries.LiveRoomStateQuery{
					LessonmgmtDB:      db,
					LiveRoomStateRepo: liveRoomStateRepo,
				},
				OrganizationQuery: queries.OrganizationQuery{
					WrapperDBConnection: wrapperConnection,
					OrganizationRepo:    orgRepo,
				},
			}
			a.CallBack(ctx)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 200, w.Code)
			}
			mock.AssertExpectationsForObjects(
				t,
				db,
				lessonRoomStateRepo,
				liveRoomStateRepo,
				orgRepo,
				mockUnleashClient,
			)
		})
	}
}
