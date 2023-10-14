package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	IStudyPlanItemRepository interface {
		BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) error
		BulkSync(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) ([]*entities.StudyPlanItem, error)
		FindByStudyPlanID(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text) ([]*entities.StudyPlanItem, error)
		BulkCopy(ctx context.Context, db database.QueryExecer, originalStudyPlanIDs pgtype.TextArray, newStudyPlanIDs pgtype.TextArray) error
		UpdateWithCopiedFromItem(ctx context.Context, db database.QueryExecer, studyPlanItems []*entities.StudyPlanItem) error
		UpdateCompletedAtByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, completedAt pgtype.Timestamptz) error
		SoftDeleteByStudyPlanItemIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error
		FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.StudyPlanItem, error)
		DeleteStudyPlanItemsByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
		DeleteStudyPlanItemsByStudyPlans(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		UpdateSchoolDate(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, studentID pgtype.Text, schoolDate pgtype.Timestamptz) error
		UpdateStatus(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, studentID pgtype.Text, status pgtype.Text) error
	}

	IStudyPlanRepository interface {
		FindByID(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text) (*entities.StudyPlan, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlan) error
		RetrieveStudyPlanItemInfo(ctx context.Context, db database.QueryExecer, args repositories.StudyPlanItemInfoArgs) ([]*repositories.StudyPlanItemInfo, error)
	}

	IStudentStudyPlanRepository interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, studentStudyPlans []*entities.StudentStudyPlan) error
	}

	IAssignmentStudyPlanItemRepository interface {
		BulkInsert(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error
		FindByStudyPlanItemIDs(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*entities.AssignmentStudyPlanItem, error)
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		EditAssignmentTime(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, studyPlanItemIDs pgtype.TextArray, startDate, endDate pgtype.Timestamptz) error
		SoftDeleteByAssigmentIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) (pgtype.TextArray, error)
		BulkEditAssignmentTime(ctx context.Context, db database.QueryExecer, studentID pgtype.Text, ens []*entities.StudyPlanItem) error
		BulkUpsertByStudyPlanItem(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error
	}

	ILoStudyPlanItemRepository interface {
		BulkInsert(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.LoStudyPlanItem) error
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		FindByStudyPlanItemIDs(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*entities.LoStudyPlanItem, error)
		DeleteLoStudyPlanItemsByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
		DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
	}

	ITopicsAssignmentsRepository interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, topicsAssignmentsList []*entities.TopicsAssignments) error
	}

	IAssignmentRepository interface {
		RetrieveAssignmentsByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.Assignment, error)
		RetrieveAssignments(ctx context.Context, db database.QueryExecer, assignmentIDs pgtype.TextArray) ([]*entities.Assignment, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, assignments []*entities.Assignment) error
	}

	ILearningObjectiveRepository interface {
		RetrieveLearningObjectivesByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.LearningObjective, error)
	}

	IBookRepository interface {
		FindByID(ctx context.Context, db database.QueryExecer, bookID pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Book, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string]*entities.Book, error)
		RetrieveBookTreeByBookID(ctx context.Context, db database.QueryExecer, bookID pgtype.Text) ([]*repositories.BookTreeInfo, error)
		Upsert(ctx context.Context, db database.Ext, cc []*entities.Book) error
		UpdateCurrentChapterDisplayOrder(ctx context.Context, db database.QueryExecer, totalGeneratedChapterDisplayOrder pgtype.Int4, bookID pgtype.Text) error
		RetrieveAdHocBookByCourseIDAndStudentID(ctx context.Context, db database.QueryExecer, courseID, studentID pgtype.Text) (*entities.Book, error)
	}

	IChapterRepository interface {
		FindByID(ctx context.Context, db database.QueryExecer, chapterID pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Chapter, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, chapterIDs []string) (map[string]*entities.Chapter, error)
		Upsert(ctx context.Context, db database.QueryExecer, cc []*entities.Chapter) error
		UpsertWithoutDisplayOrderWhenUpdate(ctx context.Context, db database.QueryExecer, cc []*entities.Chapter) error
		UpdateCurrentTopicDisplayOrder(ctx context.Context, db database.QueryExecer, totalGeneratedTopicDisplayOrder pgtype.Int4, chapterID pgtype.Text) error
	}

	IBookChapterRepository interface {
		Upsert(ctx context.Context, db database.Ext, cc []*entities.BookChapter) error
		SoftDelete(ctx context.Context, db database.QueryExecer, chapterIDs, bookIDs pgtype.TextArray) error
		SoftDeleteByChapterIDs(ctx context.Context, db database.QueryExecer, chapterIDs []string) error
		RetrieveContentStructuresByTopics(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) (map[string][]entities.ContentStructure, error)
	}

	ITopicRepository interface {
		RetrieveByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.Topic, error)
		BulkImport(ctx context.Context, db database.QueryExecer, topics []*entities.Topic) error
		UpdateTotalLOs(ctx context.Context, db database.QueryExecer, topicID pgtype.Text) error
		BulkUpsertWithoutDisplayOrder(ctx context.Context, db database.QueryExecer, topics []*entities.Topic) error
		UpdateStatus(ctx context.Context, db database.Ext, ids pgtype.TextArray, topicStatus pgtype.Text) error
		RetrieveByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Topic, error)
		UpdateLODisplayOrderCounter(ctx context.Context, db database.QueryExecer, topicID pgtype.Text, number pgtype.Int4) error
		FindByIDsV2(ctx context.Context, db database.QueryExecer, ids []string, isAll bool) (map[string]*entities.Topic, error)
		FindByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs, topicIDs pgtype.TextArray, limit, offset pgtype.Int4) ([]*entities.Topic, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, topicIDs []string) (int, error)
	}

	ICourseBookRepository interface {
		FindByCourseIDAndBookID(ctx context.Context, db database.QueryExecer, bookID, courseID pgtype.Text) (*entities.CoursesBooks, error)
		Upsert(ctx context.Context, db database.Ext, cc []*entities.CoursesBooks) error
	}

	IMasterStudyPlanRepository interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.MasterStudyPlan) error
	}

	IIndividualStudyPlanRepository interface {
		BulkSync(ctx context.Context, db database.QueryExecer, args []*entities.IndividualStudyPlan) ([]*entities.IndividualStudyPlan, error)
	}
)

type InternalModifierService struct {
	DB database.Ext

	LoStudyPlanItemRepo interface {
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, loStudyPlanItems []*entities.LoStudyPlanItem) error
		DeleteLoStudyPlanItemsByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
		DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
	}
	StudyPlanItemRepo interface {
		BulkCopy(ctx context.Context, db database.QueryExecer, originalStudyPlanIDs pgtype.TextArray, newStudyPlanIDs pgtype.TextArray) error
		BulkSync(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) ([]*entities.StudyPlanItem, error)
		DeleteStudyPlanItemsByLoIDs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) error
		DeleteStudyPlanItemsByStudyPlans(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) error
		UpdateWithCopiedFromItem(ctx context.Context, db database.QueryExecer, studyPlanItems []*entities.StudyPlanItem) error
		UpdateSchoolDate(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, studentID pgtype.Text, schoolDate pgtype.Timestamptz) error
		UpdateStatus(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray, studentID pgtype.Text, status pgtype.Text) error
	}

	StudyPlanRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text) (*entities.StudyPlan, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlan) error
		RetrieveStudyPlanItemInfo(ctx context.Context, db database.QueryExecer, args repositories.StudyPlanItemInfoArgs) ([]*repositories.StudyPlanItemInfo, error)
	}

	CourseBookRepo interface {
		FindByCourseIDAndBookID(ctx context.Context, db database.QueryExecer, bookID, courseID pgtype.Text) (*entities.CoursesBooks, error)
		Upsert(ctx context.Context, db database.Ext, cc []*entities.CoursesBooks) error
	}

	StudentRepo interface {
		FindStudentsByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*pgtype.TextArray, error)
	}

	BookRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, bookID pgtype.Text, enhancers ...repositories.QueryEnhancer) (*entities.Book, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, bookIDs []string) (map[string]*entities.Book, error)
		RetrieveBookTreeByBookID(ctx context.Context, db database.QueryExecer, bookID pgtype.Text) ([]*repositories.BookTreeInfo, error)
		Upsert(ctx context.Context, db database.Ext, cc []*entities.Book) error
		UpdateCurrentChapterDisplayOrder(ctx context.Context, db database.QueryExecer, totalGeneratedChapterDisplayOrder pgtype.Int4, bookID pgtype.Text) error
		RetrieveAdHocBookByCourseIDAndStudentID(ctx context.Context, db database.QueryExecer, courseID, studentID pgtype.Text) (*entities.Book, error)
	}

	AssignmentRepo interface {
		RetrieveAssignmentsByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.Assignment, error)
		RetrieveAssignments(ctx context.Context, db database.QueryExecer, assignmentIDs pgtype.TextArray) ([]*entities.Assignment, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, assignments []*entities.Assignment) error
	}

	LearningObjectiveRepo interface {
		RetrieveLearningObjectivesByTopicIDs(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) ([]*entities.LearningObjective, error)
	}

	AssignmentStudyPlanItemRepo interface {
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error
		BulkUpsertByStudyPlanItem(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error
	}

	StudentStudyPlanRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, studentStudyPlans []*entities.StudentStudyPlan) error
	}
}

// DeleteLOStudyPlanItems for yasuo call when deletelos in currently
func (s *InternalModifierService) DeleteLOStudyPlanItems(ctx context.Context, req *epb.DeleteLOStudyPlanItemsRequest) (*epb.DeleteLOStudyPlanItemsResponse, error) {
	if req.LoIds == nil {
		return nil, status.Error(codes.InvalidArgument, "lo_ids must not be empty")
	}

	err := s.LoStudyPlanItemRepo.DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs(ctx, s.DB, database.TextArray(req.GetLoIds()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "DeleteLoStudyPlanItemsAndStudyPlanItemByLoIDs: %s", err.Error())
	}

	return &epb.DeleteLOStudyPlanItemsResponse{}, nil
}

func convertUpsertAdHocIndividualStudyPlanRequestToStudyPlanEntity(src *epb.UpsertAdHocIndividualStudyPlanRequest) (*entities.StudyPlan, error) {
	now := timeutil.Now()
	e := &entities.StudyPlan{}
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.ID.Set(idutil.ULIDNow()),
		e.Name.Set(src.Name),
		e.SchoolID.Set(src.SchoolId),
		e.CourseID.Set(src.CourseId),
		e.BookID.Set(src.BookId),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
		e.StudyPlanType.Set(epb.StudyPlanType_STUDY_PLAN_TYPE_INDIVIDUAL.String()),
		e.Status.Set(src.Status.String()),
		e.TrackSchoolProgress.Set(false),
		e.Grades.Set(src.Grades),
	); err != nil {
		return nil, fmt.Errorf("error create study plan: %w", err)
	}
	if len(src.Grades) == 0 {
		e.Grades.Set("{}")
	}
	return e, nil
}

func verifyUpsertAdHocIndividualStudyPlan(
	ctx context.Context,
	db database.Ext,
	req *epb.UpsertAdHocIndividualStudyPlanRequest,
	courseBookRepo ICourseBookRepository,
	bookRepo IBookRepository,
) error {
	if req.BookId == "" {
		return status.Errorf(codes.InvalidArgument, "req must have book id")
	}
	if req.CourseId == "" {
		return status.Errorf(codes.InvalidArgument, "req must have course id")
	}

	if req.StudentId == "" {
		return status.Errorf(codes.InvalidArgument, "req must have student id")
	}

	if _, err := courseBookRepo.FindByCourseIDAndBookID(ctx, db, database.Text(req.BookId), database.Text(req.CourseId)); err != nil {
		if err.Error() == pgx.ErrNoRows.Error() {
			return status.Errorf(codes.NotFound, "course id %s not found or book id %s not found", req.CourseId, req.BookId)
		}
		return status.Errorf(codes.Internal, fmt.Errorf("unable to retrieve course book by course id and book id: %w", err).Error())
	}

	book, err := bookRepo.FindByID(ctx, db, database.Text(req.BookId))
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Errorf("s.BookRepo.FindByID: %w", err).Error())
	}

	if book.BookType.String != cpb.BookType_BOOK_TYPE_ADHOC.String() {
		return status.Errorf(codes.Internal, "book must have type adhoc")
	}
	return nil
}

func (s *InternalModifierService) UpsertAdHocIndividualStudyPlan(ctx context.Context, req *epb.UpsertAdHocIndividualStudyPlanRequest) (*epb.UpsertAdHocIndividualStudyPlanResponse, error) {
	return UpsertAdHocIndividualStudyPlan(ctx, s.DB, req, s)
}

func UpsertAdHocIndividualStudyPlan(
	ctx context.Context,
	db database.Ext,
	req *epb.UpsertAdHocIndividualStudyPlanRequest,
	s *InternalModifierService,
) (*epb.UpsertAdHocIndividualStudyPlanResponse, error) {
	var studyPlanID string

	if err := verifyUpsertAdHocIndividualStudyPlan(ctx, db, req, s.CourseBookRepo, s.BookRepo); err != nil {
		return nil, err
	}

	studyPlan, err := convertUpsertAdHocIndividualStudyPlanRequestToStudyPlanEntity(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	if req.StudyPlanId != nil {
		studyPlanID = req.StudyPlanId.Value
		if _, err := s.StudyPlanRepo.FindByID(ctx, db, database.Text(studyPlanID)); err != nil {
			if err.Error() == pgx.ErrNoRows.Error() {
				return nil, status.Errorf(codes.NotFound, "study plan id %v does not exists", studyPlanID)
			}
			return nil, status.Errorf(codes.Internal, fmt.Errorf("studyPlanRepo.FindByID: %w", err).Error())
		}
		if err := studyPlan.ID.Set(studyPlanID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "unable to set study plan id")
		}
		if err := s.StudyPlanRepo.BulkUpsert(ctx, db, []*entities.StudyPlan{studyPlan}); err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Errorf("studyPlanRepo.BulkUpsert: %w", err).Error())
		}
		return &epb.UpsertAdHocIndividualStudyPlanResponse{
			StudyPlanId: studyPlanID,
		}, nil
	} else {
		studyPlanID = studyPlan.ID.String
	}

	if err := database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
		studyPlans := []*entities.StudyPlan{studyPlan}
		if err := s.StudyPlanRepo.BulkUpsert(ctx, tx, studyPlans); err != nil {
			return fmt.Errorf("studyPlanRepo.BulkUpsert: %w", err)
		}

		ssp, err := toStudentStudyPlanEn(database.Text(req.StudentId), studyPlanID)
		if err != nil {
			return fmt.Errorf("toStudentStudyPlanEn: %w", err)
		}
		if err := s.StudentStudyPlanRepo.BulkUpsert(ctx, tx, []*entities.StudentStudyPlan{ssp}); err != nil {
			return fmt.Errorf("studentStudyPlan.BulkUpsert: %w", err)
		}

		if err := UpsertStudyPlanItems(ctx, tx, req.BookId, req.CourseId, studyPlans,
			&InternalModifierService{
				BookRepo:                    s.BookRepo,
				AssignmentRepo:              s.AssignmentRepo,
				StudyPlanItemRepo:           s.StudyPlanItemRepo,
				AssignmentStudyPlanItemRepo: s.AssignmentStudyPlanItemRepo,
				LoStudyPlanItemRepo:         s.LoStudyPlanItemRepo,
				LearningObjectiveRepo:       s.LearningObjectiveRepo,
			}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &epb.UpsertAdHocIndividualStudyPlanResponse{
		StudyPlanId: studyPlanID,
	}, nil
}

func UpsertStudyPlanItems(
	ctx context.Context,
	db database.QueryExecer,
	bookID, courseID string,
	studyPlans []*entities.StudyPlan,
	s *InternalModifierService,
) error {
	booktreeInfos, err := s.BookRepo.RetrieveBookTreeByBookID(ctx, db, database.Text(bookID))
	if err != nil {
		return err
	}

	topicChapterMap := make(map[string]string)
	topicIDs := make([]string, 0, len(booktreeInfos))
	var (
		studyPlanItems        []*entities.StudyPlanItem
		lospItems             []*entities.LoStudyPlanItem
		aspItems              []*entities.AssignmentStudyPlanItem
		studyPlanDisplayOrder int
	)
	now := time.Now()
	for _, info := range booktreeInfos {
		topicID := info.TopicID.String
		if _, ok := topicChapterMap[topicID]; !ok {
			topicChapterMap[topicID] = info.ChapterID.String
			topicIDs = append(topicIDs, topicID)
		}
		// if info.LoID.Status == pgtype.Present {
	}

	learningObjectives, err := s.LearningObjectiveRepo.RetrieveLearningObjectivesByTopicIDs(ctx, db, database.TextArray(topicIDs))
	if err != nil {
		if err.Error() != pgx.ErrNoRows.Error() {
			return fmt.Errorf("learningObjectiveRepo.RetrieveLearningObjectivesByTopicIDs: %w", err)
		}
		learningObjectives = []*entities.LearningObjective{}
	}

	for _, learningObjective := range learningObjectives {
		id := idutil.ULIDNow()

		for _, studyPlan := range studyPlans {
			studyPlanDisplayOrder++
			var e entities.StudyPlanItem
			database.AllNullEntity(&e)
			if err := multierr.Combine(
				e.ID.Set(idutil.ULIDNow()),
				e.StudyPlanID.Set(studyPlan.ID.String),
				e.ContentStructure.Set(&entities.ContentStructure{
					CourseID:  courseID,
					BookID:    bookID,
					ChapterID: topicChapterMap[learningObjective.TopicID.String],
					TopicID:   learningObjective.TopicID.String,
					LoID:      learningObjective.ID.String,
				}),
				e.DisplayOrder.Set(studyPlanDisplayOrder),
				e.CreatedAt.Set(now),
				e.UpdatedAt.Set(now),
				e.Status.Set(epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE.String()),
			); err != nil {
				return fmt.Errorf("unable to set value for study plan item: %w", err)
			}

			if studyPlan.MasterStudyPlan.Status != pgtype.Present {
				if err := e.ID.Set(id); err != nil {
					return fmt.Errorf("unable to set id for study plan item: %w", err)
				}
			} else {
				if err := e.CopyStudyPlanItemID.Set(id); err != nil {
					return fmt.Errorf("unable to set copy study plan item id: %w", err)
				}
			}
			studyPlanItems = append(studyPlanItems, &e)

			var lospItem entities.LoStudyPlanItem
			database.AllNullEntity(&lospItem)
			if err := multierr.Combine(
				lospItem.LoID.Set(learningObjective.ID.String),
				lospItem.StudyPlanItemID.Set(e.ID.String),
				lospItem.CreatedAt.Set(now),
				lospItem.UpdatedAt.Set(now),
			); err != nil {
				return fmt.Errorf("unable to set value for lo study plan item: %w", err)
			}
			lospItems = append(lospItems, &lospItem)
		}
	}

	assignments, err := s.AssignmentRepo.RetrieveAssignmentsByTopicIDs(ctx, db, database.TextArray(topicIDs))
	if err != nil {
		if err.Error() != pgx.ErrNoRows.Error() {
			return fmt.Errorf("assignmentRepo.RetrieveAssignmentsByTopicIDs: %w", err)
		}
		assignments = []*entities.Assignment{}
	}

	for _, assignment := range assignments {
		id := idutil.ULIDNow()
		var content entities.AssignmentContent
		if err := json.Unmarshal(assignment.Content.Bytes, &content); err != nil {
			return fmt.Errorf("unable to unmarshal content: %w", err)
		}
		for _, studyPlan := range studyPlans {
			studyPlanDisplayOrder++
			var e entities.StudyPlanItem
			database.AllNullEntity(&e)
			if err := multierr.Combine(
				e.ID.Set(idutil.ULIDNow()),
				e.StudyPlanID.Set(studyPlan.ID.String),
				e.ContentStructure.Set(&entities.ContentStructure{
					CourseID:     courseID,
					BookID:       bookID,
					ChapterID:    topicChapterMap[content.TopicID],
					TopicID:      content.TopicID,
					AssignmentID: assignment.ID.String,
				}),
				e.DisplayOrder.Set(studyPlanDisplayOrder),
				e.CreatedAt.Set(now),
				e.UpdatedAt.Set(now),
				e.Status.Set(epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE.String()),
			); err != nil {
				return fmt.Errorf("unable to set value for study plan item: %w", err)
			}
			if studyPlan.MasterStudyPlan.Status != pgtype.Present {
				if err := e.ID.Set(id); err != nil {
					return fmt.Errorf("unable to set id for study plan item: %w", err)
				}
			} else {
				if err := e.CopyStudyPlanItemID.Set(id); err != nil {
					return fmt.Errorf("unable to set copy study plan item id: %w", err)
				}
			}
			studyPlanItems = append(studyPlanItems, &e)

			var aspItem entities.AssignmentStudyPlanItem
			database.AllNullEntity(&aspItem)
			if err := multierr.Combine(
				aspItem.AssignmentID.Set(assignment.ID.String),
				aspItem.StudyPlanItemID.Set(e.ID.String),
				aspItem.CreatedAt.Set(now),
				aspItem.UpdatedAt.Set(now),
			); err != nil {
				return fmt.Errorf("unable to set value for lo study plan item: %w", err)
			}
			aspItems = append(aspItems, &aspItem)
		}
	}
	if len(studyPlanItems) > 0 {
		if err := s.StudyPlanItemRepo.BulkInsert(ctx, db, studyPlanItems); err != nil {
			return fmt.Errorf("studyPlanItemRepo.BulkInsert: %w", err)
		}
	}

	if len(aspItems) > 0 {
		if err := s.AssignmentStudyPlanItemRepo.BulkInsert(ctx, db, aspItems); err != nil {
			return fmt.Errorf("assignmentStudyPlanItemRepo.BulkInsert: %w", err)
		}
	}

	if len(lospItems) > 0 {
		if err := s.LoStudyPlanItemRepo.BulkInsert(ctx, db, lospItems); err != nil {
			return fmt.Errorf("loStudyPlanItemRepo.BulkInsert: %w", err)
		}
	}
	return nil
}
