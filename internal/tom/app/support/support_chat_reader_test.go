package support

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	golibtype "github.com/manabie-com/backend/internal/golibs/types"
	"github.com/manabie-com/backend/internal/tom/constants"
	tom_const "github.com/manabie-com/backend/internal/tom/constants"
	"github.com/manabie-com/backend/internal/tom/domain/core"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	sentities "github.com/manabie-com/backend/internal/tom/domain/support"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_elastic "github.com/manabie-com/backend/mock/golibs/elastic"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_support "github.com/manabie-com/backend/mock/tom/app/support"
	mock_repositories "github.com/manabie-com/backend/mock/tom/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/gogo/protobuf/types"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestConversationReader_ListConversationByUsers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	messageRepo := new(mock_repositories.MockMessageRepo)
	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationStudentRepo := new(mock_repositories.MockConversationStudentRepo)
	db := new(mock_database.Ext)
	s := &ChatReader{
		DB:                      db,
		ConversationMemberRepo:  conversationMemberRepo,
		ConversationRepo:        conversationRepo,
		MessageRepo:             messageRepo,
		ConversationStudentRepo: conversationStudentRepo,
	}
	var pgEndAt pgtype.Timestamptz
	_ = pgEndAt.Set(time.Now())
	teacherID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, teacherID)

	conversationIDs := []string{"conversation-1", "conversation-2"}

	userIDs := []string{"user-1", "user-2"}

	validConversationReq := &tpb.ListConversationByUsersRequest{
		ConversationIds: conversationIDs,
	}
	validUserReq := &tpb.ListConversationByUsersRequest{
		UserIds: userIDs,
	}

	mapConversationMember := make(map[pgtype.Text][]*core.ConversationMembers)
	mapConversationMember[database.Text("conversation-1")] = []*core.ConversationMembers{
		{
			UserID:         database.Text("student-1"),
			ConversationID: database.Text("conversation-1"),
		},
		{
			UserID:         database.Text("teacher-1"),
			ConversationID: database.Text("conversation-1"),
		},
	}
	mapConversationMember[database.Text("conversation-2")] = []*core.ConversationMembers{
		{
			UserID:         database.Text("parent-2"),
			ConversationID: database.Text("conversation-1"),
		},
		{
			UserID:         database.Text("teacher-2"),
			ConversationID: database.Text("conversation-1"),
		},
	}
	messages := make([]*core.Message, 0)
	messages = append(messages, &core.Message{
		ID:             database.Text("message-1"),
		ConversationID: database.Text("conversation-1"),
		UserID:         database.Text("student-1"),
		Message:        database.Text("message content 1"),
	})
	messages = append(messages, &core.Message{
		ID:             database.Text("message-2"),
		ConversationID: database.Text("conversation-2"),
		UserID:         database.Text("student-2"),
		Message:        database.Text("message content 2"),
	})
	//
	conversationFull := make(map[pgtype.Text]core.ConversationFull)
	conversationFull[database.Text("conversation-1")] = core.ConversationFull{
		Conversation: core.Conversation{
			ID:               database.Text("conversation-1"),
			ConversationType: database.Text("CONVERSATION_STUDENT"),
		},
		StudentID: database.Text("student-1"),
	}
	conversationFull[database.Text("conversation-2")] = core.ConversationFull{
		Conversation: core.Conversation{
			ID:               database.Text("conversation-2"),
			ConversationType: database.Text("CONVERSATION_PARENT"),
		},
		StudentID: database.Text("student-2"),
	}
	validResp := &tpb.ListConversationByUsersResponse{
		Items: []*tpb.Conversation{
			{
				ConversationId: "conversation-1",
				StudentId:      "student-1",
				Users: []*tpb.Conversation_User{
					{
						Id: "student-1",
					},
					{
						Id: "teacher-1",
					},
				},
				LastMessage: &tpb.MessageResponse{
					MessageId:      "message-1",
					ConversationId: "conversation-1",
					UserId:         "student-1",
					Content:        "message content 1",
				},
			},
			{
				ConversationId: "conversation-2",
				StudentId:      "student-2",
				Users: []*tpb.Conversation_User{
					{
						Id: "parent-2",
					},
					{
						Id: "teacher-2",
					},
				},
				LastMessage: &tpb.MessageResponse{
					MessageId:      "message-2",
					ConversationId: "conversation-2",
					UserId:         "student-2",
					Content:        "message content 2",
				},
			},
		},
	}
	var conversationType pgtype.Text
	_ = conversationType.Set(nil)
	testCases := map[string]TestCase{
		"err query find conversation by conversation ids": {
			ctx:          ctx,
			req:          validConversationReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Errorf("ConversationMemberRepo.FindByConversationIDs: %w", pgx.ErrTxClosed).Error()),
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray(validConversationReq.GetConversationIds())).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"err query find conversation by users ids": {
			ctx:          ctx,
			req:          validUserReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Errorf("ConversationStudentRepo.FindByStudentIDs: %w", pgx.ErrTxClosed).Error()),
			setup: func(ctx context.Context) {
				conversationStudentRepo.On("FindByStudentIDs", mock.Anything, mock.Anything, database.TextArray(validUserReq.GetUserIds()), conversationType).Once().Return(nil, pgx.ErrTxClosed)
			},
		},

		"err conversation find by conversation ids": {
			ctx:          ctx,
			req:          validConversationReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Errorf("ConversationRepo.FindByIDsReturnMapByID: %w", pgx.ErrTxClosed).Error()),
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray(validConversationReq.GetConversationIds())).Once().Return(mapConversationMember, nil)
				conversationRepo.On("FindByIDsReturnMapByID", mock.Anything, mock.Anything, database.TextArray(validConversationReq.GetConversationIds())).Once().Return(conversationFull, pgx.ErrTxClosed)
			},
		},
		"err conversation GetLastMessageByConversationIDs": {
			ctx:          ctx,
			req:          validConversationReq,
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Errorf("MessageRepo.GetLastMessageByConversationIDs: %w", pgx.ErrTxClosed).Error()),
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray(validConversationReq.GetConversationIds())).Once().Return(mapConversationMember, nil)
				conversationRepo.On("FindByIDsReturnMapByID", mock.Anything, mock.Anything, database.TextArray(validConversationReq.GetConversationIds())).Once().Return(conversationFull, nil)
				messageRepo.On(
					"GetLastMessageByConversationIDs",
					mock.Anything, mock.Anything, database.TextArray(validConversationReq.GetConversationIds()), uint(len(conversationIDs)), mock.AnythingOfType("pgtype.Timestamptz"), true).
					Once().Return(messages, pgx.ErrTxClosed)
			},
		},
		"success": {
			ctx:          ctx,
			req:          validConversationReq,
			expectedResp: validResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				conversationMemberRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray(validConversationReq.GetConversationIds())).Once().Return(mapConversationMember, nil)
				conversationRepo.On("FindByIDsReturnMapByID", mock.Anything, mock.Anything, database.TextArray(validConversationReq.GetConversationIds())).Once().Return(conversationFull, nil)
				messageRepo.On(
					"GetLastMessageByConversationIDs", mock.Anything, mock.Anything, database.TextArray(validConversationReq.GetConversationIds()), uint(len(conversationIDs)), mock.AnythingOfType("pgtype.Timestamptz"), true).
					Once().Return(messages, nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.ListConversationByUsers(testCase.ctx, testCase.req.(*tpb.ListConversationByUsersRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, resp)
			} else {
				counter := 0
				expectedRespCasted := testCase.expectedResp.(*tpb.ListConversationByUsersResponse)
				assert.Equal(t, len(expectedRespCasted.GetItems()), len(resp.GetItems()), "the length item not equal")
				for _, item1 := range expectedRespCasted.GetItems() {
					for _, item2 := range resp.GetItems() {
						if item1.ConversationId == item2.ConversationId && len(item1.GetUsers()) == len(item2.GetUsers()) {
							counter++
						}
					}
				}
				assert.Equal(t, len(expectedRespCasted.GetItems()), counter, "some items attribute is not equal")
			}
		})
	}
}

func TestChatServiceReader_ListConversationsInSchool(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	messageRepo := new(mock_repositories.MockMessageRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationStudentRepo := new(mock_repositories.MockConversationStudentRepo)
	searchRepo := new(mock_repositories.MockSearchRepo)
	externalConfigurationServiceMock := new(ExternalConfigurationServiceMock)
	locationRepo := new(mock_repositories.MockLocationRepo)
	db := &mock_database.Ext{}
	mockEs := &mock_elastic.SearchFactory{}
	s := &ChatReader{
		DB:                           db,
		SearchClient:                 mockEs,
		MessageRepo:                  messageRepo,
		ConversationMemberRepo:       conversationMemberRepo,
		ConversationRepo:             conversationRepo,
		ConversationSearchRepo:       searchRepo,
		ConversationStudentRepo:      conversationStudentRepo,
		ExternalConfigurationService: externalConfigurationServiceMock,
		LocationRepo:                 locationRepo,
	}

	orgIds := []string{"org-id"}
	getExternalConfigByKeysAndLocationsFun := func(configs ...*mpb.LocationConfiguration) func(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsRequest, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsResponse, error) {
		return func(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsRequest, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsResponse, error) {
			assert.ElementsMatch(t, in.Keys, []string{tom_const.ChatConfigKeyParent, tom_const.ChatConfigKeyStudent})
			assert.ElementsMatch(t, in.LocationsIds, orgIds)
			return &mpb.GetConfigurationByKeysAndLocationsResponse{
				Configurations: configs,
			}, nil
		}
	}
	// map[pgtype.Text]core.ConversationFull
	userID := idutil.ULIDNow()
	studentID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	cID := idutil.ULIDNow()

	seenAt := time.Now()
	conversationMap := map[pgtype.Text]core.ConversationFull{
		database.Text(cID): {
			Conversation: core.Conversation{
				ID: database.Text(cID),
			},
		},
	}
	conversationMemberMap := map[pgtype.Text][]*core.ConversationMembers{
		database.Text(cID): {
			{
				UserID:         database.Text(userID),
				ConversationID: database.Text(cID),
				SeenAt:         database.Timestamptz(seenAt),
			},
		},
	}
	conversationStudentMap := map[pgtype.Text]*sentities.ConversationStudent{
		database.Text(cID): {
			ConversationID: database.Text(cID),
			StudentID:      database.Text(studentID),
		},
	}

	claims := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			SchoolIDs: []string{"1"},
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claims)
	lastOffsetTime := time.Now()
	sampleDoc := []sentities.SearchConversationDoc{
		{
			ConversationID:          cID,
			ConversationNameEnglish: "fake conversation name",
			UserIDs:                 []string{userID},
			LastMessage: sentities.SearchLastMessage{
				UpdatedAt: lastOffsetTime,
			},
			IsReplied: true,
		},
	}
	// TODO: add detail test + add test
	testCases := []TestCase{
		{
			name: "error find conversation in school",
			ctx:  ctx,
			req: &tpb.ListConversationsInSchoolRequest{
				CourseIds: []string{"c1"},
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetString{},
				},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "search service failed"),
			setup: func(ctx context.Context) {
				conversationID := pgtype.Text{}
				_ = conversationID.Set(cID)
				locationRepo.On("FindRootIDs", mock.Anything, mock.Anything).Once().Return(orgIds, nil)
				searchRepo.On("Search", mock.Anything, mock.Anything, mock.MatchedBy(func(arg sentities.ConversationFilter) bool {
					expect := sentities.ConversationFilter{
						UserID:  userID,
						School:  golibtype.NewStrArr([]string{"1"}),
						Courses: golibtype.NewStrArr([]string{"c1"}),
						SortBy:  sentities.DefaultConversationSorts,
						Limit:   golibtype.NewInt64(10),
						ConversationTypes: golibtype.NewStrArr([]string{
							tpb.ConversationType_CONVERSATION_STUDENT.String(),
							tpb.ConversationType_CONVERSATION_PARENT.String(),
						}),
					}
					return cmp.Equal(expect, arg)
				})).Once().Return(nil, fmt.Errorf("404"))
				externalConfigurationServiceMock.getConfigurationByKeysAndLocations = getExternalConfigByKeysAndLocationsFun()
			},
		},
		{
			name: "success",
			ctx:  ctx,
			req: &tpb.ListConversationsInSchoolRequest{
				Name: wrapperspb.String("fake name"),
				Type: []tpb.ConversationType{
					tpb.ConversationType_CONVERSATION_STUDENT,
				},
				TeacherStatus: tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_ALL,
				JoinStatus:    tpb.ConversationJoinStatus_CONVERSATION_JOIN_STATUS_JOINED,
				CourseIds:     []string{"c1"},
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetString{},
				},
			},
			expectedResp: &tpb.ListConversationsInSchoolResponse{
				Items: []*tpb.Conversation{
					{
						ConversationId: cID,
						StudentId:      studentID,
						Users: []*tpb.Conversation_User{
							{
								Id:     userID,
								SeenAt: timestamppb.New(seenAt),
							},
						},
					},
				},
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetCombined{
						OffsetCombined: &cpb.Paging_Combined{
							OffsetString:  cID,
							OffsetInteger: lastOffsetTime.UnixMilli(),
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				conversationID := pgtype.Text{}
				_ = conversationID.Set(cID)
				locationRepo.On("FindRootIDs", mock.Anything, mock.Anything).Once().Return(orgIds, nil)
				searchRepo.On("Search", mock.Anything, mock.Anything, mock.MatchedBy(func(arg sentities.ConversationFilter) bool {
					expect := sentities.ConversationFilter{
						UserID:            userID,
						JoinStatus:        golibtype.NewBool(true),
						School:            golibtype.NewStrArr([]string{"1"}),
						ConversationName:  golibtype.NewStr("fake name"),
						ConversationTypes: golibtype.NewStrArr([]string{tpb.ConversationType_CONVERSATION_STUDENT.String()}),
						Courses:           golibtype.NewStrArr([]string{"c1"}),
						SortBy:            sentities.DefaultConversationSorts,
						Limit:             golibtype.NewInt64(10),
					}
					return cmp.Equal(expect, arg)
				})).Once().Return(sampleDoc, nil)
				messageRepo.On("GetLastMessageByConversationIDs", mock.Anything, mock.Anything,
					database.TextArray([]string{cID}), uint(10), mock.Anything, false).Once().Return(nil, nil)
				conversationRepo.On("FindByIDsReturnMapByID", mock.Anything, mock.Anything, database.TextArray([]string{cID})).
					Once().Return(conversationMap, nil)
				conversationMemberRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray([]string{cID})).
					Once().Return(conversationMemberMap, nil)
				conversationStudentRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray([]string{cID})).
					Once().Return(conversationStudentMap, nil)
				externalConfigurationServiceMock.getConfigurationByKeysAndLocations = getExternalConfigByKeysAndLocationsFun()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.ListConversationsInSchool(testCase.ctx, testCase.req.(*tpb.ListConversationsInSchoolRequest))
			if err != nil {
				assert.ErrorIs(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
func TestChatReader_RetrieveTotalUnreadConversationWithLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	convRepo := new(mock_repositories.MockConversationRepo)
	locationRepo := new(mock_repositories.MockLocationRepo)
	conversationLocationRepo := new(mock_repositories.MockConversationLocationRepo)
	externalConfigurationServiceMock := new(ExternalConfigurationServiceMock)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	locationConfigResolver := new(mock_support.MockLocationConfigResolver)
	mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).
		Return(true, nil)

	s := &ChatReader{
		ConversationRepo:             convRepo,
		LocationRepo:                 locationRepo,
		ConversationLocationRepo:     conversationLocationRepo,
		ExternalConfigurationService: externalConfigurationServiceMock,
		LocationConfigResolver:       locationConfigResolver,
		UnleashClientIns:             mockUnleashClient,
		Env:                          tom_const.LocalEnv,
	}

	userID := idutil.ULIDNow()
	locIDs := []string{"location-1", "location-2"}
	accessPath := []string{"root-id/parent-id/location-1", "root-id/parent-id/location-2"}

	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, constant.RoleParent)
	ctx = interceptors.ContextWithUserID(ctx, userID)

	testCases := map[string]TestCase{
		"empty location id": {
			ctx: interceptors.NewIncomingContext(ctx),
			req: &tpb.RetrieveTotalUnreadConversationsWithLocationsRequest{
				LocationIds: []string{},
			},
			// req: validReq,
			expectedResp: &tpb.RetrieveTotalUnreadConversationsWithLocationsResponse{
				TotalUnreadConversations: 1,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationLocationRepo.On("GetAllLocations", ctx, mock.Anything, userID).
					Once().
					Return(
						[]*core.ConversationLocation{
							{
								ConversationID: database.Text("conversation-1"),
								LocationID:     database.Text("location-1"),
								AccessPath:     database.Text("root-id/parent-id/location-1"),
							},
							{
								ConversationID: database.Text("conversation-2"),
								LocationID:     database.Text("location-2"),
								AccessPath:     database.Text("root-id/parent-id/location-2"),
							},
						}, nil,
					)
				externalConfigurationServiceMock.getConfigurationByKeysAndLocationsV2 = func(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
					assert.ElementsMatch(t, in.Keys, []string{tom_const.ChatConfigKeyParentV2})
					assert.ElementsMatch(t, in.LocationIds, []string{"location-1", "location-2"})
					return &mpb.GetConfigurationByKeysAndLocationsV2Response{
						Configurations: []*mpb.LocationConfiguration{
							{
								Id:              "id-1",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-1",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id-2",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-2",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
						},
					}, nil
				}
				convRepo.On("CountUnreadConversations", ctx, mock.Anything, database.Text(userID), database.TextArray(supportChatTypes), database.TextArray([]string{"location-1", "location-2"}), true).Once().Return(int64(1), nil)
			},
		},
		"success with location": {
			ctx: interceptors.NewIncomingContext(ctx),
			req: &tpb.RetrieveTotalUnreadConversationsWithLocationsRequest{
				LocationIds: locIDs,
			},
			// req: validReq,
			expectedResp: &tpb.RetrieveTotalUnreadConversationsWithLocationsResponse{
				TotalUnreadConversations: 1,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationLocationRepo.On("GetAllLocations", ctx, mock.Anything, userID).
					Once().
					Return(
						[]*core.ConversationLocation{
							{
								ConversationID: database.Text("conversation-1"),
								LocationID:     database.Text("location-1"),
								AccessPath:     database.Text("root-id/parent-id/location-1"),
							},
							{
								ConversationID: database.Text("conversation-2"),
								LocationID:     database.Text("location-2"),
								AccessPath:     database.Text("root-id/parent-id/location-2"),
							},
						}, nil,
					)
				conversationTypeWithLocationMap := map[tpb.ConversationType][]string{
					tpb.ConversationType_CONVERSATION_STUDENT: accessPath,
					tpb.ConversationType_CONVERSATION_PARENT:  accessPath,
				}
				locationConfigResolver.On("GetEnabledLocationConfigsByLocations", mock.Anything, locIDs, []tpb.ConversationType{tpb.ConversationType_CONVERSATION_STUDENT, tpb.ConversationType_CONVERSATION_PARENT}).
					Once().Return(conversationTypeWithLocationMap, nil)
				convRepo.On("CountUnreadConversationsByAccessPathsV2",
					ctx, mock.Anything, database.Text(userID), database.TextArray(accessPath), database.TextArray(accessPath),
				).Once().Return(int64(1), nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.RetrieveTotalUnreadConversationsWithLocations(testCase.ctx, testCase.req.(*tpb.RetrieveTotalUnreadConversationsWithLocationsRequest))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestChatReader_RetrieveTotalUnreadMessage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	convRepo := new(mock_repositories.MockConversationRepo)
	conversationLocationRepo := new(mock_repositories.MockConversationLocationRepo)
	externalConfigurationServiceMock := new(ExternalConfigurationServiceMock)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).
		Return(true, nil)

	s := &ChatReader{
		ConversationRepo:             convRepo,
		ConversationLocationRepo:     conversationLocationRepo,
		ExternalConfigurationService: externalConfigurationServiceMock,
		UnleashClientIns:             mockUnleashClient,
		Env:                          tom_const.LocalEnv,
	}

	userID := idutil.ULIDNow()
	validReq := &tpb.RetrieveTotalUnreadMessageRequest{
		UserId: userID,
	}

	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, constant.RoleParent)
	ctx = interceptors.ContextWithUserID(ctx, userID)

	testCases := map[string]TestCase{
		"err finding unread messageNum": {
			ctx:          interceptors.NewIncomingContext(ctx),
			req:          validReq,
			expectedResp: nil,
			expectedErr:  fmt.Errorf("s.MessageRepo.CountUnreadMessages: %w", pgx.ErrTxClosed),
			setup: func(ctx context.Context) {
				conversationLocationRepo.On("GetAllLocations", ctx, mock.Anything, userID).
					Once().
					Return(
						[]*core.ConversationLocation{
							{
								ConversationID: database.Text("conversation-1"),
								LocationID:     database.Text("location-1"),
								AccessPath:     database.Text("root-id/parent-id/location-1"),
							},
							{
								ConversationID: database.Text("conversation-2"),
								LocationID:     database.Text("location-2"),
								AccessPath:     database.Text("root-id/parent-id/location-2"),
							},
						}, nil,
					)
				externalConfigurationServiceMock.getConfigurationByKeysAndLocationsV2 = func(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
					assert.ElementsMatch(t, in.Keys, []string{tom_const.ChatConfigKeyParentV2})
					assert.ElementsMatch(t, in.LocationIds, []string{"location-1", "location-2"})
					return &mpb.GetConfigurationByKeysAndLocationsV2Response{
						Configurations: []*mpb.LocationConfiguration{
							{
								Id:              "id-1",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-1",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id-2",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-2",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
						},
					}, nil
				}
				convRepo.On("CountUnreadConversations", ctx, mock.Anything, database.Text(userID), database.TextArray(supportChatTypes), database.TextArray([]string{"location-1", "location-2"}), true).Once().Return(int64(0), pgx.ErrTxClosed)
			},
		},
		"success": {
			ctx: interceptors.NewIncomingContext(ctx),
			req: validReq,
			expectedResp: &tpb.RetrieveTotalUnreadMessageResponse{
				TotalUnreadMessages: 1,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationLocationRepo.On("GetAllLocations", ctx, mock.Anything, userID).
					Once().
					Return(
						[]*core.ConversationLocation{
							{
								ConversationID: database.Text("conversation-1"),
								LocationID:     database.Text("location-1"),
								AccessPath:     database.Text("root-id/parent-id/location-1"),
							},
							{
								ConversationID: database.Text("conversation-2"),
								LocationID:     database.Text("location-2"),
								AccessPath:     database.Text("root-id/parent-id/location-2"),
							},
						}, nil,
					)
				externalConfigurationServiceMock.getConfigurationByKeysAndLocationsV2 = func(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
					assert.ElementsMatch(t, in.Keys, []string{tom_const.ChatConfigKeyParentV2})
					assert.ElementsMatch(t, in.LocationIds, []string{"location-1", "location-2"})
					return &mpb.GetConfigurationByKeysAndLocationsV2Response{
						Configurations: []*mpb.LocationConfiguration{
							{
								Id:              "id-1",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-1",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id-2",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-2",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
						},
					}, nil
				}
				convRepo.On("CountUnreadConversations", ctx, mock.Anything, database.Text(userID), database.TextArray(supportChatTypes), database.TextArray([]string{"location-1", "location-2"}), true).Once().Return(int64(1), nil)
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.RetrieveTotalUnreadMessage(testCase.ctx, testCase.req.(*tpb.RetrieveTotalUnreadMessageRequest))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestConversationReader_ListConversationIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: "manabie",
		},
	})

	conversationRepo := new(mock_repositories.MockConversationRepo)
	db := new(mock_database.Ext)
	s := &ChatReader{
		DB:               db,
		ConversationRepo: conversationRepo,
	}
	var (
		conversationID pgtype.Text
		limit          uint32 = 100
		schoolID              = database.Text("manabie")
	)
	conversationTypesAccepted := database.TextArray([]string{"CONVERSATION_STUDENT", "CONVERSATION_PARENT"})
	_ = conversationID.Set(nil)
	testCases := map[string]TestCase{
		"err query": {
			ctx:          ctx,
			req:          &tpb.ListConversationIDsRequest{},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Errorf("ConversationRepo.ListAll: %w", pgx.ErrTxClosed).Error()),
			setup: func(ctx context.Context) {
				conversationRepo.On("ListAll", mock.Anything, mock.Anything, conversationID, limit, conversationTypesAccepted, schoolID).Once().Return(nil, pgx.ErrTxClosed)
			},
		},
		"success": {
			ctx:          ctx,
			req:          &tpb.ListConversationIDsRequest{},
			expectedResp: &tpb.ListConversationIDsResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				conversationRepo.On("ListAll", mock.Anything, mock.Anything, conversationID, limit, conversationTypesAccepted, schoolID).Once().Return(nil, nil)
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.ListConversationIDs(testCase.ctx, testCase.req.(*tpb.ListConversationIDsRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, resp)
			} else {
				expectedRespCasted := testCase.expectedResp.(*tpb.ListConversationIDsResponse)
				assert.Equal(t, len(expectedRespCasted.GetConversationIds()), len(resp.ConversationIds), "length of list conversations response is not equal to expected")
			}
		})
	}
}

func TestChatService_ConversationList(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	messageRepo := new(mock_repositories.MockMessageRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationStudentRepo := new(mock_repositories.MockConversationStudentRepo)
	conversationLocationRepo := new(mock_repositories.MockConversationLocationRepo)
	externalConfigurationServiceMock := new(ExternalConfigurationServiceMock)
	locationRepoMock := new(mock_repositories.MockLocationRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).
		Return(true, nil)

	s := &ChatReader{
		MessageRepo:                  messageRepo,
		ConversationRepo:             conversationRepo,
		ConversationMemberRepo:       conversationMemberRepo,
		ConversationStudentRepo:      conversationStudentRepo,
		ConversationLocationRepo:     conversationLocationRepo,
		ExternalConfigurationService: externalConfigurationServiceMock,
		LocationRepo:                 locationRepoMock,
		Env:                          tom_const.LocalEnv,
		UnleashClientIns:             mockUnleashClient,
	}

	userID := idutil.ULIDNow()
	studentID := idutil.ULIDNow() // incase user == parent, use this
	ctx = interceptors.ContextWithUserID(ctx, userID)
	ctx = interceptors.ContextWithUserGroup(ctx, constant.RoleParent)

	cID := idutil.ULIDNow()
	now := types.TimestampNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	testCases := []TestCase{
		{
			name:         "err query db GetLastMessageEachConversation",
			ctx:          ctx,
			customCtx:    customParentCtx,
			req:          &pb.ConversationListRequest{Limit: 10, EndAt: nil},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unknown, fmt.Errorf("MessageRepo.GetLastMessageEachUserConversation: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				messageRepo.On("GetLastMessageEachUserConversation",
					ctx,
					mock.Anything,
					pgtype.Text{String: userID, Status: pgtype.Present},
					pgtype.Text{String: pb.CONVERSATION_STATUS_NONE.String(), Status: pgtype.Present},
					uint(10),
					mock.Anything,
					database.TextArray([]string{"location-1", "location-2"}),
					true,
				).
					Once().Return(nil, pgx.ErrNoRows)
				conversationLocationRepo.On("GetAllLocations", ctx, mock.Anything, userID).
					Once().
					Return(
						[]*core.ConversationLocation{
							{
								ConversationID: database.Text(cID),
								LocationID:     database.Text("location-1"),
								AccessPath:     database.Text("root-id/parent-id/location-1"),
							},
							{
								ConversationID: database.Text(cID),
								LocationID:     database.Text("location-2"),
								AccessPath:     database.Text("root-id/parent-id/location-2"),
							},
						}, nil,
					)
				externalConfigurationServiceMock.getConfigurationByKeysAndLocationsV2 = func(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
					assert.ElementsMatch(t, in.Keys, []string{tom_const.ChatConfigKeyParentV2})
					assert.ElementsMatch(t, in.LocationIds, []string{"location-1", "location-2"})
					return &mpb.GetConfigurationByKeysAndLocationsV2Response{
						Configurations: []*mpb.LocationConfiguration{
							{
								Id:              "id-1",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-1",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id-2",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-2",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
						},
					}, nil
				}
			},
		},
		{
			name:      "success list student's conversation",
			ctx:       ctx,
			customCtx: customStudentCtx,
			req:       &pb.ConversationListRequest{Limit: 10, EndAt: types.TimestampNow()},
			expectedResp: &pb.ConversationListResponse{
				Conversations: []*pb.Conversation{
					{
						ConversationId: cID,
						StudentId:      userID,
						GuestIds:       nil,
						Seen:           true,
						LastMessage: &pb.MessageResponse{
							ConversationId: cID,
							MessageId:      "",
							UserId:         userID,
							Content:        "content",
							UrlMedia:       "",
							Type:           0,
							CreatedAt:      now,
							LocalMessageId: "",
						},
						Users: []*pb.Conversation_User{
							{
								Id:        userID,
								Group:     core.ConversationRoleStudent,
								IsPresent: true,
							},
							{
								Group:     core.ConversationRoleTeacher,
								IsPresent: true,
							},
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationID := pgtype.Text{}
				_ = conversationID.Set(cID)

				conversationLocationRepo.On("GetAllLocations", ctx, mock.Anything, userID).
					Once().
					Return(
						[]*core.ConversationLocation{
							{
								ConversationID: conversationID,
								LocationID:     database.Text("location-1"),
								AccessPath:     database.Text("root-id/parent-id/location-1"),
							},
							{
								ConversationID: conversationID,
								LocationID:     database.Text("location-2"),
								AccessPath:     database.Text("root-id/parent-id/location-2"),
							},
						}, nil,
					)
				externalConfigurationServiceMock.getConfigurationByKeysAndLocationsV2 = func(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
					assert.ElementsMatch(t, in.Keys, []string{tom_const.ChatConfigKeyParentV2})
					assert.ElementsMatch(t, in.LocationIds, []string{"location-1", "location-2"})
					return &mpb.GetConfigurationByKeysAndLocationsV2Response{
						Configurations: []*mpb.LocationConfiguration{
							{
								Id:              "id-1",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-1",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id-2",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-2",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
						},
					}, nil
				}

				messageRepo.On("GetLastMessageEachUserConversation", ctx, mock.Anything, pgtype.Text{String: userID, Status: pgtype.Present}, pgtype.Text{String: pb.CONVERSATION_STATUS_NONE.String(), Status: pgtype.Present}, uint(10), mock.Anything, database.TextArray([]string{"location-1", "location-2"}), true).
					Once().Return([]*entities.Message{
					{
						ConversationID: conversationID,
						UserID:         pgtype.Text{String: userID, Status: 2},
						Message:        pgtype.Text{String: "content", Status: 2},
						CreatedAt:      pgtype.Timestamptz{Time: time.Unix(now.Seconds, int64(now.Nanos)), Status: 2},
					},
				}, nil)

				conversations := map[pgtype.Text]core.ConversationFull{
					conversationID: {
						Conversation: core.Conversation{
							ID:        conversationID,
							CreatedAt: pgtype.Timestamptz{},
							UpdatedAt: pgtype.Timestamptz{},
						},
					},
				}

				conversationRepo.On("FindByIDsReturnMapByID", ctx, mock.Anything, database.TextArray([]string{conversationID.String})).Once().Return(conversations, nil)

				conversationMembers := map[pgtype.Text][]*core.ConversationMembers{
					conversationID: {
						{
							ID:             pgtype.Text{},
							UserID:         pgtype.Text{String: userID, Status: 2},
							ConversationID: pgtype.Text{},
							Role:           pgtype.Text{String: core.ConversationRoleStudent, Status: 2},
							Status:         pgtype.Text{String: core.ConversationStatusActive, Status: 2},
							SeenAt:         pgtype.Timestamptz{Time: time.Now(), Status: 2},
							LastNotifyAt:   pgtype.Timestamptz{},
							CreatedAt:      pgtype.Timestamptz{},
							UpdatedAt:      pgtype.Timestamptz{},
						},
						{
							ID:             pgtype.Text{},
							UserID:         pgtype.Text{String: "", Status: 2},
							ConversationID: pgtype.Text{},
							Role:           pgtype.Text{String: core.ConversationRoleTeacher, Status: 2},
							Status:         pgtype.Text{String: core.ConversationStatusActive, Status: 2},
							SeenAt:         pgtype.Timestamptz{Time: time.Now(), Status: 2},
							LastNotifyAt:   pgtype.Timestamptz{},
							CreatedAt:      pgtype.Timestamptz{},
							UpdatedAt:      pgtype.Timestamptz{},
						},
					},
				}
				conversationStudents := map[pgtype.Text]*sentities.ConversationStudent{
					conversationID: {
						ConversationID: conversationID,
						StudentID:      database.Text(userID),
					},
				}

				conversationMemberRepo.On("FindByConversationIDs", ctx, mock.Anything, mock.Anything).Once().Return(conversationMembers, nil)
				conversationStudentRepo.On("FindByConversationIDs", ctx, mock.Anything, mock.Anything).Once().Return(conversationStudents, nil)
			},
		},
		{
			name:      "success list parent's conversation",
			ctx:       ctx,
			customCtx: customParentCtx,
			req:       &pb.ConversationListRequest{Limit: 10, EndAt: types.TimestampNow()},
			expectedResp: &pb.ConversationListResponse{
				Conversations: []*pb.Conversation{
					{
						ConversationId: cID,
						StudentId:      studentID,
						GuestIds:       nil,
						Seen:           true,
						LastMessage: &pb.MessageResponse{
							ConversationId: cID,
							MessageId:      "",
							UserId:         userID,
							Content:        "content",
							UrlMedia:       "",
							Type:           0,
							CreatedAt:      now,
							LocalMessageId: "",
						},
						Users: []*pb.Conversation_User{
							{
								Id:        userID,
								Group:     core.ConversationRoleParent,
								IsPresent: true,
							},
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationID := pgtype.Text{}
				_ = conversationID.Set(cID)

				conversationLocationRepo.On("GetAllLocations", ctx, mock.Anything, userID).
					Once().
					Return(
						[]*core.ConversationLocation{
							{
								ConversationID: conversationID,
								LocationID:     database.Text("location-1"),
								AccessPath:     database.Text("root-id/parent-id/location-1"),
							},
							{
								ConversationID: conversationID,
								LocationID:     database.Text("location-2"),
								AccessPath:     database.Text("root-id/parent-id/location-2"),
							},
						}, nil,
					)
				externalConfigurationServiceMock.getConfigurationByKeysAndLocationsV2 = func(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
					assert.ElementsMatch(t, in.Keys, []string{tom_const.ChatConfigKeyParentV2})
					assert.ElementsMatch(t, in.LocationIds, []string{"location-1", "location-2"})
					return &mpb.GetConfigurationByKeysAndLocationsV2Response{
						Configurations: []*mpb.LocationConfiguration{
							{
								Id:              "id-1",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-1",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id-2",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-2",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
						},
					}, nil
				}

				messageRepo.On("GetLastMessageEachUserConversation", ctx, mock.Anything, pgtype.Text{String: userID, Status: pgtype.Present}, pgtype.Text{String: pb.CONVERSATION_STATUS_NONE.String(), Status: pgtype.Present}, uint(10), mock.Anything, database.TextArray([]string{"location-1", "location-2"}), true).
					Once().Return([]*entities.Message{
					{
						ConversationID: conversationID,
						UserID:         pgtype.Text{String: userID, Status: 2},
						Message:        pgtype.Text{String: "content", Status: 2},
						CreatedAt:      pgtype.Timestamptz{Time: time.Unix(now.Seconds, int64(now.Nanos)), Status: 2},
					},
				}, nil)

				conversations := map[pgtype.Text]core.ConversationFull{
					conversationID: {
						Conversation: core.Conversation{
							ID:        conversationID,
							CreatedAt: pgtype.Timestamptz{},
							UpdatedAt: pgtype.Timestamptz{},
						},
					},
				}

				conversationRepo.On("FindByIDsReturnMapByID", ctx, mock.Anything, database.TextArray([]string{conversationID.String})).Once().Return(conversations, nil)

				conversationMembers := map[pgtype.Text][]*core.ConversationMembers{
					conversationID: {
						{
							ID:             pgtype.Text{},
							UserID:         pgtype.Text{String: userID, Status: 2},
							ConversationID: pgtype.Text{},
							Role:           pgtype.Text{String: core.ConversationRoleParent, Status: 2},
							Status:         pgtype.Text{String: core.ConversationStatusActive, Status: 2},
							SeenAt:         pgtype.Timestamptz{Time: time.Now(), Status: 2},
							LastNotifyAt:   pgtype.Timestamptz{},
							CreatedAt:      pgtype.Timestamptz{},
							UpdatedAt:      pgtype.Timestamptz{},
						},
					},
				}
				conversationMemberRepo.On("FindByConversationIDs", ctx, mock.Anything, mock.Anything).Once().Return(conversationMembers, nil)

				conversationStudents := map[pgtype.Text]*sentities.ConversationStudent{
					conversationID: {
						ConversationID: conversationID,
						StudentID:      database.Text(studentID),
					},
				}

				conversationStudentRepo.On("FindByConversationIDs", ctx, mock.Anything, mock.Anything).Once().Return(conversationStudents, nil)
			},
		},
		{
			name:      "1 allowed and 1 not allowed locations",
			ctx:       ctx,
			customCtx: customParentCtx,
			req:       &pb.ConversationListRequest{Limit: 10, EndAt: types.TimestampNow()},
			expectedResp: &pb.ConversationListResponse{
				Conversations: []*pb.Conversation{
					{
						ConversationId: cID,
						StudentId:      studentID,
						GuestIds:       nil,
						Seen:           true,
						LastMessage: &pb.MessageResponse{
							ConversationId: cID,
							MessageId:      "",
							UserId:         userID,
							Content:        "content",
							UrlMedia:       "",
							Type:           0,
							CreatedAt:      now,
							LocalMessageId: "",
						},
						Users: []*pb.Conversation_User{
							{
								Id:        userID,
								Group:     core.ConversationRoleParent,
								IsPresent: true,
							},
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationID := pgtype.Text{}
				_ = conversationID.Set(cID)

				conversationLocationRepo.On("GetAllLocations", ctx, mock.Anything, userID).
					Once().
					Return(
						[]*core.ConversationLocation{
							{
								ConversationID: conversationID,
								LocationID:     database.Text("location-1"),
								AccessPath:     database.Text("root-id/parent-id/location-1"),
							},
							{
								ConversationID: conversationID,
								LocationID:     database.Text("location-2"),
								AccessPath:     database.Text("not-allow-root-id/parent-id/location-2"),
							},
						}, nil,
					)
				externalConfigurationServiceMock.getConfigurationByKeysAndLocationsV2 = func(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
					assert.ElementsMatch(t, in.Keys, []string{tom_const.ChatConfigKeyParentV2})
					assert.ElementsMatch(t, in.LocationIds, []string{"location-1", "location-2"})
					return &mpb.GetConfigurationByKeysAndLocationsV2Response{
						Configurations: []*mpb.LocationConfiguration{
							{
								Id:              "id-1",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-1",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id-2",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-2",
								ConfigValue:     "false",
								ConfigValueType: "boolean",
							},
						},
					}, nil
				}

				messageRepo.On("GetLastMessageEachUserConversation", ctx, mock.Anything, pgtype.Text{String: userID, Status: pgtype.Present}, pgtype.Text{String: pb.CONVERSATION_STATUS_NONE.String(), Status: pgtype.Present}, uint(10), mock.Anything, database.TextArray([]string{"location-1"}), true).
					Once().Return([]*entities.Message{
					{
						ConversationID: conversationID,
						UserID:         pgtype.Text{String: userID, Status: 2},
						Message:        pgtype.Text{String: "content", Status: 2},
						CreatedAt:      pgtype.Timestamptz{Time: time.Unix(now.Seconds, int64(now.Nanos)), Status: 2},
					},
				}, nil)

				conversations := map[pgtype.Text]core.ConversationFull{
					conversationID: {
						Conversation: core.Conversation{
							ID:        conversationID,
							CreatedAt: pgtype.Timestamptz{},
							UpdatedAt: pgtype.Timestamptz{},
						},
					},
				}

				conversationRepo.On("FindByIDsReturnMapByID", ctx, mock.Anything, database.TextArray([]string{conversationID.String})).Once().Return(conversations, nil)

				conversationMembers := map[pgtype.Text][]*core.ConversationMembers{
					conversationID: {
						{
							ID:             pgtype.Text{},
							UserID:         pgtype.Text{String: userID, Status: 2},
							ConversationID: pgtype.Text{},
							Role:           pgtype.Text{String: core.ConversationRoleParent, Status: 2},
							Status:         pgtype.Text{String: core.ConversationStatusActive, Status: 2},
							SeenAt:         pgtype.Timestamptz{Time: time.Now(), Status: 2},
							LastNotifyAt:   pgtype.Timestamptz{},
							CreatedAt:      pgtype.Timestamptz{},
							UpdatedAt:      pgtype.Timestamptz{},
						},
					},
				}
				conversationMemberRepo.On("FindByConversationIDs", ctx, mock.Anything, mock.Anything).Once().Return(conversationMembers, nil)

				conversationStudents := map[pgtype.Text]*sentities.ConversationStudent{
					conversationID: {
						ConversationID: conversationID,
						StudentID:      database.Text(studentID),
					},
				}

				conversationStudentRepo.On("FindByConversationIDs", ctx, mock.Anything, mock.Anything).Once().Return(conversationStudents, nil)
			},
		},
		{
			name:      "1 allowed and 1 not found locations",
			ctx:       ctx,
			customCtx: customParentCtx,
			req:       &pb.ConversationListRequest{Limit: 10, EndAt: types.TimestampNow()},
			expectedResp: &pb.ConversationListResponse{
				Conversations: []*pb.Conversation{
					{
						ConversationId: cID,
						StudentId:      studentID,
						GuestIds:       nil,
						Seen:           true,
						LastMessage: &pb.MessageResponse{
							ConversationId: cID,
							MessageId:      "",
							UserId:         userID,
							Content:        "content",
							UrlMedia:       "",
							Type:           0,
							CreatedAt:      now,
							LocalMessageId: "",
						},
						Users: []*pb.Conversation_User{
							{
								Id:        userID,
								Group:     core.ConversationRoleParent,
								IsPresent: true,
							},
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationID := pgtype.Text{}
				_ = conversationID.Set(cID)

				conversationLocationRepo.On("GetAllLocations", ctx, mock.Anything, userID).
					Once().
					Return(
						[]*core.ConversationLocation{
							{
								ConversationID: conversationID,
								LocationID:     database.Text("location-1"),
								AccessPath:     database.Text("root-id/parent-id/location-1"),
							},
							{
								ConversationID: conversationID,
								LocationID:     database.Text("location-2"),
								AccessPath:     database.Text("not-found-root-id/parent-id/location-2"),
							},
						}, nil,
					)
				externalConfigurationServiceMock.getConfigurationByKeysAndLocationsV2 = func(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
					assert.ElementsMatch(t, in.Keys, []string{tom_const.ChatConfigKeyParentV2})
					assert.ElementsMatch(t, in.LocationIds, []string{"location-1", "location-2"})
					return &mpb.GetConfigurationByKeysAndLocationsV2Response{
						Configurations: []*mpb.LocationConfiguration{
							{
								Id:              "id-1",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-1",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id-2",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-2",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
						},
					}, nil
				}

				messageRepo.On("GetLastMessageEachUserConversation", ctx, mock.Anything, pgtype.Text{String: userID, Status: pgtype.Present}, pgtype.Text{String: pb.CONVERSATION_STATUS_NONE.String(), Status: pgtype.Present}, uint(10), mock.Anything, database.TextArray([]string{"location-1", "location-2"}), true).
					Once().Return([]*entities.Message{
					{
						ConversationID: conversationID,
						UserID:         pgtype.Text{String: userID, Status: 2},
						Message:        pgtype.Text{String: "content", Status: 2},
						CreatedAt:      pgtype.Timestamptz{Time: time.Unix(now.Seconds, int64(now.Nanos)), Status: 2},
					},
				}, nil)

				conversations := map[pgtype.Text]core.ConversationFull{
					conversationID: {
						Conversation: core.Conversation{
							ID:        conversationID,
							CreatedAt: pgtype.Timestamptz{},
							UpdatedAt: pgtype.Timestamptz{},
						},
					},
				}

				conversationRepo.On("FindByIDsReturnMapByID", ctx, mock.Anything, database.TextArray([]string{conversationID.String})).Once().Return(conversations, nil)

				conversationMembers := map[pgtype.Text][]*core.ConversationMembers{
					conversationID: {
						{
							ID:             pgtype.Text{},
							UserID:         pgtype.Text{String: userID, Status: 2},
							ConversationID: pgtype.Text{},
							Role:           pgtype.Text{String: core.ConversationRoleParent, Status: 2},
							Status:         pgtype.Text{String: core.ConversationStatusActive, Status: 2},
							SeenAt:         pgtype.Timestamptz{Time: time.Now(), Status: 2},
							LastNotifyAt:   pgtype.Timestamptz{},
							CreatedAt:      pgtype.Timestamptz{},
							UpdatedAt:      pgtype.Timestamptz{},
						},
					},
				}
				conversationMemberRepo.On("FindByConversationIDs", ctx, mock.Anything, mock.Anything).Once().Return(conversationMembers, nil)

				conversationStudents := map[pgtype.Text]*sentities.ConversationStudent{
					conversationID: {
						ConversationID: conversationID,
						StudentID:      database.Text(studentID),
					},
				}

				conversationStudentRepo.On("FindByConversationIDs", ctx, mock.Anything, mock.Anything).Once().Return(conversationStudents, nil)
			},
		},
		{
			name:      "all not allowed locations",
			ctx:       ctx,
			customCtx: customParentCtx,
			req:       &pb.ConversationListRequest{Limit: 10, EndAt: types.TimestampNow()},
			expectedResp: &pb.ConversationListResponse{
				Conversations: []*pb.Conversation{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				conversationID := pgtype.Text{}
				_ = conversationID.Set(cID)

				conversationLocationRepo.On("GetAllLocations", ctx, mock.Anything, userID).
					Once().
					Return(
						[]*core.ConversationLocation{
							{
								ConversationID: conversationID,
								LocationID:     database.Text("location-1"),
								AccessPath:     database.Text("not-allow-root-id-1/parent-id/location-1"),
							},
							{
								ConversationID: conversationID,
								LocationID:     database.Text("location-2"),
								AccessPath:     database.Text("not-allow-root-id-2/parent-id/location-2"),
							},
						}, nil,
					)
				externalConfigurationServiceMock.getConfigurationByKeysAndLocationsV2 = func(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
					assert.ElementsMatch(t, in.Keys, []string{tom_const.ChatConfigKeyParentV2})
					assert.ElementsMatch(t, in.LocationIds, []string{"location-1", "location-2"})
					return &mpb.GetConfigurationByKeysAndLocationsV2Response{
						Configurations: []*mpb.LocationConfiguration{
							{
								Id:              "id-1",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-1",
								ConfigValue:     "false",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id-2",
								ConfigKey:       tom_const.ChatConfigKeyParentV2,
								LocationId:      "location-2",
								ConfigValue:     "false",
								ConfigValueType: "boolean",
							},
						},
					}, nil
				}

				messageRepo.On("GetLastMessageEachUserConversation", ctx, mock.Anything, pgtype.Text{String: userID, Status: pgtype.Present}, pgtype.Text{String: pb.CONVERSATION_STATUS_NONE.String(), Status: pgtype.Present}, uint(10), mock.Anything, database.TextArray([]string{}), true).
					Once().Return([]*entities.Message{}, nil)

				conversationRepo.On("FindByIDsReturnMapByID", ctx, mock.Anything, database.TextArray([]string{})).Once().Return(map[pgtype.Text]core.ConversationFull{}, nil)
				conversationMemberRepo.On("FindByConversationIDs", ctx, mock.Anything, mock.Anything).Once().Return(map[pgtype.Text][]*core.ConversationMembers{}, nil)
				conversationStudentRepo.On("FindByConversationIDs", ctx, mock.Anything, mock.Anything).Once().Return(map[pgtype.Text]*sentities.ConversationStudent{}, nil)
			},
		},
		{
			name:         "could not call to external config service",
			ctx:          ctx,
			customCtx:    customParentCtx,
			req:          &pb.ConversationListRequest{Limit: 10, EndAt: types.TimestampNow()},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "GetConfigurationByKeysAndLocationsV2: Error"),
			setup: func(ctx context.Context) {
				conversationID := pgtype.Text{}
				_ = conversationID.Set(cID)

				conversationLocationRepo.On("GetAllLocations", ctx, mock.Anything, userID).
					Once().
					Return(
						[]*core.ConversationLocation{
							{
								ConversationID: conversationID,
								LocationID:     database.Text("location-1"),
								AccessPath:     database.Text("root-id/parent-id/location-1"),
							},
							{
								ConversationID: conversationID,
								LocationID:     database.Text("location-2"),
								AccessPath:     database.Text("root-id/parent-id/location-2"),
							},
						}, nil,
					)
				externalConfigurationServiceMock.getConfigurationByKeysAndLocationsV2 = func(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
					assert.ElementsMatch(t, in.Keys, []string{tom_const.ChatConfigKeyParentV2})
					assert.ElementsMatch(t, in.LocationIds, []string{"location-1", "location-2"})
					return nil, fmt.Errorf("Error")
				}

				messageRepo.On("GetLastMessageEachUserConversation", ctx, mock.Anything, pgtype.Text{String: userID, Status: pgtype.Present}, pgtype.Text{String: pb.CONVERSATION_STATUS_NONE.String(), Status: pgtype.Present}, uint(10), mock.Anything, database.TextArray([]string{"location-1", "location-2"}), true).
					Once().Return([]*entities.Message{
					{
						ConversationID: conversationID,
						UserID:         pgtype.Text{String: userID, Status: 2},
						Message:        pgtype.Text{String: "content", Status: 2},
						CreatedAt:      pgtype.Timestamptz{Time: time.Unix(now.Seconds, int64(now.Nanos)), Status: 2},
					},
				}, nil)

				conversations := map[pgtype.Text]core.ConversationFull{
					conversationID: {
						Conversation: core.Conversation{
							ID:        conversationID,
							CreatedAt: pgtype.Timestamptz{},
							UpdatedAt: pgtype.Timestamptz{},
						},
					},
				}

				conversationRepo.On("FindByIDsReturnMapByID", ctx, mock.Anything, database.TextArray([]string{conversationID.String})).Once().Return(conversations, nil)

				conversationMembers := map[pgtype.Text][]*core.ConversationMembers{
					conversationID: {
						{
							ID:             pgtype.Text{},
							UserID:         pgtype.Text{String: userID, Status: 2},
							ConversationID: pgtype.Text{},
							Role:           pgtype.Text{String: core.ConversationRoleParent, Status: 2},
							Status:         pgtype.Text{String: core.ConversationStatusActive, Status: 2},
							SeenAt:         pgtype.Timestamptz{Time: time.Now(), Status: 2},
							LastNotifyAt:   pgtype.Timestamptz{},
							CreatedAt:      pgtype.Timestamptz{},
							UpdatedAt:      pgtype.Timestamptz{},
						},
					},
				}
				conversationMemberRepo.On("FindByConversationIDs", ctx, mock.Anything, mock.Anything).Once().Return(conversationMembers, nil)

				conversationStudents := map[pgtype.Text]*sentities.ConversationStudent{
					conversationID: {
						ConversationID: conversationID,
						StudentID:      database.Text(studentID),
					},
				}

				conversationStudentRepo.On("FindByConversationIDs", ctx, mock.Anything, mock.Anything).Once().Return(conversationStudents, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.ctx = interceptors.NewIncomingContext(testCase.ctx)
			if testCase.customCtx != nil {
				testCase.ctx = testCase.customCtx(testCase.ctx)
			}
			testCase.setup(testCase.ctx)
			resp, err := s.ConversationList(testCase.ctx, testCase.req.(*pb.ConversationListRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestChatServiceReader_ListConversationsInSchoolWithLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	messageRepo := new(mock_repositories.MockMessageRepo)
	conversationRepo := new(mock_repositories.MockConversationRepo)
	conversationMemberRepo := new(mock_repositories.MockConversationMemberRepo)
	conversationStudentRepo := new(mock_repositories.MockConversationStudentRepo)
	// convLocationRepo := new(mock_repositories.MockConversationLocationRepo)
	locationRepo := new(mock_repositories.MockLocationRepo)
	searchRepo := new(mock_repositories.MockSearchRepo)
	externalConfigurationServiceMock := new(ExternalConfigurationServiceMock)
	locationConfigResolver := new(mock_support.MockLocationConfigResolver)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	db := &mock_database.Ext{}
	mockEs := &mock_elastic.SearchFactory{}
	s := &ChatReader{
		DB:                           db,
		SearchClient:                 mockEs,
		MessageRepo:                  messageRepo,
		ConversationMemberRepo:       conversationMemberRepo,
		ConversationRepo:             conversationRepo,
		ConversationSearchRepo:       searchRepo,
		ConversationStudentRepo:      conversationStudentRepo,
		LocationRepo:                 locationRepo,
		ExternalConfigurationService: externalConfigurationServiceMock,
		LocationConfigResolver:       locationConfigResolver,
		Env:                          "env",
		UnleashClientIns:             mockUnleashClient,
	}
	// map[pgtype.Text]core.ConversationFull
	userID := idutil.ULIDNow()
	studentID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	cID := idutil.ULIDNow()

	seenAt := time.Now()
	conversationMap := map[pgtype.Text]core.ConversationFull{
		database.Text(cID): {
			Conversation: core.Conversation{
				ID: database.Text(cID),
			},
		},
	}
	conversationMemberMap := map[pgtype.Text][]*core.ConversationMembers{
		database.Text(cID): {
			{
				UserID:         database.Text(userID),
				ConversationID: database.Text(cID),
				SeenAt:         database.Timestamptz(seenAt),
			},
		},
	}
	conversationStudentMap := map[pgtype.Text]*sentities.ConversationStudent{
		database.Text(cID): {
			ConversationID: database.Text(cID),
			StudentID:      database.Text(studentID),
		},
	}

	claims := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			SchoolIDs: []string{"1"},
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claims)
	lastOffsetTime := time.Now()
	sampleDoc := []sentities.SearchConversationDoc{
		{
			ConversationID:          cID,
			ConversationNameEnglish: "fake conversation name",
			UserIDs:                 []string{userID},
			LastMessage: sentities.SearchLastMessage{
				UpdatedAt: lastOffsetTime,
			},
			IsReplied: true,
		},
	}
	locationIDs := []string{"loc1", "loc2"}
	accessPath := []string{"org/loc1", "org/loca/loc2"}

	// TODO: add detail test + add test
	testCases := []TestCase{
		{
			env:  constants.StagingEnv,
			name: "stag - no config found",
			ctx:  ctx,
			req: &tpb.ListConversationsInSchoolRequest{
				Name: wrapperspb.String("fake name"),
				Type: []tpb.ConversationType{
					tpb.ConversationType_CONVERSATION_STUDENT,
				},
				TeacherStatus: tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_ALL,
				JoinStatus:    tpb.ConversationJoinStatus_CONVERSATION_JOIN_STATUS_JOINED,
				CourseIds:     []string{"c1"},
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetString{},
				},
				LocationIds: locationIDs,
			},
			expectedResp: &tpb.ListConversationsInSchoolResponse{
				Items: []*tpb.Conversation{},
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetCombined{
						OffsetCombined: &cpb.Paging_Combined{
							OffsetString:  "",
							OffsetInteger: int64(0),
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", tom_const.ChatThreadLaunchingFeatureFlag, mock.Anything).Once().Return(
					true, nil,
				)

				conversationID := pgtype.Text{}
				_ = conversationID.Set(cID)
				conversationTypeWithLocationMap := make(map[tpb.ConversationType][]string, 0)
				locationConfigResolver.On("GetEnabledLocationConfigsByLocations", mock.Anything, locationIDs, []tpb.ConversationType{tpb.ConversationType_CONVERSATION_STUDENT}).Once().Return(
					conversationTypeWithLocationMap,
					nil,
				)
				searchRepo.On("SearchV2", mock.Anything, mock.Anything, mock.MatchedBy(func(arg sentities.ConversationFilter) bool {
					expect := sentities.ConversationFilter{
						UserID:           userID,
						JoinStatus:       golibtype.NewBool(true),
						School:           golibtype.NewStrArr([]string{"1"}),
						ConversationName: golibtype.NewStr("fake name"),
						Courses:          golibtype.NewStrArr([]string{"c1"}),
						SortBy:           sentities.DefaultConversationSorts,
						Limit:            golibtype.NewInt64(10),
						LocationConfigs: []sentities.LocationConfigFilter{
							{
								ConversationType: golibtype.NewStr(tpb.ConversationType_CONVERSATION_STUDENT.String()),
								AccessPaths:      golibtype.NewStrArr(accessPath),
							},
						},
					}
					return cmp.Equal(expect, arg)
				})).Once().Return(sampleDoc, nil)
				messageRepo.On("GetLastMessageByConversationIDs", mock.Anything, mock.Anything,
					database.TextArray([]string{cID}), uint(10), mock.Anything, false).Once().Return(nil, nil)
				conversationRepo.On("FindByIDsReturnMapByID", mock.Anything, mock.Anything, database.TextArray([]string{cID})).
					Once().Return(conversationMap, nil)
				conversationMemberRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray([]string{cID})).
					Once().Return(conversationMemberMap, nil)
				conversationStudentRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray([]string{cID})).
					Once().Return(conversationStudentMap, nil)
			},
		},
		{
			env:  constants.StagingEnv,
			name: "stag - student type",
			ctx:  ctx,
			req: &tpb.ListConversationsInSchoolRequest{
				Name: wrapperspb.String("fake name"),
				Type: []tpb.ConversationType{
					tpb.ConversationType_CONVERSATION_STUDENT,
				},
				TeacherStatus: tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_ALL,
				JoinStatus:    tpb.ConversationJoinStatus_CONVERSATION_JOIN_STATUS_JOINED,
				CourseIds:     []string{"c1"},
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetString{},
				},
				LocationIds: locationIDs,
			},
			expectedResp: &tpb.ListConversationsInSchoolResponse{
				Items: []*tpb.Conversation{
					{
						ConversationId: cID,
						StudentId:      studentID,
						Users: []*tpb.Conversation_User{
							{
								Id:     userID,
								SeenAt: timestamppb.New(seenAt),
							},
						},
					},
				},
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetCombined{
						OffsetCombined: &cpb.Paging_Combined{
							OffsetString:  cID,
							OffsetInteger: lastOffsetTime.UnixMilli(),
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", tom_const.ChatThreadLaunchingFeatureFlag, mock.Anything).Once().Return(
					true, nil,
				)

				conversationID := pgtype.Text{}
				_ = conversationID.Set(cID)
				conversationTypeWithLocationMap := map[tpb.ConversationType][]string{
					tpb.ConversationType_CONVERSATION_STUDENT: accessPath,
				}
				locationConfigResolver.On("GetEnabledLocationConfigsByLocations", mock.Anything, locationIDs, []tpb.ConversationType{tpb.ConversationType_CONVERSATION_STUDENT}).Once().Return(
					conversationTypeWithLocationMap,
					nil,
				)
				searchRepo.On("SearchV2", mock.Anything, mock.Anything, mock.MatchedBy(func(arg sentities.ConversationFilter) bool {
					expect := sentities.ConversationFilter{
						UserID:           userID,
						JoinStatus:       golibtype.NewBool(true),
						School:           golibtype.NewStrArr([]string{"1"}),
						ConversationName: golibtype.NewStr("fake name"),
						Courses:          golibtype.NewStrArr([]string{"c1"}),
						SortBy:           sentities.DefaultConversationSorts,
						Limit:            golibtype.NewInt64(10),
						LocationConfigs: []sentities.LocationConfigFilter{
							{
								ConversationType: golibtype.NewStr(tpb.ConversationType_CONVERSATION_STUDENT.String()),
								AccessPaths:      golibtype.NewStrArr(accessPath),
							},
						},
					}
					return cmp.Equal(expect, arg)
				})).Once().Return(sampleDoc, nil)
				messageRepo.On("GetLastMessageByConversationIDs", mock.Anything, mock.Anything,
					database.TextArray([]string{cID}), uint(10), mock.Anything, false).Once().Return(nil, nil)
				conversationRepo.On("FindByIDsReturnMapByID", mock.Anything, mock.Anything, database.TextArray([]string{cID})).
					Once().Return(conversationMap, nil)
				conversationMemberRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray([]string{cID})).
					Once().Return(conversationMemberMap, nil)
				conversationStudentRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray([]string{cID})).
					Once().Return(conversationStudentMap, nil)
			},
		},
		{
			env:  constants.StagingEnv,
			name: "stag - parent type",
			ctx:  ctx,
			req: &tpb.ListConversationsInSchoolRequest{
				Name: wrapperspb.String("fake name"),
				Type: []tpb.ConversationType{
					tpb.ConversationType_CONVERSATION_PARENT,
				},
				TeacherStatus: tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_ALL,
				JoinStatus:    tpb.ConversationJoinStatus_CONVERSATION_JOIN_STATUS_JOINED,
				CourseIds:     []string{"c1"},
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetString{},
				},
				LocationIds: locationIDs,
			},
			expectedResp: &tpb.ListConversationsInSchoolResponse{
				Items: []*tpb.Conversation{
					{
						ConversationId: cID,
						StudentId:      studentID,
						Users: []*tpb.Conversation_User{
							{
								Id:     userID,
								SeenAt: timestamppb.New(seenAt),
							},
						},
					},
				},
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetCombined{
						OffsetCombined: &cpb.Paging_Combined{
							OffsetString:  cID,
							OffsetInteger: lastOffsetTime.UnixMilli(),
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", tom_const.ChatThreadLaunchingFeatureFlag, mock.Anything).Once().Return(
					true, nil,
				)

				conversationID := pgtype.Text{}
				_ = conversationID.Set(cID)
				conversationTypeWithLocationMap := map[tpb.ConversationType][]string{
					tpb.ConversationType_CONVERSATION_PARENT: accessPath,
				}
				locationConfigResolver.On("GetEnabledLocationConfigsByLocations", mock.Anything, locationIDs, []tpb.ConversationType{tpb.ConversationType_CONVERSATION_PARENT}).Once().Return(
					conversationTypeWithLocationMap,
					nil,
				)
				searchRepo.On("SearchV2", mock.Anything, mock.Anything, mock.MatchedBy(func(arg sentities.ConversationFilter) bool {
					expect := sentities.ConversationFilter{
						UserID:           userID,
						JoinStatus:       golibtype.NewBool(true),
						School:           golibtype.NewStrArr([]string{"1"}),
						ConversationName: golibtype.NewStr("fake name"),
						Courses:          golibtype.NewStrArr([]string{"c1"}),
						SortBy:           sentities.DefaultConversationSorts,
						Limit:            golibtype.NewInt64(10),
						LocationConfigs: []sentities.LocationConfigFilter{
							{
								ConversationType: golibtype.NewStr(tpb.ConversationType_CONVERSATION_PARENT.String()),
								AccessPaths:      golibtype.NewStrArr(accessPath),
							},
						},
					}
					return cmp.Equal(expect, arg)
				})).Once().Return([]sentities.SearchConversationDoc{
					{
						ConversationID:          cID,
						ConversationNameEnglish: "fake conversation name",
						UserIDs:                 []string{userID},
						LastMessage: sentities.SearchLastMessage{
							UpdatedAt: lastOffsetTime,
						},
						IsReplied: true,
					},
				}, nil)
				messageRepo.On("GetLastMessageByConversationIDs", mock.Anything, mock.Anything,
					database.TextArray([]string{cID}), uint(10), mock.Anything, false).Once().Return(nil, nil)
				conversationRepo.On("FindByIDsReturnMapByID", mock.Anything, mock.Anything, database.TextArray([]string{cID})).
					Once().Return(conversationMap, nil)
				conversationMemberRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray([]string{cID})).
					Once().Return(conversationMemberMap, nil)
				conversationStudentRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray([]string{cID})).
					Once().Return(conversationStudentMap, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			s.Env = ""
			if testCase.env != "" {
				s.Env = testCase.env
			}
			testCase.setup(testCase.ctx)
			testCase.ctx = interceptors.NewIncomingContext(testCase.ctx)
			resp, err := s.ListConversationsInSchoolWithLocations(testCase.ctx, testCase.req.(*tpb.ListConversationsInSchoolRequest))
			if err != nil {
				assert.ErrorIs(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
