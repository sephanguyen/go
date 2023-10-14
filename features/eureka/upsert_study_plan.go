package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	CourseType                         = "course"
	IndividualType                     = "individual"
	LearningObjectiveType              = "learning_objective"
	AssignmentType                     = "assignment"
	LearningObjectiveAndAssignmentType = "learning_objective and assignment"
)

type StudyPlanItemInfo struct {
	BookID       string
	CourseID     string
	StudyPlanID  string
	ChapterID    string
	TopicID      string
	LoID         string
	AssignmentID string
}

func (s *suite) userCreateAValidStudyPlanWith(ctx context.Context, studyPlanType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	e := &entities.StudyPlan{}

	stepState.StudyPlanID = idutil.ULIDNow()
	now := time.Now()
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.ID.Set(stepState.StudyPlanID),
		e.Name.Set(fmt.Sprintf("study-plan-name-%s", stepState.StudyPlanID)),
		e.SchoolID.Set(12),
		e.BookID.Set("book-1"),
		e.CourseID.Set(stepState.CourseID),
		e.Grades.Set([]int{1, 2}),
		e.Status.Set(pb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE.String()),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
		e.TrackSchoolProgress.Set(false),
	); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to set course value: %w", err)
	}

	switch studyPlanType {
	case CourseType:
		stepState.StudyPlanType = pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE.String()
		if err := e.StudyPlanType.Set(pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE.String()); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to set studyplan type course: %w", err)
		}
	case IndividualType:
		stepState.StudyPlanType = pb.StudyPlanType_STUDY_PLAN_TYPE_INDIVIDUAL.String()
		if err := e.StudyPlanType.Set(pb.StudyPlanType_STUDY_PLAN_TYPE_INDIVIDUAL.String()); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to set studyplan type individual: %w", err)
		}
	}
	if _, err := database.Insert(ctx, e, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a valid studyplan: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createABook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := &entities.Book{}
	stepState.BookID = idutil.ULIDNow()
	now := time.Now()
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.ID.Set(stepState.BookID),
		e.Name.Set(fmt.Sprintf("book-name-%s", stepState.BookID)),
		e.Country.Set(cpb.Country_COUNTRY_VN.String()),
		e.Subject.Set(cpb.Subject_SUBJECT_CHEMISTRY.String()),
		e.Grade.Set(1),
		e.SchoolID.Set(12),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
		e.CurrentChapterDisplayOrder.Set(0),
	); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to set value for book: %w", err)
	}

	if _, err := database.Insert(ctx, e, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a book: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createAChapter(ctx context.Context, displayOrder int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := &entities.Chapter{}
	stepState.ChapterID = idutil.ULIDNow()
	now := time.Now()
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.ID.Set(stepState.ChapterID),
		e.Name.Set(fmt.Sprintf("chapter-name-%s", stepState.ChapterID)),
		e.Country.Set(cpb.Country_COUNTRY_VN.String()),
		e.Subject.Set(cpb.Subject_SUBJECT_CHEMISTRY.String()),
		e.Grade.Set(1),
		e.DisplayOrder.Set(displayOrder),
		e.SchoolID.Set(12),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
		e.CurrentTopicDisplayOrder.Set(0),
	); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to set value for chapter: %w", err)
	}
	if _, err := database.Insert(ctx, e, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a chapter: %w", err)
	}
	bce := &entities.BookChapter{}
	database.AllNullEntity(bce)
	if err := multierr.Combine(
		bce.BookID.Set(stepState.BookID),
		bce.ChapterID.Set(stepState.ChapterID),
		bce.UpdatedAt.Set(now),
		bce.CreatedAt.Set(now),
	); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to set value for book chapter: %w", err)
	}
	if _, err := database.Insert(ctx, bce, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a book chapter: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkStudyPlansHaveBeenStoredCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := &entities.StudyPlan{}
	fields, _ := e.FieldMap()
	req := stepState.Request.(*pb.UpsertStudyPlanRequest)
	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE study_plan_id = $1 AND deleted_at IS NULL`, strings.Join(fields, ","), e.TableName())
	var sps entities.StudyPlans
	if err := database.Select(ctx, s.DB, stmt, &stepState.StudyPlanID).ScanAll(&sps); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve study plan from database: %w", err)
	}
	if len(sps) != 1 {
		s.ZapLogger.Info(fmt.Sprintf("s.StudyPlanID: %v", stepState.StudyPlanID))
		return StepStateToContext(ctx, stepState), fmt.Errorf("amount of study plan is wrong, expect %v but got %v", 1, len(sps))
	}
	for _, sp := range sps {
		if req.Name != sp.Name.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("name of study plan %s have been stored incorrect, expect %s but got %s", sp.ID.String, req.Name, sp.Name.String)
		}
		if req.TrackSchoolProgress != sp.TrackSchoolProgress.Bool {
			return StepStateToContext(ctx, stepState), fmt.Errorf("track school progress of study plan %s have been stored incorrect, expect %v but got %v", sp.ID.String, req.TrackSchoolProgress, sp.TrackSchoolProgress.Bool)
		}
		grades := make([]int32, 0, len(sp.Grades.Elements))
		for _, v := range sp.Grades.Elements {
			grades = append(grades, v.Int)
		}
		if !reflect.DeepEqual(req.Grades, grades) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("grades of study plan %s have been stored incorrect, expect %v but got %v", sp.ID.String, req.Grades, grades)
		}
		if sp.StudyPlanType.String != pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("study plan type of study plan %s have been stored incorrect, expect %v but got %v", sp.ID.String, pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE.String(), sp.StudyPlanType.String)
		}
	}
	sps = entities.StudyPlans{}
	stmt = fmt.Sprintf(`SELECT %s FROM %s WHERE master_study_plan_id = $1 AND deleted_at IS NULL`, strings.Join(fields, ","), e.TableName())
	if err := database.Select(ctx, s.DB, stmt, &stepState.StudyPlanID).ScanAll(&sps); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve study plan from database: %w", err)
	}
	if len(sps) != len(stepState.StudentIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("amount of study plan is wrong, expect %v but got %v", len(stepState.StudentIDs), len(sps))
	}
	for _, sp := range sps {
		if req.Name != sp.Name.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("name of study plan %s have been stored incorrect, expect %s but got %s", sp.ID.String, req.Name, sp.Name.String)
		}
		if req.TrackSchoolProgress != sp.TrackSchoolProgress.Bool {
			return StepStateToContext(ctx, stepState), fmt.Errorf("track school progress of study plan %s have been stored incorrect, expect %v but got %v", sp.ID.String, req.TrackSchoolProgress, sp.TrackSchoolProgress.Bool)
		}
		grades := make([]int32, 0, len(sp.Grades.Elements))
		for _, v := range sp.Grades.Elements {
			grades = append(grades, v.Int)
		}
		if !reflect.DeepEqual(req.Grades, grades) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("grades of study plan %s have been stored incorrect, expect %v but got %v", sp.ID.String, req.Grades, grades)
		}
		if sp.StudyPlanType.String != pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE.String() {
			return StepStateToContext(ctx, stepState), fmt.Errorf("study plan type of study plan %s have been stored incorrect, expect %v but got %v", sp.ID.String, pb.StudyPlanType_STUDY_PLAN_TYPE_COURSE.String(), sp.StudyPlanType.String)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkCourseStudyPlanHasBeenStoredCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := &entities.CourseStudyPlan{}
	stmt := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE study_plan_id = $1 AND course_id = $2 AND deleted_at IS NULL`, e.TableName())
	var count pgtype.Int8
	if err := s.DB.QueryRow(ctx, stmt, &stepState.StudyPlanID, &stepState.CourseID).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve course study plan from database: %w", err)
	}

	if count.Status != pgtype.Present {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve course study plan")
	}
	if count.Int != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("course study plan not stored correctly, expect %v but got %v", 1, count.Int)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkStudentStudyPlansHaveBeenStoredCorrecly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := &entities.StudentStudyPlan{}
	spe := &entities.StudyPlan{}
	stmt := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %s  as ssp
		JOIN %s as sp
		USING(study_plan_id)
		WHERE sp.master_study_plan_id = $1 
		AND student_id = ANY($2::TEXT[]) 
		AND ssp.deleted_at IS NULL
		AND sp.deleted_at IS NULL`,
		e.TableName(), spe.TableName())
	var count pgtype.Int8
	if err := s.DB.QueryRow(ctx, stmt, &stepState.StudyPlanID, &stepState.StudentIDs).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve student study plan from database: %w", err)
	}

	if count.Status != pgtype.Present {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve student study plan")
	}
	if count.Int != int64(len(stepState.StudentIDs)) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student study plan not stored correctly, expect %v but got %v", len(stepState.StudentIDs), count.Int)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkStudyPlanItemsHaveBeenStoredCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	getKeyStudyPlanItem := func(studyPlanID, loID, assignmentID string) string {
		return strings.Join([]string{studyPlanID, loID, assignmentID}, "|")
	}
	getKeyLoStudyPlanItem := func(studyPlanItemID, loID string) string {
		return strings.Join([]string{studyPlanItemID, loID}, "|")
	}
	getKeyAssignmentStudyPlanItem := func(studyPlanItemID, assignmentID string) string {
		return strings.Join([]string{studyPlanItemID, assignmentID}, "|")
	}
	var (
		loIDs         []string
		assignmentIDs []string
	)
	for _, spItem := range stepState.StudyPlanItemInfos {
		if spItem.LoID != "" {
			loIDs = append(loIDs, spItem.LoID)
		} else {
			assignmentIDs = append(assignmentIDs, spItem.AssignmentID)
		}
	}
	e := &entities.StudyPlanItem{}
	fields, _ := e.FieldMap()
	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE study_plan_id = $1 AND deleted_at IS NULL`, strings.Join(fields, ","), e.TableName())
	var spItems entities.StudyPlanItems
	if err := database.Select(ctx, s.DB, stmt, &stepState.StudyPlanID).ScanAll(&spItems); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve study plan item from database: %w", err)
	}

	lospie := &entities.LoStudyPlanItem{}
	fields, _ = lospie.FieldMap()
	stmt = fmt.Sprintf(`SELECT %s FROM %s WHERE lo_id = ANY($1::TEXT[]) AND deleted_at IS NULL`, strings.Join(fields, ","), lospie.TableName())
	var lospItems entities.LoStudyPlanItems
	if err := database.Select(ctx, s.DB, stmt, &loIDs).ScanAll(&lospItems); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve lo study plan item from database: %w", err)
	}

	aspie := &entities.AssignmentStudyPlanItem{}
	fields, _ = aspie.FieldMap()
	stmt = fmt.Sprintf(`SELECT %s FROM %s WHERE assignment_id = ANY($1::TEXT[]) AND deleted_at IS NULL`, strings.Join(fields, ","), aspie.TableName())
	var aspItems entities.AssignmentStudyPlanItems
	if err := database.Select(ctx, s.DB, stmt, &assignmentIDs).ScanAll(&aspItems); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve assignment study plan item from database: %w", err)
	}

	type ContentStructure struct {
		CourseID     string `json:"course_id,omitempty"`
		BookID       string `json:"book_id,omitempty"`
		ChapterID    string `json:"chapter_id,omitempty"`
		TopicID      string `json:"topic_id,omitempty"`
		LoID         string `json:"lo_id,omitempty"`
		AssignmentID string `json:"assignment_id,omitempty"`
	}

	type StudyPlanItem struct {
		ID      string
		Content *ContentStructure
	}
	spItemsMap := make(map[string]*StudyPlanItem)
	lospItemsMap := make(map[string]bool)
	aspItemsMap := make(map[string]bool)

	for _, item := range lospItems {
		lospItemsMap[getKeyLoStudyPlanItem(item.StudyPlanItemID.String, item.LoID.String)] = true
	}
	for _, item := range aspItems {
		aspItemsMap[getKeyAssignmentStudyPlanItem(item.StudyPlanItemID.String, item.AssignmentID.String)] = true
	}
	for _, item := range spItems {
		var content ContentStructure
		if err := item.ContentStructure.AssignTo(&content); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("item.ContentStructure.AssignTo: %w", err)
		}
		if content.LoID != "" {
			if _, ok := lospItemsMap[getKeyLoStudyPlanItem(item.ID.String, content.LoID)]; !ok {
				return StepStateToContext(ctx, stepState), fmt.Errorf(" lo study plan item does not store correctly,missing study plan item %v lo %v", item.ID.String, content.LoID)
			}
			spItemsMap[getKeyStudyPlanItem(item.StudyPlanID.String, content.LoID, "")] = &StudyPlanItem{
				ID:      item.StudyPlanID.String,
				Content: &content,
			}
		}
		if content.AssignmentID != "" {
			if _, ok := aspItemsMap[getKeyAssignmentStudyPlanItem(item.ID.String, content.AssignmentID)]; !ok {
				return StepStateToContext(ctx, stepState), fmt.Errorf(" assignment study plan item does not store correctly,missing study plan item %v assignment %v", item.ID.String, content.AssignmentID)
			}
			spItemsMap[getKeyStudyPlanItem(item.StudyPlanID.String, "", content.AssignmentID)] = &StudyPlanItem{
				ID:      item.StudyPlanID.String,
				Content: &content,
			}
		}
	}

	for _, item := range stepState.StudyPlanItemInfos {
		spItem, ok := spItemsMap[getKeyStudyPlanItem(stepState.StudyPlanID, item.LoID, item.AssignmentID)]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("study plan item does not store correctly, missing study plan %v lo %v assignment %v", stepState.StudyPlanID, item.LoID, item.AssignmentID)
		}
		if spItem.Content.CourseID != stepState.CourseID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("study plan item %v does not store correctly,course id is wrong expect %v but got %v", spItem.ID, stepState.CourseID, spItem.Content.CourseID)
		}
		if spItem.Content.BookID != stepState.BookID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("study plan item %v does not store correctly,book id is wrong expect %v but got %v", spItem.ID, stepState.BookID, spItem.Content.BookID)
		}
		if spItem.Content.ChapterID != item.ChapterID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("study plan item %v does not store correctly,course id is wrong expect %v but got %v", spItem.ID, item.ChapterID, spItem.Content.CourseID)
		}
		if spItem.Content.TopicID != item.TopicID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("study plan item %v does not store correctly,book id is wrong expect %v but got %v", spItem.ID, item.TopicID, spItem.Content.BookID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studyPlansAndRelatedItemsHaveBeenStored(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if ctx, err := s.checkStudyPlansHaveBeenStoredCorrectly(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if ctx, err := s.checkCourseStudyPlanHasBeenStoredCorrectly(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if ctx, err := s.checkStudentStudyPlansHaveBeenStoredCorrecly(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if ctx, err := s.checkStudyPlanItemsHaveBeenStoredCorrectly(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studyPlansHaveBeenUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.UpsertStudyPlanRequest)
	e := &entities.StudyPlan{}
	fields, _ := e.FieldMap()
	stmt := fmt.Sprintf(`SELECT %s FROM %s WHERE study_plan_id = $1 OR master_study_plan_id = $1 AND deleted_at IS NULL`, strings.Join(fields, ","), e.TableName())
	var result entities.StudyPlans
	if err := database.Select(ctx, s.DB, stmt, &stepState.StudyPlanID).ScanAll(&result); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve study plan by id from database: %w", err)
	}

	for _, res := range result {
		if req.Name != res.Name.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("name of study plan %s have been stored incorrect, expect %s but got %s", res.ID.String, req.Name, res.Name.String)
		}
		if req.TrackSchoolProgress != res.TrackSchoolProgress.Bool {
			return StepStateToContext(ctx, stepState), fmt.Errorf("track school progress of study plan %s have been stored incorrect, expect %v but got %v", res.ID.String, req.TrackSchoolProgress, res.TrackSchoolProgress.Bool)
		}
		if req.Status.String() != res.Status.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("status of study plan %s have been stored incorrect, expect %v but got %v", res.ID.String, req.Status.String(), res.Status.String)
		}
		grades := make([]int32, 0, len(res.Grades.Elements))
		for _, v := range res.Grades.Elements {
			grades = append(grades, v.Int)
		}
		if !reflect.DeepEqual(req.Grades, grades) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("grades of study plan %s have been stored incorrect, expect %v but got %v", res.ID.String, req.Grades, grades)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateAValidStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &pb.UpsertStudyPlanRequest{
		Name:                fmt.Sprintf("studyplan-%s", stepState.StudyPlanID),
		SchoolId:            constants.ManabieSchool,
		TrackSchoolProgress: true,
		Grades:              []int32{3, 4},
		Status:              pb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
		BookId:              stepState.BookID,
		CourseId:            stepState.CourseID,
	}

	resp, err := pb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlan(contextWithToken(s, ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan: %w", err)
	}

	stepState.Request = req
	stepState.StudyPlanID = resp.StudyPlanId
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateAStudyPlanWithInvalidStudy_plan_id(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &pb.UpsertStudyPlanRequest{
		StudyPlanId:         wrapperspb.String("invalid-study-plan-id"),
		Name:                fmt.Sprintf("studyplan-%s", stepState.StudyPlanID),
		SchoolId:            constants.ManabieSchool,
		TrackSchoolProgress: true,
		Grades:              []int32{3, 4},
		Status:              pb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
	}
	stepState.Response, stepState.ResponseErr = pb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlan(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &pb.UpsertStudyPlanRequest{
		StudyPlanId:         wrapperspb.String(stepState.StudyPlanID),
		Name:                fmt.Sprintf("studyplan-%s", stepState.StudyPlanID),
		SchoolId:            constants.ManabieSchool,
		TrackSchoolProgress: true,
		Grades:              []int32{3, 4},
		Status:              pb.StudyPlanStatus_STUDY_PLAN_STATUS_ARCHIVED,
	}
	if _, err := pb.NewStudyPlanModifierServiceClient(s.Conn).UpsertStudyPlan(contextWithToken(s, ctx), req); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert study plan: %w", err)
	}
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userAddABookToCourseDoesNotHaveAnyStudents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := &bob_entities.Course{}
	stepState.CourseID = idutil.ULIDNow()
	database.AllNullEntity(e)
	now := time.Now()
	if err := multierr.Combine(
		e.ID.Set(stepState.ChapterID),
		e.Name.Set(fmt.Sprintf("chapter-name-%s", stepState.ChapterID)),
		e.Country.Set(cpb.Country_COUNTRY_VN.String()),
		e.Subject.Set(cpb.Subject_SUBJECT_CHEMISTRY.String()),
		e.Grade.Set(1),
		e.DisplayOrder.Set(1),
		e.SchoolID.Set(12),
		e.CourseType.Set(bob_pb.COURSE_TYPE_CONTENT.String()),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to set value to course: %w", err)
	}
	if _, err := database.Insert(ctx, e, s.BobDB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a course: %w", err)
	}
	if ctx, err := s.addBookToCourse(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudentIDs = []string{}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addAValidBookWithSomeLearningObjectivesToCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.createABook(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var (
		topics             []*pb.Topic
		los                []*cpb.LearningObjective
		assignments        []*pb.Assignment
		studyPlanItemInfos []*StudyPlanItemInfo
		topicIDs           []string
	)
	n := rand.Intn(2) + 1
	for i := 1; i <= n; i++ {
		if ctx, err := s.createAChapter(ctx, i); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		m := rand.Intn(2) + 1
		for j := 1; j <= m; j++ {
			topicID := idutil.ULIDNow()
			topicIDs = append(topicIDs, topicID)
			topic := &pb.Topic{
				Id:           topicID,
				Name:         fmt.Sprintf("topic-name-%s", topicID),
				Country:      pb.Country_COUNTRY_VN,
				Subject:      pb.Subject_SUBJECT_BIOLOGY,
				DisplayOrder: int32(i),
				SchoolId:     constants.ManabieSchool,

				ChapterId: stepState.ChapterID,
			}
			topics = append(topics, topic)
			for g := 1; g <= 2; g++ {
				loID := idutil.ULIDNow()
				lo := &cpb.LearningObjective{
					Info: &cpb.ContentBasicInfo{
						Id:           loID,
						Name:         fmt.Sprintf("lo-name-%s", loID),
						Country:      cpb.Country_COUNTRY_VN,
						Subject:      cpb.Subject_SUBJECT_BIOLOGY,
						DisplayOrder: int32(i),
						SchoolId:     constants.ManabieSchool,
					},
					Type:    cpb.LearningObjectiveType(int32(g)),
					TopicId: topicID,
				}
				los = append(los, lo)
				studyPlanItemInfos = append(studyPlanItemInfos, &StudyPlanItemInfo{
					BookID:    stepState.BookID,
					ChapterID: stepState.ChapterID,
					TopicID:   topicID,
					LoID:      lo.Info.Id,
				})
			}
			assignment := &pb.Assignment{}
			ctx, assignment = s.generateAssignment(ctx, idutil.ULIDNow(), true, true, true)
			assignment.Content.TopicId = topicID
			assignment.DisplayOrder = 4
			assignments = append(assignments, assignment)
			studyPlanItemInfos = append(studyPlanItemInfos, &StudyPlanItemInfo{
				BookID:       stepState.BookID,
				ChapterID:    stepState.ChapterID,
				TopicID:      topicID,
				AssignmentID: assignment.AssignmentId,
			})
		}
	}
	if ctx, err := s.aSignedIn(ctx, "school admin"); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if _, err := pb.NewTopicModifierServiceClient(s.Conn).Upsert(contextWithToken(s, ctx), &pb.UpsertTopicsRequest{
		Topics: topics,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topics: %w", err)
	}
	if _, err := pb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(contextWithToken(s, ctx), &pb.UpsertLOsRequest{
		LearningObjectives: los,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create los: %w", err)
	}
	if _, err := pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(contextWithToken(s, ctx), &pb.UpsertAssignmentsRequest{
		Assignments: assignments,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create assignments: %w", err)
	}
	if ctx, err := s.addBookToCourse(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudyPlanItemInfos = studyPlanItemInfos
	stepState.TopicIDs = topicIDs
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userAddAValidBookDoesNotHaveAnyToCourse(ctx context.Context, loType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.createABook(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	var (
		topics             []*pb.Topic
		los                []*cpb.LearningObjective
		assignments        []*pb.Assignment
		studyPlanItemInfos []*StudyPlanItemInfo
	)
	n := rand.Intn(2) + 1
	for i := 1; i <= n; i++ {
		if ctx, err := s.createAChapter(ctx, i); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		m := rand.Intn(2) + 1
		for j := 1; j <= m; j++ {
			topicID := idutil.ULIDNow()
			topic := &pb.Topic{
				Id:           topicID,
				Name:         fmt.Sprintf("topic-name-%s", topicID),
				Country:      pb.Country_COUNTRY_VN,
				Subject:      pb.Subject_SUBJECT_BIOLOGY,
				DisplayOrder: int32(i),
				SchoolId:     constants.ManabieSchool,

				ChapterId: stepState.ChapterID,
			}
			topics = append(topics, topic)
			if loType == AssignmentType {
				for g := 1; g <= 2; g++ {
					loID := idutil.ULIDNow()
					lo := &cpb.LearningObjective{
						Info: &cpb.ContentBasicInfo{
							Id:           loID,
							Name:         fmt.Sprintf("lo-name-%s", loID),
							Country:      cpb.Country_COUNTRY_VN,
							Subject:      cpb.Subject_SUBJECT_BIOLOGY,
							DisplayOrder: int32(i),
							SchoolId:     constants.ManabieSchool,
						},
						TopicId: topicID,
						Type:    cpb.LearningObjectiveType(int32(g)),
					}
					los = append(los, lo)
					studyPlanItemInfos = append(studyPlanItemInfos, &StudyPlanItemInfo{
						BookID:    stepState.BookID,
						ChapterID: stepState.ChapterID,
						TopicID:   topicID,
						LoID:      lo.Info.Id,
					})
				}
			}

			if loType == LearningObjectiveType {
				assignment := &pb.Assignment{}
				ctx, assignment = s.generateAssignment(ctx, idutil.ULIDNow(), true, true, true)
				assignment.Content.TopicId = topicID
				assignment.DisplayOrder = 1
				assignments = append(assignments, assignment)
				studyPlanItemInfos = append(studyPlanItemInfos, &StudyPlanItemInfo{
					BookID:       stepState.BookID,
					ChapterID:    stepState.ChapterID,
					TopicID:      topicID,
					AssignmentID: assignment.AssignmentId,
				})
			}
		}
	}
	if _, err := pb.NewTopicModifierServiceClient(s.Conn).Upsert(contextWithToken(s, ctx), &pb.UpsertTopicsRequest{
		Topics: topics,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create topics: %w", err)
	}
	switch loType {
	case LearningObjectiveType:
		if _, err := pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(contextWithToken(s, ctx), &pb.UpsertAssignmentsRequest{
			Assignments: assignments,
		}); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create assignments: %w", err)
		}
	case AssignmentType:
		if _, err := pb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(contextWithToken(s, ctx), &pb.UpsertLOsRequest{
			LearningObjectives: los,
		}); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create los: %w", err)
		}
	}
	stepState.StudyPlanItemInfos = studyPlanItemInfos
	if ctx, err := s.addBookToCourse(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addBookToCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()
	cbe := &entities.CoursesBooks{}
	database.AllNullEntity(cbe)
	if err := multierr.Combine(
		cbe.BookID.Set(stepState.BookID),
		cbe.CourseID.Set(stepState.CourseID),
		cbe.CreatedAt.Set(now),
		cbe.UpdatedAt.Set(now),
	); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to set value for course book: %w", err)
	}
	if _, err := database.Insert(ctx, cbe, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to crete course book: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}
