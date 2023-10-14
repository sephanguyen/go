package support

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/types"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/tom/app"
	tom_const "github.com/manabie-com/backend/internal/tom/constants"
	"github.com/manabie-com/backend/internal/tom/domain/core"
	sentities "github.com/manabie-com/backend/internal/tom/domain/support"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	supportChatTypes = []string{tpb.ConversationType_CONVERSATION_PARENT.String(), tpb.ConversationType_CONVERSATION_STUDENT.String()}
)

type ChatReader struct {
	Logger           *zap.Logger
	DB               database.Ext
	SearchClient     elastic.SearchFactory
	UnleashClientIns unleashclient.ClientInstance
	Env              string

	LocationRepo interface {
		FindAccessPaths(ctx context.Context, db database.Ext, locationIDs []string) ([]string, error)
		FindRootIDs(ctx context.Context, db database.Ext) ([]string, error)
		FindLowestAccessPathByLocationIDs(ctx context.Context, db database.Ext, locationIDs []string) ([]string, map[string]string, error)
	}

	ConversationMemberRepo interface {
		FindByConversationID(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text) (mapUserID map[pgtype.Text]core.ConversationMembers, err error)
		FindByConversationIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray) (mapConversationID map[pgtype.Text][]*core.ConversationMembers, err error)
	}
	ConversationSearchRepo interface {
		Search(ctx context.Context, cl elastic.SearchFactory, f sentities.ConversationFilter) ([]sentities.SearchConversationDoc, error)
		SearchV2(ctx context.Context, cl elastic.SearchFactory, f sentities.ConversationFilter) ([]sentities.SearchConversationDoc, error)
	}
	ConversationRepo interface {
		FindByID(context.Context, database.QueryExecer, pgtype.Text) (*core.Conversation, error)
		FindByIDsReturnMapByID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (map[pgtype.Text]core.ConversationFull, error)
		ListAll(ctx context.Context, db database.QueryExecer, offsetID pgtype.Text, limit uint32, conversationTypesAccepted pgtype.TextArray, resourcePath pgtype.Text) ([]*core.Conversation, error)
		CountUnreadConversations(ctx context.Context, db database.QueryExecer, userID pgtype.Text, conversationType pgtype.TextArray, locationIDs pgtype.TextArray, enableChatThreadLaunching bool) (int64, error)
		CountUnreadConversationsByAccessPaths(ctx context.Context, db database.QueryExecer, userID pgtype.Text, conversationType pgtype.TextArray, accessPath pgtype.TextArray) (int64, error)
		CountUnreadConversationsByAccessPathsV2(ctx context.Context, db database.QueryExecer, userID pgtype.Text, studentAccessPaths pgtype.TextArray, parentAccessPaths pgtype.TextArray) (int64, error)
	}
	MessageRepo interface {
		GetLastMessageByConversationIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray, limit uint, endAt pgtype.Timestamptz, includeSystemMsg bool) ([]*core.Message, error)
		Create(context.Context, database.QueryExecer, *core.Message) error
		FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (c *core.Message, err error)
		GetLastMessageEachUserConversation(ctx context.Context, db database.QueryExecer, userID, status pgtype.Text, limit uint, endAt pgtype.Timestamptz, locationIDs pgtype.TextArray, enableChatThreadLaunching bool) ([]*core.Message, error)
		FindAllMessageByConversation(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, limit uint, endAt pgtype.Timestamptz) ([]*core.Message, error)
		CountMessagesSince(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text, since *pgtype.Timestamptz) (int, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, userID, id pgtype.Text) error
		FindMessages(ctx context.Context, db database.QueryExecer, args *core.FindMessagesArgs) ([]*core.Message, error)
	}
	ConversationStudentRepo interface {
		FindByConversationIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray) (map[pgtype.Text]*sentities.ConversationStudent, error)
		FindByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray, conversationType pgtype.Text) ([]string, error)
	}

	ConversationLocationRepo interface {
		GetAllLocations(ctx context.Context, db database.QueryExecer, userID string) ([]*core.ConversationLocation, error)
	}

	ExternalConfigurationService interface {
		GetConfigurationByKeysAndLocations(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsRequest, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsResponse, error)
		GetConfigurationByKeysAndLocationsV2(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error)
	}

	LocationConfigResolver interface {
		// for Chat Thread Launching Phase 2
		GetEnabledLocationConfigsByLocations(ctx context.Context, locationIDs []string, conversationTypes []tpb.ConversationType) (map[tpb.ConversationType][]string, error)

		// for Chat Thread Launching Phase 1
		GetEnabledLocationConfigsByOrg(ctx context.Context, locationIDs []string) ([]tpb.ConversationType, []string, error)
	}
}

func (rcv *ChatReader) ConversationList(ctx context.Context, req *pb.ConversationListRequest) (*pb.ConversationListResponse, error) {
	conversations, err := rcv.getLastMessageEachSupportConversation(ctx, uint(req.Limit), req.EndAt)
	if err != nil {
		return nil, err
	}

	return &pb.ConversationListResponse{
		Conversations: conversations,
	}, nil
}

// HACK: initially, will return all location ids if there is an error when calling ExternalConfigurationService
// plz consider to remove this later
func (rcv *ChatReader) getAllowedLocationIDs(ctx context.Context, userID string, key string) ([]string, error) {
	// get conversation's locations list
	locations, err := rcv.ConversationLocationRepo.GetAllLocations(ctx, rcv.DB, userID)
	if err != nil {
		return nil, fmt.Errorf("ConversationLocationRepo.GetAllLocations: %w", err)
	}

	// call external configuration service to get configs by keys and locations
	mdCtx, err := interceptors.GetOutgoingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetOutgoingContext: %v", err)
	}

	enableChatThreadLaunching, err := rcv.UnleashClientIns.IsFeatureEnabled(tom_const.ChatThreadLaunchingFeatureFlag, rcv.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if enableChatThreadLaunching {
		locationIds := []string{}
		for _, loc := range locations {
			locationIds = append(locationIds, loc.LocationID.String)
		}

		res, err := rcv.ExternalConfigurationService.GetConfigurationByKeysAndLocationsV2(mdCtx, &mpb.GetConfigurationByKeysAndLocationsV2Request{
			Keys:        []string{key},
			LocationIds: locationIds,
		})
		if err != nil {
			return nil, fmt.Errorf("GetConfigurationByKeysAndLocationsV2: %v", err)
		}

		// filter list allow location ids
		allowIDs := make([]string, 0)
		// first check in config
		for _, config := range res.Configurations {
			boolValue, err := strconv.ParseBool(config.ConfigValue)
			if err != nil {
				boolValue = false
			}
			if boolValue {
				allowIDs = append(allowIDs, config.LocationId)
			}
		}

		return allowIDs, nil
	}

	// list all root path of above locations
	rootIDsMap := make(map[string][]string)
	for _, loc := range locations {
		s := strings.SplitN(loc.AccessPath.String, "/", 2)
		if len(s) > 0 {
			rootIDsMap[s[0]] = append(rootIDsMap[s[0]], loc.LocationID.String)
		}
	}
	// case: student have been created on org level
	if len(locations) == 0 {
		rootIDs, err := rcv.LocationRepo.FindRootIDs(ctx, rcv.DB)
		if err != nil {
			return nil, fmt.Errorf("LocationRepo.FindRootIDs: %w", err)
		}
		for _, id := range rootIDs {
			rootIDsMap[id] = append(rootIDsMap[id], id)
		}
	}
	rootIDs := make([]string, 0, len(rootIDsMap))
	for k := range rootIDsMap {
		rootIDs = append(rootIDs, k)
	}

	res, err := rcv.ExternalConfigurationService.GetConfigurationByKeysAndLocations(mdCtx, &mpb.GetConfigurationByKeysAndLocationsRequest{
		Keys:         []string{key},
		LocationsIds: rootIDs,
	})
	if err != nil {
		// TODO: plz consider to return error later
		ctxzap.Extract(ctx).Warn("ExternalConfigurationService.GetConfigurationByKeysAndLocations:", zap.Error(err))
		allowIDs := make([]string, 0, len(locations))
		for _, loc := range locations {
			allowIDs = append(allowIDs, loc.LocationID.String)
		}
		return allowIDs, nil
	}

	// filter list allow location ids
	allowIDs := make([]string, 0)
	// first check in config
	for _, config := range res.Configurations {
		boolValue, err := strconv.ParseBool(config.ConfigValue)
		if err != nil {
			boolValue = false
		}
		if boolValue {
			allowIDs = append(allowIDs, rootIDsMap[config.LocationId]...)
		}
		delete(rootIDsMap, config.LocationId)
	}
	// if not exist in config, default is allowed
	for _, v := range rootIDsMap {
		allowIDs = append(allowIDs, v...)
	}

	return allowIDs, nil
}

// TODO: rename Support conversation == student conversation
func (rcv *ChatReader) getLastMessageEachSupportConversation(ctx context.Context, limit uint, endAt *pbtypes.Timestamp) ([]*pb.Conversation, error) {
	enableChatThreadLaunching, err := rcv.UnleashClientIns.IsFeatureEnabled(tom_const.ChatThreadLaunchingFeatureFlag, rcv.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var isParent bool
	var key string
	userGroup, userID, _ := interceptors.GetUserInfoFromContext(ctx)
	studentLocationConfigKey, parentLocationConfigKey := app.GetLocationConfigKeys(enableChatThreadLaunching)

	if userGroup == cpb.UserGroup_USER_GROUP_PARENT.String() {
		isParent = true
		key = parentLocationConfigKey
	} else {
		key = studentLocationConfigKey
	}
	logger := ctxzap.Extract(ctx)

	var pgEndAt pgtype.Timestamptz
	if endAt != nil {
		_ = pgEndAt.Set(time.Unix(endAt.Seconds, 0))
	} else {
		_ = pgEndAt.Set(time.Now())
	}

	allowedLocationIDs, err := rcv.getAllowedLocationIDs(ctx, userID, key)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var messages []*core.Message
	messages, err = rcv.MessageRepo.GetLastMessageEachUserConversation(ctx, rcv.DB, database.Text(userID), database.Text(pb.CONVERSATION_STATUS_NONE.String()), limit, pgEndAt, database.TextArray(allowedLocationIDs), enableChatThreadLaunching)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Error(codes.Unknown, fmt.Errorf("MessageRepo.GetLastMessageEachUserConversation: %w", err).Error())
	}

	// get all conversationID
	conversationIDs := make([]string, 0, len(messages))
	for _, message := range messages {
		conversationIDs = append(conversationIDs, message.ConversationID.String)
	}

	conversationMap, _ := rcv.ConversationRepo.FindByIDsReturnMapByID(ctx, rcv.DB, database.TextArray(conversationIDs))
	conversationStatusMap, _ := rcv.ConversationMemberRepo.FindByConversationIDs(ctx, rcv.DB, database.TextArray(conversationIDs))
	var resp = make([]*pb.Conversation, 0, len(messages))

	var mapConversationStudents map[pgtype.Text]*sentities.ConversationStudent

	// only query if current user is the parent
	if isParent {
		mapConversationStudents, err = rcv.ConversationStudentRepo.FindByConversationIDs(ctx, rcv.DB, database.TextArray(conversationIDs))
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("ConversationStudentRepo.FindByConversationIDs: %w", err).Error())
		}
	}

	for _, message := range messages {
		conversation := conversationMap[message.ConversationID]

		conversationMembers := conversationStatusMap[message.ConversationID]
		studentID, seen := getStudentIDAndSeenStatusFromConversationMembers(conversationMembers, message, userID)

		if studentID == "" && isParent {
			convStudent, ok := mapConversationStudents[message.ConversationID]
			if ok {
				studentID = convStudent.StudentID.String
			}
		}

		var users []*pb.Conversation_User
		for _, u := range conversationMembers {
			users = append(users, &pb.Conversation_User{
				Id:        u.UserID.String,
				Group:     u.Role.String,
				IsPresent: u.Status.String == core.ConversationStatusActive,
			})
		}

		resp = append(resp, &pb.Conversation{
			ConversationId:   message.ConversationID.String,
			ConversationName: conversation.Conversation.Name.String,
			StudentId:        studentID,
			GuestIds:         nil,
			Seen:             seen,
			LastMessage:      toMessageResponse(message),
			// StudentQuestionId: conversation.StudentQuestionID.String,
			Status: pb.ConversationStatus(pb.CodesMessageType_value[conversation.Conversation.Status.String]),
			Users:  users,
			// ClassId:           uint32(conversation.ClassID.Int),
			ConversationType: pb.ConversationType(pb.ConversationType_value[conversation.Conversation.ConversationType.String]),
		})
	}
	return resp, nil
}

func (rcv *ChatReader) ListConversationIDs(ctx context.Context, req *tpb.ListConversationIDsRequest) (*tpb.ListConversationIDsResponse, error) {
	var (
		conversationID pgtype.Text
		limit          uint32 = 100
	)
	// temporary, until rls is enabled for all env
	resourcePath, err := interceptors.ResourcePathFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	if len(resourcePath) == 0 {
		return nil, status.Error(codes.PermissionDenied, "nil resourcePath id in context")
	}
	_ = conversationID.Set(nil)
	if req.Paging.GetOffsetString() != "" {
		_ = conversationID.Set(req.Paging.GetOffsetString())
	}
	if req.GetPaging().GetLimit() > 100 {
		limit = req.GetPaging().Limit
	}
	conversationTypesAccepted := database.TextArray([]string{"CONVERSATION_STUDENT", "CONVERSATION_PARENT"})

	conversations, err := rcv.ConversationRepo.ListAll(ctx, rcv.DB, conversationID, limit, conversationTypesAccepted, database.Text(resourcePath))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("ConversationRepo.ListAll: %w", err).Error())
	}
	if len(conversations) == 0 {
		return &tpb.ListConversationIDsResponse{}, nil
	}
	return &tpb.ListConversationIDsResponse{
		ConversationIds: retrieveConversationIDsFromConversation(conversations),
		NextPage: &cpb.Paging{
			Limit: limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: conversations[len(conversations)-1].ID.String,
			},
		},
	}, nil
}

func (rcv *ChatReader) RetrieveTotalUnreadMessage(ctx context.Context, req *tpb.RetrieveTotalUnreadMessageRequest) (*tpb.RetrieveTotalUnreadMessageResponse, error) {
	enableChatThreadLaunching, err := rcv.UnleashClientIns.IsFeatureEnabled(tom_const.ChatThreadLaunchingFeatureFlag, rcv.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var key string
	userGroup, userID, _ := interceptors.GetUserInfoFromContext(ctx)
	if userID == "" {
		return nil, fmt.Errorf("empty userID")
	}

	studentLocationConfigKey, parentLocationConfigKey := app.GetLocationConfigKeys(enableChatThreadLaunching)
	if userGroup == cpb.UserGroup_USER_GROUP_PARENT.String() {
		key = parentLocationConfigKey
	} else {
		key = studentLocationConfigKey
	}

	allowedLocationIDs, err := rcv.getAllowedLocationIDs(ctx, userID, key)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	count, err := rcv.ConversationRepo.CountUnreadConversations(ctx, rcv.DB, database.Text(req.UserId), database.TextArray(supportChatTypes), database.TextArray(allowedLocationIDs), enableChatThreadLaunching)
	if err != nil {
		return nil, fmt.Errorf("s.MessageRepo.CountUnreadMessages: %w", err)
	}
	return &tpb.RetrieveTotalUnreadMessageResponse{
		TotalUnreadMessages: count,
	}, nil
}

func (rcv *ChatReader) RetrieveTotalUnreadConversationsWithLocations(ctx context.Context, req *tpb.RetrieveTotalUnreadConversationsWithLocationsRequest) (*tpb.RetrieveTotalUnreadConversationsWithLocationsResponse, error) {
	enableChatThreadLaunching, err := rcv.UnleashClientIns.IsFeatureEnabled(tom_const.ChatThreadLaunchingFeatureFlag, rcv.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	studentLocationConfigKey, parentLocationConfigKey := app.GetLocationConfigKeys(enableChatThreadLaunching)

	if enableChatThreadLaunching {
		userGroup, userID, _ := interceptors.GetUserInfoFromContext(ctx)
		if userID == "" {
			return nil, fmt.Errorf("empty userID")
		}

		if len(req.GetLocationIds()) == 0 && slices.Contains([]string{cpb.UserGroup_USER_GROUP_STUDENT.String(), cpb.UserGroup_USER_GROUP_PARENT.String()}, userGroup) {
			var key string
			if userGroup == cpb.UserGroup_USER_GROUP_PARENT.String() {
				key = parentLocationConfigKey
			} else {
				key = studentLocationConfigKey
			}

			allowedLocationIDs, err := rcv.getAllowedLocationIDs(ctx, userID, key)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			count, err := rcv.ConversationRepo.CountUnreadConversations(ctx, rcv.DB, database.Text(userID), database.TextArray(supportChatTypes), database.TextArray(allowedLocationIDs), enableChatThreadLaunching)
			if err != nil {
				return nil, fmt.Errorf("rcv.ConversationRepo.CountUnreadConversations %w", err)
			}
			return &tpb.RetrieveTotalUnreadConversationsWithLocationsResponse{
				TotalUnreadConversations: count,
			}, nil
		}

		conversationTypeWithLocationMap, err := rcv.LocationConfigResolver.GetEnabledLocationConfigsByLocations(ctx, req.GetLocationIds(), []tpb.ConversationType{tpb.ConversationType_CONVERSATION_STUDENT, tpb.ConversationType_CONVERSATION_PARENT})
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed GetEnabledLocationConfigsByLocations: %+v", err))
		}

		studentAps := conversationTypeWithLocationMap[tpb.ConversationType_CONVERSATION_STUDENT]
		parentAps := conversationTypeWithLocationMap[tpb.ConversationType_CONVERSATION_PARENT]

		count, err := rcv.ConversationRepo.CountUnreadConversationsByAccessPathsV2(
			ctx, rcv.DB, database.Text(userID), database.TextArray(studentAps), database.TextArray(parentAps))
		if err != nil {
			return nil, fmt.Errorf("s.ConversationRepo.CountUnreadConversationsByAccessPathsV2: %w", err)
		}
		return &tpb.RetrieveTotalUnreadConversationsWithLocationsResponse{
			TotalUnreadConversations: count,
		}, nil
	}

	userID := interceptors.UserIDFromContext(ctx)
	if userID == "" {
		return nil, fmt.Errorf("empty userID")
	}
	if len(req.GetLocationIds()) == 0 {
		count, err := rcv.ConversationRepo.CountUnreadConversations(ctx, rcv.DB, database.Text(userID), database.TextArray(supportChatTypes), database.TextArray([]string{}), enableChatThreadLaunching)
		if err != nil {
			return nil, fmt.Errorf("rcv.ConversationRepo.CountUnreadConversations %w", err)
		}
		return &tpb.RetrieveTotalUnreadConversationsWithLocationsResponse{
			TotalUnreadConversations: count,
		}, nil
	}
	aps, err := rcv.LocationRepo.FindAccessPaths(ctx, rcv.DB, req.GetLocationIds())
	if err != nil {
		return nil, fmt.Errorf("rcv.LocationRepo.FIndAccessPaths %w", err)
	}
	if len(aps) != len(req.GetLocationIds()) {
		return nil, fmt.Errorf("finding access path of locations %v returned %d items", req.GetLocationIds(), len(aps))
	}
	count, err := rcv.ConversationRepo.CountUnreadConversationsByAccessPaths(
		ctx, rcv.DB, database.Text(userID), database.TextArray(supportChatTypes), database.TextArray(aps))
	if err != nil {
		return nil, fmt.Errorf("s.MessageRepo.CountUnreadConversationsWithAccessPaths: %w", err)
	}
	return &tpb.RetrieveTotalUnreadConversationsWithLocationsResponse{
		TotalUnreadConversations: count,
	}, nil
}

// ListConversationByStudents list all conversations by students
// Flow:1. getConversationMembers by user_ids || conversation_ids
// 2. get conservations
// 3. get latest message
// convert -> full conversations
// some fields were omitted
func (rcv *ChatReader) ListConversationByUsers(ctx context.Context, req *tpb.ListConversationByUsersRequest) (*tpb.ListConversationByUsersResponse, error) {
	if len(req.GetUserIds()) == 0 && len(req.GetConversationIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty params")
	}
	var (
		mapConversationMembers map[pgtype.Text][]*core.ConversationMembers
		conversationIDs        []string
		err                    error
	)

	if len(req.GetConversationIds()) != 0 {
		mapConversationMembers, err = rcv.ConversationMemberRepo.FindByConversationIDs(ctx, rcv.DB, database.TextArray(req.GetConversationIds()))
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("ConversationMemberRepo.FindByConversationIDs: %w", err).Error())
		}
		conversationIDs = req.GetConversationIds()
	} else if len(req.GetUserIds()) != 0 {
		var conversationType pgtype.Text
		_ = conversationType.Set(nil)
		// check the conversation on conversation_students (include coversation type: parent && student)
		conversationIDs, err = rcv.ConversationStudentRepo.FindByStudentIDs(ctx, rcv.DB, database.TextArray(req.GetUserIds()), conversationType)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("ConversationStudentRepo.FindByStudentIDs: %w", err).Error())
		}
		mapConversationMembers, err = rcv.ConversationMemberRepo.FindByConversationIDs(ctx, rcv.DB, database.TextArray(conversationIDs))
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("ConversationMemberRepo.FindByConversationIDs: %w", err).Error())
		}
	}

	conversationMap, err := rcv.ConversationRepo.FindByIDsReturnMapByID(ctx, rcv.DB, database.TextArray(conversationIDs))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("ConversationRepo.FindByIDsReturnMapByID: %w", err).Error())
	}
	var pgEndAt pgtype.Timestamptz
	_ = pgEndAt.Set(time.Now())
	messages, err := rcv.MessageRepo.GetLastMessageByConversationIDs(ctx, rcv.DB, database.TextArray(conversationIDs), uint(len(conversationIDs)), pgEndAt, true)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("MessageRepo.GetLastMessageByConversationIDs: %w", err).Error())
	}
	messageMap := convertToMapMessages(messages)

	var resp = make([]*tpb.Conversation, 0, len(messages))
	for _, c := range conversationMap {
		conversationMembers := mapConversationMembers[c.Conversation.ID]
		var users = make([]*tpb.Conversation_User, 0, len(conversationMembers))
		for _, u := range conversationMembers {
			users = append(users, &tpb.Conversation_User{
				Id:        u.UserID.String,
				Group:     cpb.UserGroup(cpb.UserGroup_value[u.Role.String]),
				IsPresent: u.Status.String == core.ConversationStatusActive,
			})
		}
		var lastmessage *tpb.MessageResponse
		if m, ok := messageMap[c.Conversation.ID]; ok {
			lastmessage = toMessagePb(m)
		}
		resp = append(resp, &tpb.Conversation{
			ConversationId:   c.Conversation.ID.String,
			LastMessage:      lastmessage,
			Status:           tpb.ConversationStatus(tpb.CodesMessageType_value[c.Conversation.Status.String]),
			ConversationType: tpb.ConversationType(tpb.ConversationType_value[c.Conversation.ConversationType.String]),
			Users:            users,
			ConversationName: c.Conversation.Name.String,
			IsReplied:        c.IsReply.Bool,
			Owner:            c.Conversation.Owner.String,
			StudentId:        c.StudentID.String,
		})
	}
	return &tpb.ListConversationByUsersResponse{
		Items: resp,
	}, nil
}

func (rcv *ChatReader) requestFilterToRepoFilterV2(ctx context.Context, req *tpb.ListConversationsInSchoolRequest, locationConfigs map[tpb.ConversationType][]string) (ret sentities.ConversationFilter, err error) {
	_, userID, schoolIDs := interceptors.GetUserInfoFromContext(ctx)
	if len(schoolIDs) == 0 {
		return ret, fmt.Errorf("empty school ids from request")
	}
	ret.UserID = userID
	logger := ctxzap.Extract(ctx)
	courseIDs := req.CourseIds
	joinStatus := req.JoinStatus
	repliedStatus := req.TeacherStatus

	if repliedStatus != tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_ALL && repliedStatus != tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_NONE {
		ret.RepliedStatus = types.NewBool(repliedStatus == tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_REPLIED)
	}
	ret.School = types.NewStrArr(schoolIDs)

	if len(courseIDs) != 0 {
		ret.Courses = types.NewStrArr(courseIDs)
	}

	if joinStatus != tpb.ConversationJoinStatus_CONVERSATION_JOIN_STATUS_NONE {
		ret.JoinStatus = types.NewBool(joinStatus == tpb.ConversationJoinStatus_CONVERSATION_JOIN_STATUS_JOINED)
	}

	if req.Name != nil && req.Name.Value != "" {
		ret.ConversationName = types.NewStr(req.Name.Value)
	}

	if len(locationConfigs) > 0 {
		for convType, accessPath := range locationConfigs {
			ret.LocationConfigs = append(ret.LocationConfigs, sentities.LocationConfigFilter{
				ConversationType: types.NewStr(convType.String()),
				AccessPaths:      types.NewStrArr(accessPath),
			})
		}
	}

	if req.Paging != nil {
		ret.Limit = types.NewInt64(int64(req.Paging.Limit))
		if req.Paging.GetOffsetCombined() != nil {
			conversationID := req.Paging.GetOffsetCombined().GetOffsetString()
			lastMessageTime := req.Paging.GetOffsetCombined().GetOffsetInteger()
			ret.OffsetTime = types.NewInt64(lastMessageTime)
			ret.OffsetConverstionID = types.NewStr(conversationID)
			logger.Info("Paging: ", zap.Int64("last message time", lastMessageTime))
			logger.Info("Paging: ", zap.String("conversationID", conversationID))
		}
	}
	ret.SortBy = append(ret.SortBy,
		sentities.ConversationSortItem{
			Key: sentities.SortKey_LatestMsgTime,
			Asc: false,
		}, sentities.ConversationSortItem{
			Key: sentities.SortKey_ConversationID,
			Asc: true,
		})
	return
}

func (rcv *ChatReader) requestFilterToRepoFilter(ctx context.Context, req *tpb.ListConversationsInSchoolRequest, accessPaths []string, excludeTypes []tpb.ConversationType) (ret sentities.ConversationFilter, err error) {
	_, userID, schoolIDs := interceptors.GetUserInfoFromContext(ctx)
	if len(schoolIDs) == 0 {
		return ret, fmt.Errorf("empty school ids from request")
	}
	ret.UserID = userID
	logger := ctxzap.Extract(ctx)
	courseIDs := req.CourseIds
	joinStatus := req.JoinStatus
	repliedStatus := req.TeacherStatus

	if repliedStatus != tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_ALL && repliedStatus != tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_NONE {
		ret.RepliedStatus = types.NewBool(repliedStatus == tpb.TeacherConversationStatus_TEACHER_CONVERSATION_STATUS_REPLIED)
	}
	ret.School = types.NewStrArr(schoolIDs)

	if len(courseIDs) != 0 {
		ret.Courses = types.NewStrArr(courseIDs)
	}
	if joinStatus != tpb.ConversationJoinStatus_CONVERSATION_JOIN_STATUS_NONE {
		ret.JoinStatus = types.NewBool(joinStatus == tpb.ConversationJoinStatus_CONVERSATION_JOIN_STATUS_JOINED)
	}

	// Add all support conversation types when request have empty conversation types filter
	if len(req.Type) == 0 {
		req.Type = []tpb.ConversationType{
			tpb.ConversationType_CONVERSATION_STUDENT,
			tpb.ConversationType_CONVERSATION_PARENT,
		}
	}
	conversationTypes := make([]string, 0, len(req.Type))
	for _, ct := range req.Type {
		if slices.Contains(excludeTypes, ct) {
			continue
		}

		conversationTypes = append(conversationTypes, ct.String())
	}
	ret.ConversationTypes = types.NewStrArr(conversationTypes)

	if req.Name != nil && req.Name.Value != "" {
		ret.ConversationName = types.NewStr(req.Name.Value)
	}
	if len(accessPaths) > 0 {
		ret.AccessPaths = types.NewStrArr(accessPaths)
	}

	if req.Paging != nil {
		ret.Limit = types.NewInt64(int64(req.Paging.Limit))
		if req.Paging.GetOffsetCombined() != nil {
			conversationID := req.Paging.GetOffsetCombined().GetOffsetString()
			lastMessageTime := req.Paging.GetOffsetCombined().GetOffsetInteger()
			ret.OffsetTime = types.NewInt64(lastMessageTime)
			ret.OffsetConverstionID = types.NewStr(conversationID)
			logger.Info("Paging: ", zap.Int64("last message time", lastMessageTime))
			logger.Info("Paging: ", zap.String("conversationID", conversationID))
		}
	}
	ret.SortBy = append(ret.SortBy,
		sentities.ConversationSortItem{
			Key: sentities.SortKey_LatestMsgTime,
			Asc: false,
		}, sentities.ConversationSortItem{
			Key: sentities.SortKey_ConversationID,
			Asc: true,
		})
	return
}

func sortConversationsByIDs(conversations []*tpb.Conversation, conversationIDs []string) []*tpb.Conversation {
	conversationsMap := make(map[string]*tpb.Conversation)
	for _, conversation := range conversations {
		conversationsMap[conversation.ConversationId] = conversation
	}

	sortedConversations := make([]*tpb.Conversation, 0, len(conversations))
	for _, conversationID := range conversationIDs {
		conversation, ok := conversationsMap[conversationID]
		if ok {
			sortedConversations = append(sortedConversations, conversation)
		}
	}
	return sortedConversations
}

func (rcv *ChatReader) ListConversationsInSchoolWithLocations(ctx context.Context, req *tpb.ListConversationsInSchoolRequest) (*tpb.ListConversationsInSchoolResponse, error) {
	var filter sentities.ConversationFilter

	logger := ctxzap.Extract(ctx)
	featureEnabled, err := rcv.UnleashClientIns.IsFeatureEnabled(tom_const.ChatThreadLaunchingFeatureFlag, rcv.Env)
	if err != nil {
		return nil, fmt.Errorf("failed get UnleashClient: %+v", err)
	}
	useV2Search := false
	if featureEnabled {
		logger.Info("Communication_Chat_ChatThreadLaunching_Phase2 enabled")
		conversationTypeWithLocationMap, err := rcv.LocationConfigResolver.GetEnabledLocationConfigsByLocations(ctx, req.GetLocationIds(), req.GetType())
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed GetEnabledLocationConfigsByLocations: %+v", err))
		}

		// no location configuration found
		if len(conversationTypeWithLocationMap) == 0 {
			return &tpb.ListConversationsInSchoolResponse{
				Items: []*tpb.Conversation{},
				NextPage: &cpb.Paging{
					Limit: req.Paging.Limit,
					Offset: &cpb.Paging_OffsetCombined{
						OffsetCombined: &cpb.Paging_Combined{
							OffsetString:  "",
							OffsetInteger: int64(0),
						},
					},
				},
			}, nil
		}

		filter, err = rcv.requestFilterToRepoFilterV2(ctx, req, conversationTypeWithLocationMap)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed requestFilterToRepoFilterV2: %+v", err))
		}
		useV2Search = true
	} else {
		excludeConversationTypes, accessPaths, err := rcv.LocationConfigResolver.GetEnabledLocationConfigsByOrg(ctx, req.GetLocationIds())
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed GetEnabledLocationConfigsByOrg: %+v", err))
		}
		filter, err = rcv.requestFilterToRepoFilter(ctx, req, accessPaths, excludeConversationTypes)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed requestFilterToRepoFilter: %+v", err))
		}
	}

	return rcv.searchWithFilterAndFetchFromDB(ctx, filter, req, useV2Search)
}

func (rcv *ChatReader) searchWithFilterAndFetchFromDB(ctx context.Context, filter sentities.ConversationFilter, originalReq *tpb.ListConversationsInSchoolRequest, useV2Search bool) (*tpb.ListConversationsInSchoolResponse, error) {
	logger := ctxzap.Extract(ctx)
	var (
		docs []sentities.SearchConversationDoc
		err  error
	)
	if useV2Search {
		docs, err = rcv.ConversationSearchRepo.SearchV2(ctx, rcv.SearchClient, filter)
	} else {
		docs, err = rcv.ConversationSearchRepo.Search(ctx, rcv.SearchClient, filter)
	}
	if err != nil {
		logger.Error("rcv.ConversationSearchRepo.Search", zap.Error(err))
		return nil, status.Error(codes.Internal, "search service failed")
	}

	lastConversationID := ""
	lastMessageTime := int64(0)
	if len(docs) > 0 {
		lastConversationID = docs[len(docs)-1].ConversationID
		lastMessageTime = docs[len(docs)-1].LastMessage.UpdatedAt.UnixMilli()
	}
	sortedConversationIDs := make([]string, 0, len(docs))
	for _, item := range docs {
		sortedConversationIDs = append(sortedConversationIDs, item.ConversationID)
	}
	var sortedConversations []*tpb.Conversation

	if len(sortedConversationIDs) == 0 {
		sortedConversations = []*tpb.Conversation{}
	} else {
		conversations, err := rcv.getSupportChatLastestMessage(ctx, sortedConversationIDs, uint(originalReq.Paging.Limit), time.Now())
		if err != nil {
			return nil, fmt.Errorf("s.getLastestMessage: %w", err)
		}
		sortedConversations = sortConversationsByIDs(conversations, sortedConversationIDs)
	}

	return &tpb.ListConversationsInSchoolResponse{
		Items: sortedConversations,
		NextPage: &cpb.Paging{
			Limit: originalReq.Paging.Limit,
			Offset: &cpb.Paging_OffsetCombined{
				OffsetCombined: &cpb.Paging_Combined{
					OffsetString:  lastConversationID,
					OffsetInteger: lastMessageTime,
				},
			},
		},
	}, nil
}

func (rcv *ChatReader) ListConversationsInSchool(ctx context.Context, req *tpb.ListConversationsInSchoolRequest) (*tpb.ListConversationsInSchoolResponse, error) {
	filter, err := rcv.requestFilterToRepoFilter(ctx, req, []string{}, []tpb.ConversationType{})
	if err != nil {
		return nil, err
	}
	useV2Search := false
	return rcv.searchWithFilterAndFetchFromDB(ctx, filter, req, useV2Search)
}

func (rcv *ChatReader) getSupportChatLastestMessage(ctx context.Context, cIDs []string, limit uint, endAt time.Time) ([]*tpb.Conversation, error) {
	userID := interceptors.UserIDFromContext(ctx)
	logger := ctxzap.Extract(ctx)

	var pgEndAt pgtype.Timestamptz
	_ = pgEndAt.Set(endAt)

	var (
		messages []*core.Message
		err      error
	)

	messages, err = rcv.MessageRepo.GetLastMessageByConversationIDs(ctx, rcv.DB, database.TextArray(cIDs), limit, pgEndAt, false)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Error(codes.Unknown, err.Error())
	}

	conversationMap, _ := rcv.ConversationRepo.FindByIDsReturnMapByID(ctx, rcv.DB, database.TextArray(cIDs))

	// get student id per conversation
	mapConversationStudents, err := rcv.ConversationStudentRepo.FindByConversationIDs(ctx, rcv.DB, database.TextArray(cIDs))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("ConversationStudentRepo.FindByConversationIDs: %w", err).Error())
	}
	mapConversationMembers, err := rcv.ConversationMemberRepo.FindByConversationIDs(ctx, rcv.DB, database.TextArray(cIDs))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("ConversationMemberRepo.FindByConversationIDs: %w", err).Error())
	}

	var resp []*tpb.Conversation
	messageMap := convertToMapMessages(messages)

	for _, c := range conversationMap {
		conversation := conversationMap[c.Conversation.ID]
		conversationMembers := mapConversationMembers[c.Conversation.ID]
		var users = make([]*tpb.Conversation_User, 0, len(conversationMembers))
		for _, u := range conversationMembers {
			users = append(users, &tpb.Conversation_User{
				Id:        u.UserID.String,
				Group:     cpb.UserGroup(cpb.UserGroup_value[u.Role.String]),
				IsPresent: u.Status.String == core.ConversationStatusActive,
				SeenAt:    timestamppb.New(u.SeenAt.Time),
			})
		}
		studentID := ""
		if convStudent, ok := mapConversationStudents[c.Conversation.ID]; ok {
			studentID = convStudent.StudentID.String
		}
		m, ok := messageMap[c.Conversation.ID]

		if ok {
			lastmessage := toMessagePb(m)
			var seen bool
			for _, cStatus := range conversationMembers {
				if cStatus.UserID.String == userID {
					seen = cStatus.SeenAt.Time.After(m.CreatedAt.Time)
				}
			}
			resp = append(resp, &tpb.Conversation{
				ConversationId:   c.Conversation.ID.String,
				StudentId:        studentID,
				LastMessage:      lastmessage,
				Status:           tpb.ConversationStatus(tpb.CodesMessageType_value[conversation.Conversation.Status.String]),
				ConversationType: tpb.ConversationType(tpb.ConversationType_value[conversation.Conversation.ConversationType.String]),
				Seen:             seen,
				Users:            users,
				ConversationName: conversation.Conversation.Name.String,
				IsReplied:        c.IsReply.Bool, // client handle
				Owner:            conversation.Conversation.Owner.String,
			})
		} else {
			resp = append(resp, &tpb.Conversation{
				ConversationId:   c.Conversation.ID.String,
				StudentId:        studentID,
				Status:           tpb.ConversationStatus(tpb.CodesMessageType_value[conversation.Conversation.Status.String]),
				ConversationType: tpb.ConversationType(tpb.ConversationType_value[conversation.Conversation.ConversationType.String]),
				Users:            users,
				ConversationName: conversation.Conversation.Name.String,
				IsReplied:        c.IsReply.Bool, // client handle
				Owner:            conversation.Conversation.Owner.String,
			})
		}
	}
	return resp, nil
}
