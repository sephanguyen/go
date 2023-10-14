package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type AssignStudyPlanTaskModifierService struct {
	DB            database.Ext
	StudyPlanRepo interface {
		FindByIDs(ctx context.Context, db database.QueryExecer, studyPlanID pgtype.TextArray) ([]*entities.StudyPlan, error)
		BulkCopy(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) ([]string, []string, error)
		BulkUpsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlan) error
		FindDependStudyPlan(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]string, error)
		BulkUpdateBook(ctx context.Context, db database.QueryExecer, spbs []*repositories.StudyPlanBook) error
	}
	StudentRepo interface {
		FindStudentsByCourseID(ctx context.Context, db database.QueryExecer, courseID pgtype.Text) (*pgtype.TextArray, error)
	}
	CourseStudyPlanRepo interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, courseStudyPlans []*entities.CourseStudyPlan) error
	}
	StudentStudyPlan interface {
		BulkUpsert(ctx context.Context, db database.QueryExecer, studentStudyPlans []*entities.StudentStudyPlan) error
	}
	StudyPlanItemRepo interface {
		BulkCopy(ctx context.Context, db database.QueryExecer, originalStudyPlanIDs pgtype.TextArray, newStudyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) error
		UpdateWithCopiedFromItem(ctx context.Context, db database.QueryExecer, studyPlanItems []*entities.StudyPlanItem) error
		FindByIDs(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) ([]*entities.StudyPlanItem, error)
		FindStudyPlanIDByItemIDs(ctx context.Context, db database.QueryExecer, studyPlanItemIDs pgtype.TextArray) (map[string]string, error)
		SoftDeleteWithStudyPlanIDs(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
	}
	AssignmentStudyPlanItemRepo interface {
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error
	}
	LoStudyPlanItemRepo interface {
		CopyFromStudyPlan(ctx context.Context, db database.QueryExecer, studyPlanIDs pgtype.TextArray) error
		BulkInsert(ctx context.Context, db database.QueryExecer, loStudyPlanItems []*entities.LoStudyPlanItem) error
	}
	AssignStudyPlanTaskRepo interface {
		Create(ctx context.Context, db database.QueryExecer, e *entities.AssignStudyPlanTask) (pgtype.Text, error)
		UpdateStatus(ctx context.Context, db database.QueryExecer, id pgtype.Text, status pgtype.Text) error
		UpdateDetailError(ctx context.Context, db database.QueryExecer, errDetail *repositories.AssignStudyPlanTaskDetailErrorArgs) error
	}
}

func (s *AssignStudyPlanTaskModifierService) assignStudyPlan(ctx context.Context, studyPlanIDs []string, courseID string, studentIDs []string, studyPlanType pb.StudyPlanType, scheduleStudyPlanItems []*pb.ScheduleStudyPlan, tx pgx.Tx, r *IAssignStudyPlan) error {
	err := ScheduleStudyPlanWithTx(ctx, tx, &pb.ScheduleStudyPlanRequest{
		Schedule: scheduleStudyPlanItems,
	}, s.AssignmentStudyPlanItemRepo.BulkInsert, s.LoStudyPlanItemRepo.BulkInsert)
	if err != nil {
		return fmt.Errorf("s.ScheduleStudyPlanWithTx: %w", err)
	}

	if studyPlanType == pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE {
		err := HandleAssignCourseStudyPlan(ctx, courseID, studyPlanIDs[0], tx, r)
		if err != nil {
			return fmt.Errorf("HandleAssignCourseStudyPlan: %w", err)
		}
	}
	if studyPlanType == pb.StudyPlanType_STUDY_PLAN_TYPE_INDIVIDUAL {
		for i, studyPlanID := range studyPlanIDs {
			assignStudyPlanReq := &pb.AssignStudyPlanRequest{
				StudyPlanId: studyPlanID,
				Data: &pb.AssignStudyPlanRequest_StudentId{
					StudentId: studentIDs[i],
				},
			}
			err = AssignStudyPlanWithTx(ctx, assignStudyPlanReq, tx, r)
			if err != nil {
				return fmt.Errorf("AssignStudyPlanWithTx: %w", err)
			}
		}
	}
	return nil
}

func makeCreateCourseStudyPlanData(studyPlanID string, studyPlanItems []*pb.StudyPlanItem,
	scheduleStudyPlanItems []*pb.ScheduleStudyPlan) ([]*pb.StudyPlanItem, []*pb.ScheduleStudyPlan) {
	for i, item := range studyPlanItems {
		item.StudyPlanId = studyPlanID
		studyPlanItemID := idutil.ULIDNow()
		item.StudyPlanItemId = studyPlanItemID
		scheduleStudyPlanItems[i].StudyPlanItemId = studyPlanItemID
	}

	return studyPlanItems, scheduleStudyPlanItems
}

func makeCreateIndividualStudyPlanData(studyPlanIDs []string,
	studyPlanItems []*pb.StudyPlanItem, scheduleStudyPlanItems []*pb.ScheduleStudyPlan) (
	[]*pb.StudyPlanItem, []*pb.ScheduleStudyPlan) {
	var newScheduleStudyPlanItems []*pb.ScheduleStudyPlan
	var newStudyPlanItems []*pb.StudyPlanItem
	for _, studyPlanID := range studyPlanIDs {
		for i, item := range studyPlanItems {
			newStudyPlanItem := copyStudyPlanItem(item)
			newStudyPlanItem.StudyPlanId = studyPlanID
			scheduleStudyPlanItem := &pb.ScheduleStudyPlan{
				StudyPlanItemId: newStudyPlanItem.StudyPlanItemId,
				Item:            scheduleStudyPlanItems[i].Item,
			}
			newScheduleStudyPlanItems = append(newScheduleStudyPlanItems, scheduleStudyPlanItem)
			newStudyPlanItems = append(newStudyPlanItems, newStudyPlanItem)
		}
	}
	return newStudyPlanItems, newScheduleStudyPlanItems
}

func (s *AssignStudyPlanTaskModifierService) createStudyPlan(ctx context.Context, req *pb.AssignStudyPlanEvent) error {
	r := &IAssignStudyPlan{
		StudyPlanRepo:               s.StudyPlanRepo,
		CourseStudyPlanRepo:         s.CourseStudyPlanRepo,
		StudentRepo:                 s.StudentRepo,
		StudentStudyPlan:            s.StudentStudyPlan,
		StudyPlanItemRepo:           s.StudyPlanItemRepo,
		AssignmentStudyPlanItemRepo: s.AssignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         s.LoStudyPlanItemRepo,
	}

	var newStudyPlanItems []*pb.StudyPlanItem
	var newScheduleStudyPlanItem []*pb.ScheduleStudyPlan
	studyPlanIDs := req.StudyPlanIds

	err := database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if req.Type == pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE {
			newStudyPlanItems, newScheduleStudyPlanItem = makeCreateCourseStudyPlanData(req.StudyPlanIds[0], req.StudyPlanItems, req.ScheduleStudyPlanItems)
		} else {
			newStudyPlanItems, newScheduleStudyPlanItem = makeCreateIndividualStudyPlanData(req.StudyPlanIds, req.StudyPlanItems, req.ScheduleStudyPlanItems)
		}
		_, errTx := UpsertStudyPlanItemWithTx(ctx, &pb.UpsertStudyPlanItemRequest{
			StudyPlanItems: newStudyPlanItems,
		}, tx, r)
		if errTx != nil {
			return errTx
		}

		errTx = s.assignStudyPlan(ctx, studyPlanIDs, req.CourseId, req.StudentIds, req.Type, newScheduleStudyPlanItem, tx, r)
		if errTx != nil {
			return fmt.Errorf("s.assignStudyPlan: %v", errTx)
		}

		completeStatus := database.Text(pb.StudyPlanTaskStatus_STUDY_PLAN_TASK_STATUS_COMPLETED.String())
		errTx = s.AssignStudyPlanTaskRepo.UpdateStatus(ctx, tx, database.Text(req.TaskId), completeStatus)
		if errTx != nil {
			return fmt.Errorf("s.AssignStudyPlanTaskRepo.UpdateStatus: %v", errTx)
		}
		return errTx
	})
	if err != nil {
		// failed to upsert then update status error and detail error to study plan task
		errorStatus := database.Text(pb.StudyPlanTaskStatus_STUDY_PLAN_TASK_STATUS_ERROR.String())
		errDetail := &repositories.AssignStudyPlanTaskDetailErrorArgs{
			ID:          database.Text(req.TaskId),
			Status:      errorStatus,
			ErrorDetail: database.Text(err.Error()),
		}
		// if when update fail again we have to combine two error and return
		if err2 := s.AssignStudyPlanTaskRepo.UpdateDetailError(ctx, s.DB, errDetail); err2 != nil {
			return fmt.Errorf(multierr.Combine(err, fmt.Errorf("AssignStudyPlanTaskRepo.UpdateDetailError: %w", err2)).Error())
		}
		return err
	}
	return nil
}

func (s *AssignStudyPlanTaskModifierService) makeCourseStudyPlanUpdateData(
	ctx context.Context, courseStudyPlanID string,
	studyPlanItems []*pb.StudyPlanItem, scheduleStudyPlanItems []*pb.ScheduleStudyPlan,
) ([]*pb.StudyPlanItem, []*pb.StudyPlanItem, []*pb.ScheduleStudyPlan, error) {
	newStudyPlanItems := make([]*pb.StudyPlanItem, 0, len(studyPlanItems))
	updateStudyPlanItems := make([]*pb.StudyPlanItem, 0, len(studyPlanItems))
	newScheduleStudyPlanItems := make([]*pb.ScheduleStudyPlan, 0, len(studyPlanItems))

	var studyPlanItemIDs []string
	for _, studyPlanItem := range studyPlanItems {
		if studyPlanItem.StudyPlanItemId != "" {
			studyPlanItemIDs = append(studyPlanItemIDs, studyPlanItem.StudyPlanItemId)
		}
	}

	oldStudyPlanItems, err := s.StudyPlanItemRepo.FindByIDs(ctx, s.DB, database.TextArray(studyPlanItemIDs))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("s.StudyPlanItemRepo.FindByIDs: %w", err)
	}

	oldStudyPlanItemMap := make(map[string]*entities.StudyPlanItem)
	for _, oldStudyPlanItem := range oldStudyPlanItems {
		oldStudyPlanItemMap[oldStudyPlanItem.ID.String] = oldStudyPlanItem
	}

	for i, studyPlanItem := range studyPlanItems {
		studyPlanItem.StudyPlanId = courseStudyPlanID
		if studyPlanItem.StudyPlanItemId == "" {
			studyPlanItem.StudyPlanItemId = idutil.ULIDNow()
			newStudyPlanItems = append(newStudyPlanItems, studyPlanItem)
			newScheduleStudyPlanItems = append(newScheduleStudyPlanItems, scheduleStudyPlanItems[i])
		} else {
			_studyPlanItem, ok := oldStudyPlanItemMap[studyPlanItem.StudyPlanItemId]
			if !ok {
				continue
			}
			completedAt := _studyPlanItem.CompletedAt.Time
			if !completedAt.IsZero() {
				studyPlanItem.CompletedAt = timestamppb.New(completedAt)
			}

			status := _studyPlanItem.Status.String
			if status != "" {
				studyPlanItem.Status = pb.StudyPlanItemStatus(pb.StudyPlanItemStatus_value[status])
			}

			schoolDate := _studyPlanItem.SchoolDate.Time
			if !schoolDate.IsZero() {
				studyPlanItem.SchoolDate = timestamppb.New(schoolDate)
			}
			updateStudyPlanItems = append(updateStudyPlanItems, studyPlanItem)
		}
	}
	return newStudyPlanItems, updateStudyPlanItems, newScheduleStudyPlanItems, nil
}

func (s *AssignStudyPlanTaskModifierService) individualStudyPlanUpdate(ctx context.Context, studyPlanIDs []string, studyPlanItems []*pb.StudyPlanItem, scheduleStudyPlanItems []*pb.ScheduleStudyPlan) ([]*pb.StudyPlanItem, []*pb.StudyPlanItem, []*pb.ScheduleStudyPlan, error) {
	studyPlanItemIDs := make([]string, len(studyPlanItems))
	for i, item := range studyPlanItems {
		studyPlanItemIDs[i] = item.StudyPlanItemId
	}

	oldStudyPlanItems, err := s.StudyPlanItemRepo.FindByIDs(ctx, s.DB, database.TextArray(studyPlanItemIDs))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("s.StudyPlanItemRepo.FindByIDs: %w", err)
	}

	oldStudyPlanItemMap := make(map[string]*entities.StudyPlanItem)
	for _, oldStudyPlanItem := range oldStudyPlanItems {
		oldStudyPlanItemMap[oldStudyPlanItem.ID.String] = oldStudyPlanItem
	}
	var newStudyPlanItems, updateStudyPlanItems []*pb.StudyPlanItem
	var newScheduleStudyPlanItems []*pb.ScheduleStudyPlan
	for _, studyPlanID := range studyPlanIDs {
		for i, studyPlanItem := range studyPlanItems {
			oldStudyPlanItem := oldStudyPlanItemMap[studyPlanItem.StudyPlanItemId]
			if oldStudyPlanItem != nil && oldStudyPlanItem.StudyPlanID.String == studyPlanID {
				studyPlanItem.StudyPlanId = studyPlanID
				completedAt := oldStudyPlanItem.CompletedAt.Time
				if !completedAt.IsZero() {
					studyPlanItem.CompletedAt = timestamppb.New(completedAt)
				}
				schoolDate := oldStudyPlanItem.SchoolDate.Time
				if !schoolDate.IsZero() {
					studyPlanItem.SchoolDate = timestamppb.New(schoolDate)
				}
				status := oldStudyPlanItem.Status.String
				if status != "" {
					studyPlanItem.Status = pb.StudyPlanItemStatus(pb.StudyPlanItemStatus_value[status])
				}
				updateStudyPlanItems = append(updateStudyPlanItems, studyPlanItem)
			} else {
				newStudyPlanItem := &pb.StudyPlanItem{
					StudyPlanId:      studyPlanItem.StudyPlanId,
					StudyPlanItemId:  idutil.ULIDNow(),
					AvailableFrom:    studyPlanItem.AvailableFrom,
					AvailableTo:      studyPlanItem.AvailableTo,
					StartDate:        studyPlanItem.StartDate,
					EndDate:          studyPlanItem.EndDate,
					ContentStructure: studyPlanItem.ContentStructure,
					DisplayOrder:     studyPlanItem.DisplayOrder,
				}
				newStudyPlanItems = append(newStudyPlanItems, newStudyPlanItem)
				newScheduleStudyPlanItem := &pb.ScheduleStudyPlan{
					StudyPlanItemId: newStudyPlanItem.StudyPlanId,
					Item:            scheduleStudyPlanItems[i].Item,
				}
				newScheduleStudyPlanItem.StudyPlanItemId = newStudyPlanItem.StudyPlanId
				newScheduleStudyPlanItems = append(newScheduleStudyPlanItems, newScheduleStudyPlanItem)
				continue
			}
		}
	}
	return newStudyPlanItems, updateStudyPlanItems, newScheduleStudyPlanItems, nil
}

func (s *AssignStudyPlanTaskModifierService) InsertNewStudyPlanItemToAllDependStudyPlan(ctx context.Context, courseStudyPlanID string, dependStudyPlanIDs []string,
	newStudyPlanItems []*pb.StudyPlanItem, scheduleStudyPlanItems []*pb.ScheduleStudyPlan, tx pgx.Tx) error {
	var insertStudyPlanItems []*entities.StudyPlanItem
	var insertScheduleStudyPlanItems []*pb.ScheduleStudyPlan
	for i, newStudyPlanItem := range newStudyPlanItems {
		studyPlanEn, err := ToStudyPlanItem(newStudyPlanItem)
		if err != nil {
			return err
		}
		if courseStudyPlanID != "" {
			studyPlanEn.StudyPlanID.Set(courseStudyPlanID)
			item := &pb.ScheduleStudyPlan{
				StudyPlanItemId: newStudyPlanItem.StudyPlanItemId,
				Item:            scheduleStudyPlanItems[i].Item,
			}
			insertScheduleStudyPlanItems = append(insertScheduleStudyPlanItems, item)
			insertStudyPlanItems = append(insertStudyPlanItems, studyPlanEn)
		}

		for _, individualStudyPlanID := range dependStudyPlanIDs {
			t := *studyPlanEn
			if courseStudyPlanID != "" {
				t.CopyStudyPlanItemID.Set(studyPlanEn.ID)
			}
			err := multierr.Combine(
				t.StudyPlanID.Set(individualStudyPlanID),
				t.ID.Set(idutil.ULIDNow()),
			)
			if err != nil {
				return err
			}
			scheduleStudyPlanItem := &pb.ScheduleStudyPlan{
				StudyPlanItemId: t.ID.String,
				Item:            scheduleStudyPlanItems[i].Item,
			}
			insertStudyPlanItems = append(insertStudyPlanItems, &t)
			insertScheduleStudyPlanItems = append(insertScheduleStudyPlanItems, scheduleStudyPlanItem)
		}
	}

	err := s.StudyPlanItemRepo.BulkInsert(ctx, tx, insertStudyPlanItems)
	if err != nil {
		return fmt.Errorf("s.StudyPlanItemRepo.BulkInsert: %w", err)
	}
	err = ScheduleStudyPlanWithTx(ctx, tx, &pb.ScheduleStudyPlanRequest{
		Schedule: insertScheduleStudyPlanItems,
	}, s.AssignmentStudyPlanItemRepo.BulkInsert, s.LoStudyPlanItemRepo.BulkInsert)
	if err != nil {
		return fmt.Errorf("s.ScheduleStudyPlanWithTx: %w", err)
	}

	return nil
}

func (s *AssignStudyPlanTaskModifierService) updateStudyPlan(ctx context.Context, req *pb.AssignStudyPlanEvent) error {
	r := &IAssignStudyPlan{
		StudyPlanRepo:               s.StudyPlanRepo,
		CourseStudyPlanRepo:         s.CourseStudyPlanRepo,
		StudentRepo:                 s.StudentRepo,
		StudentStudyPlan:            s.StudentStudyPlan,
		StudyPlanItemRepo:           s.StudyPlanItemRepo,
		AssignmentStudyPlanItemRepo: s.AssignmentStudyPlanItemRepo,
		LoStudyPlanItemRepo:         s.LoStudyPlanItemRepo,
	}

	var newStudyPlanItems, updateStudyPlanItems []*pb.StudyPlanItem
	var newScheduleStudyPlanItems []*pb.ScheduleStudyPlan
	var allStudyPlans, individualStudyPlanIDs []string
	var courseStudyPlanID string

	if req.Type == pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE {
		courseStudyPlanID = req.StudyPlanIds[0]
		var err error
		individualStudyPlanIDs, err = s.StudyPlanRepo.FindDependStudyPlan(ctx, s.DB, database.TextArray(req.StudyPlanIds))
		if err != nil {
			return fmt.Errorf("s.StudyPlanRepo.FindDependStudyPlan: %w", err)
		}

		allStudyPlans = append(individualStudyPlanIDs, req.StudyPlanIds...)

		newStudyPlanItems, updateStudyPlanItems, newScheduleStudyPlanItems, err = s.makeCourseStudyPlanUpdateData(ctx, courseStudyPlanID,
			req.StudyPlanItems, req.ScheduleStudyPlanItems)
		if err != nil {
			return err
		}
	}
	if req.Type == pb.StudyPlanType_STUDY_PLAN_TYPE_INDIVIDUAL {
		courseStudyPlanID = ""
		var err error
		newStudyPlanItems, updateStudyPlanItems, newScheduleStudyPlanItems, err = s.individualStudyPlanUpdate(ctx, req.StudyPlanIds,
			req.StudyPlanItems, req.ScheduleStudyPlanItems)
		if err != nil {
			return err
		}
		allStudyPlans = req.StudyPlanIds
		individualStudyPlanIDs = req.StudyPlanIds
	}
	err := database.ExecInTxWithRetry(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		studyPlans := make([]*pb.UpsertStudyPlansRequest_StudyPlan, len(req.StudyPlanIds))
		for i, studyPlanID := range req.StudyPlanIds {
			studyPlans[i] = &pb.UpsertStudyPlansRequest_StudyPlan{
				StudyPlanId: wrapperspb.String(studyPlanID),
				SchoolId:    req.SchoolId,
				Name:        req.Name,
				Country:     cpb.Country_COUNTRY_ID,
				CourseId:    req.CourseId,
				Type:        req.Type,
			}
		}
		_, errTx := UpsertStudyPlansInTx(ctx, &pb.UpsertStudyPlansRequest{
			StudyPlans: studyPlans,
		}, tx, s.StudyPlanRepo.BulkUpsert)

		if errTx != nil {
			return fmt.Errorf("UpsertStudyPlansInTx: %w", errTx)
		}

		// remove all study plan item then upsert again
		errTx = s.StudyPlanItemRepo.SoftDeleteWithStudyPlanIDs(ctx, tx, database.TextArray(allStudyPlans))
		if errTx != nil {
			return fmt.Errorf("StudyPlanItemRepo.SoftDeleteWithStudyPlanIDs: %w", errTx)
		}

		_, errTx = UpsertStudyPlanItemWithTx(ctx, &pb.UpsertStudyPlanItemRequest{
			StudyPlanItems: updateStudyPlanItems,
		}, tx, r)
		if errTx != nil {
			return fmt.Errorf("UpsertStudyPlanItemWithTx: %w", errTx)
		}
		// insert new study plan item
		errTx = s.InsertNewStudyPlanItemToAllDependStudyPlan(ctx, courseStudyPlanID, individualStudyPlanIDs, newStudyPlanItems, newScheduleStudyPlanItems, tx)
		if errTx != nil {
			return fmt.Errorf("InsertNewStudyPlanItemToAllDependStudyPlan: %w", errTx)
		}

		completeStatus := database.Text(pb.StudyPlanTaskStatus_STUDY_PLAN_TASK_STATUS_COMPLETED.String())
		errTx = s.AssignStudyPlanTaskRepo.UpdateStatus(ctx, tx, database.Text(req.TaskId), completeStatus)
		if errTx != nil {
			return fmt.Errorf("s.AssignStudyPlanTaskRepo.UpdateStatus: %v", errTx)
		}

		return nil
	})

	if err != nil {
		// failed to upsert then update status error to study plan task
		errorStatus := database.Text(pb.StudyPlanTaskStatus_STUDY_PLAN_TASK_STATUS_ERROR.String())
		errDetail := &repositories.AssignStudyPlanTaskDetailErrorArgs{
			ID:          database.Text(req.TaskId),
			Status:      errorStatus,
			ErrorDetail: database.Text(err.Error()),
		}
		// if when update fail again we have to combine two error and return
		if err2 := s.AssignStudyPlanTaskRepo.UpdateDetailError(ctx, s.DB, errDetail); err2 != nil {
			return fmt.Errorf(multierr.Combine(err, fmt.Errorf("AssignStudyPlanTaskRepo.UpdateDetailError: %w", err2)).Error())
		}
		return err
	}

	return nil
}
