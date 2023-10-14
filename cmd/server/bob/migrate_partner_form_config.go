package bob

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"

	"go.uber.org/zap"
)

func init() {
	bootstrap.RegisterJob("bob_migrate_partner_form_configs", runMigratePartnerFormConfigs)
}

func runMigratePartnerFormConfigs(ctx context.Context, bobCfg configurations.Config, rsc *bootstrap.Resources) error {
	l := rsc.Logger()
	bobDB := rsc.DB()

	orgGaQuery := "select organization_id, name from organizations"
	organizations, err := bobDB.Query(ctx, orgGaQuery)
	if err != nil {
		return fmt.Errorf("failed to get orgs: %s", err)
	}
	defer organizations.Close()
	// todo: loop publish for each org
	for organizations.Next() {
		var organizationID, name string
		err := organizations.Scan(&organizationID, &name)
		if err != nil {
			return fmt.Errorf("failed to scan an orgs row: %s", err)
		}
		ctx = auth.InjectFakeJwtToken(ctx, organizationID)
		total, err := migratePartnerFormConfigs(ctx, l, &bobCfg, organizationID, bobDB)
		if err != nil {
			return err
		}
		l.Sugar().Infof("There is/are %d new partner_form_configs migrated from org %s. ", total, name)
	}
	return nil
}

func migratePartnerFormConfigs(ctx context.Context, l *zap.Logger, c *configurations.Config, organizationID string, bobConn database.Ext) (int, error) {
	if c.Common.Environment != "prod" && c.Common.Environment != "uat" && organizationID == "" {
		return 0, fmt.Errorf("running in non (production/uat) requires a school id")
	}

	queries := getQueries(organizationID)
	total := 0
	if len(queries) > 0 {
		for _, query := range queries {
			_, err := bobConn.Exec(ctx, query)
			if err != nil {
				l.Sugar().Errorf("failed to insert partner_form_configs: %s", err)
				continue
			}
			total++
		}
	}
	return total, nil
}

func getQueries(organizationID string) (queries []string) {
	switch organizationID {
	case fmt.Sprintf("%v", constants.GASchool):
		// prod GA got wrong ID, therefore we must set this ID manually
		gaIDProdString := "-2147483644"
		return []string{fmt.Sprintf(`INSERT INTO public.partner_form_configs (form_config_id,partner_id,feature_name,created_at,updated_at,deleted_at,form_config_data,resource_path) VALUES
		('01FTCP0VPV85CV5C5RH7FKQ1BW',%d,'FEATURE_NAME_INDIVIDUAL_LESSON_REPORT',now(),now(),NULL,'{"sections": [{"fields": [{"label": {"i18n": {"translations": {"en": "Attendance", "ja": "出席情報", "vi": "Attendance"}, "fallback_language": "ja"}}, "field_id": "attendance_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "attendance_status", "value_type": "VALUE_TYPE_STRING", "is_required": true, "display_config": {"size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_STATUS"}}, {"field_id": "attendance_remark", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_config": {"type": "ATTENDANCE_REMARK"}}], "section_id": "attendance_section_id", "section_name": "attendance"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Homework Submission", "ja": "課題", "vi": "Homework Submission"}, "fallback_language": "ja"}}, "field_id": "homework_submission_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Homework Status", "ja": "提出状況", "vi": "Homework Status"}, "fallback_language": "ja"}}, "field_id": "homework_submission_status", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"options": [{"key": "COMPLETED", "label": {"i18n": {"translations": {"en": "Completed", "ja": "完了", "vi": "Completed"}, "fallback_language": "ja"}}}, {"key": "INCOMPLETE", "label": {"i18n": {"translations": {"en": "Incomplete", "ja": "未完了", "vi": "Incomplete"}, "fallback_language": "ja"}}}], "optionLabelKey": "label"}, "component_config": {"type": "AUTOCOMPLETE"}}], "section_id": "homework_submission_section_id", "section_name": "homework_submission"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Lesson", "ja": "授業", "vi": "Lesson"}, "fallback_language": "ja"}}, "field_id": "lesson_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"field_id": "lesson_view_study_plan_action", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 3, "xs": 3}}, "component_config": {"type": "LINK_VIEW_STUDY_PLAN"}}, {"field_id": "lesson_previous_report_action", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 3, "xs": 3}}, "component_config": {"type": "BUTTON_PREVIOUS_REPORT"}}, {"label": {"i18n": {"translations": {"en": "Content", "ja": "追加教材", "vi": "Content"}, "fallback_language": "ja"}}, "field_id": "lesson_content", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}, {"label": {"i18n": {"translations": {"en": "Homework", "ja": "追加課題", "vi": "Homework"}, "fallback_language": "ja"}}, "field_id": "lesson_homework", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 6, "xs": 6}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "lesson_section_id", "section_name": "lesson"}, {"fields": [{"label": {"i18n": {"translations": {"en": "Remarks", "ja": "備考", "vi": "Remarks"}, "fallback_language": "ja"}}, "field_id": "remarks_section_label", "value_type": "VALUE_TYPE_NULL", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"variant": "body2"}, "component_config": {"type": "TYPOGRAPHY"}}, {"label": {"i18n": {"translations": {"en": "Remarks", "ja": "備考", "vi": "Remarks"}, "fallback_language": "ja"}}, "field_id": "remarks", "value_type": "VALUE_TYPE_STRING", "is_required": false, "display_config": {"size": {"md": 12, "xs": 12}}, "component_props": {"InputProps": {"rows": 6, "multiline": true}}, "component_config": {"type": "TEXT_FIELD_AREA"}}], "section_id": "remarks_section_id", "section_name": "remarks"}]}',%s) ON CONFLICT ON CONSTRAINT partner_form_configs_pk DO NOTHING;`,
			constants.TestingSchool, gaIDProdString)}
	default:
		return []string{}
	}
}
