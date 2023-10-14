package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
)

type ConfigPermissionCommandChecker struct {
	Lesson              *domain.VirtualLesson
	Ctx                 context.Context
	WrapperDBConnection *support.WrapperDBConnection
	StudentsRepo        infrastructure.StudentsRepo
}

type CommandChecker interface {
	Check(command StateModifyCommand) error
}

type PermissionCommandChecker struct {
	*ConfigPermissionCommandChecker
}

func Create(cf *ConfigPermissionCommandChecker) *PermissionCommandChecker {
	return &PermissionCommandChecker{&ConfigPermissionCommandChecker{
		Lesson:              cf.Lesson,
		Ctx:                 cf.Ctx,
		WrapperDBConnection: cf.WrapperDBConnection,
		StudentsRepo:        cf.StudentsRepo,
	}}
}

func (p *PermissionCommandChecker) isTeacher(commander string) error {
	conn, err := p.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(p.Ctx))
	if err != nil {
		return err
	}
	if !p.Lesson.TeacherIDs.HaveID(commander) {
		isUserAStudent, err := p.StudentsRepo.IsUserIDAStudent(p.Ctx, conn, commander)
		if err != nil {
			return fmt.Errorf("error in StudentsRepo.IsUserIDAStudent, user %s: %w", commander, err)
		}
		if isUserAStudent {
			return fmt.Errorf("permission denied: user %s is a student and cannot execute this command", commander)
		}
	}
	return nil
}

func (p *PermissionCommandChecker) Check(command StateModifyCommand) error {
	commander := command.GetCommander()
	switch v := command.(type) {
	case *UpdateHandsUpCommand:
		if p.Lesson.LearnerIDs.HaveID(commander) {
			// learner can only update self-state
			if commander != v.UserID {
				return fmt.Errorf("permission denied: learner %s can not change other member's status", commander)
			}
		} else if err := p.isTeacher(commander); err != nil {
			return err
		}
	case *SubmitPollingAnswerCommand:
		if p.Lesson.LearnerIDs.HaveID(commander) {
			// learner can only submit self-state
			if commander != v.UserID {
				return fmt.Errorf("permission denied: other learner %s can not submit polling answer", commander)
			}
		} else {
			return fmt.Errorf("permission denied: user %s can not submit polling answer", commander)
		}
	default:
		if err := p.isTeacher(commander); err != nil {
			return err
		}
	}
	return nil
}
