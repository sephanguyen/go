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

type StaffTransportationExpenses struct {
	ID                 string
	StaffID            string
	LocationID         string
	TransportationType string
	TransportationFrom string
	TransportationTo   string
	CostAmount         int32
	RoundTrip          bool
	Remarks            string
	IsDeleted          bool
}

func NewStaffTransportExpenseFromRPCRequest(staffID string, req *pb.StaffTransportationExpenseRequest) *StaffTransportationExpenses {
	return &StaffTransportationExpenses{
		ID:                 req.GetId(),
		StaffID:            staffID,
		LocationID:         req.GetLocationId(),
		TransportationType: req.GetType().String(),
		TransportationFrom: req.GetFrom(),
		TransportationTo:   req.GetTo(),
		CostAmount:         req.GetCostAmount(),
		RoundTrip:          req.GetRoundTrip(),
		Remarks:            req.GetRemarks(),
	}
}

func (t *StaffTransportationExpenses) ValidateUpsertInfo() error {
	switch {
	case t.LocationID == "":
		return fmt.Errorf("location id must not be empty")

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

func (t *StaffTransportationExpenses) IsEqual(newTE *StaffTransportationExpenses) bool {
	return t.CostAmount == newTE.CostAmount &&
		t.Remarks == newTE.Remarks &&
		t.RoundTrip == newTE.RoundTrip &&
		t.TransportationFrom == newTE.TransportationFrom &&
		t.TransportationTo == newTE.TransportationTo &&
		t.TransportationType == newTE.TransportationType
}

func (t *StaffTransportationExpenses) ToEntity() *entity.StaffTransportationExpense {
	transportExpenseE := &entity.StaffTransportationExpense{
		ID:                 database.Text(t.ID),
		LocationID:         database.Text(t.LocationID),
		StaffID:            database.Text(t.StaffID),
		TransportationType: database.Text(t.TransportationType),
		TransportationFrom: database.Text(t.TransportationFrom),
		TransportationTo:   database.Text(t.TransportationTo),
		CostAmount:         database.Int4(t.CostAmount),
		RoundTrip:          database.Bool(t.RoundTrip),
		Remarks:            database.Text(t.Remarks),
		CreatedAt:          pgtype.Timestamptz{Status: pgtype.Null},
		UpdatedAt:          pgtype.Timestamptz{Status: pgtype.Null},
		DeletedAt:          pgtype.Timestamptz{Status: pgtype.Null},
	}

	if t.IsDeleted {
		transportExpenseE.DeletedAt = database.Timestamptz(time.Now())
	}

	return transportExpenseE
}

type ListStaffTransportationExpenses []*StaffTransportationExpenses

func NewListStaffTransportExpensesFromRPCRequest(staffID string, listStaffTransportExpenseReq []*pb.StaffTransportationExpenseRequest) ListStaffTransportationExpenses {
	listStaffTransportExpenses := make(ListStaffTransportationExpenses, 0, len(listStaffTransportExpenseReq))

	for _, staffTransportExpensesReq := range listStaffTransportExpenseReq {
		staffTransportExpenseDto := NewStaffTransportExpenseFromRPCRequest(staffID, staffTransportExpensesReq)
		listStaffTransportExpenses = append(listStaffTransportExpenses, staffTransportExpenseDto)
	}

	return listStaffTransportExpenses
}

func (ls ListStaffTransportationExpenses) ValidateUpsertInfo() error {
	if len(ls) > constant.KListStaffTransportExpensesLimit {
		return fmt.Errorf("list staff transportation expenses config must be limit to 10 rows")
	}

	for _, staffTransportExpense := range ls {
		err := staffTransportExpense.ValidateUpsertInfo()
		if err != nil {
			return err
		}
	}
	return nil
}

func (ls ListStaffTransportationExpenses) ToEntities() []*entity.StaffTransportationExpense {
	listStaffTransportExpenses := make(entity.ListStaffTransportationExpense, 0, len(ls))

	for _, transportExpense := range ls {
		transportExpenseE := transportExpense.ToEntity()
		listStaffTransportExpenses = append(listStaffTransportExpenses, transportExpenseE)
	}

	return listStaffTransportExpenses
}

func NewListStaffTransportExpensesFromEntities(listStaffTransportExpensesEntity []*entity.StaffTransportationExpense) ListStaffTransportationExpenses {
	listStaffTransportExpenses := make(ListStaffTransportationExpenses, 0, len(listStaffTransportExpensesEntity))
	for _, staffTransportExpenseReq := range listStaffTransportExpensesEntity {
		staffTransportExpense := NewStaffTransportExpensesFromEntity(staffTransportExpenseReq)
		listStaffTransportExpenses = append(listStaffTransportExpenses, staffTransportExpense)
	}
	return listStaffTransportExpenses
}

func NewStaffTransportExpensesFromEntity(transportExpenseEntity *entity.StaffTransportationExpense) *StaffTransportationExpenses {
	isDeleted := false
	if transportExpenseEntity.DeletedAt.Status != pgtype.Null {
		isDeleted = true
	}

	return &StaffTransportationExpenses{
		ID:                 transportExpenseEntity.ID.String,
		StaffID:            transportExpenseEntity.StaffID.String,
		LocationID:         transportExpenseEntity.LocationID.String,
		TransportationType: transportExpenseEntity.TransportationType.String,
		TransportationFrom: transportExpenseEntity.TransportationFrom.String,
		TransportationTo:   transportExpenseEntity.TransportationTo.String,
		CostAmount:         transportExpenseEntity.CostAmount.Int,
		Remarks:            transportExpenseEntity.Remarks.String,
		RoundTrip:          transportExpenseEntity.RoundTrip.Bool,

		IsDeleted: isDeleted,
	}
}
