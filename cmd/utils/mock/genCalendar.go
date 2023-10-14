package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/calendar/application/command"
	"github.com/manabie-com/backend/internal/calendar/application/queries"
	"github.com/manabie-com/backend/internal/calendar/infrastructure/repositories"
	"github.com/manabie-com/backend/internal/golibs/tools"

	"github.com/spf13/cobra"
)

func genCalendarRepo(cmd *cobra.Command, args []string) error {
	calendarRepos := map[string]interface{}{
		"date_info_repo":        &repositories.DateInfoRepo{},
		"date_type_repo":        &repositories.DateTypeRepo{},
		"location_repo":         &repositories.LocationRepo{},
		"scheduler_repo":        &repositories.SchedulerRepo{},
		"user_repo":             &repositories.UserRepo{},
		"lesson_repo":           &repositories.LessonRepo{},
		"lesson_member_repo":    &repositories.LessonMemberRepo{},
		"lesson_teacher_repo":   &repositories.LessonTeacherRepo{},
		"lesson_classroom_repo": &repositories.LessonClassroomRepo{},
		"lesson_group_repo":     &repositories.LessonGroupRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "repositories"), "calendar", calendarRepos)

	calendarCmd := map[string]interface{}{
		"upsert_date_info_command": &command.UpsertDateInfoCommand{},
		"create_scheduler_command": &command.CreateSchedulerCommand{},
		"update_scheduler_command": &command.UpdateSchedulerCommand{},
	}
	tools.MockRepository("mock_application", filepath.Join(args[0], "application/command"), "calendar", calendarCmd)

	calendarQry := map[string]interface{}{
		"date_info_query_handler": &queries.DateInfoQueryHandler{},
		"lesson_query_handler":    &queries.LessonQueryHandler{},
		"staff_query_handler":     &queries.GetStaff{},
	}
	tools.MockRepository("mock_application", filepath.Join(args[0], "application/queries"), "calendar", calendarQry)

	return nil
}

func newGenCalendarCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "calendar [../../mock/calendar]",
		Short: "generate calendar repository type",
		Args:  cobra.ExactArgs(1),
		RunE:  genCalendarRepo,
	}
}
