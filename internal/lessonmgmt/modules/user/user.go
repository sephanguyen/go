package user

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/controller"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"
	masterClassRepo "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure/repo"
	lbv1 "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc"
)

type Module struct {
	wrapperConnection                        *support.WrapperDBConnection
	UserGRPCService                          lbv1.UserServiceServer
	StudentSubscriptionGRPCLessonmgmtService lbv1.StudentSubscriptionServiceServer
}

func New(_ grpc.ServiceRegistrar, bobDB database.Ext, wrapperConnection *support.WrapperDBConnection, env string, unleashClientIns unleashclient.ClientInstance) *Module {
	m := &Module{
		wrapperConnection: wrapperConnection,
		UserGRPCService:   controller.NewUserGRPCService(bobDB, wrapperConnection, &repo.TeacherRepo{}, &repo.UserRepo{}, &repo.UserBasicInfoRepo{}),
		StudentSubscriptionGRPCLessonmgmtService: controller.NewStudentSubscriptionGRPCService(
			wrapperConnection,
			&repo.StudentSubscriptionRepo{},
			&repo.StudentSubscriptionAccessPathRepo{},
			&masterClassRepo.ClassMemberRepo{},
			&masterClassRepo.ClassRepo{},
			env,
			unleashClientIns,
		),
	}

	// TODO: move Student Subscription Service from old code base after
	// pbv1.RegisterStudentSubscriptionServiceServer(s, m.StudentSubscriptionGRPCService)

	return m
}
