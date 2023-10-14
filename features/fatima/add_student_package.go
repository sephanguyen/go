package fatima

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/fatima/entities"
	"github.com/manabie-com/backend/internal/fatima/repositories"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) ServerMustStoreThisPackageForThisStudent() error {
	return s.serverMustStoreThisPackageForThisStudent()
}

func (s *suite) serverMustStoreThisPackageForThisStudent() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	token, err := s.GenerateValidAuthenticationToken(s.Request.(*pb.AddStudentPackageRequest).StudentId)
	if err != nil {
		return err
	}

	s.AuthToken = token
	err = s.userRetrieveAccessibleCourse()
	if err != nil {
		return err
	}

	_, err = s.returnsAllCourseAccessibleResponseOfThisUser(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *suite) UserAddAPackageForAStudent(packageID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := s.userAddAPackageForAStudent(ctx, packageID)
	return err
}

func (s *suite) userAddAPackageForAStudent(ctx context.Context, packageID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// togglePackageRequest, err := s.toggleActivePackageRequest(ctx, packageID)
	// if err != nil {
	// 	return ctx, err
	// }

	req := &pb.AddStudentPackageRequest{
		StudentId: ksuid.New().String(),
		PackageId: ksuid.New().String(),
	}

	s.Response, s.ResponseErr = pb.NewSubscriptionModifierServiceClient(s.Conn).AddStudentPackage(contextWithToken(s, ctx), req)
	s.Request = req

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) ServerMustStoreTheseCoursesForThisStudent(ctx context.Context) error {
	return s.serverMustStoreTheseCoursesForThisStudent(ctx)
}

func (s *suite) serverMustStoreTheseCoursesForThisStudent(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	token, err := s.GenerateValidAuthenticationToken(s.Request.(*pb.AddStudentPackageCourseRequest).StudentId)
	if err != nil {
		return err
	}

	spID := s.Response.(*pb.AddStudentPackageCourseResponse).StudentPackageId
	repo := &repositories.StudentPackageRepo{}
	sp, err := repo.Get(ctx, s.DB, database.Text(spID))
	if err != nil {
		return fmt.Errorf("err find package: %w", err)
	}

	err = s.validateStudentPackageAccessPath(ctx, sp)
	if err != nil {
		return err
	}

	s.AuthToken = token
	err = s.userRetrieveAccessibleCourse()
	if err != nil {
		return fmt.Errorf("err s.userRetrieveAccessibleCourse: %w", err)
	}

	_, err = s.returnsAllCourseAccessibleResponseOfThisUser(ctx)
	if err != nil {
		return fmt.Errorf("err s.returnsAllCourseAccessibleResponseOfThisUser: %w", err)
	}

	err = s.validateStudentPackageWithLocationIds(ctx)
	if err != nil {
		return fmt.Errorf("err s.validateStudentPackageAccessPath: %w", err)
	}

	return nil
}

func (s *suite) serverMustStoreTheseCoursesAndClassForThisStudent(ctx context.Context) error {
	spID := s.Response.(*pb.AddStudentPackageCourseResponse).StudentPackageId
	repo := &repositories.StudentPackageRepo{}
	sp, err := repo.Get(ctx, s.DB, database.Text(spID))
	if err != nil {
		return fmt.Errorf("err find package: %w", err)
	}

	err = s.validateStudentPackageClass(ctx, sp)
	if err != nil {
		return err
	}
	err = s.validateStudentPackageAccessPath(ctx, sp)
	if err != nil {
		return err
	}

	select {
	case <-s.FoundChanForJetStream:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for event to be published")
	}
}

func (s *suite) validateStudentPackageClass(ctx context.Context, sp *entities.StudentPackage) error {
	studentPackageClassEntity := &entities.StudentPackageClass{}

	courseIDs, err := sp.GetCourseIDs()
	if err != nil {
		return fmt.Errorf("err parse sp.GetCourseIDs(): %w", err)
	}

	var locationIDs []string
	if err := sp.LocationIDs.AssignTo(&locationIDs); err != nil {
		return fmt.Errorf("err parse locationIDs: %w", err)
	}

	query := fmt.Sprintf(`SELECT %s FROM student_package_class WHERE student_package_id = $1 AND deleted_at is NULL`, strings.Join(database.GetFieldNames(studentPackageClassEntity), ", "))
	rows, err := s.DB.Query(ctx, query, sp.ID)
	if err != nil {
		return fmt.Errorf("err s.DB.Query: %w", err)
	}
	defer rows.Close()
	studentPackageClasses := make([]*entities.StudentPackageClass, 0)

	for rows.Next() {
		studentPackageClassEntity = &entities.StudentPackageClass{}
		err := rows.Scan(database.GetScanFields(studentPackageClassEntity, database.GetFieldNames(studentPackageClassEntity))...)
		if err != nil {
			return fmt.Errorf("err rows.Scan: %w", err)
		}
		if studentPackageClassEntity.StudentID.String != sp.StudentID.String {
			return fmt.Errorf("err expected student_id is %s, but actual is %v", sp.StudentID.String, studentPackageClassEntity.StudentID.String)
		}
		if !golibs.InArrayString(studentPackageClassEntity.ClassID.String, s.ClassIDs) {
			return fmt.Errorf("err expected class_id %s in %v", studentPackageClassEntity.ClassID.String, s.ClassIDs)
		}
		if !(golibs.InArrayString(studentPackageClassEntity.LocationID.String, locationIDs)) {
			return fmt.Errorf("err expected location_id %s in %v", studentPackageClassEntity.LocationID.String, locationIDs)
		}
		if !(golibs.InArrayString(studentPackageClassEntity.CourseID.String, courseIDs)) {
			return fmt.Errorf("err expected course_id %s in %v", studentPackageClassEntity.CourseID.String, courseIDs)
		}
		studentPackageClasses = append(studentPackageClasses, studentPackageClassEntity)
	}
	if len(courseIDs) != len(studentPackageClasses) {
		return fmt.Errorf("total student_package_class records is not expected")
	}

	return nil
}

func (s *suite) validateStudentPackageAccessPath(ctx context.Context, sp *entities.StudentPackage) error {
	e := &entities.StudentPackageAccessPath{}

	courseIDs, err := sp.GetCourseIDs()
	if err != nil {
		return fmt.Errorf("err parse sp.GetCourseIDs(): %w", err)
	}

	var locationIDs []string
	if err := sp.LocationIDs.AssignTo(&locationIDs); err != nil {
		return fmt.Errorf("err parse locationIDs: %w", err)
	}

	query := fmt.Sprintf(`SELECT %s FROM student_package_access_path WHERE student_package_id = $1 AND deleted_at is NULL`, strings.Join(database.GetFieldNames(e), ", "))
	rows, err := s.DB.Query(
		ctx,
		query,
		sp.ID,
	)
	if err != nil {
		return fmt.Errorf("err s.DB.Query: %w", err)
	}
	defer rows.Close()

	studentPackageAccessPaths := make([]*entities.StudentPackageAccessPath, 0)
	for rows.Next() {
		e := &entities.StudentPackageAccessPath{}
		err := rows.Scan(database.GetScanFields(e, database.GetFieldNames(e))...)
		if err != nil {
			return fmt.Errorf("err rows.Scan: %w", err)
		}
		if e.StudentID.String != sp.StudentID.String {
			return fmt.Errorf("err expected student_id is %s, but actual is %v", sp.StudentID.String, e.StudentID.String)
		}
		if len(locationIDs) > 0 {
			if !(golibs.InArrayString(e.LocationID.String, locationIDs)) {
				return fmt.Errorf("err expected location_id %s in %v", e.LocationID.String, locationIDs)
			}
		} else {
			if e.LocationID.String != "" {
				return fmt.Errorf("err expected location_id is '', but actual is %v", e.LocationID.String)
			}
		}
		if !(golibs.InArrayString(e.CourseID.String, courseIDs)) {
			return fmt.Errorf("err expected course_id %s in %v", e.CourseID.String, courseIDs)
		}
		studentPackageAccessPaths = append(studentPackageAccessPaths, e)
	}

	count := len(studentPackageAccessPaths)
	if len(locationIDs) > 0 {
		if count != (len(courseIDs) * len(locationIDs)) {
			return fmt.Errorf("total student_package_access_path records is not expected actual is %d, expect %d", count, len(courseIDs)*len(locationIDs))
		}
	} else {
		if count != len(courseIDs) {
			return fmt.Errorf("total student_package_access_path records is not expected actual is %d, expect %d", count, len(courseIDs))
		}
	}

	return nil
}

func (s *suite) validateStudentPackageWithLocationIds(ctx context.Context) error {
	var studentPackageID string
	query := "SELECT student_package_id FROM student_packages WHERE student_id = $1 and location_ids @> $2"
	if err := database.Select(
		ctx,
		s.DB,
		query,
		s.Request.(*pb.AddStudentPackageCourseRequest).StudentId,
		database.TextArray(s.Request.(*pb.AddStudentPackageCourseRequest).LocationIds),
	).ScanFields(&studentPackageID); err != nil {
		return err
	}

	return nil
}

func (s *suite) UserAddAPackageForAStudentV2(packageID string) error {
	s.StudentID = ksuid.New().String()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	req := &pb.AddStudentPackageRequest{
		StudentId: s.StudentID,
		PackageId: s.Response.(*pb.CreatePackageResponse).PackageId,
	}

	s.Response, s.ResponseErr = pb.NewSubscriptionModifierServiceClient(s.Conn).AddStudentPackage(contextWithToken(s, ctx), req)
	s.Request = req

	if s.Response != nil {
		s.StudentPackageID = s.Response.(*pb.AddStudentPackageResponse).StudentPackageId
	}
	return nil
}

func (s *suite) UserAddACourseForAStudent() error {
	return s.userAddACourseForAStudent()
}

func (s *suite) userAddACourseForAStudent() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now()
	startAt := timestamppb.Now()
	endAt := timestamppb.New(now.Add(7 * 24 * time.Hour))
	s.StudentID = ksuid.New().String()

	req := &pb.AddStudentPackageCourseRequest{
		StudentId:   s.StudentID,
		CourseIds:   s.CourseIDs,
		StartAt:     startAt,
		EndAt:       endAt,
		LocationIds: []string{constants.ManabieOrgLocation},
	}
	s.Response, s.ResponseErr = pb.NewSubscriptionModifierServiceClient(s.Conn).AddStudentPackageCourse(contextWithToken(s, ctx), req)
	s.Request = req

	return nil
}

func (s *suite) userAddCourseWithStudentPackageExtraForAStudent() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	now := time.Now()
	startAt := timestamppb.Now()
	endAt := timestamppb.New(now.Add(7 * 24 * time.Hour))
	studentPackageExtras := make([]*pb.AddStudentPackageCourseRequest_AddStudentPackageExtra, 0)

	newClassId := idutil.ULIDNow()
	newLocationID := constants.ManabieOrgLocation
	studentPackageExtras = append(studentPackageExtras, &pb.AddStudentPackageCourseRequest_AddStudentPackageExtra{
		CourseId:   s.CourseIDs[0],
		LocationId: constants.ManabieOrgLocation,
		ClassId:    newClassId,
	})
	s.LocationIDs = append(s.LocationIDs, newLocationID)
	s.ClassIDs = append(s.ClassIDs, newClassId)

	req := &pb.AddStudentPackageCourseRequest{
		CourseIds:           s.CourseIDs,
		StudentId:           s.StudentID,
		StartAt:             startAt,
		EndAt:               endAt,
		StudentPackageExtra: studentPackageExtras,
	}
	s.Request = req
	err := s.createStudentPackageUpsertedV2Subscribe()
	if err != nil {
		return err
	}

	s.Response, s.ResponseErr = pb.NewSubscriptionModifierServiceClient(s.Conn).AddStudentPackageCourse(contextWithToken(s, ctx), req)
	return nil
}
