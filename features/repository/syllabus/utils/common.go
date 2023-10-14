package utils

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"

	"go.uber.org/multierr"
)

// import "context"
// be attention: it use default school ID

func AUserInsertSomeBooksToDatabase(ctx context.Context, db database.Ext, schoolID int32, numOfBooks int) ([]*entities.Book, error) {
	books := make([]*entities.Book, 0, numOfBooks)
	for i := 0; i < numOfBooks; i++ {
		book, err := generateBook(schoolID)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	bookRepo := repositories.BookRepo{}
	if err := bookRepo.Upsert(ctx, db, books); err != nil {
		return nil, fmt.Errorf("unable to create a book: %w", err)
	}
	return books, nil
}

func AUserInsertSomeCourseStudentToDatabase(ctx context.Context, db database.Ext, courseIDs []string) ([]*entities.CourseStudent, error) {
	n := len(courseIDs)
	courseStudents := make([]*entities.CourseStudent, 0, n)

	for i := 0; i < n; i++ {
		courseStudent, err := generateCourseStudent(courseIDs[i])
		if err != nil {
			return nil, err
		}
		courseStudents = append(courseStudents, courseStudent)
	}
	courseStudentRepo := repositories.CourseStudentRepo{}
	if _, err := courseStudentRepo.BulkUpsert(ctx, db, courseStudents); err != nil {
		return nil, fmt.Errorf("unable to create a courseStudent: %w", err)
	}
	return courseStudents, nil
}

func AUserInsertSomeAssignmentsToDatabase(ctx context.Context, db database.Ext, numOfAssignments int) ([]*entities.Assignment, error) {
	assignments := make([]*entities.Assignment, 0, numOfAssignments)
	for i := 0; i < numOfAssignments; i++ {
		assignment, err := generateAssignment()
		if err != nil {
			return nil, fmt.Errorf("unable to setup an assignment: %w", err)
		}
		assignments = append(assignments, assignment)
	}

	assignmentRepo := repositories.AssignmentRepo{}
	if err := assignmentRepo.BulkUpsert(ctx, db, assignments); err != nil {
		return nil, fmt.Errorf("unable to create a assignment: %w", err)
	}
	return assignments, nil
}

func spiWithContentStructure(c *entities.ContentStructure) StudyPlanItemOption {
	return func(e *entities.StudyPlanItem) error {
		err := e.ContentStructure.Set(c)
		return err
	}
}

func spiWithStudyPlanID(id string) StudyPlanItemOption {
	return func(u *entities.StudyPlanItem) error {
		err := u.StudyPlanID.Set(id)
		return err
	}
}

func AUserInsertSomeStudyPlanItemsToDatabase(ctx context.Context, db database.Ext, courseID string, studyPlanIDs []string,
	contentStructures []*entities.ContentStructure) ([]*entities.StudyPlanItem, error) {
	numOfStudyPlanItems := len(studyPlanIDs)
	studyPlanItems := make([]*entities.StudyPlanItem, 0, numOfStudyPlanItems)
	for i := 0; i < numOfStudyPlanItems; i++ {
		contentStructure := contentStructures[i]
		studyPlanItem, err := generateStudyPlanItem(spiWithContentStructure(contentStructure), spiWithStudyPlanID(studyPlanIDs[i]))
		if err != nil {
			return nil, fmt.Errorf("unable to setup a study plan item: %w", err)
		}
		studyPlanItems = append(studyPlanItems, studyPlanItem)
		contentStructures = append(contentStructures, contentStructure)
	}

	studyPlanItemRepo := repositories.StudyPlanItemRepo{}
	if err := studyPlanItemRepo.BulkInsert(ctx, db, studyPlanItems); err != nil {
		return nil, fmt.Errorf("unable to create a study plan item: %w", err)
	}
	return studyPlanItems, nil
}

func withDisplayOrder(o int) StudyPlanItemOption {
	return func(u *entities.StudyPlanItem) error {
		err := u.DisplayOrder.Set(o)
		return err
	}
}

func AUserInsertSomeStudyPlanItemsToDatabaseWithStudyPlanID(ctx context.Context, db database.Ext, opts ...StudyPlanItemOption) ([]*entities.StudyPlanItem, error) {
	n := rand.Intn(5) + 10
	studyPlanItems := make([]*entities.StudyPlanItem, 0, n)
	for i := 0; i < n; i++ {
		studyPlanItem, err := generateStudyPlanItem(opts...)
		if errs := multierr.Combine(
			err,
			withDisplayOrder(i)(studyPlanItem)); errs != nil {
			return nil, fmt.Errorf("unable to setup a study plan item: %w", errs)
		}
		studyPlanItems = append(studyPlanItems, studyPlanItem)
	}
	studyPlanItemRepo := repositories.StudyPlanItemRepo{}
	if err := studyPlanItemRepo.BulkInsert(ctx, db, studyPlanItems); err != nil {
		return nil, fmt.Errorf("unable to create some course study plan items: %w", err)
	}
	return studyPlanItems, nil
}

func spWithCreatedAt(t time.Time) StudyPlanOption {
	return func(s *entities.StudyPlan) error {
		err := s.CreatedAt.Set(t)
		return err
	}
}

func AUserInsertSomeStudyPlanToDatabase(ctx context.Context, db database.Ext, opts ...StudyPlanOption) ([]*entities.StudyPlan, error) {
	n := rand.Intn(5) + 10
	studyPlans := make([]*entities.StudyPlan, 0, n)
	for i := 0; i < n; i++ {
		studyPlan, err := generateStudyPlan(opts...)
		if err != nil {
			return nil, fmt.Errorf("unable to setup a study plan: %w", err)
		}
		spWithCreatedAt(time.Now().Add(time.Duration(i) * time.Millisecond))(studyPlan)
		studyPlans = append(studyPlans, studyPlan)
	}

	studyPlanRepo := repositories.StudyPlanRepo{}
	if err := studyPlanRepo.BulkUpsert(ctx, db, studyPlans); err != nil {
		return nil, fmt.Errorf("unable to create a study plan: %w", err)
	}
	return studyPlans, nil
}

func AUserInsertSomeStudentStudyPlanToDatabase(ctx context.Context, db database.Ext, studentIDs []string, studyPlanIDs []string) ([]*entities.StudentStudyPlan, error) {
	studentStudyPlans := make([]*entities.StudentStudyPlan, 0, len(studentIDs))
	n := len(studyPlanIDs)
	for i := 0; i < n; i++ {
		studentStudyPlan, err := generateStudentStudyPlan(studentIDs[i], studyPlanIDs[i])
		if err != nil {
			return nil, fmt.Errorf("unable to setup a student study plan: %w", err)
		}
		if err := studentStudyPlan.CreatedAt.Set(time.Now().Add(time.Millisecond * time.Duration(i))); err != nil {
			return nil, fmt.Errorf("unable to setup a student study plan: %w", err)
		}
		studentStudyPlans = append(studentStudyPlans, studentStudyPlan)
	}
	studentStudyPlanRepo := repositories.StudentStudyPlanRepo{}
	if err := studentStudyPlanRepo.BulkUpsert(ctx, db, studentStudyPlans); err != nil {
		return nil, fmt.Errorf("unable to create a student study plan: %w", err)
	}
	return studentStudyPlans, nil
}

func AUserInsertSomeCourseStudyPlansToDatabase(ctx context.Context, db database.Ext, courseID string, studyPlanIDs []string) ([]*entities.CourseStudyPlan, error) {
	n := rand.Intn(5) + 10
	courseStudyPlans := make([]*entities.CourseStudyPlan, 0, n)
	for i, studyPlanID := range studyPlanIDs {
		courseStudyPlan, err := generateCourseStudyPlan(studyPlanID, courseID)
		errs := multierr.Combine(
			err,
			courseStudyPlan.CreatedAt.Set(time.Now().Add(time.Millisecond*time.Duration(i))),
		)
		if errs != nil {
			return nil, errs
		}
		courseStudyPlans = append(courseStudyPlans, courseStudyPlan)
	}
	CourseStudyPlanRepo := repositories.CourseStudyPlanRepo{}
	if err := CourseStudyPlanRepo.BulkUpsert(ctx, db, courseStudyPlans); err != nil {
		return nil, fmt.Errorf("unable to create a course study plan: %w", err)
	}

	return courseStudyPlans, nil
}

func AUserInsertSomeLoStudyPlanItemsToDatabase(ctx context.Context, db database.Ext, studyPlanItemIDs []string) ([]*entities.LoStudyPlanItem, error) {
	loStudyPlanItems := make([]*entities.LoStudyPlanItem, 0, len(studyPlanItemIDs))
	for i := 0; i < len(studyPlanItemIDs); i++ {
		loStudyPlanItem, err := generateLoStudyPlanItem(studyPlanItemIDs[i])
		if err != nil {
			return nil, fmt.Errorf("unable to setup a lo study plan item: %w", err)
		}
		loStudyPlanItems = append(loStudyPlanItems, loStudyPlanItem)
	}

	loStudyPlanItemRepo := repositories.LoStudyPlanItemRepo{}
	if err := loStudyPlanItemRepo.BulkInsert(ctx, db, loStudyPlanItems); err != nil {
		return nil, fmt.Errorf("unable to create a lo study plan item: %w", err)
	}
	return loStudyPlanItems, nil
}

func AUserInsertSomeAssignmentStudyPlanItemsToDatabase(ctx context.Context, db database.Ext, assignmentIDs, studyPlanItemIDs []string) ([]*entities.AssignmentStudyPlanItem, error) {
	assignmentStudyPlanItems := make([]*entities.AssignmentStudyPlanItem, 0, len(assignmentIDs))
	for i := 0; i < len(assignmentIDs); i++ {
		assignmentStudyPlanItem, err := generateAssignmentStudyPlanItem(assignmentIDs[i], studyPlanItemIDs[i])
		if err != nil {
			return nil, fmt.Errorf("unable to setup an assignment study plan item: %w", err)
		}
		assignmentStudyPlanItems = append(assignmentStudyPlanItems, assignmentStudyPlanItem)
	}

	assignmentStudyPlanItemRepo := repositories.AssignmentStudyPlanItemRepo{}
	if err := assignmentStudyPlanItemRepo.BulkInsert(ctx, db, assignmentStudyPlanItems); err != nil {
		return nil, fmt.Errorf("unable to create a assignment study plan item: %w", err)
	}
	return assignmentStudyPlanItems, nil
}
