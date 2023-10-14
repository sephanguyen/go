package debezium

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"

	"go.uber.org/zap"
)

// Register service as source connector
type IncrementalSnapshotSubscription struct {
	Logger   *zap.Logger
	JSM      nats.JetStreamManagement
	DB       database.QueryExecer
	SourceID string
}

func InitDebeziumIncrementalSnapshot(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	db database.QueryExecer,
	sourceID string,
) error {
	internalDebeziumIncrementalSnapshot := &IncrementalSnapshotSubscription{
		JSM:      jsm,
		Logger:   logger,
		DB:       db,
		SourceID: sourceID,
	}
	err := internalDebeziumIncrementalSnapshot.Subscribe(sourceID)
	if err != nil {
		return fmt.Errorf("internalDebeziumIncrementalSnapshot.Subscribe: %w", err)
	}

	return nil
}

func (rcv *IncrementalSnapshotSubscription) Subscribe(sourceID string) error {
	var optsDebeziumIncremetalSnapshotSub nats.Option
	var queueName string

	switch sourceID {
	case "bob":
		queueName = constants.QueueBobDebeziumIncrementalSnapshotSend
		optsDebeziumIncremetalSnapshotSub = nats.Option{
			JetStreamOptions: []nats.JSSubOption{
				nats.ManualAck(),
				nats.Bind(constants.StreamDebeziumIncrementalSnapshot, constants.DurableBobDebeziumIncrementalSnapshotSend),
				nats.DeliverSubject(constants.DeliverDebeziumIncrementalSnapshotSend),
				nats.MaxDeliver(10),
				nats.AckWait(30 * time.Second),
			},
		}

	case "calendar":
		queueName = constants.QueueCalendarDebeziumIncrementalSnapshotSend
		optsDebeziumIncremetalSnapshotSub = nats.Option{
			JetStreamOptions: []nats.JSSubOption{
				nats.ManualAck(),
				nats.Bind(constants.StreamDebeziumIncrementalSnapshot, constants.DurableCalendarDebeziumIncrementalSnapshotSend),
				nats.DeliverSubject(constants.DeliverDebeziumIncrementalSnapshotSend),
				nats.MaxDeliver(10),
				nats.AckWait(30 * time.Second),
			},
		}

	case "fatima":
		queueName = constants.QueueFatimaDebeziumIncrementalSnapshotSend
		optsDebeziumIncremetalSnapshotSub = nats.Option{
			JetStreamOptions: []nats.JSSubOption{
				nats.ManualAck(),
				nats.Bind(constants.StreamDebeziumIncrementalSnapshot, constants.DurableFatimaDebeziumIncrementalSnapshotSend),
				nats.DeliverSubject(constants.DeliverDebeziumIncrementalSnapshotSend),
				nats.MaxDeliver(10),
				nats.AckWait(30 * time.Second),
			},
		}

	case "mastermgmt":
		queueName = constants.QueueMastermgmtDebeziumIncrementalSnapshotSend
		optsDebeziumIncremetalSnapshotSub = nats.Option{
			JetStreamOptions: []nats.JSSubOption{
				nats.ManualAck(),
				nats.Bind(constants.StreamDebeziumIncrementalSnapshot, constants.DurableMastermgmtDebeziumIncrementalSnapshotSend),
				nats.DeliverSubject(constants.DeliverDebeziumIncrementalSnapshotSend),
				nats.MaxDeliver(10),
				nats.AckWait(30 * time.Second),
			},
		}

	default:
		return fmt.Errorf("cannot set config for source service %s", sourceID)
	}

	_, err := rcv.JSM.QueueSubscribe(constants.SubjectDebeziumIncrementalSnapshotSend, queueName, optsDebeziumIncremetalSnapshotSub, rcv.HandlerNatsMessageDebeziumIncrementalSnapshot)
	if err != nil {
		return fmt.Errorf("subDebeziumIncrementalSnapshot.QueueSubscribe: %w", err)
	}

	return nil
}

func (rcv *IncrementalSnapshotSubscription) HandlerNatsMessageDebeziumIncrementalSnapshot(ctx context.Context, data []byte) (bool, error) {
	// TODO:
	// remove this workflow
	// ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	// defer cancel()

	// dataCollection := DataCollection{}

	// err := json.Unmarshal(data, &dataCollection)
	// if err != nil {
	// 	return false, err
	// }

	// if dataCollection.SourceID == rcv.SourceID {
	// 	err := try.Do(func(attempt int) (bool, error) {
	// 		fmt.Println("waiting for replication slot to active", dataCollection.RepName)
	// 		if isActive, err := isReplicationSlotActive(ctx, rcv.DB, dataCollection.RepName); err != nil || !isActive {
	// 			time.Sleep(300 * time.Millisecond)
	// 			return true, err
	// 		}
	// 		return false, nil
	// 	})

	// 	if err != nil {
	// 		return true, err
	// 	}

	// 	err = IncrementalSnapshot(ctx, rcv.DB, dataCollection)
	// 	if err != nil {
	// 		return false, err
	// 	}
	// }

	return false, nil
}
