package lesson

import (
	calendar_repo "github.com/manabie-com/backend/internal/calendar/infrastructure/repositories"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/controller"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	lesson_nats "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/nats"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	lesson_report_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/infrastructure/repo"
	masterdata_course_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/course/repository"
	masterdata_location_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/location/repository"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	user_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"
	zoom_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/infrastructure/repo"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/service"
	zoom_service "github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/service"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	pbv1 "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	plv1 "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc"
)

type ModuleWriter struct {
	LessonManagementGRPCService pbv1.LessonManagementServiceServer
	LessonModifierService       plv1.LessonModifierServiceServer
	LessonExecutorService       plv1.LessonExecutorServiceServer
}

type ModuleReader struct {
	LessonReaderService plv1.LessonReaderServiceServer
}

func NewModuleWriter(_ grpc.ServiceRegistrar, wrapperConnection *support.WrapperDBConnection, jsm nats.JetStreamManagement, userModule infrastructure.UserModulePort, mediaModule infrastructure.MediaModulePort, env string, unleashClientIns unleashclient.ClientInstance, zoomService service.ZoomServiceInterface, schedulerClient clients.SchedulerClientInterface) *ModuleWriter {
	m := &ModuleWriter{
		LessonManagementGRPCService: controller.NewLessonManagementGRPCService(
			wrapperConnection,
			&repo.LessonRepo{},
			userModule,
			mediaModule,
			&repo.LessonRoomStateRepo{},
		),
		LessonModifierService: controller.NewLessonModifierService(
			wrapperConnection,
			jsm,
			&repo.LessonRepo{},
			&repo.MasterDataRepo{},
			userModule,
			mediaModule,
			&calendar_repo.DateInfoRepo{},
			&repo.ClassroomRepo{},
			&lesson_report_repo.LessonReportRepo{},
			env,
			unleashClientIns,
			&calendar_repo.SchedulerRepo{},
			&user_repo.StudentSubscriptionRepo{},
			&repo.ReallocationRepo{},
			&repo.LessonMemberRepo{},
			zoomService,
			&zoom_repo.ZoomAccountRepo{},
			&user_repo.UserAccessPathRepo{},
			&repository.DomainEnrollmentStatusHistoryRepo{},
			schedulerClient,
			&lesson_nats.LessonPublisher{},
		),
		LessonExecutorService: controller.NewLessonExecutorService(
			wrapperConnection,
			&repo.ClassroomRepo{},
			&repo.LessonRepo{},
			jsm,
			&repo.MasterDataRepo{},
			userModule,
			&lesson_report_repo.LessonReportRepo{},
			env,
			unleashClientIns,
			&calendar_repo.SchedulerRepo{},
			&user_repo.StudentSubscriptionRepo{},
			&repo.ReallocationRepo{},
			&calendar_repo.DateInfoRepo{},
			*zoom_service.NewZoomAccountService(wrapperConnection, zoomService, &zoom_repo.ZoomAccountRepo{}, &repo.LessonRepo{}),
			&masterdata_location_repo.LocationRepository{},
			&masterdata_course_repo.CourseRepository{},
			&user_repo.UserBasicInfoRepo{},
			&user_repo.TeacherRepo{},
			schedulerClient,
			&repo.CourseRepo{},
		),
	}

	// TODO: move Lesson Management Service from old code base after
	// pbv1.RegisterLessonManagementServiceServer(s, m.LessonManagementGRPCService)

	return m
}

func NewModuleReader(_ grpc.ServiceRegistrar, wrapperConnection *support.WrapperDBConnection, env string, unleashClientIns unleashclient.ClientInstance) *ModuleReader {
	m := &ModuleReader{
		LessonReaderService: controller.NewLessonReaderService(
			wrapperConnection,
			&repo.LessonRepo{},
			nil,
			&repo.LessonTeacherRepo{},
			&repo.LessonMemberRepo{},
			&repo.LessonGroupRepo{},
			&repo.LessonClassroomRepo{},
			&user_repo.UserRepo{},
			env,
			unleashClientIns,
		),
	}
	return m
}
