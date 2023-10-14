package queries

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/schedule_class/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type ReserveClassQueryHandler struct {
	DB                      database.Ext
	ReserveClassRepo        infrastructure.ReserveClassRepo
	CourseRepo              infrastructure.CourseRepo
	ClassRepo               infrastructure.ClassRepo
	StudentPackageClassRepo infrastructure.StudentPackageClassRepo
}

func (r *ReserveClassQueryHandler) RetrieveScheduledClass(ctx context.Context, studentID string) (*mpb.RetrieveScheduledStudentClassResponse, error) {
	reserveClasses, err := r.ReserveClassRepo.GetByStudentIDs(ctx, r.DB, studentID)
	if err != nil {
		return nil, fmt.Errorf("query reserve class fail: %w", err)
	}

	querySpc := make([]string, 0)
	courseIDs := make([]string, 0)
	classIDs := make([]string, 0)

	if len(reserveClasses) == 0 {
		return &mpb.RetrieveScheduledStudentClassResponse{
			ScheduledClasses: []*mpb.RetrieveScheduledStudentClassResponse_ScheduledClassInfo{},
		}, nil
	}

	for _, rc := range reserveClasses {
		querySpc = append(querySpc, fmt.Sprintf("('%s', '%s', '%s')", rc.StudentPackageID, studentID, rc.CourseID))
		courseIDs = append(courseIDs, rc.CourseID)
		classIDs = append(classIDs, rc.ClassID)
	}
	querySpcString := strings.Join(querySpc, ", ")

	spcList, mapSpc, err := r.StudentPackageClassRepo.GetManyByStudentPackageIDAndStudentIDAndCourseID(ctx, r.DB, querySpcString)
	if err != nil {
		return nil, fmt.Errorf("query student package class fail: %w", err)
	}

	classIDs = append(classIDs, sliceutils.Map(spcList, func(spc *repo.StudentPackageClassDTO) string {
		return spc.ClassID.String
	})...)

	uniqCourseIDs := golibs.Uniq(courseIDs)
	uniqClassIDs := golibs.Uniq(classIDs)

	mapCourse, err := r.CourseRepo.GetMapCourseByIDs(ctx, r.DB, uniqCourseIDs)
	if err != nil {
		return nil, fmt.Errorf("query course fail: %w", err)
	}

	mapClass, err := r.ClassRepo.GetMapClassByIDs(ctx, r.DB, uniqClassIDs)
	if err != nil {
		return nil, fmt.Errorf("query class fail: %w", err)
	}

	itemList := make([]*mpb.RetrieveScheduledStudentClassResponse_ScheduledClassInfo, 0)

	sort.Slice(reserveClasses, func(i, j int) bool {
		return reserveClasses[i].EffectiveDate.Before(reserveClasses[j].EffectiveDate)
	})

	for _, rc := range reserveClasses {
		course, ok := mapCourse[rc.CourseID]

		if !ok {
			return nil, fmt.Errorf("not found course")
		}

		scheduleClass, ok := mapClass[rc.ClassID]

		if !ok {
			return nil, fmt.Errorf("not found scheduled class")
		}

		spc, ok := mapSpc[r.StudentPackageClassRepo.GetStudentPackageClassID(rc.StudentPackageID, rc.StudentID, rc.CourseID)]

		if !ok {
			return nil, fmt.Errorf("not found student package class")
		}

		currentClassID := spc.ClassID.String

		currentClass, ok := mapClass[currentClassID]

		if !ok {
			return nil, fmt.Errorf("not found current active class")
		}

		item := &mpb.RetrieveScheduledStudentClassResponse_ScheduledClassInfo{
			CourseId:   course.CourseID.String,
			CourseName: course.Name.String,
			CurrentClass: &mpb.RetrieveScheduledStudentClassResponse_ClassInfo{
				ClassId: currentClass.ClassID.String,
				Name:    currentClass.Name.String,
			},
			ScheduledClass: &mpb.RetrieveScheduledStudentClassResponse_ClassInfo{
				ClassId: scheduleClass.ClassID.String,
				Name:    scheduleClass.Name.String,
			},
			EffectiveDate: &timestamppb.Timestamp{Seconds: rc.EffectiveDate.Unix()},
		}

		itemList = append(itemList, item)
	}
	result := &mpb.RetrieveScheduledStudentClassResponse{
		ScheduledClasses: itemList,
	}

	return result, nil
}

func (r *ReserveClassQueryHandler) GetReserveClassesByEffectiveDate(ctx context.Context, date string) ([]*mpb.GetReserveClassesByEffectiveDateResponse_ReserveClass, error) {
	resp, err := r.ReserveClassRepo.GetByEffectiveDate(ctx, r.DB, date)

	if err != nil {
		return nil, fmt.Errorf("ReserveClassRepo.GetByEffectiveDate %w", err)
	}

	reserveClasses := sliceutils.Map(resp, func(rc *domain.ReserveClass) *mpb.GetReserveClassesByEffectiveDateResponse_ReserveClass {
		return &mpb.GetReserveClassesByEffectiveDateResponse_ReserveClass{
			StudentPackageId: rc.StudentPackageID,
			StudentId:        rc.StudentID,
			CourseId:         rc.CourseID,
			ClassId:          rc.ClassID,
		}
	})

	return reserveClasses, nil
}
