package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InternalReaderService struct {
	bpb.UnimplementedInternalReaderServiceServer

	CheckClientVersions []string
	DB                  database.Ext
	EurekaDBTrace       database.Ext
	BookChapterRepo     interface {
		RetrieveContentStructuresByLOs(context.Context, database.QueryExecer, pgtype.TextArray) (map[string][]repositories.ContentStructure, error)
		RetrieveContentStructuresByTopics(context.Context, database.QueryExecer, pgtype.TextArray) (map[string][]repositories.ContentStructure, error)
	}
	TopicRepo interface {
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs, topicIDs pgtype.TextArray, limit, offset pgtype.Int4) ([]*entities.Topic, error)
		RetrieveBookTopic(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.BookTopic, error)
	}
	TopicsLearningObjectivesRepo interface {
		RetrieveByLoIDs(
			ctx context.Context, db database.QueryExecer,
			loIDs pgtype.TextArray,
		) ([]*repositories.TopicLearningObjective, error)
	}
	CoursesBooksRepo interface {
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) ([]*entities.CoursesBooks, error)
	}
	LearningObjectiveRepo interface {
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningObjective, error)
		RetrieveByTopicIDs(ctx context.Context, db database.QueryExecer, topicIds pgtype.TextArray) ([]*entities.LearningObjective, error)
		RetrieveBookLoByIntervalTime(ctx context.Context, db database.QueryExecer, intervalTime pgtype.Text) ([]*entities.BookLearningObjective, error)
	}
	QuizSetRepo interface {
		CountQuizOnLO(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (map[string]int32, error)
	}
	StudentsLearningObjectivesCompletenessRepo interface {
		Find(ctx context.Context, db database.QueryExecer, studentId pgtype.Text, loIds pgtype.TextArray) (map[pgtype.Text]*entities.StudentsLearningObjectivesCompleteness, error)
	}

	CourseReaderServiceClient interface {
		RetrieveLOs(ctx context.Context, req *epb.RetrieveLOsRequest, opts ...grpc.CallOption) (*epb.RetrieveLOsResponse, error)
	}
}

func toContentStructuresPb(contentStructures []repositories.ContentStructure) []*epb.ContentStructure {
	ret := make([]*epb.ContentStructure, 0, len(contentStructures))
	for _, cs := range contentStructures {
		ret = append(ret, &epb.ContentStructure{
			ChapterId: cs.ChapterID,
			BookId:    cs.BookID,
			TopicId:   cs.TopicID,
		})
	}
	return ret
}

func (s *InternalReaderService) RetrieveTopics(ctx context.Context, req *bpb.RetrieveTopicsRequest) (_ *bpb.RetrieveTopicsResponse, err error) {
	var (
		topics   []*entities.Topic
		nextPage *cpb.Paging
	)

	if req.Paging != nil {
		if req.Paging.GetOffsetInteger() < 0 {
			return nil, status.Error(codes.InvalidArgument, "offset must be positive")
		}

		if req.Paging.Limit <= 0 {
			req.Paging.Limit = 100
		}

		offset := req.Paging.GetOffsetInteger()
		limit := req.Paging.Limit

		nextPage = &cpb.Paging{
			Limit: uint32(limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: int64(limit) + int64(offset),
			},
		}

		topics, err = s.TopicRepo.FindByBookIDs(ctx, s.EurekaDBTrace, database.TextArray(req.BookIds), database.TextArray(req.TopicIds), database.Int4(int32(limit)), database.Int4(int32(offset)))
	} else {
		topics, err = s.TopicRepo.FindByBookIDs(ctx, s.EurekaDBTrace, database.TextArray(req.BookIds), database.TextArray(req.TopicIds), pgtype.Int4{Status: pgtype.Null}, pgtype.Int4{Status: pgtype.Null})
	}

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.TopicRepo.FindByBookID: %w", err).Error())
	}

	topicPbs := make([]*cpb.Topic, 0, len(topics))
	for _, topic := range topics {
		topicPbs = append(topicPbs, ToTopicPbV1(topic))
	}

	resp := &bpb.RetrieveTopicsResponse{
		Items:    topicPbs,
		NextPage: nextPage,
	}

	return resp, nil
}

func (i *InternalReaderService) VerifyAppVersion(ctx context.Context, req *bpb.VerifyAppVersionRequest) (*bpb.VerifyAppVersionResponse, error) {
	joinClientVersions := strings.Join(i.CheckClientVersions, ",")
	if err := interceptors.CheckForceUpdateApp(ctx, joinClientVersions); err != nil {
		return nil, err
	}

	return &bpb.VerifyAppVersionResponse{
		IsValid: true,
	}, nil
}
