package services

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	natsJS "github.com/nats-io/nats.go"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func isEqualTopicEnAndPb(e *entities.Topic, topic *pb.Topic) bool {
	if topic.UpdatedAt == nil {
		topic.UpdatedAt = &timestamppb.Timestamp{Seconds: e.UpdatedAt.Time.Unix()}
	}
	if topic.CreatedAt == nil {
		topic.CreatedAt = &timestamppb.Timestamp{Seconds: e.CreatedAt.Time.Unix()}
	}
	updatedAt := &timestamppb.Timestamp{Seconds: e.UpdatedAt.Time.Unix()}
	createdAt := &timestamppb.Timestamp{Seconds: e.CreatedAt.Time.Unix()}

	return (e.ID.String == topic.Id) &&
		(e.Name.String == topic.Name) &&
		(pb.Country(pb.Country_value[e.Country.String]) == topic.Country) &&
		(pb.Subject(pb.Subject_value[e.Subject.String]) == topic.Subject) &&
		(pb.TopicType(pb.TopicType_value[e.TopicType.String]) == topic.Type) &&
		updatedAt.AsTime().Equal(topic.UpdatedAt.AsTime()) &&
		createdAt.AsTime().Equal(topic.CreatedAt.AsTime())
}

func Test_toTopicEntity(t *testing.T) {
	t.Parallel()
	req1 := &pb.Topic{
		Id:        fmt.Sprintf("%d", 1),
		Name:      "Random name",
		Subject:   pb.Subject_SUBJECT_MATHS,
		Type:      pb.TopicType_TOPIC_TYPE_LEARNING,
		CreatedAt: nil,
		UpdatedAt: nil,
	}
	req2 := &pb.Topic{
		Id:        fmt.Sprintf("%d", 1),
		Name:      "Random name",
		Subject:   pb.Subject_SUBJECT_MATHS,
		Type:      pb.TopicType_TOPIC_TYPE_LEARNING,
		CreatedAt: &timestamppb.Timestamp{Seconds: time.Now().Unix()},
		UpdatedAt: &timestamppb.Timestamp{Seconds: time.Now().Unix()},
	}
	e1 := toTopicEntity(req1)
	e2 := toTopicEntity(req2)
	require.True(t, isEqualTopicEnAndPb(e1, req1))
	require.True(t, isEqualTopicEnAndPb(e2, req2))
}

func generateTopic() *pb.Topic {
	rand.Seed(time.Now().UnixNano())
	num := rand.Int()
	now := time.Now()
	return &pb.Topic{
		Id:           fmt.Sprintf("%d", num),
		Name:         "Random name",
		Subject:      pb.Subject_SUBJECT_MATHS,
		Type:         pb.TopicType_TOPIC_TYPE_LEARNING,
		CreatedAt:    &timestamppb.Timestamp{Seconds: now.Unix()},
		UpdatedAt:    &timestamppb.Timestamp{Seconds: now.Unix()},
		Status:       pb.TopicStatus_TOPIC_STATUS_DRAFT,
		ChapterId:    "mock-chapter-id",
		DisplayOrder: 1,
	}
}

func TestTopicModifierService_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	topicRepo := new(mock_repositories.MockTopicRepo)
	chapterRepo := new(mock_repositories.MockChapterRepo)
	t1 := generateTopic()
	t1.Id = ""
	t2 := generateTopic()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	topics := []*entities.Topic{
		{
			ID: database.Text("mock-topic-id-1"),
		},
		{
			ID: database.Text("mock-topic-id-2"),
		},
	}
	chapter := &entities.Chapter{
		ID:   database.Text("mock-chapter-id"),
		Name: database.Text("mock-chapter-name"),
	}
	mapChapter := make(map[string]*entities.Chapter)
	mapChapter["mock-chapter-id"] = &entities.Chapter{
		ID:   database.Text("mock-chapter-id"),
		Name: database.Text("mock-chapter-name"),
	}
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "admin"),
			req: &pb.UpsertTopicsRequest{
				Topics: []*pb.Topic{
					t1, t2,
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Commit", mock.Anything).Once().Return(nil)

				chapterRepo.On("FindByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(mapChapter, nil)
				chapterRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(chapter, nil)
				topicRepo.On("RetrieveByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("BulkImport", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				topicRepo.On("RetrieveByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(topics, nil)
				chapterRepo.On("UpdateCurrentTopicDisplayOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				topicRepo.On("UpdateTotalLOs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(2).Return(nil)
			},
		},
		{
			name: "error insert",
			ctx:  interceptors.ContextWithUserID(ctx, "error insert"),
			req: &pb.UpsertTopicsRequest{
				Topics: []*pb.Topic{
					t1, t2,
				},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("TopicRepo.BulkImport: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", mock.Anything, mock.Anything).Once().Return(tx, nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				chapterRepo.On("FindByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(mapChapter, nil)
				chapterRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(chapter, nil)
				topicRepo.On("RetrieveByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("RetrieveByIDs", mock.Anything, mock.Anything, mock.Anything).Once().Return(topics, nil)
				topicRepo.On("BulkImport", ctx, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
				topicRepo.On("UpdateTotalLOs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
				chapterRepo.On("UpdateCurrentTopicDisplayOrder", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}

	s := &TopicModifierService{
		DB:          db,
		TopicRepo:   topicRepo,
		ChapterRepo: chapterRepo,
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.UpsertTopicsRequest)
			_, err := s.Upsert(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestTopicModifierService_Publish(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	topicRepo := new(mock_repositories.MockTopicRepo)

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "happy case"),
			req: &pb.PublishTopicsRequest{
				TopicIds: []string{"1", "2"},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				topicRepo.On("UpdateStatus", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "err query case",
			ctx:  interceptors.ContextWithUserID(ctx, "err query case"),
			req: &pb.PublishTopicsRequest{
				TopicIds: []string{"1", "2"},
			},
			expectedErr: status.Error(codes.Internal, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				topicRepo.On("UpdateStatus", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
	}

	s := &TopicModifierService{
		TopicRepo: topicRepo,
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.PublishTopicsRequest)
			_, err := s.Publish(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), errors.Cause(err).Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestTopicModifierService_DeleteTopics(t *testing.T) {
	t.Parallel()

	topicRepo := new(mock_repositories.MockTopicRepo)

	db := &mock_database.Ext{}
	s := &TopicModifierService{
		DB:        db,
		TopicRepo: topicRepo,
	}

	validReq := &pb.DeleteTopicsRequest{
		TopicIds: []string{"topic-1"},
	}

	m := map[string]*entities.Topic{
		"topic-1": {
			ID: database.Text("topic-1"),
		},
	}

	testCases := map[string]TestCase{
		"happy case": {
			req: validReq,
			setup: func(ctx context.Context) {
				topicRepo.On("FindByIDsV2", ctx, db, validReq.TopicIds, false).Once().Return(m, nil)
				topicRepo.On("SoftDelete", ctx, db, validReq.TopicIds).Once().Return(0, nil)
			},
		},
		"chapters not exist": {
			req: validReq,
			setup: func(ctx context.Context) {
				topicRepo.On("FindByIDsV2", ctx, db, validReq.TopicIds, false).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedErr: status.Errorf(codes.InvalidArgument, fmt.Errorf("unable to find topics by ids: %w", pgx.ErrNoRows).Error()),
		},
		"err ChapterRepo.SoftDelete": {
			req: validReq,
			setup: func(ctx context.Context) {
				topicRepo.On("FindByIDsV2", ctx, db, validReq.TopicIds, false).Once().Return(m, nil)
				topicRepo.On("SoftDelete", ctx, db, validReq.TopicIds).Once().Return(0, ErrSomethingWentWrong)
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("unable to delete topics: %w", ErrSomethingWentWrong).Error()),
		},
	}
	for caseName, testCase := range testCases {
		t.Run(caseName, func(t *testing.T) {
			ctx := context.Background()
			testCase.setup(ctx)
			req := testCase.req.(*pb.DeleteTopicsRequest)
			if _, err := s.DeleteTopics(ctx, req); testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestCTopicModifierService_AssignTopicItems(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db := &mock_database.Ext{}
	jsm := new(mock_nats.JetStreamManagement)

	learningObjectiveRepo := &mock_repositories.MockLearningObjectiveRepo{}
	topicLearningObjectiveRepo := &mock_repositories.MockTopicsLearningObjectivesRepo{}
	svc := TopicModifierService{
		JSM:                          jsm,
		DB:                           db,
		LearningObjectiveRepo:        learningObjectiveRepo,
		TopicsLearningObjectivesRepo: topicLearningObjectiveRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &pb.AssignTopicItemsRequest{
				TopicId: "topic_id",
				Items: []*pb.AssignTopicItemsRequest_Item{
					{
						ItemId: &pb.AssignTopicItemsRequest_Item_LoId{
							LoId: "lo_id",
						},
						DisplayOrder: 1,
					},
				},
			},
			expectedErr: nil,
			setup: func(c context.Context) {
				topicLearningObjectiveRepo.On("BulkImport", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				learningObjectiveRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.LearningObjective{
					{
						ID:            pgtype.Text{String: "topic_id"},
						Name:          pgtype.Text{String: "hello"},
						Country:       pgtype.Text{String: "country"},
						Grade:         pgtype.Int2{Int: 5},
						Subject:       pgtype.Text{String: "subject"},
						TopicID:       pgtype.Text{String: "topic_id"},
						MasterLoID:    pgtype.Text{String: "master loid"},
						DisplayOrder:  pgtype.Int2{Int: 5},
						VideoScript:   pgtype.Text{String: "video script"},
						Video:         pgtype.Text{String: "video"},
						StudyGuide:    pgtype.Text{String: "study guides"},
						SchoolID:      pgtype.Int4{Int: 5},
						Type:          pgtype.Text{String: "type"},
						Prerequisites: database.TextArray([]string{"hello"}),
					},
				}, nil)
				jsm.On("PublishContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(&natsJS.PubAck{}, nil)
			},
		},
		{
			name: "err topicLearningObjectiveRepo.BulkImport",
			ctx:  ctx,
			req: &pb.AssignTopicItemsRequest{
				TopicId: "topic_id",
				Items: []*pb.AssignTopicItemsRequest_Item{
					{
						ItemId: &pb.AssignTopicItemsRequest_Item_LoId{
							LoId: "lo_id",
						},
						DisplayOrder: 1,
					},
				},
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("TopicLearningObjectiveRepo.BulkImport: %w", fmt.Errorf("%v", "BulkImport error")).Error()),
			setup: func(c context.Context) {
				topicLearningObjectiveRepo.On("BulkImport", ctx, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("%v", "BulkImport error"))
				learningObjectiveRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.LearningObjective{
					{
						ID:            pgtype.Text{String: "topic_id"},
						Name:          pgtype.Text{String: "hello"},
						Country:       pgtype.Text{String: "country"},
						Grade:         pgtype.Int2{Int: 5},
						Subject:       pgtype.Text{String: "subject"},
						TopicID:       pgtype.Text{String: "topic_id"},
						MasterLoID:    pgtype.Text{String: "master loid"},
						DisplayOrder:  pgtype.Int2{Int: 5},
						VideoScript:   pgtype.Text{String: "video script"},
						Video:         pgtype.Text{String: "video"},
						StudyGuide:    pgtype.Text{String: "study guides"},
						SchoolID:      pgtype.Int4{Int: 5},
						Type:          pgtype.Text{String: "type"},
						Prerequisites: database.TextArray([]string{"hello"}),
					},
				}, nil)
			},
		},
		{
			name: "err learningObjectiveRepo.RetrieveByIDs",
			ctx:  ctx,
			req: &pb.AssignTopicItemsRequest{
				TopicId: "topic_id",
				Items: []*pb.AssignTopicItemsRequest_Item{
					{
						ItemId: &pb.AssignTopicItemsRequest_Item_LoId{
							LoId: "lo_id",
						},
						DisplayOrder: 1,
					},
				},
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("LearningObjectiveRep.RetrieveByIDs: %w", fmt.Errorf("%v", "RerieveByIDs error")).Error()),
			setup: func(c context.Context) {
				learningObjectiveRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("%v", "RerieveByIDs error"))
			},
		},
		{
			name: "err bus.Publish",
			ctx:  ctx,
			req: &pb.AssignTopicItemsRequest{
				TopicId: "topic_id",
				Items: []*pb.AssignTopicItemsRequest_Item{
					{
						ItemId: &pb.AssignTopicItemsRequest_Item_LoId{
							LoId: "lo_id",
						},
						DisplayOrder: 1,
					},
				},
			},
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("JSM.PublishContext: subject: %q, %v", constants.SubjectLearningObjectivesCreated, fmt.Errorf("%v", "Publish error")).Error()),
			setup: func(c context.Context) {
				topicLearningObjectiveRepo.On("BulkImport", ctx, mock.Anything, mock.Anything).Once().Return(nil)
				var pre pgtype.TextArray
				pre.Set([]string{"hello"})
				learningObjectiveRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything).Once().Return([]*entities.LearningObjective{
					{
						ID:            pgtype.Text{String: "topic_id"},
						Name:          pgtype.Text{String: "hello"},
						Country:       pgtype.Text{String: "country"},
						Grade:         pgtype.Int2{Int: 5},
						Subject:       pgtype.Text{String: "subject"},
						TopicID:       pgtype.Text{String: "topic_id"},
						MasterLoID:    pgtype.Text{String: "master loid"},
						DisplayOrder:  pgtype.Int2{Int: 5},
						VideoScript:   pgtype.Text{String: "video script"},
						Video:         pgtype.Text{String: "video"},
						StudyGuide:    pgtype.Text{String: "study guides"},
						SchoolID:      pgtype.Int4{Int: 5},
						Type:          pgtype.Text{String: "type"},
						Prerequisites: pre,
					},
				}, nil)
				jsm.On("PublishContext", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("%v", "Publish error"))
			},
		},
		{
			name: "err empty topicID",
			ctx:  ctx,
			req: &pb.AssignTopicItemsRequest{
				TopicId: "",
				Items: []*pb.AssignTopicItemsRequest_Item{
					{
						ItemId: &pb.AssignTopicItemsRequest_Item_LoId{
							LoId: "lo_id",
						},
						DisplayOrder: 1,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "request missing topic_id"),
			setup:       func(c context.Context) {},
		},
		{
			name: "err empty loID",
			ctx:  ctx,
			req: &pb.AssignTopicItemsRequest{
				TopicId: "topic_id",
				Items: []*pb.AssignTopicItemsRequest_Item{
					{
						ItemId: &pb.AssignTopicItemsRequest_Item_LoId{
							LoId: "",
						},
						DisplayOrder: 1,
					},
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "lo_id is required"),
			setup:       func(c context.Context) {},
		},
	}

	for _, testCase := range testCases {
		t.Log("Test case: " + testCase.name)
		testCase.setup(testCase.ctx)
		_, err := svc.AssignTopicItems(testCase.ctx, testCase.req.(*pb.AssignTopicItemsRequest))
		assert.Equal(t, testCase.expectedErr, err)
	}
}
