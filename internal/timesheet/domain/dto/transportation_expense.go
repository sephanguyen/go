package dto

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
)

type TransportationExpenses struct {
	TransportExpenseID string
	TimesheetID        string
	TransportationType string
	TransportationFrom string
	TransportationTo   string
	CostAmount         int32
	RoundTrip          bool
	Remarks            string
	IsDeleted          bool
}

func (t *TransportationExpenses) ToEntity() *entity.TransportationExpense {
	transportExpenseE := &entity.TransportationExpense{
		TransportationExpenseID: database.Text(t.TransportExpenseID),
		TimesheetID:             database.Text(t.TimesheetID),
		TransportationType:      database.Text(t.TransportationType),
		TransportationFrom:      database.Text(t.TransportationFrom),
		TransportationTo:        database.Text(t.TransportationTo),
		CostAmount:              database.Int4(t.CostAmount),
		RoundTrip:               database.Bool(t.RoundTrip),
		Remarks:                 database.Text(t.Remarks),
		CreatedAt:               pgtype.Timestamptz{Status: pgtype.Null},
		UpdatedAt:               pgtype.Timestamptz{Status: pgtype.Null},
		DeletedAt:               pgtype.Timestamptz{Status: pgtype.Null},
	}

	if t.IsDeleted {
		transportExpenseE.DeletedAt = database.Timestamptz(time.Now())
	}
	return transportExpenseE
}

func (t *TransportationExpenses) Validate() error {
	switch {
	case t.TransportationType == "":
		return fmt.Errorf("transportation type must not be empty")
	case t.TransportationType == pb.TransportationType_TYPE_INVALID.String():
		return fmt.Errorf("transportation type is invalid")

	case t.TransportationFrom == "":
		return fmt.Errorf("transportation from must not be nil")
	case len([]rune(t.TransportationFrom)) > constant.KTransportExpensesFromToLimit:
		return fmt.Errorf("transportation from must limit to 100 characters")

	case t.TransportationTo == "":
		return fmt.Errorf("transportation to must not be nil")
	case len([]rune(t.TransportationTo)) > constant.KTransportExpensesFromToLimit:
		return fmt.Errorf("transportation to must limit to 100 characters")

	case t.CostAmount <= 0:
		return fmt.Errorf("transportation cost amount must be greater than 0")

	case len([]rune(t.Remarks)) > constant.KTransportExpensesRemarksLimit:
		return fmt.Errorf("transportation remarks must limit to 100 characters")
	}
	return nil
}

func NewTransportExpenseFromRPCRequest(timesheetID string, req *pb.TransportationExpensesRequest) *TransportationExpenses {
	return &TransportationExpenses{
		TransportExpenseID: req.GetTransportationExpenseId(),
		TimesheetID:        timesheetID,
		TransportationType: req.GetType().String(),
		TransportationFrom: req.GetFrom(),
		TransportationTo:   req.GetTo(),
		CostAmount:         req.GetAmount(),
		RoundTrip:          req.GetRoundTrip(),
		Remarks:            req.GetRemarks(),
	}
}

type ListTransportationExpenses []*TransportationExpenses

func (lt ListTransportationExpenses) ToEntities() []*entity.TransportationExpense {
	listTransportExpenses := make(entity.ListTransportationExpenses, 0, len(lt))

	for _, transportExpense := range lt {
		transportExpenseE := transportExpense.ToEntity()
		listTransportExpenses = append(listTransportExpenses, transportExpenseE)
	}
	return listTransportExpenses
}

func (lt ListTransportationExpenses) Validate() error {
	if len(lt) > constant.KListTransportExpensesLimit {
		return fmt.Errorf("list transportation expenses must be limit to 10 rows")
	}

	for _, transportExpense := range lt {
		err := transportExpense.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

func NewListTransportExpensesFromRPCRequest(timesheetID string, listTransportExpenseReq []*pb.TransportationExpensesRequest) ListTransportationExpenses {
	listTransportExpenses := make(ListTransportationExpenses, 0, len(listTransportExpenseReq))
	for _, transportExpensesReq := range listTransportExpenseReq {
		transportExpensesReq := NewTransportExpenseFromRPCRequest(timesheetID, transportExpensesReq)
		listTransportExpenses = append(listTransportExpenses, transportExpensesReq)
	}
	return listTransportExpenses
}

func NewListTransportExpensesFromEntities(listTransportExpensesEntity []*entity.TransportationExpense) ListTransportationExpenses {
	listTransportExpenses := make(ListTransportationExpenses, 0, len(listTransportExpensesEntity))
	for _, transportExpenseReq := range listTransportExpensesEntity {
		transportExpense := NewTransportExpensesFromEntity(transportExpenseReq)
		listTransportExpenses = append(listTransportExpenses, transportExpense)
	}
	return listTransportExpenses
}

func NewTransportExpensesFromEntity(transportExpenseEntity *entity.TransportationExpense) *TransportationExpenses {
	isDeleted := false
	if transportExpenseEntity.DeletedAt.Status != pgtype.Null {
		isDeleted = true
	}

	return &TransportationExpenses{
		TransportExpenseID: transportExpenseEntity.TransportationExpenseID.String,
		TimesheetID:        transportExpenseEntity.TimesheetID.String,
		TransportationType: transportExpenseEntity.TransportationType.String,
		TransportationFrom: transportExpenseEntity.TransportationFrom.String,
		TransportationTo:   transportExpenseEntity.TransportationTo.String,
		CostAmount:         transportExpenseEntity.CostAmount.Int,
		Remarks:            transportExpenseEntity.Remarks.String,

		IsDeleted: isDeleted,
	}
}
