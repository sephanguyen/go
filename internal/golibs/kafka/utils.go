package kafka

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/segmentio/kafka-go/protocol"
	"go.opentelemetry.io/otel"
)

func GetTopicNameWithPrefix(topicName, prefix string) string {
	return prefix + topicName
}

func TraceCarrierFromContext(ctx context.Context) *B3TraceCarrierImpl {
	carrier := &B3TraceCarrierImpl{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier
}

func ContextWithTraceCarrier(ctx context.Context, carrier *B3TraceCarrierImpl) context.Context {
	newCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)
	return newCtx
}

func TraceCarrierAndClaimInfoFromMessageHeaders(headers []protocol.Header) (*B3TraceCarrierImpl, *interceptors.CustomClaims) {
	userID, resourcePath := "", ""
	carrier := &B3TraceCarrierImpl{}
	for _, header := range headers {
		switch header.Key {
		case kafkaUserIDHeaderName:
			userID = string(header.Value)
		case kafkaResourcePathHeaderName:
			resourcePath = string(header.Value)
		default:
			carrier.Set(header.Key, string(header.Value))
		}
	}

	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			UserID:       userID,
			ResourcePath: resourcePath,
		},
	}

	return carrier, claim
}

func MessageHeadersFromContext(ctx context.Context, isTracing bool) []protocol.Header {
	userInfo := golibs.UserInfoFromCtx(ctx)
	headers := []protocol.Header{
		{
			Key:   kafkaUserIDHeaderName,
			Value: []byte(userInfo.UserID),
		},
		{
			Key:   kafkaResourcePathHeaderName,
			Value: []byte(userInfo.ResourcePath),
		},
	}

	if isTracing {
		traceCarrier := TraceCarrierFromContext(ctx)
		traceHeaders := make([]protocol.Header, 0)
		for key, val := range traceCarrier.GetAllValues() {
			traceHeaders = append(traceHeaders, protocol.Header{
				Key:   key,
				Value: []byte(val),
			})
		}

		headers = append(headers, traceHeaders...)
	}

	return headers
}
