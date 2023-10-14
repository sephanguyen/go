package domain

import (
	"context"
	"fmt"
)

type StreamingProviderPort interface {
	CreateStreamingRoom(ctx context.Context, room *VirtualRoom) (roomID string, err error)
	//	Implementations of methods not be modified room entity
	AttendeeConsumeAStreamingSlot(ctx context.Context, room *VirtualRoom, attendeeID string) error
}

type StreamingProvider struct {
	StreamingRoomID     string
	TotalStreamingSlots int

	// ports
	streamingProviderPort StreamingProviderPort
}

func (s *StreamingProvider) isValid() error {
	if len(s.StreamingRoomID) == 0 {
		return fmt.Errorf("streaming room ID cannot be empty")
	}

	return nil
}
