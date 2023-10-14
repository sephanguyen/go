package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	_ "github.com/manabie-com/backend/features/accesscontrol"
	_ "github.com/manabie-com/backend/features/auth"
	_ "github.com/manabie-com/backend/features/bob"
	_ "github.com/manabie-com/backend/features/calendar"
	"github.com/manabie-com/backend/features/common"
	_ "github.com/manabie-com/backend/features/communication"
	_ "github.com/manabie-com/backend/features/conversationmgmt"
	_ "github.com/manabie-com/backend/features/discount"
	_ "github.com/manabie-com/backend/features/draft"
	_ "github.com/manabie-com/backend/features/eibanam/communication"
	_ "github.com/manabie-com/backend/features/eibanam/entryexitmanagement"
	_ "github.com/manabie-com/backend/features/eibanam/lesson"
	_ "github.com/manabie-com/backend/features/eibanam/multitenant"
	_ "github.com/manabie-com/backend/features/eibanam/usermanagement"
	_ "github.com/manabie-com/backend/features/enigma"
	_ "github.com/manabie-com/backend/features/entryexitmgmt"
	_ "github.com/manabie-com/backend/features/eureka"
	_ "github.com/manabie-com/backend/features/eurekav2"
	_ "github.com/manabie-com/backend/features/fatima"
	"github.com/manabie-com/backend/features/gandalf"
	"github.com/manabie-com/backend/features/gandalf/jprep"
	"github.com/manabie-com/backend/features/gandalf/learning"
	"github.com/manabie-com/backend/features/gandalf/managing"
	_ "github.com/manabie-com/backend/features/hephaestus"
	_ "github.com/manabie-com/backend/features/invoicemgmt"
	_ "github.com/manabie-com/backend/features/lessonmgmt"
	_ "github.com/manabie-com/backend/features/mastermgmt"
	_ "github.com/manabie-com/backend/features/payment"
	_ "github.com/manabie-com/backend/features/platform"
	repoSyllabus "github.com/manabie-com/backend/features/repository/syllabus"
	_ "github.com/manabie-com/backend/features/syllabus"
	_ "github.com/manabie-com/backend/features/timesheet"
	_ "github.com/manabie-com/backend/features/tom"
	_ "github.com/manabie-com/backend/features/unleash"
	_ "github.com/manabie-com/backend/features/usermgmt"
	_ "github.com/manabie-com/backend/features/virtualclassroom"
	_ "github.com/manabie-com/backend/features/yasuo"
	_ "github.com/manabie-com/backend/features/zeus"
	_ "github.com/manabie-com/backend/internal/golibs/automaxprocs"
	"github.com/manabie-com/backend/internal/golibs/configs"
	dpb "github.com/manabie-com/backend/pkg/manabuf/draft/v1"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/structpb"
)

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "pretty",
	Strict: true,
}

var (
	service          string
	commonConfigPath string
	configPath       string
	secretsPath      string
	fakeFirebaseAddr string
	shamirAddr       string
	applicantID      string

	traceEnabled bool
	otelEndpoint string

	pushgatewayEndpoint    string
	collectBDDTestsMetrics bool

	draftEndpoint string

	ciPullRequestID string
	ciRunID         string
	ciActor         string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	godog.BindFlags("godog.", flag.CommandLine, &opts)
	flag.StringVar(&service, "manabie.service", "bob", "service to run test [bob, draft, eureka, fatima, gandalf, tom, yasuo, usermgmt, payment, entryexitmgmt, mastermgmt, enigma, invoicemgmt, timesheet, calendar, virtualclassroom, accesscontrol, discount, auth]")
	flag.StringVar(&commonConfigPath, "manabie.commonConfigPath", "/configs/bob.common.config.yaml", "path to common configuration file")
	flag.StringVar(&configPath, "manabie.configPath", "/configs/bob.config.yaml", "path to configuration file")
	flag.StringVar(&secretsPath, "manabie.secretsPath", "/configs/bob.secrets.yaml.encrypted", "path to encrypted configuration file")
	flag.StringVar(&fakeFirebaseAddr, "manabie.fakeFirebaseAddr", "firebase.emulator.svc.cluster.local:40401", "fake firebase authenthication address")
	flag.StringVar(&shamirAddr, "manabie.shamirAddr", "shamir:5650", "shamir exchange token address")
	flag.BoolVar(&traceEnabled, "manabie.traceEnabled", false, "enable tracing report to Jaeger")
	flag.StringVar(&otelEndpoint, "manabie.otelEndpoint", "", "OpenTelemetry endpoint to send traces to")
	flag.StringVar(&applicantID, "applicantID", "manabie-local", "shamir exchange token applicant id")
	flag.StringVar(&pushgatewayEndpoint, "manabie.pushgatewayEndpoint", "", "Prometheus pushgateway endpoint to send metrics to")
	flag.BoolVar(&collectBDDTestsMetrics, "manabie.collectBDDTestsMetrics", false, "Set to true to collect BDD tests metrics")

	flag.StringVar(&draftEndpoint, "manabie.draftEndpoint", "", "Endpoint of draft GRPC service")
	flag.StringVar(&ciPullRequestID, "manabie.ciPullRequestID", "", "Github CI pull request title")
	flag.StringVar(&ciActor, "manabie.ciActor", "", "Github CI actor")
	flag.StringVar(&ciRunID, "manabie.ciRunID", "", "Github CI run id")
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	flag.Parse()
	opts.Paths = flag.Args()
	opts.DefaultContext = ctx

	ts := godog.TestSuite{
		Name:    service,
		Options: &opts,
	}

	if len(ts.Options.Paths) == 1 && ts.Options.Paths[0] == "." {
		ts.Options.Paths = []string{service}
	}

	if service == "gandalf" {
		printMessageForDisabledGandalf()
		return
	}
	printMessageIfServiceDeprecated(service)

	var (
		b *bddService
		m *metricsService
	)
	if collectBDDTestsMetrics {
		cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		conn, err := grpc.DialContext(
			cctx,
			draftEndpoint,
			grpc.WithBlock(),
			// grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
		)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		b = newBDDService(ctx, dpb.NewBDDSuiteServiceClient(conn))

		if pushgatewayEndpoint != "" {
			m = &metricsService{pusher: push.New(pushgatewayEndpoint, "manabie_bdd_tests")}
		}
	}

	var tp *tracesdk.TracerProvider
	if traceEnabled {
		exp, err := otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithEndpoint(otelEndpoint),
			otlptracegrpc.WithDialOption(grpc.WithBlock(), grpc.WithTimeout(10*time.Second)),
			otlptracegrpc.WithHeaders(map[string]string{
				"auth_token": "45fb58ce040a477d4cb9",
			}),
			otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
		)
		if err != nil {
			log.Fatalf("otlptracegrpc.New: %v", err) //nolint:gocritic
		}

		res, err := resource.New(ctx, resource.WithAttributes(
			semconv.ServiceNameKey.String("gandalf"),
		))
		if err != nil {
			log.Fatalf("resource.New: %v", err) //nolint:gocritic
		}

		tp = tracesdk.NewTracerProvider(
			tracesdk.WithSampler(tracesdk.AlwaysSample()),
			tracesdk.WithResource(res),
			tracesdk.WithSpanProcessor(tracesdk.NewBatchSpanProcessor(exp)),
		)
		defer func() {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err := tp.Shutdown(ctx); err != nil {
				log.Printf("tp.Shutdown: %v\n", err)
			}
		}()

		otel.SetTracerProvider(tp)
	}

	switch service {
	case "gandalf":
		c := &gandalf.Config{}
		configs.MustLoadConfig(
			context.Background(),
			commonConfigPath,
			configPath,
			secretsPath,
			&c.Config,
		)

		// ignore zeus' tests if zeus (and NATS Jetstream) itself is not enabled
		// note that OR op has higher precedence than AND op in godog tag parser
		if ts.Options.Tags == "" {
			ts.Options.Tags = "~@zeus"
		} else {
			ts.Options.Tags = fmt.Sprintf("%s && ~@zeus", ts.Options.Tags)
		}

		ts.TestSuiteInitializer = func(tsc *godog.TestSuiteContext) {
			managing.TestSuiteInitializer(c, fakeFirebaseAddr)(tsc)
			jprep.TestSuiteInitializer(c, fakeFirebaseAddr)(tsc)
			learning.TestSuiteInitializer(c, fakeFirebaseAddr)(tsc)
		}
		ts.ScenarioInitializer = func(sc *godog.ScenarioContext) {
			managing.ScenarioInitializer(c)(sc)
			jprep.ScenarioInitializer(c)(sc)
			learning.ScenarioInitializer(c)(sc)
		}

	case "repository.syllabus":
		c := &gandalf.Config{}
		configs.MustLoadConfig(
			context.Background(),
			commonConfigPath,
			configPath,
			secretsPath,
			&c.Config,
		)
		ts.TestSuiteInitializer = b.suiteHook(
			m.suiteHook(service,
				repoSyllabus.TestSuiteInitializer(c),
			),
		)
		ts.ScenarioInitializer = b.scenarioHook(
			m.scenarioHook(
				scenarioTracingInitializer(
					tp,
					repoSyllabus.ScenarioInitializer(c),
				),
			),
		)
	default:
		runtimeflag := common.RunTimeFlag{
			ApplicantID:  applicantID,
			FirebaseAddr: fakeFirebaseAddr,
			OtelEnabled:  traceEnabled,
		}
		if builder, ok := common.RegisteredTests[service]; ok {
			builtSuite := builder(commonConfigPath, configPath, secretsPath, runtimeflag)

			ts.TestSuiteInitializer = b.suiteHook(
				m.suiteHook(service, builtSuite.SuiteInitializer),
			)
			ts.ScenarioInitializer = b.scenarioHook(
				m.scenarioHook(
					scenarioTracingInitializer(tp, builtSuite.ScenarioInitializer),
				),
			)
			break
		}
		if scenarioIniter, ok := common.RegisteredTestWithCommonConnections[service]; ok {
			c := configs.MustLoadAll[common.Config](commonConfigPath, configPath, secretsPath)

			var conn common.Connections
			suiteIniter := func(ctx *godog.TestSuiteContext) {
				ctx.BeforeSuite(func() {
					conn = common.SetupConnections(c, fakeFirebaseAddr)
				})

				ctx.AfterSuite(func() {
					conn.CloseAllConnections()
				})
			}
			ts.TestSuiteInitializer = b.suiteHook(
				m.suiteHook(service, suiteIniter),
			)
			ts.ScenarioInitializer = b.scenarioHook(
				m.scenarioHook(
					scenarioTracingInitializer(tp, scenarioIniter(c, &conn)),
				),
			)
		}
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	os.Exit(ts.Run())
}

func printMessageIfServiceDeprecated(service string) {
	deprecatedServices := []string{"bob", "fatima", "gandalf", "yasuo"}
	for _, s := range deprecatedServices {
		if s == service {
			yellowColorCode := "\033[0;33m"
			resetColorCode := "\033[0m"
			log.Printf("%sWARNING%s: %q test suite is deprecated and will be phased out eventually. "+
				"Please write your tests in a directory of a bounded-context instead (even if you are adding new code to internal/%s). "+
				"Example of bounded-context directories: entryexitmgmt, mastermgmt, payment, ...", yellowColorCode, resetColorCode, service, service)
			return
		}
	}
}

func printMessageForDisabledGandalf() {
	yellowColorCode := "\033[0;33m"
	resetColorCode := "\033[0m"
	log.Printf("%sWARNING%s: gandalf test suite is disabled. Please move your tests to another test suite.", yellowColorCode, resetColorCode)
}

type testStatus string

const (
	passed testStatus = "PASSED"
	failed testStatus = "FAILED"
)

type skippedScenario struct {
	createdBy string
}

type skippedFeature struct {
	// If all is true, then a whole feature can be skipped.
	all       bool
	createdBy string

	// scenarios is used to store a list of scenario names
	// that can be skipped, in case of all is false.
	scenarios map[string]*skippedScenario
}

type bddService struct {
	totalFailedScenarios uint32
	totalPassedScenarios uint32

	// ctx is context used by godog cli
	ctx context.Context

	// instanceID is the ID generated by each time running the BDD test suite.
	// It will be used to record BDD test suite metrics, such as how many test
	// scenarios pass or failed, how long each scenario runs...etc.
	instanceID string

	bddSuiteClient dpb.BDDSuiteServiceClient

	// features is used to store test result of a feature, with
	// key is the feature id, and value is its test result, which can
	// be either passed or failed.
	//
	// Since all features are run concurrently, so we use sync.Map
	// instead of a plain Go map, to make it safely load and store
	// the test result between go routines.
	features sync.Map

	// skippedFeatures is used to store all features/scenarios that
	// can be skipped.
	skippedFeatures map[string]*skippedFeature
}

func newBDDService(ctx context.Context, c dpb.BDDSuiteServiceClient) *bddService {
	cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := c.RetrieveSkippedBDDTests(cctx, &dpb.RetrieveSkippedBDDTestsRequest{
		Repository: "github.com/manabie-com/backend",
	})
	if err != nil {
		log.Fatalf("cannot retrieve skipped BDD tests: %v", err)
	}

	b := &bddService{
		ctx:            ctx,
		bddSuiteClient: c,
	}
	if len(resp.SkippedBddTests) == 0 {
		return b
	}

	b.skippedFeatures = make(map[string]*skippedFeature)
	for _, t := range resp.SkippedBddTests {
		if b.skippedFeatures[t.FeaturePath] == nil {
			b.skippedFeatures[t.FeaturePath] = &skippedFeature{}
		}

		// if scenario name is empty, that means a whole feature will be skipped.
		if t.ScenarioName == "" {
			b.skippedFeatures[t.FeaturePath].all = true
			b.skippedFeatures[t.FeaturePath].createdBy = t.CreatedBy
		} else {
			if b.skippedFeatures[t.FeaturePath].scenarios == nil {
				b.skippedFeatures[t.FeaturePath].scenarios = make(map[string]*skippedScenario)
			}
			b.skippedFeatures[t.FeaturePath].scenarios[t.ScenarioName] = &skippedScenario{createdBy: t.CreatedBy}
		}
	}

	return b
}

func (b *bddService) canSkip(featurePath, scenarioName string) (string, bool) {
	if b.skippedFeatures == nil {
		return "", false
	}
	if b.skippedFeatures[featurePath] == nil {
		return "", false
	}
	if b.skippedFeatures[featurePath].all {
		return b.skippedFeatures[featurePath].createdBy, true
	}
	if v, ok := b.skippedFeatures[featurePath].scenarios[scenarioName]; ok {
		return v.createdBy, true
	}
	return "", false
}

func (b *bddService) suiteHook(next func(*godog.TestSuiteContext)) func(*godog.TestSuiteContext) {
	if b == nil {
		return next
	}

	type stat struct {
		Failed uint32 `json:"FAILED"`
		Passed uint32 `json:"PASSED"`
	}

	type stats struct {
		Feature  stat `json:"feature"`
		Scenario stat `json:"scenario"`
	}

	makeInstanceStats := func(totalFailedFeatures, totalPassedFeatures, totalFailedScenarios, totalPassedScenarios uint32) []byte {
		st, _ := json.Marshal(stats{
			Feature: stat{
				Failed: totalFailedFeatures,
				Passed: totalPassedFeatures,
			},
			Scenario: stat{
				Failed: totalFailedScenarios,
				Passed: totalPassedScenarios,
			},
		})
		return st
	}

	return func(sc *godog.TestSuiteContext) {
		sc.BeforeSuite(func() {
			ctx, cancel := context.WithTimeout(b.ctx, 10*time.Second)
			defer cancel()

			flavor, _ := structpb.NewStruct(map[string]interface{}{
				"env":             "CI",
				"organization":    "manabie",
				"pull_request_id": ciPullRequestID,
				"repository":      "backend",
				"run_id":          ciRunID,
				"actor":           ciActor,
			})
			resp, err := b.bddSuiteClient.AddInstance(ctx, &dpb.AddInstanceRequest{
				Name:   fmt.Sprintf("github.com/manabie-com/backend/pull/%s", ciPullRequestID),
				Stats:  makeInstanceStats(0, 0, 0, 0),
				Flavor: flavor,
				Tags:   []string{"backend"}, // TODO: set correct tags?
			})
			if err != nil {
				log.Printf("client.AddInstance: %v\n", err)
				return
			}

			// set instanceID for using later
			b.instanceID = resp.Id
		})

		sc.AfterSuite(func() {
			var (
				totalFailedFeatures uint32
				totalPassedFeatures uint32
			)

			var wg sync.WaitGroup
			b.features.Range(func(key, value any) bool {
				featureID := key.(string)
				status, ok := value.(testStatus)
				if !ok {
					return true
				}
				switch status {
				case failed:
					totalFailedFeatures++
				case passed:
					totalPassedFeatures++
				}

				wg.Add(1)
				go func() {
					defer wg.Done()

					ctx, cancel := context.WithTimeout(b.ctx, 30*time.Second)
					defer cancel()

					_, err := b.bddSuiteClient.SetFeatureStatus(ctx, &dpb.SetFeatureStatusRequest{
						Id:     featureID,
						Status: string(status),
					})
					if err != nil {
						log.Printf("cannot set feature status: id: %s: %v", featureID, err)
					}
				}()

				return true
			})
			wg.Wait()

			instanceStatus := failed
			if totalFailedFeatures == 0 && b.totalFailedScenarios == 0 {
				instanceStatus = passed
			}

			ctx, cancel := context.WithTimeout(b.ctx, 30*time.Second)
			defer cancel()

			_, err := b.bddSuiteClient.MarkInstanceEnded(ctx, &dpb.MarkInstanceEndedRequest{
				Id:     b.instanceID,
				Status: string(instanceStatus),
				Stats:  makeInstanceStats(totalFailedFeatures, totalPassedFeatures, b.totalFailedScenarios, b.totalPassedScenarios),
			})
			if err != nil {
				log.Printf("cannot mark instance ended: %v", err)
			}
		})

		next(sc)
	}
}

func (b *bddService) scenarioHook(next func(*godog.ScenarioContext)) func(*godog.ScenarioContext) {
	if b == nil {
		return next
	}

	return func(sc *godog.ScenarioContext) {
		sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			if b.instanceID == "" {
				return ctx, nil
			}

			if createdBy, ok := b.canSkip(sc.Uri, sc.Name); ok {
				return ctx, fmt.Errorf("feature %s is skipped by %s", sc.Uri, createdBy)
			}

			cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			featureResp, err := b.bddSuiteClient.AddFeature(cctx, &dpb.AddFeatureRequest{
				InstanceId: b.instanceID,
				Name:       uriToFeatureName(sc.Uri),
				Uri:        sc.Uri,
				Keyword:    "Feature",
			})
			if err != nil {
				log.Printf("cannot record current BDD feature: %v\n", err)
				return ctx, nil
			}

			steps, err := json.Marshal(sc.Steps)
			if err != nil {
				log.Printf("cannot marshal scenario steps: %v\n", err)
				return ctx, nil
			}
			tags := make([]string, 0, len(sc.Tags))
			for _, pt := range sc.Tags {
				tags = append(tags, pt.Name)
			}

			scenarioResp, err := b.bddSuiteClient.AddScenario(cctx, &dpb.AddScenarioRequest{
				FeatureId: featureResp.Id,
				Name:      sc.Name,
				Steps:     steps,
				Keyword:   "Scenario",
				Tags:      tags,
			})
			if err != nil {
				log.Printf("cannot record current BDD scenario: %v\n", err)
				return ctx, nil
			}

			featureID := featureResp.Id
			b.features.LoadOrStore(featureID, passed) // mark feature passed by default
			ctx = context.WithValue(ctx, featureIDKey{}, featureID)

			ctx = context.WithValue(ctx, scenarioIDKey{}, scenarioResp.Id)
			ctx = context.WithValue(ctx, scenarioURIKey{}, sc.Uri)

			return ctx, nil
		})

		sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
			if b.instanceID == "" {
				return ctx, err
			}

			if featureID, ok := ctx.Value(featureIDKey{}).(string); ok {
				if err != nil {
					if val, ok := b.features.Load(featureID); ok {
						if status, ok := val.(testStatus); ok && status != failed {
							// mark feature failed if any scenario failed
							b.features.Store(featureID, failed)
						}
					}
				}

				// Since we don't have a Feature (before and after) hook so we won't know
				// when a feature will end, so we just keep calling the MarkFeatureEnded API
				// to let it set the ended_at field to the database.
				//
				// Since all scenarios in this feature may not done yet, so we need to use
				// the suite after hook to update the feature status later, that's when all
				// scenarios in a feature are done.
				//
				// See AfterSuite hook function above.
				_, merr := b.bddSuiteClient.MarkFeatureEnded(ctx, &dpb.MarkFeatureEndedRequest{
					Id: featureID,
				})
				if merr != nil {
					log.Printf("cannot mark feature ended: %v\n", merr)
				}
			}
			if scenarioID, ok := ctx.Value(scenarioIDKey{}).(string); ok {
				var status testStatus
				if err != nil {
					status = failed
					atomic.AddUint32(&b.totalFailedScenarios, 1)
				} else {
					status = passed
					atomic.AddUint32(&b.totalPassedScenarios, 1)
				}

				_, merr := b.bddSuiteClient.MarkScenarioEnded(ctx, &dpb.MarkScenarioEndedRequest{
					Id:     scenarioID,
					Status: string(status),
				})
				if merr != nil {
					log.Printf("cannot mark scenario ended: %v\n", merr)
				}
			}
			return ctx, err
		})

		sc.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
			if b.instanceID == "" {
				return ctx, nil
			}

			if scenarioID, ok := ctx.Value(scenarioIDKey{}).(string); ok {
				resp, err := b.bddSuiteClient.AddStep(ctx, &dpb.AddStepRequest{
					ScenarioId: scenarioID,
					Name:       st.Text,
					Uri:        ctx.Value(scenarioURIKey{}).(string),
				})
				if err != nil {
					log.Printf("cannot record current BDD step: %v\n", err)
					return ctx, nil
				}

				ctx = context.WithValue(ctx, stepIDKey{}, resp.Id)
			}
			return ctx, nil
		})

		sc.StepContext().After(func(ctx context.Context, _ *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
			if b.instanceID == "" {
				return ctx, err
			}

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}
			if stepID, ok := ctx.Value(stepIDKey{}).(string); ok {
				if _, merr := b.bddSuiteClient.MarkStepEnded(ctx, &dpb.MarkStepEndedRequest{
					Id:      stepID,
					Status:  strings.ToUpper(status.String()),
					Message: errMessage,
				}); merr != nil {
					log.Printf("cannot mark step ended: %v\n", merr)
				}
			}
			return ctx, err
		})

		next(sc)
	}
}

type (
	featureIDKey   struct{}
	scenarioIDKey  struct{}
	scenarioURIKey struct{}
	stepIDKey      struct{}
)

// uriToFeatureName converts a BDD feature uri, e.g. bob/student_leave_live_lesson.feature,
// to a BDD feature name. For example it converts the uri above to: "bob: student leave live lesson"
func uriToFeatureName(uri string) string {
	slashIdx := strings.IndexByte(uri, '/')
	svcName := uri[:slashIdx]
	path := uri[slashIdx+1 : strings.IndexByte(uri, '.')]
	path = strings.ReplaceAll(path, "_", " ")
	return fmt.Sprintf("%s: %s", svcName, path)
}
