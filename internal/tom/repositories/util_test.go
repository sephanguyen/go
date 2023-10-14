package repositories

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
)

func randomText() pgtype.Text {
	return database.Text(idutil.ULIDNow())
}
func randomTime() pgtype.Timestamptz {
	return database.Timestamptz(time.Now())
}
