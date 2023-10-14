package seqnumberservice

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"

	"go.uber.org/zap"
)

type ISequenceNumberService interface {
	GetPaymentSequenceNumberService() IPaymentSequenceNumberService
}

type SequenceNumberService struct {
	PaymentRepo interface {
		GetLatestPaymentSequenceNumber(ctx context.Context, db database.QueryExecer) (int32, error)
		PaymentSeqNumberLockAdvisory(ctx context.Context, db database.QueryExecer) (bool, error)
		PaymentSeqNumberUnLockAdvisory(ctx context.Context, db database.QueryExecer) error
	}
	Logger zap.SugaredLogger
}

func (s *SequenceNumberService) GetPaymentSequenceNumberService() IPaymentSequenceNumberService {
	return &PaymentSequenceNumberService{
		PaymentRepo: s.PaymentRepo,
		Logger:      s.Logger,
	}
}
