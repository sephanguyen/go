package common

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/logger"

	firebase "firebase.google.com/go"
	"github.com/cucumber/godog"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var RegisteredTestWithCommonConnections = map[string]func(c *Config, dep *Connections) func(*godog.ScenarioContext){}

func RegisterTestWithCommonConnection(name string, h func(c *Config, dep *Connections) func(*godog.ScenarioContext)) {
	RegisteredTestWithCommonConnections[name] = h
}

type SuiteBuilder[T any] struct {
	SuiteInitFunc    func(*T, RunTimeFlag) func(ctx *godog.TestSuiteContext)
	ScenarioInitFunc func(*T, RunTimeFlag) func(ctx *godog.ScenarioContext)
}
type BuiltSuite struct {
	SuiteInitializer    func(ctx *godog.TestSuiteContext)
	ScenarioInitializer func(ctx *godog.ScenarioContext)
}
type RunTimeFlag struct {
	FirebaseAddr string
	ApplicantID  string
	OtelEnabled  bool
}

var RegisteredTests = map[string]func(string, string, string, RunTimeFlag) BuiltSuite{}

func RegisterTest[T any](service string, s *SuiteBuilder[T]) {
	RegisteredTests[service] = func(commonConfigPath, configPath, secretsPath string, f RunTimeFlag) BuiltSuite {
		var c T
		configs.MustLoadConfig(
			context.Background(),
			commonConfigPath,
			configPath,
			secretsPath,
			&c,
		)
		return BuiltSuite{
			SuiteInitializer:    s.SuiteInitFunc(&c, f),
			ScenarioInitializer: s.ScenarioInitFunc(&c, f),
		}
	}
}

// Find all fields with the given T type matching given suffix
// for example
//
//	struct {
//		BoDB Config
//		TomDB Config
//	}
func extractFieldMapWithSuffix[T any](config interface{}, fieldSuffix string) (map[string]T, error) {
	v := reflect.ValueOf(config)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("config must be a struct")
	}
	t := reflect.TypeOf(config)
	ret := map[string]T{}
	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		if strings.HasSuffix(fieldName, fieldSuffix) {
			dbName := fieldName[:len(fieldName)-2]
			fv := v.FieldByName(fieldName)
			if !fv.IsValid() {
				return nil, fmt.Errorf("field %s not found", fieldName)
			}
			out, ok := fv.Interface().(T)
			if !ok {
				tv := fv.Type()
				return nil, fmt.Errorf("expected field %q to be of \"%T\" type, got \"%s.%s\"", fieldName, out, tv.PkgPath(), tv.Name())
			}
			ret[dbName] = out
		}
	}
	return ret, nil
}

func SetupConnections(c *Config, fakeFirebaseAddr string) Connections {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	deps := Connections{
		FirebaseAddr: fakeFirebaseAddr,
		ApplicantID:  c.JWTApplicant,
		Logger:       logger.NewZapLogger(c.Common.Log.ApplicationLevel, true),
	}

	zapLogger := logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)
	grpConnMap, err := extractFieldMapWithSuffix[string](*c, "SrvAddr")
	if err != nil {
		zapLogger.Panic("extractFieldMapWithSuffx", zap.Error(err))
	}

	commonDialOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock()}
	for serviceName, addr := range grpConnMap {
		if addr != "" {
			connField := reflect.ValueOf(&deps).Elem().FieldByName(serviceName + "Conn")
			if connField.CanSet() {
				conn, err := grpc.DialContext(ctx, addr, commonDialOptions...)
				if err != nil {
					zapLogger.Sugar().Panicf("DialContext %s\n", addr)
				}
				connField.Set(reflect.ValueOf(conn))
			}
		}
	}

	const cantConnectFirebase = "cannot connect to firebase: %v"

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		zapLogger.Panic(fmt.Sprintf(cantConnectFirebase, err))
	}
	deps.FirebaseClient, err = app.Auth(ctx)
	if err != nil {
		zapLogger.Panic(fmt.Sprintf("cannot create firebase client: %v", err))
	}

	deps.GCPApp, err = gcp.NewApp(ctx, "", c.Common.IdentityPlatformProject)
	if err != nil {
		zapLogger.Panic(fmt.Sprintf(cantConnectFirebase, err))
	}
	deps.FirebaseAuthClient, err = internal_auth_tenant.NewFirebaseAuthClientFromGCP(ctx, deps.GCPApp)
	if err != nil {
		zapLogger.Panic(fmt.Sprintf(cantConnectFirebase, err))
	}
	deps.TenantManager, err = internal_auth_tenant.NewTenantManagerFromGCP(ctx, deps.GCPApp)
	if err != nil {
		zapLogger.Panic(fmt.Sprintf("cannot create tenant manager: %v", err))
	}

	// TODO: there are some common steps that need to run before test suites, but not required by some other services
	// if we run the script below, unleash test will fails on CI, because it does not setup for bob tables
	// // Init auth info
	// stmt :=
	// 	`
	// 	INSERT INTO organization_auths
	// 		(organization_id, auth_project_id, auth_tenant_id)
	// 	SELECT
	// 		school_id, 'fake_aud', ''
	// 	FROM
	// 		schools
	// 	UNION
	// 	SELECT
	// 		school_id, 'dev-manabie-online', ''
	// 	FROM
	// 		schools
	// 	ON CONFLICT
	// 		DO NOTHING
	// 	;
	// 	`
	// _, err = deps.BobPostgresDBTrace.Exec(ctx, stmt)
	// if err != nil {
	// 	zapLogger.Panic(fmt.Sprintf("cannot init auth info: %v", err))
	// }
	return deps
}
