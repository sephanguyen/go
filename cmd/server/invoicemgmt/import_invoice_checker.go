package invoicemgmt

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	invoiceAlert "github.com/manabie-com/backend/internal/invoicemgmt/alert"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/filestorage"
	invoiceService "github.com/manabie-com/backend/internal/invoicemgmt/services/invoice"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	bootstrap.RegisterJob("invoicemgmt_import_invoice_checker", ImportInvoiceChecker)
}

// ImportInvoiceCheckerConfig similar to the default invoicemgmt's Config object,
// but we remove the Postgres config to avoid its auto initialization.
type ImportInvoiceCheckerConfig struct {
	Common                configs.CommonConfig
	Issuers               []configs.TokenIssuerConfig
	Storage               configs.StorageConfig
	UnleashClientConfig   configs.UnleashClientConfig `yaml:"unleash_client"`
	PostgresV2            configs.PostgresConfigV2    `yaml:"postgres_v2"`
	InvoiceScheduleConfig InvoiceScheduleConfig       `yaml:"invoice_schedule_config"`
}

type InvoiceScheduleConfig struct {
	SlackChannel string `yaml:"slack_channel"`
	SlackWebhook string `yaml:"slack_webhook"`
}

// ImportInvoiceChecker is the main function for invoicemgmt_import_invoice_checker job.
func ImportInvoiceChecker(ctx context.Context, c ImportInvoiceCheckerConfig, rsc *bootstrap.Resources) error {
	zlogger := rsc.Logger().Sugar()

	zlogger.Info("start process scheduled ImportInvoiceChecker: %v", time.Now())
	unleashClient := rsc.Unleash()

	enableImproveCronImportInvoiceChecker, err := unleashClient.IsFeatureEnabled(constant.EnableImproveCronImportInvoiceChecker, c.Common.Environment)
	if err != nil {
		zlogger.Fatal(fmt.Sprintf("failed to check %s unleash feature flag", constant.EnableImproveCronImportInvoiceChecker), zap.Error(err))
	}

	// Init slack client
	slackClient := &alert.SlackImpl{
		WebHookURL: c.InvoiceScheduleConfig.SlackWebhook,
		HTTPClient: http.Client{Timeout: time.Duration(10) * time.Second},
	}

	// Init slack alert manager
	slackAlertManager := invoiceAlert.NewInvoiceScheduleSlackAlert(slackClient, invoiceAlert.Config{
		Environment:  c.Common.Environment,
		SlackChannel: c.InvoiceScheduleConfig.SlackChannel,
	})

	alertDependency := &alertDependency{
		unleashClient: unleashClient,
		zlogger:       zlogger,
		alertManager:  slackAlertManager,
		env:           c.Common.Environment,
	}

	// use the direct call on service method invoice checker
	if enableImproveCronImportInvoiceChecker {
		repositories := initRepositories()
		invoicemgmtDB := rsc.DBWith("invoicemgmt")
		paymentConn := rsc.GRPCDial("payment")
		internalOrderService := payment_pb.NewInternalServiceClient(paymentConn)

		// Initialize the file store to be used. The default storage is Google Cloud Storage
		fileStorageName := filestorage.GoogleCloudStorageService
		if strings.Contains(c.Storage.Endpoint, "minio") {
			fileStorageName = filestorage.MinIOService
		}

		fileStorage, err := filestorage.GetFileStorage(fileStorageName, &c.Storage)
		if err != nil {
			zlogger.Fatal(fmt.Sprintf("failed to init %s file storage", fileStorageName), zap.Error(err))
		}

		invoiceSvc := invoiceService.NewInvoiceModifierService(
			*zlogger,
			invoicemgmtDB,
			internalOrderService,
			fileStorage,
			getInvoiceServiceRepositories(repositories),
			unleashClient,
			c.Common.Environment,
			&utils.TempFileCreator{TempDirPattern: constant.InvoicemgmtTemporaryDir},
		)

		req := &invoice_pb.InvoiceScheduleCheckerRequest{
			InvoiceDate: timestamppb.New(time.Now()),
		}
		_, err = invoiceSvc.InvoiceScheduleChecker(ctx, req)
		if err != nil {
			zlogger.Error(err)
			notifySlack(alertDependency, invoiceAlert.Failed, err)
			return err
		}

		notifySlack(alertDependency, invoiceAlert.Success, nil)
		return nil
	}
	// use the previous approach for grpc endpoint
	invoiceMgmtConn := rsc.GRPCDial("invoicemgmt")

	req := &invoice_pb.InvoiceScheduleCheckerRequest{
		InvoiceDate: timestamppb.New(time.Now()),
	}
	_, err = invoice_pb.NewInternalServiceClient(invoiceMgmtConn).InvoiceScheduleChecker(ctx, req)
	if err != nil {
		zlogger.Error(err)
		notifySlack(alertDependency, invoiceAlert.Failed, err)
		return err
	}

	notifySlack(alertDependency, invoiceAlert.Success, nil)
	zlogger.Info("end process scheduled ImportInvoiceChecker: %v", time.Now())
	zlogger.Info("process done")
	return nil
}

type alertDependency struct {
	zlogger       *zap.SugaredLogger
	unleashClient unleashclient.ClientInstance
	alertManager  *invoiceAlert.InvoiceScheduleSlackAlert
	env           string
}

func notifySlack(d *alertDependency, status invoiceAlert.Status, resultErr error) {
	enableInvoiceScheduleCronJobAlert, err := d.unleashClient.IsFeatureEnabled(constant.EnableInvoiceScheduleCronJobAlert, d.env)
	if err != nil {
		d.zlogger.Fatal(fmt.Sprintf("failed to check %s unleash feature flag", constant.EnableImproveCronImportInvoiceChecker), zap.Error(err))
	}

	if !enableInvoiceScheduleCronJobAlert {
		return
	}

	switch status {
	case invoiceAlert.Success:
		err = d.alertManager.SendSuccessNotification()
	case invoiceAlert.Failed:
		err = d.alertManager.SendFailNotification(resultErr)
	}

	if err != nil {
		d.zlogger.Warn(err)
	}
}
