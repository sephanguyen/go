package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lesson_infras "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/application"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/constant"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LessonReportModifierService struct {
	wrapperConnection *support.WrapperDBConnection
	// commands
	lessonReportCommand          application.LessonReportCommand
	UpdateLessonSchedulingStatus func(ctx context.Context, req *lpb.UpdateLessonSchedulingStatusRequest) (*lpb.UpdateLessonSchedulingStatusResponse, error)
	UnleashClientIns             unleashclient.ClientInstance
	Env                          string
}

func NewLessonReportModifierService(
	wrapperConnection *support.WrapperDBConnection,
	lessonRepo lesson_infras.LessonRepo,
	lessonMemberRepo lesson_infras.LessonMemberRepo,
	lessonReportRepo infrastructure.LessonReportRepo,
	lessonReportDetailRepo infrastructure.LessonReportDetailRepo,
	partnerFormConfigRepo infrastructure.PartnerFormConfigRepo,
	reallocationRepo lesson_infras.ReallocationRepo,
	updateLessonSchedulingStatus func(ctx context.Context, req *lpb.UpdateLessonSchedulingStatusRequest) (*lpb.UpdateLessonSchedulingStatusResponse, error),
	unleashClientIns unleashclient.ClientInstance,
	env string,
	masterDataPort lesson_infras.MasterDataPort,

) *LessonReportModifierService {
	return &LessonReportModifierService{
		wrapperConnection: wrapperConnection,
		lessonReportCommand: application.LessonReportCommand{
			LessonRepo:             lessonRepo,
			LessonMemberRepo:       lessonMemberRepo,
			LessonReportRepo:       lessonReportRepo,
			LessonReportDetailRepo: lessonReportDetailRepo,
			PartnerFormConfigRepo:  partnerFormConfigRepo,
			ReallocationRepo:       reallocationRepo,
			MasterDataPort:         masterDataPort,
			Logger:                 zap.NewNop(),
		},
		UpdateLessonSchedulingStatus: updateLessonSchedulingStatus,
		UnleashClientIns:             unleashClientIns,
		Env:                          env,
	}
}

type NewLessonReportOption func(*domain.LessonReport) error

// Individual
func (l *LessonReportModifierService) SaveDraftIndividualLessonReport(ctx context.Context, req *lpb.WriteIndividualLessonReportRequest) (*lpb.SaveDraftIndividualLessonReportResponse, error) {
	isUnleashToggledOptimisticLocking, err := l.UnleashClientIns.IsFeatureEnabled(constant.OptimisticLockingLessonReport, l.Env)

	if err != nil {
		return nil, fmt.Errorf("l.connectToUnleash: %w", err)
	}

	lessonReport, err := NewLessonReport(ByIndividualLessonReportGRPCMessage(req))
	lessonReport.UnleashToggles[constant.OptimisticLockingLessonReport] = isUnleashToggledOptimisticLocking

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	conn, err := l.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	res := &lpb.SaveDraftIndividualLessonReportResponse{}
	err = l.lessonReportCommand.SaveDraft(ctx, conn, lessonReport)
	if err != nil {
		errorString := fmt.Sprintf("could not save draft lesson report: %v", err)
		if strings.Contains(err.Error(), "validateOptimisticLockingLessonReport") {
			res.Error = errorString
			return res, nil
		}
		return nil, status.Error(codes.Internal, errorString)
	}
	res.LessonReportId = lessonReport.LessonReportID
	return res, nil
}

func ByIndividualLessonReportGRPCMessage(req *lpb.WriteIndividualLessonReportRequest) NewLessonReportOption {
	return func(l *domain.LessonReport) (err error) {
		l.LessonReportID = req.LessonReportId
		l.LessonID = req.LessonId
		l.FeatureName = req.FeatureName
		details := make([]*domain.LessonReportDetail, 0, len(l.Details))
		for _, detail := range req.Details {
			lrd := &domain.LessonReportDetail{
				StudentID:        detail.StudentId,
				AttendanceStatus: constant.StudentAttendStatus(detail.AttendanceStatus.String()),
				AttendanceRemark: detail.AttendanceRemark,
				AttendanceNotice: constant.StudentAttendanceNotice(detail.AttendanceNotice.String()),
				AttendanceReason: constant.StudentAttendanceReason(detail.AttendanceReason.String()),
				AttendanceNote:   detail.AttendanceNote,
				ReportVersion:    int(detail.ReportVersion),
			}
			lrd.Fields, err = domain.LessonReportFieldsFromDynamicFieldValueGRPC(detail.FieldValues...)
			if err != nil {
				return fmt.Errorf("got error when parse DynamicFieldValue GRPC message to LessonReportFields: %v", err)
			}
			details = append(details, lrd)
		}
		l.Details = append(l.Details, details...)
		l.IsUpdateMembersInfo = true
		l.IsSavePerStudent = req.IsSavePerStudent
		return nil
	}
}
func (l *LessonReportModifierService) SubmitIndividualLessonReport(ctx context.Context, req *lpb.WriteIndividualLessonReportRequest) (*lpb.SubmitIndividualLessonReportResponse, error) {
	res := &lpb.SubmitIndividualLessonReportResponse{}
	lessonReport, err := NewLessonReport(ByIndividualLessonReportGRPCMessage(req))

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	isPermissionUnleashToggled, err := l.UnleashClientIns.IsFeatureEnabled(constant.PermissionToSubmitReport, l.Env)
	isUnleashToggledOptimisticLocking, err := l.UnleashClientIns.IsFeatureEnabled(constant.OptimisticLockingLessonReport, l.Env)

	if err != nil {
		return nil, fmt.Errorf("l.connectToUnleash: %w", err)
	}

	lessonReport.UnleashToggles[constant.PermissionToSubmitReport] = isPermissionUnleashToggled
	lessonReport.UnleashToggles[constant.OptimisticLockingLessonReport] = isUnleashToggledOptimisticLocking

	conn, err := l.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if err = l.lessonReportCommand.Submit(ctx, conn, lessonReport); err != nil {
		errorString := fmt.Sprintf("could not submit lesson report: %v", err)
		if strings.Contains(err.Error(), "validateOptimisticLockingLessonReport") {
			res.Error = errorString
			return res, nil
		}
		return nil, status.Error(codes.Internal, errorString)
	}
	isUnleashToggled, err := l.UnleashClientIns.IsFeatureEnabled("BACKEND_Lesson_HandleUpdateStatusWhenUserSubmitLessonReport", l.Env)
	if err != nil {
		return nil, fmt.Errorf("l.connectToUnleash: %w", err)
	}

	if isUnleashToggled {
		publishedSchedulingStatuses := []constant.LessonSchedulingStatus{}
		schedulingStatus := constant.LessonSchedulingStatus(lessonReport.Lesson.SchedulingStatus)
		if schedulingStatus == constant.LessonSchedulingStatusDraft {
			publishedSchedulingStatuses = []constant.LessonSchedulingStatus{constant.LessonSchedulingStatusPublished, constant.LessonSchedulingStatusCompleted}
		} else if schedulingStatus == constant.LessonSchedulingStatusPublished {
			publishedSchedulingStatuses = []constant.LessonSchedulingStatus{constant.LessonSchedulingStatusCompleted}
		}

		for _, v := range publishedSchedulingStatuses {
			if _, err := l.UpdateLessonSchedulingStatus(ctx, &lpb.UpdateLessonSchedulingStatusRequest{
				LessonId:         lessonReport.LessonID,
				SchedulingStatus: cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(v)]),
			}); err != nil {
				return nil, err
			}
		}
	}

	res.LessonReportId = lessonReport.LessonReportID
	return res, nil
}

// Group
func (l *LessonReportModifierService) SaveDraftGroupLessonReport(ctx context.Context, req *lpb.WriteGroupLessonReportRequest) (*lpb.SaveDraftGroupLessonReportResponse, error) {
	res := &lpb.SaveDraftGroupLessonReportResponse{}
	lessonReport, err := NewLessonReport(ByGroupLessonReportGRPCMessage(req))

	isUnleashToggledOptimisticLocking, err := l.UnleashClientIns.IsFeatureEnabled(constant.OptimisticLockingLessonReport, l.Env)
	lessonReport.UnleashToggles[constant.OptimisticLockingLessonReport] = isUnleashToggledOptimisticLocking

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	conn, err := l.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	err = l.lessonReportCommand.SaveDraft(ctx, conn, lessonReport)
	if err != nil {
		errorString := fmt.Sprintf("could not save draft lesson report: %v", err)
		if strings.Contains(err.Error(), "validateOptimisticLockingLessonReport") {
			res.Error = errorString
			return res, nil
		}
		return nil, status.Error(codes.Internal, errorString)
	}
	res.LessonReportId = lessonReport.LessonReportID
	return res, nil
}

func (l *LessonReportModifierService) SubmitGroupLessonReport(ctx context.Context, req *lpb.WriteGroupLessonReportRequest) (*lpb.SubmitGroupLessonReportResponse, error) {
	res := &lpb.SubmitGroupLessonReportResponse{}
	lessonReport, err := NewLessonReport(ByGroupLessonReportGRPCMessage(req))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	isPermissionUnleashToggled, err1 := l.UnleashClientIns.IsFeatureEnabled(constant.PermissionToSubmitReport, l.Env)
	isAttendanceUnleashToggled, err2 := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_BackOffice_ValidationLessonBeforeCompleted", l.Env)
	isUnleashToggledOptimisticLocking, err3 := l.UnleashClientIns.IsFeatureEnabled(constant.OptimisticLockingLessonReport, l.Env)

	err = multierr.Combine(err1, err2, err3)
	if err != nil {
		return nil, fmt.Errorf("l.connectToUnleash: %w", err)
	}
	lessonReport.UnleashToggles[constant.PermissionToSubmitReport] = isPermissionUnleashToggled
	lessonReport.UnleashToggles["Lesson_LessonManagement_BackOffice_ValidationLessonBeforeCompleted"] = isAttendanceUnleashToggled
	lessonReport.UnleashToggles[constant.OptimisticLockingLessonReport] = isUnleashToggledOptimisticLocking

	conn, err := l.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	err = l.lessonReportCommand.Submit(ctx, conn, lessonReport)
	if err != nil {
		errorString := fmt.Sprintf("could not submit lesson report: %v", err)
		if strings.Contains(err.Error(), "validateOptimisticLockingLessonReport") {
			res.Error = errorString
			return res, nil
		}
		return nil, status.Error(codes.Internal, errorString)
	}
	isUnleashToggled, err := l.UnleashClientIns.IsFeatureEnabled("BACKEND_Lesson_HandleUpdateStatusWhenUserSubmitLessonReport", l.Env)
	if err != nil {
		return nil, fmt.Errorf("l.connectToUnleash: %w", err)
	}

	if isUnleashToggled {
		schedulingStatus := lesson_domain.LessonSchedulingStatus(lessonReport.Lesson.SchedulingStatus)
		publishedSchedulingStatuses := []lesson_domain.LessonSchedulingStatus{}
		isAttendance := true

		if isAttendanceUnleashToggled {
			isAttendance = lessonReport.Details.CheckStudentsAttendance()
		}

		if schedulingStatus == lesson_domain.LessonSchedulingStatusDraft {
			publishedSchedulingStatuses = []lesson_domain.LessonSchedulingStatus{lesson_domain.LessonSchedulingStatusPublished, lesson_domain.LessonSchedulingStatusCompleted}
			if isAttendanceUnleashToggled && !isAttendance {
				publishedSchedulingStatuses = []lesson_domain.LessonSchedulingStatus{lesson_domain.LessonSchedulingStatusPublished}
			}
		} else if schedulingStatus == lesson_domain.LessonSchedulingStatusPublished {
			if !isAttendanceUnleashToggled || (isAttendanceUnleashToggled && isAttendance) {
				publishedSchedulingStatuses = []lesson_domain.LessonSchedulingStatus{lesson_domain.LessonSchedulingStatusCompleted}
			}
		}

		for _, v := range publishedSchedulingStatuses {
			if _, err := l.UpdateLessonSchedulingStatus(ctx, &lpb.UpdateLessonSchedulingStatusRequest{
				LessonId:         lessonReport.LessonID,
				SchedulingStatus: cpb.LessonSchedulingStatus(cpb.LessonSchedulingStatus_value[string(v)]),
			}); err != nil {
				return nil, err
			}
		}
	}
	res.LessonReportId = lessonReport.LessonReportID
	return res, nil
}

func NewLessonReport(opts ...NewLessonReportOption) (*domain.LessonReport, error) {
	lessonRp := &domain.LessonReport{}
	for _, opt := range opts {
		if err := opt(lessonRp); err != nil {
			return nil, err
		}
	}
	lessonRp.UnleashToggles = make(map[string]bool)
	return lessonRp, nil
}

// ByLessonReportGRPCMessage will create lesson report object by LessonReport GRPC message
func ByGroupLessonReportGRPCMessage(req *lpb.WriteGroupLessonReportRequest) NewLessonReportOption {
	return func(l *domain.LessonReport) (err error) {
		l.LessonReportID = req.LessonReportId
		l.LessonID = req.LessonId

		details := make([]*domain.LessonReportDetail, 0, len(l.Details))
		for _, detail := range req.Details {
			lrd := &domain.LessonReportDetail{
				StudentID:        detail.StudentId,
				AttendanceStatus: constant.StudentAttendStatus(detail.GetAttendanceStatus().String()),
				AttendanceRemark: detail.GetAttendanceRemark(),
				AttendanceNotice: constant.StudentAttendanceNotice(detail.GetAttendanceNotice().String()),
				AttendanceReason: constant.StudentAttendanceReason(detail.GetAttendanceReason().String()),
				AttendanceNote:   detail.GetAttendanceNote(),
				ReportVersion:    int(detail.ReportVersion),
			}
			lrd.Fields, err = domain.LessonReportFieldsFromDynamicFieldValueGRPC(detail.FieldValues...)
			if err != nil {
				return fmt.Errorf("SaveDraftGroupLessonReport().ByGroupLessonReportGRPCMessage():error parsing DynamicFieldValue GRPC message to LessonReportFields: %v", err)
			}
			if lrd.IsValid() != nil {
				return fmt.Errorf("SaveDraftGroupLessonReport().ByGroupLessonReportGRPCMessage(): invalid lesson report detail: %v", err)
			}
			details = append(details, lrd)
		}
		l.Details = append(l.Details, details...)
		l.IsUpdateMembersInfo = true

		return nil
	}
}
