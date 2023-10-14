package mock

import (
	"path/filepath"

	"github.com/manabie-com/backend/internal/golibs/tools"
	lesson_allocation_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/infrastructure/repo"
	asg_student_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/infrastructure/repo"
	class_do_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/infrastructure/repo"
	course_location_schedule_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/infrastructure/repo"
	elastic_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/elasticsearch"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/mediaadapter"
	nats_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/nats"
	lesson_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/usermodadapter"
	lesson_report_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/infrastructure/repo"
	masterdata_academic_week_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/academic_week/repository"
	masterdata_academic_year_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/academic_year/repository"
	class_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/class/repository"
	masterdata_course_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/course/repository"
	masterdata_location_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/location/repository"
	media_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/media/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"

	"github.com/spf13/cobra"
)

func genLessonmgmtRepo(cmd *cobra.Command, args []string) error {
	assignedStudentRepo := map[string]interface{}{
		"assigned_student_repo": &asg_student_repo.AssignedStudentRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "assigned_student/repositories"), "asgstudent", assignedStudentRepo)

	userModuleRepos := map[string]interface{}{
		"teacher_repo":                     &repo.TeacherRepo{},
		"student_subscription_repo":        &repo.StudentSubscriptionRepo{},
		"user_repo":                        &repo.UserRepo{},
		"student_subscription_access_path": &repo.StudentSubscriptionAccessPathRepo{},
		"user_access_path":                 &repo.UserAccessPathRepo{},
		"user_basic_info":                  &repo.UserBasicInfo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "user/repositories"), "usermgmt", userModuleRepos)

	lessonModuleRepos := map[string]interface{}{
		"lesson_repo":            &lesson_repo.LessonRepo{},
		"lesson_group_repo":      &lesson_repo.LessonGroupRepo{},
		"master_data_repo":       &lesson_repo.MasterDataRepo{},
		"lesson_room_state_repo": &lesson_repo.LessonRoomStateRepo{},
		"course_repo":            &lesson_repo.CourseRepo{},
		"lesson_teacher":         &lesson_repo.LessonTeacherRepo{},
		"lesson_member":          &lesson_repo.LessonMemberRepo{},
		"lesson_classroom":       &lesson_repo.LessonClassroomRepo{},
		"classroom":              &lesson_repo.ClassroomRepo{},
		"reallocation":           &lesson_repo.ReallocationRepo{},
	}

	tools.MockRepository("mock_repositories", filepath.Join(args[0], "lesson/repositories"), "usermgmt", lessonModuleRepos)

	searchModuleRepos := map[string]interface{}{
		"search_repo": &elastic_repo.SearchRepo{},
	}
	tools.MockRepository("mock_elasticsearch", filepath.Join(args[0], "lesson/elasticsearch"), "lessonmgmt", searchModuleRepos)
	userModuleAdapter := map[string]interface{}{
		"user_module_adapter": &usermodadapter.UserModuleAdapter{},
	}
	tools.MockRepository("mock_user_module_adapter", filepath.Join(args[0], "lesson/usermodadapter"), "usermgmt", userModuleAdapter)

	mediaModule := map[string]interface{}{
		"media_module": &mediaadapter.MediaModuleAdapter{},
		"media_repo":   &media_repo.MediaRepo{},
	}
	tools.MockRepository("mock_media_module", filepath.Join(args[0], "lesson/media_module_adapter"), "lessonmgmt", mediaModule)

	lessonReportRepos := map[string]interface{}{
		"lesson_report_repo":        &lesson_report_repo.LessonReportRepo{},
		"lesson_report_detail_repo": &lesson_report_repo.LessonReportDetailRepo{},
		"partner_form_config_repo":  &lesson_report_repo.PartnerFormConfigRepo{},
	}
	tools.MockRepository("mock_lesson_report", filepath.Join(args[0], "lesson_report/repositories"), "lessonmgmt", lessonReportRepos)

	masterDataRepo := map[string]interface{}{
		"location_repo": &masterdata_location_repo.LocationRepository{},
		"course_repo":   &masterdata_course_repo.CourseRepository{},
		"academic_week": &masterdata_academic_week_repo.AcademicWeekRepository{},
		"academic_year": &masterdata_academic_year_repo.AcademicYearRepository{},
		"class":         &class_repo.ClassRepository{},
	}
	tools.MockRepository("mock_master_data", filepath.Join(args[0], "master_data/repositories"), "lessonmgmt", masterDataRepo)

	natRepo := map[string]interface{}{
		"lesson_publisher": &nats_repo.LessonPublisher{},
	}
	tools.MockRepository("mock_nats", filepath.Join(args[0], "lesson/nats"), "lessonmgmt", natRepo)
	lessonAllocationRepo := map[string]interface{}{
		"lesson_allocation_repo": &lesson_allocation_repo.LessonAllocationRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "lesson_allocation/repositories"), "lessonmgmt", lessonAllocationRepo)

	courseLocationScheduleRepo := map[string]interface{}{
		"course_location_schedule_repo": &course_location_schedule_repo.CourseLocationScheduleRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "course_location_schedule/repositories"), "lessonmgmt", courseLocationScheduleRepo)

	classDoRepo := map[string]interface{}{
		"classdo_account_repo": &class_do_repo.ClassDoAccountRepo{},
	}
	tools.MockRepository("mock_repositories", filepath.Join(args[0], "classdo/repositories"), "lessonmgmt", classDoRepo)

	return nil
}

func newGenLessonmgmtCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lessonmgmt [../../mock/lessonmgmt]",
		Short: "generate lessonmgmt repository type",
		Args:  cobra.ExactArgs(1),
		RunE:  genLessonmgmtRepo,
	}
}
