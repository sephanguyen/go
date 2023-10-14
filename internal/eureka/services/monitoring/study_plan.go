package services

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/eureka/entities"
	monitor_ent "github.com/manabie-com/backend/internal/eureka/entities/monitors"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	monitor_repo "github.com/manabie-com/backend/internal/eureka/repositories/monitors"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type StudyPlanMonitorService struct {
	DB database.Ext

	Cfg    *configurations.Config
	Logger zap.Logger

	Alert alert.SlackFactory

	StudentStudyPlanRepo interface {
		RetrieveByStudentCourse(ctx context.Context, db database.QueryExecer, studentIDs, courseIDs pgtype.TextArray) ([]*entities.StudentStudyPlan, error)
	}
	CourseStudentRepo interface {
		RetrieveByIntervalTime(ctx context.Context, db database.QueryExecer, intervalTime pgtype.Text) ([]*entities.CourseStudent, error)
	}
	StudyPlanRepo interface {
		RetrieveMasterByCourseIDs(ctx context.Context, db database.QueryExecer, studyPlanType pgtype.Text, courseIDs pgtype.TextArray) ([]*entities.StudyPlan, error)
		RetrieveCombineStudent(ctx context.Context, db database.QueryExecer, bookIDs pgtype.TextArray) ([]*entities.StudyPlanCombineStudentID, error)
	}
	StudyPlanMonitorRepo interface {
		RetrieveByFilter(ctx context.Context, db database.QueryExecer, filter *monitor_repo.RetrieveFilter) ([]*monitor_ent.StudyPlanMonitor, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, studyPlanMonitorIDs pgtype.TextArray) error
		BulkUpsert(ctx context.Context, db database.QueryExecer, studyPlanMonitorItems []*monitor_ent.StudyPlanMonitor) error
		SoftDeleteTypeStudyPlan(ctx context.Context, db database.QueryExecer, filter *monitor_repo.RetrieveFilter) error
		MarkItemsAutoUpserted(ctx context.Context, db database.QueryExecer, studyPlanMonitorIDs pgtype.TextArray) error
	}
	AssignmentRepo interface {
		RetrieveBookAssignmentByIntervalTime(ctx context.Context, db database.QueryExecer, intervalTime pgtype.Text) ([]*entities.BookAssignment, error)
	}
	StudyPlanItemRepo interface {
		RetrieveByBookContent(ctx context.Context, db database.QueryExecer, bookIDs, loIDs, assignmentIDs pgtype.TextArray) ([]*entities.StudyPlanItem, error)
		FindWithFilterV2(ctx context.Context, db database.QueryExecer, filter *repositories.FilterStudyPlanItemArgs) ([]*entities.StudyPlanItem, error)
		BulkSync(ctx context.Context, db database.QueryExecer, items []*entities.StudyPlanItem) (insertItems []*entities.StudyPlanItem, err error)
	}
	LearningObjectiveRepo interface {
		RetrieveBookLoByIntervalTime(ctx context.Context, db database.QueryExecer, intervalTime pgtype.Text) ([]*entities.BookLearningObjective, error)
	}
	LoStudyPlanItemRepo interface {
		BulkInsert(ctx context.Context, db database.QueryExecer, loStudyPlanItems []*entities.LoStudyPlanItem) error
	}
	AssignmentStudyPlanItemRepo interface {
		BulkInsert(ctx context.Context, db database.QueryExecer, assignmentStudyPlanItems []*entities.AssignmentStudyPlanItem) error
	}
}

// UpsertStudentCourse will collect invalid data when upsert student to course
// in cronjob -> 15 min each times(default)
func (s *StudyPlanMonitorService) UpsertStudentCourse(ctx context.Context, timeCron int) error {
	// we have 2 phases
	// + confirm again previous
	// + re-confirm again the course student still alive or not

	// on phase 2:
	// collect invalid data and write to database
	errChan := make(chan error, 2)
	wg := sync.WaitGroup{}
	wg.Add(2)
	//TODO:  re-confirm again the course student still alive or not
	go func() {
		defer wg.Done()
		err := s.ReVerifyMissingStudentStudyPlan(ctx, timeCron)
		if err != nil {
			errChan <- err
		}
	}()
	//TODO: make flex time interval
	go func() {
		defer wg.Done()
		err := s.CollectMissingStudentStudyPlan(ctx, timeCron)
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()
	var err error
	for er := range errChan {
		if er == nil {
			continue
		}

		err = multierr.Append(err, er)
	}

	return err
}

func convCourseStudyPlanMap(studyPlans []*entities.StudyPlan) map[string][]string {
	mapCourseStudyPlan := make(map[string][]string)
	for _, sp := range studyPlans {
		mapCourseStudyPlan[sp.CourseID.String] = make([]string, 0)
		mapCourseStudyPlan[sp.CourseID.String] = append(mapCourseStudyPlan[sp.CourseID.String], sp.ID.String)
	}
	return mapCourseStudyPlan
}

type StudentMasterStudyPlan struct {
	MasterStudyPlanID string
	StudentID         string
}

func convStudentMasterStudyPlanMap(studentStudyPlans []*entities.StudentStudyPlan) map[StudentMasterStudyPlan]bool {
	mapStudentMasterStudyPlan := make(map[StudentMasterStudyPlan]bool)
	for _, ssp := range studentStudyPlans {
		mapStudentMasterStudyPlan[StudentMasterStudyPlan{
			MasterStudyPlanID: ssp.MasterStudyPlanID.String,
			StudentID:         ssp.StudentID.String}] = true
	}
	return mapStudentMasterStudyPlan
}

func (s *StudyPlanMonitorService) ReVerifyMissingStudentStudyPlan(ctx context.Context, timeCron int) error {
	CLCTime, UCLTime := genTimeLCLTimeUCLReverify(timeCron)
	err := s.StudyPlanMonitorRepo.SoftDeleteTypeStudyPlan(ctx, s.DB, &monitor_repo.RetrieveFilter{
		StudyPlanMonitorType: database.Text(monitor_ent.StudyPlanMonitorType_STUDENT_STUDY_PLAN),
		IntervalTimeLCL:      &CLCTime,
		IntervalTimeULC:      &UCLTime,
	})
	if err != nil {
		return fmt.Errorf("StudyPlanMonitorService.ReVerifyMissingStudentStudyPlan.StudyPlanMonitorRepo.SoftDeleteTypeStudyPlan: %w", err)
	}
	return nil
}

func (s *StudyPlanMonitorService) CollectMissingStudentStudyPlan(ctx context.Context, timeCron int) error {
	missingData := make([]*monitor_ent.StudyPlanMonitor, 0)
	LCLTime, _ := genTimeLCLTimeUCL(timeCron)
	courseStudents, err := s.CourseStudentRepo.RetrieveByIntervalTime(ctx, s.DB, LCLTime)
	if err != nil {
		return fmt.Errorf("StudyPlanMonitorService.CourseStudentRepo.RetrieveByIntervalTime: %w", err)
	}
	courseIDs, studentIDs := retrieveCourseIDStudentID(courseStudents)

	studentStudyPlans, err := s.StudentStudyPlanRepo.RetrieveByStudentCourse(ctx, s.DB, database.TextArray(studentIDs), database.TextArray(courseIDs))
	if err != nil {
		return fmt.Errorf("StudyPlanMonitorService.StudentStudyPlanRepo.RetrieveByStudentCourse: %w", err)
	}
	masterStudyPlans, err := s.StudyPlanRepo.RetrieveMasterByCourseIDs(ctx, s.DB, database.Text(epb.StudyPlanType_STUDY_PLAN_TYPE_COURSE.String()), database.TextArray(courseIDs))
	if err != nil {
		return fmt.Errorf("StudyPlanMonitorService.StudyPlanRepo.RetrieveMasterByCourseIDs: %w", err)
	}
	mapCourseStudyPlan := convCourseStudyPlanMap(masterStudyPlans)
	mapStudentMasterStudyPlan := convStudentMasterStudyPlanMap(studentStudyPlans)
	for _, cs := range courseStudents {
		// TODO: improvement, we can call to bob and check content each course book then warning
		// TODO: avoid duplicate when run > 1 times
		if masterStudyPlanIDs, ok := mapCourseStudyPlan[cs.CourseID.String]; ok {
			for _, spID := range masterStudyPlanIDs {
				if _, isCreated := mapStudentMasterStudyPlan[StudentMasterStudyPlan{
					StudentID:         cs.StudentID.String,
					MasterStudyPlanID: spID,
				}]; !isCreated {
					studyPlanmonitorEnt, err := generateUpsertCourseStudentMissingData(cs.CourseID.String, cs.StudentID.String, spID)
					if err != nil {
						return fmt.Errorf("unable to generate study_plan_monitor ent: %w", err)
					}
					//TODO: verify before insert (avoid duplicate if run >1)
					missingData = append(missingData, studyPlanmonitorEnt)
				}
			}
		}
	}
	if len(missingData) > 0 {
		if err = s.StudyPlanMonitorRepo.BulkUpsert(ctx, s.DB, missingData); err != nil {
			return fmt.Errorf("StudyPlanMonitorService.StudyPlanMonitorRepo.BulkUpsert: %w", err)
		}
		att := alert.InitAttachment("error")
		att.AddSourceInfo(s.Cfg.SchoolInformation.SchoolName, s.Cfg.Common.Environment)
		att.AddDetailInfo(len(missingData), monitor_ent.StudyPlanMonitorType_STUDENT_STUDY_PLAN)
		err := s.Alert.Send(alert.Payload{
			Text: "Missing items",
			Attachments: []alert.IAttachment{
				att,
			},
		})
		if err != nil {
			s.Logger.Error("CollectMissingStudentStudyPlan: fail to notify to slack ", zap.Error(err))
		}
	}

	return nil
}

func generateUpsertCourseStudentMissingData(courseID, studentID, masterStudyPlanID string) (*monitor_ent.StudyPlanMonitor, error) {
	e := &monitor_ent.StudyPlanMonitor{}
	database.AllNullEntity(e)
	now := timeutil.Now()
	studyPlanMonitorID := idutil.ULIDNow()
	p := &monitor_ent.StudyPlanMonitorPayload{
		BookID:            pgtype.Text{Status: pgtype.Null},
		StudyPlanID:       pgtype.Text{Status: pgtype.Null},
		MasterStudyPlanID: pgtype.Text{Status: pgtype.Null},
		LoID:              pgtype.Text{Status: pgtype.Null},
		TopicID:           pgtype.Text{Status: pgtype.Null},
		ChapterID:         pgtype.Text{Status: pgtype.Null},
		AssignmentID:      pgtype.Text{Status: pgtype.Null},
		LMDisplayOrder:    pgtype.Int4{Status: pgtype.Null},
	}
	err := multierr.Combine(
		e.CourseID.Set(courseID),
		e.StudentID.Set(studentID),
		e.StudyPlanMonitorID.Set(studyPlanMonitorID),
		e.Type.Set(monitor_ent.StudyPlanMonitorType_STUDENT_STUDY_PLAN),
		e.Payload.Set(p),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	return e, err
}

func retrieveCourseIDStudentID(css []*entities.CourseStudent) (courseIDs []string, studentIDs []string) {
	courseIDs = make([]string, 0, len(css))
	studentIDs = make([]string, 0, len(css))

	for _, cs := range css {
		courseIDs = append(courseIDs, cs.CourseID.String)
		studentIDs = append(studentIDs, cs.StudentID.String)
	}
	return
}

// genTimeLCLTimeUCL from a given time -> gen timeLCL and timeUCL
// each time excute Monitoring ex 15'
// time1----15' later-----time2
// because the record save to db depend on lasttime (time1), so the ULC time only need less than the begin time of time2 (now - UCL)
// the range time when excute maybe need some seconds to start, so 1m enough.
func genTimeLCLTimeUCL(time int) (timeLCL, timeUCL pgtype.Text) {
	timeLCL = database.Text(strconv.Itoa(time+1) + " mins")
	timeUCL = database.Text(strconv.Itoa(1) + " mins")
	return
}

// like above but we will cover 2 previous times
func genTimeLCLTimeUCLReverify(time int) (timeLCL, timeUCL pgtype.Text) {
	timeLCL = database.Text(strconv.Itoa(time*2+1) + " mins")
	timeUCL = database.Text(strconv.Itoa(1) + " mins")
	return
}

/*


 */

func (s *StudyPlanMonitorService) UpsertLearningItems(ctx context.Context, timeCron int, orgID string) error {
	// re verify again
	// get
	// retrieve lo - student - master study plan
	// retrieve assignment -student- master study plan

	// collect data
	// -> call bob get LO
	// -> repo get assignment
	err := s.CollectMissingLearningItems(ctx, timeCron, orgID)
	if err != nil {
		return fmt.Errorf("StudyPlanMonitorService.UpsertLearningItems: %w", err)
	}
	return nil
}

func (s *StudyPlanMonitorService) ReVerifyMissingLearningItems(ctx context.Context, timeCron int) error {
	//TODO: re-verify
	return nil
}

type StudyPlanItemLearningItem struct {
	StudyPlanID  string `json:"study_plan_id,omitempty"`
	LoID         string `json:"lo_id,omitempty"`
	AssignmentID string `json:"assignment_id,omitempty"`
}

// NOTE 1: with monitor, so I don't need to create new index, waste resource
/*WRONG IDEA: you can reference when check bug, or something, hope it can help you: at first sight I think about retrieve
all study plan item and clarify which master study plan(course), which child(student) study plan.
But it hard/slow:  hard to retrieve data, hard to travel the data, and the important is we can't determine if
the study plan items of the root is missing. If we want to do it, you have to retrieve all the book tree,
it very complicated, you can optimize but it not work a lot.
*/
/*
BETTER IDEA: from lo and assignment we can get the BookID
divide to 2 parts:
 + 1 book - assignments(slice) (1)
 + 2 book - los (slice) 		(2)
 -> get BookIDs, get loIDs, assignmentIDs
- from BookIDs -> get all study plans with this book, (in repo - left join to get student_id) (3)
- from BookIDs assignmentIDs loIDs: get study plan items (4)
MAIN idea: travel all study plans from (3)
check the book (1) or (2) or both (1)and(2) have value or not (5)
travel value(s) (slice) from (5) check data on (4) if not existed -> missing
*/
func (s *StudyPlanMonitorService) CollectMissingLearningItems(ctx context.Context, timeCron int, orgID string) error {
	missingData := make([]*monitor_ent.StudyPlanMonitor, 0)
	LCLTime, _ := genTimeLCLTimeUCL(timeCron)
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		bookAssignments, err := s.AssignmentRepo.RetrieveBookAssignmentByIntervalTime(ctx, tx, LCLTime)
		if err != nil {
			return fmt.Errorf("StudyPlanMonitorService.CollectMissingLearningItems.AssignmentRepo.RetrieveBookAssignmentByIntervalTime: %w", err)
		}
		bookLos, err := s.LearningObjectiveRepo.RetrieveBookLoByIntervalTime(ctx, tx, LCLTime)
		if err != nil {
			return fmt.Errorf("StudyPlanMonitorService.CollectMissingLearningItems.LearningObjectiveRepo.RetrieveBookLoByIntervalTime: %w", err)
		}
		bookIDs := make([]string, 0)
		// studyPlans(+ student id) it have student id
		mapBookLo, mapLoContentStructure, mapLoDisplayOrder, loIDs, bookIDs1 := retrieveInfoBookLoIDs(bookLos)
		mapBookAssignment, mapAssignmentContentStructure, mapAssignmentDisplayOrder, assignmentIDs, bookIDs2 := retrieveInfoBookAssignmentIDs(bookAssignments)
		bookIDs = append(bookIDs, bookIDs1...)
		bookIDs = append(bookIDs, bookIDs2...)
		bookIDs = golibs.GetUniqueElementStringArray(bookIDs)

		studyPlans, err := s.StudyPlanRepo.RetrieveCombineStudent(ctx, tx, database.TextArray(bookIDs))
		if err != nil {
			return fmt.Errorf("StudyPlanRepo.RetrieveCombineStudent: %w", err)
		}
		if len(studyPlans) == 0 {
			return nil // nothing to check
		}
		studyPlanItems, err := s.StudyPlanItemRepo.RetrieveByBookContent(ctx, tx, database.TextArray(bookIDs), database.TextArray(loIDs), database.TextArray(assignmentIDs))
		if err != nil {
			return fmt.Errorf("StudyPlanItemRepo.RetrieveCombineStudent: %w", err)
		}
		mapStudyPlanItemLearningItem := convert2MapStudyPlanItemLearningItem(studyPlanItems)

		for _, sp := range studyPlans {
			if loIDs, ok1 := mapBookLo[sp.BookID.String]; ok1 {
				for _, loID := range loIDs {
					if _, ok2 := mapStudyPlanItemLearningItem[StudyPlanItemLearningItem{
						StudyPlanID: sp.ID.String,
						LoID:        loID,
					}]; !ok2 {
						missingStudyPlanItemInfo, err := generateStudyPlanItemMissingData(mapLoContentStructure, mapLoDisplayOrder, sp, loID, "")
						if err != nil {
							return fmt.Errorf("generateStudyPlanItemMissingData-lo: %w", err)
						}
						missingData = append(missingData, missingStudyPlanItemInfo)
					}
				}
			}
			if assignmentIDs, ok1 := mapBookAssignment[sp.BookID.String]; ok1 {
				for _, assignmentID := range assignmentIDs {
					if _, ok2 := mapStudyPlanItemLearningItem[StudyPlanItemLearningItem{
						StudyPlanID:  sp.ID.String,
						AssignmentID: assignmentID,
					}]; !ok2 {
						missingStudyPlanItemInfo, err := generateStudyPlanItemMissingData(mapAssignmentContentStructure, mapAssignmentDisplayOrder, sp, "", assignmentID)
						if err != nil {
							return fmt.Errorf("generateStudyPlanItemMissingData-assignment: %w", err)
						}
						missingData = append(missingData, missingStudyPlanItemInfo)
					}
				}
			}
		}
		if len(missingData) > 0 {
			if err = s.StudyPlanMonitorRepo.BulkUpsert(ctx, tx, missingData); err != nil {
				return fmt.Errorf("StudyPlanMonitorService.CollectMissingLearningItems.StudyPlanMonitorRepo.BulkUpsert: %w", err)
			}
			att := alert.InitAttachment("error")

			att.AddSourceInfo(s.Cfg.SchoolInformation.SchoolName, s.Cfg.Common.Environment)
			att.AddDetailInfo(len(missingData), monitor_ent.StudyPlanMonitorType_STUDY_PLAN_ITEM)
			err := s.Alert.Send(alert.Payload{
				Text: "Missing items",
				Attachments: []alert.IAttachment{
					att,
				},
			})
			if err != nil {
				s.Logger.Error("CollectMissingLearningItems: fail to notify to slack ", zap.Error(err))
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if len(missingData) > 0 {
		if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			if err := s.autoUpsertSPIs(ctx, tx, missingData); err != nil {
				return fmt.Errorf("autoUpsertStudyPlanItems: %w", err)
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *StudyPlanMonitorService) autoUpsertSPIs(ctx context.Context, tx database.QueryExecer, missingData []*monitor_ent.StudyPlanMonitor) (err error) {
	masterSpis := []*entities.StudyPlanItem{}
	individualSpis := []*entities.StudyPlanItem{}
	studyPlanMonitorIDs := []string{}
	loStudyPlanItems := []*entities.LoStudyPlanItem{}
	assignmentStudyPlanItems := []*entities.AssignmentStudyPlanItem{}
	mapContentStructureFlatten := map[string]bool{}
	for _, item := range missingData {
		if item.StudentID.String == "" {
			loSPI, assSPI, studyPlanItem, err := s.generateStudyPlanItem(ctx, tx, item, false)
			if err != nil {
				return fmt.Errorf("error generateMasterSPI: %w", err)
			}

			key := fmt.Sprintf("%s-%s", studyPlanItem.StudyPlanID.String, studyPlanItem.ContentStructureFlatten.String)
			if _, ok := mapContentStructureFlatten[key]; !ok {
				mapContentStructureFlatten[key] = true
				masterSpis = append(masterSpis, studyPlanItem)
				studyPlanMonitorIDs = append(studyPlanMonitorIDs, item.StudyPlanMonitorID.String)

				if loSPI != nil {
					loStudyPlanItems = append(loStudyPlanItems, loSPI)
				} else {
					assignmentStudyPlanItems = append(assignmentStudyPlanItems, assSPI)
				}
			}
		}
	}

	// we need upsert master spis first to individual spis can get copy_study_plan_item_id field
	_, err = s.StudyPlanItemRepo.BulkSync(ctx, tx, masterSpis)
	if err != nil {
		return fmt.Errorf("error BulkSync: %w", err)
	}

	for _, item := range missingData {
		if item.StudentID.String != "" {
			loSPI, assSPI, studyPlanItem, err := s.generateStudyPlanItem(ctx, tx, item, true)
			if err != nil {
				return fmt.Errorf("error generateIndividualSPI: %w", err)
			}

			key := fmt.Sprintf("%s-%s", studyPlanItem.StudyPlanID.String, studyPlanItem.ContentStructureFlatten.String)
			if _, ok := mapContentStructureFlatten[key]; !ok {
				mapContentStructureFlatten[key] = true
				individualSpis = append(individualSpis, studyPlanItem)
				studyPlanMonitorIDs = append(studyPlanMonitorIDs, item.StudyPlanMonitorID.String)

				if loSPI != nil {
					loStudyPlanItems = append(loStudyPlanItems, loSPI)
				} else {
					assignmentStudyPlanItems = append(assignmentStudyPlanItems, assSPI)
				}
			}
		}
	}

	// upsert individual spis
	_, err = s.StudyPlanItemRepo.BulkSync(ctx, tx, individualSpis)
	if err != nil {
		return fmt.Errorf("error BulkSync: %w", err)
	}

	err = s.LoStudyPlanItemRepo.BulkInsert(ctx, tx, loStudyPlanItems)
	if err != nil {
		return fmt.Errorf("error s.LoStudyPlanItemRepo.BulkInsert: %w", err)
	}

	err = s.AssignmentStudyPlanItemRepo.BulkInsert(ctx, tx, assignmentStudyPlanItems)
	if err != nil {
		return fmt.Errorf("error s.AssignmentStudyPlanItemRepo.BulkInsert: %w", err)
	}

	err = s.StudyPlanMonitorRepo.MarkItemsAutoUpserted(ctx, tx, database.TextArray(studyPlanMonitorIDs))
	if err != nil {
		return fmt.Errorf("error MarkItemsAutoUpserted: %w", err)
	}

	return nil
}

func (s *StudyPlanMonitorService) generateStudyPlanItem(ctx context.Context, tx database.QueryExecer, item *monitor_ent.StudyPlanMonitor, isIndividualSpi bool) (*entities.LoStudyPlanItem, *entities.AssignmentStudyPlanItem, *entities.StudyPlanItem, error) {
	payload := &monitor_ent.StudyPlanMonitorPayload{}
	item.Payload.AssignTo(payload)
	studyPlanItem := &entities.StudyPlanItem{}
	var loStudyPlanItem *entities.LoStudyPlanItem
	var assStudyPlanItem *entities.AssignmentStudyPlanItem
	var contentStructureFlatten string
	database.AllNullEntity(studyPlanItem)

	cs := &epb.ContentStructure{
		CourseId:  item.CourseID.String,
		BookId:    payload.BookID.String,
		ChapterId: payload.ChapterID.String,
		TopicId:   payload.TopicID.String,
	}

	if payload.AssignmentID.String != "" {
		cs.ItemId = &epb.ContentStructure_AssignmentId{
			AssignmentId: &wrapperspb.StringValue{
				Value: payload.AssignmentID.String,
			},
		}
		contentStructureFlatten = toContentStructureFlattenAssignment(cs, payload.AssignmentID.String)
	}

	if payload.LoID.String != "" {
		cs.ItemId = &epb.ContentStructure_LoId{
			LoId: &wrapperspb.StringValue{
				Value: payload.LoID.String,
			},
		}
		contentStructureFlatten = toContentStructureFlattenLO(cs, payload.LoID.String)
	}

	id := idutil.ULIDNow()
	if err := multierr.Combine(
		studyPlanItem.ID.Set(id),
		studyPlanItem.StudyPlanID.Set(payload.StudyPlanID),
		studyPlanItem.Status.Set(epb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE),
		studyPlanItem.ContentStructure.Set(toContentStructure(cs)),
		studyPlanItem.ContentStructureFlatten.Set(contentStructureFlatten),
		studyPlanItem.DisplayOrder.Set(payload.LMDisplayOrder),
		studyPlanItem.CreatedAt.Set(time.Now()),
		studyPlanItem.UpdatedAt.Set(time.Now()),
	); err != nil {
		return nil, nil, nil, err
	}

	// get more info from master if have
	if isIndividualSpi {
		filterArgs := &repositories.FilterStudyPlanItemArgs{
			StudyPlanID:  database.Text(payload.MasterStudyPlanID.String),
			TopicID:      database.Text(payload.TopicID.String),
			LoID:         pgtype.Text{Status: pgtype.Null},
			AssignmentID: pgtype.Text{Status: pgtype.Null},
		}

		if payload.LoID.String != "" {
			filterArgs.LoID.Set(payload.LoID.String)
		}

		if payload.AssignmentID.String != "" {
			filterArgs.AssignmentID.Set(payload.AssignmentID.String)
		}

		masterSpis, _ := s.StudyPlanItemRepo.FindWithFilterV2(ctx, tx, filterArgs)
		if len(masterSpis) > 0 {
			if err := multierr.Combine(
				studyPlanItem.CopyStudyPlanItemID.Set(masterSpis[0].ID),
				studyPlanItem.SchoolDate.Set(masterSpis[0].SchoolDate),
			); err != nil {
				return nil, nil, nil, err
			}
		}
	}

	now := time.Now()
	if payload.AssignmentID.String != "" {
		assStudyPlanItem = &entities.AssignmentStudyPlanItem{
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
			AssignmentID:    payload.AssignmentID,
			StudyPlanItemID: database.Text(id),
		}
	} else {
		loStudyPlanItem = &entities.LoStudyPlanItem{
			BaseEntity: entities.BaseEntity{
				CreatedAt: database.Timestamptz(now),
				UpdatedAt: database.Timestamptz(now),
				DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
			},
			LoID:            payload.LoID,
			StudyPlanItemID: database.Text(id),
		}
	}

	return loStudyPlanItem, assStudyPlanItem, studyPlanItem, nil
}

// convert2MapBookAssignmentIDs input []*entities.Assignments
// output mapBook - AssignmentIDs (according book) and AssignmentIDs
func retrieveInfoBookAssignmentIDs(bookAssignments []*entities.BookAssignment) (mapBookAssignment map[string][]string, mapAssignmentContentStructure map[string]*entities.ContentStructure, mapAssignmentDisplayOrder map[string]int, assignmentIDs, bookIDs []string) {
	mapBookAssignment = make(map[string][]string)                               // key: book_id - value: []assignment_ids
	mapAssignmentContentStructure = make(map[string]*entities.ContentStructure) // key assignment_id - value content_structure
	mapAssignmentDisplayOrder = make(map[string]int)
	assignmentIDs = make([]string, 0, len(bookAssignments))
	bookIDs = make([]string, 0)

	for _, bookAssignment := range bookAssignments {
		if _, ok := mapBookAssignment[bookAssignment.BookID.String]; !ok {
			mapBookAssignment[bookAssignment.BookID.String] = make([]string, 0)
			bookIDs = append(bookIDs, bookAssignment.BookID.String)
		}
		mapBookAssignment[bookAssignment.BookID.String] = append(mapBookAssignment[bookAssignment.BookID.String], bookAssignment.ID.String)
		assignmentIDs = append(assignmentIDs, bookAssignment.ID.String)

		if _, ok := mapAssignmentContentStructure[bookAssignment.ID.String]; !ok {
			mapAssignmentContentStructure[bookAssignment.ID.String] = &entities.ContentStructure{
				TopicID:      bookAssignment.TopicID.String,
				ChapterID:    bookAssignment.ChapterID.String,
				AssignmentID: bookAssignment.ID.String,
				BookID:       bookAssignment.BookID.String,
			}
		}

		if _, ok := mapAssignmentDisplayOrder[bookAssignment.ID.String]; !ok {
			mapAssignmentDisplayOrder[bookAssignment.ID.String] = int(bookAssignment.Assignment.DisplayOrder.Int)
		}
	}
	return
}

func retrieveInfoBookLoIDs(bookLos []*entities.BookLearningObjective) (mapBookLo map[string][]string, mapLoContentStructure map[string]*entities.ContentStructure, mapLoDisplayOrder map[string]int, loIDs, bookIDs []string) {
	mapBookLo = make(map[string][]string)                               // key book_id - value []string los
	mapLoContentStructure = make(map[string]*entities.ContentStructure) // key lo_id - value content_structure
	mapLoDisplayOrder = make(map[string]int)
	loIDs = make([]string, 0, len(bookLos))
	bookIDs = make([]string, 0)

	for _, booklo := range bookLos {
		if _, ok := mapBookLo[booklo.BookID.String]; !ok {
			mapBookLo[booklo.BookID.String] = make([]string, 0)
			bookIDs = append(bookIDs, booklo.BookID.String)
		}
		mapBookLo[booklo.BookID.String] = append(mapBookLo[booklo.BookID.String], booklo.ID.String)
		loIDs = append(loIDs, booklo.ID.String)

		if _, ok := mapLoContentStructure[booklo.ID.String]; !ok {
			mapLoContentStructure[booklo.ID.String] = &entities.ContentStructure{
				TopicID:   booklo.TopicID.String,
				ChapterID: booklo.ChapterID.String,
				LoID:      booklo.ID.String,
				BookID:    booklo.BookID.String,
			}
		}

		if _, ok := mapLoDisplayOrder[booklo.ID.String]; !ok {
			mapLoDisplayOrder[booklo.ID.String] = int(booklo.LearningObjective.DisplayOrder.Int)
		}
	}
	return
}

func convert2MapStudyPlanItemLearningItem(spis []*entities.StudyPlanItem) map[StudyPlanItemLearningItem]bool {
	mapStudyPlanItemLearningItem := make(map[StudyPlanItemLearningItem]bool)
	for _, spi := range spis {
		contentTemp := &entities.ContentStructure{}
		_ = spi.ContentStructure.AssignTo(&contentTemp)
		if contentTemp == nil {
			continue
		}
		if contentTemp.LoID != "" {
			mapStudyPlanItemLearningItem[StudyPlanItemLearningItem{
				StudyPlanID: spi.StudyPlanID.String,
				LoID:        contentTemp.LoID,
			}] = true
		} else if contentTemp.AssignmentID != "" {
			mapStudyPlanItemLearningItem[StudyPlanItemLearningItem{
				StudyPlanID:  spi.StudyPlanID.String,
				AssignmentID: contentTemp.AssignmentID,
			}] = true
		}
	}
	return mapStudyPlanItemLearningItem
}

func generateStudyPlanItemMissingData(mapContentStructure map[string]*entities.ContentStructure, mapDisplayOrder map[string]int, sp *entities.StudyPlanCombineStudentID, loID, assignmentID string) (*monitor_ent.StudyPlanMonitor, error) {
	e := &monitor_ent.StudyPlanMonitor{}
	database.AllNullEntity(e)
	now := timeutil.Now()
	studyPlanMonitorID := idutil.ULIDNow()
	p := &monitor_ent.StudyPlanMonitorPayload{
		BookID:            sp.BookID,
		StudyPlanID:       sp.ID,
		MasterStudyPlanID: pgtype.Text{Status: pgtype.Null},
		LoID:              pgtype.Text{Status: pgtype.Null},
		TopicID:           pgtype.Text{Status: pgtype.Null},
		ChapterID:         pgtype.Text{Status: pgtype.Null},
		AssignmentID:      pgtype.Text{Status: pgtype.Null},
		LMDisplayOrder:    pgtype.Int4{Status: pgtype.Null},
	}
	var er error
	if loID != "" {
		er = multierr.Combine(
			p.LoID.Set(loID),
			p.ChapterID.Set(mapContentStructure[loID].ChapterID),
			p.TopicID.Set(mapContentStructure[loID].TopicID),
			p.LMDisplayOrder.Set(mapDisplayOrder[loID]),
		)
	} else if assignmentID != "" {
		er = multierr.Combine(
			p.AssignmentID.Set(assignmentID),
			p.ChapterID.Set(mapContentStructure[assignmentID].ChapterID),
			p.TopicID.Set(mapContentStructure[assignmentID].TopicID),
			p.LMDisplayOrder.Set(mapDisplayOrder[assignmentID]),
		)
	}
	err := multierr.Combine(
		e.CourseID.Set(sp.CourseID.String),
		e.StudentID.Set(sp.StudentID.String),
		e.StudyPlanMonitorID.Set(studyPlanMonitorID),
		e.Type.Set(monitor_ent.StudyPlanMonitorType_STUDY_PLAN_ITEM),
		p.MasterStudyPlanID.Set(sp.MasterStudyPlan.String),
		e.Payload.Set(p),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
		er,
	)
	return e, err
}

func toContentStructure(src *epb.ContentStructure) (output *entities.ContentStructure) {
	if src == nil {
		return nil
	}

	output = &entities.ContentStructure{
		BookID:    src.BookId,
		ChapterID: src.ChapterId,
		CourseID:  src.CourseId,
		TopicID:   src.TopicId,
	}
	if src.GetLoId() != nil {
		output.LoID = src.GetLoId().GetValue()
	}
	if src.GetAssignmentId() != nil {
		output.AssignmentID = src.GetAssignmentId().GetValue()
	}

	return output
}

func toContentStructureFlattenLO(cs *epb.ContentStructure, loID string) string {
	// contentStructureFlatten format:
	return fmt.Sprintf("book::%stopic::%schapter::%scourse::%slo::%s", cs.BookId, cs.TopicId, cs.ChapterId, cs.CourseId, loID)
}

func toContentStructureFlattenAssignment(cs *epb.ContentStructure, assignmentID string) string {
	// contentStructureFlatten format:
	return fmt.Sprintf("book::%stopic::%schapter::%scourse::%sassignment::%s", cs.BookId, cs.TopicId, cs.ChapterId, cs.CourseId, assignmentID)
}
