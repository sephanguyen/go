package services

import (
	"context"
	"fmt"
	"time"

	entities "github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type LearningObjectiveModifierService struct {
	DB  database.Ext
	JSM nats.JetStreamManagement

	TopicRepo interface {
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.Topic, error)
		UpdateTotalLOs(ctx context.Context, db database.QueryExecer, topicID pgtype.Text) error
		RetrieveByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Topic, error)
		UpdateLODisplayOrderCounter(ctx context.Context, db database.QueryExecer, topicID pgtype.Text, number pgtype.Int4) error
	}

	LearningObjectiveRepo interface {
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningObjective, error)
		BulkImport(ctx context.Context, db database.QueryExecer, learningObjectives []*entities.LearningObjective) error
		SoftDeleteWithLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (int64, error)
		UpdateDisplayOrders(ctx context.Context, db database.QueryExecer, mDisplayOrder map[pgtype.Text]pgtype.Int2) error
		CountTotal(ctx context.Context, db database.QueryExecer) (*pgtype.Int8, error)
		UpdateName(ctx context.Context, db database.QueryExecer, loID pgtype.Text, name pgtype.Text) (int64, error)
	}

	BookChapterRepo interface {
		RetrieveContentStructuresByLOs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (map[string]entities.ContentStructure, error)
	}

	TopicsLearningObjectivesRepo interface {
		SoftDeleteByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
		BulkImport(context.Context, database.QueryExecer, []*entities.TopicsLearningObjectives) error
		BulkUpdateDisplayOrder(ctx context.Context, db database.QueryExecer, topicsLearningsObjectives []*entities.TopicsLearningObjectives) error
	}

	LoStudyPlanItemRepo interface {
		DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
	}
}

func NewLearningObjectiveModifierService(db database.Ext, jsm nats.JetStreamManagement) *LearningObjectiveModifierService {
	return &LearningObjectiveModifierService{
		DB:                           db,
		JSM:                          jsm,
		TopicRepo:                    new(repositories.TopicRepo),
		LearningObjectiveRepo:        new(repositories.LearningObjectiveRepo),
		BookChapterRepo:              new(repositories.BookChapterRepo),
		LoStudyPlanItemRepo:          new(repositories.LoStudyPlanItemRepo),
		TopicsLearningObjectivesRepo: new(repositories.TopicsLearningObjectivesRepo),
	}
}
func (s *LearningObjectiveModifierService) UpsertLOs(ctx context.Context, req *epb.UpsertLOsRequest) (*epb.UpsertLOsResponse, error) {
	ids := make([]string, 0, len(req.LearningObjectives))
	var (
		topicIDs    []string
		isAutoGenDo bool
		isInserted  bool
	)
	type Group struct {
		Los      []*entities.LearningObjective
		TopicLos []*entities.TopicsLearningObjectives
		LoIDs    []string
	}

	topicMap := make(map[string]*Group)
	lodoMap := make(map[string]int32)

	for i, lo := range req.LearningObjectives {
		// validate
		if err := validateUpsertLO(lo); err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("learningObjectives[%d].%s", i, err.Error()))
		}

		e, isAutoGenID, err := toLOEntity(lo)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if isAutoGenID {
			isInserted = true
		}

		loTopicE := &entities.TopicsLearningObjectives{
			TopicID:      e.TopicID,
			LoID:         e.ID,
			DisplayOrder: e.DisplayOrder,
			CreatedAt:    e.CreatedAt,
			UpdatedAt:    e.UpdatedAt,
			DeletedAt:    e.DeletedAt,
		}
		topicID := e.TopicID.String
		if _, ok := topicMap[topicID]; !ok {
			topicIDs = append(topicIDs, topicID)
			topicMap[topicID] = &Group{}
		}
		topicMap[topicID].Los = append(topicMap[topicID].Los, e)
		topicMap[topicID].TopicLos = append(topicMap[topicID].TopicLos, loTopicE)
		topicMap[topicID].LoIDs = append(topicMap[topicID].LoIDs, e.ID.String)
		ids = append(ids, e.ID.String)
	}

	topics, err := s.TopicRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(topicIDs))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("unable to retrieve topics by ids: %w", err).Error())
	}

	if !isTopicsExisted(topicIDs, topics) {
		return nil, status.Errorf(codes.InvalidArgument, "some topics does not exists")
	}

	for topicID, group := range topicMap {
		los := group.Los
		topicLos := group.TopicLos
		if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			var (
				insertNum int32
				count     int16
				total     int16
			)
			topic, err := s.TopicRepo.RetrieveByID(ctx, tx, database.Text(topicID), repositories.WithUpdateLock())
			if err != nil {
				return fmt.Errorf("unable to retrieve topic by id: %w", err)
			}
			if isAutoGenLODisplayOrder(los) {
				isAutoGenDo = true
				existedLos, err := s.LearningObjectiveRepo.RetrieveByIDs(ctx, tx, database.TextArray(ids))
				if err != nil {
					return fmt.Errorf("unable to retrieve los by ids: %w", err)
				}
				m := make(map[string]*entities.LearningObjective)
				for _, lo := range existedLos {
					m[lo.ID.String] = lo
				}

				if topic.LODisplayOrderCounter.Status == pgtype.Present {
					total = int16(topic.LODisplayOrderCounter.Int)
				}

				for i, lo := range los {
					if e, ok := m[lo.ID.String]; !ok {
						lo.DisplayOrder = database.Int2(total + count + 1)
						topicLos[i].DisplayOrder = lo.DisplayOrder
						lodoMap[lo.ID.String] = int32(lo.DisplayOrder.Int)
						count++
					} else {
						lo.DisplayOrder = e.DisplayOrder
						topicLos[i].DisplayOrder = lo.DisplayOrder
						lodoMap[lo.ID.String] = int32(lo.DisplayOrder.Int)
					}
				}
				insertNum = int32(count)
			}
			if err := s.LearningObjectiveRepo.BulkImport(ctx, tx, los); err != nil {
				return fmt.Errorf("unable to bulk import learning objective: %w", err)
			}
			if err := s.TopicsLearningObjectivesRepo.BulkImport(ctx, tx, topicLos); err != nil {
				return fmt.Errorf("unable to bulk import topic learning objective: %w", err)
			}
			if err := s.TopicRepo.UpdateLODisplayOrderCounter(ctx, tx, database.Text(topicID), database.Int4(insertNum)); err != nil {
				return fmt.Errorf("unable to update lo display order counter: %w", err)
			}

			if err := s.TopicRepo.UpdateTotalLOs(ctx, tx, database.Text(topicID)); err != nil {
				return fmt.Errorf("unable to update total learing objectives: %w", err)
			}
			return nil
		}); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ForeignKeyViolation {
				return nil, status.Error(codes.FailedPrecondition, err.Error())
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if isInserted {
		for _, lo := range req.LearningObjectives {
			if isAutoGenDo {
				lo.Info.DisplayOrder = lodoMap[lo.Info.Id]
			}
		}

		resp := map[string]*npb.ContentStructures{}

		data, err := s.BookChapterRepo.RetrieveContentStructuresByLOs(
			ctx,
			s.DB,
			database.TextArray(ids),
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cm.BookChapterRepo.RetrieveContentStructuresByLOs: %v", err)
		}

		for loID, contentStructures := range data {
			resp[loID] = &npb.ContentStructures{
				ContentStructures: []*epb.ContentStructure{s.toContentStructuresPb(contentStructures)},
			}
		}
		msg, err := proto.Marshal(&npb.EventLearningObjectivesCreated{
			LearningObjectives:  req.LearningObjectives,
			LoContentStructures: resp,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("UpsertLOs: proto.Marshal: %v", err).Error())
		}

		if _, err = s.JSM.PublishContext(ctx, constants.SubjectLearningObjectivesCreated, msg); err != nil {
			return nil, fmt.Errorf("s.JSM.PublishContext: subject: %q, %v", constants.SubjectLearningObjectivesCreated, err)
		}
	}

	return &epb.UpsertLOsResponse{
		LoIds: ids,
	}, nil
}

func (s *LearningObjectiveModifierService) toContentStructuresPb(cs entities.ContentStructure) *epb.ContentStructure {
	return &epb.ContentStructure{
		ChapterId: cs.ChapterID,
		BookId:    cs.BookID,
		TopicId:   cs.TopicID,
	}
}

func toLOEntity(src *cpb.LearningObjective) (*entities.LearningObjective, bool, error) {
	isAutoGenID := false
	if src.Info.Id == "" {
		src.Info.Id = idutil.ULIDNow()
		isAutoGenID = true
	}

	if src.Type == cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_NONE {
		src.Type = cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_LEARNING
	}

	if src.VendorType == cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_NONE {
		src.VendorType = cpb.LearningMaterialVendorType_LM_VENDOR_TYPE_MANABIE
	}

	e := new(entities.LearningObjective)
	database.AllNullEntity(e)

	err := multierr.Combine(
		e.ID.Set(src.Info.Id),
		e.Name.Set(src.Info.Name),
		e.Country.Set(src.Info.Country.String()),
		e.Grade.Set(src.Info.Grade),
		e.Subject.Set(src.Info.Subject.String()),
		e.TopicID.Set(src.TopicId),
		// e.VideoScript.Set(src.Video),
		e.DisplayOrder.Set(src.Info.DisplayOrder),
		e.Prerequisites.Set(src.Prerequisites),
		e.Video.Set(src.Video),
		e.StudyGuide.Set(src.StudyGuide),
		e.SchoolID.Set(src.Info.SchoolId),
		e.Type.Set(src.Type.String()),
		e.ManualGrading.Set(src.ManualGrading),
		e.ApproveGrading.Set(src.ApproveGrading),
		e.GradeCapping.Set(src.GradeCapping),
		e.ReviewOption.Set(src.GetReviewOption()),
		e.VendorType.Set(src.VendorType.String()),
	)
	if src.TimeLimit != nil {
		err = multierr.Append(err, e.TimeLimit.Set(src.TimeLimit.Value))
	}

	if src.Instruction != "" {
		err = multierr.Append(err, e.Instruction.Set(src.Instruction))
	}

	if src.GradeToPass != nil {
		err = multierr.Append(err, e.GradeToPass.Set(src.GradeToPass.Value))
	}

	if src.Info.MasterId != "" {
		err = multierr.Append(err, e.MasterLoID.Set(src.Info.MasterId))
	}

	if src.MaximumAttempt != nil {
		err = multierr.Append(err, e.MaximumAttempt.Set(src.MaximumAttempt.Value))
	}

	if src.Info.CreatedAt != nil {
		err = multierr.Append(err, e.CreatedAt.Set(src.Info.CreatedAt.AsTime()))
	} else {
		err = multierr.Append(err, e.CreatedAt.Set(time.Now().UTC()))
	}

	if src.Info.UpdatedAt != nil {
		err = multierr.Append(err, e.UpdatedAt.Set(src.Info.UpdatedAt.AsTime()))
	} else {
		e.UpdatedAt = e.CreatedAt
	}

	if err != nil {
		return nil, isAutoGenID, fmt.Errorf("toLOEntity: %v", err)
	}

	return e, isAutoGenID, nil
}

func validateUpsertLO(lo *cpb.LearningObjective) error {
	if ma := lo.MaximumAttempt; ma != nil && (ma.Value < 1 || ma.Value > 99) {
		return fmt.Errorf("maximum_attempt must be Null or between 1 to 99")
	}
	return nil
}

func (s *LearningObjectiveModifierService) DeleteLos(ctx context.Context, req *epb.DeleteLosRequest) (*epb.DeleteLosResponse, error) {
	existedLos, err := s.LearningObjectiveRepo.RetrieveByIDs(ctx, s.DB, database.TextArray(req.GetLoIds()))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Errorf("LearningObjectiveRepo.RetrieveByIDs: %w", err).Error())
	}

	m := make(map[string]bool)
	for _, lo := range existedLos {
		m[lo.ID.String] = true
	}

	for _, loID := range req.GetLoIds() {
		if ok := m[loID]; !ok {
			return nil, status.Errorf(codes.NotFound, "lo %v is not exist", loID)
		}
	}

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := s.LearningObjectiveRepo.SoftDeleteWithLoIDs(ctx, tx, database.TextArray(req.GetLoIds())); err != nil {
			return fmt.Errorf("s.LoRepo.SoftDeleteWithLoIDs: %w", err)
		}
		if err := s.LoStudyPlanItemRepo.DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs(ctx, tx, database.TextArray(req.GetLoIds())); err != nil {
			return fmt.Errorf("LoStudyPlanItemRepo.DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs: %w", err)
		}
		if err := s.TopicsLearningObjectivesRepo.SoftDeleteByLoIDs(ctx, tx, database.TextArray(req.GetLoIds())); err != nil {
			return fmt.Errorf("TopicLearningObjectiveRepo.SoftDeleteByLoIDs: %w", err)
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &epb.DeleteLosResponse{Successful: true}, nil
}

func (s *LearningObjectiveModifierService) UpdateDisplayOrdersOfLOs(ctx context.Context, topicLODisplayOrders []*epb.TopicLODisplayOrder) ([]*epb.TopicLO, error) {
	type UpdateData struct {
		TopicLOs          []*entities.TopicsLearningObjectives
		LODisplayOrderMap map[pgtype.Text]pgtype.Int2
	}

	now := database.Timestamptz(time.Now())
	mLOsByTopic := make(map[string]*UpdateData)
	los := make([]*epb.TopicLO, 0, len(topicLODisplayOrders))

	for _, lo := range topicLODisplayOrders {
		// Ignore invalid data
		if lo.TopicId == "" || lo.LoId == "" {
			continue
		}

		if _, ok := mLOsByTopic[lo.TopicId]; !ok {
			mLOsByTopic[lo.TopicId] = &UpdateData{LODisplayOrderMap: make(map[pgtype.Text]pgtype.Int2)}
		}
		data := mLOsByTopic[lo.TopicId]
		data.TopicLOs = append(
			data.TopicLOs,
			&entities.TopicsLearningObjectives{
				TopicID:      database.Text(lo.TopicId),
				LoID:         database.Text(lo.LoId),
				DisplayOrder: database.Int2(int16(lo.DisplayOrder)),
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		)
		data.LODisplayOrderMap[database.Text(lo.LoId)] = database.Int2(int16(lo.DisplayOrder))

		los = append(los, &epb.TopicLO{
			LoId:    lo.LoId,
			TopicId: lo.TopicId,
		})
	}

	// Should execute in separated transaction for each group of TopicLOs by Topic
	// to avoid redundance blocking time for another queries.
	for _, data := range mLOsByTopic {
		if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			// Update TopicsLOs
			if err := s.TopicsLearningObjectivesRepo.BulkUpdateDisplayOrder(ctx, tx, data.TopicLOs); err != nil {
				return fmt.Errorf("s.TopicsLearningObjectivesRepo.BulkUpdateDisplayOrder: %w", err)
			}
			// Update DisplayOrder of LearningObjectives
			if err := s.LearningObjectiveRepo.UpdateDisplayOrders(ctx, tx, data.LODisplayOrderMap); err != nil {
				return fmt.Errorf("s.LearningObjectiveRepo.UpdateDisplayOrders: %w", err)
			}

			return nil
		}); err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Sprintf("ExecInTx: %s", err.Error()))
		}
	}

	return los, nil
}

func validateUpdateLearningObjectiveNameRequest(req *epb.UpdateLearningObjectiveNameRequest) error {
	if req.LoId == "" {
		return fmt.Errorf("missing field LoId")
	}
	if req.NewLearningObjectiveName == "" {
		return fmt.Errorf("missing field NewLearningObjectiveName")
	}

	return nil
}

func (s *LearningObjectiveModifierService) UpdateLearningObjectiveName(ctx context.Context, req *epb.UpdateLearningObjectiveNameRequest) (*epb.UpdateLearningObjectiveNameResponse, error) {
	if err := validateUpdateLearningObjectiveNameRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("validateUpdateLearningObjectiveNameRequest: %w", err).Error())
	}
	rowAffected, err := s.LearningObjectiveRepo.UpdateName(ctx, s.DB, database.Text(req.LoId), database.Text(req.NewLearningObjectiveName))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.LearningObjectiveRepo.UpdateName: %w", err).Error())
	}
	if rowAffected == 0 {
		return nil, status.Error(codes.NotFound, fmt.Errorf("s.LearningObjectiveRepo.UpdateName not found any learning objective to update name: %w", pgx.ErrNoRows).Error())
	}
	return &epb.UpdateLearningObjectiveNameResponse{}, nil
}
