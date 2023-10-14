package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ImportService struct {
	DB  database.Ext
	JSM nats.JetStreamManagement
	Env string

	ImportStudyPlanTaskRepo interface {
		Update(ctx context.Context, db database.QueryExecer, e *entities.ImportStudyPlanTask) error
	}

	MasterStudyPlanRepo interface {
		BulkUpdateTime(ctx context.Context, db database.QueryExecer, items []*entities.MasterStudyPlan) error
	}

	IndividualStudyPlanRepo interface {
		BulkUpdateTime(ctx context.Context, db database.QueryExecer, items []*entities.IndividualStudyPlan) error
	}

	StudyPlanRepo interface {
		BulkCopy(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) ([]string, []string, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlan) error
		FindDependStudyPlan(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]string, error)
		FindByID(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text) (*entities.StudyPlan, error)
		FindByIDs(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.TextArray) ([]*entities.StudyPlan, error)
		BulkUpdateBook(ctx context.Context, db database.QueryExecer, spbs []*repositories.StudyPlanBook) error
		RetrieveMasterByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs pgtype.TextArray) ([]*entities.StudyPlan, error)
		RetrieveByBookIDs(ctx context.Context, db database.QueryExecer, bookIDs pgtype.TextArray) ([]*entities.StudyPlan, error)
		RetrieveStudyPlanItemInfo(ctx context.Context, db database.QueryExecer, args repositories.StudyPlanItemInfoArgs) ([]*repositories.StudyPlanItemInfo, error)
	}
	StudentRepo interface {
		FindStudentsByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*pgtype.TextArray, error)
	}
	CourseStudyPlanRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, courseStudyPlans []*entities.CourseStudyPlan) error
		FindByCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs pgtype.TextArray) ([]*entities.CourseStudyPlan, error)
		ListCourseStudyPlans(ctx context.Context, db database.QueryExecer, args *repositories.ListCourseStudyPlansArgs) ([]*entities.CourseStudyPlan, error)
	}
	StudentStudyPlan interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, studentStudyPlans []*entities.StudentStudyPlan) error
	}
	StudyPlanItemRepo interface {
		BulkCopy(ctx context.Context, db database.QueryExecer, originalStudyPlanIDs pgtype.TextArray, newStudyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) error
		BulkSync(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) ([]*entities.StudyPlanItem, error)
		UpdateWithCopiedFromItem(ctx context.Context, db database.QueryExecer, studyPlanItems []*entities.StudyPlanItem) error
		FindStudyPlanIDByItemIDs(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) (map[string]string, error)
		SoftDeleteWithStudyPlanIDs(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		RetrieveStudyPlanContentStructuresByBooks(ctx context.Context, db database.QueryExecer, books pgtype.TextArray) (map[string][]entities.ContentStructure, error)
		CopyItemsForCopiedStudyPlans(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) error
		FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.StudyPlanItem, error)
		FindByStudyPlanID(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.Text) ([]*entities.StudyPlanItem, error)
	}
	AssignmentStudyPlanItemRepo interface {
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error
		CountAssignment(ctx context.Context, db database.QueryExecer, assignmentIDs pgtype.TextArray) (int, error)
		BulkCopy(ctx context.Context, db database.QueryExecer, items []*entities.AssignmentStudyPlanItem) error
		BulkUpsertByStudyPlanItem(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error
	}
	LoStudyPlanItemRepo interface {
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, loStudyPlanItems []*entities.LoStudyPlanItem) error
		BulkCopy(ctx context.Context, db database.QueryExecer, items []*entities.LoStudyPlanItem) error
		BulkUpsertByStudyPlanItem(ctx context.Context, db database.QueryExecer, loStudyPlanItems []*entities.LoStudyPlanItem) error
	}
	AssignStudyPlanTaskRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.AssignStudyPlanTask) (pgtype.Text, error)
	}

	BookChapterRepo interface {
		Upsert(ctx context.Context, db database.Ext, cc []*entities.BookChapter) error
		SoftDelete(ctx context.Context, db database.QueryExecer, chapterIDs, bookIDs pgtype.TextArray) error
		SoftDeleteByChapterIDs(ctx context.Context, db database.QueryExecer, chapterIDs []string) error
		RetrieveContentStructuresByLOs(ctx context.Context, db database.QueryExecer, loIDs pgtype.TextArray) (map[string]entities.ContentStructure, error)
		RetrieveContentStructuresByTopics(ctx context.Context, db database.QueryExecer, topicIDs pgtype.TextArray) (map[string][]entities.ContentStructure, error)
	}

	AssignmentRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, assignments []*entities.Assignment) error
		SoftDelete(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error
		RetrieveAssignments(ctx context.Context, db database.QueryExecer, assignmentIDs pgtype.TextArray) ([]*entities.Assignment, error)
	}

	TopicsAssignmentsRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, topicsAssignments *entities.TopicsAssignments) error
		BulkUpsert(ctx context.Context, db database.QueryExecer, topicsAssignmentsList []*entities.TopicsAssignments) error
		SoftDeleteByAssignmentIDs(ctx context.Context, db database.QueryExecer, assignmentIDs pgtype.TextArray) error
	}
}

func (s *ImportService) ImportStudyPlan(ctx context.Context, req *pb.ImportStudyPlanRequest) (*pb.ImportStudyPlanResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, codes.Unimplemented.String())
}

func copyStudyPlanItem(org *pb.StudyPlanItem) *pb.StudyPlanItem {
	return &pb.StudyPlanItem{
		StudyPlanId:      org.StudyPlanId,
		StudyPlanItemId:  idutil.ULIDNow(),
		AvailableFrom:    org.AvailableFrom,
		AvailableTo:      org.AvailableTo,
		StartDate:        org.StartDate,
		EndDate:          org.EndDate,
		ContentStructure: org.ContentStructure,
		DisplayOrder:     org.DisplayOrder,
		Status:           org.Status,
	}
}

func editTimeStudyPlanItem(item *sspb.StudyPlanItemImport, spiItem *entities.StudyPlanItem) (*entities.StudyPlanItem, error) {
	if err := multierr.Combine(
		spiItem.StartDate.Set(nil),
		spiItem.EndDate.Set(nil),
		spiItem.AvailableFrom.Set(nil),
		spiItem.AvailableTo.Set(nil),
	); err != nil {
		return nil, fmt.Errorf("multiErr.Combine.SetImportStudyPlanTaskEnt: %v", err)
	}

	if item.StartDate != nil {
		_ = spiItem.StartDate.Set(item.StartDate.AsTime())
	}

	if item.EndDate != nil {
		_ = spiItem.EndDate.Set(item.EndDate.AsTime())
	}

	if item.AvailableFrom != nil {
		_ = spiItem.AvailableFrom.Set(item.AvailableFrom.AsTime())
	}

	if item.AvailableTo != nil {
		_ = spiItem.AvailableTo.Set(item.AvailableTo.AsTime())
	}

	return spiItem, nil
}

func (s *ImportService) ImportStudyPlanItems(ctx context.Context, data *npb.EventImportStudyPlan) error {
	// receive task and update status of task -> IN_PROGRESS
	importStudyPlanTaskEnt := &entities.ImportStudyPlanTask{}
	database.AllNullEntity(importStudyPlanTaskEnt)
	if err := multierr.Combine(
		importStudyPlanTaskEnt.TaskID.Set(data.TaskId),
		importStudyPlanTaskEnt.Status.Set(pb.StudyPlanTaskStatus_STUDY_PLAN_TASK_STATUS_IN_PROGRESS.String()),
		importStudyPlanTaskEnt.UpdatedAt.Set(time.Now()),
	); err != nil {
		return fmt.Errorf("multiErr.Combine.SetImportStudyPlanTaskEnt: %v", err)
	}

	if err := s.ImportStudyPlanTaskRepo.Update(ctx, s.DB, importStudyPlanTaskEnt); err != nil {
		return fmt.Errorf("ImportService.ImportStudyPlanItems.UpdateStatus: %v", err)
	}

	// update time for master item
	// update time for individual study plan item
	// update time for study plan item (old flow)
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		/* Using following code if continue to refactoring system

		masterSpis := []*entities.MasterStudyPlan{}
		individualSpis := []*entities.IndividualStudyPlan{}
		for _, item := range data.StudyPlanItems {
			// init master
			masterSpi := &entities.MasterStudyPlan{}
			database.AllNullEntity(masterSpi)
			_ = masterSpi.StudyPlanID.Set(item.StudyPlanId)
			_ = masterSpi.LearningMaterialID.Set(item.LearningMaterialId)

			// init individual
			individualSpi := &entities.IndividualStudyPlan{}
			database.AllNullEntity(individualSpi)
			_ = individualSpi.ID.Set(item.StudyPlanId)
			_ = individualSpi.LearningMaterialID.Set(item.LearningMaterialId)

			if item.StartDate != nil {
				_ = masterSpi.StartDate.Set(item.StartDate.AsTime())
				_ = individualSpi.StartDate.Set(item.StartDate.AsTime())
			}

			if item.EndDate != nil {
				_ = masterSpi.EndDate.Set(item.EndDate.AsTime())
				_ = individualSpi.EndDate.Set(item.EndDate.AsTime())
			}

			if item.AvailableFrom != nil {
				_ = masterSpi.AvailableFrom.Set(item.AvailableFrom.AsTime())
				_ = individualSpi.AvailableFrom.Set(item.AvailableFrom.AsTime())
			}

			if item.AvailableTo != nil {
				_ = masterSpi.AvailableTo.Set(item.AvailableTo.AsTime())
				_ = individualSpi.AvailableTo.Set(item.AvailableTo.AsTime())
			}

			masterSpis = append(masterSpis, masterSpi)
			individualSpis = append(individualSpis, individualSpi)
		}

		if err := s.MasterStudyPlanRepo.BulkUpdateTime(ctx, tx, masterSpis); err != nil {
			return fmt.Errorf("s.MasterStudyPlanRepo.BulkUpdateTime: %w", err)
		}

		if err := s.IndividualStudyPlanRepo.BulkUpdateTime(ctx, tx, individualSpis); err != nil {
			return fmt.Errorf("s.IndividualStudyPlanRepo.BulkUpdateTime: %w", err)
		}

		*/

		// update time for study_plan_items table
		spis, err := s.StudyPlanItemRepo.FindByStudyPlanID(ctx, tx, database.Text(data.StudyPlanItems[0].StudyPlanId))
		if err != nil {
			return fmt.Errorf("s.StudyPlanItemRepo.FindByStudyPlanID: %v", err)
		}

		editedSpis := []*entities.StudyPlanItem{}
		for _, reqItem := range data.GetStudyPlanItems() {
			for _, item := range spis {
				cs := new(entities.ContentStructure)
				if err := item.ContentStructure.AssignTo(&cs); err != nil {
					return fmt.Errorf("item.ContentStructure.AssignTo: %v", err)
				}
				if reqItem.LearningMaterialId == cs.AssignmentID {
					spi, err := editTimeStudyPlanItem(reqItem, item)
					if err != nil {
						return err
					}
					editedSpis = append(editedSpis, spi)
				}

				if reqItem.LearningMaterialId == cs.LoID {
					spi, err := editTimeStudyPlanItem(reqItem, item)
					if err != nil {
						return err
					}

					editedSpis = append(editedSpis, spi)
				}
			}
		}

		if err := s.StudyPlanItemRepo.BulkInsert(ctx, tx, editedSpis); err != nil {
			return fmt.Errorf("s.StudyPlanItemRepo.BulkInsert: %v", err)
		}

		if err := s.StudyPlanItemRepo.UpdateWithCopiedFromItem(ctx, tx, editedSpis); err != nil {
			return fmt.Errorf("s.StudyPlanItemRepo.UpdateWithCopiedFromItem: %v", err)
		}

		return nil
	}); err != nil {
		// update status of task and write err log
		if err := multierr.Combine(
			importStudyPlanTaskEnt.Status.Set(pb.StudyPlanTaskStatus_STUDY_PLAN_TASK_STATUS_ERROR.String()),
			importStudyPlanTaskEnt.ErrorDetail.Set(err.Error()),
			importStudyPlanTaskEnt.UpdatedAt.Set(time.Now()),
		); err != nil {
			return fmt.Errorf("multiErr.Combine.SetImportStudyPlanTaskEnt: %v", err)
		}

		if err := s.ImportStudyPlanTaskRepo.Update(ctx, s.DB, importStudyPlanTaskEnt); err != nil {
			return fmt.Errorf("ImportService.ImportStudyPlanItems.WriteError: %v", err)
		}

		return fmt.Errorf("database.ExecInTx: %w", err)
	}

	// update task completed
	if err := multierr.Combine(
		importStudyPlanTaskEnt.Status.Set(pb.StudyPlanTaskStatus_STUDY_PLAN_TASK_STATUS_COMPLETED.String()),
		importStudyPlanTaskEnt.UpdatedAt.Set(time.Now()),
	); err != nil {
		return fmt.Errorf("multiErr.Combine.SetImportStudyPlanTaskEnt: %v", err)
	}

	if err := s.ImportStudyPlanTaskRepo.Update(ctx, s.DB, importStudyPlanTaskEnt); err != nil {
		return fmt.Errorf("ImportService.ImportStudyPlanItems.UpdateStatus: %v", err)
	}

	return nil
}

func (s *ImportService) SyncStudyPlanItemsOnLOsCreated(ctx context.Context, data *npb.EventLearningObjectivesCreated) error {
	mapLoContentStructure := map[string]*pb.ContentStructure{}
	if data.LoContentStructures == nil {
		var ids = make([]string, 0)
		for _, lo := range data.LearningObjectives {
			ids = append(ids, lo.Info.Id)
		}

		contentStructures, err := s.BookChapterRepo.RetrieveContentStructuresByLOs(
			ctx,
			s.DB,
			database.TextArray(ids),
		)

		if err != nil {
			return status.Errorf(codes.Internal, "cm.BookChapterRepo.RetrieveContentStructuresByLOs: %v", err)
		}

		for loID, contentStructures := range contentStructures {
			mapLoContentStructure[loID] = toContentStructuresPb(contentStructures)
		}
	} else {
		for loID, lcs := range data.LoContentStructures {
			for _, contentStructure := range lcs.ContentStructures {
				mapLoContentStructure[loID] = &pb.ContentStructure{
					CourseId:  contentStructure.CourseId,
					BookId:    contentStructure.BookId,
					ChapterId: contentStructure.ChapterId,
					TopicId:   contentStructure.TopicId,
				}
			}
		}
	}
	defer timeutil.CalculateRunTime(ctx, "SyncStudyPlanItemsOnLOsCreated")()

	loDisplayOrder := make(map[string]int) // lo id => display order
	loIDs := make([]string, 0, len(data.LearningObjectives))
	for _, lo := range data.LearningObjectives {
		loDisplayOrder[lo.Info.Id] = int(lo.Info.DisplayOrder)
		loIDs = append(loIDs, lo.Info.Id)
	}

	loIDs = golibs.GetUniqueElementStringArray(loIDs)
	var bookIDs []string
	for _, cs := range mapLoContentStructure {
		bookIDs = append(bookIDs, cs.BookId)
	}

	studyPlanLoMap := make(map[string]string)
	studyPlanBookMap := make(map[string]*repositories.StudyPlanItemInfo)
	bookStudyPlanItemInfoMap := make(map[string][]*repositories.StudyPlanItemInfo)
	studyPlanMasterStudyPlanItemMap := make(map[string]string)
	getKey := func(keys ...string) string {
		return strings.Join(keys, "|")
	}
	infos, err := s.StudyPlanRepo.RetrieveStudyPlanItemInfo(ctx, s.DB, repositories.StudyPlanItemInfoArgs{
		BookIDs:       database.TextArray(bookIDs),
		LoIDs:         database.TextArray(loIDs),
		AssignmentIDs: database.TextArray(nil),
	})
	if err != nil {
		return fmt.Errorf("StudyPlanRepo.RetrieveStudyPlanItemInfo: %v", err)
	}
	for _, info := range infos {
		bookID := info.BookID.String
		studyPlanID := info.StudyPlanID.String
		if info.StudyPlanItem.ContentStructure.Status == pgtype.Present {
			var ct entities.ContentStructure
			info.StudyPlanItem.ContentStructure.AssignTo(&ct)
			studyPlanLoMap[getKey(studyPlanID, ct.LoID)] = info.StudyPlanItem.ID.String
		}
		if _, ok := studyPlanBookMap[getKey(studyPlanID, bookID)]; !ok {
			studyPlanBookMap[getKey(studyPlanID, bookID)] = info
			bookStudyPlanItemInfoMap[bookID] = append(bookStudyPlanItemInfoMap[bookID], info)
		}
	}

	var (
		studyPlanItems   []*entities.StudyPlanItem
		loStudyPlanItems []*entities.LoStudyPlanItem
	)
	loMap := make(map[string]bool)
	for _, lo := range data.LearningObjectives {
		loID := lo.Info.Id
		if _, ok := mapLoContentStructure[loID]; !ok {
			continue
		}
		cs := mapLoContentStructure[loID]
		infos := bookStudyPlanItemInfoMap[cs.BookId]
		for _, info := range infos {
			if cs.BookId != info.BookID.String {
				continue
			}
			cs.CourseId = info.CourseID.String
			item := newItem(info.StudyPlanID.String, cs)
			item.DisplayOrder.Set(loDisplayOrder[loID])
			item.ContentStructureFlatten.Set(toContentStructureFlattenLO(cs, loID))
			if studyPlanItemID, ok := studyPlanLoMap[getKey(info.StudyPlanID.String, loID)]; ok {
				item.ID.Set(studyPlanItemID)
			} else if ok := loMap[loID]; !ok {
				loItem := newLOItem(loID, item.ID.String)
				loStudyPlanItems = append(loStudyPlanItems, loItem)
			}
			if info.MasterStudyPlanID.Status == pgtype.Null {
				studyPlanMasterStudyPlanItemMap[getKey(info.StudyPlanID.String, loID)] = item.ID.String
			}
			if info.MasterStudyPlanID.Status == pgtype.Present {
				item.CopyStudyPlanItemID.Set(studyPlanMasterStudyPlanItemMap[getKey(info.MasterStudyPlanID.String, loID)])
			}
			studyPlanItems = append(studyPlanItems, item)
		}
		loMap[loID] = true
	}
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := s.StudyPlanItemRepo.BulkSync(ctx, tx, studyPlanItems); err != nil {
			return fmt.Errorf("s.StudyPlanItemRepo.BulkSync: %v", err)
		}

		if len(loStudyPlanItems) > 0 {
			if err := s.LoStudyPlanItemRepo.BulkUpsertByStudyPlanItem(ctx, tx, loStudyPlanItems); err != nil {
				return fmt.Errorf("s.LoStudyPlanItemRepo.BulkUpsertByStudyPlanItem: %v", err)
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("database.ExecInTx: %v", err)
	}
	return nil
}

func newItem(studyPlanID string, cs *pb.ContentStructure) *entities.StudyPlanItem {
	now := time.Now()

	item := &entities.StudyPlanItem{}
	database.AllNullEntity(item)

	item.ID.Set(idutil.ULIDNow())
	item.StudyPlanID.Set(studyPlanID)
	item.ContentStructure.Set(toContentStructure(cs))
	item.ContentStructureFlatten.Set(nil)
	item.DisplayOrder.Set(0)
	item.Status.Set(pb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE.String())
	item.CreatedAt.Set(now)
	item.UpdatedAt.Set(now)

	return item
}

func newLOItem(loID, itemID string) *entities.LoStudyPlanItem {
	now := time.Now()

	loItem := &entities.LoStudyPlanItem{}
	loItem.LoID.Set(loID)
	loItem.StudyPlanItemID.Set(itemID)
	loItem.CreatedAt.Set(now)
	loItem.UpdatedAt.Set(now)
	loItem.DeletedAt.Set(nil)

	return loItem
}

func (s *ImportService) SyncStudyPlanItemsOnAssignmentsCreated(ctx context.Context, data *npb.EventAssignmentsCreated) error {
	defer timeutil.CalculateRunTime(ctx, "SyncStudyPlanItemsOnAssignmentsCreated")()
	assignmentAndTimesMap := make(map[*pb.Assignment]*StudyPlanItemTimes)
	for _, assignment := range data.Assignments {
		assignmentAndTimesMap[assignment] = nil
	}

	return s.HandedStudyPlanItemsWithTimesOnAssignmentsCreated(ctx, assignmentAndTimesMap)
}

type StudyPlanItemTimes struct {
	AvailableFrom *time.Time
	AvailableTo   *time.Time
	StartDate     *time.Time
	EndDate       *time.Time
}

func (s *ImportService) HandedStudyPlanItemsWithTimesOnAssignmentsCreated(ctx context.Context, assignmentAndTimesMap map[*pb.Assignment]*StudyPlanItemTimes) error {
	return HandedStudyPlanItemsWithTimesOnAssignmentsCreated(ctx, s.DB, assignmentAndTimesMap, s)
}

// nolint
func HandedStudyPlanItemsWithTimesOnAssignmentsCreated(
	ctx context.Context,
	db database.Ext,
	assignmentAndTimesMap map[*pb.Assignment]*StudyPlanItemTimes,
	s *ImportService,
) error {
	assignments := make([]*pb.Assignment, 0, len(assignmentAndTimesMap))
	for assignment := range assignmentAndTimesMap {
		assignments = append(assignments, assignment)
	}

	topicAssignments := make(map[string][]string)
	topicIDs := make([]string, 0, len(assignments))
	mDisplayOrder := make(map[string]int32)
	assignmentIDs := make([]string, 0, len(assignments))
	for _, a := range assignments {
		topicID := a.GetContent().GetTopicId()
		topicIDs = append(topicIDs, topicID)
		assignmentIDs = append(assignmentIDs, a.AssignmentId)
		topicAssignments[topicID] = append(topicAssignments[topicID], a.AssignmentId)

		mDisplayOrder[a.AssignmentId] = a.DisplayOrder
	}

	contentStructures, err := s.BookChapterRepo.RetrieveContentStructuresByTopics(ctx, db, database.TextArray(golibs.Uniq(topicIDs)))
	if err != nil {
		return status.Errorf(codes.Internal, "bookChapterRepo.RetrieveContentStructuresByTopics: %v", err)
	}

	resp := make(map[string]*bpb.ContentStructures)
	for topicID, contentStructure := range contentStructures {
		resp[topicID] = &bpb.ContentStructures{
			ContentStructures: toContentStructuresPbV2(contentStructure),
		}
	}

	var bookIDs []string
	for _, lcs := range resp {
		for _, cs := range lcs.GetContentStructures() {
			bookIDs = append(bookIDs, cs.BookId)
		}
	}
	studyPlanAssignmentMap := make(map[string]string)
	studyPlanBookMap := make(map[string]*repositories.StudyPlanItemInfo)
	bookStudyPlanItemInfoMap := make(map[string][]*repositories.StudyPlanItemInfo)
	studyPlanMasterStudyPlanItemMap := make(map[string]string)
	getKey := func(keys ...string) string {
		return strings.Join(keys, "|")
	}
	infos, err := s.StudyPlanRepo.RetrieveStudyPlanItemInfo(ctx, db, repositories.StudyPlanItemInfoArgs{
		BookIDs:       database.TextArray(bookIDs),
		AssignmentIDs: database.TextArray(assignmentIDs),
		LoIDs:         database.TextArray(nil),
	})
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Errorf("studyPlanRepo.RetrieveStudyPlanItemInfo: %v", err).Error())
	}

	for _, info := range infos {
		bookID := info.BookID.String
		studyPlanID := info.StudyPlanID.String

		if info.StudyPlanItem.ContentStructure.Status == pgtype.Present {
			var ct entities.ContentStructure
			info.StudyPlanItem.ContentStructure.AssignTo(&ct)
			studyPlanAssignmentMap[getKey(studyPlanID, ct.AssignmentID)] = info.StudyPlanItem.ID.String
		}
		if _, ok := studyPlanBookMap[getKey(studyPlanID, bookID)]; !ok {
			studyPlanBookMap[getKey(studyPlanID, bookID)] = info
			bookStudyPlanItemInfoMap[bookID] = append(bookStudyPlanItemInfoMap[bookID], info)
		}
	}

	var (
		studyPlanItems           []*entities.StudyPlanItem
		assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem
	)
	for _, a := range assignments {
		assignmentID := a.AssignmentId
		topicID := a.GetContent().GetTopicId()
		if _, ok := resp[topicID]; !ok {
			continue
		}
		loContentStructures := resp[topicID].GetContentStructures()
		for _, cs := range loContentStructures {
			infos := bookStudyPlanItemInfoMap[cs.BookId]
			for _, info := range infos {
				if cs.BookId != info.BookID.String {
					continue
				}
				cs.CourseId = info.CourseID.String
				item := newItem(info.StudyPlanID.String, cs)
				item.DisplayOrder.Set(mDisplayOrder[assignmentID])
				item.ContentStructureFlatten.Set(toContentStructureFlattenAssignment(cs, assignmentID))
				times := assignmentAndTimesMap[a]
				if times != nil {
					if times.AvailableFrom != nil {
						item.AvailableFrom.Set(*times.AvailableFrom)
					}
					if times.AvailableTo != nil {
						item.AvailableTo.Set(*times.AvailableTo)
					}
					if times.StartDate != nil {
						item.StartDate.Set(*times.StartDate)
					}
					if times.EndDate != nil {
						item.EndDate.Set(*times.EndDate)
					}
				}
				if studyPlanItemID, ok := studyPlanAssignmentMap[getKey(info.StudyPlanID.String, assignmentID)]; ok {
					item.ID.Set(studyPlanItemID)
				} else {
					aItem := newAssignmentItem(assignmentID, item.ID.String)
					assignmentStudyPlanItems = append(assignmentStudyPlanItems, aItem)
				}
				if info.MasterStudyPlanID.Status == pgtype.Null {
					studyPlanMasterStudyPlanItemMap[getKey(info.StudyPlanID.String, assignmentID)] = item.ID.String
				}
				if info.MasterStudyPlanID.Status == pgtype.Present {
					item.CopyStudyPlanItemID.Set(studyPlanMasterStudyPlanItemMap[getKey(info.MasterStudyPlanID.String, assignmentID)])
				}
				studyPlanItems = append(studyPlanItems, item)
			}
		}
	}
	if err := database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {

		if len(studyPlanItems) > 0 {
			if _, err := s.StudyPlanItemRepo.BulkSync(ctx, tx, studyPlanItems); err != nil {
				return fmt.Errorf("studyPlanItemRepo.BulkSync: %v", err)
			}
		}
		if len(assignmentStudyPlanItems) > 0 {
			if err := s.AssignmentStudyPlanItemRepo.BulkUpsertByStudyPlanItem(ctx, tx, assignmentStudyPlanItems); err != nil {
				return fmt.Errorf("assignmentStudyPlanItemRepo.BulkUpsertByStudyPlanItem: %v", err)
			}
		}
		return nil
	}); err != nil {
		return status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %v", err).Error())
	}

	return nil
}

func newAssignmentItem(assignmentID, itemID string) *entities.AssignmentStudyPlanItem {
	now := time.Now()

	e := &entities.AssignmentStudyPlanItem{}
	e.AssignmentID.Set(assignmentID)
	e.StudyPlanItemID.Set(itemID)
	e.CreatedAt.Set(now)
	e.UpdatedAt.Set(now)
	e.DeletedAt.Set(nil)

	return e
}

func toContentStructuresPb(cs entities.ContentStructure) *pb.ContentStructure {
	ret := &pb.ContentStructure{
		ChapterId: cs.ChapterID,
		BookId:    cs.BookID,
		TopicId:   cs.TopicID,
	}
	return ret
}

func toContentStructuresPbV2(contentStructures []entities.ContentStructure) []*pb.ContentStructure {
	ret := make([]*pb.ContentStructure, 0, len(contentStructures))
	for _, cs := range contentStructures {
		ret = append(ret, &pb.ContentStructure{
			ChapterId: cs.ChapterID,
			BookId:    cs.BookID,
			TopicId:   cs.TopicID,
		})
	}
	return ret
}
