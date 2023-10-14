package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/tom/domain/core"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.opencensus.io/trace"
)

type ConversationLocationRepo struct {
}

func (r *ConversationLocationRepo) RemoveLocationsForConversation(
	ctx context.Context, db database.QueryExecer, conversationID string, locations []string) error {
	ctx, span := trace.StartSpan(ctx, "ConversationLocatioRepo.RemoveLocationsForConversation")
	defer span.End()
	updateStmt := `update conversation_locations set deleted_at=now() where conversation_id=$1 and location_id=ANY($2)`
	_, err := db.Exec(ctx, updateStmt, database.Text(conversationID), database.TextArray(locations))
	if err != nil {
		return fmt.Errorf("db.Exec %w", err)
	}
	return nil
}
func (r *ConversationLocationRepo) FindByConversationIDs(
	ctx context.Context, db database.QueryExecer, conversationIDs pgtype.TextArray) (map[string][]core.ConversationLocation, error) {
	ctx, span := trace.StartSpan(ctx, "ConversationLocationRepo.FindByLocationIDs")
	defer span.End()

	var e core.ConversationLocation
	allfield, _ := e.FieldMap()
	query := `select %s from conversation_locations where conversation_id=ANY($1) and deleted_at is null`
	rows, err := db.Query(ctx, fmt.Sprintf(query, strings.Join(allfield, ",")), conversationIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query %w", err)
	}
	defer rows.Close()
	convIDLocationsMap := make(map[string][]core.ConversationLocation)

	for rows.Next() {
		var e core.ConversationLocation
		err := rows.Scan(database.GetScanFields(&e, allfield)...)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan %w", err)
		}
		convIDLocationsMap[e.ConversationID.String] = append(convIDLocationsMap[e.ConversationID.String], e)
	}
	return convIDLocationsMap, nil
}

func (r *ConversationLocationRepo) GetAllLocations(ctx context.Context, db database.QueryExecer, userID string) ([]*core.ConversationLocation, error) {
	ctx, span := trace.StartSpan(ctx, "ConversationLocationRepo.GetAllLocations")
	defer span.End()

	fields, _ := (&core.ConversationLocation{}).FieldMap()
	query := fmt.Sprintf(`select cl.%s from conversation_locations  cl 
join conversation_members cm on cl.conversation_id = cm.conversation_id 
join conversations c on c.conversation_id = cl.conversation_id 
where c.status = 'CONVERSATION_STATUS_NONE' AND (c.conversation_type <> 'CONVERSATION_LESSON' OR c.conversation_type IS NULL)
and cm.status = 'CONVERSATION_STATUS_ACTIVE' and cm.user_id = $1`, strings.Join(fields, ", cl."))
	rows, err := db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("db.Query %w", err)
	}
	defer rows.Close()

	var res []*core.ConversationLocation
	for rows.Next() {
		var e core.ConversationLocation
		if err = rows.Scan(database.GetScanFields(&e, fields)...); err != nil {
			return nil, fmt.Errorf("rows.Scan %w", err)
		}
		res = append(res, &e)
	}

	return res, nil
}

func (r *ConversationLocationRepo) BulkUpsert(ctx context.Context, db database.QueryExecer, conversationLoc []core.ConversationLocation) error {
	ctx, span := trace.StartSpan(ctx, "ConversationLocationRepo.BulkUpsert")
	defer span.End()
	locIDs := make([]string, 0, len(conversationLoc))
	for _, ent := range conversationLoc {
		locIDs = append(locIDs, ent.LocationID.String)
	}
	findAccessPath := `select access_path, location_id from locations where location_id=any($1) and deleted_at is null`
	rows, err := db.Query(ctx, findAccessPath, database.TextArray(locIDs))
	if err != nil {
		return fmt.Errorf("db.Query %w", err)
	}
	defer rows.Close()
	locAccessPathMap := map[string]string{}
	for rows.Next() {
		var loc, accp string
		err := rows.Scan(&accp, &loc)
		if err != nil {
			return fmt.Errorf("rows.Scan %w", err)
		}
		locAccessPathMap[loc] = accp
	}
	rows.Close()
	for idx := range conversationLoc {
		location := conversationLoc[idx].LocationID.String
		accp, exist := locAccessPathMap[location]
		if !exist {
			return fmt.Errorf("cannot find access path for location %s", location)
		}
		err := conversationLoc[idx].AccessPath.Set(accp)
		if err != nil {
			return fmt.Errorf("AccessPath.Set %w", err)
		}
	}

	queueFn := func(b *pgx.Batch, e core.ConversationLocation) {
		fieldNames, values := e.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO conversation_locations (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT conversation_locations_pk
			DO UPDATE SET deleted_at = NULL, updated_at = $5`,
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		b.Queue(query, values...)
	}

	b := &pgx.Batch{}

	for idx := range conversationLoc {
		queueFn(b, conversationLoc[idx])
	}

	result := db.SendBatch(ctx, b)
	defer result.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := result.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}

	return nil
}
