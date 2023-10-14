package usermgmt

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

func RunMigrateUserPhoneNumber(ctx context.Context, bobCfg *configurations.Config, organizationID, userType string) {
	var phoneNumberType, userGroup string

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)

	defer stop()

	zLogger = logger.NewZapLogger("debug", bobCfg.Common.Environment == "local")
	zLogger.Sugar().Infof("-----START: Migrate userPhoneNumber - userType: %v-----", userType)
	defer zLogger.Sugar().Sync()

	dbPool, dbcancel, err := database.NewPool(ctx, zLogger, bobCfg.PostgresV2.Databases["bob"])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := dbcancel(); err != nil {
			zLogger.Error("dbcancel() failed", zap.Error(err))
		}
	}()

	org, err := (&repository.OrganizationRepo{}).Find(ctx, dbPool, database.Text(strings.TrimSpace(organizationID)))
	if err != nil {
		zLogger.Fatal(fmt.Sprintf("find organization %s failed: %s", organizationID, err.Error()))
	}

	ctx = auth.InjectFakeJwtToken(ctx, org.OrganizationID.String)

	switch userType {
	case student:
		phoneNumberType = entity.StudentPhoneNumber
		userGroup = `JOIN students s ON s.student_id  = u.user_id  `

	case staff:
		phoneNumberType = entity.StaffPrimaryPhoneNumber
		userGroup = `JOIN staff s ON s.staff_id  = u.user_id `

	case parent:
		phoneNumberType = entity.ParentPrimaryPhoneNumber
		userGroup = `JOIN parents s ON s.parent_id  = u.user_id `

	default:
		zLogger.Sugar().Infof("Must have userType !!!")
		return
	}

	stmt := fmt.Sprintf(`
			INSERT INTO user_phone_number ( user_phone_number_id, user_id, phone_number, updated_at, created_at, resource_path, "type") 
			SELECT gen_random_uuid(), u.user_id ,u.phone_number, now(), now(), u.resource_path, $1
			FROM users u
			%v
			WHERE NOT EXISTS (
				select *
				FROM user_phone_number upn
				WHERE upn.user_id = u.user_id
			) 
		  	AND u.phone_number is not null AND u.phone_number ~ '^[0-9+]+$'
		`, userGroup)

	err = database.ExecInTx(ctx, dbPool, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, stmt, phoneNumberType)
		return err
	})
	if err != nil {
		zLogger.Sugar().Fatalf("RunMigrateUserPhoneNumber err: %v", err)
	}

	zLogger.Sugar().Infof("-----DONE: Migrate userPhoneNumber - userType: %v-----", userType)
}
