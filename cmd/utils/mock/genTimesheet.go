package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/golibs/tools"
	"github.com/manabie-com/backend/internal/timesheet/infrastructure/repository"
	importMasterData "github.com/manabie-com/backend/internal/timesheet/service/import_master_data"
	"github.com/manabie-com/backend/internal/timesheet/service/mastermgmt"
	"github.com/manabie-com/backend/internal/timesheet/service/timesheet"

	"github.com/spf13/cobra"
)

func genTimesheetRepo(cmd *cobra.Command, args []string) error {
	repos := map[string]interface{}{
		"timesheet":                           &repository.TimesheetRepoImpl{},
		"other_working_hours":                 &repository.OtherWorkingHoursRepoImpl{},
		"timesheet_lesson_hours":              &repository.TimesheetLessonHoursRepoImpl{},
		"timesheet_config":                    &repository.TimesheetConfigRepoImpl{},
		"lesson":                              &repository.LessonRepoImpl{},
		"auto_create_timesheet_flag":          &repository.AutoCreateFlagRepoImpl{},
		"transportation_expense":              &repository.TransportationExpenseRepoImpl{},
		"auto_create_timesheet_flag_log":      &repository.AutoCreateFlagActivityLogRepoImpl{},
		"staff_transportation_expense":        &repository.StaffTransportationExpenseRepoImpl{},
		"timesheet_confirmation_cut_off_date": &repository.TimesheetConfirmationCutOffDateRepoImpl{},
		"timesheet_confirmation_period":       &repository.TimesheetConfirmationPeriodRepoImpl{},
		"timesheet_confirmation_info":         &repository.TimesheetConfirmationInfoRepoImpl{},
		"timesheet_location_list":             &repository.TimesheetLocationListRepoImpl{},
		"partner_auto_create_timesheet_flag":  &repository.PartnerAutoCreateTimesheetFlagRepoImpl{},
		"timesheet_action_log":                &repository.TimesheetActionLogRepoImpl{},
		"location":                            &repository.LocationRepoImpl{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repository"), "timesheet", repos)

	structs := map[string][]interface{}{
		"internal/timesheet/services/timesheet":                   {&timesheet.ServiceImpl{}},
		"internal/timesheet/services/import_master_data":          {&importMasterData.ImportTimesheetConfigService{}},
		"internal/timesheet/service/gettimesheet":                 {&timesheet.GetTimesheetServiceImpl{}},
		"internal/timesheet/service/autocreatetimesheet":          {&timesheet.AutoCreateTimesheetServiceImpl{}},
		"internal/timesheet/service/autocreatetimesheetflag":      {&timesheet.AutoCreateTimesheetFlagServiceImpl{}},
		"internal/timesheet/service/timesheet_state_machine":      {&timesheet.TimesheetStateMachineService{}},
		"internal/timesheet/service/mastermgmt":                   {&mastermgmt.MasterConfigurationServiceImpl{}},
		"internal/timesheet/service/staff_transportation_expense": {&timesheet.StaffTransportationExpenseServiceImpl{}},
		"internal/timesheet/service/timesheet_confirmation":       {&timesheet.ConfirmationWindowServiceImpl{}},
		"internal/timesheet/service/timesheet_action_log":         {&timesheet.ActionLogServiceImpl{}},
		"internal/timesheet/service/location":                     {&timesheet.LocationServiceImpl{}},
	}

	if err := tools.GenMockStructs(structs); err != nil {
		return err
	}

	return nil
}

func newGenTimesheetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "timesheet [../../mock/timesheet]",
		Short: "generate timesheet repository type",
		Args:  cobra.ExactArgs(1),
		RunE:  genTimesheetRepo,
	}
}
