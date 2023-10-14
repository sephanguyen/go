package rlsscan

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

func init() {
	bootstrap.RegisterJob("rls_check", startRLSScan)
}

type ConfigV2 struct {
	Common          configs.CommonConfig
	Issuers         []configs.TokenIssuerConfig
	RLSScanPostgres configs.PostgresConfigV2 `yaml:"postgres_v2"`
}

func handleScan(ctx context.Context, db *pgxpool.Pool, zapLogger *zap.Logger) {
	// init logger
	err := database.ScanRLS(ctx, db)
	if err != nil {
		zapLogger.Error("rls_scan is error", zap.Error(err))
	}
	err = database.ScanPostgresAC(ctx, db)
	if err != nil {
		zapLogger.Error("rls_scan_ac is error", zap.Error(err))
	}
	zapLogger.Info("rls_scan is running")
}

func startRLSScan(ctx context.Context, c ConfigV2, rsc *bootstrap.Resources) error {
	l := rsc.Logger()
	targetDB := c.Common.Name
	switch c.Common.Name {
	case "enigma", "notificationmgmt", "shamir", "virtualclassroom", "usermgmt", "yasuo":
		targetDB = "bob"
	case "payment", "discount":
		targetDB = "fatima"
	case "conversationmgmt":
		targetDB = "tom"
	case "spike":
		targetDB = "notificationmgmt"
	}
	dbconf, ok := c.RLSScanPostgres.Databases[targetDB]
	if !ok {
		return fmt.Errorf("missing config for database %s", targetDB)
	}
	dbconf.MaxConns = 1
	db, dbcancel, err := database.NewPool(ctx, l, dbconf)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %s", err)
	}
	defer func() {
		if err := dbcancel(); err != nil {
			l.Error("dbcancel() failed", zap.Error(err))
		}
	}()
	defer db.Close()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	handleScan(ctx, db, l)
	for {
		select {
		case <-ctx.Done():
			l.Info("rls_scan is stopped")
			return nil
		case <-ticker.C:
			handleScan(ctx, db, l)
		}
	}
}
