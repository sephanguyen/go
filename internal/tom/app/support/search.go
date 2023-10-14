package support

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/tom/domain/support"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
	"go.opentelemetry.io/otel/trace"

	"github.com/abadojack/whatlanggo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatReaderService interface {
	ListConversationByUsers(ctx context.Context, req *tpb.ListConversationByUsersRequest) (*tpb.ListConversationByUsersResponse, error)
}

// TODO: consider replacing this with CDC
type SearchIndexer struct {
	Logger        *zap.Logger
	SearchFactory elastic.SearchFactory
	ChatReader    ChatReaderService
	EurekaCourseReaderService
	DB database.Ext

	SearchRepo interface {
		// upsert multiple document at a time (using bulk index of Elasticsearch)
		BulkUpsert(ctx context.Context, s elastic.SearchFactory, docs []support.SearchConversationDoc) (int, error)
	}
	ConversationLocationRepo interface {
		FindByConversationIDs(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray) (map[string][]core.ConversationLocation, error)
	}
	MessageRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (c *core.Message, err error)
	}
	ConversationRepo interface {
		FindByIDsReturnMapByID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (map[pgtype.Text]core.ConversationFull, error)
	}
	ConversationMemberRepo interface {
		FindByConversationID(ctx context.Context, db database.QueryExecer, conversationID pgtype.Text) (map[pgtype.Text]core.ConversationMembers, error)
	}
	ConversationStudentRepo interface {
		UpdateSearchIndexTime(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray, time pgtype.Timestamptz) error
		FindSearchIndexTime(ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray) (map[string]pgtype.Timestamptz, error)
	}
}

type EurekaCourseReaderService interface {
	ListCourseIDsByStudents(ctx context.Context, in *epb.ListCourseIDsByStudentsRequest, opts ...grpc.CallOption) (*epb.ListCourseIDsByStudentsResponse, error)
}

func (s *SearchIndexer) HandleStudentCourseUpdated(ctx context.Context, req *npb.EventCourseStudent) error {
	if len(req.GetStudentIds()) == 0 {
		return status.Error(codes.InvalidArgument, "empty student ids")
	}

	newCheckTime := time.Now()
	conversations, err := s.ChatReader.ListConversationByUsers(ctx, &tpb.ListConversationByUsersRequest{
		UserIds: req.GetStudentIds(),
	})
	if err != nil {
		return fmt.Errorf("s.ChatReader.ListConversationByUsers unable to get conversations: %w", err)
	}
	// Student conversation is not yet created, let this msg be ignored, the creation of conversation
	// will revisit this logic anyway
	if len(conversations.GetItems()) == 0 {
		return nil
	}
	cIDs := make([]string, 0, len(conversations.GetItems()))
	for _, item := range conversations.GetItems() {
		cIDs = append(cIDs, item.ConversationId)
	}

	_, err = s.rebuildDocumentsFromConversationEntities(ctx, cIDs, conversations)
	if err != nil {
		return err
	}

	err = s.ConversationStudentRepo.UpdateSearchIndexTime(ctx, s.DB, database.TextArray(cIDs), database.Timestamptz(newCheckTime))
	if err != nil {
		s.Logger.Warn("ConversationStudentRepo.UpdateSearchIndexTime", zap.Error(err))
	}
	return nil
}

// because not really one user have one conversation: so we have total and total success to know
// from one or more users -> find conversation members -> retrieve studentcourse -> avoid missing course in conversation
func (s *SearchIndexer) BuildConversationDocument(ctx context.Context, req *tpb.BuildConversationDocumentRequest) (*tpb.BuildConversationDocumentResponse, error) {
	if req.GetConversationIds() == nil && req.GetUserIds() == nil {
		return nil, status.Error(codes.InvalidArgument, "empty params")
	}

	total, err := s.rebuildConversationDocument(ctx, req.GetConversationIds())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &tpb.BuildConversationDocumentResponse{
		Total:        int32(len(req.GetConversationIds())),
		TotalSuccess: int32(total),
	}, nil
}

func retrieveStudentIDsFromConversations(conversations []*tpb.Conversation) []string {
	studentIDs := make([]string, 0, len(conversations))
	for _, c := range conversations {
		studentIDs = append(studentIDs, c.GetStudentId())
	}
	return golibs.GetUniqueElementStringArray(studentIDs)
}

func getAccessPaths(sl []core.ConversationLocation) []string {
	ret := make([]string, 0, len(sl))
	for _, item := range sl {
		ret = append(ret, item.AccessPath.String)
	}
	return ret
}
func convertToEsConversations(rawConversations []*tpb.Conversation, courseIDs []*epb.ListCourseIDsByStudentsResponse_StudentCourses, convLoc map[string][]core.ConversationLocation) []support.SearchConversationDoc {
	res := make([]support.SearchConversationDoc, 0, len(rawConversations))
	// first convert course ids to each student
	// define: if one student exist in conservation --> add all course related the student

	// convert mapStudentCourse
	mapStudentCourse := convertStudentCourseToMap(courseIDs)
	for _, c := range rawConversations {
		updatedAt := c.GetLastMessage().GetUpdatedAt().AsTime()
		userIDs := convertTpbUsersToUserIds(c.GetUsers())
		locs := convLoc[c.GetConversationId()]

		esConversation := support.SearchConversationDoc{
			ConversationID: c.GetConversationId(),
			//TODO: we will handle which belong to country later
			CourseIDs: retrieveCourseIDs(c, mapStudentCourse), //TODO: add all course related to student to ..,
			UserIDs:   userIDs,                                // convert to array,
			LastMessage: support.SearchLastMessage{
				UpdatedAt: updatedAt,
			},
			IsReplied:        c.GetIsReplied(),
			Owner:            c.GetOwner(),
			ConversationType: c.GetConversationType().String(),
			AccessPath:       getAccessPaths(locs),
		}
		switch detectLanguage(c.ConversationName) {
		case constants.English:
			esConversation.ConversationNameEnglish = c.GetConversationName()
		case constants.Japanese, constants.Mandarin:
			esConversation.ConversationNameJapanese = c.GetConversationName()
		default:
			esConversation.ConversationNameEnglish = c.GetConversationName()
		}
		res = append(res, esConversation)
	}
	return res
}
func detectLanguage(s string) string {
	info := whatlanggo.DetectLang(s)
	return info.String()
}

func convertStudentCourseToMap(scs []*epb.ListCourseIDsByStudentsResponse_StudentCourses) map[string][]string {
	mapStudentCourse := make(map[string][]string, len(scs))
	for _, sc := range scs {
		mapStudentCourse[sc.StudentId] = sc.CourseIds
	}
	return mapStudentCourse
}

// TODO: improve performance
// from userids + mapStudent -> get all courses with no duplicate
func retrieveCourseIDs(c *tpb.Conversation, mapStudentCourse map[string][]string) []string {
	return golibs.GetUniqueElementStringArray(mapStudentCourse[c.StudentId])
}

// func sumUpCourseIdsFromUsers(userIds []string,epb.ListCourseIDsByStudentsResponse_StudentCourses)
func convertTpbUsersToUserIds(usrs []*tpb.Conversation_User) []string {
	usrIDs := make([]string, 0, len(usrs))
	for _, u := range usrs {
		if u.IsPresent {
			usrIDs = append(usrIDs, u.GetId())
		}
	}
	return usrIDs
}

func (s *SearchIndexer) reindexConversationDocument(ctx context.Context, checkTime time.Time, cIDs []string) error {
	mapIDTime, err := s.ConversationStudentRepo.FindSearchIndexTime(ctx, s.DB, database.TextArray(cIDs))
	if err != nil {
		return fmt.Errorf("ConversationStudentRepo.FindSearchIndexTime: %w", err)
	}
	needIndex := []string{}
	for cid, t := range mapIDTime {
		// null update time or update time < checktime
		if t.Status == pgtype.Null || t.Time.Before(checkTime) {
			needIndex = append(needIndex, cid)
		}
	}
	if len(needIndex) == 0 {
		return nil
	}
	newCheckTime := time.Now()
	_, err = s.rebuildConversationDocument(ctx, needIndex)
	if err != nil {
		return err
	}
	err = s.ConversationStudentRepo.UpdateSearchIndexTime(ctx, s.DB, database.TextArray(needIndex), database.Timestamptz(newCheckTime))
	if err != nil {
		s.Logger.Warn("ConversationStudentRepo.UpdateSearchIndexTime", zap.Error(err))
	}
	return nil
}

func (s *SearchIndexer) rebuildDocumentsFromConversationEntities(ctx context.Context, cIDs []string, conversations *tpb.ListConversationByUsersResponse) (int, error) {
	resourcePath, _ := interceptors.ResourcePathFromContext(ctx)
	studentIDs := retrieveStudentIDsFromConversations(conversations.GetItems())
	if len(studentIDs) == 0 {
		return 0, fmt.Errorf("retrieveStudentIDsFromConversations: unable to get students for conversatios: %v", cIDs)
	}
	studentCourses, err := s.EurekaCourseReaderService.ListCourseIDsByStudents(ctx, &epb.ListCourseIDsByStudentsRequest{
		StudentIds:     studentIDs,
		OrganizationId: resourcePath,
	})
	if resourcePath == "" {
		sp := trace.SpanFromContext(ctx)
		fmt.Printf("ident %s\n", sp.SpanContext().TraceID())
	}
	if err != nil {
		return 0, fmt.Errorf("unable to list course ids: %w", err)
	}
	convLocationMap := make(map[string][]core.ConversationLocation)
	locationMap, err := s.ConversationLocationRepo.FindByConversationIDs(ctx, s.DB, database.TextArray(cIDs))
	if err != nil {
		return 0, fmt.Errorf("s.ConversationLocationRepo.FindByConversationIDs %w", err)
	}
	convLocationMap = locationMap

	esConversations := convertToEsConversations(conversations.GetItems(), studentCourses.GetStudentCourses(), convLocationMap)

	total, err := s.SearchRepo.BulkUpsert(ctx, s.SearchFactory, esConversations)
	if err != nil {
		return 0, fmt.Errorf("s.SearchRepo.BulkUpsert unable to bulk insert conversations to elasticsearch: %w", err)
	}
	return total, nil
}

func (s *SearchIndexer) rebuildConversationDocument(ctx context.Context, cIDs []string) (int, error) {
	conversations, err := s.ChatReader.ListConversationByUsers(ctx, &tpb.ListConversationByUsersRequest{
		ConversationIds: cIDs,
	})
	if err != nil {
		return 0, fmt.Errorf("s.ChatReader.ListConversationByUsers unable to get conversations: %w", err)
	}
	return s.rebuildDocumentsFromConversationEntities(ctx, cIDs, conversations)
}

func (s *SearchIndexer) ReindexConversationDocument(ctx context.Context, timeCheck time.Time, conversationIDs []string) error {
	return s.reindexConversationDocument(ctx, timeCheck, conversationIDs)
}
