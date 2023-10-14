package invoicemgmt

import (
	"context"
	"time"

	"github.com/manabie-com/backend/cmd/server/invoicemgmt"
	"github.com/manabie-com/backend/features/common"
	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
)

func (s *suite) setupBootstrapResourceForImportInvoiceChecker(c *common.Config) *bootstrap.Resources {
	rsc := bootstrap.NewResources().WithLoggerC(&c.Common)
	rsc.WithServiceName("invoicemgmt")
	rsc.WithUnleashC(&c.UnleashClientConfig).Unleash()
	rsc.WithDatabase(map[string]*database.DBTrace{
		"invoicemgmt": s.InvoiceMgmtDBTrace,
	})

	return rsc
}

func (s *suite) cronjobRunFuncImportInvoiceChecker(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsc := s.setupBootstrapResourceForImportInvoiceChecker(s.Cfg)
	defer rsc.Cleanup() //nolint:errcheck
	err := invoicemgmt.ImportInvoiceChecker(
		ctx,
		invoicemgmt.ImportInvoiceCheckerConfig{
			Common:              s.Cfg.Common,
			UnleashClientConfig: s.Cfg.UnleashClientConfig,
		},
		rsc,
	)

	rsc.WithDatabase(map[string]*database.DBTrace{
		"invoicemgmt": s.InvoiceMgmtPostgresDBTrace,
	})

	time.Sleep(invoiceConst.ReselectSleepDuration)

	return StepStateToContext(ctx, stepState), err
}
