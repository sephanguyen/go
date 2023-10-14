package application

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	user_infrastructure "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"
)

type UpdaterLessonMember struct {
	DB database.Ext

	// ports
	LessonMemberRepo infrastructure.LessonMemberRepo
	UserRepo         user_infrastructure.UserRepo
}

func (l *UpdaterLessonMember) UpdateLessonMemberNames(ctx context.Context, lessonMembers *domain.LessonMembers) error {
	updateLessonMembers := make([]*domain.UpdateLessonMemberName, 0, len(*lessonMembers))
	userIds := lessonMembers.GetStudentIDs()

	userDetails, err := l.UserRepo.Retrieve(ctx, l.DB, database.TextArray(userIds))
	if err != nil {
		return err
	}
	fields := lessonMembers.GetMapFieldValuesOfStudent()
	if len(fields) == 0 {
		return fmt.Errorf("could not get info of user")
	}
	for _, user := range userDetails {
		lessonMember, ok := fields[user.ID.String]
		if !ok {
			return fmt.Errorf("could not map info of user: %s", user.ID.String)
		}
		updateLessonMember := domain.UpdateLessonMemberName{
			LessonID:      lessonMember.LessonID,
			StudentID:     lessonMember.StudentID,
			UserFirstName: user.FirstName.String,
			UserLastName:  user.LastName.String,
		}
		updateLessonMembers = append(updateLessonMembers, &updateLessonMember)
	}

	if err = l.LessonMemberRepo.UpdateLessonMemberNames(ctx, l.DB, updateLessonMembers); err != nil {
		return fmt.Errorf("LessonRepo.GetLessonByID err: %w", err)
	}

	return nil
}
