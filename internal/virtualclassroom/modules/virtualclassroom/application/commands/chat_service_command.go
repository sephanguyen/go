package commands

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/manabie-com/backend/internal/golibs/clients"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
)

type ChatServiceCommand struct {
	LessonmgmtDB database.Ext

	ConversationClient         clients.ConversationClientInterface
	LiveLessonConversationRepo infrastructure.LiveLessonConversationRepo
}

type SuccessfulPrivateConversation struct {
	ConversationID string
	ParticipantID  string
}

type FailedPrivateConversation struct {
	FailedParticipantID string
	ErrorMsg            string
}

func (c *ChatServiceCommand) GetConversationID(ctx context.Context, lessonID string, participants []string, convType domain.LiveLessonConversationType) (conversationID string, err error) {
	userID := interceptors.UserIDFromContext(ctx)
	participants = append(participants, userID)

	con := domain.NewLiveLessonConversation(lessonID, participants, convType)
	con.RemoveDuplicates()

	switch convType {
	case domain.LiveLessonConversationTypePublic:
		storedCon, err := c.LiveLessonConversationRepo.GetConversationByLessonIDAndConvType(ctx, c.LessonmgmtDB, con.LessonID, string(con.ConversationType))
		if err == domain.ErrNoConversationFound {
			createConvResp, err := c.ConversationClient.CreateConversation(ctx, clients.CreateRequestCreateConversation(con.ConversationName, con.ParticipantList))
			if err != nil {
				return conversationID, fmt.Errorf("error in ConversationModifierService.CreateConversation, name %s participants %s: %w",
					con.ConversationName,
					con.ParticipantList,
					err,
				)
			}
			conversationID = createConvResp.ConversationId
			con.AddConversationID(conversationID)

			if err := c.LiveLessonConversationRepo.UpsertConversation(ctx, c.LessonmgmtDB, con); err != nil {
				return conversationID, fmt.Errorf("error in LiveLessonConversationRepo.UpsertConversation, lesson %s participants %s conversation id %s: %w",
					con.LessonID,
					con.ParticipantList,
					con.ConversationID,
					err,
				)
			}

			break
		} else if err != nil {
			return conversationID, fmt.Errorf("error in LiveLessonConversationRepo.GetConversationByLessonIDAndConvType, lesson %s conv type %s: %w",
				con.LessonID,
				string(con.ConversationType),
				err,
			)
		}

		// the conversation is existing
		conversationID = storedCon.ConversationID
		con.AddConversationID(conversationID)

		// check if there are new participants; if there are, then update live lesson conversation
		newParticipants := sliceutils.Filter(con.ParticipantList,
			func(participant_id string) bool {
				return !sliceutils.Contains(storedCon.ParticipantList, participant_id)
			},
		)

		if len(newParticipants) > 0 {
			// TODO: add conversation members from chat server, temporarily do nothing
			// use newParticipants as the new members

			// so the existing participants would not be replaced by only the participants from the request
			con.UpdateParticipants(append(storedCon.ParticipantList, newParticipants...))

			if err := c.LiveLessonConversationRepo.UpsertConversation(ctx, c.LessonmgmtDB, con); err != nil {
				return conversationID, fmt.Errorf("error in LiveLessonConversationRepo.UpsertConversation, lesson %s participants %s conversation id %s: %w",
					con.LessonID,
					con.ParticipantList,
					con.ConversationID,
					err,
				)
			}
		}
	case domain.LiveLessonConversationTypePrivate:
		conversationID, err = c.createPrivateConversation(ctx, con)
		if err != nil {
			return conversationID, err
		}
	default:
		return conversationID, fmt.Errorf("conversation type is not supported: %s", convType)
	}

	return conversationID, nil
}

func (c *ChatServiceCommand) GetPrivateConversationIDs(ctx context.Context, lessonID string, participantIDs []string) (map[string]string, []string, error) {
	userID := interceptors.UserIDFromContext(ctx)
	participantIDs = sliceutils.RemoveDuplicates(participantIDs)

	cleanParticipantIDs := sliceutils.Remove(participantIDs, func(participantID string) bool {
		return participantID == userID
	})
	participantIDslength := len(cleanParticipantIDs)
	if participantIDslength == 0 {
		return nil, nil, fmt.Errorf("participant list should contain at least one user ID excluding the current user")
	}

	participantConvMap := make(map[string]string, participantIDslength)
	failedParticipantIDs := make([]string, 0, participantIDslength)
	chanSuccessConv := make(chan SuccessfulPrivateConversation, participantIDslength)
	chanFailedConv := make(chan FailedPrivateConversation, participantIDslength)
	var (
		err          error
		wg           sync.WaitGroup
		errorMsgTemp atomic.Value
	)
	errorMsgTemp.Store("")

	// worker pool is used to limit the number of goroutines
	numWorkers := 50
	if numWorkers > participantIDslength {
		numWorkers = participantIDslength
	}
	workerPool := make(chan struct{}, numWorkers)
	// populate the worker pool by the number of workers
	// to block the available of workers
	for i := 0; i < numWorkers; i++ {
		workerPool <- struct{}{}
	}

	for _, participantID := range cleanParticipantIDs {
		<-workerPool // free up a space in the worker pool
		wg.Add(1)
		go c.addBatchCreatePrivateConversation(ctx, lessonID, userID, participantID, &wg, chanSuccessConv, chanFailedConv, workerPool)
	}

	go func() {
		wg.Wait()
		close(chanSuccessConv)
		close(chanFailedConv)
	}()

	for i := 0; i < participantIDslength; i++ {
		select {
		case successConv, ok := <-chanSuccessConv:
			if ok {
				participantConvMap[successConv.ParticipantID] = successConv.ConversationID
			}
		case failedConv, ok := <-chanFailedConv:
			if ok {
				failedParticipantIDs = append(failedParticipantIDs, failedConv.FailedParticipantID)
				errorMsgTemp.Store(errorMsgTemp.Load().(string) + failedConv.ErrorMsg)
			}
		}
	}

	if errMsg := errorMsgTemp.Load().(string); errMsg != "" {
		err = errors.New(errMsg)
	}

	return participantConvMap, failedParticipantIDs, err
}

func (c *ChatServiceCommand) createPrivateConversation(ctx context.Context, con domain.LiveLessonConversation) (conversationID string, err error) {
	conversationID, err = c.LiveLessonConversationRepo.GetConversationIDByExactInfo(ctx, c.LessonmgmtDB, con.LessonID, con.ParticipantList, string(con.ConversationType))
	if err == domain.ErrNoConversationFound {
		createConvResp, err := c.ConversationClient.CreateConversation(ctx, clients.CreateRequestCreateConversation(con.ConversationName, con.ParticipantList))
		if err != nil {
			return conversationID, fmt.Errorf("error in ConversationModifierService.CreateConversation, name %s participants %s: %w",
				con.ConversationName,
				con.ParticipantList,
				err,
			)
		}
		conversationID = createConvResp.ConversationId
		con.AddConversationID(conversationID)

		if err := c.LiveLessonConversationRepo.UpsertConversation(ctx, c.LessonmgmtDB, con); err != nil {
			return conversationID, fmt.Errorf("error in LiveLessonConversationRepo.UpsertConversation, lesson %s participants %s conversation id %s: %w",
				con.LessonID,
				con.ParticipantList,
				con.ConversationID,
				err,
			)
		}
	} else if err != nil {
		return conversationID, fmt.Errorf("error in LiveLessonConversationRepo.GetConversationIDByExactInfo for private chat, lesson %s participants %s: %w",
			con.LessonID,
			con.ParticipantList,
			err,
		)
	}

	return conversationID, nil
}

func (c *ChatServiceCommand) addBatchCreatePrivateConversation(
	ctx context.Context,
	lessonID, currentUserID, participantID string,
	wg *sync.WaitGroup,
	chanSuccessConv chan SuccessfulPrivateConversation,
	chanFailedConv chan FailedPrivateConversation,
	workerPool chan struct{}) {
	defer func() {
		workerPool <- struct{}{}
		wg.Done()
	}()

	participants := []string{currentUserID, participantID}
	con := domain.NewLiveLessonConversation(lessonID, participants, domain.LiveLessonConversationTypePrivate)

	conversationID, err := c.createPrivateConversation(ctx, con)
	if err != nil {
		chanFailedConv <- FailedPrivateConversation{
			FailedParticipantID: participantID,
			ErrorMsg:            fmt.Sprintf("%s ; ", err.Error()),
		}
		return
	}

	chanSuccessConv <- SuccessfulPrivateConversation{
		ConversationID: conversationID,
		ParticipantID:  participantID,
	}
}
