package tom

import (
	"context"
	"fmt"

	pb "github.com/manabie-com/backend/pkg/genproto/tom"
)

func (s *suite) teacherDoesNotReceivesNotification(ctx context.Context) error {
	var deviceToken string
	row := s.DB.QueryRow(ctx, "SELECT token FROM user_device_tokens WHERE user_id = $1", s.teacherWhoSentMessage)
	if err := row.Scan(&deviceToken); err != nil {
		return err
	}
	for try := 0; try < 5; try++ {
		resp, err := pb.NewChatServiceClient(s.Conn).RetrievePushedNotificationMessages(
			contextWithToken(ctx, s.TeacherToken),
			&pb.RetrievePushedNotificationMessageRequest{
				DeviceToken: deviceToken,
			})
		if err != nil {
			return err
		}
		if len(resp.GetMessages()) > 0 {
			return fmt.Errorf("exptect non notification, got %d", len(resp.GetMessages()))
		}
	}
	return nil
}
