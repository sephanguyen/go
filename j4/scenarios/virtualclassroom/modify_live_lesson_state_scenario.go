package virtualclassroom

import (
	"context"

	"github.com/manabie-com/backend/j4/infras"
	"github.com/manabie-com/backend/j4/serviceutil"
	"github.com/manabie-com/backend/j4/serviceutil/virtualclassroom"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	j4 "github.com/manabie-com/j4/pkg/runner"
)

type ModifyLiveLessonStateScenario struct {
	tokenGenerator *serviceutil.TokenGenerator
	j4cfg          *infras.ManabieJ4Config
	conns          *infras.Connections
}

func (m *ModifyLiveLessonStateScenario) getOneLessonTestScenario(ctx context.Context) (*j4.Scenario, error) {
	schoolID := m.j4cfg.VirtualClassroomConfig.SchoolID
	adminID := m.j4cfg.VirtualClassroomConfig.AdminID

	runConfig, err := m.j4cfg.GetScenarioConfig("Virtualclassroom_ModifyLiveLessonState")
	if err != nil {
		return nil, err
	}
	runCfg := infras.MustOptionFromConfig(&runConfig)

	list := &virtualclassroom.LiveLessonList{
		ListCfg: m.j4cfg.VirtualClassroomConfig.LessonInfo,
	}
	db := m.conns.DBConnPools["bob"]

	adminToken, err := m.tokenGenerator.GetTokenFromShamir(ctx, adminID, schoolID)
	if err != nil {
		return nil, err
	}

	lessonID, studentID, err := list.GetOneLessonWithStudentID(contextWithToken(ctx, adminToken), db, schoolID)
	if err != nil {
		return nil, err
	}

	runCfg.TestFunc = func(ctx context.Context) error {
		conn, err := m.conns.PoolToGateWay.Get(ctx)
		if err != nil {
			return err
		}
		defer conn.Close()

		studentToken, err := m.tokenGenerator.GetTokenFromShamir(ctx, studentID, schoolID)
		if err != nil {
			return err
		}

		vpbClient := vpb.NewVirtualClassroomModifierServiceClient(conn)
		_, err = vpbClient.ModifyVirtualClassroomState(contextWithToken(ctx, studentToken), &vpb.ModifyVirtualClassroomStateRequest{
			Id: lessonID,
			Command: &vpb.ModifyVirtualClassroomStateRequest_RaiseHand{
				RaiseHand: true,
			},
		})
		return err
	}

	scenario, err := j4.NewScenario("Virtualclassroom_ModifyLiveLessonState_OneLessonScenario", *runCfg)
	if err != nil {
		return nil, err
	}

	return scenario, nil
}

func (m *ModifyLiveLessonStateScenario) getMultipleLessonTestScenario(ctx context.Context) (*j4.Scenario, error) {
	schoolID := m.j4cfg.VirtualClassroomConfig.SchoolID
	adminID := m.j4cfg.VirtualClassroomConfig.AdminID

	runConfig, err := m.j4cfg.GetScenarioConfig("Virtualclassroom_ModifyLiveLessonState")
	if err != nil {
		return nil, err
	}
	runCfg := infras.MustOptionFromConfig(&runConfig)

	list := &virtualclassroom.LiveLessonList{
		ListCfg: m.j4cfg.VirtualClassroomConfig.LessonInfo,
	}
	db := m.conns.DBConnPools["bob"]

	adminToken, err := m.tokenGenerator.GetTokenFromShamir(ctx, adminID, schoolID)
	if err != nil {
		return nil, err
	}

	if err := list.GetMultipleLessons(contextWithToken(ctx, adminToken), db, schoolID); err != nil {
		return nil, err
	}

	runCfg.TestFunc = func(ctx context.Context) error {
		lessonID, studentID, err := list.GetOneFromLessonList()
		if err != nil {
			return err
		}

		conn, err := m.conns.PoolToGateWay.Get(ctx)
		if err != nil {
			return err
		}
		defer conn.Close()

		studentToken, err := m.tokenGenerator.GetTokenFromShamir(ctx, studentID, schoolID)
		if err != nil {
			return err
		}

		vpbClient := vpb.NewVirtualClassroomModifierServiceClient(conn)
		_, err = vpbClient.ModifyVirtualClassroomState(contextWithToken(ctx, studentToken), &vpb.ModifyVirtualClassroomStateRequest{
			Id: lessonID,
			Command: &vpb.ModifyVirtualClassroomStateRequest_RaiseHand{
				RaiseHand: true,
			},
		})
		return err
	}

	scenario, err := j4.NewScenario("Virtualclassroom_ModifyLiveLessonState_MultipleLessonScenario", *runCfg)
	if err != nil {
		return nil, err
	}

	return scenario, nil
}
