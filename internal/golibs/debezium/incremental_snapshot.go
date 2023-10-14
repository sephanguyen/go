package debezium

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"

	"github.com/jackc/pgtype"
)

// incremental snapshot is debezium's feature which snapshot new sync table to kafka
// Working with hephaestus to upsert kafka connect
// it allow us to snapshot new added tables without recreating new source connector
// of course we only allow it to snapshot not existed synced tables by checking topics data in kafka

type DataCollection struct {
	SourceID string
	Tables   []string
	RepName  string
}

func (dc DataCollection) String() string {
	for i := range dc.Tables {
		dc.Tables[i] = fmt.Sprintf(`"%s"`, dc.Tables[i])
	}
	tbs := strings.Join(dc.Tables, ",")
	return fmt.Sprintf(`{"data-collections": [%s]}`, tbs)
}

// snapshotWaitDuration is the wait duration to wait for debezium task in kafka connect to be ready.
// In unit tests, it is reduced to speed up tests.
var snapshotWaitDuration = time.Second * 10

func IncrementalSnapshot(ctx context.Context, db database.QueryExecer, snapshotSignalTable string, data DataCollection) error {
	// wait for the debezium task in kafka connect to be ready
	time.Sleep(snapshotWaitDuration)

	err := try.Do(func(attempt int) (bool, error) {
		fmt.Println("waiting for replication slot to active", data.RepName)
		if isActive, err := isReplicationSlotActive(ctx, db, data.RepName); err != nil || !isActive {
			time.Sleep(300 * time.Millisecond)
			return true, err
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	sourceID := data.SourceID

	idText := database.Text(fmt.Sprintf("%s%s", sourceID, idutil.ULIDNow()))
	tpText := database.Text("execute-snapshot")
	dataCollectionText := database.Text(data.String())

	sql := fmt.Sprintf("INSERT INTO %s VALUES ($1, $2, $3)", snapshotSignalTable)

	cmd, err := db.Exec(ctx, sql, idText, tpText, dataCollectionText)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("cannot insert signal to trigger snapshot new synced table")
	}

	return nil
}

func isReplicationSlotActive(ctx context.Context, db database.QueryExecer, repName string) (bool, error) {
	sql := `SELECT active FROM pg_replication_slots WHERE slot_name=$1`
	var active pgtype.Bool
	err := db.QueryRow(ctx, sql, database.Text(repName)).Scan(&active)
	if err != nil {
		return false, err
	}
	return active.Bool, nil
}
