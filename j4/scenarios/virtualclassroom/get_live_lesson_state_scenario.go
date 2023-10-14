package virtualclassroom

import (
	"context"

	"github.com/manabie-com/backend/j4/infras"
	"github.com/manabie-com/backend/j4/serviceutil"
	"github.com/manabie-com/backend/j4/serviceutil/virtualclassroom"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	j4 "github.com/manabie-com/j4/pkg/runner"
)

type GetLiveLessonStateScenario struct {
	tokenGenerator *serviceutil.TokenGenerator
	j4cfg          *infras.ManabieJ4Config
	conns          *infras.Connections
}

func (g *GetLiveLessonStateScenario) getOneLessonTestScenario(ctx context.Context) (*j4.Scenario, error) {
	schoolID := g.j4cfg.VirtualClassroomConfig.SchoolID
	adminID := g.j4cfg.VirtualClassroomConfig.AdminID

	runConfig, err := g.j4cfg.GetScenarioConfig("Virtualclassroom_GetLiveLessonState")
	if err != nil {
		return nil, err
	}
	runCfg := infras.MustOptionFromConfig(&runConfig)

	list := &virtualclassroom.LiveLessonList{
		ListCfg: g.j4cfg.VirtualClassroomConfig.LessonInfo,
	}
	db := g.conns.DBConnPools["bob"]

	adminToken, err := g.tokenGenerator.GetTokenFromShamir(ctx, adminID, schoolID)
	if err != nil {
		return nil, err
	}

	lessonID, studentID, err := list.GetOneLessonWithStudentID(contextWithToken(ctx, adminToken), db, schoolID)
	if err != nil {
		return nil, err
	}

	runCfg.TestFunc = func(ctx context.Context) error {
		conn, err := g.conns.PoolToGateWay.Get(ctx)
		if err != nil {
			return err
		}
		defer conn.Close()

		studentToken, err := g.tokenGenerator.GetTokenFromShamir(ctx, studentID, schoolID)
		if err != nil {
			return err
		}

		vpbClient := vpb.NewVirtualClassroomReaderServiceClient(conn)
		_, err = vpbClient.GetLiveLessonState(contextWithToken(ctx, studentToken), &vpb.GetLiveLessonStateRequest{
			LessonId: lessonID,
		})
		return err
	}

	scenario, err := j4.NewScenario("Virtualclassroom_GetLiveLessonState_OneLessonScenario", *runCfg)
	if err != nil {
		return nil, err
	}

	return scenario, nil
}

func (g *GetLiveLessonStateScenario) getMultipleLessonTestScenario(ctx context.Context) (*j4.Scenario, error) {
	schoolID := g.j4cfg.VirtualClassroomConfig.SchoolID
	adminID := g.j4cfg.VirtualClassroomConfig.AdminID

	runConfig, err := g.j4cfg.GetScenarioConfig("Virtualclassroom_GetLiveLessonState")
	if err != nil {
		return nil, err
	}
	runCfg := infras.MustOptionFromConfig(&runConfig)

	list := &virtualclassroom.LiveLessonList{
		ListCfg: g.j4cfg.VirtualClassroomConfig.LessonInfo,
	}
	db := g.conns.DBConnPools["bob"]

	adminToken, err := g.tokenGenerator.GetTokenFromShamir(ctx, adminID, schoolID)
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

		conn, err := g.conns.PoolToGateWay.Get(ctx)
		if err != nil {
			return err
		}
		defer conn.Close()

		studentToken, err := g.tokenGenerator.GetTokenFromShamir(ctx, studentID, schoolID)
		if err != nil {
			return err
		}

		vpbClient := vpb.NewVirtualClassroomReaderServiceClient(conn)
		_, err = vpbClient.GetLiveLessonState(contextWithToken(ctx, studentToken), &vpb.GetLiveLessonStateRequest{
			LessonId: lessonID,
		})
		return err
	}

	scenario, err := j4.NewScenario("Virtualclassroom_GetLiveLessonState_MultipleLessonScenario", *runCfg)
	if err != nil {
		return nil, err
	}

	return scenario, nil
}
