package helper

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/eibanam/communication/util"
	legacytpb "github.com/manabie-com/backend/pkg/genproto/tom"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/durationpb"
)

var (
	MsgTypeImage     = "image"
	MsgTypePdf       = "pdf"
	MsgTypeText      = "text"
	MsgTypeHyperlink = "hyperlink"
	MsgTypeSystem    = "system"
)

func (h *CommunicationHelper) SendTextMsgToConversation(ctx context.Context, token, content, convID string) error {
	req := &legacytpb.SendMessageRequest{
		ConversationId: convID,
	}

	req.Message = content
	req.Type = legacytpb.MESSAGE_TYPE_TEXT

	ctx, cancel := util.ContextWithTokenAndTimeOut(ctx, token)
	defer cancel()

	_, err := legacytpb.NewChatServiceClient(h.tomGRPCConn).SendMessage(ctx, req)
	if err != nil {
		return err
	}
	return err
}

func (h *CommunicationHelper) SendMsgToConversation(ctx context.Context, token, msgType, convID string) error {
	req := &legacytpb.SendMessageRequest{
		ConversationId: convID,
	}

	switch msgType {
	case MsgTypeText:
		req.Message = "hello world"
		req.Type = legacytpb.MESSAGE_TYPE_TEXT
	case MsgTypeHyperlink:
		req.Message = "https://google.com"
		req.Type = legacytpb.MESSAGE_TYPE_TEXT
	case MsgTypeImage: // mobile is treating image == file
		req.Type = legacytpb.MESSAGE_TYPE_FILE
		url, err := h.generateURLforFile(ctx, token, MsgTypeImage)
		if err != nil {
			return err
		}
		req.Message = fmt.Sprintf("<a href=\"%s\">%s</a>", url, "testimage.jpg")
	case MsgTypePdf:
		req.Type = legacytpb.MESSAGE_TYPE_FILE
		url, err := h.generateURLforFile(ctx, token, MsgTypePdf)
		if err != nil {
			return err
		}
		req.Message = fmt.Sprintf("<a href=\"%s\">%s</a>", url, "testpdf.pdf")
	default:
		return fmt.Errorf("unsupported msg type %s", msgType)
	}

	ctx, cancel := util.ContextWithTokenAndTimeOut(ctx, token)
	defer cancel()

	_, err := legacytpb.NewChatServiceClient(h.tomGRPCConn).SendMessage(ctx, req)
	if err != nil {
		return err
	}
	return err
}

type Message struct {
	Content        string
	ConversationID string
	Sender         string
	Type           legacytpb.MessageType
	TargetUser     string
}

func (h *CommunicationHelper) CheckMsgType(msgType string, givenType legacytpb.MessageType) error {
	var wantType legacytpb.MessageType
	switch msgType {
	case MsgTypeText:
		wantType = legacytpb.MESSAGE_TYPE_TEXT
	case MsgTypeHyperlink:
		wantType = legacytpb.MESSAGE_TYPE_TEXT
	case MsgTypeImage: // mobile is treating image == file
		wantType = legacytpb.MESSAGE_TYPE_FILE
	case MsgTypePdf:
		wantType = legacytpb.MESSAGE_TYPE_FILE
	case MsgTypeSystem:
		wantType = legacytpb.MESSAGE_TYPE_SYSTEM
	default:
		return fmt.Errorf("cannot check unknown msg type: %s", msgType)
	}
	if givenType != wantType {
		return fmt.Errorf("want %s, has %s", wantType.String(), givenType.String())
	}
	return nil
}

func (h *CommunicationHelper) ListConversationMessages(ctx context.Context, token string, convID string) ([]Message, error) {
	res, err := legacytpb.NewChatServiceClient(h.tomGRPCConn).ConversationDetail(util.ContextWithToken(ctx, token), &legacytpb.ConversationDetailRequest{
		ConversationId: convID,
		Limit:          10,
	})
	if err != nil {
		return nil, err
	}
	var ret = make([]Message, 0, len(res.GetMessages()))
	for _, item := range res.GetMessages() {
		ret = append(ret, Message{
			Content:        item.Content,
			ConversationID: item.ConversationId,
			Sender:         item.UserId,
			Type:           item.GetType(),
			TargetUser:     item.GetTargetUser(),
		})
	}
	return ret, nil
}
func (h *CommunicationHelper) MsgChanFromStream(ctx context.Context, stream legacytpb.ChatService_SubscribeV2Client) chan *legacytpb.MessageResponse {
	msgChan := make(chan *legacytpb.MessageResponse)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				res, err := stream.Recv()
				if err != nil {
					close(msgChan)
					fmt.Printf("recv: %s\n", err)
					return
				}
				// ignore ping msg
				if res.GetEvent().GetEventPing() != nil {
					continue
				}
				msgRes := res.Event.GetEventNewMessage()
				msgChan <- msgRes
			}
		}
	}()
	return msgChan
}

func (h *CommunicationHelper) DrainMsgFromStream(stream legacytpb.ChatService_SubscribeV2Client) (Message, error) {
	msgChan := make(chan *legacytpb.MessageResponse)

	go func() {
		for {
			res, err := stream.Recv()
			if err != nil {
				return
			}
			// ignore ping msg
			if res.GetEvent().GetEventPing() != nil {
				continue
			}
			msgRes := res.Event.GetEventNewMessage()
			msgChan <- msgRes
			return
		}
	}()

	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
		return Message{}, fmt.Errorf("no new msg signal from upstream")
	case newMsg := <-msgChan:
		return Message{
			Content:        newMsg.Content,
			ConversationID: newMsg.ConversationId,
			Sender:         newMsg.UserId,
			Type:           newMsg.Type,
			TargetUser:     newMsg.TargetUser,
		}, nil
	}
}
func (h *CommunicationHelper) ConnectChatStreamWithHashAndPings(ctx context.Context, token string, hash string, pingPerSec int) error {
	ctx = util.ContextWithToken(ctx, token)

	chatSvc := legacytpb.NewChatServiceClient(h.tomGRPCConn)
	ctx = metadata.AppendToOutgoingContext(ctx, "x-chat-userhash", hash)
	reconnect := func() (string, legacytpb.ChatService_SubscribeV2Client, error) {
		streamV2, err := chatSvc.SubscribeV2(ctx, &legacytpb.SubscribeV2Request{})
		if err != nil {
			return "", nil, err
		}

		sessionID := ""
		resp, err := streamV2.Recv()
		if err != nil {
			return "", nil, err
		}

		if resp.Event.GetEventPing() == nil {
			return "", nil, fmt.Errorf("stream must receive pingEvent first")
		}
		sessionID = resp.Event.GetEventPing().SessionId
		return sessionID, streamV2, nil
	}

	for i := 0; i < pingPerSec; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					break
				}
				sessionID, stream, err := reconnect()
				if err != nil {
					time.Sleep(2 * time.Second)
					continue
				}
			pingloop:
				for {
					select {
					case <-ctx.Done():
						stream.CloseSend()
						return
					default:
						ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), token)
						defer cancel()
						_, err := chatSvc.PingSubscribeV2(ctx, &legacytpb.PingSubscribeV2Request{SessionId: sessionID})
						if err != nil {
							fmt.Printf("failed to ping using subscribe v2 api: %s, reconnecting\n", err)
							stream.CloseSend()
							break pingloop
							// continue
							// return
						}
						time.Sleep(1 * time.Second)
					}
				}
			}

		}()
	}

	return nil
}

func (h *CommunicationHelper) ConnectChatStreamWithHash(ctx context.Context, token string, hash string) (legacytpb.ChatService_SubscribeV2Client, error) {
	ctx = util.ContextWithToken(ctx, token)

	chatSvc := legacytpb.NewChatServiceClient(h.tomGRPCConn)
	ctx = metadata.AppendToOutgoingContext(ctx, "x-chat-userhash", hash)

	streamV2, err := chatSvc.SubscribeV2(ctx, &legacytpb.SubscribeV2Request{})
	if err != nil {
		return nil, err
	}

	sessionID := ""
	resp, err := streamV2.Recv()
	if err != nil {
		return nil, err
	}

	if resp.Event.GetEventPing() == nil {
		return nil, fmt.Errorf("stream must receive pingEvent first")
	}

	sessionID = resp.Event.GetEventPing().SessionId

	go func() {
		for {
			ctx, cancel := util.ContextWithTokenAndTimeOut(context.Background(), token)
			defer cancel()
			_, err := chatSvc.PingSubscribeV2(ctx, &legacytpb.PingSubscribeV2Request{SessionId: sessionID})
			if err != nil {
				fmt.Printf("failed to ping using subscribe v2 api: %s\n", err)
				return
			}
			time.Sleep(2 * time.Second)
		}
	}()

	return streamV2, nil
}

func (h *CommunicationHelper) ConnectChatStream(ctx context.Context, token string) (legacytpb.ChatService_SubscribeV2Client, error) {
	ctx = util.ContextWithToken(ctx, token)

	chatSvc := legacytpb.NewChatServiceClient(h.tomGRPCConn)

	streamV2, err := chatSvc.SubscribeV2(ctx, &legacytpb.SubscribeV2Request{})
	if err != nil {
		return nil, err
	}

	sessionID := ""
	resp, err := streamV2.Recv()
	if err != nil {
		return nil, err
	}

	if resp.Event.GetEventPing() == nil {
		return nil, fmt.Errorf("stream must receive pingEvent first")
	}

	sessionID = resp.Event.GetEventPing().SessionId

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			ctx, cancel := util.ContextWithTokenAndTimeOut(ctx, token)
			defer cancel()
			_, err := chatSvc.PingSubscribeV2(ctx, &legacytpb.PingSubscribeV2Request{SessionId: sessionID})
			if err != nil {
				fmt.Printf("failed to ping using subscribe v2 api: %s\n", err)
				return
			}
			time.Sleep(2 * time.Second)
		}
	}()

	return streamV2, nil
}

func (h *CommunicationHelper) generateURLforFile(ctx context.Context, token string, fileType string) (string, error) {
	resumableUpload := &bpb.ResumableUploadURLRequest{
		AllowOrigin:   "http://localhost:3001",
		ContentType:   "",
		Expiry:        durationpb.New(time.Second * 1800),
		FileExtension: "",
		PrefixName:    "",
	}
	switch fileType {
	case MsgTypeImage:
		resumableUpload.FileExtension = "jpg"
		resumableUpload.PrefixName = "testimage"
	case MsgTypePdf:
		resumableUpload.FileExtension = MsgTypePdf
		resumableUpload.PrefixName = "testpdf"
	default:
		return "", fmt.Errorf("unsupported file type %s", fileType)
	}
	ctx2, cancel := util.ContextWithTokenAndTimeOut(ctx, token)
	defer cancel()

	res, err := bpb.NewUploadServiceClient(h.bobGRPCConn).GenerateResumableUploadURL(ctx2, resumableUpload)
	if err != nil {
		return "", err
	}
	return res.ResumableUploadUrl, nil
}
