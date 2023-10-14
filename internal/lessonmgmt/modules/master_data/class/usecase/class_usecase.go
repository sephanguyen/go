package usecase

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	allocation_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
	user_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"
)

type ClassUseCase struct {
	ClassRepo               domain.ClassRepository
	CourseRepo              domain.CourseRepository
	StudentSubscriptionRepo infrastructure.StudentSubscriptionRepo
}

func (c *ClassUseCase) GetByStudentSubscription(ctx context.Context, db database.Ext, studentSubID []string) ([]*domain.ClassUnassigned, error) {
	studentSub, err := c.StudentSubscriptionRepo.GetByStudentSubscriptionIDs(ctx, db, studentSubID)
	if err != nil {
		return nil, fmt.Errorf("StudentSubscriptionRepo.GetByStudentSubscriptionIDs: %v", err)
	}

	courseIDs := sliceutils.Map(studentSub, func(ss *user_domain.StudentSubscription) string {
		return ss.CourseID
	})

	courses, err := c.CourseRepo.GetByIDs(ctx, db, courseIDs)
	if err != nil {
		return nil, fmt.Errorf("CourseRepo.GetByIDs: %v", err)
	}
	courseMap := make(map[string]string, len(courses))
	for _, c := range courses {
		courseMap[c.CourseID.String] = c.TeachingMethod.String
	}

	studentCourse := make([]string, 0, len(studentSub)*2)
	for i := range studentSub {
		if courseMap[studentSub[i].CourseID] == string(allocation_domain.Group) {
			studentID := studentSub[i].StudentID
			courseID := studentSub[i].CourseID
			studentCourse = append(studentCourse, studentID, courseID)
		}
	}

	studentCourseWithClass := map[string]string{}
	studentCourseWithReserveClass := map[string]string{}
	if len(studentCourse) > 0 {
		studentCourseWithClass, err = c.ClassRepo.GetByStudentCourse(ctx, db, studentCourse)
		if err != nil {
			return nil, fmt.Errorf("ClassRepo.GetByStudentCourse: %v", err)
		}
		studentCourseWithReserveClass, err = c.ClassRepo.GetReserveClass(ctx, db, studentCourse)
		if err != nil {
			return nil, fmt.Errorf("ClassRepo.GetReserveClass: %v", err)
		}
	}

	classUnassigned := make([]*domain.ClassUnassigned, 0, len(studentSub))
	for _, ss := range studentSub {
		c := &domain.ClassUnassigned{
			StudentSubscriptionID: ss.StudentSubscriptionID,
		}
		if courseMap[ss.CourseID] == string(allocation_domain.Group) {
			c.IsClassUnAssigned = true
		}
		if _, ok := studentCourseWithClass[ss.StudentWithCourseID()]; ok {
			c.IsClassUnAssigned = false
		}
		if _, ok := studentCourseWithReserveClass[ss.StudentWithCourseID()]; ok {
			c.IsClassUnAssigned = false
		}
		classUnassigned = append(classUnassigned, c)
	}
	return classUnassigned, nil
}
