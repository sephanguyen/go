package domain

import (
	"context"
	"fmt"
	"time"
)

type VirtualClassroom struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time

	Room   *VirtualRoom
	Lesson *VirtualLesson
}

func (v *VirtualClassroom) IsValid(ctx context.Context) error {
	if len(v.ID) == 0 {
		return fmt.Errorf("ID cannot be empty")
	}

	if v.Room == nil {
		return fmt.Errorf("room cannot be empty")
	}
	if err := v.Room.isValid(ctx); err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}

	if v.Lesson == nil {
		return fmt.Errorf("lesson cannot be empty")
	}
	if err := v.Lesson.IsValid(ctx); err != nil {
		return fmt.Errorf("invalid lesson: %w", err)
	}

	if len(v.Room.AttendeeStates) != 0 {
		attendeeIDs := v.Room.AttendeeStates.GetAttendeeIDs()
		if err := v.Lesson.CheckLessonMemberIDs(ctx, attendeeIDs); err != nil {
			return fmt.Errorf("CheckLessonMemberIDs: %w", err)
		}
	}

	return nil
}

func (v *VirtualClassroom) AttendeeConsumeAStreamingSlot(ctx context.Context, attendeeID string) error {
	if err := v.Lesson.CheckLessonMemberIDs(ctx, []string{attendeeID}); err != nil {
		return fmt.Errorf("CheckLessonMemberIDs: %w", err)
	}

	if err := v.Room.AttendeeConsumeAStreamingSlot(ctx, attendeeID); err != nil {
		return fmt.Errorf("Room.AttendeeConsumeAStreamingSlot: %w", err)
	}

	return nil
}

type VirtualClassroomPort interface {
	IsExistent(ctx context.Context, id string) (bool, error)
}

type VirtualClassRoomBuilder struct {
	virtualClassRoom *VirtualClassroom
}

func NewVirtualClassRoomBuilder() *VirtualClassRoomBuilder {
	return &VirtualClassRoomBuilder{
		virtualClassRoom: &VirtualClassroom{},
	}
}

func (v *VirtualClassRoomBuilder) Build(ctx context.Context) (*VirtualClassroom, error) {
	if err := v.virtualClassRoom.IsValid(ctx); err != nil {
		return nil, fmt.Errorf("invalid virtual class room: %v", err)
	}
	return v.virtualClassRoom, nil
}

// BuildDraft will skip validate data
// only use to load VirtualClassRoom object from trusted data sources
func (v *VirtualClassRoomBuilder) BuildDraft() *VirtualClassroom {
	return v.virtualClassRoom
}

func (v *VirtualClassRoomBuilder) WithID(id string) *VirtualClassRoomBuilder {
	v.virtualClassRoom.ID = id
	return v
}

func (v *VirtualClassRoomBuilder) WithModificationTime(createdAt, updatedAt time.Time) *VirtualClassRoomBuilder {
	v.virtualClassRoom.CreatedAt = createdAt
	v.virtualClassRoom.UpdatedAt = updatedAt
	return v
}

func (v *VirtualClassRoomBuilder) WithLessonID(id string) *VirtualClassRoomBuilder {
	v.defaultLesson()
	v.virtualClassRoom.Lesson.LessonID = id
	return v
}

func (v *VirtualClassRoomBuilder) WithStreamingProvider(streamingProvider *StreamingProvider) *VirtualClassRoomBuilder {
	v.defaultRoom()
	v.virtualClassRoom.Room.StreamingProvider = streamingProvider
	return v
}

func (v *VirtualClassRoomBuilder) WithEndedAt(ended time.Time) *VirtualClassRoomBuilder {
	v.defaultRoom()
	v.virtualClassRoom.Room.EndedAt = &ended
	return v
}

func (v *VirtualClassRoomBuilder) WithMaterials(materials Materials) *VirtualClassRoomBuilder {
	v.defaultRoom()
	v.virtualClassRoom.Room.Materials = materials
	return v
}

func (v *VirtualClassRoomBuilder) WithPresentMaterialState(presentMaterialState presentMaterialState) *VirtualClassRoomBuilder {
	v.defaultRoom()
	v.virtualClassRoom.Room.PresentMaterialState = presentMaterialState
	return v
}

func (v *VirtualClassRoomBuilder) WithCurrentPolling(currentPolling *CurrentPolling) *VirtualClassRoomBuilder {
	v.defaultRoom()
	v.virtualClassRoom.Room.CurrentPolling = currentPolling
	return v
}

func (v *VirtualClassRoomBuilder) WithRecordingState(recordingState *RecordingState) *VirtualClassRoomBuilder {
	v.defaultRoom()
	v.virtualClassRoom.Room.RecordingState = recordingState
	return v
}

func (v *VirtualClassRoomBuilder) WithAttendeeStates(attendeeStates AttendeeStates) *VirtualClassRoomBuilder {
	v.defaultRoom()
	v.virtualClassRoom.Room.AttendeeStates = attendeeStates
	return v
}

func (v *VirtualClassRoomBuilder) WithUserModulePort(userModulePort UserModulePort) *VirtualClassRoomBuilder {
	v.defaultRoom()
	v.virtualClassRoom.Room.userModulePort = userModulePort
	return v
}

func (v *VirtualClassRoomBuilder) WithVirtualLessonPort(virtualLessonPort VirtualLessonPort) *VirtualClassRoomBuilder {
	v.defaultLesson()
	v.virtualClassRoom.Lesson.virtualLessonPort = virtualLessonPort
	return v
}

func (v *VirtualClassRoomBuilder) defaultRoom() {
	if v.virtualClassRoom.Room == nil {
		v.virtualClassRoom.Room = &VirtualRoom{}
	}
}

func (v *VirtualClassRoomBuilder) defaultLesson() {
	if v.virtualClassRoom.Lesson == nil {
		v.virtualClassRoom.Lesson = &VirtualLesson{}
	}
}
