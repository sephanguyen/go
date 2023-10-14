package seqnumberservice

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	"go.uber.org/zap"
)

var PaymentSeqNumberLockAcquiredErr = "payment sequence number lock is currently acquired by another process"

type IPaymentSequenceNumberService interface {
	InitLatestSeqNumber(ctx context.Context, db database.QueryExecer) error
	InitLatestSeqNumberWithLock(ctx context.Context, db database.QueryExecer) (releaseLockFunc func(), err error)
	AssignSeqNumberToPayment(payment *entities.Payment) error
	AssignSeqNumberToPayments(payments []*entities.Payment) error
}

type PaymentSequenceNumberService struct {
	latestPaymentSeqNumber int32
	PaymentRepo            interface {
		GetLatestPaymentSequenceNumber(ctx context.Context, db database.QueryExecer) (int32, error)
		PaymentSeqNumberLockAdvisory(ctx context.Context, db database.QueryExecer) (bool, error)
		PaymentSeqNumberUnLockAdvisory(ctx context.Context, db database.QueryExecer) error
	}
	Logger zap.SugaredLogger
}

func (s *PaymentSequenceNumberService) InitLatestSeqNumber(ctx context.Context, db database.QueryExecer) error {
	// Set the latest sequence number
	return s.setLatestSeqNumber(ctx, db)
}

func (s *PaymentSequenceNumberService) InitLatestSeqNumberWithLock(ctx context.Context, db database.QueryExecer) (releaseLockFunc func(), err error) {
	lockAcquired, err := s.PaymentRepo.PaymentSeqNumberLockAdvisory(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("paymentRepo.LockLatestPaymentSequenceNumber err: %v", err)
	}

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	if !lockAcquired {
		return nil, fmt.Errorf("%v in resource path %v", PaymentSeqNumberLockAcquiredErr, resourcePath)
	}

	// Set the latest sequence number
	err = s.setLatestSeqNumber(ctx, db)
	if err != nil {
		return nil, err
	}

	return func() {
		if err := s.PaymentRepo.PaymentSeqNumberUnLockAdvisory(ctx, db); err != nil {
			s.Logger.Warnf("%v unable to release lock", err.Error())
		}
	}, nil
}

func (s *PaymentSequenceNumberService) setLatestSeqNumber(ctx context.Context, db database.QueryExecer) error {
	latestPaymentSequenceNumber, err := s.PaymentRepo.GetLatestPaymentSequenceNumber(ctx, db)
	if err != nil {
		return fmt.Errorf("paymentRepo.GetLatestPaymentSequenceNumber err: %v", err)
	}

	s.latestPaymentSeqNumber = latestPaymentSequenceNumber
	return nil
}

func (s *PaymentSequenceNumberService) incrementSeqNumber() {
	s.latestPaymentSeqNumber++
}

func (s *PaymentSequenceNumberService) AssignSeqNumberToPayment(payment *entities.Payment) error {
	s.incrementSeqNumber()
	return payment.PaymentSequenceNumber.Set(s.latestPaymentSeqNumber)
}

func (s *PaymentSequenceNumberService) AssignSeqNumberToPayments(payments []*entities.Payment) error {
	for _, p := range payments {
		s.incrementSeqNumber()
		err := p.PaymentSequenceNumber.Set(s.latestPaymentSeqNumber)
		if err != nil {
			return err
		}
	}

	return nil
}
