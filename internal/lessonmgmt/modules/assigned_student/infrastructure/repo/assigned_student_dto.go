package repo

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/application/queries/payloads"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/domain"

	"github.com/jackc/pgtype"
)

type AsgStudents []*AssignedStudentDTO

const (
	UnderAssignedStatus = "Under assigned"
	JustAssignedStatus  = "Just assigned"
	OverAssignedStatus  = "Over assigned"
)

func (a *AsgStudents) Add() database.Entity {
	e := &AssignedStudentDTO{}
	*a = append(*a, e)
	return e
}

type AssignedStudentDTO struct {
	PurchaseType          pgtype.Text
	StudentID             pgtype.Text
	CourseID              pgtype.Text
	LocationID            pgtype.Text
	StartDate             pgtype.Date
	EndDate               pgtype.Date
	Duration              pgtype.Text
	PurchaseSlot          pgtype.Int4
	AssignedSlot          pgtype.Int4
	SlotGap               pgtype.Int4
	Status                pgtype.Text
	StudentSubscriptionID pgtype.Text
}

func (a *AssignedStudentDTO) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id",
			"course_id",
			"location_id",
			"start_date",
			"end_date",
			"duration",
			"purchased_slot",
			"assigned_slot",
			"slot_gap",
			"status",
			"student_subscription_id",
		}, []interface{}{
			&a.StudentID,
			&a.CourseID,
			&a.LocationID,
			&a.StartDate,
			&a.EndDate,
			&a.Duration,
			&a.PurchaseSlot,
			&a.AssignedSlot,
			&a.SlotGap,
			&a.Status,
			&a.StudentSubscriptionID,
		}
}

func (a *AssignedStudentDTO) TableName() string {
	if a.PurchaseType.String == string(domain.PurchaseMethodRecurring) {
		return "student_course_recurring_slot_info"
	}
	return "student_course_slot_info"
}

type ListAsgStudentArgs struct {
	Limit                     uint32
	StudentSubscriptionID     pgtype.Text
	Courses                   pgtype.TextArray
	Students                  pgtype.TextArray
	FromDate                  pgtype.Date
	ToDate                    pgtype.Date
	KeyWord                   pgtype.Text
	LocationIDs               pgtype.TextArray
	AssignedStudentStatus     pgtype.TextArray
	Timezone                  pgtype.Text
}

func ToListAsgStudentArgsDto(ap *payloads.GetAssignedStudentListArg) *ListAsgStudentArgs {
	args := &ListAsgStudentArgs{
		Limit:                     ap.Limit,
		StudentSubscriptionID:     pgtype.Text{Status: pgtype.Null},
		Courses:                   pgtype.TextArray{Status: pgtype.Null},
		Students:                  pgtype.TextArray{Status: pgtype.Null},
		FromDate:                  pgtype.Date{Status: pgtype.Null},
		ToDate:                    pgtype.Date{Status: pgtype.Null},
		KeyWord:                   pgtype.Text{Status: pgtype.Null},
		LocationIDs:               pgtype.TextArray{Status: pgtype.Null},
		AssignedStudentStatus:     pgtype.TextArray{Status: pgtype.Null},
		Timezone:                  pgtype.Text{Status: pgtype.Null},
	}

	if len(ap.CourseIDs) > 0 {
		args.Courses = database.TextArray(ap.CourseIDs)
	}

	if len(ap.StudentIDs) > 0 {
		args.Students = database.TextArray(ap.StudentIDs)
	}

	if !ap.FromDate.IsZero() {
		args.FromDate = pgtype.Date(database.Timestamptz(ap.FromDate))
	}

	if !ap.ToDate.IsZero() {
		args.ToDate = pgtype.Date(database.Timestamptz(ap.ToDate))
	}

	if ap.KeyWord != "" {
		args.KeyWord = database.Text(ap.KeyWord)
	}

	if len(ap.LocationIDs) > 0 {
		args.LocationIDs = database.TextArray(ap.LocationIDs)
	}
	if ap.StudentSubscriptionID != "" {
		args.StudentSubscriptionID = database.Text(ap.StudentSubscriptionID)
	}

	if len(ap.AssignedStudentStatuses) > 0 {
		status := make([]string, 0, len(ap.AssignedStudentStatuses))
		for _, v := range ap.AssignedStudentStatuses {
			switch v {
			case domain.AssignedStudentStatusUnderAssigned:
				v = UnderAssignedStatus
			case domain.AssignedStudentStatusJustAssigned:
				v = JustAssignedStatus
			case domain.AssignedStudentStatusOverAssigned:
				v = OverAssignedStatus
			default:
				v = UnderAssignedStatus
			}
			status = append(status, string(v))
		}

		args.AssignedStudentStatus = database.TextArray(status)
	}

	if ap.Timezone == "" {
		args.Timezone = database.Text("UTC")
	} else {
		args.Timezone = database.Text(ap.Timezone)
	}

	return args
}
