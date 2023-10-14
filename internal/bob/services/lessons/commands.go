package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type StateModifyCommand interface {
	getCommander() string
	getLessonID() string
	initBasicData(commanderID, lessonID string)
}

// list commands
var (
	_ StateModifyCommand = new(ShareMaterialCommand)
	_ StateModifyCommand = new(StopSharingMaterialCommand)
	_ StateModifyCommand = new(FoldHandAllCommand)
	_ StateModifyCommand = new(UpdateHandsUpCommand)
	_ StateModifyCommand = new(ResetAllStatesCommand)
	_ StateModifyCommand = new(UpdateAnnotationCommand)
	_ StateModifyCommand = new(DisableAllAnnotationCommand)
	_ StateModifyCommand = new(StartPollingCommand)
	_ StateModifyCommand = new(StopPollingCommand)
	_ StateModifyCommand = new(EndPollingCommand)
	_ StateModifyCommand = new(SubmitPollingAnswerCommand)
	_ StateModifyCommand = new(ResetPollingCommand)
	_ StateModifyCommand = new(RequestRecordingCommand)
	_ StateModifyCommand = new(StopRecordingCommand)
	_ StateModifyCommand = new(SpotlightCommand)
	_ StateModifyCommand = new(WhiteboardZoomStateCommand)
	_ StateModifyCommand = new(UpdateChatCommand)
	_ StateModifyCommand = new(ResetAllChatCommand)
)

type ShareMaterialCommand struct {
	CommanderID string
	LessonID    string
	State       *CurrentMaterial
}

type StopSharingMaterialCommand struct {
	CommanderID string
	LessonID    string
}

type FoldHandAllCommand struct {
	CommanderID string
	LessonID    string
}

type DisableAllAnnotationCommand struct {
	CommanderID string
	LessonID    string
}

type UpdateHandsUpCommand struct {
	CommanderID string
	UserID      string // user who will be changed hands up state
	LessonID    string
	State       *UserHandsUp
}

type ResetAllStatesCommand struct {
	CommanderID string
	LessonID    string
}

type UpdateAnnotationCommand struct {
	CommanderID string
	UserIDs     []string // list user who will be changed annotation state
	LessonID    string
	State       *UserAnnotation
}

type StartPollingCommand struct {
	CommanderID string
	Options     []*PollingOption
	LessonID    string
}

type StopPollingCommand struct {
	CommanderID string
	LessonID    string
}

type SubmitPollingAnswerCommand struct {
	CommanderID string
	LessonID    string
	UserID      string // user who will be submit answer
	Answers     []string
}

type EndPollingCommand struct {
	CommanderID string
	LessonID    string
}

type ResetPollingCommand struct {
	CommanderID string
	LessonID    string
}

type RequestRecordingCommand struct {
	CommanderID string
	LessonID    string
}

type StopRecordingCommand struct {
	CommanderID string
	LessonID    string
}

type SpotlightCommand struct {
	CommanderID     string
	LessonID        string
	SpotlightedUser string
	IsEnable        bool
}

type WhiteboardZoomStateCommand struct {
	CommanderID         string
	LessonID            string
	WhiteboardZoomState *domain.WhiteboardZoomState
}

type UpdateChatCommand struct {
	CommanderID string
	LessonID    string
	UserIDs     []string // list user whose chat permission will be changed
	State       *UserChat
}

type ResetAllChatCommand struct {
	CommanderID string
	LessonID    string
}

type PollingOptions []*PollingOption

func (c *ShareMaterialCommand) getCommander() string {
	return c.CommanderID
}

func (c *ShareMaterialCommand) getLessonID() string {
	return c.LessonID
}

func (c *ShareMaterialCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *SubmitPollingAnswerCommand) getCommander() string {
	return c.CommanderID
}

func (c *SubmitPollingAnswerCommand) getLessonID() string {
	return c.LessonID
}

func (c *SubmitPollingAnswerCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *StopSharingMaterialCommand) getCommander() string {
	return c.CommanderID
}

func (c *StopSharingMaterialCommand) getLessonID() string {
	return c.LessonID
}

func (c *StopSharingMaterialCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *FoldHandAllCommand) getCommander() string {
	return c.CommanderID
}

func (c *FoldHandAllCommand) getLessonID() string {
	return c.LessonID
}

func (c *FoldHandAllCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *UpdateHandsUpCommand) getCommander() string {
	return c.CommanderID
}

func (c *UpdateHandsUpCommand) getLessonID() string {
	return c.LessonID
}

func (c *UpdateHandsUpCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *ResetAllStatesCommand) getCommander() string {
	return c.CommanderID
}

func (c *ResetAllStatesCommand) getLessonID() string {
	return c.LessonID
}

func (c *ResetAllStatesCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *UpdateAnnotationCommand) getCommander() string {
	return c.CommanderID
}

func (c *UpdateAnnotationCommand) getLessonID() string {
	return c.LessonID
}

func (c *UpdateAnnotationCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *DisableAllAnnotationCommand) getCommander() string {
	return c.CommanderID
}

func (c *DisableAllAnnotationCommand) getLessonID() string {
	return c.LessonID
}

func (c *DisableAllAnnotationCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *StartPollingCommand) getCommander() string {
	return c.CommanderID
}

func (c *StartPollingCommand) getLessonID() string {
	return c.LessonID
}

func (c *StartPollingCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *StopPollingCommand) getCommander() string {
	return c.CommanderID
}

func (c *StopPollingCommand) getLessonID() string {
	return c.LessonID
}

func (c *StopPollingCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *EndPollingCommand) getCommander() string {
	return c.CommanderID
}

func (c *EndPollingCommand) getLessonID() string {
	return c.LessonID
}

func (c *EndPollingCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *ResetPollingCommand) getCommander() string {
	return c.CommanderID
}

func (c *ResetPollingCommand) getLessonID() string {
	return c.LessonID
}

func (c *ResetPollingCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *RequestRecordingCommand) getCommander() string {
	return c.CommanderID
}

func (c *RequestRecordingCommand) getLessonID() string {
	return c.LessonID
}

func (c *RequestRecordingCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *StopRecordingCommand) getCommander() string {
	return c.CommanderID
}

func (c *StopRecordingCommand) getLessonID() string {
	return c.LessonID
}

func (c *StopRecordingCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *SpotlightCommand) getCommander() string {
	return c.CommanderID
}

func (c *SpotlightCommand) getLessonID() string {
	return c.LessonID
}

func (c *SpotlightCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *WhiteboardZoomStateCommand) getCommander() string {
	return c.CommanderID
}

func (c *WhiteboardZoomStateCommand) getLessonID() string {
	return c.LessonID
}

func (c *WhiteboardZoomStateCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *UpdateChatCommand) getCommander() string {
	return c.CommanderID
}

func (c *UpdateChatCommand) getLessonID() string {
	return c.LessonID
}

func (c *UpdateChatCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

func (c *ResetAllChatCommand) getCommander() string {
	return c.CommanderID
}

func (c *ResetAllChatCommand) getLessonID() string {
	return c.LessonID
}

func (c *ResetAllChatCommand) initBasicData(commanderID, lessonID string) {
	c.CommanderID = commanderID
	c.LessonID = lessonID
}

// list command handler

type CommandHandler interface {
	Execute(ctx context.Context) error
}

var (
	_ CommandHandler = new(ShareMaterialCommandHandler)
	_ CommandHandler = new(StopSharingMaterialCommandHandler)
	_ CommandHandler = new(FoldHandAllCommandHandler)
	_ CommandHandler = new(UpdateHandsUpCommandHandler)
	_ CommandHandler = new(ResetAllStatesCommandHandler)
	_ CommandHandler = new(UpdateAnnotationCommandHandler)
	_ CommandHandler = new(DisableAllAnnotationCommandHandler)
	_ CommandHandler = new(StartPollingCommandHandler)
	_ CommandHandler = new(StopPollingCommandHandler)
	_ CommandHandler = new(EndPollingCommandHandler)
	_ CommandHandler = new(SubmitPollingAnswerCommandHandler)
	_ CommandHandler = new(ResetPollingCommandHandler)
	_ CommandHandler = new(SpotlightCommandHandler)
)

type ShareMaterialCommandHandler struct {
	command *ShareMaterialCommand

	DB         database.Ext
	LessonRepo interface {
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
		UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error
		GrantRecordingPermission(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, recordingState pgtype.JSONB) error
		StopRecording(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, creator pgtype.Text, recordingState pgtype.JSONB) error
	}
	LessonGroupRepo interface {
		Get(ctx context.Context, db database.QueryExecer, lessonGroupID, courseID pgtype.Text) (*entities.LessonGroup, error)
	}
	LessonMemberRepo interface {
		GetLessonMemberStatesWithParams(ctx context.Context, db database.QueryExecer, filter *repositories.MemberStatesFilter) (entities.LessonMemberStates, error)
		UpsertLessonMemberState(ctx context.Context, db database.QueryExecer, state *entities.LessonMemberState) error
		UpsertAllLessonMemberStateByStateType(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, state *entities.StateValue) error
		UpsertMultiLessonMemberStateByState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, userIds pgtype.TextArray, state *entities.StateValue) error
	}
	LessonRoomStateRepo interface {
		UpsertCurrentMaterialState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, currentMaterial pgtype.JSONB) error
	}
}

func (h *ShareMaterialCommandHandler) Execute(ctx context.Context) error {
	switch h.DB.(type) {
	case pgx.Tx:
		break
	default:
		if err := database.ExecInTx(ctx, h.DB, func(ctx context.Context, tx pgx.Tx) error {
			handler := &ShareMaterialCommandHandler{
				command:             h.command,
				DB:                  tx,
				LessonRepo:          h.LessonRepo,
				LessonMemberRepo:    h.LessonMemberRepo,
				LessonGroupRepo:     h.LessonGroupRepo,
				LessonRoomStateRepo: h.LessonRoomStateRepo,
			}
			if err := handler.Execute(ctx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		return nil
	}

	// get the latest data of lesson
	lesson, err := h.LessonRepo.FindByID(ctx, h.DB, database.Text(h.command.LessonID))
	if err != nil {
		return fmt.Errorf("error in LessonRepo.FindByID, lesson %s: %w", h.command.LessonID, err)
	}

	// check media belong to lesson group of lesson or not
	if h.command.State != nil {
		gr, err := h.LessonGroupRepo.Get(ctx, h.DB, lesson.LessonGroupID, lesson.CourseID)
		if err != nil {
			return fmt.Errorf("error in LessonGroupRepo.Get, lesson %s: %w", h.command.LessonID, err)
		}

		isValid := false
		for _, media := range gr.MediaIDs.Elements {
			if media.String == h.command.State.MediaID {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("media %s not belong to lesson %s", h.command.State.MediaID, h.command.LessonID)
		}
	}

	newCurrentMaterial := h.command.State
	if newCurrentMaterial != nil {
		newCurrentMaterial.UpdatedAt = time.Now()

		if err := newCurrentMaterial.IsValid(); err != nil {
			return fmt.Errorf("invalid current material state: %w", err)
		}
	}

	src := pgtype.JSONB{}
	if err := src.Set(database.JSONB(newCurrentMaterial)); err != nil {
		return fmt.Errorf("could not marshal current material to jsonb: %w", err)
	}

	if err := h.LessonRoomStateRepo.UpsertCurrentMaterialState(ctx, h.DB, lesson.LessonID, src); err != nil {
		return fmt.Errorf("error in LessonRoomStateRepo.UpsertCurrentMaterialState, lesson %s: %w", h.command.LessonID, err)
	}

	return nil
}

type StopSharingMaterialCommandHandler struct {
	command *StopSharingMaterialCommand

	DB         database.Ext
	LessonRepo interface {
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
		UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error
		GrantRecordingPermission(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, recordingState pgtype.JSONB) error
		StopRecording(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, creator pgtype.Text, recordingState pgtype.JSONB) error
	}
	LessonMemberRepo interface {
		GetLessonMemberStatesWithParams(ctx context.Context, db database.QueryExecer, filter *repositories.MemberStatesFilter) (entities.LessonMemberStates, error)
		UpsertLessonMemberState(ctx context.Context, db database.QueryExecer, state *entities.LessonMemberState) error
		UpsertAllLessonMemberStateByStateType(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, state *entities.StateValue) error
		UpsertMultiLessonMemberStateByState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, userIds pgtype.TextArray, state *entities.StateValue) error
	}
	LessonRoomStateRepo interface {
		Spotlight(ctx context.Context, db database.QueryExecer, lessonID, userID pgtype.Text) error
		UnSpotlight(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error
		UpsertWhiteboardZoomState(ctx context.Context, db database.QueryExecer, lessonID string, whiteboardZoomState *domain.WhiteboardZoomState) error
		UpsertCurrentMaterialState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, currentMaterial pgtype.JSONB) error
	}
}

func (h *StopSharingMaterialCommandHandler) Execute(ctx context.Context) error {
	command := &ShareMaterialCommand{
		CommanderID: h.command.CommanderID,
		LessonID:    h.command.LessonID,
	}

	commandDp := &CommandDispatcher{
		DB:                  h.DB,
		LessonRepo:          h.LessonRepo,
		LessonMemberRepo:    h.LessonMemberRepo,
		LessonRoomStateRepo: h.LessonRoomStateRepo,
	}
	if err := commandDp.Execute(ctx, command); err != nil {
		return err
	}

	return nil
}

type FoldHandAllCommandHandler struct {
	command *FoldHandAllCommand

	DB               database.Ext
	LessonMemberRepo interface {
		UpsertAllLessonMemberStateByStateType(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, state *entities.StateValue) error
		UpsertMultiLessonMemberStateByState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, userIds pgtype.TextArray, state *entities.StateValue) error
	}
}

func (h *FoldHandAllCommandHandler) Execute(ctx context.Context) error {
	if err := h.LessonMemberRepo.UpsertAllLessonMemberStateByStateType(
		ctx,
		h.DB,
		database.Text(h.command.LessonID),
		database.Text(string(LearnerStateTypeHandsUp)),
		&entities.StateValue{
			BoolValue:        database.Bool(false),
			StringArrayValue: database.TextArray([]string{}),
		},
	); err != nil {
		return fmt.Errorf("LessonMemberRepo.UpsertAllLessonMemberStateByStateType: %w", err)
	}
	return nil
}

type DisableAllAnnotationCommandHandler struct {
	command *DisableAllAnnotationCommand

	DB               database.Ext
	LessonMemberRepo interface {
		UpsertAllLessonMemberStateByStateType(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, state *entities.StateValue) error
	}
}

func (h *DisableAllAnnotationCommandHandler) Execute(ctx context.Context) error {
	if err := h.LessonMemberRepo.UpsertAllLessonMemberStateByStateType(
		ctx,
		h.DB,
		database.Text(h.command.LessonID),
		database.Text(string(LearnerStateTypeAnnotation)),
		&entities.StateValue{
			BoolValue:        database.Bool(false),
			StringArrayValue: database.TextArray([]string{}),
		},
	); err != nil {
		return fmt.Errorf("LessonMemberRepo.UpsertAllLessonMemberStateByStateType: %w", err)
	}
	return nil
}

type UpdateHandsUpCommandHandler struct {
	command *UpdateHandsUpCommand

	DB               database.Ext
	LessonMemberRepo interface {
		UpsertLessonMemberState(ctx context.Context, db database.QueryExecer, state *entities.LessonMemberState) error
	}
}

func (h *UpdateHandsUpCommandHandler) Execute(ctx context.Context) error {
	state := &entities.LessonMemberState{}
	database.AllNullEntity(state)

	now := time.Now()
	if err := multierr.Combine(
		state.LessonID.Set(h.command.LessonID),
		state.UserID.Set(h.command.UserID),
		state.StateType.Set(LearnerStateTypeHandsUp),
		state.CreatedAt.Set(now),
		state.UpdatedAt.Set(now),
		state.BoolValue.Set(h.command.State.Value),
		state.StringArrayValue.Set([]string{}),
	); err != nil {
		return err
	}

	if err := h.LessonMemberRepo.UpsertLessonMemberState(
		ctx,
		h.DB,
		state,
	); err != nil {
		return fmt.Errorf("LessonMemberRepo.UpsertLessonMemberState: %w", err)
	}

	return nil
}

type ResetAllStatesCommandHandler struct {
	command *ResetAllStatesCommand

	DB         database.Ext
	LessonRepo interface {
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
		UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error
		GrantRecordingPermission(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, recordingState pgtype.JSONB) error
		StopRecording(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, creator pgtype.Text, recordingState pgtype.JSONB) error
	}
	LessonMemberRepo interface {
		GetLessonMemberStatesWithParams(ctx context.Context, db database.QueryExecer, filter *repositories.MemberStatesFilter) (entities.LessonMemberStates, error)
		UpsertLessonMemberState(ctx context.Context, db database.QueryExecer, state *entities.LessonMemberState) error
		UpsertAllLessonMemberStateByStateType(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, state *entities.StateValue) error
		UpsertMultiLessonMemberStateByState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, userIds pgtype.TextArray, state *entities.StateValue) error
	}
	LessonRoomStateRepo interface {
		Spotlight(ctx context.Context, db database.QueryExecer, lessonID, userID pgtype.Text) error
		UnSpotlight(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error
		UpsertWhiteboardZoomState(ctx context.Context, db database.QueryExecer, lessonID string, whiteboardZoomState *domain.WhiteboardZoomState) error
		UpsertCurrentMaterialState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, currentMaterial pgtype.JSONB) error
	}
}

func (h *ResetAllStatesCommandHandler) Execute(ctx context.Context) error {
	switch h.DB.(type) {
	case pgx.Tx:
		break
	default:
		if err := database.ExecInTx(ctx, h.DB, func(ctx context.Context, tx pgx.Tx) error {
			handler := &ResetAllStatesCommandHandler{
				command:             h.command,
				DB:                  tx,
				LessonRepo:          h.LessonRepo,
				LessonMemberRepo:    h.LessonMemberRepo,
				LessonRoomStateRepo: h.LessonRoomStateRepo,
			}
			if err := handler.Execute(ctx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		return nil
	}

	var command StateModifyCommand
	command = &StopSharingMaterialCommand{
		CommanderID: h.command.CommanderID,
		LessonID:    h.command.LessonID,
	}

	commandDp := &CommandDispatcher{
		DB:                  h.DB,
		LessonRepo:          h.LessonRepo,
		LessonMemberRepo:    h.LessonMemberRepo,
		LessonRoomStateRepo: h.LessonRoomStateRepo,
	}
	if err := commandDp.Execute(ctx, command); err != nil {
		return err
	}

	command = &DisableAllAnnotationCommand{
		CommanderID: h.command.CommanderID,
		LessonID:    h.command.LessonID,
	}
	if err := commandDp.Execute(ctx, command); err != nil {
		return err
	}

	command = &FoldHandAllCommand{
		CommanderID: h.command.CommanderID,
		LessonID:    h.command.LessonID,
	}
	if err := commandDp.Execute(ctx, command); err != nil {
		return err
	}

	command = &ResetPollingCommand{
		CommanderID: h.command.CommanderID,
		LessonID:    h.command.LessonID,
	}
	if err := commandDp.Execute(ctx, command); err != nil {
		return err
	}

	command = &SpotlightCommand{
		CommanderID: h.command.CommanderID,
		LessonID:    h.command.LessonID,
		IsEnable:    false,
	}
	if err := commandDp.Execute(ctx, command); err != nil {
		return err
	}

	command = &WhiteboardZoomStateCommand{
		CommanderID:         h.command.CommanderID,
		LessonID:            h.command.LessonID,
		WhiteboardZoomState: new(domain.WhiteboardZoomState).SetDefault(),
	}
	if err := commandDp.Execute(ctx, command); err != nil {
		return err
	}

	command = &ResetAllChatCommand{
		CommanderID: h.command.CommanderID,
		LessonID:    h.command.LessonID,
	}
	if err := commandDp.Execute(ctx, command); err != nil {
		return err
	}

	// TODO: remove current recording and change to new recording state
	// stop recording state
	lesson, err := h.LessonRepo.FindByID(ctx, h.DB, database.Text(h.command.LessonID))
	if err != nil {
		return fmt.Errorf("LessonRepo.FindByID: %w", err)
	}
	state, err := NewLiveLessonState(lesson.LessonID, lesson.RoomState, nil)
	if err != nil {
		return err
	}
	state.RoomState.Recording = &RecordingState{
		IsRecording: false,
		Creator:     nil,
	}
	if err = state.RoomState.IsValid(); err != nil {
		return fmt.Errorf("invalid room state: %w", err)
	}

	src := pgtype.JSONB{}
	if err = src.Set(state.RoomState); err != nil {
		return fmt.Errorf("could not marshal room state to jsonb: %w", err)
	}

	if err = h.LessonRepo.UpdateLessonRoomState(ctx, h.DB, lesson.LessonID, src); err != nil {
		return fmt.Errorf("LessonRepo.UpdateLessonRoomState: %w", err)
	}

	return nil
}

type UpdateAnnotationCommandHandler struct {
	command *UpdateAnnotationCommand

	DB         database.Ext
	LessonRepo interface {
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
		UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error
	}
	LessonMemberRepo interface {
		UpsertMultiLessonMemberStateByState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, userIds pgtype.TextArray, state *entities.StateValue) error
	}
}

func (h *UpdateAnnotationCommandHandler) Execute(ctx context.Context) error {
	// get the latest data of lesson
	switch h.DB.(type) {
	case pgx.Tx:
		break
	default:
		if err := database.ExecInTx(ctx, h.DB, func(ctx context.Context, tx pgx.Tx) error {
			handler := &UpdateAnnotationCommandHandler{
				command:          h.command,
				DB:               tx,
				LessonRepo:       h.LessonRepo,
				LessonMemberRepo: h.LessonMemberRepo,
			}
			if err := handler.Execute(ctx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		return nil
	}

	if err := h.LessonMemberRepo.UpsertMultiLessonMemberStateByState(
		ctx,
		h.DB,
		database.Text(h.command.LessonID),
		database.Text(string(LearnerStateTypeAnnotation)),
		database.TextArray(h.command.UserIDs),
		&entities.StateValue{
			BoolValue:        database.Bool(h.command.State.Value),
			StringArrayValue: database.TextArray([]string{}),
		},
	); err != nil {
		return fmt.Errorf("LessonMemberRepo.UpsertMultiLessonMemberStateByState: %w", err)
	}
	return nil
}

type StartPollingCommandHandler struct {
	command *StartPollingCommand

	DB         database.Ext
	LessonRepo interface {
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
		UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error
	}
}

func (h *StartPollingCommandHandler) Execute(ctx context.Context) error {
	// get the latest data of lesson
	switch h.DB.(type) {
	case pgx.Tx:
		break
	default:
		if err := database.ExecInTx(ctx, h.DB, func(ctx context.Context, tx pgx.Tx) error {
			handler := &StartPollingCommandHandler{
				command:    h.command,
				DB:         tx,
				LessonRepo: h.LessonRepo,
			}
			if err := handler.Execute(ctx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		return nil
	}
	if len(h.command.Options) < 2 {
		return fmt.Errorf("Option must be larger than 1")
	}
	if len(h.command.Options) > 10 {
		return fmt.Errorf("Option can not be larger than 10")
	}
	var options PollingOptions = h.command.Options
	if err := options.ValidatePollingOptions([]string{}); err != nil {
		return err
	}

	lesson, err := h.LessonRepo.FindByID(ctx, h.DB, database.Text(h.command.LessonID))
	if err != nil {
		return fmt.Errorf("LessonRepo.FindByID: %w", err)
	}
	state, err := NewLiveLessonState(lesson.LessonID, lesson.RoomState, nil)
	if err != nil {
		return err
	}
	if state.RoomState.CurrentPolling != nil {
		return fmt.Errorf("The Polling already exists")
	}
	state.RoomState.CurrentPolling = &CurrentPolling{
		Options:   h.command.Options,
		Status:    PollingStateStarted,
		CreatedAt: time.Now(),
	}
	if err := state.RoomState.IsValid(); err != nil {
		return fmt.Errorf("invalid room state: %w", err)
	}

	src := pgtype.JSONB{}
	if err := src.Set(state.RoomState); err != nil {
		return fmt.Errorf("could not marshal room state to jsonb: %w", err)
	}

	if err := h.LessonRepo.UpdateLessonRoomState(ctx, h.DB, lesson.LessonID, src); err != nil {
		return fmt.Errorf("LessonRepo.UpdateLessonRoomState: %w", err)
	}

	return nil
}

func (pos PollingOptions) ValidatePollingOptions(answers []string) error {
	hasCorrect := false
	allOptions := make(map[string]bool)
	for _, o := range pos {
		allOptions[o.Answer] = true
		if o.IsCorrect {
			hasCorrect = true
		}
	}
	if !hasCorrect {
		return fmt.Errorf("At least 1 correct answer")
	}
	for _, answer := range answers {
		if _, ok := allOptions[answer]; !ok {
			return fmt.Errorf("The answer %s doesn't belong to options", answer)
		}
	}

	return nil
}

type StopPollingCommandHandler struct {
	command *StopPollingCommand

	DB         database.Ext
	LessonRepo interface {
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
		UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error
	}
}

func (h *StopPollingCommandHandler) Execute(ctx context.Context) error {
	switch h.DB.(type) {
	case pgx.Tx:
		break
	default:
		if err := database.ExecInTx(ctx, h.DB, func(ctx context.Context, tx pgx.Tx) error {
			handler := &StopPollingCommandHandler{
				command:    h.command,
				DB:         tx,
				LessonRepo: h.LessonRepo,
			}
			if err := handler.Execute(ctx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		return nil
	}
	lesson, err := h.LessonRepo.FindByID(ctx, h.DB, database.Text(h.command.LessonID))
	if err != nil {
		return fmt.Errorf("LessonRepo.FindByID: %w", err)
	}
	state, err := NewLiveLessonState(lesson.LessonID, lesson.RoomState, nil)
	if err != nil {
		return err
	}
	if state.RoomState.CurrentPolling == nil {
		return fmt.Errorf("The Polling not exists")
	}
	if state.RoomState.CurrentPolling.Status != PollingStateStarted {
		return fmt.Errorf("permission denied: Can't stop polling when polling not start")
	}

	// update room state
	state.RoomState.CurrentPolling.StoppedAt = time.Now()
	state.RoomState.CurrentPolling.Status = PollingStateStopped

	if err := state.RoomState.IsValid(); err != nil {
		return fmt.Errorf("invalid room state: %w", err)
	}

	src := pgtype.JSONB{}
	if err := src.Set(state.RoomState); err != nil {
		return fmt.Errorf("could not marshal room state to jsonb: %w", err)
	}

	if err := h.LessonRepo.UpdateLessonRoomState(ctx, h.DB, lesson.LessonID, src); err != nil {
		return fmt.Errorf("LessonRepo.UpdateLessonRoomState: %w", err)
	}
	return nil
}

type EndPollingCommandHandler struct {
	command *EndPollingCommand

	DB         database.Ext
	LessonRepo interface {
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
		UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error
	}
	LessonMemberRepo interface {
		GetLessonMemberStatesWithParams(ctx context.Context, db database.QueryExecer, filter *repositories.MemberStatesFilter) (entities.LessonMemberStates, error)
		UpsertAllLessonMemberStateByStateType(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, state *entities.StateValue) error
	}
	LessonPollingRepo interface {
		Create(ctx context.Context, db database.Ext, polling *entities.LessonPolling) (*entities.LessonPolling, error)
	}
}

func (h *EndPollingCommandHandler) Execute(ctx context.Context) error {
	switch h.DB.(type) {
	case pgx.Tx:
		break
	default:
		if err := database.ExecInTx(ctx, h.DB, func(ctx context.Context, tx pgx.Tx) error {
			handler := &EndPollingCommandHandler{
				command:           h.command,
				DB:                tx,
				LessonRepo:        h.LessonRepo,
				LessonMemberRepo:  h.LessonMemberRepo,
				LessonPollingRepo: h.LessonPollingRepo,
			}
			if err := handler.Execute(ctx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		return nil
	}
	lesson, err := h.LessonRepo.FindByID(ctx, h.DB, database.Text(h.command.LessonID))
	if err != nil {
		return fmt.Errorf("LessonRepo.FindByID: %w", err)
	}
	state, err := NewLiveLessonState(lesson.LessonID, lesson.RoomState, nil)
	if err != nil {
		return err
	}
	if state.RoomState.CurrentPolling == nil {
		return fmt.Errorf("The Polling not exists")
	}
	if state.RoomState.CurrentPolling.Status != PollingStateStopped {
		return fmt.Errorf("permission denied: Can't end polling when polling not stop")
	}

	// save polling
	var learnerStates entities.LessonMemberStates
	filter := repositories.MemberStatesFilter{}
	errFilter := multierr.Combine(
		filter.LessonID.Set(nil),
		filter.UserID.Set(nil),
		filter.StateType.Set(nil),
	)
	if errFilter != nil {
		return fmt.Errorf("could not filter to get lesson member states: %w", errFilter)
	}
	filter.LessonID.Set(lesson.LessonID)
	filter.StateType.Set(LearnerStateTypePollingAnswer)
	learnerStates, err = h.LessonMemberRepo.GetLessonMemberStatesWithParams(ctx, h.DB, &filter)
	e := &entities.LessonPolling{}
	database.AllNullEntity(e)
	srcOptions := pgtype.JSONB{}
	if err := srcOptions.Set(state.RoomState.CurrentPolling.Options); err != nil {
		return fmt.Errorf("could not marshal options to jsonb: %w", err)
	}
	srcAnswers := pgtype.JSONB{}
	if err := srcAnswers.Set(learnerStates); err != nil {
		return fmt.Errorf("could not marshal answers to jsonb: %w", err)
	}

	pollId := idutil.ULIDNow()
	if err := multierr.Combine(
		e.PollID.Set(pollId),
		e.LessonID.Set(lesson.LessonID),
		e.Options.Set(srcOptions),
		e.StudentsAnswers.Set(srcAnswers),
		e.CreatedAt.Set(state.RoomState.CurrentPolling.CreatedAt),
		e.StoppedAt.Set(state.RoomState.CurrentPolling.StoppedAt),
		e.UpdatedAt.Set(state.RoomState.CurrentPolling.StoppedAt),
		e.EndedAt.Set(time.Now()),
	); err != nil {
		return err
	}
	e, err = h.LessonPollingRepo.Create(ctx, h.DB, e)
	if err != nil {
		return fmt.Errorf("LessonPollingRepo.Create: %v", err)
	}

	// update room state
	state.RoomState.CurrentPolling = nil
	src := pgtype.JSONB{}
	if err := src.Set(state.RoomState); err != nil {
		return fmt.Errorf("could not marshal room state to jsonb: %w", err)
	}

	if err := h.LessonRepo.UpdateLessonRoomState(ctx, h.DB, lesson.LessonID, src); err != nil {
		return fmt.Errorf("LessonRepo.UpdateLessonRoomState: %w", err)
	}

	// update user state
	if err := h.LessonMemberRepo.UpsertAllLessonMemberStateByStateType(
		ctx,
		h.DB,
		database.Text(h.command.LessonID),
		database.Text(string(LearnerStateTypePollingAnswer)),
		&entities.StateValue{
			BoolValue:        database.Bool(false),
			StringArrayValue: database.TextArray([]string{}),
		},
	); err != nil {
		return fmt.Errorf("LessonMemberRepo.UpsertAllLessonMemberStateByStateType: %w", err)
	}

	return nil
}

type SubmitPollingAnswerCommandHandler struct {
	command *SubmitPollingAnswerCommand

	DB         database.Ext
	LessonRepo interface {
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
	}
	LessonMemberRepo interface {
		GetLessonMemberStatesWithParams(ctx context.Context, db database.QueryExecer, filter *repositories.MemberStatesFilter) (entities.LessonMemberStates, error)
		UpsertLessonMemberState(ctx context.Context, db database.QueryExecer, state *entities.LessonMemberState) error
	}
}

func (h *SubmitPollingAnswerCommandHandler) Execute(ctx context.Context) error {
	if len(h.command.Answers) == 0 {
		return fmt.Errorf("At least 1 answer")
	}
	lesson, err := h.LessonRepo.FindByID(ctx, h.DB, database.Text(h.command.LessonID))
	if err != nil {
		return fmt.Errorf("LessonRepo.FindByID: %w", err)
	}
	state, err := NewLiveLessonState(lesson.LessonID, lesson.RoomState, nil)
	if err != nil {
		return err
	}
	if state.RoomState.CurrentPolling == nil {
		return fmt.Errorf("The Polling not exists")
	}
	if state.RoomState.CurrentPolling.Status != PollingStateStarted {
		return fmt.Errorf("permission denied: Can't submit answer when polling not start")
	}
	var options PollingOptions = state.RoomState.CurrentPolling.Options
	if err := options.ValidatePollingOptions(h.command.Answers); err != nil {
		return err
	}

	var learnerStates entities.LessonMemberStates
	filter := repositories.MemberStatesFilter{}
	errFilter := multierr.Combine(
		filter.LessonID.Set(nil),
		filter.UserID.Set(nil),
		filter.StateType.Set(nil),
	)
	if errFilter != nil {
		return errFilter
	}
	filter.LessonID.Set(lesson.LessonID)
	filter.UserID.Set(h.command.UserID)
	filter.StateType.Set(LearnerStateTypePollingAnswer)

	learnerStates, err = h.LessonMemberRepo.GetLessonMemberStatesWithParams(ctx, h.DB, &filter)
	if len(learnerStates) > 0 && len(learnerStates[0].StringArrayValue.Elements) > 0 {
		return fmt.Errorf("permission denied: Only submit 1 time")
	}
	memberState := &entities.LessonMemberState{}
	database.AllNullEntity(memberState)

	now := time.Now()
	if err := multierr.Combine(
		memberState.LessonID.Set(h.command.LessonID),
		memberState.UserID.Set(h.command.UserID),
		memberState.StateType.Set(LearnerStateTypePollingAnswer),
		memberState.CreatedAt.Set(now),
		memberState.UpdatedAt.Set(now),
		memberState.BoolValue.Set(false),
		memberState.StringArrayValue.Set(h.command.Answers),
	); err != nil {
		return err
	}

	if err := h.LessonMemberRepo.UpsertLessonMemberState(
		ctx,
		h.DB,
		memberState,
	); err != nil {
		return fmt.Errorf("LessonMemberRepo.UpsertLessonMemberState: %w", err)
	}

	return nil
}

type ResetPollingCommandHandler struct {
	command *ResetPollingCommand

	DB         database.Ext
	LessonRepo interface {
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
		UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error
	}
	LessonMemberRepo interface {
		UpsertAllLessonMemberStateByStateType(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, state *entities.StateValue) error
	}
}

func (h *ResetPollingCommandHandler) Execute(ctx context.Context) error {
	switch h.DB.(type) {
	case pgx.Tx:
		break
	default:
		if err := database.ExecInTx(ctx, h.DB, func(ctx context.Context, tx pgx.Tx) error {
			handler := &ResetPollingCommandHandler{
				command:          h.command,
				DB:               tx,
				LessonRepo:       h.LessonRepo,
				LessonMemberRepo: h.LessonMemberRepo,
			}
			if err := handler.Execute(ctx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		return nil
	}

	// get the latest data of lesson
	lesson, err := h.LessonRepo.FindByID(ctx, h.DB, database.Text(h.command.LessonID))
	if err != nil {
		return fmt.Errorf("LessonRepo.FindByID: %w", err)
	}
	state, err := NewLiveLessonState(lesson.LessonID, lesson.RoomState, nil)
	if err != nil {
		return err
	}
	if state.RoomState.CurrentPolling != nil {
		state.RoomState.CurrentPolling = nil
		if err := state.RoomState.IsValid(); err != nil {
			return fmt.Errorf("invalid room state: %w", err)
		}

		src := pgtype.JSONB{}
		if err := src.Set(state.RoomState); err != nil {
			return fmt.Errorf("could not marshal room state to jsonb: %w", err)
		}

		if err := h.LessonRepo.UpdateLessonRoomState(ctx, h.DB, lesson.LessonID, src); err != nil {
			return fmt.Errorf("LessonRepo.UpdateLessonRoomState: %w", err)
		}
		if err := h.LessonMemberRepo.UpsertAllLessonMemberStateByStateType(
			ctx,
			h.DB,
			database.Text(h.command.LessonID),
			database.Text(string(LearnerStateTypePollingAnswer)),
			&entities.StateValue{
				BoolValue:        database.Bool(false),
				StringArrayValue: database.TextArray([]string{}),
			},
		); err != nil {
			return fmt.Errorf("LessonMemberRepo.UpsertAllLessonMemberStateByStateType: %w", err)
		}
	}
	return nil
}

type RequestRecordingHandler struct {
	command *RequestRecordingCommand
	DB      database.Ext

	LessonRepo interface {
		GrantRecordingPermission(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, recordingState pgtype.JSONB) error
	}
}

func (h *RequestRecordingHandler) Execute(ctx context.Context) error {
	switch h.DB.(type) {
	case pgx.Tx:
		break
	default:
		if err := database.ExecInTx(ctx, h.DB, func(ctx context.Context, tx pgx.Tx) error {
			handler := &RequestRecordingHandler{
				command:    h.command,
				DB:         tx,
				LessonRepo: h.LessonRepo,
			}
			if err := handler.Execute(ctx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		return nil
	}

	state := &LessonRoomState{
		Recording: &RecordingState{
			IsRecording: true,
			Creator:     &h.command.CommanderID,
		},
	}
	src := pgtype.JSONB{}
	if err := src.Set(state); err != nil {
		return fmt.Errorf("could not marshal recording state to jsonb: %w", err)
	}
	if err := h.LessonRepo.GrantRecordingPermission(ctx, h.DB, database.Text(h.command.LessonID), src); err != nil {
		return fmt.Errorf("LessonRepo.GrantRecordingPermission: %w", err)
	}

	return nil
}

type StopRecordingHandler struct {
	command *StopRecordingCommand
	DB      database.Ext

	LessonRepo interface {
		StopRecording(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, creator pgtype.Text, recordingState pgtype.JSONB) error
	}
}

func (h *StopRecordingHandler) Execute(ctx context.Context) error {
	switch h.DB.(type) {
	case pgx.Tx:
		break
	default:
		if err := database.ExecInTx(ctx, h.DB, func(ctx context.Context, tx pgx.Tx) error {
			handler := &StopRecordingHandler{
				command:    h.command,
				DB:         tx,
				LessonRepo: h.LessonRepo,
			}
			if err := handler.Execute(ctx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		return nil
	}

	state := &LessonRoomState{
		Recording: &RecordingState{
			IsRecording: false,
		},
	}
	src := pgtype.JSONB{}
	if err := src.Set(state); err != nil {
		return fmt.Errorf("could not marshal recording state to jsonb: %w", err)
	}
	if err := h.LessonRepo.StopRecording(ctx, h.DB, database.Text(h.command.LessonID), database.Text(h.command.CommanderID), src); err != nil {
		return fmt.Errorf("LessonRepo.StopRecording: %w", err)
	}

	return nil
}

type SpotlightCommandHandler struct {
	command             *SpotlightCommand
	DB                  database.Ext
	LessonRoomStateRepo interface {
		Spotlight(ctx context.Context, db database.QueryExecer, lessonID, userID pgtype.Text) error
		UnSpotlight(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) error
	}
}

func (h *SpotlightCommandHandler) Execute(ctx context.Context) error {
	switch h.DB.(type) {
	case pgx.Tx:
		break
	default:
		if err := database.ExecInTx(ctx, h.DB, func(ctx context.Context, tx pgx.Tx) error {
			handler := &SpotlightCommandHandler{
				command:             h.command,
				DB:                  tx,
				LessonRoomStateRepo: h.LessonRoomStateRepo,
			}
			if err := handler.Execute(ctx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}
	var err error
	if h.command.IsEnable {
		err = h.LessonRoomStateRepo.Spotlight(ctx, h.DB, database.Text(h.command.LessonID), database.Text(h.command.SpotlightedUser))
	} else {
		err = h.LessonRoomStateRepo.UnSpotlight(ctx, h.DB, database.Text(h.command.LessonID))
	}
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

type WhiteboardZoomStateCommandHandler struct {
	command *WhiteboardZoomStateCommand

	DB                  database.Ext
	LessonRoomStateRepo interface {
		UpsertWhiteboardZoomState(ctx context.Context, db database.QueryExecer, lessonID string, whiteboardZoomState *domain.WhiteboardZoomState) error
	}
}

func (w *WhiteboardZoomStateCommandHandler) Execute(ctx context.Context) error {
	switch w.DB.(type) {
	case pgx.Tx:
		break
	default:
		if err := database.ExecInTx(ctx, w.DB, func(ctx context.Context, tx pgx.Tx) error {
			handler := &WhiteboardZoomStateCommandHandler{
				command:             w.command,
				DB:                  tx,
				LessonRoomStateRepo: w.LessonRoomStateRepo,
			}
			if err := handler.Execute(ctx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	return w.LessonRoomStateRepo.UpsertWhiteboardZoomState(ctx, w.DB, w.command.getLessonID(), w.command.WhiteboardZoomState)
}

type UpdateChatCommandHandler struct {
	command *UpdateChatCommand

	DB         database.Ext
	LessonRepo interface {
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
		UpdateLessonRoomState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, state pgtype.JSONB) error
	}
	LessonMemberRepo interface {
		UpsertMultiLessonMemberStateByState(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, userIds pgtype.TextArray, state *entities.StateValue) error
	}
}

func (h *UpdateChatCommandHandler) Execute(ctx context.Context) error {
	switch h.DB.(type) {
	case pgx.Tx:
		break
	default:
		if err := database.ExecInTx(ctx, h.DB, func(ctx context.Context, tx pgx.Tx) error {
			handler := &UpdateChatCommandHandler{
				command:          h.command,
				DB:               tx,
				LessonRepo:       h.LessonRepo,
				LessonMemberRepo: h.LessonMemberRepo,
			}
			if err := handler.Execute(ctx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		return nil
	}

	if err := h.LessonMemberRepo.UpsertMultiLessonMemberStateByState(
		ctx,
		h.DB,
		database.Text(h.command.LessonID),
		database.Text(string(LearnerStateTypeChat)),
		database.TextArray(h.command.UserIDs),
		&entities.StateValue{
			BoolValue:        database.Bool(h.command.State.Value),
			StringArrayValue: database.TextArray([]string{}),
		},
	); err != nil {
		return fmt.Errorf("LessonMemberRepo.UpsertMultiLessonMemberStateByState: %w", err)
	}
	return nil
}

type ResetAllChatCommandHandler struct {
	command *ResetAllChatCommand

	DB               database.Ext
	LessonMemberRepo interface {
		UpsertAllLessonMemberStateByStateType(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text, stateType pgtype.Text, state *entities.StateValue) error
	}
}

func (h *ResetAllChatCommandHandler) Execute(ctx context.Context) error {
	switch h.DB.(type) {
	case pgx.Tx:
		break
	default:
		if err := database.ExecInTx(ctx, h.DB, func(ctx context.Context, tx pgx.Tx) error {
			handler := &ResetAllChatCommandHandler{
				command:          h.command,
				DB:               tx,
				LessonMemberRepo: h.LessonMemberRepo,
			}
			if err := handler.Execute(ctx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		return nil
	}

	if err := h.LessonMemberRepo.UpsertAllLessonMemberStateByStateType(
		ctx,
		h.DB,
		database.Text(h.command.LessonID),
		database.Text(string(LearnerStateTypeChat)),
		&entities.StateValue{
			BoolValue:        database.Bool(true),
			StringArrayValue: database.TextArray([]string{}),
		},
	); err != nil {
		return fmt.Errorf("LessonMemberRepo.UpsertAllLessonMemberStateByStateType: %w", err)
	}
	return nil
}
