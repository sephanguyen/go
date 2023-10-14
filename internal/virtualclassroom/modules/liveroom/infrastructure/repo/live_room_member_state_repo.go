package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type LiveRoomMemberStateRepo struct{}

func (l *LiveRoomMemberStateRepo) GetLiveRoomMemberStatesByChannelID(ctx context.Context, db database.QueryExecer, channelID string) (domain.LiveRoomMemberStates, error) {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomMemberStateRepo.GetLiveRoomMemberStatesByChannelID")
	defer span.End()

	filter := &domain.SearchLiveRoomMemberStateParams{
		ChannelID: channelID,
	}

	return l.GetLiveRoomMemberStatesWithParams(ctx, db, filter)
}

func (l *LiveRoomMemberStateRepo) GetLiveRoomMemberStatesWithParams(ctx context.Context, db database.QueryExecer, filter *domain.SearchLiveRoomMemberStateParams) (domain.LiveRoomMemberStates, error) {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomMemberStateRepo.GetLiveRoomMemberStatesWithParams")
	defer span.End()

	dto := &LiveRoomMemberState{}
	fields, _ := dto.FieldMap()
	dtoStates := LiveRoomMemberStates{}

	query := fmt.Sprintf(`SELECT %s FROM %s 
			WHERE deleted_at IS NULL `,
		strings.Join(fields, ","),
		dto.TableName(),
	)
	args := []interface{}{}

	if len(filter.ChannelID) > 0 {
		query += fmt.Sprintf(" AND channel_id = $%d ", len(args)+1)
		args = append(args, &filter.ChannelID)
	}
	if len(filter.UserIDs) > 0 {
		query += fmt.Sprintf(" AND user_id = ANY($%d) ", len(args)+1)
		args = append(args, &filter.UserIDs)
	}
	if len(filter.StateType) > 0 {
		query += fmt.Sprintf(" AND state_type = $%d ", len(args)+1)
		args = append(args, &filter.StateType)
	}

	err := database.Select(ctx, db, query, args...).ScanAll(&dtoStates)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return dtoStates.ToLiveRoomMemberStatesDomain(), nil
}

func (l *LiveRoomMemberStateRepo) BulkUpsertLiveRoomMembersState(ctx context.Context, db database.QueryExecer, channelID string, userIDs []string, stateType vc_domain.LearnerStateType, state *vc_domain.StateValue) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomMemberStateRepo.BulkUpsertLiveRoomMembersState")
	defer span.End()

	queueFn := func(b *pgx.Batch, memberState *LiveRoomMemberState) {
		fieldNames := database.GetFieldNamesExcepts(memberState, []string{"deleted_at"})
		args := database.GetScanFields(memberState, fieldNames)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO live_room_member_state (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT live_room_member_state_pk
			DO UPDATE SET 
				bool_value = $4, 
				string_array_value = $5, 
				updated_at = now()
			WHERE live_room_member_state.state_type = $3 
			AND live_room_member_state.deleted_at IS NULL`,
			strings.Join(fieldNames, ","),
			placeHolders,
		)
		b.Queue(query, args...)
	}

	b := &pgx.Batch{}
	now := time.Now()
	for _, userID := range userIDs {
		memberStateDTO := &LiveRoomMemberState{
			ChannelID:        database.Text(channelID),
			UserID:           database.Text(userID),
			StateType:        database.Text(string(stateType)),
			BoolValue:        database.Bool(state.BoolValue),
			StringArrayValue: database.TextArray(state.StringArrayValue),
			CreatedAt:        database.Timestamptz(now),
			UpdatedAt:        database.Timestamptz(now),
		}
		queueFn(b, memberStateDTO)
	}
	batchResult := db.SendBatch(ctx, b)
	defer batchResult.Close()

	for i := 0; i < b.Len(); i++ {
		_, err := batchResult.Exec()
		if err != nil {
			return fmt.Errorf("batchResult.Exec[%d]: %w", i, err)
		}
	}

	return nil
}

func (l *LiveRoomMemberStateRepo) UpdateAllLiveRoomMembersState(ctx context.Context, db database.QueryExecer, channelID string, stateType vc_domain.LearnerStateType, state *vc_domain.StateValue) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomMemberStateRepo.UpdateAllLiveRoomMembersState")
	defer span.End()

	memberStateDTO := &LiveRoomMemberState{
		StateType:        database.Text(string(stateType)),
		BoolValue:        database.Bool(state.BoolValue),
		StringArrayValue: database.TextArray(state.StringArrayValue),
	}

	query := fmt.Sprintf(`UPDATE %s SET
				bool_value = $3, 
				string_array_value = $4, 
				updated_at = now()
			WHERE channel_id = $1 
			AND state_type = $2
			AND deleted_at IS NULL`,
		memberStateDTO.TableName(),
	)

	_, err := db.Exec(ctx, query, &channelID, &memberStateDTO.StateType, &memberStateDTO.BoolValue, &memberStateDTO.StringArrayValue)
	if err != nil {
		return err
	}

	return nil
}

func (l *LiveRoomMemberStateRepo) CreateLiveRoomMemberState(ctx context.Context, db database.QueryExecer, channelID, userID string, stateType vc_domain.LearnerStateType, state *vc_domain.StateValue) error {
	ctx, span := interceptors.StartSpan(ctx, "LiveRoomMemberStateRepo.CreateLiveRoomMemberState")
	defer span.End()

	memberStateDTO := &LiveRoomMemberState{
		ChannelID:        database.Text(channelID),
		UserID:           database.Text(userID),
		StateType:        database.Text(string(stateType)),
		BoolValue:        database.Bool(state.BoolValue),
		StringArrayValue: database.TextArray(state.StringArrayValue),
	}
	if err := memberStateDTO.PreInsert(); err != nil {
		return err
	}

	fields := database.GetFieldNamesExcepts(memberStateDTO, []string{"deleted_at"})
	values := database.GetScanFields(memberStateDTO, fields)
	placeHolders := database.GeneratePlaceholders(len(fields))

	query := fmt.Sprintf(`INSERT INTO %s (%s)
			VALUES (%s) ON CONFLICT ON CONSTRAINT live_room_member_state_pk DO NOTHING `,
		memberStateDTO.TableName(),
		strings.Join(fields, ", "),
		placeHolders)

	_, err := db.Exec(ctx, query, values...)
	if err != nil {
		return err
	}

	return nil
}
