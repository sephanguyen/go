package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type TopicModifierService struct {
	DB  database.Ext
	JSM nats.JetStreamManagement

	ChapterRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, chapterID pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Chapter, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, chapterIDs []string) (map[string]*entities.Chapter, error)
		UpdateCurrentTopicDisplayOrder(ctx context.Context, db database.QueryExecer, totalGeneratedTopicDisplayOrder pgtype.Int4, chapterID pgtype.Text) error
	}

	TopicRepo interface {
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.Topic, error)
		BulkImport(ctx context.Context, db database.QueryExecer, topics []*entities.Topic) error
		UpdateTotalLOs(ctx context.Context, db database.QueryExecer, topicID pgtype.Text) error
		BulkUpsertWithoutDisplayOrder(ctx context.Context, db database.QueryExecer, topics []*entities.Topic) error
		UpdateStatus(ctx context.Context, db database.Ext, ids pgtype.TextArray, topicStatus pgtype.Text) error
		FindByIDsV2(ctx context.Context, db database.QueryExecer, ids []string, isAll bool) (map[string]*entities.Topic, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, topicIDs []string) (int, error)
	}

	LearningObjectiveRepo interface {
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningObjective, error)
	}

	TopicsLearningObjectivesRepo interface {
		BulkImport(context.Context, database.QueryExecer, []*entities.TopicsLearningObjectives) error
	}
}

func NewTopicModifierService(db database.Ext, jsm nats.JetStreamManagement) *TopicModifierService {
	return &TopicModifierService{
		DB:                           db,
		JSM:                          jsm,
		ChapterRepo:                  new(repositories.ChapterRepo),
		TopicRepo:                    new(repositories.TopicRepo),
		LearningObjectiveRepo:        new(repositories.LearningObjectiveRepo),
		TopicsLearningObjectivesRepo: new(repositories.TopicsLearningObjectivesRepo),
	}
}

func toTopicEntity(p *epb.Topic) *entities.Topic {
	ep := new(entities.Topic)
	database.AllNullEntity(ep)
	if p.Id == "" {
		p.Id = idutil.ULIDNow()
	}

	attachmentNames := make([]string, 0)
	attachmentURLs := make([]string, 0)
	if len(p.Attachments) != 0 {
		for _, each := range p.Attachments {
			attachmentNames = append(attachmentNames, each.Name)
			attachmentURLs = append(attachmentURLs, each.Url)
		}
	}

	err := multierr.Combine(
		ep.ID.Set(p.Id),
		ep.Name.Set(p.Name),
		ep.Country.Set(cpb.Country_COUNTRY_NONE.String()),
		ep.Grade.Set(-1),
		ep.Subject.Set(p.Subject.String()),
		ep.TopicType.Set(p.Type.String()),
		ep.ChapterID.Set(p.ChapterId),
		ep.DisplayOrder.Set(int16(p.DisplayOrder)),
		ep.IconURL.Set(p.IconUrl),
		ep.SchoolID.Set(p.SchoolId),
		ep.Instruction.Set(p.Instruction),
		ep.AttachmentNames.Set(attachmentNames),
		ep.AttachmentURLs.Set(attachmentURLs),
		ep.EssayRequired.Set(p.EssayRequired),
		ep.TotalLOs.Set(p.TotalLos),
		ep.LODisplayOrderCounter.Set(0),
	)

	if p.Status != epb.TopicStatus_TOPIC_STATUS_NONE {
		err = multierr.Append(err, ep.Status.Set(p.Status.String()))
	}

	if p.ChapterId == "" {
		err = multierr.Append(err, ep.ChapterID.Set(nil))
	}

	if p.CreatedAt != nil {
		err = multierr.Append(err, ep.CreatedAt.Set(time.Unix(p.CreatedAt.Seconds, int64(p.CreatedAt.Nanos))))
	} else {
		err = multierr.Append(err, ep.CreatedAt.Set(time.Now()))
	}

	if p.UpdatedAt != nil {
		err = multierr.Append(err, ep.UpdatedAt.Set(time.Unix(p.UpdatedAt.Seconds, int64(p.UpdatedAt.Nanos))))
	} else {
		err = multierr.Append(err, ep.UpdatedAt.Set(time.Now()))
	}

	if p.PublishedAt != nil {
		err = multierr.Append(err, ep.PublishedAt.Set(time.Unix(p.PublishedAt.Seconds, int64(p.PublishedAt.Nanos))))
	} else {
		err = multierr.Append(err, ep.PublishedAt.Set(time.Now()))
	}

	if err != nil {
		return nil
	}

	return ep
}

func (s *TopicModifierService) convertTopics2MapTopics(topics []*entities.Topic) map[string]*entities.Topic {
	mapTopics := make(map[string]*entities.Topic)
	for _, topic := range topics {
		mapTopics[topic.ID.String] = topic
	}
	return mapTopics
}

// convertTopics2MapChapterTopics: with key: chapterID and value is a array topics belong to its
func (s *TopicModifierService) convertTopics2MapChapterTopics(topics []*entities.Topic) map[string][]*entities.Topic {
	mapTopics := make(map[string][]*entities.Topic)
	for _, topic := range topics {
		if topic.ChapterID.String != "" {
			mapTopics[topic.ChapterID.String] = append(mapTopics[topic.ChapterID.String], topic)
		} else {
			mapTopics[""] = append(mapTopics[""], topic)
		}
	}
	return mapTopics
}

// validateExistedTopicOnDB validate the topic existed is have chapter_id correctly on its request or not
// targetTopics is request topics
// sourceTopics is topics retrieved on DB
func (s *TopicModifierService) validateExistedTopicOnDB(targetTopics, sourceTopics []*entities.Topic) bool {
	mapSourceTopics := s.convertTopics2MapTopics(sourceTopics)
	for _, targetTopic := range targetTopics {
		if topic, ok := mapSourceTopics[targetTopic.ID.String]; ok {
			if topic.ChapterID.String != targetTopic.ChapterID.String {
				return false
			}
		}
	}
	return true
}

func (s *TopicModifierService) retrieveChapterIDsFromTopics(topics []*entities.Topic) []string {
	chapterIDs := make([]string, 0)
	for _, t := range topics {
		chapterIDs = append(chapterIDs, t.ChapterID.String)
	}
	return golibs.GetUniqueElementStringArray(chapterIDs)
}

func (s *TopicModifierService) validateChapterOnDB(chapterIDs []string, mapChapters map[string]*entities.Chapter) error {
	for _, cID := range chapterIDs {
		if cID == "" {
			return fmt.Errorf("chapter_id must not empty")
		}
	}
	if len(mapChapters) != len(chapterIDs) {
		return fmt.Errorf("some chapters aren't exist")
	}
	return nil
}

func (s *TopicModifierService) retrieveTopicIDsFromTopicEntities(topics []*entities.Topic) []string {
	topicIDs := make([]string, 0, len(topics))
	for _, topic := range topics {
		topicIDs = append(topicIDs, topic.ID.String)
	}
	return topicIDs
}

// isAutomaticDisplayOrderOnTopic check for each element on map with key and value
func (s *TopicModifierService) isAutomaticDisplayOrderOnTopic(key string, topics []*entities.Topic) bool {
	if key == "" {
		return false
	}
	for _, topic := range topics {
		if topic.DisplayOrder.Int != 0 {
			return false
		}
	}
	return true
}

func (s *TopicModifierService) updateTotalLOs(db database.QueryExecer, topicIDs []string, logger *zap.Logger) {
	if len(topicIDs) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, topicID := range topicIDs {
		if err := s.TopicRepo.UpdateTotalLOs(ctx, db, database.Text(topicID)); err != nil {
			logger.Error("c.TopicRepo.UpdateTotalLOs", zap.String("topic_id", topicID), zap.Error(err))
		}
	}
}

// Upsert upsert topics with according chapter
// if the topics existed -> will check the chapter info with topic on the request is correct on db or not
// because on a request have multiple chapters so we have to divide to each batch according to chapter
// check each batch should automatically or not, if not -> use old flow
// -> yes, go to new flow, lock the chapter row, add display order to the topics which not existed on database-> upsert
func (s *TopicModifierService) Upsert(ctx context.Context, req *epb.UpsertTopicsRequest) (*epb.UpsertTopicsResponse, error) {
	topicInputs := make([]*entities.Topic, 0, len(req.Topics))
	ids := make([]string, 0, len(req.Topics))

	for _, topic := range req.Topics {
		e := toTopicEntity(topic)
		if e == nil {
			// Ignore line with wrong data
			continue
		}

		ids = append(ids, e.ID.String)
		topicInputs = append(topicInputs, e)
	}
	// valid the existed is correct with chapter id or not after convert it to
	allTopicsFromDB, err := s.TopicRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(ids))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("TopicRepo.RetrieveByIDs: %w", err).Error())
	}
	if !s.validateExistedTopicOnDB(topicInputs, allTopicsFromDB) {
		return nil, status.Error(codes.InvalidArgument, "the existed topic_id have to according with chapter_id")
	}
	chapterIDs := s.retrieveChapterIDsFromTopics(topicInputs)
	chapters, err := s.ChapterRepo.FindByIDs(ctx, s.DB, chapterIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("ChapterRepo.FindByIDs: %w", err).Error())
	}
	if err := s.validateChapterOnDB(chapterIDs, chapters); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mapChapterTopics := s.convertTopics2MapChapterTopics(topicInputs)
	mapChapterTopicsFromDB := s.convertTopics2MapChapterTopics(allTopicsFromDB)
	for cID, topics := range mapChapterTopics {
		topicIDs := s.retrieveTopicIDsFromTopicEntities(topics)

		if err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			chapter, err := s.ChapterRepo.FindByID(ctx, tx, database.Text(cID), repositories.WithUpdateLock())
			if err != nil {
				return fmt.Errorf("UpsertTopics.ChapterRepo.FindByID: %w", err)
			}
			topicsFromDB := make([]*entities.Topic, 0)
			if topics, ok := mapChapterTopicsFromDB[cID]; ok {
				topicsFromDB = topics
			}
			if !s.isAutomaticDisplayOrderOnTopic(cID, topics) {
				totalInsertTopic := len(topics) - len(topicsFromDB)
				if err := s.TopicRepo.BulkImport(ctx, tx, topics); err != nil {
					return fmt.Errorf("TopicRepo.BulkImport: %w", err)
				}
				if err = s.ChapterRepo.UpdateCurrentTopicDisplayOrder(ctx, tx, database.Int4(int32(totalInsertTopic)), database.Text(cID)); err != nil {
					return fmt.Errorf("ChapterRepo.UpdateCurrentTopicDisplayOrder: %w", err)
				}
			} else {
				currenDisplayOrder := chapter.CurrentTopicDisplayOrder.Int
				var totalGeneratedTopicDisplayOrder int32
				mapTopicsFromDB := s.convertTopics2MapTopics(topicsFromDB)
				for _, c := range topics {
					if _, ok := mapTopicsFromDB[c.ID.String]; !ok {
						totalGeneratedTopicDisplayOrder++
						c.DisplayOrder.Set(currenDisplayOrder + totalGeneratedTopicDisplayOrder)
					}
				}
				if err = s.TopicRepo.BulkUpsertWithoutDisplayOrder(ctx, tx, topics); err != nil {
					return fmt.Errorf("TopicRepo.BulkUpsertWithoutDisplayOrder: %w", err)
				}
				if err = s.ChapterRepo.UpdateCurrentTopicDisplayOrder(ctx, tx, database.Int4(totalGeneratedTopicDisplayOrder), database.Text(cID)); err != nil {
					return fmt.Errorf("ChapterRepo.UpdateCurrentTopicDisplayOrder: %w", err)
				}
			}
			s.updateTotalLOs(tx, topicIDs, ctxzap.Extract(ctx))
			return nil
		}); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &epb.UpsertTopicsResponse{
		TopicIds: ids,
	}, nil
}

func (s *TopicModifierService) Publish(ctx context.Context, req *epb.PublishTopicsRequest) (*epb.PublishTopicsResponse, error) {
	publishStatus := cpb.TopicStatus_TOPIC_STATUS_PUBLISHED.String()
	topicIds := req.TopicIds

	err := s.TopicRepo.UpdateStatus(ctx, s.DB, database.TextArray(topicIds), database.Text(publishStatus))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &epb.PublishTopicsResponse{
		IsSuccess: true,
	}, nil
}

func isAutoGenLODisplayOrder(los entities.LearningObjectives) bool {
	for _, lo := range los {
		if lo.DisplayOrder.Int != 0 {
			return false
		}
	}
	return true
}

func (s *TopicModifierService) DeleteTopics(ctx context.Context, req *epb.DeleteTopicsRequest) (*epb.DeleteTopicsResponse, error) {
	topicMap, err := s.TopicRepo.FindByIDsV2(ctx, s.DB, req.GetTopicIds(), false)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("unable to find topics by ids: %w", err).Error())
	}

	for _, topicID := range req.TopicIds {
		if _, ok := topicMap[topicID]; !ok {
			return nil, status.Errorf(codes.InvalidArgument, "topic %v does not exists", topicID)
		}
	}

	_, err = s.TopicRepo.SoftDelete(ctx, s.DB, req.TopicIds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("unable to delete topics: %w", err).Error())
	}

	return &epb.DeleteTopicsResponse{Successful: true}, nil
}

func (s *TopicModifierService) AssignTopicItems(ctx context.Context, req *epb.AssignTopicItemsRequest) (*epb.AssignTopicItemsResponse, error) {
	if req.TopicId == "" {
		return nil, status.Error(codes.InvalidArgument, "request missing topic_id")
	}

	var topicsLearningObjects []*entities.TopicsLearningObjectives
	loIDs := make([]string, 0, len(req.Items))
	mDisplayOrder := make(map[string]int32)

	for _, item := range req.Items {
		if item.GetLoId() != "" {
			topicsLearningObjects = append(topicsLearningObjects, &entities.TopicsLearningObjectives{
				TopicID:      database.Text(req.TopicId),
				LoID:         database.Text(item.GetLoId()),
				DisplayOrder: database.Int2(int16(item.GetDisplayOrder())),
			})

			loIDs = append(loIDs, item.GetLoId())
			mDisplayOrder[item.GetLoId()] = item.GetDisplayOrder()
		}
	}

	if len(topicsLearningObjects) == 0 {
		return nil, status.Error(codes.InvalidArgument, "lo_id is required")
	}
	los, err := s.LearningObjectiveRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(loIDs))
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("LearningObjectiveRep.RetrieveByIDs: %w", err).Error())
	}

	if err := s.TopicsLearningObjectivesRepo.BulkImport(ctx, s.DB, topicsLearningObjects); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("TopicLearningObjectiveRepo.BulkImport: %w", err).Error())
	}

	loPbs := make([]*cpb.LearningObjective, 0, len(los))
	for _, lo := range los {
		var prerequisites []string
		_ = lo.Prerequisites.AssignTo(&prerequisites)
		loPbs = append(loPbs, &cpb.LearningObjective{
			Info: &cpb.ContentBasicInfo{
				Id:           lo.ID.String,
				Name:         lo.Name.String,
				SchoolId:     lo.SchoolID.Int,
				DisplayOrder: mDisplayOrder[lo.ID.String],
				MasterId:     lo.MasterLoID.String,
			},
			TopicId:       lo.TopicID.String,
			Video:         lo.Video.String,
			StudyGuide:    lo.StudyGuide.String,
			Prerequisites: prerequisites,
			Type:          cpb.LearningObjectiveType(cpb.LearningObjectiveType_value[lo.Type.String]),
		})
	}

	data := &npb.EventLearningObjectivesCreated{
		LearningObjectives: loPbs,
	}
	msg, _ := proto.Marshal(data)
	if _, err = s.JSM.PublishContext(ctx, constants.SubjectLearningObjectivesCreated, msg); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("JSM.PublishContext: subject: %q, %v", constants.SubjectLearningObjectivesCreated, err).Error())
	}

	return &epb.AssignTopicItemsResponse{}, nil
}
