package support

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/tom/domain/support"
	mock_repositories "github.com/manabie-com/backend/mock/tom/repositories"
	mock_services "github.com/manabie-com/backend/mock/tom/services"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestSearchIndexer_reindexConversationDocument(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	mockChatReader := &mock_services.ChatReaderService{}
	mockEurekaCourseReaderService := &mock_services.EurekaCourseReaderService{}
	mockSearchRepo := &mock_repositories.MockSearchRepo{}
	mockConversationRepo := &mock_repositories.MockConversationRepo{}
	mockConStuRepo := &mock_repositories.MockConversationStudentRepo{}
	mockMessageRepo := &mock_repositories.MockMessageRepo{}
	mockLocationRepo := &mock_repositories.MockConversationLocationRepo{}

	s := &SearchIndexer{
		ChatReader:                mockChatReader,
		EurekaCourseReaderService: mockEurekaCourseReaderService,
		SearchRepo:                mockSearchRepo,
		ConversationRepo:          mockConversationRepo,
		MessageRepo:               mockMessageRepo,
		ConversationStudentRepo:   mockConStuRepo,
		ConversationLocationRepo:  mockLocationRepo,
	}

	lastMsg := &tpb.MessageResponse{
		MessageId: "message-1",
		UpdatedAt: timestamppb.Now(),
	}

	convID1 := "conversation-1"
	convID2 := "conversation-2"
	fullDoc := []support.SearchConversationDoc{
		{
			ConversationID:          "conversation-1",
			ConversationNameEnglish: "english",
			CourseIDs:               []string{"course-1", "course-2", "course-3"},
			LastMessage: support.SearchLastMessage{
				UpdatedAt: lastMsg.UpdatedAt.AsTime(),
			},
			ConversationType: tpb.ConversationType_CONVERSATION_PARENT.String(),
			UserIDs:          []string{},
			AccessPath:       []string{"loc-1", "loc-2"},
		},
		{
			ConversationID:           "conversation-2",
			ConversationNameJapanese: "ドラえもん",
			CourseIDs:                []string{"course-1", "course-5", "course-6"},
			LastMessage: support.SearchLastMessage{
				UpdatedAt: lastMsg.UpdatedAt.AsTime(),
			},
			ConversationType: tpb.ConversationType_CONVERSATION_STUDENT.String(),
			UserIDs:          []string{},
			AccessPath:       []string{"loc-2"},
		},
	}
	loc1 := "loc-1"
	loc2 := "loc-2"
	conversationIDs := []string{"conversation-1", "conversation-2"}
	locationMap := map[string][]core.ConversationLocation{
		"conversation-1": {
			{
				LocationID:     dbText(loc1),
				ConversationID: dbText("conversation-1"),
				AccessPath:     dbText("loc-1"),
			},
			{
				LocationID:     dbText(loc2),
				ConversationID: dbText("conversation-1"),
				AccessPath:     dbText("loc-2"),
			},
		},
		"conversation-2": {
			{
				LocationID:     dbText(loc2),
				ConversationID: dbText("conversation-2"),
				AccessPath:     dbText("loc-2"),
			},
		},
	}
	studentIDs := []string{"student-1", "student-2"}
	conversationResp := &tpb.ListConversationByUsersResponse{

		Items: []*tpb.Conversation{
			{
				ConversationId:   "conversation-1",
				ConversationName: "english",
				StudentId:        "student-1",
				LastMessage:      lastMsg,
				ConversationType: tpb.ConversationType_CONVERSATION_PARENT,
			},
			{
				ConversationId:   "conversation-2",
				StudentId:        "student-2",
				ConversationName: "ドラえもん",
				LastMessage:      lastMsg,
				ConversationType: tpb.ConversationType_CONVERSATION_STUDENT,
			},
		},
	}
	courseStudentResp := &epb.ListCourseIDsByStudentsResponse{
		StudentCourses: []*epb.ListCourseIDsByStudentsResponse_StudentCourses{
			{
				StudentId: "student-1",
				CourseIds: []string{"course-1", "course-2", "course-3"},
			},
			{
				StudentId: "student-2",
				CourseIds: []string{"course-1", "course-5", "course-6"},
			},
		},
	}

	testCases := []TestCase{
		{
			name:        "success",
			ctx:         ctx,
			req:         conversationIDs,
			expectedErr: nil,
			setup: func(ctx context.Context) {

				mockConStuRepo.On("FindSearchIndexTime", mock.Anything, mock.Anything, database.TextArray(conversationIDs)).Once().Return(map[string]pgtype.Timestamptz{
					convID1: {Status: pgtype.Null},
					convID2: {Status: pgtype.Null},
				}, nil)

				mockChatReader.On("ListConversationByUsers", mock.Anything, mock.MatchedBy(func(req *tpb.ListConversationByUsersRequest) bool {
					return stringutil.SliceElementsMatch(conversationIDs, req.GetConversationIds())
				})).Once().Return(conversationResp, nil)
				mockEurekaCourseReaderService.On("ListCourseIDsByStudents", mock.Anything, &epb.ListCourseIDsByStudentsRequest{StudentIds: studentIDs}).Once().Return(courseStudentResp, nil)
				mockLocationRepo.On("FindByConversationIDs", ctx, s.DB, database.TextArray(conversationIDs)).Once().Return(locationMap, nil)

				mockSearchRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.MatchedBy(func(docs []support.SearchConversationDoc) bool {
					return cmp.Equal(docs, fullDoc)
				})).Once().Return(2, nil)

				mockConStuRepo.On("UpdateSearchIndexTime", mock.Anything, mock.Anything, database.TextArray(conversationIDs), mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.reindexConversationDocument(ctx, time.Now(), testCase.req.([]string))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestSearchIndexer_HandleCourseStudentUpdated(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	mockChatReader := &mock_services.ChatReaderService{}
	mockEurekaCourseReaderService := &mock_services.EurekaCourseReaderService{}
	mockSearchRepo := &mock_repositories.MockSearchRepo{}
	mockConvLocationRepo := &mock_repositories.MockConversationLocationRepo{}
	mockConvStuRepo := &mock_repositories.MockConversationStudentRepo{}

	s := &SearchIndexer{
		ChatReader:                mockChatReader,
		EurekaCourseReaderService: mockEurekaCourseReaderService,
		SearchRepo:                mockSearchRepo,
		ConversationLocationRepo:  mockConvLocationRepo,
		ConversationStudentRepo:   mockConvStuRepo,
	}
	studentIDs := []string{"student-1", "student-2"}
	lastMsg := &tpb.MessageResponse{
		MessageId: "message-1",
		UpdatedAt: timestamppb.Now(),
	}

	cIDs := []string{"conversation-1", "conversation-2"}
	conversationResp := &tpb.ListConversationByUsersResponse{
		Items: []*tpb.Conversation{
			{
				ConversationId:   "conversation-1",
				ConversationName: "english",
				StudentId:        "student-1",
				LastMessage:      lastMsg,
				ConversationType: tpb.ConversationType_CONVERSATION_PARENT,
			},
			{
				ConversationId:   "conversation-2",
				StudentId:        "student-2",
				ConversationName: "ドラえもん",
				LastMessage:      lastMsg,
				ConversationType: tpb.ConversationType_CONVERSATION_STUDENT,
			},
		},
	}
	courseStudentResp := &epb.ListCourseIDsByStudentsResponse{
		StudentCourses: []*epb.ListCourseIDsByStudentsResponse_StudentCourses{
			{
				StudentId: "student-1",
				CourseIds: []string{"course-1", "course-2", "course-3"},
			},
			{
				StudentId: "student-2",
				CourseIds: []string{"course-1", "course-5", "course-6"},
			},
		},
	}
	fullDoc := []support.SearchConversationDoc{
		{
			ConversationID:          "conversation-1",
			ConversationNameEnglish: "english",
			CourseIDs:               []string{"course-1", "course-2", "course-3"},
			LastMessage: support.SearchLastMessage{
				UpdatedAt: lastMsg.UpdatedAt.AsTime(),
			},
			ConversationType: tpb.ConversationType_CONVERSATION_PARENT.String(),
			UserIDs:          []string{},
			AccessPath:       []string{"loc-1", "loc-2"},
		},
		{
			ConversationID:           "conversation-2",
			ConversationNameJapanese: "ドラえもん",
			CourseIDs:                []string{"course-1", "course-5", "course-6"},
			LastMessage: support.SearchLastMessage{
				UpdatedAt: lastMsg.UpdatedAt.AsTime(),
			},
			ConversationType: tpb.ConversationType_CONVERSATION_STUDENT.String(),
			UserIDs:          []string{},
			AccessPath:       []string{"loc-2"},
		},
	}
	loc1, loc2 := "loc1", "loc2"
	locationMap := map[string][]core.ConversationLocation{
		"conversation-1": {
			{
				LocationID:     dbText(loc1),
				ConversationID: dbText("conversation-1"),
				AccessPath:     dbText("loc-1"),
			},
			{
				LocationID:     dbText(loc2),
				ConversationID: dbText("conversation-1"),
				AccessPath:     dbText("loc-2"),
			},
		},
		"conversation-2": {
			{
				LocationID:     dbText(loc2),
				ConversationID: dbText("conversation-2"),
				AccessPath:     dbText("loc-2"),
			},
		},
	}

	testCases := []TestCase{
		{
			name: "success",
			ctx:  ctx,
			req: &npb.EventCourseStudent{
				StudentIds: studentIDs,
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				var dbReadTime time.Time
				mockChatReader.On("ListConversationByUsers", mock.Anything, &tpb.ListConversationByUsersRequest{UserIds: studentIDs}).Once().
					Run(func(mock.Arguments) {
						dbReadTime = time.Now()
					}).
					Return(conversationResp, nil)
				mockEurekaCourseReaderService.On("ListCourseIDsByStudents", mock.Anything, &epb.ListCourseIDsByStudentsRequest{StudentIds: studentIDs}).Once().Return(courseStudentResp, nil)
				mockConvLocationRepo.On("FindByConversationIDs", mock.Anything, mock.Anything, database.TextArray(cIDs)).Return(locationMap, nil)
				mockSearchRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.MatchedBy(func(docs []support.SearchConversationDoc) bool {
					return cmp.Equal(docs, fullDoc)
				})).Once().Return(2, nil)
				// this timestamp must be set before reading db
				mockConvStuRepo.On("UpdateSearchIndexTime", mock.Anything, mock.Anything, database.TextArray(cIDs), mock.MatchedBy(func(t pgtype.Timestamptz) bool {
					return t.Time.Before(dbReadTime)
				})).Once().Return(nil)
			},
		},
		{
			name: "empty params",
			ctx:  ctx,
			req: &npb.EventCourseStudent{
				StudentIds: nil,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "empty student ids"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "error when call eureka-ListCourseIDsByStudents",
			ctx:  ctx,
			req: &npb.EventCourseStudent{
				StudentIds: studentIDs,
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("unable to list course ids: Err Eureka"),
			setup: func(ctx context.Context) {
				mockChatReader.On("ListConversationByUsers", mock.Anything, &tpb.ListConversationByUsersRequest{UserIds: studentIDs}).Once().Return(conversationResp, nil)
				mockEurekaCourseReaderService.On("ListCourseIDsByStudents", mock.Anything, &epb.ListCourseIDsByStudentsRequest{StudentIds: studentIDs}).Once().Return(courseStudentResp, fmt.Errorf("Err Eureka"))
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.HandleStudentCourseUpdated(ctx, testCase.req.(*npb.EventCourseStudent))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			}
		})
	}
}

func TestSearchIndexer_BuildConversationDocument(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	mockChatReader := &mock_services.ChatReaderService{}
	mockEurekaCourseReaderService := &mock_services.EurekaCourseReaderService{}
	mockSearchRepo := &mock_repositories.MockSearchRepo{}
	mockLocationRepo := &mock_repositories.MockConversationLocationRepo{}

	s := &SearchIndexer{
		ChatReader:                mockChatReader,
		EurekaCourseReaderService: mockEurekaCourseReaderService,
		SearchRepo:                mockSearchRepo,
		ConversationLocationRepo:  mockLocationRepo,
	}
	conversationIDs := []string{"conversation-1", "conversation-2"}
	studentIDs := []string{"student-1", "student-2"}
	lastMsg := &tpb.MessageResponse{
		MessageId: "message-1",
		UpdatedAt: timestamppb.Now(),
	}
	loc1 := "loc-1"
	loc2 := "loc-2"
	locationMap := map[string][]core.ConversationLocation{
		"conversation-1": {
			{
				LocationID:     dbText(loc1),
				ConversationID: dbText("conversation-1"),
				AccessPath:     dbText("loc-1"),
			},
			{
				LocationID:     dbText(loc2),
				ConversationID: dbText("conversation-1"),
				AccessPath:     dbText("loc-2"),
			},
		},
		"conversation-2": {
			{
				LocationID:     dbText(loc2),
				ConversationID: dbText("conversation-2"),
				AccessPath:     dbText("loc-2"),
			},
		},
	}

	conversationResp := &tpb.ListConversationByUsersResponse{
		Items: []*tpb.Conversation{
			{
				ConversationId:   "conversation-1",
				ConversationName: "english",
				StudentId:        "student-1",
				LastMessage:      lastMsg,
				ConversationType: tpb.ConversationType_CONVERSATION_PARENT,
			},
			{
				ConversationId:   "conversation-2",
				StudentId:        "student-2",
				ConversationName: "ドラえもん",
				LastMessage:      lastMsg,
				ConversationType: tpb.ConversationType_CONVERSATION_STUDENT,
			},
		},
	}
	courseStudentResp := &epb.ListCourseIDsByStudentsResponse{
		StudentCourses: []*epb.ListCourseIDsByStudentsResponse_StudentCourses{
			{
				StudentId: "student-1",
				CourseIds: []string{"course-1", "course-2", "course-3"},
			},
			{
				StudentId: "student-2",
				CourseIds: []string{"course-1", "course-5", "course-6"},
			},
		},
	}
	fullDoc := []support.SearchConversationDoc{
		{
			ConversationID:          "conversation-1",
			ConversationNameEnglish: "english",
			CourseIDs:               []string{"course-1", "course-2", "course-3"},
			LastMessage: support.SearchLastMessage{
				UpdatedAt: lastMsg.UpdatedAt.AsTime(),
			},
			ConversationType: tpb.ConversationType_CONVERSATION_PARENT.String(),
			UserIDs:          []string{},
			AccessPath:       []string{"loc-1", "loc-2"},
		},
		{
			ConversationID:           "conversation-2",
			ConversationNameJapanese: "ドラえもん",
			CourseIDs:                []string{"course-1", "course-5", "course-6"},
			LastMessage: support.SearchLastMessage{
				UpdatedAt: lastMsg.UpdatedAt.AsTime(),
			},
			ConversationType: tpb.ConversationType_CONVERSATION_STUDENT.String(),
			UserIDs:          []string{},
			AccessPath:       []string{"loc-2"},
		},
	}

	testCases := []TestCase{
		{
			name: "success",
			ctx:  ctx,
			req: &tpb.BuildConversationDocumentRequest{
				ConversationIds: conversationIDs,
				UserIds:         nil,
			},
			expectedResp: &tpb.BuildConversationDocumentResponse{Total: 2, TotalSuccess: 2},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockChatReader.On("ListConversationByUsers", mock.Anything, &tpb.ListConversationByUsersRequest{ConversationIds: conversationIDs}).Once().Return(conversationResp, nil)
				mockEurekaCourseReaderService.On("ListCourseIDsByStudents", mock.Anything, &epb.ListCourseIDsByStudentsRequest{StudentIds: studentIDs}).Once().Return(courseStudentResp, nil)

				mockLocationRepo.On("FindByConversationIDs", ctx, s.DB, database.TextArray(conversationIDs)).Once().Return(locationMap, nil)
				mockSearchRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.MatchedBy(func(docs []support.SearchConversationDoc) bool {
					return cmp.Equal(docs, fullDoc)
				})).Once().Return(2, nil)
			},
		},
		{
			name: "empty params",
			ctx:  ctx,
			req: &tpb.BuildConversationDocumentRequest{
				ConversationIds: nil,
				UserIds:         nil,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "empty params"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "error when call tom-ListConversationByUsers",
			ctx:  ctx,
			req: &tpb.BuildConversationDocumentRequest{
				ConversationIds: conversationIDs,
				UserIds:         nil,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "s.ChatReader.ListConversationByUsers unable to get conversations: Err Tom"),
			setup: func(ctx context.Context) {
				mockChatReader.On("ListConversationByUsers", mock.Anything, &tpb.ListConversationByUsersRequest{ConversationIds: conversationIDs}).Once().Return(conversationResp, fmt.Errorf("Err Tom"))
			},
		},
		{
			name: "error when call eureka-ListCourseIDsByStudents",
			ctx:  ctx,
			req: &tpb.BuildConversationDocumentRequest{
				ConversationIds: conversationIDs,
				UserIds:         nil,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, "unable to list course ids: Err Eureka"),
			setup: func(ctx context.Context) {
				mockChatReader.On("ListConversationByUsers", mock.Anything, &tpb.ListConversationByUsersRequest{ConversationIds: conversationIDs}).Once().Return(conversationResp, nil)
				mockEurekaCourseReaderService.On("ListCourseIDsByStudents", mock.Anything, &epb.ListCourseIDsByStudentsRequest{StudentIds: studentIDs}).Once().Return(courseStudentResp, fmt.Errorf("Err Eureka"))
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.BuildConversationDocument(ctx, testCase.req.(*tpb.BuildConversationDocumentRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, resp)
			} else {
				expectedResp := testCase.expectedResp.(*tpb.BuildConversationDocumentResponse)
				assert.Equal(t, expectedResp.Total, resp.Total, "total not equal")
				assert.Equal(t, expectedResp.TotalSuccess, resp.TotalSuccess, "total success not equal")
			}
		})
	}
}
