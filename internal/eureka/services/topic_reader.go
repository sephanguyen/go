package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/i18n"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type TopicReaderService struct {
	Env string
	DB  database.Ext

	StudyPlanRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text) (*entities.StudyPlan, error)
	}

	StudyPlanItemRepo interface {
		FindWithFilter(ctx context.Context, db database.QueryExecer, filter *repositories.StudyPlanItemArgs) ([]*entities.StudyPlanItem, error)
	}

	AssignmentRepo interface {
		RetrieveAssignmentsByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.Assignment, error)
	}

	LearningObjectiveRepo interface {
		RetrieveLearningObjectivesByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.LearningObjective, error)
	}

	BobCourseReaderClient interface {
		RetrieveBookTreeByTopicIDs(ctx context.Context, in *bpb.RetrieveBookTreeByTopicIDsRequest, opts ...grpc.CallOption) (*bpb.RetrieveBookTreeByTopicIDsResponse, error)
	}

	TopicRepo interface {
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.Topic, error)
	}

	BookRepo interface {
		RetrieveBookTreeByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*repositories.BookTreeInfo, error)
	}
}

var (
	ErrMustHaveStudyPlanID = fmt.Errorf("study_plan_id can't be null")
)

func (s *TopicReaderService) ListToDoItemsByTopics(ctx context.Context, req *pb.ListToDoItemsByTopicsRequest) (*pb.ListToDoItemsByTopicsResponse, error) {
	return s.ListToDoItemsByTopicsV2(ctx, req)
}

func getTodoItems(studyPlanItems []*entities.StudyPlanItem) ([]*pb.ToDoItem, error) {
	todoItems := make([]*pb.ToDoItem, 0, len(studyPlanItems))

	for _, item := range studyPlanItems {
		var content entities.ContentStructure
		if err := item.ContentStructure.AssignTo(&content); err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("unable to unmarshal content structure: %w", err).Error())
		}
		todoItem := &pb.ToDoItem{
			StudyPlanItem: toStudyPlanItemPb(item),
		}
		if content.LoID != "" {
			todoItem.Type = pb.ToDoItemType_TO_DO_ITEM_TYPE_LO
			todoItem.ResourceId = content.LoID
		}
		if content.AssignmentID != "" {
			todoItem.Type = pb.ToDoItemType_TO_DO_ITEM_TYPE_ASSIGNMENT
			todoItem.ResourceId = content.AssignmentID
		}

		todoItems = append(todoItems, todoItem)
	}
	return todoItems, nil
}

func (c *TopicReaderService) RetrieveTopics(ctx context.Context, req *pb.RetrieveTopicsRequest) (*pb.RetrieveTopicsResponse, error) {
	topics, err := c.TopicRepo.RetrieveByIDs(ctx, c.DB, database.TextArray(req.TopicIds))
	if err != nil {
		return nil, errors.Wrap(err, "c.TopicRepo.RetrieveByIDs")
	}

	ret := make([]*pb.Topic, 0, len(topics))
	for _, topic := range topics {
		ret = append(ret, ToTopicPb(topic))
	}

	return &pb.RetrieveTopicsResponse{Topics: ret}, nil
}

func ToTopicPb(p *entities.Topic) *pb.Topic {
	var (
		country                = pb.Country(pb.Country_value[p.Country.String])
		bCountry               = bob_pb.Country(bob_pb.Country_value[p.Country.String])
		grade, _               = i18n.ConvertIntGradeToString(bCountry, int(p.Grade.Int))
		publishedAt, updatedAt *timestamppb.Timestamp
	)

	if p.PublishedAt.Get() != nil {
		publishedAt = &timestamppb.Timestamp{Seconds: p.PublishedAt.Time.Unix()}
	} else {
		publishedAt = nil
	}

	if p.UpdatedAt.Get() != nil {
		updatedAt = &timestamppb.Timestamp{Seconds: p.UpdatedAt.Time.Unix()}
	} else {
		updatedAt = nil
	}

	topic := &pb.Topic{
		Id:           p.ID.String,
		Name:         p.Name.String,
		Country:      country,
		Grade:        grade,
		Subject:      pb.Subject(pb.Subject_value[p.Subject.String]),
		Type:         pb.TopicType(pb.TopicType_value[p.TopicType.String]),
		Status:       pb.TopicStatus(pb.TopicStatus_value[p.Status.String]),
		DisplayOrder: int32(p.DisplayOrder.Int),
		CreatedAt: &timestamppb.Timestamp{
			Seconds: p.CreatedAt.Time.Unix(),
		},
		UpdatedAt:     updatedAt,
		PublishedAt:   publishedAt,
		TotalLos:      p.TotalLOs.Int,
		ChapterId:     p.ChapterID.String,
		SchoolId:      getSchool(p.SchoolID.Int),
		Instruction:   p.Instruction.String,
		IconUrl:       p.IconURL.String,
		EssayRequired: p.EssayRequired.Bool,
	}
	if p.CopiedTopicID.Status == pgtype.Present {
		topic.CopiedTopicId = &wrapperspb.StringValue{
			Value: p.CopiedTopicID.String,
		}
	} else {
		topic.CopiedTopicId = nil
	}

	numAttachment := len(p.AttachmentNames.Elements)
	if n := len(p.AttachmentURLs.Elements); n < numAttachment {
		numAttachment = n
	}

	for i := 0; i < numAttachment; i++ {
		topic.Attachments = append(topic.Attachments, &pb.Attachment{
			Name: p.AttachmentNames.Elements[i].String,
			Url:  p.AttachmentURLs.Elements[i].String,
		})
	}

	return topic
}

func getSchool(schoolID int32) int32 {
	if schoolID != 0 {
		return schoolID
	}
	return constants.ManabieSchool
}

/*
1. retrieve book tree from topics include chapterId, topicId, loId and them order
2. retrieve learning objective by topics
3. retrieve assignment by topics
4. sort todo item by lo, assignment using topic map
5. return todo item
*/
func (s *TopicReaderService) ListToDoItemsByTopicsV2(ctx context.Context, req *pb.ListToDoItemsByTopicsRequest) (*pb.ListToDoItemsByTopicsResponse, error) {
	logger := ctxzap.Extract(ctx)
	if req.StudyPlanId == nil {
		return nil, status.Errorf(codes.InvalidArgument, ErrMustHaveStudyPlanID.Error())
	}

	if _, err := s.StudyPlanRepo.FindByID(ctx, s.DB, database.Text(req.StudyPlanId.Value)); err != nil {
		if err.Error() == pgx.ErrNoRows.Error() {
			return nil, status.Errorf(codes.NotFound, "study plan not exist")
		}
		return nil, status.Errorf(codes.Internal, fmt.Errorf("unable to retrieve study plan: %w", err).Error())
	}

	t := time.Now()
	bookInfos, err := s.BookRepo.RetrieveBookTreeByTopicIDs(ctx, s.DB, database.TextArray(req.TopicIds))
	timeRetrieveBookTree := time.Since(t).Milliseconds()

	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("BookRepo.RetrieveBookTreeByTopicIDs: %w", err).Error())
	}

	type Topic struct {
		ChapterID           string
		TopicID             string
		ChapterDisplayOrder int32
		TopicDisplayOrder   int32
	}

	type Assignment struct {
		ChapterID              string
		TopicID                string
		ChapterDisplayOrder    int32
		TopicDisplayOrder      int32
		AssignmentID           string
		AssignmentDisplayOrder int32
	}

	type LearningObjective struct {
		ChapterID           string
		TopicID             string
		ChapterDisplayOrder int32
		TopicDisplayOrder   int32
		LoID                string
		LoDisplayOrder      int32
	}

	topicMap := make(map[string]*Topic)
	loMap := make(map[string]*LearningObjective)
	assignmentMap := make(map[string]*Assignment)

	for _, info := range bookInfos {
		if _, ok := topicMap[info.TopicID.String]; !ok {
			topicMap[info.TopicID.String] = &Topic{
				ChapterID:           info.ChapterID.String,
				TopicID:             info.TopicID.String,
				ChapterDisplayOrder: int32(info.ChapterDisplayOrder.Int),
				TopicDisplayOrder:   int32(info.TopicDisplayOrder.Int),
			}
		}
	}

	t = time.Now()
	learningObjectives, err := s.LearningObjectiveRepo.RetrieveLearningObjectivesByTopicIDs(ctx, s.DB, database.TextArray(req.TopicIds))
	if err != nil {
		timeRetrieveLearningObjectives := time.Since(t).Milliseconds()
		return nil, status.Errorf(codes.Internal, fmt.Errorf("unable to retrieve learning objectives by topics: err %w req %v time retrieve book tree %v time retrieve learning objectives %v", err, req, timeRetrieveBookTree, timeRetrieveLearningObjectives).Error())
	}
	loIDs := make([]string, 0, len(learningObjectives))
	for _, learningObjective := range learningObjectives {
		tp, ok := topicMap[learningObjective.TopicID.String]
		if !ok {
			return nil, status.Errorf(codes.Internal, "topic not in map: %s", learningObjective.TopicID.String)
		}
		loMap[learningObjective.ID.String] = &LearningObjective{
			ChapterID:           tp.ChapterID,
			TopicID:             tp.TopicID,
			LoID:                learningObjective.ID.String,
			ChapterDisplayOrder: tp.ChapterDisplayOrder,
			TopicDisplayOrder:   tp.TopicDisplayOrder,
			LoDisplayOrder:      int32(learningObjective.DisplayOrder.Int),
		}
		loIDs = append(loIDs, learningObjective.ID.String)
	}

	t = time.Now()
	assignments, err := s.AssignmentRepo.RetrieveAssignmentsByTopicIDs(ctx, s.DB, database.TextArray(req.TopicIds))
	if err != nil {
		timeRetrieveAssignments := time.Since(t).Milliseconds()
		return nil, status.Errorf(codes.Internal, fmt.Errorf("unable to retrieve assignments by topics: err %w req %v time retrieve book tree %v time retrieve assignment %v", err, req, timeRetrieveBookTree, timeRetrieveAssignments).Error())
	}
	assignmentIDs := make([]string, 0, len(assignments))
	for _, assignment := range assignments {
		var content entities.AssignmentContent
		if err := assignment.Content.AssignTo(&content); err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("unable to unmarshal assignment content: %w", err).Error())
		}
		tp, ok := topicMap[content.TopicID]
		if !ok {
			return nil, status.Errorf(codes.Internal, "topic not in map: %s", content.TopicID)
		}
		assignmentMap[assignment.ID.String] = &Assignment{
			ChapterID:              tp.ChapterID,
			TopicID:                tp.TopicID,
			AssignmentID:           assignment.ID.String,
			ChapterDisplayOrder:    tp.ChapterDisplayOrder,
			TopicDisplayOrder:      tp.TopicDisplayOrder,
			AssignmentDisplayOrder: assignment.DisplayOrder.Int,
		}
		assignmentIDs = append(assignmentIDs, assignment.ID.String)
	}
	args := &repositories.StudyPlanItemArgs{
		StudyPlanIDs:  database.TextArray([]string{req.StudyPlanId.Value}),
		TopicIDs:      database.TextArray(req.TopicIds),
		LoIDs:         database.TextArray(loIDs),
		AssignmentIDs: database.TextArray(assignmentIDs),
	}

	switch req.StudyPlanItemFilter {
	case pb.StudyPlanItemFilter_STUDY_PLAN_ITEM_FILTER_AVAILABLE:
		args.AvailableDateFilter = true
	}

	studyPlanItems, err := s.StudyPlanItemRepo.FindWithFilter(ctx, s.DB, args)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("unable to retrieve study plan items: %w", err).Error())
	}
	if len(studyPlanItems) == 0 {
		return &pb.ListToDoItemsByTopicsResponse{
			Items: []*pb.ListToDoItemsByTopicsResponse_ToDoItemsByTopic{},
		}, nil
	}
	todoItems, err := getTodoItems(studyPlanItems)
	if err != nil {
		return nil, err
	}
	getDO := func(item *pb.ToDoItem) (int, int, int) {
		switch item.Type {
		case pb.ToDoItemType_TO_DO_ITEM_TYPE_LO:
			lo, ok := loMap[item.ResourceId]
			if !ok {
				logger.Error("TopicReaderService.ListToDoItemsByTopics: item.ResourceId not found in loMap", zap.String("lo_id", item.ResourceId))
				return 0, 0, 0
			}
			return int(lo.ChapterDisplayOrder), int(lo.TopicDisplayOrder), int(lo.LoDisplayOrder)
		case pb.ToDoItemType_TO_DO_ITEM_TYPE_ASSIGNMENT:
			assignment, ok := assignmentMap[item.ResourceId]
			if !ok {
				logger.Error("TopicReaderService.ListToDoItemsByTopics: item.ResourceId not found in assignmentMap", zap.String("assignment_id", item.ResourceId))
				return 0, 0, 0
			}
			return int(assignment.ChapterDisplayOrder), int(assignment.TopicDisplayOrder), int(assignment.AssignmentDisplayOrder)
		}
		return 0, 0, 0
	}

	sort.SliceStable(todoItems, func(i, j int) bool {
		cdo1, tdo1, ldo1 := getDO(todoItems[i])
		cdo2, tdo2, ldo2 := getDO(todoItems[j])
		if cdo1 != cdo2 {
			return cdo1 < cdo2
		}
		if tdo1 != tdo2 {
			return tdo1 < tdo2
		}
		return ldo1 < ldo2
	})
	var result []*pb.ListToDoItemsByTopicsResponse_ToDoItemsByTopic
	currentItem := &pb.ListToDoItemsByTopicsResponse_ToDoItemsByTopic{}
	for _, item := range todoItems {
		topicID := item.StudyPlanItem.ContentStructure.TopicId
		if currentItem.TopicId != topicID {
			currentItem = &pb.ListToDoItemsByTopicsResponse_ToDoItemsByTopic{
				TopicId:   topicID,
				TodoItems: []*pb.ToDoItem{},
			}
			result = append(result, currentItem)
		}
		currentItem.TodoItems = append(currentItem.TodoItems, item)
	}
	return &pb.ListToDoItemsByTopicsResponse{
		Items: result,
	}, nil
}
