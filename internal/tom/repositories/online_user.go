package repositories

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	"go.uber.org/zap"

	"github.com/jackc/pgtype"
)

type OnlineUserRepo struct {
	*OnlineUserDBRepo
	*OnlineUserCacheRepo
}

func (r *OnlineUserRepo) Find(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray, since pgtype.Timestamptz, msg *pb.Event) (mapNodeUserIDs map[pgtype.Text][]string, err error) {
	ctx, span := interceptors.StartSpan(ctx, "OnlineUserRepo.Find")
	defer span.End()

	mapNodeUserIDs, pgOfflineUserIDs := r.OnlineUserCacheRepo.Find(userIDs, since)
	msgID := ""
	if msg != nil {
		if newMsg := msg.GetEventNewMessage(); newMsg != nil {
			msgID = newMsg.GetMessageId()
		}
	}
	userInCache := make([]string, 0)
	for _, users := range mapNodeUserIDs {
		userInCache = append(userInCache, users...)
	}
	ctxzap.Info(
		ctx,
		"OnlineUserRepo.Find OnlineUserCacheRepo.Find find user in cache",
		zap.String("message_id", msgID),
		zap.Strings("user_ids", userInCache),
	)
	if len(pgOfflineUserIDs.Elements) == 0 && len(mapNodeUserIDs) > 0 {
		return mapNodeUserIDs, nil
	}

	nodeUserIDs, err := r.OnlineUserDBRepo.Find(ctx, db, pgOfflineUserIDs, since)
	if err != nil {
		return nil, fmt.Errorf("r.OnlineUserDBRepo.Find: %w", err)
	}

	mapUserIDNodes := map[pgtype.Text][]string{}
	for node, userIDs := range nodeUserIDs {
		mapNodeUserIDs[node] = append(mapNodeUserIDs[node], userIDs...)

		for _, userID := range userIDs {
			pgUserID := database.Text(userID)
			mapUserIDNodes[pgUserID] = append(mapUserIDNodes[pgUserID], node.String)
			r.OnlineUserCacheRepo.Add(pgUserID, database.TextArray(mapUserIDNodes[pgUserID]))
		}
	}

	totalUserIDs := make([]string, 0)
	for _, userIDs := range mapNodeUserIDs {
		totalUserIDs = append(totalUserIDs, userIDs...)
	}
	ctxzap.Info(
		ctx,
		"OnlineUserRepo.Find total user at the end",
		zap.String("message_id", msgID),
		zap.Strings("user_ids", totalUserIDs),
	)
	return mapNodeUserIDs, nil
}
