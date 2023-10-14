package utils

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"go.uber.org/multierr"
)

const DefaultSchoolID = 1

func generateBook(schoolID int32) (*entities.Book, error) {
	book := &entities.Book{}
	bookID := idutil.ULIDNow()
	database.AllNullEntity(book)
	now := timeutil.Now()
	if err := multierr.Combine(book.ID.Set(bookID),
		book.BookType.Set(cpb.BookType_BOOK_TYPE_GENERAL),
		book.Name.Set(FormatName(bookID)),
		book.SchoolID.Set(schoolID),
		book.CreatedAt.Set(now),
		book.UpdatedAt.Set(now),
		book.CurrentChapterDisplayOrder.Set(0)); err != nil {
		return nil, fmt.Errorf("unable to setup a book: %w", err)
	}
	return book, nil
}

func generateAssignment() (*entities.Assignment, error) {
	assignment := &entities.Assignment{}
	assignmentID := idutil.ULIDNow()
	topicID := idutil.ULIDNow()
	database.AllNullEntity(assignment)
	assignment.Now()
	if err := multierr.Combine(assignment.ID.Set(assignmentID),
		assignment.Name.Set(FormatName(assignmentID)),
		assignment.Content.Set(entities.AssignmentContent{
			TopicID: topicID,
			LoIDs:   []string{idutil.ULIDNow(), idutil.ULIDNow()},
		})); err != nil {
		return nil, fmt.Errorf("unable to setup a assignment: %w", err)
	}
	return assignment, nil
}

func generateCourseStudent(courseID string) (*entities.CourseStudent, error) {
	courseStudent := &entities.CourseStudent{}
	id := idutil.ULIDNow()
	database.AllNullEntity(courseStudent)
	courseStudent.Now()
	if err := multierr.Combine(
		courseStudent.ID.Set(id),
		courseStudent.StudentID.Set(id),
		courseStudent.CourseID.Set(courseID),
	); err != nil {
		return nil, fmt.Errorf("unable to setup a course student: %w", err)
	}
	return courseStudent, nil
}

func generateStudyPlan(opts ...StudyPlanOption) (*entities.StudyPlan, error) {
	studyPlan := &entities.StudyPlan{}
	id := idutil.ULIDNow()
	database.AllNullEntity(studyPlan)
	studyPlan.Now()
	if err := multierr.Combine(
		studyPlan.ID.Set(id),
		studyPlan.MasterStudyPlan.Set(id),
		studyPlan.Name.Set(FormatName(id)),
		studyPlan.SchoolID.Set(DefaultSchoolID),
		studyPlan.CourseID.Set(id),
		studyPlan.BookID.Set(id),
		studyPlan.Status.Set("STUDY_PLAN_STATUS_ACTIVE"),
	); err != nil {
		return nil, err
	}
	for _, opt := range opts {
		err := opt(studyPlan)
		if err != nil {
			return nil, fmt.Errorf("unable to setup a study plan: %w", err)
		}
	}
	return studyPlan, nil
}

func generateStudentStudyPlan(studentID, studyPlanID string) (*entities.StudentStudyPlan, error) {
	studentStudyPlan := &entities.StudentStudyPlan{}
	database.AllNullEntity(studentStudyPlan)
	studentStudyPlan.Now()
	if err := multierr.Combine(
		studentStudyPlan.MasterStudyPlanID.Set(studyPlanID),
		studentStudyPlan.StudyPlanID.Set(studyPlanID),
		studentStudyPlan.StudentID.Set(studentID),
	); err != nil {
		return nil, fmt.Errorf("unable to setup a student study plan: %w", err)
	}
	return studentStudyPlan, nil
}

func generateLoStudyPlanItem(studyPlanItemID string) (*entities.LoStudyPlanItem, error) {
	loStudyPlanItem := &entities.LoStudyPlanItem{}
	database.AllNullEntity(loStudyPlanItem)
	loID := idutil.ULIDNow()
	loStudyPlanItem.Now()
	if err := multierr.Combine(
		loStudyPlanItem.StudyPlanItemID.Set(studyPlanItemID),
		loStudyPlanItem.LoID.Set(loID),
	); err != nil {
		return nil, fmt.Errorf("unable to setup a lo plan item: %w", err)
	}
	return loStudyPlanItem, nil
}

func generateAssignmentStudyPlanItem(assignmentID, studyPlanItemID string) (*entities.AssignmentStudyPlanItem, error) {
	assignmentStudyPlanItem := &entities.AssignmentStudyPlanItem{}
	database.AllNullEntity(assignmentStudyPlanItem)
	assignmentStudyPlanItem.Now()
	if err := multierr.Combine(
		assignmentStudyPlanItem.StudyPlanItemID.Set(studyPlanItemID),
		assignmentStudyPlanItem.AssignmentID.Set(assignmentID),
	); err != nil {
		return nil, fmt.Errorf("unable to setup a assignment study plan item: %w", err)
	}
	return assignmentStudyPlanItem, nil
}

func generateCourseStudyPlan(studyPlanID, courseID string) (*entities.CourseStudyPlan, error) {
	courseStudyPlan := &entities.CourseStudyPlan{}
	database.AllNullEntity(courseStudyPlan)
	courseStudyPlan.Now()
	err := multierr.Combine(
		courseStudyPlan.CourseID.Set(courseID),
		courseStudyPlan.StudyPlanID.Set(studyPlanID),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to setup a course study plan: %w", err)
	}
	return courseStudyPlan, nil
}

func generateStudyPlanItem(opts ...StudyPlanItemOption) (*entities.StudyPlanItem, error) {
	now := time.Now()
	id := idutil.ULIDNow()
	e := &entities.StudyPlanItem{}
	contentStructure := entities.ContentStructure{
		CourseID:     fmt.Sprintf("CourseID_%v", id),
		BookID:       fmt.Sprintf("BookID_%v", id),
		ChapterID:    fmt.Sprintf("ChapterID_%v", id),
		TopicID:      fmt.Sprintf("TopicID_%v", id),
		LoID:         fmt.Sprintf("LoID_%v", id),
		AssignmentID: fmt.Sprintf("AssignmentID_%v", id),
	}

	database.AllNullEntity(e)
	e.Now()
	err := multierr.Combine(
		e.ID.Set(id),
		e.StudyPlanID.Set(id),
		e.ContentStructure.Set(contentStructure),
		e.DisplayOrder.Set(0),
		e.CopyStudyPlanItemID.Set(id),
		e.Status.Set("STUDY_PLAN_ITEM_STATUS_NONE"),
		e.SchoolDate.Set(now),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to setup a study plan item: %w", err)
	}
	for _, opt := range opts {
		err := opt(e)
		if err != nil {
			return nil, fmt.Errorf("unable to setup a study plan item: %w", err)
		}
	}
	return e, nil
}
