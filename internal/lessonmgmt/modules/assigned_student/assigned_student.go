package user

import (
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/controller"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/infrastructure/repo"
	lesson_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	academic_year_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/academic_year/repository"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"
	lbv1 "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc"
)

type Module struct {
	AssignedStudentGRPCService lbv1.AssignedStudentListServiceServer
}

func New(_ grpc.ServiceRegistrar, wrapperConnection *support.WrapperDBConnection, env string, unleashClientIns unleashclient.ClientInstance) *Module {
	m := &Module{
		AssignedStudentGRPCService: controller.NewAssignedStudentGRPCService(
			wrapperConnection,
			&repo.AssignedStudentRepo{},
			&lesson_repo.ReallocationRepo{},
			&user_repo.StudentSubscriptionRepo{},
			env,
			unleashClientIns,
			&academic_year_repo.AcademicYearRepository{},
		),
	}
	return m
}
