package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LessonReportModifierService struct {
	DB database.Ext

	PartnerFormConfigRepo interface {
		FindByPartnerAndFeatureName(ctx context.Context, db database.Ext, partnerID pgtype.Int4, featureName pgtype.Text) (*entities.PartnerFormConfig, error)
		FindByFeatureName(ctx context.Context, db database.Ext, featureName pgtype.Text) (*entities.PartnerFormConfig, error)
	}
	LessonRepo interface {
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
		GetLearnerIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error)
	}
	LessonReportRepo interface {
		Create(ctx context.Context, db database.Ext, report *entities.LessonReport) (*entities.LessonReport, error)
		Update(ctx context.Context, db database.Ext, report *entities.LessonReport) (*entities.LessonReport, error)
		UpdateLessonReportSubmittingStatusByID(ctx context.Context, db database.QueryExecer, id pgtype.Text, status entities.ReportSubmittingStatus) error
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.LessonReport, error)
		Delete(ctx context.Context, db database.Ext, id pgtype.Text) error
		FindByLessonID(ctx context.Context, db database.Ext, lessonID pgtype.Text) (*entities.LessonReport, error)
	}
	LessonReportDetailRepo interface {
		GetByLessonReportID(ctx context.Context, db database.Ext, lessonReportID pgtype.Text) (entities.LessonReportDetails, error)
		Upsert(ctx context.Context, db database.Ext, lessonReportID pgtype.Text, details entities.LessonReportDetails) error
		UpsertFieldValues(ctx context.Context, db database.Ext, values []*entities.PartnerDynamicFormFieldValue) error
		DeleteByLessonReportID(ctx context.Context, db database.Ext, lessonReportID pgtype.Text) error
		DeleteFieldValuesByDetails(ctx context.Context, db database.Ext, detailIDs pgtype.TextArray) error
		GetFieldValuesByDetailIDs(ctx context.Context, db database.Ext, detailIDs pgtype.TextArray) (entities.PartnerDynamicFormFieldValues, error)
	}
	LessonMemberRepo interface {
		GetLessonMembersInLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (entities.LessonMembers, error)
		UpdateLessonMembersFields(ctx context.Context, db database.QueryExecer, e []*entities.LessonMember, updateFields entities.UpdateLessonMemberFields) error
	}
	TeacherRepo interface {
		FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entities.Teacher, error)
	}
	LessonReportApprovalRecordRepo interface {
		Create(ctx context.Context, db database.QueryExecer, l *entities.LessonReportApprovalRecord) error
	}

	UpdateLessonSchedulingStatus func(ctx context.Context, req *lpb.UpdateLessonSchedulingStatusRequest) (*lpb.UpdateLessonSchedulingStatusResponse, error)
	Env                          string
	UnleashClientIns             unleashclient.ClientInstance
}

func (l *LessonReportModifierService) CreateIndividualLessonReport(ctx context.Context, req *bpb.CreateIndividualLessonReportRequest) (*bpb.CreateIndividualLessonReportResponse, error) {
	ctxzap.Extract(ctx).Debug("create report",
		zap.String("start time", req.StartTime.String()),
		zap.String("end time", req.EndTime.String()),
		zap.String("teacher", strings.Join(req.TeacherIds, ", ")),
	)

	for _, detail := range req.ReportDetail {
		ctxzap.Extract(ctx).Debug("create report",
			zap.String("StudentId ", detail.StudentId),
			zap.String("CourseId ", detail.CourseId),
			zap.String("AttendanceStatus ", detail.AttendanceStatus.String()),
			zap.String("AttendanceStatus enum ", detail.AttendanceStatus.Enum().String()),
			zap.String("AttendanceRemark ", detail.AttendanceRemark),
		)
		for _, dField := range detail.FieldValues {
			ctxzap.Extract(ctx).Debug("create report",
				zap.String("FieldID ", dField.DynamicFieldId),
				zap.String("ValueType ", dField.ValueType.String()),
			)
			if dField.ValueType == bpb.ValueType_VALUE_TYPE_INT {
				ctxzap.Extract(ctx).Debug("create report",
					zap.String("FieldValue ", fmt.Sprintf("%d", dField.GetIntValue())))
			}
			if dField.ValueType == bpb.ValueType_VALUE_TYPE_STRING {
				ctxzap.Extract(ctx).Debug("create report",
					zap.String("FieldValue ", dField.GetStringValue()))
			}
		}
	}

	return &bpb.CreateIndividualLessonReportResponse{Id: ""}, nil
}

func (l *LessonReportModifierService) SubmitLessonReport(ctx context.Context, req *bpb.WriteLessonReportRequest) (*bpb.SubmitLessonReportResponse, error) {
	lessonReport, err := NewLessonReport(ByLessonReportGRPCMessage(req))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	lessonReport.PartnerFormConfigRepo = l.PartnerFormConfigRepo
	lessonReport.LessonRepo = l.LessonRepo
	lessonReport.LessonReportRepo = l.LessonReportRepo
	lessonReport.LessonReportDetailRepo = l.LessonReportDetailRepo
	lessonReport.LessonMemberRepo = l.LessonMemberRepo
	lessonReport.TeacherRepo = l.TeacherRepo

	isAttendanceUnleashToggled, err := l.UnleashClientIns.IsFeatureEnabled("Lesson_LessonManagement_BackOffice_ValidationLessonBeforeCompleted", l.Env)
	if err != nil {
		return nil, fmt.Errorf("l.connectToUnleash: %w", err)
	}

	if err := lessonReport.Submit(ctx, l.DB); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("could not submit lesson report: %v", err))
	}
	isUnleashToggled, err := l.UnleashClientIns.IsFeatureEnabled("BACKEND_Lesson_HandleUpdateStatusWhenUserSubmitLessonReport", l.Env)
	if err != nil {
		return nil, fmt.Errorf("l.connectToUnleash: %w", err)
	}

	if isUnleashToggled {
		publishedSchedulingStatuses := []entities.SchedulingStatus{}
		schedulingStatus := entities.SchedulingStatus(lessonReport.Lesson.SchedulingStatus.String)
		isAttendance := lessonReport.Details.CheckStudentsAttendance()

		if schedulingStatus == entities.LessonSchedulingStatusDraft {
			publishedSchedulingStatuses = []entities.SchedulingStatus{entities.LessonSchedulingStatusPublished, entities.LessonSchedulingStatusCompleted}
			if isAttendanceUnleashToggled && !isAttendance {
				publishedSchedulingStatuses = []entities.SchedulingStatus{entities.LessonSchedulingStatusPublished}
			}
		} else if schedulingStatus == entities.LessonSchedulingStatusPublished {
			if !isAttendanceUnleashToggled || (isAttendanceUnleashToggled && isAttendance) {
				publishedSchedulingStatuses = []entities.SchedulingStatus{entities.LessonSchedulingStatusCompleted}
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

	return &bpb.SubmitLessonReportResponse{
		LessonReportId: lessonReport.LessonReportID,
	}, nil
}

func (l *LessonReportModifierService) SaveDraftLessonReport(ctx context.Context, req *bpb.WriteLessonReportRequest) (*bpb.SaveDraftLessonReportResponse, error) {
	lessonReport, err := NewLessonReport(ByLessonReportGRPCMessage(req))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	lessonReport.PartnerFormConfigRepo = l.PartnerFormConfigRepo
	lessonReport.LessonRepo = l.LessonRepo
	lessonReport.LessonReportRepo = l.LessonReportRepo
	lessonReport.LessonReportDetailRepo = l.LessonReportDetailRepo
	lessonReport.LessonMemberRepo = l.LessonMemberRepo
	lessonReport.TeacherRepo = l.TeacherRepo

	if err := lessonReport.SaveDraft(ctx, l.DB); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("could not save draft lesson report: %v", err))
	}

	return &bpb.SaveDraftLessonReportResponse{
		LessonReportId: lessonReport.LessonReportID,
	}, nil
}
