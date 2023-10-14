package shamir

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	unleash_client "github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/shamir/configurations"
	"github.com/manabie-com/backend/internal/shamir/services"
	grpctrans "github.com/manabie-com/backend/internal/shamir/transports/grpc"
	"github.com/manabie-com/backend/internal/shamir/transports/rest"
	authRepository "github.com/manabie-com/backend/internal/usermgmt/modules/auth/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/features"
	pb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"

	"github.com/gin-gonic/gin"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.opencensus.io/plugin/ocgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	s := &server{}
	bootstrap.
		WithGRPC[configurations.Config](s).
		WithHTTP(s).
		WithMonitorServicer(s).
		Register(s)
}

type server struct {
	tokVerifier *services.TokenVerifier
	service     *grpctrans.Service
	bootstrap.DefaultMonitorService[configurations.Config]
}

func (s *server) WithUnaryServerInterceptors(_ configurations.Config, rsc *bootstrap.Resources) []grpc.UnaryServerInterceptor {
	customs := []grpc.UnaryServerInterceptor{
		grpc_zap.PayloadUnaryServerInterceptor(rsc.Logger(), func(ctx context.Context, fullMethod string, _ interface{}) bool {
			return true
		}),
		tracer.UnaryActivityLogRequestInterceptor(rsc.NATS(), rsc.Logger(), s.ServerName()),
	}

	grpcUnary := bootstrap.DefaultUnaryServerInterceptor(rsc)
	grpcUnary = append(grpcUnary, customs...)

	return grpcUnary
}

func (s *server) WithStreamServerInterceptors(_ configurations.Config, rsc *bootstrap.Resources) []grpc.StreamServerInterceptor {
	return bootstrap.DefaultStreamServerInterceptor(rsc)
}

func (s *server) WithServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}
}

func (*server) ServerName() string {
	return "shamir"
}

func dbConnectionWithError(rsc *bootstrap.Resources, dbname string) (db *database.DBTrace, err error) {
	defer func() {
		if r := recover(); r != nil {
			db = nil
			err = fmt.Errorf("cannot init db: %v", r)
		}
	}()
	db = rsc.DBWith(dbname)
	return
}

func (s *server) InitDependencies(c configurations.Config, rsc *bootstrap.Resources) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	zapLogger := rsc.Logger()
	db := rsc.DBWith("bob")
	authDB, err := dbConnectionWithError(rsc, "auth")
	if err != nil {
		zapLogger.Error("can't connect auth db: ", zap.Error(err))
	}

	var privateKeys map[string]*rsa.PrivateKey
	var primaryKeyID string
	privateKeys, primaryKeyID, err = configurations.LoadPrivateKeysWithSopsFormat(c.KeysGlob, c.PrimaryKeyFile)
	if err != nil {
		zapLogger.Panic("Failed to load private keys", zap.Error(err))
	}

	organizationAuthRepo := (&repository.OrganizationRepo{}).WithDefaultValue(c.Common.Environment)

	organizationAuths, getAllOrganizationAuthsErr := organizationAuthRepo.GetAll(ctx, db, 1000)
	if getAllOrganizationAuthsErr == nil {
		c.Issuers = appendAdditionalIssuer(c.Issuers, organizationAuths)
	} else {
		zapLogger.Error("can't get all organization auth information")
	}

	// Init unleash client
	var unleashClientInstance unleash_client.ClientInstance
	unleashClientInstance, err = unleash_client.NewUnleashClientInstance(
		c.UnleashClientConfig.URL,
		c.UnleashClientConfig.AppName,
		c.UnleashClientConfig.APIToken,
		zapLogger)
	if err != nil {
		zapLogger.Fatal("failed at new unleash instance:", zap.Error(err))
	}
	err = unleashClientInstance.ConnectToUnleashClient()
	if err != nil {
		zapLogger.Fatal("failed to connect to unleash as a client:", zap.Error(err))
	}

	v, err := services.NewTokenVerifier(unleashClientInstance, c.Common.Environment, c.Vendor, privateKeys, primaryKeyID, c.Issuers)
	if err != nil {
		zapLogger.Panic("Failed to create token verifier", zap.Error(err))
	}

	s.tokVerifier = v

	svc := &grpctrans.Service{
		Verifier: v,
		DB:       db,
		AuthDB:   authDB,
		SalesforceService: services.SalesforceService{
			Config: c.SalesforceConfigs,
		},

		UnleashClient:                 unleashClientInstance,
		DefaultOrganizationAuthValues: (&repository.OrganizationRepo{}).DefaultOrganizationAuthValues(c.Common.Environment),
		UserRepoV2:                    &repository.UserRepoV2{},
		DomainAPIKeypairRepo: &repository.DomainAPIKeypairRepo{
			EncryptedKey:  c.OpenAPI.AESKey,
			InitialVector: c.OpenAPI.AESIV,
		},
		OrganizationRepo:   &repository.OrganizationRepo{},
		OrganizationRepoV2: &authRepository.OrganizationRepo{},
		Env:                c.Common.Environment,
		FeatureManager: &features.FeatureManager{
			UnleashClient:             unleashClientInstance,
			Env:                       c.Common.Environment,
			DB:                        db,
			InternalConfigurationRepo: &repository.DomainInternalConfigurationRepo{},
		},
	}

	s.service = svc

	return nil
}

func (s *server) SetupGRPC(_ context.Context, grpcserver *grpc.Server, _ configurations.Config, _ *bootstrap.Resources) error {
	pb.RegisterTokenReaderServiceServer(grpcserver, s.service)
	pb.RegisterInternalServiceServer(grpcserver, s.service)
	return nil
}

func (s *server) SetupHTTP(_ configurations.Config, g *gin.Engine, rsc *bootstrap.Resources) error {
	return rest.SetupGinEngine(g, s.tokVerifier, rsc.Logger(), rsc.GetHTTPAddress("shamir"))
}

func (*server) GracefulShutdown(context.Context) {}

func appendAdditionalIssuer(currentIssuers []configs.TokenIssuerConfig, organizationAuths []*entity.OrganizationAuth) []configs.TokenIssuerConfig {
	additionalIssuers := make([]configs.TokenIssuerConfig, 0, len(organizationAuths))
	for _, organizationAuth := range organizationAuths {
		for _, issuer := range currentIssuers {
			if issuer.Audience == organizationAuth.AuthProjectID.String {
				continue
			}
			additionalIssuer := configs.TokenIssuerConfig{
				Issuer:       fmt.Sprintf("https://securetoken.google.com/%s", organizationAuth.AuthProjectID.String),
				Audience:     organizationAuth.AuthProjectID.String,
				JWKSEndpoint: "https://www.googleapis.com/service_accounts/v1/jwk/securetoken@system.gserviceaccount.com",
			}
			additionalIssuers = append(additionalIssuers, additionalIssuer)
		}
	}
	return append(currentIssuers, additionalIssuers...)
}
