package study_plan

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// nolint

var (
	active    = "active"
	overdue   = "overdue"
	completed = "completed"
)

func (s *Suite) validCourseAndStudyPlanInDB(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	var err error

	stepState.CourseID, err = utils.GenerateCourse(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.YasuoConn)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't generate course: %v", err)
	}

	stepState.NumberOfStudyPlan = rand.Intn(3) + 2
	studyPlanIDs, err := utils.GenerateStudyPlans(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, stepState.CourseID, stepState.BookID, stepState.NumberOfStudyPlan)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't generate study plan: %v", err)
	}

	stepState.StudyPlanIDs = studyPlanIDs

	courseStudents, err := utils.AValidCourseWithIDs(ctx, s.EurekaDB, []string{stepState.StudentID}, stepState.CourseID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidCourseWithIDs: %w", err)
	}

	if err := utils.GenerateCourseBooks(s.AuthHelper.SignedCtx(ctx, stepState.SchoolAdminToken), stepState.CourseID, []string{stepState.BookID}, s.EurekaConn); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateCourseBooks: %w", err)
	}
	m := rand.Int31n(3) + 1
	locationIDs := make([]string, 0, m)
	for i := int32(1); i <= m; i++ {
		id := idutil.ULIDNow()
		locationIDs = append(locationIDs, id)
	}
	for _, courseStudent := range courseStudents {
		for _, locationID := range locationIDs {
			now := time.Now()
			e := &entities.CourseStudentsAccessPath{}
			database.AllNullEntity(e)
			if err := multierr.Combine(
				e.CourseStudentID.Set(courseStudent.ID.String),
				e.CourseID.Set(courseStudent.CourseID.String),
				e.StudentID.Set(courseStudent.StudentID.String),
				e.LocationID.Set(locationID),
				e.CreatedAt.Set(now),
				e.UpdatedAt.Set(now),
			); err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
			}
			if _, err := database.Insert(ctx, e, s.EurekaDB.Exec); err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("database.Insert: %w", err)
			}
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminAssignStudyPlanToAStudent(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	for _, studyPlanID := range stepState.StudyPlanIDs {
		err := utils.UserAssignStudyPlanToAStudent(s.AuthHelper.SignedCtx(ctx, stepState.Token), s.EurekaConn, stepState.StudentID, studyPlanID)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("cannot assign study plan to student")
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminInsertIndividualStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	for i := 0; i < len(stepState.StudyPlanIDs); i++ {
		for _, loID := range stepState.LoIDs {
			req := generateIndividualStudyPlanRequest(stepState.StudyPlanIDs[i], loID, stepState.StudentID)
			_, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).UpsertIndividual(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
			if stepState.ResponseErr != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert individual study plan to student:%s ", stepState.ResponseErr.Error())
			}
		}
	}

	// generate overdue individual study plan with lastStudyPlanID
	for _, loID := range stepState.LoIDs {
		req := generateOverdueIndividualStudyPlanRequest(stepState.StudyPlanIDs[len(stepState.StudyPlanIDs)-1], loID, stepState.StudentID)
		_, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).UpsertIndividual(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
		if stepState.ResponseErr != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("can't insert over due individual study plan to student:%s ", stepState.ResponseErr.Error())
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
func generateIndividualStudyPlanRequest(spID, lmID, studentID string) *sspb.UpsertIndividualInfoRequest {
	req := &sspb.UpsertIndividualInfoRequest{
		IndividualItems: []*sspb.StudyPlanItem{
			{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId:        spID,
					LearningMaterialId: lmID,
					StudentId: &wrapperspb.StringValue{
						Value: studentID,
					},
				},
				AvailableFrom: timestamppb.New(time.Now().Add(-24 * time.Hour)),
				AvailableTo:   timestamppb.New(time.Now().AddDate(0, 0, 10)),
				StartDate:     timestamppb.New(time.Now().Add(-23 * time.Hour)),
				EndDate:       timestamppb.New(time.Now().AddDate(0, 0, 1)),
				Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
			},
		},
	}

	return req
}
func generateOverdueIndividualStudyPlanRequest(spID, lmID, studentID string) *sspb.UpsertIndividualInfoRequest {
	req := &sspb.UpsertIndividualInfoRequest{
		IndividualItems: []*sspb.StudyPlanItem{
			{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId:        spID,
					LearningMaterialId: lmID,
					StudentId: &wrapperspb.StringValue{
						Value: studentID,
					},
				},
				AvailableFrom: timestamppb.New(time.Now().Add(-24 * time.Hour)),
				AvailableTo:   timestamppb.New(time.Now().AddDate(0, 0, 10)),
				StartDate:     timestamppb.New(time.Now().Add(-23 * time.Hour)),
				EndDate:       timestamppb.New(time.Now().Add(-20 * time.Hour)),
				Status:        sspb.StudyPlanItemStatus_STUDY_PLAN_ITEM_STATUS_ACTIVE,
			},
		},
	}

	return req
}

func (s *Suite) listPaginatedStudentToDoItems(ctx context.Context, studentID string, status sspb.StudyPlanItemToDoStatus) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	paging := &cpb.Paging{
		Limit: uint32(2),
	}
	for {
		resp, err := sspb.NewStudyPlanClient(s.EurekaConn).ListToDoItem(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListToDoItemRequest{
			StudentId: studentID,
			CourseIds: []string{stepState.CourseID},
			Status:    status,
			Page:      paging,
		})
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
		if len(resp.TodoItems) == 0 {
			break
		}
		if len(resp.TodoItems) > int(paging.Limit) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total study plan items: got: %d, want: %d", len(resp.TodoItems), paging.Limit)
		}
		stepState.PaginatedToDoItems = append(stepState.PaginatedToDoItems, resp.TodoItems)
		paging = resp.NextPage
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentsIsAssignedSomeValidStudyPlans(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx, err := s.validCourseAndStudyPlanInDB(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.validCourseAndStudyPlanInDB: %w", err)
	}

	ctx, err = s.adminAssignStudyPlanToAStudent(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.adminAssignStudyPlanToAStudent: %w", err)
	}

	ctx, err = s.adminInsertIndividualStudyPlan(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.adminInsertIndividualStudyPlan: %w", err)
	}

	ctx, err = s.aValidAssignment(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.aValidAssignment: %w", err)
	}

	ctx, err = s.studentSubmitAssignment(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.studentSubmitAssignment: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userListToDoItems(ctx context.Context, arg string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Status = arg
	switch arg {
	case active:
		stepState.Status = "active"
		stepState.RequestSentAt = time.Now()
		if ctx, err := s.listPaginatedStudentToDoItems(ctx, stepState.StudentID, sspb.StudyPlanItemToDoStatus_STUDY_PLAN_ITEM_TO_DO_STATUS_ACTIVE); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	case completed:
		stepState.Status = "completed"
		if ctx, err := s.listPaginatedStudentToDoItems(ctx, stepState.StudentID, sspb.StudyPlanItemToDoStatus_STUDY_PLAN_ITEM_TO_DO_STATUS_COMPLETED); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	case overdue:
		stepState.Status = "overdue"
		if ctx, err := s.listPaginatedStudentToDoItems(ctx, stepState.StudentID, sspb.StudyPlanItemToDoStatus_STUDY_PLAN_ITEM_TO_DO_STATUS_OVERDUE); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	default:
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("invalid status")
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) schoolAdminAndStudentLogin(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	ctx, err := s.aSignedIn(ctx, "student")
	stepState.StudentID = stepState.Student.ID
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.StudentToken = stepState.Token
	ctx, err = s.aSignedIn(ctx, "school admin")
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.SchoolAdminToken = stepState.Token
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnToDoItemsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	switch stepState.Status {
	case active:
		if ctx, err := s.ourSystemReturnActiveIndividualStudyPlan(ctx); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	case completed:
		if ctx, err := s.ourSystemReturnCompletedIndividualStudyPlan(ctx); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	case overdue:
		if ctx, err := s.ourSystemReturnOverdueIndividualStudyPlan(ctx); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnActiveIndividualStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	now := time.Now().UTC()
	ctx, allItems, err := s.getIndividualStudyPlanByStudent(ctx, stepState.StudentID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	isActiveItem := func(item *repositories.IndividualStudyPlanItem) bool {
		return item.CompletedAt.Status == pgtype.Null &&
			item.StartDate.Status != pgtype.Null &&
			item.StartDate.Time.Before(now) &&
			item.EndDate.Time.UTC().After(stepState.RequestSentAt) &&
			item.AvailableFrom.Time.UTC().Before(now) &&
			(item.AvailableTo.Time.UTC().After(stepState.RequestSentAt) || item.AvailableTo.Status == pgtype.Null)
	}
	var activeIDs []string
	var activeItems []*repositories.IndividualStudyPlanItem
	for _, item := range allItems {
		if isActiveItem(item) {
			activeIDs = append(activeIDs, item.StudyPlanID.String)
			activeItems = append(activeItems, item)
		}
	}
	items := stepState.PaginatedToDoItems
	if len(items) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("got empty active to do items")
	}
	var total int
	for _, paginatedItems := range items {
		for _, pi := range paginatedItems {
			if pi.IndividualStudyPlanItem.StartDate == nil {
				continue
			}
			if pi.LearningMaterialType != 0 {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning material type %d", pi.LearningMaterialType)
			}
			if !golibs.InArrayString(pi.IndividualStudyPlanItem.StudyPlanItemIdentity.StudyPlanId, activeIDs) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study plan item id: %q in list study plan items %v of student: %q", pi.IndividualStudyPlanItem.StudyPlanItemIdentity.StudyPlanId, activeIDs, stepState.StudentID)
			}
			if t := pi.IndividualStudyPlanItem.StartDate.AsTime(); !t.Before(now) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("start_date must be before current time: got: %v, current time: %v", t, now)
			}
			if t := pi.IndividualStudyPlanItem.EndDate.AsTime(); !t.After(stepState.RequestSentAt) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("end_date must be after current time: got: %v, current time: %v", t, now)
			}

			total++
		}

		// check order by (start_date ASC, learning_material_id ASC)
		for i := 1; i < len(paginatedItems); i++ {
			prevItem := paginatedItems[i-1].IndividualStudyPlanItem
			item := paginatedItems[i].IndividualStudyPlanItem

			prevStartDate := prevItem.StartDate.AsTime()
			prevLearningMaterialID := prevItem.StudyPlanItemIdentity.LearningMaterialId

			startDate := item.StartDate.AsTime()
			learningMaterialID := item.StudyPlanItemIdentity.LearningMaterialId

			if prevStartDate.Before(startDate) {
				continue
			}
			if prevStartDate.After(startDate) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("start_date of individual study plan items %v (%v) must be before %v (%v)",
					prevLearningMaterialID, prevStartDate, learningMaterialID, startDate)
			}

			if prevLearningMaterialID > learningMaterialID {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("learning_material_id of  individual study plan items %v (%v) must be less than %v (%v)",
					prevLearningMaterialID, prevLearningMaterialID, learningMaterialID, learningMaterialID)
			}
		}
	}
	if total != len(activeItems) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("active individual study plan items mismatch of student %q, got: %d, want: %d", stepState.StudentID, total, len(activeItems))
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnCompletedIndividualStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx, allItems, err := s.getIndividualStudyPlanByStudent(ctx, stepState.StudentID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	isCompletedItem := func(item *repositories.IndividualStudyPlanItem) bool {
		return item.CompletedAt.Status == pgtype.Present
	}
	var completedIDs []string
	var completedItems []*repositories.IndividualStudyPlanItem
	for _, item := range allItems {
		if isCompletedItem(item) {
			completedIDs = append(completedIDs, item.StudyPlanID.String)
			completedItems = append(completedItems, item)
		}
	}
	items := stepState.PaginatedToDoItems
	if len(items) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("got empty completed to do items")
	}

	var total int
	for _, paginatedItems := range items {
		for _, pi := range paginatedItems {
			if !golibs.InArrayString(pi.IndividualStudyPlanItem.StudyPlanItemIdentity.StudyPlanId, completedIDs) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected learning material id: %q in list study plans %v of student: %q", pi.IndividualStudyPlanItem.StudyPlanItemIdentity.StudyPlanId, completedIDs, stepState.StudentID)
			}
		}

		// check order by (start_date DESC, learning_material_id DESC)
		for i := 1; i < len(paginatedItems); i++ {
			prevItem := paginatedItems[i-1].IndividualStudyPlanItem
			item := paginatedItems[i].IndividualStudyPlanItem

			prevStartDate := prevItem.StartDate.AsTime()
			prevLearningMaterialID := prevItem.StudyPlanItemIdentity.LearningMaterialId

			startDate := item.StartDate.AsTime()
			learningMaterialID := item.StudyPlanItemIdentity.LearningMaterialId

			if prevStartDate.After(startDate) {
				continue
			}

			if prevStartDate.Before(startDate) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("start_date of learning_material_id %v (%v) must be after %v (%v)",
					prevLearningMaterialID, prevStartDate, learningMaterialID, startDate)
			}

			if prevLearningMaterialID > learningMaterialID {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("learning_material_id of individual_study_plan_item %v (%v) must be greater than %v (%v)",
					prevLearningMaterialID, prevLearningMaterialID, learningMaterialID, learningMaterialID)
			}
		}

		total += len(paginatedItems)
	}

	if total != len(completedItems) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("completed individual study plan items mismatch of student %q, got: %d, want: %d", stepState.StudentID, total, len(completedItems))
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnOverdueIndividualStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	now := time.Now().UTC()
	ctx, allItems, err := s.getIndividualStudyPlanByStudent(ctx, stepState.StudentID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	isOverDueItem := func(item *repositories.IndividualStudyPlanItem) bool {
		return item.StartDate.Status != pgtype.Null &&
			item.StartDate.Time.Before(now) &&
			item.EndDate.Time.UTC().Before(now) &&
			item.AvailableFrom.Time.UTC().Before(now) &&
			item.AvailableTo.Time.UTC().After(now)
	}

	var overdueIDs []string
	var overdueItems []*repositories.IndividualStudyPlanItem
	for _, item := range allItems {
		if isOverDueItem(item) {
			overdueIDs = append(overdueIDs, item.StudyPlanID.String)
			overdueItems = append(overdueItems, item)
		}
	}
	items := stepState.PaginatedToDoItems

	var total int
	for _, paginatedItems := range items {
		for _, pi := range paginatedItems {
			if !golibs.InArrayString(pi.IndividualStudyPlanItem.StudyPlanItemIdentity.StudyPlanId, overdueIDs) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study plan item id: %q in list study plan items %v of student: %q", pi.IndividualStudyPlanItem.StudyPlanItemIdentity.StudyPlanId, overdueIDs, stepState.StudentID)
			}
			if t := pi.IndividualStudyPlanItem.StartDate.AsTime(); !t.Before(now) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("start_date must be before current time: got: %v, current time: %v", t, now)
			}
			if t := pi.IndividualStudyPlanItem.EndDate.AsTime(); t.After(now) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("end_date must be before current time: got: %v, current time: %v", t, now)
			}

			total++
		}
	}

	if total != len(overdueItems) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("overdue study plan items mismatch of student %q, got: %d, want: %d", stepState.StudentID, total, len(overdueItems))
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) getIndividualStudyPlanByStudent(ctx context.Context, studentID string) (context.Context, []*repositories.IndividualStudyPlanItem, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	query := `
		SELECT distinct on (study_plan_id, learning_material_id, student_id)
		isp.study_plan_id,  isp.learning_material_id,  isp.student_id,  isp.status,  isp.start_date,  isp.end_date,  isp.available_from,  isp.available_to, isp.school_date, gsl.completed_at
		FROM list_available_learning_material() isp
		LEFT JOIN get_student_completion_learning_material() gsl using(student_id, study_plan_id, learning_material_id)
		WHERE isp.student_id = $1
	`
	rows, err := s.EurekaDB.Query(ctx, query, studentID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()
	var items []*repositories.IndividualStudyPlanItem
	for rows.Next() {
		e := new(repositories.IndividualStudyPlanItem)
		if err := rows.Scan(&e.StudyPlanID, &e.LearningMaterialID, &e.StudentID, &e.Status, &e.StartDate, &e.EndDate, &e.AvailableFrom, &e.AvailableTo, &e.SchoolDate, &e.CompletedAt); err != nil {
			return utils.StepStateToContext(ctx, stepState), nil, err
		}
		items = append(items, e)
	}
	return utils.StepStateToContext(ctx, stepState), items, nil
}

func (s *Suite) aValidAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.LearningMaterialIDs = append(stepState.LearningMaterialIDs, stepState.LoIDs[0:(len(stepState.LoIDs)/2)+1]...)
	assignmentResult, err := utils.GenerateAssignment(
		s.AuthHelper.SignedCtx(ctx, stepState.Student.Token),
		stepState.TopicIDs[0],
		len(stepState.LearningMaterialIDs),
		stepState.LoIDs,
		s.EurekaConn,
		nil,
	)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.GenerateAssignment: %w", err)
	}
	stepState.AssignmentIDs = assignmentResult.AssignmentIDs
	return utils.StepStateToContext(ctx, stepState), nil
}
func (s *Suite) studentSubmitAssignment(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	assignmentSubmissions := s.getAssignmentSubmission(ctx)
	for _, req := range assignmentSubmissions {
		stepState.Request = req
		stepState.Response, stepState.ResponseErr = sspb.NewAssignmentClient(s.EurekaConn).
			SubmitAssignment(s.AuthHelper.SignedCtx(ctx, stepState.Student.Token), req)
		if stepState.ResponseErr != nil {
			return utils.StepStateToContext(ctx, stepState), stepState.ResponseErr
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) getAssignmentSubmission(ctx context.Context) []*sspb.SubmitAssignmentRequest {
	stepState := utils.StepStateFromContext[StepState](ctx)
	assignmentSubmission := []*sspb.SubmitAssignmentRequest{}
	for _, learningMaterialID := range stepState.LearningMaterialIDs {
		req := &sspb.SubmitAssignmentRequest{
			Submission: &sspb.StudentSubmission{
				StudyPlanItemIdentity: &sspb.StudyPlanItemIdentity{
					StudyPlanId:        stepState.StudyPlanIDs[0],
					LearningMaterialId: learningMaterialID,
					StudentId:          wrapperspb.String(stepState.StudentID),
				},
				SubmissionContent:  []*sspb.SubmissionContent{},
				Note:               "submit",
				CompleteDate:       timestamppb.Now(),
				Duration:           int32(rand.Intn(99) + 1),
				CorrectScore:       wrapperspb.Float(rand.Float32() * 10),
				TotalScore:         wrapperspb.Float(rand.Float32() * 100),
				UnderstandingLevel: sspb.SubmissionUnderstandingLevel(rand.Intn(len(sspb.SubmissionUnderstandingLevel_value))),
			},
		}
		assignmentSubmission = append(assignmentSubmission, req)
	}

	return assignmentSubmission
}

func (s *Suite) userTryListToDoItems(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	paging := &cpb.Paging{
		Limit: uint32(100),
	}
	stepState.Response, stepState.ResponseErr = sspb.NewStudyPlanClient(s.EurekaConn).ListToDoItem(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListToDoItemRequest{
		StudentId: stepState.StudentID,
		CourseIds: []string{stepState.CourseID},
		Status:    sspb.StudyPlanItemToDoStatus_STUDY_PLAN_ITEM_TO_DO_STATUS_ACTIVE,
		Page:      paging,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}
