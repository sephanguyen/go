package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LessonManagementService struct {
	CreateLessonV2    func(ctx context.Context, req *lpb.CreateLessonRequest) (*lpb.CreateLessonResponse, error)
	UpdateLessonV2    func(context.Context, *lpb.UpdateLessonRequest) (*lpb.UpdateLessonResponse, error)
	DeleteLessonV2    func(context.Context, *lpb.DeleteLessonRequest) (*lpb.DeleteLessonResponse, error)
	RetrieveLessonsV2 func(context.Context, *lpb.RetrieveLessonsRequest) (*lpb.RetrieveLessonsResponse, error)
}

func (l *LessonManagementService) CreateLesson(ctx context.Context, req *bpb.CreateLessonRequest) (*bpb.CreateLessonResponse, error) {
	// mapping
	studentInfoList := make([]*lpb.CreateLessonRequest_StudentInfo, 0, len(req.StudentInfoList))
	for _, v := range req.StudentInfoList {
		studentInfoList = append(studentInfoList, &lpb.CreateLessonRequest_StudentInfo{
			StudentId:        v.StudentId,
			CourseId:         v.CourseId,
			LocationId:       v.LocationId,
			AttendanceStatus: lpb.StudentAttendStatus(v.AttendanceStatus),
			AttendanceNote:   v.AttendanceNote,
			AttendanceNotice: lpb.StudentAttendanceNotice(v.AttendanceNotice),
			AttendanceReason: lpb.StudentAttendanceReason(v.AttendanceReason),
		})
	}

	materials := make([]*lpb.Material, 0, len(req.Materials))
	for _, v := range req.Materials {
		switch resource := v.Resource.(type) {
		case *bpb.Material_BrightcoveVideo_:
			material := &lpb.Material{
				Resource: &lpb.Material_BrightcoveVideo_{
					BrightcoveVideo: &lpb.Material_BrightcoveVideo{
						Name: resource.BrightcoveVideo.Name,
						Url:  resource.BrightcoveVideo.Url,
					}}}
			materials = append(materials, material)
		case *bpb.Material_MediaId:
			material := &lpb.Material{
				Resource: &lpb.Material_MediaId{
					MediaId: resource.MediaId,
				}}
			materials = append(materials, material)
		default:
			return nil, status.Error(codes.Internal, fmt.Errorf(`unexpected material's type %T`, resource).Error())
		}
	}
	lReq := &lpb.CreateLessonRequest{
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		TeachingMedium:   req.TeachingMedium,
		TeachingMethod:   req.TeachingMethod,
		TeacherIds:       req.TeacherIds,
		LocationId:       req.CenterId,
		StudentInfoList:  studentInfoList,
		Materials:        materials,
		SavingOption:     &lpb.CreateLessonRequest_SavingOption{},
		ClassId:          req.ClassId,
		CourseId:         req.CourseId,
		SchedulingStatus: lpb.LessonStatus(req.SchedulingStatus),
		ClassroomIds:     req.ClassroomIds,
	}

	zoomInfo := req.GetZoomInfo()
	if zoomInfo != nil {
		lReq.ZoomInfo = &lpb.ZoomInfo{
			ZoomAccountOwner: zoomInfo.GetZoomAccountOwner(),
			ZoomLink:         zoomInfo.GetZoomLink(),
			ZoomId:           zoomInfo.GetZoomId(),
			Occurrences: sliceutils.Map(zoomInfo.Occurrences, func(zoomOccurrence *bpb.ZoomInfo_OccurrenceZoom) *lpb.ZoomInfo_OccurrenceZoom {
				return &lpb.ZoomInfo_OccurrenceZoom{
					OccurrenceId: zoomOccurrence.GetOccurrenceId(),
					StartTime:    zoomOccurrence.StartTime,
				}
			}),
		}
	}

	if req.SavingOption != nil {
		lReq.SavingOption = &lpb.CreateLessonRequest_SavingOption{
			Method: lpb.CreateLessonSavingMethod(*req.SavingOption.Method.Enum()),
			Recurrence: &lpb.Recurrence{
				EndDate: func() (endDate *timestamppb.Timestamp) {
					if req.SavingOption.Recurrence != nil {
						endDate = req.SavingOption.Recurrence.EndDate
					}
					return endDate
				}(),
			},
		}
	}

	lRes, err := l.CreateLessonV2(ctx, lReq)
	if err != nil {
		return nil, err
	}

	res := &bpb.CreateLessonResponse{
		Id: lRes.Id,
	}

	return res, nil
}

func (l *LessonManagementService) UpdateLesson(ctx context.Context, req *bpb.UpdateLessonRequest) (*bpb.UpdateLessonResponse, error) {
	studentInfoList := make([]*lpb.UpdateLessonRequest_StudentInfo, 0, len(req.StudentInfoList))
	for _, v := range req.StudentInfoList {
		studentInfoList = append(studentInfoList, &lpb.UpdateLessonRequest_StudentInfo{
			StudentId:        v.StudentId,
			CourseId:         v.CourseId,
			LocationId:       v.LocationId,
			AttendanceStatus: lpb.StudentAttendStatus(v.AttendanceStatus),
			AttendanceNote:   v.AttendanceNote,
			AttendanceNotice: lpb.StudentAttendanceNotice(v.AttendanceNotice),
			AttendanceReason: lpb.StudentAttendanceReason(v.AttendanceReason),
		})
	}

	materials := make([]*lpb.Material, 0, len(req.Materials))
	for _, v := range req.Materials {
		switch resource := v.Resource.(type) {
		case *bpb.Material_BrightcoveVideo_:
			material := &lpb.Material{
				Resource: &lpb.Material_BrightcoveVideo_{
					BrightcoveVideo: &lpb.Material_BrightcoveVideo{
						Name: resource.BrightcoveVideo.Name,
						Url:  resource.BrightcoveVideo.Url,
					}}}
			materials = append(materials, material)
		case *bpb.Material_MediaId:
			material := &lpb.Material{
				Resource: &lpb.Material_MediaId{
					MediaId: resource.MediaId,
				}}
			materials = append(materials, material)
		default:
			return nil, status.Error(codes.Internal, fmt.Errorf(`unexpected material's type %T`, resource).Error())
		}
	}

	lReq := &lpb.UpdateLessonRequest{
		LessonId:         req.LessonId,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		TeachingMedium:   req.TeachingMedium,
		TeachingMethod:   req.TeachingMethod,
		TeacherIds:       req.TeacherIds,
		LocationId:       req.CenterId,
		StudentInfoList:  studentInfoList,
		Materials:        materials,
		ClassId:          req.ClassId,
		CourseId:         req.CourseId,
		SchedulingStatus: lpb.LessonStatus(req.SchedulingStatus),
		ClassroomIds:     req.ClassroomIds,
	}
	zoomInfo := req.GetZoomInfo()
	if zoomInfo != nil {
		lReq.ZoomInfo = &lpb.ZoomInfo{
			ZoomAccountOwner: zoomInfo.GetZoomAccountOwner(),
			ZoomLink:         zoomInfo.GetZoomLink(),
			ZoomId:           zoomInfo.GetZoomId(),
			Occurrences: sliceutils.Map(zoomInfo.Occurrences, func(zoomOccurrence *bpb.ZoomInfo_OccurrenceZoom) *lpb.ZoomInfo_OccurrenceZoom {
				return &lpb.ZoomInfo_OccurrenceZoom{
					OccurrenceId: zoomOccurrence.GetOccurrenceId(),
					StartTime:    zoomOccurrence.StartTime,
				}
			}),
		}
	}

	if req.SavingOption != nil {
		lReq.SavingOption = &lpb.UpdateLessonRequest_SavingOption{
			Method: lpb.CreateLessonSavingMethod(*req.SavingOption.Method.Enum()),
			Recurrence: &lpb.Recurrence{
				EndDate: func() (endDate *timestamppb.Timestamp) {
					if req.SavingOption.Recurrence != nil {
						endDate = req.SavingOption.Recurrence.EndDate
					}
					return endDate
				}(),
			},
		}
	}

	_, err := l.UpdateLessonV2(ctx, lReq)
	if err != nil {
		return nil, err
	}
	return &bpb.UpdateLessonResponse{}, nil
}

func (l *LessonManagementService) RetrieveLessons(ctx context.Context, req *bpb.RetrieveLessonsRequestV2) (*bpb.RetrieveLessonsResponseV2, error) {
	lReq := &lpb.RetrieveLessonsRequest{
		Paging:      req.GetPaging(),
		Keyword:     req.GetKeyword(),
		LessonTime:  lpb.LessonTime(req.GetLessonTime()),
		CurrentTime: req.GetCurrentTime(),
		LocationIds: req.GetLocationIds(),
	}

	if req.GetFilter() != nil {
		filter := &lpb.RetrieveLessonsFilter{
			CourseIds:        req.Filter.GetCourseIds(),
			TeacherIds:       req.Filter.GetTeacherIds(),
			StudentIds:       req.Filter.GetStudentIds(),
			LocationIds:      req.Filter.GetCenterIds(),
			ClassIds:         req.Filter.GetClassIds(),
			Grades:           req.Filter.GetGrades(),
			GradesV2:         req.Filter.GetGradesV2(),
			SchedulingStatus: req.Filter.GetSchedulingStatus(),
			FromDate:         req.Filter.GetFromDate(),
			ToDate:           req.Filter.GetToDate(),
			TimeZone:         req.Filter.GetTimeZone(),
			FromTime:         req.Filter.GetFromTime(),
			ToTime:           req.Filter.GetToTime(),
			DateOfWeeks:      req.Filter.GetDateOfWeeks(),
		}
		lReq.Filter = filter
	}

	lRes, err := l.RetrieveLessonsV2(ctx, lReq)
	if err != nil {
		return nil, err
	}

	items := make([]*bpb.RetrieveLessonsResponseV2_Lesson, 0, len(lRes.Items))
	for _, v := range lRes.Items {
		item := &bpb.RetrieveLessonsResponseV2_Lesson{
			Id:               v.Id,
			Name:             v.Name,
			StartTime:        v.StartTime,
			EndTime:          v.EndTime,
			CenterId:         v.CenterId,
			TeacherIds:       v.TeacherIds,
			TeachingMethod:   v.TeachingMethod,
			TeachingMedium:   v.TeachingMedium,
			CourseId:         v.CourseId,
			ClassId:          v.ClassId,
			SchedulingStatus: v.SchedulingStatus,
		}
		items = append(items, item)
	}

	return &bpb.RetrieveLessonsResponseV2{
		Items:      items,
		TotalItems: lRes.TotalItems,
		NextPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: lRes.NextPage.GetOffsetString(),
			},
		},
		PreviousPage: &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: lRes.PreviousPage.GetOffsetString(),
			},
		},
		TotalLesson: lRes.TotalLesson,
	}, nil
}

func (l *LessonManagementService) DeleteLesson(ctx context.Context, req *bpb.DeleteLessonRequest) (*bpb.DeleteLessonResponse, error) {
	lReq := &lpb.DeleteLessonRequest{
		LessonId: req.LessonId,
	}

	if req.SavingOption != nil {
		lReq.SavingOption = &lpb.DeleteLessonRequest_SavingOption{Method: lpb.CreateLessonSavingMethod(req.SavingOption.Method)}
	}

	_, err := l.DeleteLessonV2(ctx, lReq)

	return &bpb.DeleteLessonResponse{}, err
}
