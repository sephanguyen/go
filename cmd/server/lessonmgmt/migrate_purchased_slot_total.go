package lessonmgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/configurations"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

var (
	org                    string
	studentSubscriptionIDs string
	limitNum               int
)

func init() {
	bootstrap.RegisterJob("migrate_purchased_slot_total", MigratePurchasedSlotTotal).
		StringVar(&org, "organizationID", "", "specific organization").
		StringVar(&studentSubscriptionIDs, "studentSubscriptionID", "", "specific studentSubscriptionID").
		IntVar(&limitNum, "limit", 1000, "limit records")
}

type StudentSub struct {
	StudentSubID     string
	StudentID        string
	CourseID         string
	StartAt          time.Time
	EndAt            time.Time
	LocationID       string
	StudentPackageID string
}

// Step 0 : get total record of lesson student subscription
// Step 1 : get number of loop total / expected limit = number loop
// Step 2 : loop
//
//	        2.1 list lesson student sub record by limit and offset
//	        2.2 Calculate number total of lesson for every student course
//					package_type = PACKAGE_TYPE_ONE_TIME => total = course_location_schedule.total_no_lessons
//					package_type = PACKAGE_TYPE_SCHEDULED => total = course_location_schedule.frequency * number of weeks
//					package_type = PACKAGE_TYPE_SLOT_BASED => total = student_course.course_slot
//					package_type = PACKAGE_TYPE_FREQUENCY => total = student_course.course_slot_per_week * number of weeks
//			   2.3 Update purchased_slot_total by student_subscription_id
func MigratePurchasedSlotTotal(ctx context.Context, cfg configurations.Config, rsc *bootstrap.Resources) error {
	sugaredLogger := rsc.Logger().Sugar()
	sugaredLogger.Infof("Running migrate purchased slot total on env: %s", cfg.Common.Environment)

	start := time.Now()
	lessonDB := rsc.DBWith("lessonmgmt")
	if strings.TrimSpace(org) == "" {
		sugaredLogger.Error("org cannot be empty")
		return fmt.Errorf("org cannot be empty")
	}

	ctx = auth.InjectFakeJwtToken(ctx, org)
	query := "select count(*) from lesson_student_subscriptions where deleted_at is null "

	var subscriptionID []string
	if strings.TrimSpace(studentSubscriptionIDs) != "" {
		subscriptionID = strings.Split(studentSubscriptionIDs, "; ")
	}

	args := []interface{}{}
	if len(subscriptionID) > 0 {
		query += "and student_subscription_id = ANY($1) "
		args = append(args, subscriptionID)
	}
	var total int
	err := lessonDB.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		return err
	}

	numOffset := total / limitNum

	if remaining := total % limitNum; remaining != 0 {
		numOffset++
	}
	offset := 0
	listStudentSub := []*StudentSub{}

	paging := "limit $1 offset $2"
	for i := 0; i < numOffset; i++ {
		queryByLimit := `select lss.student_subscription_id ,student_id ,start_at ,end_at,course_id,location_id ,subscription_id
										from lesson_student_subscriptions lss join lesson_student_subscription_access_path lssap
										on lss.student_subscription_id  = lssap.student_subscription_id where lss.deleted_at is null and lssap.deleted_at is null `
		arg := []interface{}{limit, offset}
		if len(subscriptionID) > 0 {
			queryByLimit += "and lss.student_subscription_id = ANY($3) "
			arg = append(arg, subscriptionID)
		}
		queryByLimit += paging

		rows, err := lessonDB.Query(ctx, queryByLimit, arg...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			studentSub := &StudentSub{}
			if err := rows.Scan(
				&studentSub.StudentSubID,
				&studentSub.StudentID,
				&studentSub.StartAt,
				&studentSub.EndAt,
				&studentSub.CourseID,
				&studentSub.LocationID,
				&studentSub.StudentPackageID,
			); err != nil {
				return err
			}
			listStudentSub = append(listStudentSub, studentSub)
		}

		mapTotal, err := CalculatePurchasedSlotTotal(ctx, lessonDB.DB, listStudentSub)

		if err != nil {
			return fmt.Errorf("CalculatePurchasedSlotTotal.Error %w", err)
		}
		if err = UpdatePurchasedSlotTotal(ctx, lessonDB.DB, mapTotal); err != nil {
			return fmt.Errorf("UpdatePurchasedSlotTotal.Error %w", err)
		}
		offset += limit
	}
	sugaredLogger.Infof("Done for executing the job,it's took %v", time.Since(start))
	return nil
}

type CourseLocationSchedule struct {
	CourseID       string
	LocationID     string
	Freq           pgtype.Int2
	TotalNoLessons pgtype.Int2
	PackageType    string
}

func CalculatePurchasedSlotTotal(ctx context.Context, db database.Ext, studentSub []*StudentSub) (map[string]int, error) {
	args := []interface{}{}
	courseIDWithLocationID := make([]string, 0, len(studentSub)) // will like ["($1, $2)", "($3, $4)", ...]
	count := 0
	for i := 0; i < len(studentSub); i++ {
		courseID := studentSub[i].CourseID
		locationID := studentSub[i].LocationID
		args = append(args, &courseID, &locationID)
		courseIDWithLocationID = append(courseIDWithLocationID, fmt.Sprintf("($%d, $%d)", count+1, count+2))
		count += 2
	}
	// placeHolderVar will like ($1, $2), ($3, $4), ($5, $6), ....
	placeHolderVar := strings.Join(courseIDWithLocationID, ", ")

	query := "select course_id ,location_id ,frequency ,total_no_lessons ,product_type_schedule from course_location_schedule where (course_id,location_id) IN (:PlaceHolderVar)"
	query = strings.ReplaceAll(query, ":PlaceHolderVar", placeHolderVar)
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	listCourseLocation := []*CourseLocationSchedule{}
	for rows.Next() {
		cls := &CourseLocationSchedule{}
		if err := rows.Scan(&cls.CourseID, &cls.LocationID, &cls.Freq, &cls.TotalNoLessons, &cls.PackageType); err != nil {
			return nil, err
		}
		listCourseLocation = append(listCourseLocation, cls)
	}
	defer rows.Close()

	// calculate total
	mapTotal := make(map[string]int, 0)
	for _, ss := range studentSub {
		cls := FindCourseLocationScheduleByID(listCourseLocation, ss.CourseID, ss.LocationID)
		if cls != nil {
			var purchasedSlotTotal, freq int
			switch cls.PackageType {
			case "PACKAGE_TYPE_SLOT_BASED":
				courseSlot, _, err := GetConfigSlotFromOrder(ctx, db, ss.StudentID, ss.CourseID, ss.LocationID, ss.StudentPackageID)
				if err != nil {
					return nil, err
				}
				purchasedSlotTotal = courseSlot
			case "PACKAGE_TYPE_FREQUENCY":
				_, courseSlotperWeek, err := GetConfigSlotFromOrder(ctx, db, ss.StudentID, ss.CourseID, ss.LocationID, ss.StudentPackageID)
				if err != nil {
					return nil, err
				}
				freq = courseSlotperWeek
			case "PACKAGE_TYPE_SCHEDULED":
				freq = int(cls.Freq.Int)
			case "PACKAGE_TYPE_ONE_TIME":
				purchasedSlotTotal = int(cls.TotalNoLessons.Int)
			}

			if cls.PackageType == "PACKAGE_TYPE_SCHEDULED" || cls.PackageType == "PACKAGE_TYPE_FREQUENCY" {
				args := []interface{}{freq, ss.StartAt, ss.EndAt, ss.CourseID, ss.LocationID, ss.StudentID}
				query := "select calculate_purchased_slot_total_v2($1::smallint,$2::date,$3::date,$4,$5,$6)"
				var total pgtype.Int2
				err := db.QueryRow(ctx, query, args...).Scan(&total)
				if err != nil {
					return nil, err
				}
				purchasedSlotTotal = int(total.Int)
			}
			mapTotal[ss.StudentSubID] = purchasedSlotTotal
		}
	}

	return mapTotal, nil
}

func GetConfigSlotFromOrder(ctx context.Context, db database.Ext, studentID, courseID, locationID, studentPackageID string) (int, int, error) {
	query := "select course_slot,course_slot_per_week from student_course where student_id = $1 and course_id = $2 and location_id = $3 and student_package_id = $4"
	var courseSlot, courseSlotPerWeek pgtype.Int4
	err := db.QueryRow(ctx, query, studentID, courseID, locationID, studentPackageID).Scan(&courseSlot, &courseSlotPerWeek)

	if err == pgx.ErrNoRows {
		return 0, 0, nil
	} else if err != nil {
		return 0, 0, err
	}
	return int(courseSlot.Int), int(courseSlotPerWeek.Int), nil
}

func FindCourseLocationScheduleByID(list []*CourseLocationSchedule, courseID, locationID string) *CourseLocationSchedule {
	for _, cls := range list {
		if cls.CourseID == courseID && cls.LocationID == locationID {
			return cls
		}
	}
	return nil
}

func UpdatePurchasedSlotTotal(ctx context.Context, db database.Ext, mapStudentSub map[string]int) error {
	fieldNames := []string{"student_subscription_id", "course_id", "student_id", "subscription_id", "start_at", "end_at", "created_at", "updated_at"}
	placeHolders := "$1, $2, $3, $4, $5, $6, $7, $8"
	b := &pgx.Batch{}
	for studentSubID, total := range mapStudentSub {
		args := []interface{}{studentSubID, " ", " ", " ", time.Now(), time.Now(), time.Now(), time.Now(), total}
		query := fmt.Sprintf(`INSERT INTO lesson_student_subscriptions (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT lesson_student_subscriptions_pkey
		DO UPDATE SET updated_at = now(),purchased_slot_total = $9 `,
			strings.Join(fieldNames, ","),
			placeHolders)
		b.Queue(query, args...)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()
	for i := 0; i < b.Len(); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("class is not inserted")
		}
	}
	return nil
}
