package core

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"

	"github.com/jackc/pgtype"
)

func randomDBText() pgtype.Text {
	return database.Text(idutil.ULIDNow())
}
func dbNow() pgtype.Timestamptz {
	return database.Timestamptz(time.Now())
}
func dbText(str string) pgtype.Text {
	return database.Text(str)
}

func randomConversationMembers(cid string, len int) []*domain.ConversationMembers {
	ret := make([]*domain.ConversationMembers, 0, len)
	for i := 0; i < len; i++ {
		ret = append(ret, &domain.ConversationMembers{
			ConversationID: database.Text(cid),
			UserID:         database.Text(idutil.ULIDNow()),
		})
	}
	return ret
}
