package eureka

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

/*
topic1---chapter1---book1---studyplan1

	\
	 \
	  \

topic2---chapter2---book2---studyplan2

	  /
	 /
	/

topic3---chapter3---book3---studyplan3
*/
func (s *suite) anAssignmentsCreatedEvent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if ctx, err := s.aSignedIn(ctx, "school admin"); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx = contextWithToken(s, ctx)

	books := []*entities.Book{newBook(), newBook(), newBook()}
	for _, book := range books {
		if _, err := database.Insert(ctx, book, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	book1, book2, book3 := books[0], books[1], books[2]

	chapters := []*entities.Chapter{newChapter(), newChapter(), newChapter()}
	for _, c := range chapters {
		if _, err := database.Insert(ctx, c, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	chapter1, chapter2, chapter3 := chapters[0], chapters[1], chapters[2]

	bookChapters := []*entities.BookChapter{
		newBookChapter(book1, chapter1),
		newBookChapter(book2, chapter1),
		newBookChapter(book2, chapter2),
		newBookChapter(book2, chapter3),
		newBookChapter(book3, chapter3),
	}
	for _, bc := range bookChapters {
		if _, err := database.Insert(ctx, bc, s.DB.Exec); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	topic1 := newTopic()
	topic1.ChapterID = chapter1.ID
	if _, err := database.Insert(ctx, topic1, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	topic2 := newTopic()
	topic2.ChapterID = chapter2.ID
	if _, err := database.Insert(ctx, topic2, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	topic3 := newTopic()
	topic3.ChapterID = chapter3.ID
	if _, err := database.Insert(ctx, topic3, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	sp1 := newStudyPlan(book1.ID.String, stepState.CourseID)
	if _, err := database.Insert(ctx, sp1, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	copiedSP1 := newStudyPlan(idutil.ULIDNow(), stepState.CourseID)
	copiedSP1.MasterStudyPlan = sp1.ID
	if _, err := database.Insert(ctx, copiedSP1, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	sp2 := newStudyPlan(book2.ID.String, stepState.CourseID)
	if _, err := database.Insert(ctx, sp2, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	copiedSP2 := newStudyPlan(idutil.ULIDNow(), stepState.CourseID)
	copiedSP2.MasterStudyPlan = sp2.ID
	if _, err := database.Insert(ctx, copiedSP2, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	sp3 := newStudyPlan(book3.ID.String, stepState.CourseID)
	if _, err := database.Insert(ctx, sp3, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	copiedSP3 := newStudyPlan(idutil.ULIDNow(), stepState.CourseID)
	copiedSP3.MasterStudyPlan = sp3.ID
	if _, err := database.Insert(ctx, copiedSP3, s.DB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var items []*epb.StudyPlanItem

	ctx, item := s.generateStudyPlanItem(ctx, "", sp1.ID.String)
	item.ContentStructure = &epb.ContentStructure{
		CourseId:  stepState.CourseID,
		BookId:    book1.ID.String,
		ChapterId: chapter1.ID.String,
		TopicId:   topic1.ID.String,
	}
	item.ContentStructureFlatten = toContentStructureFlatten(item.ContentStructure, "")
	items = append(items, item)

	ctx, item = s.generateStudyPlanItem(ctx, "", sp2.ID.String)
	item.ContentStructure = &epb.ContentStructure{
		CourseId:  stepState.CourseID,
		BookId:    book2.ID.String,
		ChapterId: chapter1.ID.String,
		TopicId:   topic1.ID.String,
	}
	item.ContentStructureFlatten = toContentStructureFlatten(item.ContentStructure, "")
	items = append(items, item)

	ctx, item = s.generateStudyPlanItem(ctx, "", sp2.ID.String)
	item.ContentStructure = &epb.ContentStructure{
		CourseId:  stepState.CourseID,
		BookId:    book2.ID.String,
		ChapterId: chapter2.ID.String,
		TopicId:   topic2.ID.String,
	}
	item.ContentStructureFlatten = toContentStructureFlatten(item.ContentStructure, "")
	items = append(items, item)

	ctx, item = s.generateStudyPlanItem(ctx, "", sp2.ID.String)
	item.ContentStructure = &epb.ContentStructure{
		CourseId:  stepState.CourseID,
		BookId:    book2.ID.String,
		ChapterId: chapter3.ID.String,
		TopicId:   topic3.ID.String,
	}
	item.ContentStructureFlatten = toContentStructureFlatten(item.ContentStructure, "")
	items = append(items, item)

	ctx, item = s.generateStudyPlanItem(ctx, "", sp3.ID.String)
	item.ContentStructure = &epb.ContentStructure{
		CourseId:  stepState.CourseID,
		BookId:    book3.ID.String,
		ChapterId: chapter3.ID.String,
		TopicId:   topic3.ID.String,
	}
	item.ContentStructureFlatten = toContentStructureFlatten(item.ContentStructure, "")
	items = append(items, item)

	s.insertCourseBookIntoBobWithArgs(ctx, book1.ID.String, stepState.CourseID)
	s.insertCourseBookIntoBobWithArgs(ctx, book2.ID.String, stepState.CourseID)
	s.insertCourseBookIntoBobWithArgs(ctx, book3.ID.String, stepState.CourseID)

	s.insertCourseStudyPlanIntoBobWithArgs(ctx, stepState.CourseID, sp1.ID.String)
	s.insertCourseStudyPlanIntoBobWithArgs(ctx, stepState.CourseID, sp2.ID.String)
	s.insertCourseStudyPlanIntoBobWithArgs(ctx, stepState.CourseID, sp3.ID.String)

	if _, err := epb.NewAssignmentModifierServiceClient(s.Conn).UpsertStudyPlanItem(
		contextWithValidVersion(ctx),
		&epb.UpsertStudyPlanItemRequest{
			StudyPlanItems: items,
		},
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = map[string][]entities.ContentStructure{
		topic1.ID.String: {
			{
				BookID:    book1.ID.String,
				ChapterID: chapter1.ID.String,
				TopicID:   topic1.ID.String,
				CourseID:  stepState.CourseID,
			},
			{
				BookID:    book2.ID.String,
				ChapterID: chapter1.ID.String,
				TopicID:   topic1.ID.String,
				CourseID:  stepState.CourseID,
			},
		},
		topic2.ID.String: {
			{
				BookID:    book2.ID.String,
				ChapterID: chapter2.ID.String,
				TopicID:   topic2.ID.String,
				CourseID:  stepState.CourseID,
			},
		},
		topic3.ID.String: {
			{
				BookID:    book2.ID.String,
				ChapterID: chapter3.ID.String,
				TopicID:   topic3.ID.String,
				CourseID:  stepState.CourseID,
			},
			{
				BookID:    book3.ID.String,
				ChapterID: chapter3.ID.String,
				TopicID:   topic3.ID.String,
				CourseID:  stepState.CourseID,
			},
		},
	}

	stepState.TopicID = topic1.ID.String

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemReceivesAssignmentsCreatedEvent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	contentStructuresByTopic := stepState.Request.(map[string][]entities.ContentStructure)
	topicIDs := make([]string, 0, len(contentStructuresByTopic))
	for topicID := range contentStructuresByTopic {
		topicIDs = append(topicIDs, topicID)
	}

	var (
		topic1, topic2, topic3        = topicIDs[0], topicIDs[1], topicIDs[2]
		contentStructuresByAssignment = make(map[string][]entities.ContentStructure)
	)
	_ = topic2
	_ = topic3

	// add new assignment for topic1
	ctx, asm1 := s.generateAssignment(ctx, "", true, true, true)
	asm1.Content.TopicId = topic1
	asm1.DisplayOrder = rand.Int31n(math.MaxInt16)
	contentStructuresByAssignment[asm1.AssignmentId] = contentStructuresByTopic[asm1.Content.TopicId]
	if ctx, err := s.upsertAssignments(ctx, []*epb.Assignment{asm1}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// add new assignment for topic2
	ctx, asm2 := s.generateAssignment(ctx, "", true, true, true)
	asm2.Content.TopicId = topic2
	asm2.DisplayOrder = rand.Int31n(math.MaxInt16)
	contentStructuresByAssignment[asm2.AssignmentId] = contentStructuresByTopic[asm2.Content.TopicId]
	if ctx, err := s.upsertAssignments(ctx, []*epb.Assignment{asm2}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, asm3 := s.generateAssignment(ctx, "", true, true, true)
	asm3.Content.TopicId = topic2
	asm3.DisplayOrder = rand.Int31n(math.MaxInt16)

	ctx, asm4 := s.generateAssignment(ctx, "", true, true, true)
	asm4.AssignmentId = asm3.AssignmentId // make assignment4 same as assignment3
	asm4.Content.TopicId = topic3
	asm4.DisplayOrder = rand.Int31n(math.MaxInt16)
	if ctx, err := s.upsertAssignments(ctx, []*epb.Assignment{asm3, asm4}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	contentStructuresByAssignment[asm4.AssignmentId] = contentStructuresByTopic[asm4.Content.TopicId]
	contentStructuresByAssignment[asm4.AssignmentId] = append(
		contentStructuresByAssignment[asm4.AssignmentId],
		contentStructuresByTopic[asm3.Content.TopicId]...,
	)

	ctx, asm5 := s.generateAssignment(ctx, "", true, true, true)
	asm5.Content.TopicId = topic1
	asm5.DisplayOrder = rand.Int31n(math.MaxInt16)
	contentStructuresByAssignment[asm5.AssignmentId] = contentStructuresByTopic[asm5.Content.TopicId]

	ctx, asm6 := s.generateAssignment(ctx, "", true, true, true)
	asm6.Content.TopicId = topic2
	asm6.DisplayOrder = rand.Int31n(math.MaxInt16)
	contentStructuresByAssignment[asm6.AssignmentId] = contentStructuresByTopic[asm6.Content.TopicId]

	ctx, asm7 := s.generateAssignment(ctx, "", true, true, true)
	asm7.Content.TopicId = topic3
	asm7.DisplayOrder = rand.Int31n(math.MaxInt16)
	contentStructuresByAssignment[asm7.AssignmentId] = contentStructuresByTopic[asm7.Content.TopicId]

	if ctx, err := s.upsertAssignments(ctx, []*epb.Assignment{asm5, asm6, asm7}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	s.assignments = []*epb.Assignment{asm1, asm2, asm3, asm4, asm5, asm6, asm7}
	stepState.Request = contentStructuresByAssignment

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertAssignments(ctx context.Context, asms []*epb.Assignment) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if _, err := epb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(
		contextWithToken(s, ctx),
		&epb.UpsertAssignmentsRequest{
			Assignments: asms,
		},
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemMustUpdateAssignmentStudyPlanItemsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// wait for event handler to be run
	time.Sleep(2 * time.Second)

	contentStructuresByAssignment := stepState.Request.(map[string][]entities.ContentStructure)

	inCS := func(bookID, chapterID, topicID string, cs []entities.ContentStructure) bool {
		for _, c := range cs {
			if c.BookID == bookID && c.ChapterID == chapterID && c.TopicID == topicID {
				return true
			}
		}
		return false
	}

	equalCS := func(cs1, cs2 []entities.ContentStructure) bool {
		if len(cs1) != len(cs2) {
			return false
		}

		for _, c1 := range cs1 {
			if !inCS(c1.BookID, c1.ChapterID, c1.TopicID, cs2) {
				return false
			}
		}
		return true
	}

	for asmID, expectedCS := range contentStructuresByAssignment {
		ctx, cs, err := s.getContentStructuresByAssignment(ctx, asmID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if !equalCS(cs, expectedCS) {
			return StepStateToContext(ctx, stepState), fmt.Errorf(`content structure for assignment: %q is wrong
				got: %+v,
				expected: %+v`,
				asmID, cs, expectedCS,
			)
		}

		ctx, items, err := s.getCopiedAssignmentStudyPlanItems(ctx, asmID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if len(items) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("new learning objectives for copied study plans is missing")
		}

		for itemID, cs := range items {
			// content_structure_flatten must contain loID
			if !strings.Contains(cs, asmID) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("content_structure_flatten of item %q doesn't have loID: %q", itemID, asmID)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) verifyDisplayOrders(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	mAssignments := make(map[string]*epb.Assignment)
	assignmentIDs := []string{}
	for _, assignment := range s.assignments {
		mAssignments[assignment.AssignmentId] = assignment
		assignmentIDs = append(assignmentIDs, assignment.AssignmentId)
	}

	ctx, mDisplayOrders, err := s.getMapDisplayOrders(ctx, assignmentIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for assignmentID, mStudyPlanItemIDAndDisplayOrder := range mDisplayOrders {
		originalDisplayOrder := mAssignments[assignmentID].DisplayOrder

		for studyPlanItemID, displayOrder := range mStudyPlanItemIDAndDisplayOrder {
			if displayOrder != originalDisplayOrder {
				return StepStateToContext(ctx, stepState), fmt.Errorf(`display order for study plan item (assignment_id= %v): %v is wrong
				got: %v,
				expected: %v`, assignmentID, studyPlanItemID, displayOrder, originalDisplayOrder)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getMapDisplayOrders(ctx context.Context, assigmentIDs []string) (context.Context, map[string]map[string]int32, error) {
	stepState := StepStateFromContext(ctx)
	// map assignment_id -> study_plan_item_id -> display_order
	m := make(map[string]map[string]int32)

	query := `SELECT aspi.assignment_id, aspi.study_plan_item_id, spi.display_order
FROM study_plan_items spi
JOIN assignment_study_plan_items aspi
ON spi.study_plan_item_id = aspi.study_plan_item_id 
WHERE aspi.assignment_id = ANY($1)`
	rows, err := s.DB.Query(ctx, query, assigmentIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			assignmentID, studyPlanItemID string
			displayOrder                  int32
		)
		if err := rows.Scan(&assignmentID, &studyPlanItemID, &displayOrder); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		if _, ok := m[assignmentID]; !ok {
			m[assignmentID] = make(map[string]int32)
		}
		m[assignmentID][studyPlanItemID] = displayOrder
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}

	return StepStateToContext(ctx, stepState), m, nil
}

func (s *suite) getContentStructuresByAssignment(ctx context.Context, asmID string) (context.Context, []entities.ContentStructure, error) {
	stepState := StepStateFromContext(ctx)

	query := "SELECT study_plan_item_id FROM assignment_study_plan_items WHERE assignment_id = $1"
	rows, err := s.DB.Query(ctx, query, asmID)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var itemID string
		if err := rows.Scan(&itemID); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		items = append(items, itemID)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}

	query = "SELECT DISTINCT(content_structure) FROM study_plan_items WHERE study_plan_item_id = ANY($1) AND copy_study_plan_item_id IS NULL"
	rows, err = s.DB.Query(ctx, query, database.TextArray(items))
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	var ret []entities.ContentStructure
	for rows.Next() {
		var cs entities.ContentStructure
		if err := rows.Scan(&cs); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		ret = append(ret, cs)
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}

	return StepStateToContext(ctx, stepState), ret, nil
}

func (s *suite) getCopiedAssignmentStudyPlanItems(ctx context.Context, asmID string) (context.Context, map[string]string, error) {
	stepState := StepStateFromContext(ctx)

	query := `
		SELECT spi.study_plan_item_id, content_structure_flatten
		FROM assignment_study_plan_items aspi
		INNER JOIN study_plan_items spi ON spi.copy_study_plan_item_id = aspi.study_plan_item_id
		WHERE assignment_id = $1
	`
	rows, err := s.DB.Query(ctx, query, asmID)
	if err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var itemID, cs string
		if err := rows.Scan(&itemID, &cs); err != nil {
			return StepStateToContext(ctx, stepState), nil, err
		}
		m[itemID] = cs
	}
	if err := rows.Err(); err != nil {
		return StepStateToContext(ctx, stepState), nil, err
	}

	return StepStateToContext(ctx, stepState), m, nil
}

func (s *suite) assignAssignmentsToTopic(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &epb.AssignAssignmentsToTopicRequest{
		TopicId: stepState.TopicID,
		Assignment: []*epb.AssignAssignmentsToTopicRequest_Assignment{
			{
				AssignmentId: s.assignments[0].AssignmentId,
				DisplayOrder: s.assignments[0].DisplayOrder,
			},
		},
	}

	stepState.Response, stepState.ResponseErr = epb.NewAssignmentModifierServiceClient(s.Conn).
		AssignAssignmentsToTopic(contextWithToken(s, ctx), req)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), nil
}
