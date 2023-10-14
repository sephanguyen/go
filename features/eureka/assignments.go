package eureka

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	bob_repository "github.com/manabie-com/backend/internal/bob/repositories"
	consta "github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func (s *suite) ensureStudentIsCreated(ctx context.Context, studentID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.SchoolID = strconv.Itoa(constants.ManabieSchool)
	ctx = s.setFakeClaimToContext(ctx, stepState.SchoolID, cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String())

	studentRepo := bob_repository.StudentRepo{}
	students, err := studentRepo.Retrieve(ctx, s.BobDB, database.TextArray([]string{studentID}))
	if len(students) == 0 || err != nil {
		backupStudentID := stepState.CurrentStudentID
		ctx, err := s.aValidStudentInDB(ctx, studentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("ensureStudentIsCreated.aValidStudentInDB %w", err)
		}
		stepState.CurrentStudentID = backupStudentID
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someStudentsAreAssignedSomeValidStudyPlans(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.SchoolID = strconv.Itoa(constants.ManabieSchool)

	ctx = s.setFakeClaimToContext(ctx, stepState.SchoolID, consta.RoleSchoolAdmin)
	if stepState.StudentsCourseMap == nil {
		stepState.StudentsCourseMap = make(map[string]string)
	}
	ctx, err1 := s.aValidCourseAndStudyPlanBackground(ctx)
	if err1 != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.aValidCourseAndStudyPlanBackground: %w", err1)
	}
	ctx, err2 := s.userAssignCourseWithStudyPlan(ctx)
	if err2 != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.userAssignCourseWithStudyPlan: %w", err2)
	}

	for _, st := range stepState.StudentIDs {
		if _, ok := stepState.StudentsCourseMap[st]; !ok {
			stepState.StudentsCourseMap[st] = stepState.CourseID
		}
		if ctx, err := s.ensureStudentIsCreated(ctx, st); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("ensureStudentIsCreated error: %w", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsListAvailableContents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.SchoolID = strconv.Itoa(constants.ManabieSchool)

	for _, studentID := range stepState.StudentIDs {
		token, err := s.generateExchangeToken(studentID, entities.UserGroupStudent)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.AuthToken = token
		resp, err := pb.NewAssignmentReaderServiceClient(s.Conn).ListStudentAvailableContents(contextWithToken(s, ctx), &pb.ListStudentAvailableContentsRequest{
			StudyPlanId: []string{},
			BookId:      stepState.BookID,
			ChapterId:   stepState.ChapterID,
			TopicId:     stepState.TopicID,
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Contents = append(stepState.Contents, resp.Contents)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkContentType(ctx context.Context, studyPlanItemID string, contentType pb.ContentType) (context.Context, bool, error) {
	stepState := StepStateFromContext(ctx)

	assignmentQueryCheck := `
			SELECT
				count(*)
			FROM study_plan_items sti
			JOIN assignment_study_plan_items asti ON
				sti.study_plan_item_id = asti.study_plan_item_id
			WHERE
				sti.study_plan_id = $1
	`
	loQueryCheck := `
			SELECT
				count(*)
			FROM study_plan_items sti
			JOIN lo_study_plan_items losti ON
				sti.study_plan_item_id = losti.study_plan_item_id
			WHERE
				losti.study_plan_id = $1
	`
	var query string
	if contentType == pb.ContentType_CONTENT_TYPE_ASSIGNMENT {
		query = assignmentQueryCheck
	}
	if contentType == pb.ContentType_CONTENT_TYPE_LO {
		query = loQueryCheck
	}
	var count int
	if err := db.QueryRow(ctx, query, studyPlanItemID).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), false, err
	}

	return StepStateToContext(ctx, stepState), true, nil
}

func (s *suite) returnsAListOfStudyPlanItemsContent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.Contents) == 0 {
		return StepStateToContext(ctx, stepState), errors.New("got empty contents")
	}

	ctx = auth.InjectFakeJwtToken(ctx, stepState.SchoolID)
	for i, studentID := range stepState.StudentIDs {
		contents := stepState.Contents[i]
		if len(contents) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("got empty contents for student id: %q", studentID)
		}

		ctx, studyPlanItemIDs, err := s.getStudyPlanItemsByStudent(ctx, studentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		expectedIDs := make([]string, 0, len(studyPlanItemIDs))
		for _, v := range studyPlanItemIDs {
			expectedIDs = append(expectedIDs, v.ID.String)
		}

		for _, content := range contents {
			id := content.StudyPlanItem.StudyPlanItemId
			if !golibs.InArrayString(id, expectedIDs) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study plan item id: %q in list %v of student: %q", id, studyPlanItemIDs, studentID)
			}
			ctx, valid, err := s.checkContentType(ctx, id, content.Type)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("s.checkContentType: %w", err)
			}
			if !valid {
				return StepStateToContext(ctx, stepState), fmt.Errorf("invalid resource id :%s with type: %s", id, content.Type.String())
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getStudyPlanItemsByStudent(ctx context.Context, studentID string) (context.Context, []*entities.StudyPlanItem, error) {
	stepState := StepStateFromContext(ctx)
	query := `
		SELECT study_plan_item_id, start_date, end_date, available_from, available_to, completed_at, display_order FROM study_plan_items spi
		INNER JOIN student_study_plans ssp ON spi.study_plan_id = ssp.study_plan_id
		WHERE student_id = $1
	`

	rows, err := s.DB.Query(ctx, query, studentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	var items []*entities.StudyPlanItem
	for rows.Next() {
		e := new(entities.StudyPlanItem)
		if err := rows.Scan(&e.ID, &e.StartDate, &e.EndDate, &e.AvailableFrom, &e.AvailableTo, &e.CompletedAt, &e.DisplayOrder); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		items = append(items, e)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}

	return StepStateToContext(ctx, stepState), items, nil
}

func (s *suite) getStudyPlansByStudent(ctx context.Context, studentID string) (context.Context, []string, error) {
	stepState := StepStateFromContext(ctx)
	query := `
		SELECT study_plan_id FROM student_study_plans
		WHERE student_id = $1
	`

	rows, err := s.DB.Query(ctx, query, studentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}

	return StepStateToContext(ctx, stepState), ids, nil
}

func (s *suite) listPaginatedStudentToDoItems(ctx context.Context, courseID, studentID string, status pb.ToDoStatus) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	token, err := s.generateExchangeToken(studentID, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token

	paging := &cpb.Paging{
		Limit: uint32(rand.Intn(2)) + 1,
	}
	var items [][]*pb.ToDoItem
	for {
		resp, err := pb.NewAssignmentReaderServiceClient(s.Conn).ListStudentToDoItems(contextWithToken(s, ctx), &pb.ListStudentToDoItemsRequest{
			StudentId: studentID,
			CourseIds: []string{courseID},
			Status:    status,
			Paging:    paging,
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if len(resp.Items) == 0 {
			break
		}
		if len(resp.Items) > int(paging.Limit) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected total study plan items: got: %d, want: %d", len(resp.Items), paging.Limit)
		}

		items = append(items, resp.Items)

		paging = resp.NextPage
	}

	stepState.PaginatedToDoItems = append(stepState.PaginatedToDoItems, items)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentListActiveStudyPlanItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, studentID := range stepState.StudentIDs {
		stepState.RequestSentAt = time.Now()

		if ctx, err := s.listPaginatedStudentToDoItems(ctx, stepState.StudentsCourseMap[studentID], studentID, pb.ToDoStatus_TO_DO_STATUS_ACTIVE); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsAListOfActiveStudyPlanItems(ctx context.Context, includeComplete string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now().UTC()

	for i, studentID := range stepState.StudentIDs {
		ctx, allItems, err := s.getStudyPlanItemsByStudent(ctx, studentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		isActiveItem := func(item *entities.StudyPlanItem) bool {
			return item.StartDate.Status != pgtype.Null &&
				item.StartDate.Time.Before(now) &&
				item.EndDate.Time.UTC().After(stepState.RequestSentAt) &&
				item.AvailableFrom.Time.UTC().Before(now) &&
				(item.AvailableTo.Time.UTC().After(stepState.RequestSentAt) || item.AvailableTo.Status == pgtype.Null)
		}

		var activeIDs []string
		var activeItems []*entities.StudyPlanItem
		for _, item := range allItems {
			if isActiveItem(item) {
				activeIDs = append(activeIDs, item.ID.String)
				activeItems = append(activeItems, item)
			}
		}

		items := stepState.PaginatedToDoItems[i]
		if len(items) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("got empty to do items")
		}

		var total int
		for _, paginatedItems := range items {
			for _, pi := range paginatedItems {
				if pi.StudyPlanItem.StartDate == nil {
					continue
				}

				if !golibs.InArrayString(pi.StudyPlanItem.StudyPlanItemId, activeIDs) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study plan item id: %q in list study plan items %v of student: %q", pi.StudyPlanItem.StudyPlanItemId, activeIDs, studentID)
				}
				if t := pi.StudyPlanItem.StartDate.AsTime(); !t.Before(now) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("start_date must be before current time: got: %v, current time: %v", t, now)
				}
				if t := pi.StudyPlanItem.EndDate.AsTime(); !t.After(stepState.RequestSentAt) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("end_date must be after current time: got: %v, current time: %v", t, now)
				}
				if includeComplete != "all" {
					if t := pi.StudyPlanItem.CompletedAt; t != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("completed_at must be zero, got: %v", t)
					}
				}

				total++
			}

			// check order by (start_date ASC, display_order ASC, study_plan_item_id ASC)
			for i := 1; i < len(paginatedItems); i++ {
				prevItem := paginatedItems[i-1].StudyPlanItem
				item := paginatedItems[i].StudyPlanItem

				prevStartDate := prevItem.StartDate.AsTime()
				prevDisplayOrder := prevItem.DisplayOrder
				prevStudyPlanItemID := prevItem.StudyPlanItemId

				startDate := item.StartDate.AsTime()
				displayOrder := item.DisplayOrder
				studyPlanItemID := item.StudyPlanItemId

				if prevStartDate.Before(startDate) {
					continue
				}
				if prevStartDate.After(startDate) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("start_date of study_plan_item %v (%v) must be before %v (%v)",
						prevStudyPlanItemID, prevStartDate, studyPlanItemID, startDate)
				}

				if prevDisplayOrder < displayOrder {
					continue
				}
				if prevDisplayOrder > displayOrder {
					return StepStateToContext(ctx, stepState), fmt.Errorf("display_order of study_plan_item %v (%v) must be less than %v (%v)",
						prevStudyPlanItemID, prevDisplayOrder, studyPlanItemID, displayOrder)
				}

				if prevStudyPlanItemID > studyPlanItemID {
					return StepStateToContext(ctx, stepState), fmt.Errorf("study_plan_item_id of study_plan_item %v (%v) must be less than %v (%v)",
						prevStudyPlanItemID, prevStudyPlanItemID, studyPlanItemID, studyPlanItemID)
				}
			}
		}

		if total != len(activeItems) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("active study plan items mismatch of student %q, got: %d, want: %d", studentID, total, len(activeItems))
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentListCompletedStudyPlanItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, studentID := range stepState.StudentIDs {
		if ctx, err := s.listPaginatedStudentToDoItems(ctx, stepState.CourseID, studentID, pb.ToDoStatus_TO_DO_STATUS_COMPLETED); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsAListOfCompletedStudyPlanItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// only student at index 0 submit assignment
	studentID := stepState.StudentIDs[0]
	ctx, allItems, err := s.getStudyPlanItemsByStudent(ctx, studentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	isCompletedItem := func(item *entities.StudyPlanItem) bool {
		return item.CompletedAt.Status == pgtype.Present
	}

	var completedIDs []string
	var completedItems []*entities.StudyPlanItem
	for _, item := range allItems {
		if isCompletedItem(item) {
			completedIDs = append(completedIDs, item.ID.String)
			completedItems = append(completedItems, item)
		}
	}

	items := stepState.PaginatedToDoItems[0]
	var total int
	for _, paginatedItems := range items {
		for _, pi := range paginatedItems {
			if !golibs.InArrayString(pi.StudyPlanItem.StudyPlanItemId, completedIDs) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study plan id: %q in list study plans %v of student: %q", pi.StudyPlanItem.StudyPlanId, completedIDs, studentID)
			}
			if t := pi.StudyPlanItem.CompletedAt; t == nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("completed_at must not be zero, got: %v", t)
			}
		}

		// check order by (start_date DESC, display_order ASC, study_plan_item_id DESC)
		for i := 1; i < len(paginatedItems); i++ {
			prevItem := paginatedItems[i-1].StudyPlanItem
			item := paginatedItems[i].StudyPlanItem

			prevStartDate := prevItem.StartDate.AsTime()
			prevDisplayOrder := prevItem.DisplayOrder
			prevStudyPlanItemID := prevItem.StudyPlanItemId

			startDate := item.StartDate.AsTime()
			displayOrder := item.DisplayOrder
			studyPlanItemID := item.StudyPlanItemId

			if prevStartDate.After(startDate) {
				continue
			}
			if prevStartDate.Before(startDate) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("start_date of study_plan_item %v (%v) must be after %v (%v)",
					prevStudyPlanItemID, prevStartDate, studyPlanItemID, startDate)
			}

			if prevDisplayOrder < displayOrder {
				continue
			}
			if prevDisplayOrder > displayOrder {
				return StepStateToContext(ctx, stepState), fmt.Errorf("display_order of study_plan_item %v (%v) must be less than %v (%v)",
					prevStudyPlanItemID, prevDisplayOrder, studyPlanItemID, displayOrder)
			}

			if prevStudyPlanItemID < studyPlanItemID {
				return StepStateToContext(ctx, stepState), fmt.Errorf("study_plan_item_id of study_plan_item %v (%v) must be greater than %v (%v)",
					prevStudyPlanItemID, prevStudyPlanItemID, studyPlanItemID, studyPlanItemID)
			}
		}

		total += len(paginatedItems)
	}

	if total != len(completedItems) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("completed study plan items mismatch of student %q, got: %d, want: %d", studentID, total, len(completedItems))
	}

	// other students must have zero completed_at timestamp
	for i := 1; i < len(stepState.StudentIDs); i++ {
		studentID := stepState.StudentIDs[i]

		ctx, allItems, err := s.getStudyPlanItemsByStudent(ctx, studentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		var completedIDs []string
		var completedItems []*entities.StudyPlanItem
		for _, item := range allItems {
			if isCompletedItem(item) {
				completedIDs = append(completedIDs, item.ID.String)
				completedItems = append(completedItems, item)
			}
		}

		items := stepState.PaginatedToDoItems[i]
		var total int
		for _, paginatedItems := range items {
			for _, pi := range paginatedItems {
				if !golibs.InArrayString(pi.StudyPlanItem.StudyPlanItemId, completedIDs) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study plan id: %q in list study plans %v of student: %q", pi.StudyPlanItem.StudyPlanId, completedIDs, studentID)
				}
				if t := pi.StudyPlanItem.CompletedAt; t != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("completed_at must be zero, got: %v", t)
				}
			}

			total += len(paginatedItems)
		}

		if total != len(completedItems) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("completed study plan items mismatch of student %q, got: %d, want: %d", studentID, total, len(completedItems))
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentHaventCompletedAnyStudyPlanItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// study plan item has end_date is 5 seconds from now
	// See s.generateStudyPlanItem method
	time.Sleep(5 * time.Second)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentListOverdueStudyPlanItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, studentID := range stepState.StudentIDs {
		if ctx, err := s.listPaginatedStudentToDoItems(ctx, stepState.CourseID, studentID, pb.ToDoStatus_TO_DO_STATUS_OVERDUE); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsAListOfOverdueStudyPlanItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now().UTC()

	for i, studentID := range stepState.StudentIDs[:1] {
		ctx, allItems, err := s.getStudyPlanItemsByStudent(ctx, studentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		isOverDueItem := func(item *entities.StudyPlanItem) bool {
			return item.StartDate.Status != pgtype.Null &&
				item.StartDate.Time.Before(now) &&
				item.EndDate.Time.UTC().Before(now) &&
				item.AvailableFrom.Time.UTC().Before(now) &&
				item.AvailableTo.Time.UTC().After(now)
		}

		var overdueIDs []string
		var overdueItems []*entities.StudyPlanItem
		for _, item := range allItems {
			if isOverDueItem(item) {
				overdueIDs = append(overdueIDs, item.ID.String)
				overdueItems = append(overdueItems, item)
			}
		}

		items := stepState.PaginatedToDoItems[i]
		var total int
		for _, paginatedItems := range items {
			for _, pi := range paginatedItems {
				if !golibs.InArrayString(pi.StudyPlanItem.StudyPlanItemId, overdueIDs) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study plan item id: %q in list study plan items %v of student: %q", pi.StudyPlanItem.StudyPlanItemId, overdueIDs, studentID)
				}
				if t := pi.StudyPlanItem.StartDate.AsTime(); !t.Before(now) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("start_date must be before current time: got: %v, current time: %v", t, now)
				}
				if t := pi.StudyPlanItem.EndDate.AsTime(); t.After(now) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("end_date must be before current time: got: %v, current time: %v", t, now)
				}
				if t := pi.StudyPlanItem.CompletedAt; t != nil {
					return StepStateToContext(ctx, stepState), fmt.Errorf("completed_at must be zero, got: %v", t)
				}

				total++
			}
		}

		if total != len(overdueItems) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("assigned study plan items mismatch of student %q, got: %d, want: %d", studentID, total, len(overdueItems))
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsAListOfUpcomingStudyPlanItems(ctx context.Context, includeComplete string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now().UTC()

	for i, studentID := range stepState.StudentIDs {
		ctx, allItems, err := s.getStudyPlanItemsByStudent(ctx, studentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		isUpcomingItem := func(item *entities.StudyPlanItem) bool {
			return item.StartDate.Status == pgtype.Null || item.StartDate.Time.UTC().After(now) &&
				item.AvailableFrom.Time.UTC().Before(now) &&
				item.AvailableTo.Time.UTC().After(now)
		}

		var upcomingIDs []string
		var upcomingItems []*entities.StudyPlanItem
		for _, item := range allItems {
			if isUpcomingItem(item) {
				upcomingIDs = append(upcomingIDs, item.ID.String)
				upcomingItems = append(upcomingItems, item)
			}
		}

		items := stepState.PaginatedToDoItems[i]
		if len(items) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("got empty to do items")
		}
		var total int
		for _, paginatedItems := range items {
			for _, pi := range paginatedItems {
				if !golibs.InArrayString(pi.StudyPlanItem.StudyPlanItemId, upcomingIDs) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study plan item id: %q in list study plan items %v of student: %q", pi.StudyPlanItem.StudyPlanId, upcomingIDs, studentID)
				}
				if pi.StudyPlanItem.StartDate != nil {
					if t := pi.StudyPlanItem.StartDate.AsTime(); !t.After(now) {
						return StepStateToContext(ctx, stepState), fmt.Errorf("start_date must be after current time: got: %v, current time: %v", t, now)
					}
				}
				if t := pi.StudyPlanItem.EndDate.AsTime(); !t.After(now) {
					return StepStateToContext(ctx, stepState), fmt.Errorf("end_date must be after current time: got: %v, current time: %v", t, now)
				}
				if includeComplete != "all" {
					if t := pi.StudyPlanItem.CompletedAt; t != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("completed_at must be zero, got: %v", t)
					}
				}
			}

			total += len(paginatedItems)
		}

		if total != len(upcomingItems) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("upcoming study plan items mismatch of student %q, got: %d, want: %d", studentID, total, len(upcomingItems))
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsListAvailableContentsWithIncorrectFilters(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, studentID := range stepState.StudentIDs {
		token, err := s.generateExchangeToken(studentID, entities.UserGroupStudent)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.AuthToken = token
		resp, err := pb.NewAssignmentReaderServiceClient(s.Conn).ListStudentAvailableContents(contextWithToken(s, ctx), &pb.ListStudentAvailableContentsRequest{
			StudyPlanId: []string{"invalid-study-plan-ids"},
			BookId:      "invalid-book",
			ChapterId:   "invalid-chapter-id",
			TopicId:     "invalid-topic-id",
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Contents = append(stepState.Contents, resp.Contents)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) retrieveTheirSubmissionGrade(ctx context.Context, actor string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	token, err := s.generateExchangeToken(stepState.StudentIDs[0], entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.AuthToken = token
	s.retrieveTheirOwnSubmissions(ctx, "student")
	rsp := stepState.Response.(*pb.RetrieveSubmissionsResponse)
	gradeID := rsp.Items[0].SubmissionGradeId
	stepState.Response, stepState.ResponseErr = pb.NewStudentAssignmentReaderServiceClient(s.Conn).
		RetrieveSubmissionGrades(contextWithToken(s, ctx), &pb.RetrieveSubmissionGradesRequest{
			SubmissionGradeIds: []string{gradeID.Value},
		})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentListUpcomingStudyPlanItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, studentID := range stepState.StudentIDs {
		if ctx, err := s.listPaginatedStudentToDoItems(ctx, stepState.CourseID, studentID, pb.ToDoStatus_TO_DO_STATUS_UPCOMING); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someStudentsOfDifferentCoursesAreAssignedSomeValidStudyPlans(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.StudentsCourseMap == nil {
		stepState.StudentsCourseMap = make(map[string]string)
	}
	n := 2
	for i := 0; i < n; i++ {
		ctx, err1 := s.aValidCourseAndStudyPlanBackground(ctx)
		ctx, err2 := s.userAssignCourseWithStudyPlan(ctx)
		if err := multierr.Combine(
			err1,
			err2,
		); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		for _, st := range stepState.StudentIDs {
			if _, ok := stepState.StudentsCourseMap[st]; !ok {
				stepState.StudentsCourseMap[st] = stepState.CourseID
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsListAvailableContentsWithCourseId(ctx context.Context, flag string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if flag == "invalid" {
		stepState.CourseID = "invalid-course-id"
	}
	for _, studentID := range stepState.StudentIDs {
		token, err := s.generateExchangeToken(studentID, entities.UserGroupStudent)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.AuthToken = token
		resp, err := pb.NewAssignmentReaderServiceClient(s.Conn).ListStudentAvailableContents(contextWithToken(s, ctx), &pb.ListStudentAvailableContentsRequest{
			StudyPlanId: []string{},
			BookId:      stepState.BookID,
			ChapterId:   stepState.ChapterID,
			TopicId:     stepState.TopicID,
			CourseId:    stepState.CourseID,
		})
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.Contents = append(stepState.Contents, resp.Contents)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsAListOfEmptyStudyPlanItemsContent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, content := range stepState.Contents {
		if len(content) != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("eureka does not return StepStateToContext(ctx, stepState), empty study plan items")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) allStudyPlanItemHasEmptyStartDate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `UPDATE study_plan_items SET start_date = null where study_plan_item_id = ANY(SELECT study_plan_item_id FROM study_plan_items si JOIN study_plans s ON si.study_plan_id = s.study_plan_id
		WHERE s.master_study_plan_id=$1 )`
	_, err := s.DB.Exec(ctx, query, &stepState.StudyPlanID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
