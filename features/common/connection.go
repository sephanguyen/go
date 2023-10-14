package common

import (
	"context"
	"fmt"

	internal_auth "github.com/manabie-com/backend/internal/golibs/auth"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/nats"

	"firebase.google.com/go/auth"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var rsc = bootstrap.NewResources().WithLoggerC(&configs.CommonConfig{
	Name:              "gandalf",
	Environment:       "local",
	ActualEnvironment: "local",
	Organization:      "manabie",
})

type Connections struct {
	FirebaseAddr string
	ApplicantID  string
	Logger       *zap.Logger

	BobConn                         *grpc.ClientConn
	TomConn                         *grpc.ClientConn
	YasuoConn                       *grpc.ClientConn
	EurekaConn                      *grpc.ClientConn
	FatimaConn                      *grpc.ClientConn
	ShamirConn                      *grpc.ClientConn
	UserMgmtConn                    *grpc.ClientConn
	PaymentConn                     *grpc.ClientConn
	EntryExitMgmtConn               *grpc.ClientConn
	InvoiceMgmtConn                 *grpc.ClientConn
	MasterMgmtConn                  *grpc.ClientConn
	LessonMgmtConn                  *grpc.ClientConn
	EnigmaConn                      *grpc.ClientConn
	VirtualClassroomConn            *grpc.ClientConn
	VirtualClassroomHTTPConn        *grpc.ClientConn
	CalendarConn                    *grpc.ClientConn
	TimesheetConn                   *grpc.ClientConn
	NotificationMgmtConn            *grpc.ClientConn
	SpikeConn                       *grpc.ClientConn
	DiscountConn                    *grpc.ClientConn
	ConversationMgmtConn            *grpc.ClientConn
	AuthConn                        *grpc.ClientConn
	BobDB                           *pgxpool.Pool
	TomDB                           *pgxpool.Pool
	EurekaDB                        *pgxpool.Pool
	FatimaDB                        *pgxpool.Pool
	ZeusDB                          *pgxpool.Pool
	BobPostgresDB                   *pgxpool.Pool
	AuthPostgresDB                  *pgxpool.Pool
	TomPostgresDB                   *pgxpool.Pool
	InvoiceMgmtDB                   *pgxpool.Pool
	MasterMgmtDB                    *pgxpool.Pool
	MasterMgmtPostgresDB            *pgxpool.Pool
	EntryExitMgmtDB                 *pgxpool.Pool
	TimesheetDB                     *pgxpool.Pool
	CalendarDB                      *pgxpool.Pool
	LessonmgmtDB                    *pgxpool.Pool
	NotificationMgmtDB              *pgxpool.Pool
	NotificationMgmtPostgresDB      *pgxpool.Pool
	InvoiceMgmtPostgresDB           *pgxpool.Pool
	ConversationMgmtDB              *pgxpool.Pool
	ConversationMgmtPostgresDB      *pgxpool.Pool
	InvoiceMgmtDBTrace              *database.DBTrace
	InvoiceMgmtPostgresDBTrace      *database.DBTrace
	BobDBTrace                      *database.DBTrace
	TomDBTrace                      *database.DBTrace
	EurekaDBTrace                   *database.DBTrace
	FatimaDBTrace                   *database.DBTrace
	ZeusDBTrace                     *database.DBTrace
	BobPostgresDBTrace              *database.DBTrace
	AuthPostgresDBTrace             *database.DBTrace
	MasterMgmtDBTrace               *database.DBTrace
	MasterMgmtPostgresDBTrace       *database.DBTrace
	EntryExitMgmtDBTrace            *database.DBTrace
	TimesheetDBTrace                *database.DBTrace
	CalendarDBTrace                 *database.DBTrace
	LessonmgmtDBTrace               *database.DBTrace
	NotificationMgmtDBTrace         *database.DBTrace
	NotificationMgmtPostgresDBTrace *database.DBTrace
	ConversationMgmtDBTrace         *database.DBTrace
	ConversationMgmtPostgresDBTrace *database.DBTrace
	GCPApp                          *gcp.App
	FirebaseAuthClient              internal_auth_tenant.TenantClient
	FirebaseClient                  *auth.Client
	TenantManager                   internal_auth_tenant.TenantManager
	KeycloakClient                  *internal_auth.IdentityServiceImpl
	Kafka                           kafka.KafkaManagement
	JSM                             nats.JetStreamManagement
	YasuoDB                         *pgxpool.Pool
	YasuoDBTrace                    *database.DBTrace

	OrgAndSignedInSchoolAdminToken map[string]string
}

type connectGRPCOptions struct {
	bobSvcAddress                  string
	tomSvcAddress                  string
	yasuoSvcAddress                string
	eurekaSvcAddress               string
	fatimaSvcAddress               string
	shamirSvcAddress               string
	userMgmtSvcAddress             string
	notificationMgmtSvcAddress     string
	spikeSvcAddress                string
	paymentSvcAddress              string
	entryExitMgmtSvcAddress        string
	masterMgmtSvcAddress           string
	invoiceMgmtSvcAddress          string
	lessonMgmtSvcAddress           string
	enigmaSvcAddress               string
	virtualClassroomSvcAddress     string
	virtualClassroomHttpSvcAddress string
	calendarSvcAddress             string
	timesheetSvcAddress            string
	discountSvcAddress             string
	conversationMgmtSvcAddress     string
	authSvcAddress                 string
	credentials                    grpc.DialOption
	dialOptions                    []grpc.DialOption
}

type ConnectGRPCOption interface {
	configureGRPCOpt(opts *connectGRPCOptions) error
}

type connectGRPCOptFn func(opts *connectGRPCOptions) error

func (opt connectGRPCOptFn) configureGRPCOpt(opts *connectGRPCOptions) error {
	return opt(opts)
}

func WithBobSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.bobSvcAddress = rsc.GetAddress("bob")
		return nil
	})
}

func WithTomSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.tomSvcAddress = rsc.GetAddress("tom")
		return nil
	})
}

func WithYasuoSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.yasuoSvcAddress = rsc.GetAddress("yasuo")
		return nil
	})
}

func WithEurekaSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.eurekaSvcAddress = rsc.GetAddress("eureka")
		return nil
	})
}

func WithFatimaSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.fatimaSvcAddress = rsc.GetAddress("fatima")
		return nil
	})
}

func WithShamirSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.shamirSvcAddress = rsc.GetAddress("shamir")
		return nil
	})
}

func WithUserMgmtSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.userMgmtSvcAddress = rsc.GetAddress("usermgmt")
		return nil
	})
}

func WithNotificationMgmtSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.notificationMgmtSvcAddress = rsc.GetAddress("notificationmgmt")
		return nil
	})
}

func WithPaymentSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.paymentSvcAddress = rsc.GetAddress("payment")
		return nil
	})
}

func WithEntryExitMgmtSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.entryExitMgmtSvcAddress = rsc.GetAddress("entryexitmgmt")
		return nil
	})
}

func WithMasterMgmtSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.masterMgmtSvcAddress = rsc.GetAddress("mastermgmt")
		return nil
	})
}

func WithLessonMgmtSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.lessonMgmtSvcAddress = "lessonmgmt.local-manabie-backend.svc.cluster.local:6550"
		return nil
	})
}

func WithInvoiceMgmtSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.invoiceMgmtSvcAddress = rsc.GetAddress("invoicemgmt")
		return nil
	})
}

func WithVirtualClassroomSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.virtualClassroomSvcAddress = rsc.GetAddress("virtualclassroom")
		return nil
	})
}

func WithCalendarSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.calendarSvcAddress = rsc.GetAddress("calendar")
		return nil
	})
}

func WithTimesheetSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.timesheetSvcAddress = rsc.GetAddress("timesheet")
		return nil
	})
}

func WithEnigmaSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.enigmaSvcAddress = rsc.GetHTTPAddress("enigma")
		return nil
	})
}

func WithVirtualClassroomHttpSvcAddress(virSvcAddress string) ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.virtualClassroomHttpSvcAddress = virSvcAddress
		return nil
	})
}

func WithDiscountSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.discountSvcAddress = rsc.GetAddress("discount")
		return nil
	})
}

func WithSpikeSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.spikeSvcAddress = rsc.GetAddress("spike")
		return nil
	})
}

func WithConversationMgmtSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.conversationMgmtSvcAddress = rsc.GetAddress("conversationmgmt")
		return nil
	})
}

func WithAuthSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.authSvcAddress = rsc.GetAddress("auth")
		return nil
	})
}

func WithStressTestSvcAddress() ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		address := "api.staging.manabie.io:31500"
		opts.shamirSvcAddress = address
		opts.tomSvcAddress = address
		opts.userMgmtSvcAddress = address
		opts.bobSvcAddress = address
		return nil
	})
}

func WithCredentials(credentials grpc.DialOption) ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.credentials = credentials
		return nil
	})
}

func WithDialOptions(dialOpts ...grpc.DialOption) ConnectGRPCOption {
	return connectGRPCOptFn(func(opts *connectGRPCOptions) error {
		opts.dialOptions = append(opts.dialOptions, dialOpts...)
		return nil
	})
}

func (c *Connections) ConnectGRPC(ctx context.Context, opts ...ConnectGRPCOption) error {
	var err error
	grpcAddressOption := &connectGRPCOptions{}
	for i := range opts {
		if err = opts[i].configureGRPCOpt(grpcAddressOption); err != nil {
			return fmt.Errorf("failed to configureGRPCOpt: %v", err)
		}
	}

	if grpcAddressOption.credentials == nil {
		return fmt.Errorf("grpcAddressOption.credentials is nil")
	}

	commonDialOptions := []grpc.DialOption{grpcAddressOption.credentials, grpc.WithBlock()}
	commonDialOptions = append(commonDialOptions, grpcAddressOption.dialOptions...)

	if grpcAddressOption.bobSvcAddress != "" {
		c.BobConn, err = grpc.DialContext(ctx, grpcAddressOption.bobSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to Bob at %s: %s", grpcAddressOption.bobSvcAddress, err)
		}
	}

	if grpcAddressOption.tomSvcAddress != "" {
		c.TomConn, err = grpc.DialContext(ctx, grpcAddressOption.tomSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to Tom at %s: %s", grpcAddressOption.tomSvcAddress, err)
		}
	}

	if grpcAddressOption.yasuoSvcAddress != "" {
		c.YasuoConn, err = grpc.DialContext(ctx, grpcAddressOption.yasuoSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to Yasuo at %s: %s", grpcAddressOption.yasuoSvcAddress, err)
		}
	}

	if grpcAddressOption.eurekaSvcAddress != "" {
		c.EurekaConn, err = grpc.DialContext(ctx, grpcAddressOption.eurekaSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to Eureka at %s: %s", grpcAddressOption.eurekaSvcAddress, err)
		}
	}

	if grpcAddressOption.fatimaSvcAddress != "" {
		c.FatimaConn, err = grpc.DialContext(ctx, grpcAddressOption.fatimaSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to Fatima at %s: %s", grpcAddressOption.fatimaSvcAddress, err)
		}
	}

	if grpcAddressOption.shamirSvcAddress != "" {
		c.ShamirConn, err = grpc.DialContext(ctx, grpcAddressOption.shamirSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to Shamir at %s: %s", grpcAddressOption.shamirSvcAddress, err)
		}
	}

	if grpcAddressOption.userMgmtSvcAddress != "" {
		c.UserMgmtConn, err = grpc.DialContext(ctx, grpcAddressOption.userMgmtSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to Usermgmt at %s: %s", grpcAddressOption.userMgmtSvcAddress, err)
		}
	}

	if grpcAddressOption.paymentSvcAddress != "" {
		c.PaymentConn, err = grpc.DialContext(ctx, grpcAddressOption.paymentSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to Payment at %s: %s", grpcAddressOption.paymentSvcAddress, err)
		}
	}

	if grpcAddressOption.entryExitMgmtSvcAddress != "" {
		c.EntryExitMgmtConn, err = grpc.DialContext(ctx, grpcAddressOption.entryExitMgmtSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to EntryExitMgmt at %s: %s", grpcAddressOption.entryExitMgmtSvcAddress, err)
		}
	}

	if grpcAddressOption.masterMgmtSvcAddress != "" {
		c.MasterMgmtConn, err = grpc.DialContext(ctx, grpcAddressOption.masterMgmtSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to MasterMgmt at %s: %s", grpcAddressOption.masterMgmtSvcAddress, err)
		}
	}

	if grpcAddressOption.lessonMgmtSvcAddress != "" {
		c.LessonMgmtConn, err = grpc.DialContext(ctx, grpcAddressOption.lessonMgmtSvcAddress, grpcAddressOption.credentials, grpc.WithBlock())
		if err != nil {
			return fmt.Errorf("cannot connect to lesson management at %s: %s", grpcAddressOption.lessonMgmtSvcAddress, err)
		}
	}

	if grpcAddressOption.enigmaSvcAddress != "" {
		c.EnigmaConn, err = grpc.DialContext(ctx, grpcAddressOption.enigmaSvcAddress, grpcAddressOption.credentials, grpc.WithBlock())
		if err != nil {
			return fmt.Errorf("cannot connect to enigma at %s: %s", grpcAddressOption.enigmaSvcAddress, err)
		}
	}

	if grpcAddressOption.invoiceMgmtSvcAddress != "" {
		c.InvoiceMgmtConn, err = grpc.DialContext(ctx, grpcAddressOption.invoiceMgmtSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to InvoiceMgmt at %s: %s", grpcAddressOption.invoiceMgmtSvcAddress, err)
		}
	}

	if grpcAddressOption.virtualClassroomSvcAddress != "" {
		c.VirtualClassroomConn, err = grpc.DialContext(ctx, grpcAddressOption.virtualClassroomSvcAddress, grpcAddressOption.credentials, grpc.WithBlock())
		if err != nil {
			return fmt.Errorf("cannot connect to VirtualClassroom at %s: %s", grpcAddressOption.virtualClassroomSvcAddress, err)
		}
	}

	if grpcAddressOption.virtualClassroomHttpSvcAddress != "" {
		c.VirtualClassroomHTTPConn, err = grpc.DialContext(ctx, grpcAddressOption.virtualClassroomHttpSvcAddress, grpcAddressOption.credentials, grpc.WithBlock())
		if err != nil {
			return fmt.Errorf("cannot connect to VirtualClassroom http at %s: %s", grpcAddressOption.virtualClassroomHttpSvcAddress, err)
		}
	}

	if grpcAddressOption.calendarSvcAddress != "" {
		c.CalendarConn, err = grpc.DialContext(ctx, grpcAddressOption.calendarSvcAddress, grpcAddressOption.credentials, grpc.WithBlock())
		if err != nil {
			return fmt.Errorf("cannot connect to Calendar at %s: %s", grpcAddressOption.calendarSvcAddress, err)
		}
	}

	if grpcAddressOption.notificationMgmtSvcAddress != "" {
		c.NotificationMgmtConn, err = grpc.DialContext(ctx, grpcAddressOption.notificationMgmtSvcAddress, grpcAddressOption.credentials, grpc.WithBlock())
		if err != nil {
			return fmt.Errorf("cannot connect to NotificationMgmt at %s: %s", grpcAddressOption.notificationMgmtSvcAddress, err)
		}
	}

	if grpcAddressOption.spikeSvcAddress != "" {
		c.SpikeConn, err = grpc.DialContext(ctx, grpcAddressOption.spikeSvcAddress, grpcAddressOption.credentials, grpc.WithBlock())
		if err != nil {
			return fmt.Errorf("cannot connect to NotificationMgmt at %s: %s", grpcAddressOption.spikeSvcAddress, err)
		}
	}

	if grpcAddressOption.timesheetSvcAddress != "" {
		c.TimesheetConn, err = grpc.DialContext(ctx, grpcAddressOption.timesheetSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to Timesheet at %s: %s", grpcAddressOption.timesheetSvcAddress, err)
		}
	}

	if grpcAddressOption.discountSvcAddress != "" {
		c.DiscountConn, err = grpc.DialContext(ctx, grpcAddressOption.discountSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to Discount at %s: %s", grpcAddressOption.discountSvcAddress, err)
		}
	}

	if grpcAddressOption.conversationMgmtSvcAddress != "" {
		c.ConversationMgmtConn, err = grpc.DialContext(ctx, grpcAddressOption.conversationMgmtSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to ConversationMgmt at %s: %s", grpcAddressOption.conversationMgmtSvcAddress, err)
		}
	}

	if grpcAddressOption.authSvcAddress != "" {
		c.AuthConn, err = grpc.DialContext(ctx, grpcAddressOption.authSvcAddress, commonDialOptions...)
		if err != nil {
			return fmt.Errorf("cannot connect to ConversationMgmt at %s: %s", grpcAddressOption.authSvcAddress, err)
		}
	}

	return nil
}

type connectDBOptions struct {
	bobDBConfig                      *configs.PostgresDatabaseConfig
	tomDBConfig                      *configs.PostgresDatabaseConfig
	eurekaDBConfig                   *configs.PostgresDatabaseConfig
	fatimaDBConfig                   *configs.PostgresDatabaseConfig
	invoiceMgmtDBConfig              *configs.PostgresDatabaseConfig
	zeusDBConfig                     *configs.PostgresDatabaseConfig
	bobPostgresDBConfig              *configs.PostgresDatabaseConfig
	tomPostgresDBConfig              *configs.PostgresDatabaseConfig
	entryExitMgmtDBConfig            *configs.PostgresDatabaseConfig
	timesheetDBConfig                *configs.PostgresDatabaseConfig
	mastermgmtDBConfig               *configs.PostgresDatabaseConfig
	mastermgmtPostgresDBConfig       *configs.PostgresDatabaseConfig
	yasuoDBConfig                    *configs.PostgresDatabaseConfig
	calendarDBConfig                 *configs.PostgresDatabaseConfig
	lessonmgmtDBConfig               *configs.PostgresDatabaseConfig
	notificationMgmtDBConfig         *configs.PostgresDatabaseConfig
	notificationMgmtPostgresDBConfig *configs.PostgresDatabaseConfig
	authPostgresDBConfig             *configs.PostgresDatabaseConfig
	invoiceMgmtPostgresDBConfig      *configs.PostgresDatabaseConfig
}

type ConnectDBOption interface {
	configureDBOpt(opts *connectDBOptions) error
}

type connectDBOptFn func(opts *connectDBOptions) error

func (opt connectDBOptFn) configureDBOpt(opts *connectDBOptions) error {
	return opt(opts)
}

func WithBobDBConfig(bobDBConfig configs.PostgresDatabaseConfig) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		opts.bobDBConfig = &bobDBConfig
		return nil
	})
}

func WithInvoiceMgmtPostgresDBConfig(invoiceMgmtPostgresDBConfig configs.PostgresDatabaseConfig, postgresPassword string) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		invoiceMgmtPostgresDBConfig.User = "postgres"
		invoiceMgmtPostgresDBConfig.Password = postgresPassword
		opts.invoiceMgmtPostgresDBConfig = &invoiceMgmtPostgresDBConfig

		return nil
	})
}

func WithInvoiceMgmtDBConfig(invoiceMgmtDBConfig configs.PostgresDatabaseConfig) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		opts.invoiceMgmtDBConfig = &invoiceMgmtDBConfig
		return nil
	})
}

func WithTomDBConfig(tomDBConfig configs.PostgresDatabaseConfig) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		opts.tomDBConfig = &tomDBConfig
		return nil
	})
}

func WithEurekaDBConfig(eurekaDBConfig configs.PostgresDatabaseConfig) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		opts.eurekaDBConfig = &eurekaDBConfig
		return nil
	})
}

func WithFatimaDBConfig(fatimaDBConfig configs.PostgresDatabaseConfig) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		opts.fatimaDBConfig = &fatimaDBConfig
		return nil
	})
}

func WithZeusDBConfig(zeusDBConfig configs.PostgresDatabaseConfig) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		opts.zeusDBConfig = &zeusDBConfig
		return nil
	})
}

func WithEntryExitMgmtDBConfig(entryExitMgmtDBConfig configs.PostgresDatabaseConfig) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		opts.entryExitMgmtDBConfig = &entryExitMgmtDBConfig
		return nil
	})
}

func WithMastermgmtDBConfig(masterConfig configs.PostgresDatabaseConfig) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		opts.mastermgmtDBConfig = &masterConfig
		return nil
	})
}

func WithBobPostgresDBConfig(bobPostgresDBConfig configs.PostgresDatabaseConfig, postgresPassword string) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		bobPostgresDBConfig.User = "postgres"
		bobPostgresDBConfig.Password = postgresPassword
		opts.bobPostgresDBConfig = &bobPostgresDBConfig
		return nil
	})
}

func WithTomPostgresDBConfig(tomPostgresDBConfig configs.PostgresDatabaseConfig, postgresPassword string) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		tomPostgresDBConfig.User = "postgres"
		tomPostgresDBConfig.Password = postgresPassword
		opts.tomPostgresDBConfig = &tomPostgresDBConfig
		return nil
	})
}

func WithTimesheetPostgresDBConfig(timesheetPostgresDBConfig configs.PostgresDatabaseConfig, postgresPassword string) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		timesheetPostgresDBConfig.User = "postgres"
		timesheetPostgresDBConfig.Password = postgresPassword
		opts.timesheetDBConfig = &timesheetPostgresDBConfig
		return nil
	})
}

func WithMastermgmtPostgresDBConfig(masterPostgresConfig configs.PostgresDatabaseConfig, postgresPassword string) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		masterPostgresConfig.User = "postgres"
		masterPostgresConfig.Password = postgresPassword
		opts.mastermgmtPostgresDBConfig = &masterPostgresConfig
		return nil
	})
}

func WithCalendarDBConfig(calendarDBConfig configs.PostgresDatabaseConfig) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		opts.calendarDBConfig = &calendarDBConfig
		return nil
	})
}

func WithLessonmgmtDBConfig(lessonmgmtDBConfig configs.PostgresDatabaseConfig) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		opts.lessonmgmtDBConfig = &lessonmgmtDBConfig
		return nil
	})
}

func WithNotificationmgmtDBConfig(notificationDBConfig configs.PostgresDatabaseConfig) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		opts.notificationMgmtDBConfig = &notificationDBConfig
		return nil
	})
}

func WithNotificationmgmtPostgresDBConfig(notificationMgmtPostgresDBConfig configs.PostgresDatabaseConfig, postgresPassword string) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		notificationMgmtPostgresDBConfig.User = "postgres"
		notificationMgmtPostgresDBConfig.Password = postgresPassword
		opts.notificationMgmtPostgresDBConfig = &notificationMgmtPostgresDBConfig
		return nil
	})
}

func WithAuthPostgresDBConfig(authPostgresDBConfig configs.PostgresDatabaseConfig, postgresPassword string) ConnectDBOption {
	return connectDBOptFn(func(opts *connectDBOptions) error {
		authPostgresDBConfig.User = "postgres"
		authPostgresDBConfig.Password = postgresPassword
		opts.authPostgresDBConfig = &authPostgresDBConfig
		return nil
	})
}

func (c *Connections) ConnectDB(ctx context.Context, opts ...ConnectDBOption) error {
	connectDBOption := &connectDBOptions{}
	var err error
	for i := range opts {
		if err = opts[i].configureDBOpt(connectDBOption); err != nil {
			return fmt.Errorf("failed to configureDBOpt: %v", err)
		}
	}

	if connectDBOption.bobDBConfig != nil {
		c.BobDB = getDBConnectionDemo(ctx, connectDBOption.bobDBConfig)
		c.BobDBTrace = &database.DBTrace{DB: c.BobDB}
	}

	if connectDBOption.tomDBConfig != nil {
		c.TomDB = getDBConnectionDemo(ctx, connectDBOption.tomDBConfig)
		c.TomDBTrace = &database.DBTrace{DB: c.TomDB}
	}

	if connectDBOption.eurekaDBConfig != nil {
		c.EurekaDB = getDBConnectionDemo(ctx, connectDBOption.eurekaDBConfig)
		c.EurekaDBTrace = &database.DBTrace{DB: c.EurekaDB}
	}

	if connectDBOption.fatimaDBConfig != nil {
		c.FatimaDB = getDBConnectionDemo(ctx, connectDBOption.fatimaDBConfig)
		c.FatimaDBTrace = &database.DBTrace{DB: c.FatimaDB}
	}

	if connectDBOption.invoiceMgmtDBConfig != nil {
		c.InvoiceMgmtDB = getDBConnectionDemo(ctx, connectDBOption.invoiceMgmtDBConfig)
		c.InvoiceMgmtDBTrace = &database.DBTrace{DB: c.InvoiceMgmtDB}
	}

	if connectDBOption.invoiceMgmtPostgresDBConfig != nil {
		c.InvoiceMgmtPostgresDB = getDBConnectionDemo(ctx, connectDBOption.invoiceMgmtPostgresDBConfig)
		c.InvoiceMgmtPostgresDBTrace = &database.DBTrace{DB: c.InvoiceMgmtPostgresDB}
	}

	if connectDBOption.zeusDBConfig != nil {
		c.ZeusDB = getDBConnectionDemo(ctx, connectDBOption.zeusDBConfig)
		c.ZeusDBTrace = &database.DBTrace{DB: c.ZeusDB}
	}

	if connectDBOption.entryExitMgmtDBConfig != nil {
		c.EntryExitMgmtDB = getDBConnectionDemo(ctx, connectDBOption.entryExitMgmtDBConfig)
		c.EntryExitMgmtDBTrace = &database.DBTrace{DB: c.EntryExitMgmtDB}
	}

	if connectDBOption.bobPostgresDBConfig != nil {
		c.BobPostgresDB = getDBConnectionDemo(ctx, connectDBOption.bobPostgresDBConfig)
		c.BobPostgresDBTrace = &database.DBTrace{DB: c.BobPostgresDB}
	}

	if connectDBOption.authPostgresDBConfig != nil {
		c.AuthPostgresDB = getDBConnectionDemo(ctx, connectDBOption.authPostgresDBConfig)
		c.AuthPostgresDBTrace = &database.DBTrace{DB: c.AuthPostgresDB}
	}
	if connectDBOption.tomPostgresDBConfig != nil {
		c.TomPostgresDB = getDBConnectionDemo(ctx, connectDBOption.tomPostgresDBConfig)
		c.TomDBTrace = &database.DBTrace{DB: c.TomPostgresDB}
	}

	if connectDBOption.timesheetDBConfig != nil {
		c.TimesheetDB = getDBConnectionDemo(ctx, connectDBOption.timesheetDBConfig)
		c.TimesheetDBTrace = &database.DBTrace{DB: c.TimesheetDB}
	}

	if connectDBOption.mastermgmtDBConfig != nil {
		c.MasterMgmtDB = getDBConnectionDemo(ctx, connectDBOption.mastermgmtDBConfig)
		c.MasterMgmtDBTrace = &database.DBTrace{DB: c.MasterMgmtDB}
	}
	if connectDBOption.mastermgmtPostgresDBConfig != nil {
		c.MasterMgmtPostgresDB = getDBConnectionDemo(ctx, connectDBOption.mastermgmtPostgresDBConfig)
		c.MasterMgmtPostgresDBTrace = &database.DBTrace{DB: c.MasterMgmtPostgresDB}
	}
	if connectDBOption.yasuoDBConfig != nil {
		c.YasuoDB = getDBConnectionDemo(ctx, connectDBOption.yasuoDBConfig)
		c.YasuoDBTrace = &database.DBTrace{DB: c.YasuoDB}
	}

	if connectDBOption.calendarDBConfig != nil {
		c.CalendarDB = getDBConnectionDemo(ctx, connectDBOption.calendarDBConfig)
		c.CalendarDBTrace = &database.DBTrace{DB: c.CalendarDB}
	}

	if connectDBOption.lessonmgmtDBConfig != nil {
		c.LessonmgmtDB = getDBConnectionDemo(ctx, connectDBOption.lessonmgmtDBConfig)
		c.LessonmgmtDBTrace = &database.DBTrace{DB: c.LessonmgmtDB}
	}

	if connectDBOption.notificationMgmtDBConfig != nil {
		c.NotificationMgmtDB = getDBConnectionDemo(ctx, connectDBOption.notificationMgmtDBConfig)
		c.NotificationMgmtDBTrace = &database.DBTrace{DB: c.NotificationMgmtDB}
	}

	if connectDBOption.notificationMgmtPostgresDBConfig != nil {
		c.NotificationMgmtPostgresDB = getDBConnectionDemo(ctx, connectDBOption.notificationMgmtPostgresDBConfig)
		c.NotificationMgmtPostgresDBTrace = &database.DBTrace{DB: c.NotificationMgmtPostgresDB}
	}
	return nil
}

func getDBConnectionDemo(ctx context.Context, pgCfg *configs.PostgresDatabaseConfig) *pgxpool.Pool {
	cfgCopy := *pgCfg
	dbPool, _, _ := database.NewPool(ctx, zap.NewNop(), cfgCopy)
	return dbPool
}

func (c *Connections) CloseAllConnections() {
	if c.BobDB != nil {
		c.BobDB.Close()
	}

	if c.YasuoDB != nil {
		c.YasuoDB.Close()
	}

	if c.TomDB != nil {
		c.TomDB.Close()
	}

	if c.EurekaDB != nil {
		c.EurekaDB.Close()
	}

	if c.FatimaDB != nil {
		c.FatimaDB.Close()
	}

	if c.InvoiceMgmtDB != nil {
		c.InvoiceMgmtDB.Close()
	}

	if c.ZeusDB != nil {
		c.ZeusDB.Close()
	}

	if c.EntryExitMgmtDB != nil {
		c.EntryExitMgmtDB.Close()
	}

	if c.BobPostgresDB != nil {
		c.BobPostgresDB.Close()
	}

	if c.BobConn != nil {
		c.BobConn.Close()
	}

	if c.TomConn != nil {
		c.TomConn.Close()
	}

	if c.YasuoConn != nil {
		c.YasuoConn.Close()
	}

	if c.EurekaConn != nil {
		c.EurekaConn.Close()
	}

	if c.FatimaConn != nil {
		c.FatimaConn.Close()
	}

	if c.InvoiceMgmtConn != nil {
		c.InvoiceMgmtConn.Close()
	}

	if c.EntryExitMgmtConn != nil {
		c.EntryExitMgmtConn.Close()
	}

	if c.UserMgmtConn != nil {
		c.UserMgmtConn.Close()
	}

	if c.MasterMgmtConn != nil {
		c.MasterMgmtConn.Close()
	}
	if c.VirtualClassroomConn != nil {
		c.VirtualClassroomConn.Close()
	}
	if c.VirtualClassroomHTTPConn != nil {
		c.VirtualClassroomHTTPConn.Close()
	}
	if c.TimesheetConn != nil {
		c.TimesheetConn.Close()
	}
	if c.CalendarConn != nil {
		c.CalendarConn.Close()
	}
	if c.EnigmaConn != nil {
		c.EnigmaConn.Close()
	}

	if c.ShamirConn != nil {
		c.ShamirConn.Close()
	}

	if c.NotificationMgmtConn != nil {
		c.NotificationMgmtConn.Close()
	}

	if c.JSM != nil {
		c.JSM.Close()
	}

	if c.InvoiceMgmtPostgresDB != nil {
		c.InvoiceMgmtPostgresDB.Close()
	}
	if c.AuthConn != nil {
		c.AuthConn.Close()
	}
	if c.AuthPostgresDB != nil {
		c.AuthPostgresDB.Close()
	}
}
