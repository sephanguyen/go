package eureka

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	eu_v1 "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) studyPlanItemsHaveCreatedOnAssignmentsCreatedCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	spie := &entities.StudyPlanItem{}
	fields, _ := spie.FieldMap()
	stmt := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE content_structure ->> 'assignment_id' = ANY($1)
	`, strings.Join(fields, ","), spie.TableName())
	var studyPlanItems []*entities.StudyPlanItem
	if err := try.Do(func(attempt int) (retry bool, err error) {
		var items []*entities.StudyPlanItem
		rows, err := s.DB.Query(ctx, stmt, stepState.AssignmentIDs)
		if err != nil {
			return false, err
		}
		defer rows.Close()
		for rows.Next() {
			var e entities.StudyPlanItem
			if err := rows.Scan(database.GetScanFields(&e, fields)...); err != nil {
				return false, err
			}
			items = append(items, &e)
		}
		if len(items) != len(stepState.LOs)*len(stepState.StudyPlanIDs) {
			time.Sleep(2 * time.Second)
			return attempt < 5, fmt.Errorf("study plan items created is wrong, expect %v but got %v", len(stepState.LOs)*len(stepState.StudyPlanIDs), len(items))
		}
		studyPlanItems = items
		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	getKey := func(content *entities.ContentStructure) string {
		return strings.Join([]string{
			content.CourseID,
			content.BookID,
			content.ChapterID,
			content.TopicID,
			content.AssignmentID,
		}, "|")
	}

	m := make(map[string]bool)

	for _, item := range studyPlanItems {
		var content *entities.ContentStructure
		item.ContentStructure.AssignTo(&content)
		m[getKey(content)] = true
	}

	for _, lo := range stepState.LOs {
		if ok := m[getKey(lo)]; !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("missing study plan item of content: %s", getKey(lo))
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreateSomeAssignmentsInBooks(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)
	n := 4
	assIDs := make([]string, 0, len(stepState.BookIDs)*n)
	loContents := make([]*entities.ContentStructure, 0, len(stepState.BookIDs)*n)
	for _, bookID := range stepState.BookIDs {
		stepState.BookID = bookID
		if ctx, err := s.schoolAdminCreateAtopicAndAChapter(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		pbAssignments := s.prepareAssignment(ctx, stepState.TopicID, 4)
		if _, err := eu_v1.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(s.signedCtx(ctx), &eu_v1.UpsertAssignmentsRequest{
			Assignments: pbAssignments,
		}); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unable create a assignment %v", err)
		}
		courseIDs := stepState.BookCourseMap[stepState.BookID]
		for _, ass := range pbAssignments {
			assIDs = append(assIDs, ass.AssignmentId)
		}
		for _, courseID := range courseIDs {
			for _, ass := range pbAssignments {
				loContents = append(loContents, &entities.ContentStructure{
					CourseID:     courseID,
					BookID:       stepState.BookID,
					ChapterID:    stepState.ChapterID,
					TopicID:      stepState.TopicID,
					AssignmentID: ass.AssignmentId,
				})
			}
		}
	}
	stepState.AssignmentIDs = assIDs
	stepState.LOs = loContents
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCreatesSomeAssignmentsInBook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)

	if ctx, err := s.schoolAdminCreateAtopicAndAChapter(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	pbAssignments := s.prepareAssignment(ctx, stepState.TopicID, 4)
	if _, err := eu_v1.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(s.signedCtx(ctx), &eu_v1.UpsertAssignmentsRequest{
		Assignments: pbAssignments,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable create a assignment %v", err)
	}

	assIDs := make([]string, 0, len(pbAssignments))
	assContents := make([]*entities.ContentStructure, 0, len(pbAssignments))
	courseIDs := stepState.BookCourseMap[stepState.BookID]
	for _, ass := range pbAssignments {
		assIDs = append(assIDs, ass.AssignmentId)
	}
	for _, courseID := range courseIDs {
		for _, ass := range pbAssignments {
			assContents = append(assContents, &entities.ContentStructure{
				CourseID:     courseID,
				BookID:       stepState.BookID,
				ChapterID:    stepState.ChapterID,
				TopicID:      stepState.TopicID,
				AssignmentID: ass.AssignmentId,
			})
		}
	}
	stepState.AssignmentIDs = assIDs
	stepState.LOs = assContents
	return StepStateToContext(ctx, stepState), nil
}
