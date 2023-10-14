package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

// GithubEvent struct
type GithubEvent struct {
}

// AddEventData to Add the data collecting from github events
func (g *GithubEvent) AddEventRawData(ctx context.Context, db database.QueryExecer, eventName string, data pgtype.JSONB) error {
	querystm := "INSERT INTO public.github_raw_data (event_name, data)  VALUES ($1, $2)"
	_, err := db.Exec(ctx, querystm, &eventName, &data)
	if err != nil {
		return fmt.Errorf("AddEventRawData b.Exec: %v", err)
	}
	return nil
}
