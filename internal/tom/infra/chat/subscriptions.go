package chat

import (
	"context"
	"time"

	pb "github.com/manabie-com/backend/pkg/genproto/tom"
)

func (rcv *Server) HandleInternalBroadcast(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	req := &pb.InternalSendMessageRequest{}
	err := req.Unmarshal(data)
	if err != nil {
		rcv.logger.Error(err.Error())
		return false, err
	}

	if req.NodeReceive == nil {
		return false, nil
	}

	userIDs, ok := req.NodeReceive[rcv.hostName]
	if !ok {
		return false, nil
	}

	rcv.PushLocal(ctx, userIDs.Ids, req.Event)
	return false, nil
}
