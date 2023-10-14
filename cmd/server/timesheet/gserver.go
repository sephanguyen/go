package timesheet

import (
	"context"
	"fmt"
	"net/http"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/healthcheck"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/timesheet/configuration"
	"github.com/manabie-com/backend/internal/timesheet/controller"
	"github.com/manabie-com/backend/internal/timesheet/infrastructure/repository"
	importMasterData "github.com/manabie-com/backend/internal/timesheet/service/import_master_data"
	"github.com/manabie-com/backend/internal/timesheet/service/mastermgmt"
	timesheet_nats_service "github.com/manabie-com/backend/internal/timesheet/service/nats"
	"github.com/manabie-com/backend/internal/timesheet/service/timesheet"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/interceptors"
	pb_mastermgmt "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	pb_timesheet "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/gin-gonic/gin"
	gateway "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.opencensus.io/plugin/ocgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	health "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

func init() {
	s := &server{}
	bootstrap.
		WithGRPC[configuration.Config](s).
		WithNatsServicer(s).
		WithMonitorServicer(s).
		WithHTTP(s).
		Register(s)
}

type server struct {
	bootstrap.DefaultMonitorService[configuration.Config]
	authInterceptor                      *interceptors.Auth
	mastermgmtConn                       *grpc.ClientConn
	masterMgmtConfigurationServiceClient pb_mastermgmt.ConfigurationServiceClient
	masterMgmtInternalServiceClient      pb_mastermgmt.InternalServiceClient
}

func (s *server) ServerName() string {
	return "timesheet"
}

func (s *server) GracefulShutdown(context.Context) {}

func (s *server) RegisterNatsSubscribers(_ context.Context, c configuration.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	dbTrace := rsc.DB()
	jsm := rsc.NATS()
	unleashClient := rsc.Unleash()

	err := controller.RegisterLessonEventHandler(
		jsm,
		zapLogger,
		unleashClient,
		c.Common.Environment,
		&timesheet_nats_service.LessonNatsServiceImpl{
			DB:  dbTrace,
			JSM: jsm,
			GetTimesheetService: &timesheet.GetTimesheetServiceImpl{
				DB:                        dbTrace,
				TimesheetRepo:             &repository.TimesheetRepoImpl{},
				TimesheetLessonHoursRepo:  &repository.TimesheetLessonHoursRepoImpl{},
				OtherWorkingHoursRepo:     &repository.OtherWorkingHoursRepoImpl{},
				TransportationExpenseRepo: &repository.TransportationExpenseRepoImpl{},
			},
			AutoCreateTimesheetService: &timesheet.AutoCreateTimesheetServiceImpl{
				DB:                             dbTrace,
				TimesheetRepo:                  &repository.TimesheetRepoImpl{},
				TimesheetLessonHoursRepo:       &repository.TimesheetLessonHoursRepoImpl{},
				TransportationExpenseRepo:      &repository.TransportationExpenseRepoImpl{},
				StaffTransportationExpenseRepo: &repository.StaffTransportationExpenseRepoImpl{},
			},
			AutoCreateFlagActivityLogService: &timesheet.AutoCreateTimesheetFlagServiceImpl{
				DB:                    dbTrace,
				AutoCreateFlagLogRepo: &repository.AutoCreateFlagActivityLogRepoImpl{},
			},
			ConfirmationWindowService: &timesheet.ConfirmationWindowServiceImpl{
				DB:                                  dbTrace,
				TimesheetConfirmationPeriodRepo:     &repository.TimesheetConfirmationPeriodRepoImpl{},
				TimesheetConfirmationCutOffDateRepo: &repository.TimesheetConfirmationCutOffDateRepoImpl{},
				TimesheetConfirmationInfoRepo:       &repository.TimesheetConfirmationInfoRepoImpl{},
				TimesheetRepo:                       &repository.TimesheetRepoImpl{},
			},
			LessonRepo:                     &repository.LessonRepoImpl{},
			PartnerAutoCreateTimesheetRepo: &repository.PartnerAutoCreateTimesheetFlagRepoImpl{},
		},
		&mastermgmt.MasterConfigurationServiceImpl{
			MasterMgmtInternalServiceClient: s.masterMgmtInternalServiceClient,
		},
	)
	if err != nil {
		zapLogger.Fatal("RegisterLessonEventHandler failed, err: %v", zap.Error(err))
	}

	err = controller.RegisterTimesheetActionLogEventHandler(
		jsm,
		zapLogger,
		unleashClient,
		c.Common.Environment,
		&timesheet.ActionLogServiceImpl{
			DB:            dbTrace,
			ActionLogRepo: &repository.TimesheetActionLogRepoImpl{},
		},
	)
	if err != nil {
		zapLogger.Fatal("RegisterTimesheetActionLogEventHandler failed, err: %v", zap.Error(err))
	}

	err = controller.RegisterTimesheetAutoCreateFlagEventHandler(
		jsm,
		zapLogger,
		&timesheet.AutoCreateTimesheetFlagServiceImpl{
			DB:                       dbTrace,
			AutoCreateFlagRepo:       &repository.AutoCreateFlagRepoImpl{},
			AutoCreateFlagLogRepo:    &repository.AutoCreateFlagActivityLogRepoImpl{},
			TimesheetRepo:            &repository.TimesheetRepoImpl{},
			TimesheetLessonHoursRepo: &repository.TimesheetLessonHoursRepoImpl{},
			OtherWorkingHoursRepo:    &repository.OtherWorkingHoursRepoImpl{},
		},
	)
	if err != nil {
		zapLogger.Fatal("RegisterTimesheetAutoCreateFlagEventHandler failed, err: %v", zap.Error(err))
	}

	return nil
}

func (s *server) InitDependencies(c configuration.Config, rsc *bootstrap.Resources) error {
	s.authInterceptor = authInterceptor(&c, rsc.Logger(), rsc.DB())
	s.mastermgmtConn = rsc.GRPCDial("mastermgmt")
	s.masterMgmtConfigurationServiceClient = pb_mastermgmt.NewConfigurationServiceClient(s.mastermgmtConn)
	s.masterMgmtInternalServiceClient = pb_mastermgmt.NewInternalServiceClient(s.mastermgmtConn)

	return nil
}

func (s *server) SetupGRPC(_ context.Context, grpcserv *grpc.Server, c configuration.Config, rsc *bootstrap.Resources) error {
	dbTrace := rsc.DB()
	jsm := rsc.NATS()

	health.RegisterHealthServer(grpcserv, &healthcheck.Service{DB: dbTrace.DB.(*pgxpool.Pool)})
	// Register: Timesheet Service
	pb_timesheet.RegisterTimesheetServiceServer(
		grpcserv,
		&controller.TimesheetServiceController{
			TimesheetService: &timesheet.ServiceImpl{
				DB:                        dbTrace,
				JSM:                       jsm,
				TimesheetRepo:             &repository.TimesheetRepoImpl{},
				OtherWorkingHoursRepo:     &repository.OtherWorkingHoursRepoImpl{},
				TransportationExpenseRepo: &repository.TransportationExpenseRepoImpl{},
				TimesheetLessonHoursRepo:  &repository.TimesheetLessonHoursRepoImpl{},
				GetTimesheetService: &timesheet.GetTimesheetServiceImpl{
					DB:                        dbTrace,
					TimesheetRepo:             &repository.TimesheetRepoImpl{},
					TimesheetLessonHoursRepo:  &repository.TimesheetLessonHoursRepoImpl{},
					OtherWorkingHoursRepo:     &repository.OtherWorkingHoursRepoImpl{},
					TransportationExpenseRepo: &repository.TransportationExpenseRepoImpl{},
				},
			},
			MastermgmtConfigurationService: &mastermgmt.MasterConfigurationServiceImpl{
				MasterMgmtConfigurationServiceClient: s.masterMgmtConfigurationServiceClient,
			},
			ConfirmationWindowService: &timesheet.ConfirmationWindowServiceImpl{
				DB:                                  dbTrace,
				TimesheetConfirmationPeriodRepo:     &repository.TimesheetConfirmationPeriodRepoImpl{},
				TimesheetConfirmationCutOffDateRepo: &repository.TimesheetConfirmationCutOffDateRepoImpl{},
				TimesheetConfirmationInfoRepo:       &repository.TimesheetConfirmationInfoRepoImpl{},
				TimesheetLocationListRepo:           &repository.TimesheetLocationListRepoImpl{},
				TimesheetRepo:                       &repository.TimesheetRepoImpl{},
			},
		},
	)

	pb_timesheet.RegisterImportMasterDataServiceServer(
		grpcserv,
		&controller.ImportMasterDataController{
			ImportTimesheetConfigService: &importMasterData.ImportTimesheetConfigService{
				DB:                  dbTrace,
				TimesheetConfigRepo: &repository.TimesheetConfigRepoImpl{},
			},
		},
	)

	pb_timesheet.RegisterTimesheetStateMachineServiceServer(
		grpcserv,
		&controller.TimesheetStateMachineController{
			TimesheetStateMachineService: &timesheet.TimesheetStateMachineService{
				DB:                        dbTrace,
				JSM:                       jsm,
				TimesheetRepo:             &repository.TimesheetRepoImpl{},
				TimesheetLessonHoursRepo:  &repository.TimesheetLessonHoursRepoImpl{},
				OtherWorkingHoursRepo:     &repository.OtherWorkingHoursRepoImpl{},
				LessonRepo:                &repository.LessonRepoImpl{},
				TransportationExpenseRepo: &repository.TransportationExpenseRepoImpl{},
			},
			MastermgmtConfigurationService: &mastermgmt.MasterConfigurationServiceImpl{
				MasterMgmtConfigurationServiceClient: s.masterMgmtConfigurationServiceClient,
			},
		},
	)

	pb_timesheet.RegisterAutoCreateTimesheetServiceServer(
		grpcserv,
		&controller.AutoCreateTimesheetFlagController{
			JSM: jsm,
			AutoCreateFlagService: &timesheet.AutoCreateTimesheetFlagServiceImpl{
				DB:                       dbTrace,
				AutoCreateFlagRepo:       &repository.AutoCreateFlagRepoImpl{},
				AutoCreateFlagLogRepo:    &repository.AutoCreateFlagActivityLogRepoImpl{},
				TimesheetRepo:            &repository.TimesheetRepoImpl{},
				TimesheetLessonHoursRepo: &repository.TimesheetLessonHoursRepoImpl{},
				OtherWorkingHoursRepo:    &repository.OtherWorkingHoursRepoImpl{},
			},
			MastermgmtConfigurationService: &mastermgmt.MasterConfigurationServiceImpl{
				MasterMgmtConfigurationServiceClient: s.masterMgmtConfigurationServiceClient,
			},
		},
	)

	pb_timesheet.RegisterStaffTransportationExpenseServiceServer(
		grpcserv,
		&controller.StaffTransportationExpenseController{
			StaffTransportationExpenseService: &timesheet.StaffTransportationExpenseServiceImpl{
				DB:                             dbTrace,
				JSM:                            jsm,
				StaffTransportationExpenseRepo: &repository.StaffTransportationExpenseRepoImpl{},
				TimesheetRepo:                  &repository.TimesheetRepoImpl{},
				TransportationExpenseRepo:      &repository.TransportationExpenseRepoImpl{},
			},
		},
	)

	pb_timesheet.RegisterTimesheetConfirmationServiceServer(
		grpcserv,
		&controller.TimesheetConfirmationController{
			TimesheetConfirmationWindowService: &timesheet.ConfirmationWindowServiceImpl{
				DB:                                  dbTrace,
				JSM:                                 jsm,
				TimesheetConfirmationPeriodRepo:     &repository.TimesheetConfirmationPeriodRepoImpl{},
				TimesheetConfirmationCutOffDateRepo: &repository.TimesheetConfirmationCutOffDateRepoImpl{},
				TimesheetConfirmationInfoRepo:       &repository.TimesheetConfirmationInfoRepoImpl{},
				TimesheetLocationListRepo:           &repository.TimesheetLocationListRepoImpl{},
				TimesheetRepo:                       &repository.TimesheetRepoImpl{},
			},
		},
	)

	pb_timesheet.RegisterLocationServiceServer(
		grpcserv,
		&controller.LocationController{
			LocationService: &timesheet.LocationServiceImpl{
				DB:           dbTrace,
				LocationRepo: &repository.LocationRepoImpl{},
			},
		},
	)

	return nil
}

func (s *server) SetupHTTP(c configuration.Config, r *gin.Engine, rsc *bootstrap.Resources) error {
	mux := gateway.NewServeMux(
		gateway.WithOutgoingHeaderMatcher(clients.IsHeaderAllowed),
		gateway.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			authHeader := request.Header.Get("Authorization")
			// need to forward these to prevent errors from github.com/manabie-com/backend/internal/timesheet/domain/common.SignCtx
			pkgHeader := request.Header.Get("pkg")
			versionHeader := request.Header.Get("version")

			// add defaults for pkg header
			if pkgHeader == "" {
				pkgHeader = "com.manabie.liz"
			}

			// add defaults for version header
			if versionHeader == "" {
				versionHeader = "1.0.0"
			}

			md := metadata.Pairs(
				"token", authHeader,
				"pkg", pkgHeader,
				"version", versionHeader,
			)
			return md
		}),
		gateway.WithErrorHandler(func(ctx context.Context, mux *gateway.ServeMux, marshaler gateway.Marshaler, writer http.ResponseWriter, request *http.Request, err error) {
			newError := gateway.HTTPStatusError{
				HTTPStatus: 400,
				Err:        err,
			}
			gateway.DefaultHTTPErrorHandler(ctx, mux, marshaler, writer, request, &newError)
		}))

	err := setupGrpcGateway(mux, rsc.GetGRPCPort(c.Common.Name))
	if err != nil {
		return fmt.Errorf("error setupGrpcGateway %s", err)
	}

	superGroup := r.Group("/timesheet/api/v1")
	{
		superGroup.Group("/proxy/*{grpc_gateway}").Any("", gin.WrapH(mux))
	}
	return nil
}

func setupGrpcGateway(mux *gateway.ServeMux, port string) error {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(&tracer.B3Handler{ClientHandler: &ocgrpc.ClientHandler{}}),
	}

	serviceMap := map[string]func(context.Context, *gateway.ServeMux, string, []grpc.DialOption) error{
		"TimesheetService": pb_timesheet.RegisterTimesheetServiceHandlerFromEndpoint,
	}

	for _, registerFunc := range serviceMap {
		err := registerFunc(context.Background(), mux, fmt.Sprintf("localhost%s", port), dialOpts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *server) WithUnaryServerInterceptors(c configuration.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	customs := []grpc.UnaryServerInterceptor{
		s.authInterceptor.UnaryServerInterceptor,
		tracer.UnaryActivityLogRequestInterceptor(rsc.NATS(), rsc.Logger(), s.ServerName()),
	}
	grpcUnary := bootstrap.DefaultUnaryServerInterceptor(rsc)
	grpcUnary = append(grpcUnary, customs...)

	return grpcUnary
}

func (s *server) WithStreamServerInterceptors(c configuration.Config, rsc *bootstrap.Resources) []grpc.StreamServerInterceptor {
	grpcStream := bootstrap.DefaultStreamServerInterceptor(rsc)
	grpcStream = append(grpcStream, s.authInterceptor.StreamServerInterceptor)
	return grpcStream
}

func (s *server) WithServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}
}
