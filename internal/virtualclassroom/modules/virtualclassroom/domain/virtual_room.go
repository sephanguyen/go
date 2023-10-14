package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
)

type VirtualRoom struct {
	EndedAt           *time.Time
	StreamingProvider *StreamingProvider // required field
	Materials         Materials

	// states of VirtualRoom
	// Types that are assignable to PresentMaterialState:
	//	*VideoPresentMaterialState
	//	*PDFPresentMaterialState
	PresentMaterialState presentMaterialState
	CurrentPolling       *CurrentPolling
	RecordingState       *RecordingState
	AttendeeStates       AttendeeStates

	// port
	userModulePort UserModulePort
}

func (v *VirtualRoom) GetVideoPresentMaterialState() *VideoPresentMaterialState {
	if x, ok := v.PresentMaterialState.(*VideoPresentMaterialState); ok {
		return x
	}
	return nil
}

func (v *VirtualRoom) GetPDFPresentMaterialState() *PDFPresentMaterialState {
	if x, ok := v.PresentMaterialState.(*PDFPresentMaterialState); ok {
		return x
	}
	return nil
}

func (v *VirtualRoom) isValid(ctx context.Context) error {
	if v.StreamingProvider == nil {
		return fmt.Errorf("streaming provider cannot be empty")
	}
	if err := v.StreamingProvider.isValid(); err != nil {
		return fmt.Errorf("invalid streaming provider: %w", err)
	}

	if v.Materials != nil {
		if err := v.Materials.isValid(); err != nil {
			return fmt.Errorf("invalid materials: %w", err)
		}
	}

	if err := v.isValidPresentMaterialState(); err != nil {
		return err
	}

	if v.CurrentPolling != nil {
		if err := v.CurrentPolling.isValid(); err != nil {
			return fmt.Errorf("ivalid current polling: %w", err)
		}
	}

	if err := v.isValidRecordingState(ctx); err != nil {
		return err
	}

	if err := v.isValidAttendeeStates(ctx); err != nil {
		return err
	}

	return nil
}

func (v *VirtualRoom) isValidRecordingState(ctx context.Context) error {
	if v.RecordingState == nil {
		return nil
	}

	if err := v.RecordingState.isValid(); err != nil {
		return fmt.Errorf("invalid recording state: %w", err)
	}
	if v.RecordingState.Creator != nil {
		ids, err := v.userModulePort.CheckExistedUserIDs(ctx, []string{*v.RecordingState.Creator})
		if err != nil {
			return fmt.Errorf("RecordingState.Creator: userModulePort.IsExistedUserIDs: %w", err)
		}
		if len(ids) == 0 {
			return fmt.Errorf("creator %s who is recording video is not esixted", *v.RecordingState.Creator)
		}
	}

	return nil
}

func (v *VirtualRoom) isValidAttendeeStates(ctx context.Context) error {
	if len(v.AttendeeStates) == 0 {
		return nil
	}

	if err := v.AttendeeStates.isValid(); err != nil {
		return fmt.Errorf("invalid attendee states: %w", err)
	}

	// check attendee's answer is in polling or not
	if v.CurrentPolling != nil {
		pollingAnswer := v.CurrentPolling.GetAnswerMap()
		for _, attendee := range v.AttendeeStates {
			for _, answer := range attendee.PollingAnswer.Answer {
				if _, ok := pollingAnswer[answer]; !ok {
					return fmt.Errorf("answer of anttendee %s is not exist", attendee.UserID)
				}
			}
		}
	}

	attendeeIDs := v.AttendeeStates.GetAttendeeIDs()
	res, err := v.userModulePort.CheckExistedUserIDs(ctx, attendeeIDs)
	if err != nil {
		return fmt.Errorf("AttendeeStates: userModulePort.IsExistedUserIDs: %w", err)
	}

	existedIDs := golibs.StringSliceToMap(res)
	for _, attendeeID := range attendeeIDs {
		if _, ok := existedIDs[attendeeID]; !ok {
			return fmt.Errorf("attendee ID %s is not existed", attendeeID)
		}
	}

	return nil
}

func (v *VirtualRoom) isValidPresentMaterialState() error {
	if v.PresentMaterialState == nil {
		return nil
	}

	if len(v.Materials) == 0 && v.PresentMaterialState != nil {
		return fmt.Errorf("material which is presenting is not exist")
	}

	if err := v.PresentMaterialState.IsValid(); err != nil {
		return fmt.Errorf("invalid present material state: %w", err)
	}

	// check material which is presenting is in list material or not
	isExist := false
	switch value := v.PresentMaterialState.(type) {
	case *VideoPresentMaterialState:
		materialID := value.Material.ID
		for i := 0; i < len(v.Materials); i++ {
			m := v.Materials.GetVideoMaterialElement(i)
			if m == nil {
				continue
			}
			if m.ID == materialID {
				isExist = true
			}
		}
	case *PDFPresentMaterialState:
		materialID := value.Material.ID
		for i := 0; i < len(v.Materials); i++ {
			m := v.Materials.GetPDFMaterialElement(i)
			if m == nil {
				continue
			}
			if m.ID == materialID {
				isExist = true
			}
		}
	default:
		return fmt.Errorf("not yet handle type %T", v)
	}
	if !isExist {
		return fmt.Errorf("material which is presenting is not exist")
	}

	return nil
}

func (v *VirtualRoom) AttendeeConsumeAStreamingSlot(ctx context.Context, attendeeID string) error {
	if err := v.StreamingProvider.streamingProviderPort.AttendeeConsumeAStreamingSlot(ctx, v, attendeeID); err != nil {
		return fmt.Errorf("streamingProviderPort.AttendeeConsumeAStreamingSlot: %w", err)
	}

	return nil
}

type UserModulePort interface {
	CheckExistedUserIDs(ctx context.Context, ids []string) (existed []string, err error)
}
