package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LearningMaterialService struct {
	sspb.UnimplementedLearningMaterialServer
	DB database.Ext

	LearningMaterialRepo interface {
		Delete(ctx context.Context, db database.QueryExecer, lmID pgtype.Text) error
		FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningMaterial, error)
		UpdateDisplayOrders(ctx context.Context, db database.QueryExecer, lms []*entities.LearningMaterial) error
		UpdateName(ctx context.Context, db database.QueryExecer, lmID pgtype.Text, lmName pgtype.Text) (int64, error)
	}

	BookRepo interface {
		DuplicateBook(ctx context.Context, db database.QueryExecer, bookID pgtype.Text, bookName pgtype.Text) (string, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string]*entities.Book, error)
	}

	BookChapterRepo interface {
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string][]*entities.BookChapter, error)
		Upsert(ctx context.Context, db database.Ext, cc []*entities.BookChapter) error
	}

	ChapterRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, chapters []*entities.Chapter) error
		FindByBookID(ctx context.Context, db database.QueryExecer, BookID string) (map[string]*entities.Chapter, error)
		DuplicateChapters(ctx context.Context, db database.QueryExecer, bookID string, chapterIDs []string) ([]*entities.CopiedChapter, error)
	}

	TopicRepo interface {
		RetrieveByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Topic, error)
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs, topicIDs pgtype.TextArray, limit, offset pgtype.Int4) ([]*entities.Topic, error)
		DuplicateTopics(ctx context.Context, db database.QueryExecer, chapterIDs pgtype.TextArray, newChapterIDs pgtype.TextArray) ([]*entities.CopiedTopic, error)
	}

	LearningObjectiveRepoV2 interface {
		ListByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.LearningObjectiveV2, error)
		ListByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.LearningObjectiveV2, error)

		BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.LearningObjectiveV2) error
	}

	GeneralAssignmentRepo interface {
		List(ctx context.Context, db database.QueryExecer, learningMaterialIds pgtype.TextArray) ([]*entities.GeneralAssignment, error)
		ListByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.GeneralAssignment, error)
		BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.GeneralAssignment) error
	}

	FlashcardRepo interface {
		ListFlashcard(ctx context.Context, db database.QueryExecer, args *repositories.ListFlashcardArgs) ([]*entities.Flashcard, error)
		ListByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.Flashcard, error)
		BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.Flashcard) error
	}
	ExamLORepo interface {
		ListByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.ExamLO, error)
		ListByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.ExamLO, error)
		BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.ExamLO) error
	}
	TaskAssignmentRepo interface {
		List(ctx context.Context, db database.QueryExecer, learningMaterialIds pgtype.TextArray) ([]*entities.TaskAssignment, error)
		ListByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.TaskAssignment, error)
		BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.TaskAssignment) error
	}
}

func NewLearningMaterialService(db database.Ext) *LearningMaterialService {
	return &LearningMaterialService{
		DB:                      db,
		LearningMaterialRepo:    &repositories.LearningMaterialRepo{},
		BookRepo:                &repositories.BookRepo{},
		BookChapterRepo:         &repositories.BookChapterRepo{},
		ChapterRepo:             &repositories.ChapterRepo{},
		TopicRepo:               &repositories.TopicRepo{},
		LearningObjectiveRepoV2: &repositories.LearningObjectiveRepoV2{},
		GeneralAssignmentRepo:   &repositories.GeneralAssignmentRepo{},
		FlashcardRepo:           &repositories.FlashcardRepo{},
		ExamLORepo:              &repositories.ExamLORepo{},
		TaskAssignmentRepo:      &repositories.TaskAssignmentRepo{},
	}
}

func (s *LearningMaterialService) DeleteLearningMaterial(ctx context.Context, req *sspb.DeleteLearningMaterialRequest) (*sspb.DeleteLearningMaterialResponse, error) {
	if err := validateDeleteLearningMaterialRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Errorf("validateLearningMaterialReq: %w", err).Error())
	}

	if err := s.LearningMaterialRepo.Delete(ctx, s.DB, database.Text(req.LearningMaterialId)); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Errorf("LearningMaterialRepo.Delete: %w", err).Error())
	}

	return &sspb.DeleteLearningMaterialResponse{}, nil
}

func validateDeleteLearningMaterialRequest(req *sspb.DeleteLearningMaterialRequest) error {
	if req.LearningMaterialId == "" {
		return fmt.Errorf("LearningMaterial ID must not be empty")
	}
	return nil
}

func validateSwapDisplayOrderRequest(req *sspb.SwapDisplayOrderRequest) error {
	if req.FirstLearningMaterialId == "" {
		return fmt.Errorf("missing FirstLearningMaterialId")
	}
	if req.SecondLearningMaterialId == "" {
		return fmt.Errorf("missing SecondLearningMaterialId")
	}
	return nil
}

func validateSwapDisplayOrderLearningMaterial(lms []*entities.LearningMaterial) error {
	if len(lms) != 2 {
		return fmt.Errorf("missing LearningMaterials")
	}
	if lms[0].TopicID.String != lms[1].TopicID.String {
		return fmt.Errorf("LearningMaterials not in the same topic")
	}
	return nil
}

func (s *LearningMaterialService) SwapDisplayOrder(ctx context.Context, req *sspb.SwapDisplayOrderRequest) (*sspb.SwapDisplayOrderResponse, error) {
	if err := validateSwapDisplayOrderRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("validateSwapDisplayOrderRequest: %w", err).Error())
	}
	lms, err := s.LearningMaterialRepo.FindByIDs(ctx, s.DB, database.TextArray([]string{req.FirstLearningMaterialId, req.SecondLearningMaterialId}))
	if err != nil {
		return nil, fmt.Errorf("s.LearningMaterialRepo.FindByIDs: %w", err)
	}
	if err := validateSwapDisplayOrderLearningMaterial(lms); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("validateSwapDisplayOrderLearningMaterial: %w", err).Error())
	}

	lms[0].DisplayOrder, lms[1].DisplayOrder = lms[1].DisplayOrder, lms[0].DisplayOrder

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := s.TopicRepo.RetrieveByID(ctx, tx, lms[0].TopicID, repositories.WithUpdateLock()); err != nil {
			return fmt.Errorf("s.TopicRepo.RetrieveByID: %w", err)
		}
		if err := s.LearningMaterialRepo.UpdateDisplayOrders(ctx, tx, lms); err != nil {
			return fmt.Errorf("s.LearningMaterialRepo.UpdateDisplayOrders: %w", err)
		}
		return nil
	}); err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("database.ExecInTx: %w", err).Error())
	}

	return &sspb.SwapDisplayOrderResponse{}, nil
}

func validateListLearningMaterialRequest(req *sspb.ListLearningMaterialRequest) error {
	if req.GetAssignment() == nil && req.GetExamLo() == nil && req.GetFlashcard() == nil && req.GetLearningObjective() == nil && req.GetTaskAssignment() == nil {
		return fmt.Errorf("missing ListLearningMaterialsRequest")
	}

	return nil
}

func (s *LearningMaterialService) ListLearningMaterial(ctx context.Context, req *sspb.ListLearningMaterialRequest) (*sspb.ListLearningMaterialResponse, error) {
	if err := validateListLearningMaterialRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("validateListLearningMaterialRequest: %w", err).Error())
	}
	var message interface{} = req.Message
	switch message.(type) {
	case *sspb.ListLearningMaterialRequest_Assignment:
		generalAssignments, err := s.GeneralAssignmentRepo.List(ctx, s.DB, database.TextArray(req.GetAssignment().LearningMaterialIds))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, status.Error(codes.NotFound, fmt.Errorf("s.GeneralAssignmentRepo.List: %w", err).Error())
			}
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.GeneralAssignmentRepo.List: %w", err).Error())
		}
		var baseAssignments []*sspb.AssignmentBase
		for _, assignment := range generalAssignments {
			assignmentBase, err := ToAssignmentPb(assignment)
			if err != nil {
				return nil, fmt.Errorf("cannot convert assignment to assignment base, err: %w", err)
			}
			baseAssignments = append(baseAssignments, assignmentBase)
		}
		return &sspb.ListLearningMaterialResponse{
			Message: &sspb.ListLearningMaterialResponse_Assignment{
				Assignment: &sspb.ListAssignmentResponse{
					Assignments: baseAssignments,
				},
			},
		}, nil
	case *sspb.ListLearningMaterialRequest_ExamLo:
		examLOs, err := s.ExamLORepo.ListByIDs(ctx, s.DB, database.TextArray(req.GetExamLo().LearningMaterialIds))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, status.Error(codes.NotFound, fmt.Errorf("s.ExamLORepo.ListByIDs: %w", err).Error())
			}
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.ExamLORepo.ListByIDs: %w", err).Error())
		}
		var baseExamLOs []*sspb.ExamLOBase
		for _, examLO := range examLOs {
			baseExamLOs = append(baseExamLOs, ToBaseExamLO(examLO))
		}
		resp := sspb.ListLearningMaterialResponse{
			Message: &sspb.ListLearningMaterialResponse_ExamLo{
				ExamLo: &sspb.ListExamLOResponse{
					ExamLos: baseExamLOs,
				},
			},
		}
		return &resp, nil
	case *sspb.ListLearningMaterialRequest_TaskAssignment:
		taskAssignments, err := s.TaskAssignmentRepo.List(ctx, s.DB, database.TextArray(req.GetTaskAssignment().LearningMaterialIds))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, status.Error(codes.NotFound, fmt.Errorf("s.TaskAssignmentRepo.List: %w", err).Error())
			}
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.TaskAssignmentRepo.List: %w", err).Error())
		}
		var baseTaskAssignments []*sspb.TaskAssignmentBase
		for _, taskAssignment := range taskAssignments {
			baseTaskAssignment, err := ToTaskAssignmentPb(taskAssignment)
			if err != nil {
				return nil, fmt.Errorf("cannot convert task assignment to base task assignment, err: %w", err)
			}
			baseTaskAssignments = append(baseTaskAssignments, baseTaskAssignment)
		}
		return &sspb.ListLearningMaterialResponse{
			Message: &sspb.ListLearningMaterialResponse_TaskAssignment{
				TaskAssignment: &sspb.ListTaskAssignmentResponse{
					TaskAssignments: baseTaskAssignments,
				},
			},
		}, nil
	case *sspb.ListLearningMaterialRequest_LearningObjective:
		los, err := s.LearningObjectiveRepoV2.ListByIDs(ctx, s.DB, database.TextArray(req.GetLearningObjective().LearningMaterialIds))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, status.Error(codes.NotFound, fmt.Errorf("s.LearningObjectiveRepoV2.ListByIDs: %w", err).Error())
			}
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.LearningObjectiveRepoV2.ListByIDs: %w", err).Error())
		}
		var baseLOs []*sspb.LearningObjectiveBase
		for _, lo := range los {
			baseLOs = append(baseLOs, ToBaseLearningObjectiveV2(lo))
		}
		return &sspb.ListLearningMaterialResponse{
			Message: &sspb.ListLearningMaterialResponse_LearningObjective{
				LearningObjective: &sspb.ListLearningObjectiveResponse{
					LearningObjectives: baseLOs,
				},
			},
		}, nil
	case *sspb.ListLearningMaterialRequest_Flashcard:
		flashcards, err := s.FlashcardRepo.ListFlashcard(ctx, s.DB, &repositories.ListFlashcardArgs{
			LearningMaterialIDs: database.TextArray(req.GetFlashcard().LearningMaterialIds),
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, status.Error(codes.NotFound, fmt.Errorf("s.FlashcardRepo.ListFlashcard: %w", err).Error())
			}
			return nil, status.Errorf(codes.Internal, fmt.Errorf("s.FlashcardRepo.ListFlashcard: %w", err).Error())
		}
		var baseFlashcards []*sspb.FlashcardBase
		for _, flashcard := range flashcards {
			baseFlashcards = append(baseFlashcards, ToBaseFlashcard(flashcard))
		}
		return &sspb.ListLearningMaterialResponse{
			Message: &sspb.ListLearningMaterialResponse_Flashcard{
				Flashcard: &sspb.ListFlashcardResponse{
					Flashcards: baseFlashcards,
				},
			},
		}, nil
	}
	return &sspb.ListLearningMaterialResponse{}, nil
}
func validateDuplicateBookRequest(req *sspb.DuplicateBookRequest) error {
	if req.BookId == "" {
		return fmt.Errorf("missing BookId")
	}
	if req.BookName == "" {
		return fmt.Errorf("missing BookName")
	}
	return nil
}

func (s *LearningMaterialService) DuplicateBook(ctx context.Context, req *sspb.DuplicateBookRequest) (*sspb.DuplicateBookResponse, error) {
	if err := validateDuplicateBookRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("validateDuplicateBookRequest: %w", err).Error())
	}
	orgBooks, err := s.BookRepo.FindByIDs(ctx, s.DB, []string{req.BookId})
	if err != nil {
		return nil, fmt.Errorf("s.BookRepo.FindByIDs: %w", err)
	}
	if book, ok := orgBooks[req.BookId]; !ok || book.Name.String == "" {
		return nil, fmt.Errorf("NotFound Book")
	}

	var newBookID string
	orgChapterMap, err := s.ChapterRepo.FindByBookID(ctx, s.DB, req.BookId)
	if err != nil {
		return nil, fmt.Errorf("s.ChapterRepo.FindByBookID: %w", err)
	}
	var (
		chapterIDs  = make([]string, 0)
		orgTopicIDs = make([]string, 0)
		newTopicIDs = make([]string, 0)
		mapTopicIDs map[string]string
	)
	for id := range orgChapterMap {
		chapterIDs = append(chapterIDs, id)
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// duplicate book
		newBookID, err = s.BookRepo.DuplicateBook(ctx, tx, database.Text(req.BookId), database.Text(req.BookName))
		if err != nil {
			return fmt.Errorf("s.BookRepo.DuplicateBook: %w", err)
		}
		if len(chapterIDs) > 0 && newBookID != "" { // if book contain chapters
			// duplicate chapters
			orgChapterIDs, newChapterIDs, err := s.duplicateChapters(ctx, tx, newBookID, chapterIDs)
			if err != nil {
				return fmt.Errorf("s.duplicateChapters: %w", err)
			}
			// duplicate topics, Please make sure orgChapterIDs and newChapterIDs are in the same order
			orgTopicIDs, newTopicIDs, mapTopicIDs, err = s.duplicateTopics(ctx, tx, orgChapterIDs, newChapterIDs)
			if err != nil {
				return fmt.Errorf("s.duplicateTopics: %w", err)
			}

			// get and duplicate general assignments
			if err := s.duplicateGeneralAssignment(ctx, tx, orgTopicIDs, mapTopicIDs); err != nil {
				return err
			}
			// get and duplicate flashcards
			if err := s.duplicateFlashcard(ctx, tx, orgTopicIDs, mapTopicIDs); err != nil {
				return err
			}
			// get and duplicate learning objective
			if err := s.duplicateLO(ctx, tx, orgTopicIDs, mapTopicIDs); err != nil {
				return err
			}
			// get and duplicate exam lo
			if err := s.duplicateExamLO(ctx, tx, orgTopicIDs, mapTopicIDs); err != nil {
				return err
			}

			// get and duplicate task assignment
			if err := s.duplicateTaskAssignment(ctx, tx, orgTopicIDs, mapTopicIDs); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sspb.DuplicateBookResponse{
		NewBookID:  newBookID,
		OldTopicId: orgTopicIDs,
		NewTopicId: newTopicIDs,
	}, nil
}

func (s *LearningMaterialService) duplicateGeneralAssignment(ctx context.Context, db database.QueryExecer, orgTopicIDs []string, mapTopicIDs map[string]string) error {
	generalAssignments, err := s.GeneralAssignmentRepo.ListByTopicIDs(ctx, db, database.TextArray(orgTopicIDs))
	if err != nil {
		return fmt.Errorf("s.GeneralAssignmentRepo.ListByTopicIDs: %w", err)
	}
	for _, generalAssignment := range generalAssignments {
		oldTopicID := generalAssignment.TopicID.String
		if err := multierr.Combine(
			generalAssignment.ID.Set(idutil.ULIDNow()),
			generalAssignment.TopicID.Set(database.Text(mapTopicIDs[oldTopicID])),
			generalAssignment.IsPublished.Set(false),
		); err != nil {
			return fmt.Errorf("can't set new topic id for general assignment: %w", err)
		}
	}
	if err := s.GeneralAssignmentRepo.BulkInsert(ctx, db, generalAssignments); err != nil {
		return fmt.Errorf("s.GeneralAssignmentRepo.BulkInsert: %w", err)
	}
	return nil
}

func (s *LearningMaterialService) duplicateFlashcard(ctx context.Context, db database.QueryExecer, orgTopicIDs []string, mapTopicIDs map[string]string) error {
	flashcards, err := s.FlashcardRepo.ListByTopicIDs(ctx, db, database.TextArray(orgTopicIDs))
	if err != nil {
		return fmt.Errorf("s.FlashcardRepo.ListByTopicIDs: %w", err)
	}
	for _, flashcard := range flashcards {
		oldTopicID := flashcard.TopicID.String
		if err := multierr.Combine(
			flashcard.ID.Set(idutil.ULIDNow()),
			flashcard.TopicID.Set(database.Text(mapTopicIDs[oldTopicID])),
			flashcard.IsPublished.Set(false),
		); err != nil {
			return fmt.Errorf("can't set new topic id for flashcard: %w", err)
		}
	}
	if err := s.FlashcardRepo.BulkInsert(ctx, db, flashcards); err != nil {
		return fmt.Errorf("s.FlashcardRepo.BulkInsert: %w", err)
	}
	return nil
}

func (s *LearningMaterialService) duplicateLO(ctx context.Context, db database.QueryExecer, orgTopicIDs []string, mapTopicIDs map[string]string) error {
	los, err := s.LearningObjectiveRepoV2.ListByTopicIDs(ctx, db, database.TextArray(orgTopicIDs))
	if err != nil {
		return fmt.Errorf("s.LearningObjectiveRepoV2.ListByTopicIDs: %w", err)
	}
	for _, lo := range los {
		oldTopicID := lo.TopicID.String
		if err := multierr.Combine(
			lo.ID.Set(idutil.ULIDNow()),
			lo.TopicID.Set(database.Text(mapTopicIDs[oldTopicID])),
			lo.IsPublished.Set(false),
		); err != nil {
			return fmt.Errorf("can't set new topic id for lo: %w", err)
		}
	}
	if err := s.LearningObjectiveRepoV2.BulkInsert(ctx, db, los); err != nil {
		return fmt.Errorf("s.LearningObjectiveRepoV2.BulkInsert: %w", err)
	}
	return nil
}

func (s *LearningMaterialService) duplicateExamLO(ctx context.Context, db database.QueryExecer, orgTopicIDs []string, mapTopicIDs map[string]string) error {
	examLOs, err := s.ExamLORepo.ListByTopicIDs(ctx, db, database.TextArray(orgTopicIDs))
	if err != nil {
		return fmt.Errorf("s.ExamLORepo.ListByTopicIDs: %w", err)
	}
	for _, examLO := range examLOs {
		oldTopicID := examLO.TopicID.String
		if err := multierr.Combine(
			examLO.ID.Set(idutil.ULIDNow()),
			examLO.TopicID.Set(database.Text(mapTopicIDs[oldTopicID])),
			examLO.IsPublished.Set(false),
		); err != nil {
			return fmt.Errorf("can't set new topic id for exam lo: %w", err)
		}
	}
	if err := s.ExamLORepo.BulkInsert(ctx, db, examLOs); err != nil {
		return fmt.Errorf("s.ExamLORepo.BulkInsert: %w", err)
	}
	return nil
}

func (s *LearningMaterialService) duplicateTaskAssignment(ctx context.Context, db database.QueryExecer, orgTopicIDs []string, mapTopicIDs map[string]string) error {
	taskAssignments, err := s.TaskAssignmentRepo.ListByTopicIDs(ctx, db, database.TextArray(orgTopicIDs))
	if err != nil {
		return fmt.Errorf("s.TaskAssignmentRepo.ListByTopicIDs: %w", err)
	}
	for _, taskAssignment := range taskAssignments {
		oldTopicID := taskAssignment.TopicID.String
		if err := multierr.Combine(
			taskAssignment.ID.Set(idutil.ULIDNow()),
			taskAssignment.TopicID.Set(database.Text(mapTopicIDs[oldTopicID])),
			taskAssignment.IsPublished.Set(false),
		); err != nil {
			return fmt.Errorf("can't set new topic id for task assignment: %w", err)
		}
	}
	if err := s.TaskAssignmentRepo.BulkInsert(ctx, db, taskAssignments); err != nil {
		return fmt.Errorf("s.TaskAssignmentRepo.BulkInsert: %w", err)
	}
	return nil
}

func (s *LearningMaterialService) duplicateChapters(ctx context.Context, db database.QueryExecer, newBookID string, chapterIDs []string) ([]string, []string, error) {
	copiedChapters, err := s.ChapterRepo.DuplicateChapters(ctx, db, newBookID, chapterIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("s.ChapterRepo.DuplicateChapters: %w", err)
	}

	orgChapterIDs := make([]string, 0, len(copiedChapters))
	newChapterIDs := make([]string, 0, len(copiedChapters))

	for _, copiedChapter := range copiedChapters {
		orgChapterIDs = append(orgChapterIDs, copiedChapter.CopyFromID.String)
		newChapterIDs = append(newChapterIDs, copiedChapter.ID.String)
	}

	return orgChapterIDs, newChapterIDs, nil
}

func (s *LearningMaterialService) duplicateTopics(ctx context.Context, db database.QueryExecer, orgChapterIDs []string, newChapterIDs []string) ([]string, []string, map[string]string, error) {
	copiedTopics, err := s.TopicRepo.DuplicateTopics(ctx, db, database.TextArray(orgChapterIDs), database.TextArray(newChapterIDs))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("s.TopicRepo.DuplicateTopics: %w", err)
	}

	var mapTopicID = make(map[string]string, len(copiedTopics))
	newTopicIDs := make([]string, len(copiedTopics))
	orgTopicIDs := make([]string, len(copiedTopics))
	for i, copiedTopic := range copiedTopics {
		newTopicIDs[i] = copiedTopic.ID.String
		orgTopicIDs[i] = copiedTopic.CopyFromID.String
		mapTopicID[orgTopicIDs[i]] = newTopicIDs[i]
	}
	return orgTopicIDs, newTopicIDs, mapTopicID, nil
}

func validateUpdateLearningMaterialNameRequest(req *sspb.UpdateLearningMaterialNameRequest) error {
	if req.LearningMaterialId == "" {
		return fmt.Errorf("missing field LearningMaterialId")
	}
	if req.NewLearningMaterialName == "" {
		return fmt.Errorf("missing field NewLearningMaterialName")
	}

	return nil
}

func (s *LearningMaterialService) UpdateLearningMaterialName(ctx context.Context, req *sspb.UpdateLearningMaterialNameRequest) (*sspb.UpdateLearningMaterialNameResponse, error) {
	if err := validateUpdateLearningMaterialNameRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("validateUpdateLearningMaterialNameRequest: %w", err).Error())
	}
	rowAffected, err := s.LearningMaterialRepo.UpdateName(ctx, s.DB, database.Text(req.LearningMaterialId), database.Text(req.NewLearningMaterialName))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Errorf("s.LearningMaterialRepo.UpdateName: %w", err).Error())
	}
	if rowAffected == 0 {
		return nil, status.Error(codes.NotFound, fmt.Errorf("s.LearningMaterialRepo.UpdateName not found any learning material to update name: %w", pgx.ErrNoRows).Error())
	}
	return &sspb.UpdateLearningMaterialNameResponse{}, nil
}
