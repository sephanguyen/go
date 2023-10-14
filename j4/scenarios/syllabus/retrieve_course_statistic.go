package syllabus

import (
	"context"

	"github.com/manabie-com/backend/j4/infras"
	"github.com/manabie-com/backend/j4/serviceutil"
	spb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"
	j4 "github.com/manabie-com/j4/pkg/runner"
)

func GenRetrieveCourseStatisticScenario(ctx context.Context, c *infras.ManabieJ4Config, dep *infras.Dep) (*j4.Scenario, error) {
	scenarioConf, err := c.GetScenarioConfig("Syllabus_RetrieveCourseStatistic")
	if err != nil {
		return nil, err
	}
	scenarioOpt := infras.MustOptionFromConfig(&scenarioConf)
	tokenGen := serviceutil.NewTokenGenerator(c, dep.Connections)

	scenarioOpt.TestFunc = func(ctx context.Context) error {
		conn, err := dep.PoolToGateWay.Get(ctx)
		if err != nil {
			return err
		}
		defer conn.Close()
		token, err := tokenGen.GetTokenFromShamir(ctx, c.SyllabusConfig.UserID, c.SyllabusConfig.ResourcePath)
		if err != nil {
			return err
		}

		syllabusClient := spb.NewStatisticsClient(conn)
		req := &spb.CourseStatisticRequest{
			CourseId:    c.SyllabusConfig.CourseID,
			StudyPlanId: c.SyllabusConfig.StudyPlanID,
			ClassId:     []string{},
			School: &spb.CourseStatisticRequest_Unassigned{
				Unassigned: false,
			},
		}
		syllabusClient.RetrieveCourseStatistic(contextWithToken(ctx, token), req)
		return err
	}
	scenario, err := j4.NewScenario("Syllabus_RetrieveCourseStatistic", *scenarioOpt)
	if err != nil {
		return nil, err
	}
	return scenario, nil
}
