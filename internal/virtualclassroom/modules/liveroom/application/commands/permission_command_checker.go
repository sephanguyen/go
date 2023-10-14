package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	vc_infrastructure "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
)

type ConfigPermissionCommandChecker struct {
	Ctx                 context.Context
	WrapperDBConnection *support.WrapperDBConnection
	StudentsRepo        vc_infrastructure.StudentsRepo
}

type PermissionCommandChecker struct {
	*ConfigPermissionCommandChecker
}

func Create(config *ConfigPermissionCommandChecker) *PermissionCommandChecker {
	return &PermissionCommandChecker{
		&ConfigPermissionCommandChecker{
			Ctx:                 config.Ctx,
			WrapperDBConnection: config.WrapperDBConnection,
			StudentsRepo:        config.StudentsRepo,
		},
	}
}

func (p *PermissionCommandChecker) isTeacher(commanderID string) error {
	conn, err := p.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(p.Ctx))
	if err != nil {
		return err
	}
	isUserAStudent, err := p.StudentsRepo.IsUserIDAStudent(p.Ctx, conn, commanderID)
	if err != nil {
		return fmt.Errorf("error in StudentsRepo.IsUserIDAStudent, user %s: %w", commanderID, err)
	}
	if isUserAStudent {
		return fmt.Errorf("permission denied: user %s is a student and cannot execute this command", commanderID)
	}
	return nil
}

func (p *PermissionCommandChecker) isStudent(commanderID string) error {
	conn, err := p.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(p.Ctx))
	if err != nil {
		return err
	}
	isUserAStudent, err := p.StudentsRepo.IsUserIDAStudent(p.Ctx, conn, commanderID)
	if err != nil {
		return fmt.Errorf("error in StudentsRepo.IsUserIDAStudent, user %s: %w", commanderID, err)
	}
	if !isUserAStudent {
		return fmt.Errorf("permission denied: user %s can not execute this command, only student can execute this command", commanderID)
	}
	return nil
}

func (p *PermissionCommandChecker) Check(command ModifyStateCommand) error {
	commanderID := command.GetCommander()

	switch v := command.(type) {
	case *UpdateHandsUpCommand:
		// if user is a student, check if the commander ID is equal to the user whose hand to be updated
		// otherwise if not student, can execute the command
		err := p.isStudent(commanderID)
		if err == nil {
			if commanderID != v.UserID {
				return fmt.Errorf("permission denied: other learner %s can not update hand state", commanderID)
			}
		} else if err != nil && !strings.Contains(err.Error(), "permission denied") {
			return err
		}
	case *SubmitPollingAnswerCommand:
		if err := p.isStudent(commanderID); err != nil {
			return err
		}

		if commanderID != v.UserID {
			return fmt.Errorf("permission denied: other learner %s can not submit polling answer", commanderID)
		}
	case *ShareMaterialCommand, *StopSharingMaterialCommand:
		// allow student and non-student users to start and stop share material
	default:
		if err := p.isTeacher(commanderID); err != nil {
			return err
		}
	}

	return nil
}
