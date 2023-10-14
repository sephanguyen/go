package communication

import (
	"context"

	"github.com/manabie-com/backend/j4/infras"
	serviceutil "github.com/manabie-com/backend/j4/serviceutil"
	j4 "github.com/manabie-com/j4/pkg/runner"
)

var (
	Communication_GetListInfoNotifications = `
query Communication_GetListInfoNotifications($status: String, $limit:
          Int, $offset: Int) {
            info_notifications(limit: $limit, offset: $offset, order_by: {updated_at: desc}, where: {status: {_eq: $status}}) {
              ...InfoNotificationsAttrs
            }
          }


          fragment InfoNotificationsAttrs on info_notifications {
            notification_id
            notification_msg_id
            sent_at
            receiver_ids
            status
            type
            target_groups
            updated_at
            created_at
            editor_id
            event
            scheduled_at
          }
`
)

func setupNoti(ctx context.Context, cfg *infras.ManabieJ4Config, dep *infras.Dep) (*j4.Scenario, error) {
	scenarioConf, err := cfg.GetScenarioConfig("noti_filter")
	if err != nil {
		return nil, err
	}

	tokenGenerator := serviceutil.NewTokenGenerator(cfg, dep.Connections)

	tok, err := tokenGenerator.GetTokenFromShamir(ctx, cfg.AdminID, cfg.SchoolID)
	if err != nil {
		return nil, err
	}
	scenarioOpt := infras.MustOptionFromConfig(&scenarioConf)
	scenarioOpt.TestFunc = func(ctx context.Context) error {
		_, err := dep.GetHasura("bob").QueryRawHasuraV1(ctx, tok, "Communication_GetListInfoNotifications", Communication_GetListInfoNotifications, map[string]interface{}{
			"limit":  10,
			"offset": 0,
		})
		if err != nil {
			return err
		}
		return nil
	}
	return j4.NewScenario("noti_filter", *scenarioOpt)
}
