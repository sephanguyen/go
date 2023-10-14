package usermgmt

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

func RunMigrateCurrentGradeToGradeID(ctx context.Context, c *configurations.Config, inputGradePartnerIDs, organizationID string) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	zLogger = logger.NewZapLogger("debug", c.Common.Environment == "local")
	zLogger.Sugar().Info("-----START: Migrate current_grade to grade_id-----")
	defer zLogger.Sugar().Sync()

	dbPool, dbcancel, err := database.NewPool(ctx, zLogger, c.PostgresV2.Databases["bob"])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := dbcancel(); err != nil {
			zLogger.Error("dbcancel() failed", zap.Error(err))
		}
	}()

	ctx = auth.InjectFakeJwtToken(ctx, organizationID)

	gradePartnerIDs := strings.Split(inputGradePartnerIDs, "|")

	totalGradeEnum := 17
	if len(gradePartnerIDs) != totalGradeEnum {
		zLogger.Sugar().Fatalf("invalid len(gradePartnerIDs) expected: %d, but actual: %v", totalGradeEnum, len(gradePartnerIDs))
	}

	gradeRepo := repository.DomainGradeRepo{}
	grades, err := gradeRepo.GetByPartnerInternalIDs(ctx, dbPool, gradePartnerIDs)
	if err != nil {
		zLogger.Sugar().Fatalf("gradeRepo.GetByPartnerInternalIDs err: %v", err)
	}

	if err := database.ExecInTx(ctx, dbPool, func(ctx context.Context, tx pgx.Tx) error {
		err := deleteAllGradeOrg(ctx, tx, organizationID)
		if err != nil {
			return fmt.Errorf("deleteGradeOrg err: %v", err)
		}

		mapGrade := map[int]string{}
		for gradeVal, gradePartnerID := range gradePartnerIDs {
			for _, grade := range grades {
				if grade.PartnerInternalID().String() != gradePartnerID {
					continue
				}
				err := insertGradeOrg(ctx, tx, int32(gradeVal), grade.GradeID().String(), organizationID)
				if err != nil {
					return fmt.Errorf("insertGradeOrg err: %v", err)
				}
				mapGrade[gradeVal] = grade.GradeID().String()
				break
			}
		}

		for gradeVal, gradeID := range mapGrade {
			stmt := `
				UPDATE public.students SET grade_id = $1
				WHERE current_grade = $2
				AND deleted_at IS NULL 
				AND resource_path = $3;
			`
			_, err := tx.Exec(ctx, stmt, gradeID, gradeVal, organizationID)
			if err != nil {
				return fmt.Errorf("tx.Exec err: %v", err)
			}
		}

		return nil
	}); err != nil {
		zLogger.Sugar().Fatalf("database.ExecInTx err: %v", err)
	}

	zLogger.Sugar().Info("-----DONE: Migrate current_grade to grade_id-----")
}

func insertGradeOrg(ctx context.Context, db database.Ext, gradeVal int32, gradeID, resourcePath string) error {
	id := idutil.ULIDNow()
	stmt := `INSERT INTO grade_organization (grade_organization_id, grade_id, grade_value, resource_path) VALUES ($1, $2, $3, $4)`
	cmd, err := db.Exec(ctx, stmt, id, gradeID, gradeVal, resourcePath)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

func deleteAllGradeOrg(ctx context.Context, db database.Ext, resourcePath string) error {
	stmt := `UPDATE public.grade_organization SET deleted_at = now() WHERE resource_path = $1`
	_, err := db.Exec(ctx, stmt, resourcePath)
	return err
}
