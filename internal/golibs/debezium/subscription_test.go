package debezium

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerNatsMessageDebeziumIncrementalSnapshot(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name                string
		data                []byte
		subscriptionHandler *IncrementalSnapshotSubscription
		err                 error
		mockFunc            func(*IncrementalSnapshotSubscription, []byte)
	}{
		{
			name: "happy case receive message successfully",
			data: []byte(`{"SourceID":"bob","Tables":["public.student_entryexit_records","public.prefecture"],"CurrentTopic":null}`),
			subscriptionHandler: &IncrementalSnapshotSubscription{
				SourceID: "bob",
			},
			err: nil,
			mockFunc: func(subHandler *IncrementalSnapshotSubscription, data []byte) {
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFunc(tc.subscriptionHandler, tc.data)
			_, err := tc.subscriptionHandler.HandlerNatsMessageDebeziumIncrementalSnapshot(ctx, tc.data)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
