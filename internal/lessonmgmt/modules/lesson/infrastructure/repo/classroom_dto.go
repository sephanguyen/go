package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type Classroom struct {
	ClassroomID  pgtype.Text
	Name         pgtype.Text
	LocationID   pgtype.Text
	Remarks      pgtype.Text
	IsArchived   pgtype.Bool
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	RoomArea     pgtype.Text
	SeatCapacity pgtype.Int4
}

func (c *Classroom) FieldMap() ([]string, []interface{}) {
	return []string{
			"classroom_id",
			"name",
			"location_id",
			"remarks",
			"is_archived",
			"created_at",
			"updated_at",
			"deleted_at",
			"room_area",
			"seat_capacity",
		}, []interface{}{
			&c.ClassroomID,
			&c.Name,
			&c.LocationID,
			&c.Remarks,
			&c.IsArchived,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
			&c.RoomArea,
			&c.SeatCapacity,
		}
}

func (c *Classroom) TableName() string {
	return "classroom"
}

func (c *Classroom) ToClassroomEntity() *domain.Classroom {
	classroom := domain.NewClassroom(c.ClassroomID.String).
		WithName(c.Name.String).
		WithLocationID(c.LocationID.String).
		WithRoomArea(c.RoomArea.String).
		WithSeatCapacity(int(c.SeatCapacity.Int)).
		WithRemark(c.Remarks.String).
		WithIsArchived(c.IsArchived.Bool)
	return classroom
}

type ClassroomToExport struct {
	LocationID    pgtype.Text
	LocationName  pgtype.Text
	ClassroomID   pgtype.Text
	ClassroomName pgtype.Text
	Remarks       pgtype.Text
	IsArchived    pgtype.Bool
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
	DeletedAt     pgtype.Timestamptz
	RoomArea      pgtype.Text
	SeatCapacity  pgtype.Int4
}

func (clr *ClassroomToExport) TableName() string {
	return "classroom"
}

func (clr *ClassroomToExport) FieldMap() ([]string, []interface{}) {
	return []string{
			"location_id",
			"location_name",
			"classroom_id",
			"classroom_name",
			"remarks",
			"is_archived",
			"created_at",
			"updated_at",
			"deleted_at",
			"room_area",
			"seat_capacity",
		}, []interface{}{
			&clr.LocationID,
			&clr.LocationName,
			&clr.ClassroomID,
			&clr.ClassroomName,
			&clr.Remarks,
			&clr.IsArchived,
			&clr.CreatedAt,
			&clr.UpdatedAt,
			&clr.DeletedAt,
			&clr.RoomArea,
			&clr.SeatCapacity,
		}
}

func NewClassrooomFromEntity(clr *domain.Classroom) (*Classroom, error) {
	dto := &Classroom{}
	database.AllNullEntity(dto)
	if err := multierr.Combine(
		dto.ClassroomID.Set(clr.ClassroomID),
		dto.Name.Set(clr.Name),
		dto.LocationID.Set(clr.LocationID),
		dto.Remarks.Set(clr.Remarks),
		dto.RoomArea.Set(clr.RoomArea),
		dto.SeatCapacity.Set(clr.SeatCapacity),
		dto.IsArchived.Set(clr.IsArchived),
		dto.CreatedAt.Set(clr.CreatedAt),
		dto.UpdatedAt.Set(clr.UpdatedAt),
		dto.DeletedAt.Set(clr.DeletedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from classroom entity to classroom dto: %w", err)
	}
	return dto, nil
}
