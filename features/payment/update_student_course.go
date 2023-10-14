package payment

import (
	"context"
	"fmt"
	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"
)

func (s *suite) cronUpdateStudentCourseRun(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.CreateOrderRequest)
	resourcePath, err := interceptors.ResourcePathFromContext(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	newRequest := &pb.UpdateStudentCourseRequest{
		To:             timestamppb.New(req.OrderItems[0].EffectiveDate.AsTime()),
		OrganizationId: resourcePath,
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewInternalServiceClient(s.PaymentConn).UpdateStudentCourse(contextWithToken(ctx), newRequest)
	for stepState.ResponseErr != nil && strings.Contains(stepState.ResponseErr.Error(), "(SQLSTATE 23505)") {
		time.Sleep(5000)
		stepState.Response, stepState.ResponseErr = pb.NewInternalServiceClient(s.PaymentConn).GenerateBillingItems(contextWithToken(ctx), stepState.Request.(*pb.GenerateBillingItemsRequest))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getListStudentCourseBaseOnStudentPackageByOrderID(ctx context.Context, StudentPackageByOrderID string) ([]*entities.StudentCourse, error) {
	var studentCourses []*entities.StudentCourse
	studentCourse := &entities.StudentCourse{}
	studentCourseFieldNames, _ := studentCourse.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_package_id = $1
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentCourseFieldNames, ","),
		studentCourse.TableName(),
	)
	rows, err := s.FatimaDBTrace.Query(ctx, stmt, StudentPackageByOrderID)
	if err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}
	for rows.Next() {
		tmpStudentCourse := &entities.StudentCourse{}
		_, tmpStudentCourseFieldValues := tmpStudentCourse.FieldMap()
		err = rows.Scan(tmpStudentCourseFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		studentCourses = append(studentCourses, tmpStudentCourse)
	}
	return studentCourses, nil
}

func (s *suite) convertStudentCourseToUpcomingStudentCourse(studentCourses []*entities.StudentCourse, upcomingStudentPackage *entities.UpcomingStudentPackage) (upcomingStudentCourses []*entities.UpcomingStudentCourse) {
	upcomingStudentCourses = make([]*entities.UpcomingStudentCourse, 0, len(studentCourses))
	for _, studentCourse := range studentCourses {
		tmpUpcomingStudentCourse := &entities.UpcomingStudentCourse{}
		_ = multierr.Combine(
			tmpUpcomingStudentCourse.UpcomingStudentPackageID.Set(upcomingStudentPackage.UpcomingStudentPackageID.String),
			tmpUpcomingStudentCourse.CourseID.Set(studentCourse.CourseID.String),
			tmpUpcomingStudentCourse.StudentPackageID.Set(studentCourse.StudentPackageID.String),
			tmpUpcomingStudentCourse.StudentID.Set(studentCourse.StudentID.String),
			tmpUpcomingStudentCourse.StudentStartDate.Set(upcomingStudentPackage.StartAt.Time),
			tmpUpcomingStudentCourse.StudentEndDate.Set(upcomingStudentPackage.EndAt.Time),
			tmpUpcomingStudentCourse.LocationID.Set(studentCourse.LocationID.String),
			tmpUpcomingStudentCourse.PackageType.Set(studentCourse.PackageType.String),
			tmpUpcomingStudentCourse.CreatedAt.Set(studentCourse.CreatedAt.Time),
			tmpUpcomingStudentCourse.UpdatedAt.Set(studentCourse.UpdatedAt.Time),
			tmpUpcomingStudentCourse.DeletedAt.Set(studentCourse.DeletedAt.Time),
			tmpUpcomingStudentCourse.ResourcePath.Set(studentCourse.ResourcePath.String),
			tmpUpcomingStudentCourse.CourseSlot.Set(nil),
			tmpUpcomingStudentCourse.Weight.Set(nil),
			tmpUpcomingStudentCourse.CourseSlotPerWeek.Set(nil),
		)
		if studentCourse.CourseSlot.Status == pgtype.Present {
			_ = tmpUpcomingStudentCourse.CourseSlot.Set(studentCourse.CourseSlot.Int)
		}
		if studentCourse.CourseSlotPerWeek.Status == pgtype.Present {
			_ = tmpUpcomingStudentCourse.CourseSlotPerWeek.Set(studentCourse.CourseSlotPerWeek.Int)
		}
		if studentCourse.Weight.Status == pgtype.Present {
			_ = tmpUpcomingStudentCourse.Weight.Set(studentCourse.Weight.Int)
		}
		upcomingStudentCourses = append(upcomingStudentCourses, tmpUpcomingStudentCourse)
	}
	return
}

func (s *suite) insertUpcomingStudentPackage(ctx context.Context, upcomingStudentPackage entities.UpcomingStudentPackage) (err error) {
	cmdTag, err := database.InsertExcept(ctx, &upcomingStudentPackage, []string{"resource_path"}, s.FatimaDBTrace.Exec)
	if err != nil {
		err = fmt.Errorf("err insert course: %w", err)
		return
	}

	if cmdTag.RowsAffected() != 1 {
		err = fmt.Errorf("err insert course: %d RowsAffected", cmdTag.RowsAffected())
	}
	return
}

func (s *suite) insertUpcomingStudentCourses(ctx context.Context, upcomingStudentCourses []*entities.UpcomingStudentCourse) (err error) {
	for _, upcomingStudentCourse := range upcomingStudentCourses {
		cmdTag, err := database.InsertExcept(ctx, upcomingStudentCourse, []string{"resource_path"}, s.FatimaDBTrace.Exec)
		if err != nil {
			return fmt.Errorf("err insert course: %w", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err insert course: %d RowsAffected", cmdTag.RowsAffected())
		}
	}
	return
}
