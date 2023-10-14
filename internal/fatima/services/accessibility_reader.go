package services

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/manabie-com/backend/internal/fatima/entities"
	global_constant "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccessibilityReadService struct {
	DB database.Ext

	StudentPackageRepo interface {
		CurrentPackage(ctx context.Context, db database.QueryExecer, studentID pgtype.Text) ([]*entities.StudentPackage, error)
	}
}

func (c *AccessibilityReadService) RetrieveAccessibility(ctx context.Context, req *fpb.RetrieveAccessibilityRequest) (*fpb.RetrieveAccessibilityResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	resp, err := c.retrieveUserAccessibility(ctx, &fpb.RetrieveStudentAccessibilityRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	courses := make(map[string]*fpb.RetrieveAccessibilityResponse_CourseAccessibility, len(resp.Courses))
	for key, c := range resp.Courses {
		courses[key] = convCourseAccessibilityCommonToSpecific(c)
	}

	return &fpb.RetrieveAccessibilityResponse{Courses: courses}, nil
}

func (s *AccessibilityReadService) RetrieveStudentAccessibility(ctx context.Context, req *fpb.RetrieveStudentAccessibilityRequest) (*fpb.RetrieveStudentAccessibilityResponse, error) {
	return s.retrieveUserAccessibility(ctx, req)
}

func (c *AccessibilityReadService) retrieveUserAccessibility(ctx context.Context, req *fpb.RetrieveStudentAccessibilityRequest) (*fpb.RetrieveStudentAccessibilityResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "student id cannot be empty")
	}
	studentPackages, err := c.StudentPackageRepo.CurrentPackage(ctx, c.DB, database.Text(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	courses := map[string]*cpb.CourseAccessibility{}
	for _, sp := range studentPackages {
		props, err := sp.GetProperties()
		if err != nil {
			return nil, status.Error(codes.Internal, "err GetProperties: "+err.Error())
		}

		totalCanWatchVideo := len(props.CanWatchVideo)
		totalCanViewStudyGuide := len(props.CanViewStudyGuide)
		totalCanDoQuiz := len(props.CanDoQuiz)

		// get the biggest array size
		maxCourseID := totalCanWatchVideo
		if maxCourseID < totalCanViewStudyGuide {
			maxCourseID = totalCanViewStudyGuide
		}

		if maxCourseID < totalCanDoQuiz {
			maxCourseID = totalCanDoQuiz
		}

		// loop the biggest array, then access the small array to reduce loop check each array
		for i := 0; i < maxCourseID; i++ {
			if totalCanWatchVideo > i {
				courseID := props.CanWatchVideo[i]
				if _, ok := courses[courseID]; !ok {
					courses[courseID] = &cpb.CourseAccessibility{}
				}

				courses[courseID].CanWatchVideo = true
			}

			if totalCanViewStudyGuide > i {
				courseID := props.CanViewStudyGuide[i]
				if _, ok := courses[courseID]; !ok {
					courses[courseID] = &cpb.CourseAccessibility{}
				}

				courses[courseID].CanViewStudyGuide = true
			}

			if totalCanDoQuiz > i {
				courseID := props.CanDoQuiz[i]
				if _, ok := courses[courseID]; !ok {
					courses[courseID] = &cpb.CourseAccessibility{}
				}

				courses[courseID].CanDoQuiz = true
			}
		}
	}
	return &fpb.RetrieveStudentAccessibilityResponse{
		Courses: courses,
	}, nil
}

func toStudentPackage(request *npb.EventSyncStudentPackage_StudentPackage) ([]*entities.StudentPackage, error) {
	studentPackages := []*entities.StudentPackage{}

	for index := 0; index < len(request.Packages); index++ {
		value := request.Packages[index]
		id := getStudentPackageID(request.StudentId, value)

		studentPackage := &entities.StudentPackage{}
		studentPackageProps := &entities.StudentPackageProps{CanWatchVideo: value.CourseIds, CanDoQuiz: value.CourseIds, CanViewStudyGuide: value.CourseIds}

		defaultIDs := []string{global_constant.JPREPOrgLocation}
		database.AllNullEntity(studentPackage)
		err := multierr.Combine(
			studentPackage.ID.Set(id),
			studentPackage.StartAt.Set(value.StartDate.AsTime()),
			studentPackage.EndAt.Set(value.EndDate.AsTime()),
			studentPackage.IsActive.Set(true),
			studentPackage.StudentID.Set(request.StudentId),
			studentPackage.Properties.Set(studentPackageProps),
			studentPackage.LocationIDs.Set(defaultIDs),
		)
		if err != nil {
			return nil, fmt.Errorf("err set StudentPackage: %w", err)
		}

		studentPackages = append(studentPackages, studentPackage)
	}

	return studentPackages, nil
}

func getStudentPackageID(studentID string, e *npb.EventSyncStudentPackage_Package) string {
	data := []byte(fmt.Sprintf("%s-%v", studentID, e.String()))
	return fmt.Sprintf("%x", md5.Sum(data))
}

//convCourseAccessibilityCommonToSpecific convert element from type `*cpb.CourseAccessibility` to `*fpb.RetrieveAccessibilityResponse_CourseAccessibility`
func convCourseAccessibilityCommonToSpecific(course *cpb.CourseAccessibility) *fpb.RetrieveAccessibilityResponse_CourseAccessibility {
	return &fpb.RetrieveAccessibilityResponse_CourseAccessibility{
		CanDoQuiz:         course.CanDoQuiz,
		CanWatchVideo:     course.CanWatchVideo,
		CanViewStudyGuide: course.CanViewStudyGuide,
	}
}
