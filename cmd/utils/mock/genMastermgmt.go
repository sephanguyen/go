package mock

import (
	"github.com/manabie-com/backend/internal/golibs/tools"
	academic_year_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/infrastructure/repo"
	class_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure/repo"
	config_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/infrastructure/repo"
	course_locationadapter "github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure/locationadapter"
	course_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure/repo"
	external_config_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/infrastructure/repo"
	grade_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/grade/infrastructure/repo"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	organization_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/organization/repositories"
	reserveClassDomain "github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/application/commands"
	schedule_class_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/infrastructure/repo"
	subject_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/subject/infrastructure/repo"
	time_slot_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/time_slot/infrastructure/repo"
	working_hours_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/working_hours/infrastructure/repo"

	"github.com/spf13/cobra"
)

func genmastermgmt(cmd *cobra.Command, args []string) error {
	tools.RemoveImport("github.com/manabie-com/backend/internal/mastermgmt/entities")
	tools.AddImport("github.com/manabie-com/backend/internal/mastermgmt/modules/organization/entities")

	tools.AddImport("github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain")
	tools.AddImport("github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure/repo")

	courseStruct := map[string][]interface{}{
		"internal/mastermgmt/modules/course/infrastructure/repo": {
			&course_repo.CourseAccessPathRepo{},
			&course_repo.StudentSubscriptionRepo{},
			&course_repo.CourseRepo{},
			&course_repo.CourseTypeRepo{},
		},
		"internal/mastermgmt/modules/course/infrastructure/location_adapter": {
			&course_locationadapter.LocationAdapter{},
		},
	}
	if err := tools.GenMockStructs(courseStruct); err != nil {
		return err
	}
	tools.RemoveImport("github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo")
	tools.RemoveImport("github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain")
	tools.RemoveImport("github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure/repo")

	structs := map[string][]interface{}{
		"internal/mastermgmt/modules/location/infrastructure/repo": {
			&location_repo.LocationRepo{},
			&location_repo.LocationTypeRepo{},
			&location_repo.ImportLogRepo{},
		},
		"internal/mastermgmt/modules/organization/repositories": {
			&organization_repo.OrganizationRepo{},
		},
		"internal/mastermgmt/modules/class/infrastructure/repo": {
			&class_repo.ClassRepo{},
			&class_repo.ClassMemberRepo{},
		},
		"internal/mastermgmt/modules/grade/infrastructure/repo": {
			&grade_repo.GradeRepo{},
		},
		"internal/mastermgmt/modules/subject/infrastructure/repo": {
			&subject_repo.SubjectRepo{},
		},
		"internal/mastermgmt/modules/configuration/infrastructure/repo": {
			&config_repo.ConfigRepo{},
		},
		"internal/mastermgmt/modules/external_configuration/infrastructure/repo": {
			&external_config_repo.ExternalConfigRepo{},
		},
		"internal/mastermgmt/modules/academic_year/infrastructure/repo": {
			&academic_year_repo.AcademicWeekRepo{},
			&academic_year_repo.AcademicYearRepo{},
			&academic_year_repo.AcademicClosedDayRepo{},
		},
		"internal/mastermgmt/modules/working_hours/infrastructure/repo": {
			&working_hours_repo.WorkingHoursRepo{},
		},
		"internal/mastermgmt/modules/time_slot/infrastructure/repo": {
			&time_slot_repo.TimeSlotRepo{},
		},
		"internal/mastermgmt/modules/schedule_class/infrastructure/repo": {
			&schedule_class_repo.ReserveClassRepo{},
			&schedule_class_repo.CourseRepo{},
			&schedule_class_repo.ClassRepo{},
			&schedule_class_repo.StudentPackageClassRepo{},
		},
		"internal/mastermgmt/modules/schedule_class/application/commands": {
			&reserveClassDomain.ReserveClassCommandHandler{},
		},
	}
	if err := tools.GenMockStructs(structs); err != nil {
		return err
	}
	return nil
}

func newGenMasterMgmtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mastermgmt [../../mock/mastermgmt]",
		Short: "generate mastermgmt repository type",
		Args:  cobra.NoArgs,
		RunE:  genmastermgmt,
	}
}
