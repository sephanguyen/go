package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"golang.org/x/exp/slices"
)

type LessonReport struct {
	LessonReportID   string
	LessonID         string
	FeatureName      string
	Lesson           *entities.Lesson
	SubmittingStatus entities.ReportSubmittingStatus
	FormConfig       *FormConfig
	Details          LessonReportDetails

	PartnerFormConfigRepo interface {
		FindByFeatureName(ctx context.Context, db database.Ext, featureName pgtype.Text) (*entities.PartnerFormConfig, error)
		FindByPartnerAndFeatureName(ctx context.Context, db database.Ext, partnerID pgtype.Int4, featureName pgtype.Text) (*entities.PartnerFormConfig, error)
	}
	LessonRepo interface {
		FindByID(ctx context.Context, db database.Ext, id pgtype.Text) (*entities.Lesson, error)
		GetLearnerIDsOfLesson(ctx context.Context, db database.QueryExecer, lessonID pgtype.Text) (pgtype.TextArray, error)
	}
	LessonReportRepo interface {
		Create(ctx context.Context, db database.Ext, report *entities.LessonReport) (*entities.LessonReport, error)
		Update(ctx context.Context, db database.Ext, report *entities.LessonReport) (*entities.LessonReport, error)
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

	UpdateLessonSchedulingStatus func(ctx context.Context, req *lpb.UpdateLessonSchedulingStatusRequest) (*lpb.UpdateLessonSchedulingStatusResponse, error)
}

type NewLessonReportOption func(*LessonReport) error

// ByLessonReportGRPCMessage will create lesson report object by LessonReport GRPC message
func ByLessonReportGRPCMessage(req *bpb.WriteLessonReportRequest) NewLessonReportOption {
	return func(l *LessonReport) (err error) {
		l.LessonReportID = req.LessonReportId
		l.LessonID = req.LessonId
		l.FeatureName = req.FeatureName
		details := make([]*LessonReportDetail, 0, len(l.Details))
		for _, detail := range req.Details {
			lrd := &LessonReportDetail{
				StudentID:        detail.StudentId,
				AttendanceStatus: entities.StudentAttendStatus(detail.AttendanceStatus.String()),
				AttendanceRemark: detail.AttendanceRemark,
				AttendanceNotice: entities.StudentAttendanceNotice(detail.AttendanceNotice.String()),
				AttendanceReason: entities.StudentAttendanceReason(detail.AttendanceReason.String()),
				AttendanceNote:   detail.AttendanceNote,
			}
			lrd.Fields, err = LessonReportFieldsFromDynamicFieldValueGRPC(detail.FieldValues...)
			if err != nil {
				return fmt.Errorf("got error when parse DynamicFieldValue GRPC message to LessonReportFields: %v", err)
			}
			details = append(details, lrd)
		}
		l.Details = append(l.Details, details...)

		return nil
	}
}

func NewLessonReport(opts ...NewLessonReportOption) (*LessonReport, error) {
	lessonRp := &LessonReport{}
	for _, opt := range opts {
		if err := opt(lessonRp); err != nil {
			return nil, err
		}
	}
	return lessonRp, nil
}

func (l *LessonReport) ToLessonReportEntity() (*entities.LessonReport, error) {
	now := time.Now()
	e := &entities.LessonReport{}
	database.AllNullEntity(e)
	if err := multierr.Combine(
		e.LessonID.Set(l.LessonID),
		e.ReportSubmittingStatus.Set(l.SubmittingStatus),
		e.FormConfigID.Set(l.FormConfig.FormConfigID),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return nil, err
	}

	lessonRpID := l.LessonReportID
	if len(lessonRpID) == 0 {
		lessonRpID = idutil.ULIDNow()
	}
	if err := e.LessonReportID.Set(lessonRpID); err != nil {
		return nil, err
	}

	return e, nil
}

func (l *LessonReport) IsValid(ctx context.Context, db database.Ext, isDraft bool) error {
	if len(l.LessonID) == 0 {
		return fmt.Errorf("lesson_id could not be empty")
	}

	if len(l.SubmittingStatus) == 0 {
		return fmt.Errorf("submitting_status could not be empty")
	}

	// check details of student which belong to lesson
	learnerIDs, err := l.LessonRepo.GetLearnerIDsOfLesson(ctx, db, database.Text(l.LessonID))
	if err != nil {
		return fmt.Errorf("LessonRepo.GetLearnerIDsOfLesson: %v", err)
	}

	if err = l.Details.OnlyHaveLearnerIDs(database.FromTextArray(learnerIDs)); err != nil {
		return err
	}

	allowFields := make(map[string]*FormConfigField)
	if l.FormConfig != nil {
		if err := l.FormConfig.IsValid(); err != nil {
			return fmt.Errorf("invalid form_config: %v", err)
		}
		// get field ids from form config
		allowFields = l.FormConfig.GetFieldsMap()
	}
	if err = l.Details.OnlyHaveAllowFields(allowFields); err != nil {
		return err
	}

	if err = l.Details.IsValid(); err != nil {
		return fmt.Errorf("invalid details: %v", err)
	}

	if !isDraft {
		// check required field's values of each details
		requiredFields := make(map[string]*FormConfigField)
		if l.FormConfig != nil {
			requiredFields = l.FormConfig.GetRequiredFieldsMap()
		}
		if err = l.Details.ValidateRequiredFieldsValue(requiredFields); err != nil {
			return err
		}
	}

	return nil
}

// Normalize will normalize fields in LessonReport struct and also fill missing fields
//   - Set default value for SubmittingStatus if it empties
//   - Fill FormConfig's data if it empties
//   - Remove lesson details which has StudentID not belong to lesson
//   - Normalize Details
func (l *LessonReport) Normalize(ctx context.Context, db database.Ext) error {
	if len(l.SubmittingStatus) == 0 {
		l.SubmittingStatus = entities.ReportSubmittingStatusSaved
	}

	// check lesson id
	lesson, err := l.LessonRepo.FindByID(ctx, db, database.Text(l.LessonID))
	if err != nil {
		return fmt.Errorf("LessonRepo.FindByID: %v", err)
	}

	l.Lesson = lesson

	// fill form config
	if l.FormConfig == nil {
		if err := l.getFormConfig(ctx, db, lesson); err != nil {
			return fmt.Errorf("could not get form config: %v", err)
		}
	}

	// normalize details
	l.Details.Normalize()

	return nil
}

func (l *LessonReport) Submit(ctx context.Context, db database.Ext) error {
	l.SubmittingStatus = entities.ReportSubmittingStatusSubmitted
	if err := l.preStoreToDB(ctx, db); err != nil {
		return fmt.Errorf("preStore to DB: %s", err)
	}

	if err := l.storeToDB(ctx, db); err != nil {
		return fmt.Errorf("store to DB: %s", err)
	}

	return nil
}

func (l *LessonReport) SaveDraft(ctx context.Context, db database.Ext) error {
	l.SubmittingStatus = entities.ReportSubmittingStatusSaved
	if err := l.preStoreToDB(ctx, db); err != nil {
		return fmt.Errorf("preStore to DB: %s", err)
	}

	if err := l.storeToDB(ctx, db); err != nil {
		return fmt.Errorf("store to DB: %s", err)
	}

	return nil
}

func (l *LessonReport) preStoreToDB(ctx context.Context, db database.Ext) error {
	isDraft := l.SubmittingStatus == entities.ReportSubmittingStatusSaved
	// when update current lesson report, we need
	// check existence of id and didn't be changed lesson id
	if len(l.LessonReportID) != 0 {
		currentReport, err := l.LessonReportRepo.FindByID(ctx, db, database.Text(l.LessonReportID))
		if err != nil {
			return fmt.Errorf("LessonReportRepo.FindByID: %v", err)
		}
		l.LessonID = currentReport.LessonID.String
	} else {
		// if there is no lesson report id (creating), check there is any current lesson report
		currentLessonReport, err := l.LessonReportRepo.FindByLessonID(ctx, db, database.Text(l.LessonID))
		if err == nil {
			if isDraft {
				// if ReportSubmittingStatus == save draft, assign report_id
				l.LessonReportID = currentLessonReport.LessonReportID.String
				l.LessonID = currentLessonReport.LessonID.String
			} else {
				// if ReportSubmittingStatus == submit, don't assign report_id
				return fmt.Errorf("each lesson must only have a lesson report")
			}
		} else if err != nil && err.Error() != fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows).Error() {
			return fmt.Errorf("LessonReportRepo.FindByLessonID: %v", err)
		}
	}

	if err := l.Normalize(ctx, db); err != nil {
		return err
	}

	if err := l.IsValid(ctx, db, isDraft); err != nil {
		return fmt.Errorf("LessonReport.IsValid: %v", err)
	}

	return nil
}

func (l *LessonReport) storeToDB(ctx context.Context, db database.Ext) error {
	isCreate := false
	if len(l.LessonReportID) == 0 {
		isCreate = true
	}

	e, err := l.ToLessonReportEntity()
	if err != nil {
		return fmt.Errorf("could not convert to lesson report entity: %v", err)
	}

	if isCreate {
		// insert lesson report record
		e, err = l.LessonReportRepo.Create(ctx, db, e)
		if err != nil {
			return fmt.Errorf("LessonReportRepo.Create: %v", err)
		}
		l.LessonReportID = e.LessonReportID.String
	} else {
		_, err = l.LessonReportRepo.Update(ctx, db, e)
		if err != nil {
			return fmt.Errorf("LessonReportRepo.Update: %v", err)
		}
	}

	// upsert lesson report details record
	details, err := l.Details.ToLessonReportDetailsEntity(l.LessonReportID)
	if err != nil {
		return fmt.Errorf("could not convert to lesson report details entity: %v", err)
	}
	if err = l.LessonReportDetailRepo.Upsert(ctx, db, database.Text(l.LessonReportID), details); err != nil {
		return fmt.Errorf("LessonReportDetailRepo.Upsert: %v", err)
	}

	// get lesson report detail ids
	details, err = l.LessonReportDetailRepo.GetByLessonReportID(ctx, db, database.Text(l.LessonReportID))
	if err != nil {
		return fmt.Errorf("LessonReportDetailRepo.GetByLessonReportID: %v", err)
	}

	detailByStudentID := make(map[string]*entities.LessonReportDetail)
	for i := range details {
		detailByStudentID[details[i].StudentID.String] = details[i]
	}

	// upsert field values for each lesson report detail
	var fieldValues []*entities.PartnerDynamicFormFieldValue
	for _, detail := range l.Details {
		id := detailByStudentID[detail.StudentID].LessonReportDetailID.String
		e, err := detail.Fields.ToPartnerDynamicFormFieldValueEntities(id)
		if err != nil {
			return fmt.Errorf("could not convert to partner dynamic form field value: %v", err)
		}
		fieldValues = append(fieldValues, e...)
	}
	if err = l.LessonReportDetailRepo.UpsertFieldValues(ctx, db, fieldValues); err != nil {
		return fmt.Errorf("LessonReportDetailRepo.UpsertFieldValues: %v", err)
	}

	// when lesson is locked, the report can not change the attendance status of lesson member
	if !l.Lesson.IsLocked.Bool {
		// upsert lesson member's fields
		members, err := l.Details.ToLessonMembersEntity(l.LessonID)
		if err != nil {
			return fmt.Errorf("could not convert to lesson member: %v", err)
		}
		if err = l.LessonMemberRepo.UpdateLessonMembersFields(
			ctx,
			db,
			members,
			entities.UpdateLessonMemberFields{
				entities.LessonMemberAttendanceRemark,
				entities.LessonMemberAttendanceStatus,
				entities.LessonMemberAttendanceNotice,
				entities.LessonMemberAttendanceReason,
				entities.LessonMemberAttendanceNote,
			},
		); err != nil {
			return fmt.Errorf("LessonMemberRepo.UpdateLessonMembersFields: %v", err)
		}
	}

	return nil
}

func (l *LessonReport) getFormConfig(ctx context.Context, db database.Ext, lesson *entities.Lesson) error {
	// get school id
	teacher, err := l.TeacherRepo.FindByID(ctx, db, lesson.TeacherID)
	if err != nil {
		return fmt.Errorf("TeacherRepo.FindByID: %v", err)
	}

	var feature pgtype.Text
	switch lesson.TeachingMethod.String {
	case string(entities.LessonTeachingMethodGroup):
		feature = database.Text(string(entities.FeatureNameGroupLessonReport))
	default:
		feature = database.Text(string(entities.FeatureNameIndividualLessonReport))
	}
	if len(l.FeatureName) > 0 {
		feature = database.Text(l.FeatureName)
	}
	cf, err := l.PartnerFormConfigRepo.FindByPartnerAndFeatureName(ctx, db, teacher.SchoolIDs.Elements[0], feature)
	if err != nil {
		return fmt.Errorf("PartnerFormConfigRepo.FindByPartnerAndFeatureName: %v", err)
	}

	l.FormConfig, err = NewFormConfigByPartnerFormConfig(cf)
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonReport) Delete(ctx context.Context, db database.Ext) (err error) {
	switch db.(type) {
	case pgx.Tx:
		err = l.delete(ctx, db)
	default:
		err = database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
			if err := l.delete(ctx, tx); err != nil {
				return err
			}
			return nil
		})
	}
	if err != nil {
		return err
	}

	return nil
}

func (l *LessonReport) delete(ctx context.Context, db database.Ext) error {
	if err := l.GetData(ctx, db, IgnoreFormConfig(), IgnoreDetailFieldValues()); err != nil {
		return err
	}

	// remove lesson report
	if err := l.LessonReportRepo.Delete(ctx, db, database.Text(l.LessonReportID)); err != nil {
		return fmt.Errorf("LessonReportRepo.Delete: %v", err)
	}

	// remove lesson report detail
	if l.Details != nil && len(l.Details) > 0 {
		if err := l.LessonReportDetailRepo.DeleteByLessonReportID(ctx, db, database.Text(l.LessonReportID)); err != nil {
			return fmt.Errorf("LessonReportDetailRepo.DeleteByLessonReport: %v", err)
		}

		// remove record in partner_form_dynamic_field_value table
		if err := l.LessonReportDetailRepo.DeleteFieldValuesByDetails(ctx, db, database.TextArray(l.Details.ReportDetailIDs())); err != nil {
			return fmt.Errorf("LessonReportDetailRepo.DeleteFieldValuesByDetails: %v", err)
		}

		// remove attendance_status and attendance_remark in lesson_member
		members := make(entities.LessonMembers, 0, len(l.Details))
		for _, detail := range l.Details {
			members = append(members, &entities.LessonMember{
				LessonID:         database.Text(l.LessonID),
				UserID:           database.Text(detail.StudentID),
				AttendanceStatus: database.Text(string(entities.StudentAttendStatusEmpty)),
				AttendanceRemark: database.Text(""),
			})
		}

		if err := l.LessonMemberRepo.UpdateLessonMembersFields(
			ctx,
			db,
			members,
			entities.UpdateLessonMemberFields{
				entities.LessonMemberAttendanceRemark,
				entities.LessonMemberAttendanceStatus,
			},
		); err != nil {
			return fmt.Errorf("LessonMemberRepo.UpdateLessonMembersFields: %v", err)
		}
	}

	return nil
}

type ignoreFieldsOption struct {
	FormConfig  bool
	Details     bool
	FieldValues bool
}
type GetDataOption func(*ignoreFieldsOption)

func IgnoreFormConfig() GetDataOption {
	return func(i *ignoreFieldsOption) {
		i.FormConfig = true
	}
}

func IgnoreDetails() GetDataOption {
	return func(i *ignoreFieldsOption) {
		i.Details = true
	}
}

func IgnoreDetailFieldValues() GetDataOption {
	return func(i *ignoreFieldsOption) {
		i.FieldValues = true
	}
}

// GetData will get data of a lesson include: LessonID, SubmittingStatus, FormConfig, Details.
// To ignore some properties, use options below:
//   - IgnoreFormConfig: not get form config.
//   - IgnoreDetails: not get details.
//   - IgnoreDetailFieldValues: not get field's values of each detail (available only when
//
// not ignore details).
func (l *LessonReport) GetData(ctx context.Context, db database.Ext, opts ...GetDataOption) error {
	report, err := l.LessonReportRepo.FindByID(ctx, db, database.Text(l.LessonReportID))
	if err != nil {
		return fmt.Errorf("LessonReportRepo.FindByID: %v", err)
	}

	l.LessonID = report.LessonID.String
	l.SubmittingStatus = entities.ReportSubmittingStatus(report.ReportSubmittingStatus.String)

	// default: get all
	ign := &ignoreFieldsOption{}
	for _, opt := range opts {
		opt(ign)
	}

	if !ign.FormConfig {
		lesson, err := l.LessonRepo.FindByID(ctx, db, database.Text(l.LessonID))
		if err != nil {
			return fmt.Errorf("LessonRepo.FindByID: %v", err)
		}

		if err = l.getFormConfig(ctx, db, lesson); err != nil {
			return nil
		}
	}

	if !ign.Details {
		details, err := l.LessonReportDetailRepo.GetByLessonReportID(ctx, db, database.Text(l.LessonReportID))
		if err != nil {
			return fmt.Errorf("LessonReportDetailRepo.GetByLessonReportID: %v", err)
		}

		members, err := l.LessonMemberRepo.GetLessonMembersInLesson(ctx, db, database.Text(l.LessonID))
		if err != nil {
			return fmt.Errorf("LessonMemberRepo.GetLessonMembersInLesson: %v", err)
		}
		l.Details = LessonReportDetailsFormEntity(details, members)

		if !ign.FieldValues {
			values, err := l.LessonReportDetailRepo.GetFieldValuesByDetailIDs(ctx, db, details.ReportDetailIDs())
			if err != nil {
				return fmt.Errorf("LessonReportDetailRepo.GetFieldValuesByDetailIDs: %v", err)
			}

			fieldByDetailID := make(map[string]LessonReportFields)
			for _, v := range values {
				field := &LessonReportField{
					FieldID: v.FieldID.String,
					Value: &AttributeValue{
						Int:         int(v.IntValue.Int),
						String:      v.StringValue.String,
						Bool:        v.BoolValue.Bool,
						IntArray:    database.Int4ArrayToIntArray(v.IntArrayValue),
						StringArray: database.FromTextArray(v.StringArrayValue),
						IntSet:      database.Int4ArrayToIntArray(v.IntSetValue),
						StringSet:   database.FromTextArray(v.StringSetValue),
					},
					FieldRenderGuide: nil,
				}
				if _, ok := fieldByDetailID[v.LessonReportDetailID.String]; ok {
					fieldByDetailID[v.LessonReportDetailID.String] = append(fieldByDetailID[v.LessonReportDetailID.String], field)
				} else {
					fieldByDetailID[v.LessonReportDetailID.String] = []*LessonReportField{field}
				}
			}
			l.Details.AddFieldValues(fieldByDetailID)
		}
	}

	return nil
}

type LessonReportDetail struct {
	ReportDetailID   string
	StudentID        string
	AttendanceStatus entities.StudentAttendStatus
	AttendanceRemark string
	Fields           LessonReportFields
	AttendanceNotice entities.StudentAttendanceNotice
	AttendanceReason entities.StudentAttendanceReason
	AttendanceNote   string
}

func (l *LessonReportDetail) IsValid() error {
	if len(l.StudentID) == 0 {
		return fmt.Errorf("student_id could not be empty")
	}

	if err := l.Fields.IsValid(); err != nil {
		return fmt.Errorf("invalid fields: %v", err)
	}

	return nil
}

type LessonReportDetails []*LessonReportDetail

func LessonReportDetailsFormEntity(details entities.LessonReportDetails, members entities.LessonMembers) LessonReportDetails {
	membersByID := make(map[string]*entities.LessonMember)
	for i := range members {
		membersByID[members[i].UserID.String] = members[i]
	}

	res := make(LessonReportDetails, 0, len(details))
	for _, detail := range details {
		member, ok := membersByID[detail.StudentID.String]
		v := &LessonReportDetail{
			ReportDetailID: detail.LessonReportDetailID.String,
			StudentID:      detail.StudentID.String,
		}
		if ok {
			v.AttendanceRemark = member.AttendanceRemark.String
			v.AttendanceStatus = entities.StudentAttendStatus(member.AttendanceStatus.String)
		}
		res = append(res, v)
	}

	return res
}

func (ls LessonReportDetails) AddFieldValues(fields map[string]LessonReportFields) {
	for i := range ls {
		if v, ok := fields[ls[i].ReportDetailID]; ok {
			ls[i].Fields = v
		}
	}
}

func (ls LessonReportDetails) CheckStudentsAttendance() bool {
	for _, l := range ls {
		if len(l.AttendanceStatus) == 0 || l.AttendanceStatus == entities.StudentAttendStatusEmpty {
			return false
		}
	}
	return true
}

func (ls LessonReportDetails) IsValid() error {
	studentIDs := make(map[string]bool)
	for _, l := range ls {
		if err := l.IsValid(); err != nil {
			return err
		}

		if _, ok := studentIDs[l.StudentID]; ok {
			return fmt.Errorf("lesson report detail's student id %s be duplicated", l.StudentID)
		}
		studentIDs[l.StudentID] = true
	}

	return nil
}

func (ls LessonReportDetails) OnlyHaveLearnerIDs(learnerIDs []string) error {
	learnerIDsMap := make(map[string]bool)
	for _, id := range learnerIDs {
		learnerIDsMap[id] = true
	}
	for _, detail := range ls {
		if _, ok := learnerIDsMap[detail.StudentID]; !ok {
			return fmt.Errorf("learner %s doesn't belong to lesson", detail.StudentID)
		}
	}

	return nil
}

func (ls LessonReportDetails) OnlyHaveAllowFields(allowFields map[string]*FormConfigField) error {
	// validate field id
	for _, detail := range ls {
		for _, field := range detail.Fields {
			// check field id exist or not in form config
			if _, ok := allowFields[field.FieldID]; !ok {
				return fmt.Errorf("field id %s of user %s not exist in form config", field.FieldID, detail.StudentID)
			}

			// check dynamic field not allow same system defined fields
			if slices.Contains(
				[]string{string(SystemDefinedFieldAttendanceStatus),
					string(SystemDefinedFieldAttendanceRemark),
					string(SystemDefinedFieldAttendanceNotice),
					string(SystemDefinedFieldAttendanceReason),
					string(SystemDefinedFieldAttendanceNote)},
				field.FieldID) {
				return fmt.Errorf("field id %s of user %s is not a dynamic field", field.FieldID, detail.StudentID)
			}
		}

		// check system defined fields exist or not in form config
		if _, ok := allowFields[string(SystemDefinedFieldAttendanceStatus)]; len(detail.AttendanceStatus) != 0 && !ok {
			return fmt.Errorf("system field id %s of user %s not exist in form config", SystemDefinedFieldAttendanceStatus, detail.StudentID)
		}
		if _, ok := allowFields[string(SystemDefinedFieldAttendanceRemark)]; len(detail.AttendanceRemark) != 0 && !ok {
			return fmt.Errorf("system field id %s of user %s not exist in form config", SystemDefinedFieldAttendanceRemark, detail.StudentID)
		}
		// if _, ok := allowFields[string(SystemDefinedFieldAttendanceNotice)]; len(detail.AttendanceNotice) != 0 && !ok {
		// 	return fmt.Errorf("field id %s of user %s not exist in form config", SystemDefinedFieldAttendanceNotice, detail.StudentID)
		// }
		// if _, ok := allowFields[string(SystemDefinedFieldAttendanceReason)]; len(detail.AttendanceReason) != 0 && !ok {
		// 	return fmt.Errorf("field id %s of user %s not exist in form config", SystemDefinedFieldAttendanceReason, detail.StudentID)
		// }
		// if _, ok := allowFields[string(SystemDefinedFieldAttendanceNote)]; len(detail.AttendanceNote) != 0 && !ok {
		// 	return fmt.Errorf("field id %s of user %s not exist in form config", SystemDefinedFieldAttendanceNote, detail.StudentID)
		// }
	}

	return nil
}

func (ls LessonReportDetails) ValidateRequiredFieldsValue(requiredFields map[string]*FormConfigField) error {
	for _, detail := range ls {
		inputFieldsByID := make(map[string]*LessonReportField)
		for i, field := range detail.Fields {
			inputFieldsByID[field.FieldID] = detail.Fields[i]
		}
		for id, requiredField := range requiredFields {
			// check system defined fields
			if id == string(SystemDefinedFieldAttendanceStatus) {
				if len(detail.AttendanceStatus) == 0 || detail.AttendanceStatus == entities.StudentAttendStatusEmpty {
					return fmt.Errorf("field %s is required", SystemDefinedFieldAttendanceStatus)
				}
				continue
			}
			if id == string(SystemDefinedFieldAttendanceRemark) {
				if len(detail.AttendanceRemark) == 0 {
					return fmt.Errorf("field %s is required", SystemDefinedFieldAttendanceRemark)
				}
				continue
			}
			if id == string(SystemDefinedFieldAttendanceNotice) {
				if len(detail.AttendanceNotice) == 0 {
					return fmt.Errorf("field %s is required", SystemDefinedFieldAttendanceNotice)
				}
				continue
			}
			if id == string(SystemDefinedFieldAttendanceReason) {
				if len(detail.AttendanceReason) == 0 {
					return fmt.Errorf("field %s is required", SystemDefinedFieldAttendanceReason)
				}
				continue
			}
			if id == string(SystemDefinedFieldAttendanceNote) {
				if len(detail.AttendanceNote) == 0 {
					return fmt.Errorf("field %s is required", SystemDefinedFieldAttendanceNote)
				}
				continue
			}

			v, ok := inputFieldsByID[id]
			if !ok || v.Value == nil {
				return fmt.Errorf("field %s is required", id)
			}
			switch requiredField.ValueType {
			case FieldValueTypeInt:
			case FieldValueTypeString:
				if len(v.Value.String) == 0 {
					return fmt.Errorf("field %s is required", id)
				}
			case FieldValueTypeBool:
			case FieldValueTypeIntArray:
				if len(v.Value.IntArray) == 0 {
					return fmt.Errorf("field %s is required", id)
				}
			case FieldValueTypeStringArray:
				if len(v.Value.StringArray) == 0 {
					return fmt.Errorf("field %s is required", id)
				}
			case FieldValueTypeIntSet:
				if len(v.Value.IntSet) == 0 {
					return fmt.Errorf("field %s is required", id)
				}
			case FieldValueTypeStringSet:
				if len(v.Value.StringSet) == 0 {
					return fmt.Errorf("field %s is required", id)
				}
			}
		}
	}

	return nil
}

func (ls LessonReportDetails) GetByStudentIDs(studentIDs []string) LessonReportDetails {
	if len(ls) == 0 {
		return nil
	}

	studentIDsMap := make(map[string]bool)
	for _, id := range studentIDs {
		studentIDsMap[id] = true
	}

	details := make(LessonReportDetails, 0, len(ls))
	for i, detail := range ls {
		if _, ok := studentIDsMap[detail.StudentID]; ok {
			details = append(details, ls[i])
			delete(studentIDsMap, detail.StudentID)
		}
	}

	return details
}

// Normalize will remove duplicated StudentID items and normalize Fields attribute
func (ls *LessonReportDetails) Normalize() {
	if len(*ls) == 0 {
		return
	}

	studentIDs := make([]string, 0, len(*ls))
	for _, detail := range *ls {
		studentIDs = append(studentIDs, detail.StudentID)
	}
	*ls = ls.GetByStudentIDs(studentIDs)

	for i := range *ls {
		(*ls)[i].Fields.Normalize()
	}
}

func (ls LessonReportDetails) ToLessonReportDetailsEntity(lessonReportID string) (entities.LessonReportDetails, error) {
	now := time.Now()
	res := make(entities.LessonReportDetails, 0, len(ls))
	for _, l := range ls {
		e := &entities.LessonReportDetail{}
		database.AllNullEntity(e)
		if err := multierr.Combine(
			e.LessonReportDetailID.Set(idutil.ULIDNow()),
			e.LessonReportID.Set(lessonReportID),
			e.StudentID.Set(l.StudentID),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		); err != nil {
			return nil, err
		}
		res = append(res, e)
	}

	return res, nil
}

func (ls LessonReportDetails) ToLessonMembersEntity(lessonID string) (entities.LessonMembers, error) {
	now := time.Now()
	res := make(entities.LessonMembers, 0, len(ls))
	for _, l := range ls {
		e := &entities.LessonMember{}
		database.AllNullEntity(e)
		if err := multierr.Combine(
			e.LessonID.Set(lessonID),
			e.UserID.Set(l.StudentID),
			e.AttendanceStatus.Set(l.AttendanceStatus),
			e.AttendanceRemark.Set(l.AttendanceRemark),
			e.AttendanceNotice.Set(l.AttendanceNotice),
			e.AttendanceReason.Set(l.AttendanceReason),
			e.AttendanceNote.Set(l.AttendanceNote),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
		); err != nil {
			return nil, err
		}
		res = append(res, e)
	}

	return res, nil
}

func (ls LessonReportDetails) ReportDetailIDs() []string {
	res := make([]string, 0, len(ls))
	for _, i := range ls {
		res = append(res, i.ReportDetailID)
	}

	return res
}

type LessonReportField struct {
	FieldID          string
	Value            *AttributeValue
	FieldRenderGuide []byte
	ValueType        string
}

func (l *LessonReportField) IsValid() error {
	if len(l.FieldID) == 0 {
		return fmt.Errorf("field_id could not be empty")
	}

	return nil
}

type LessonReportFields []*LessonReportField

func (ls LessonReportFields) IsValid() error {
	fieldIDs := make(map[string]bool)
	for _, l := range ls {
		if err := l.IsValid(); err != nil {
			return err
		}

		if _, ok := fieldIDs[l.FieldID]; ok {
			return fmt.Errorf("lesson report field's field id %s be duplicated", l.FieldID)
		}
		fieldIDs[l.FieldID] = true
	}

	return nil
}

func (ls LessonReportFields) GetFieldsByIDs(ids []string) LessonReportFields {
	idsMap := make(map[string]bool)
	for _, id := range ids {
		idsMap[id] = true
	}

	fields := make(LessonReportFields, 0, len(ids))
	for i, field := range ls {
		if _, ok := idsMap[field.FieldID]; ok {
			fields = append(fields, ls[i])
			delete(idsMap, field.FieldID)
		}
	}

	return fields
}

func (ls *LessonReportFields) Normalize() {
	if len(*ls) == 0 {
		return
	}

	fieldIDs := make(map[string]bool)
	notDuplicated := make(LessonReportFields, 0, len(*ls))
	for i, field := range *ls {
		if _, ok := fieldIDs[field.FieldID]; !ok {
			notDuplicated = append(notDuplicated, (*ls)[i])
			fieldIDs[field.FieldID] = true
		}
	}
	*ls = notDuplicated
}

func (ls LessonReportFields) ToPartnerDynamicFormFieldValueEntities(lessonReportDetailID string) ([]*entities.PartnerDynamicFormFieldValue, error) {
	now := time.Now()
	res := make([]*entities.PartnerDynamicFormFieldValue, 0, len(ls))
	for _, l := range ls {
		e := &entities.PartnerDynamicFormFieldValue{}
		database.AllNullEntity(e)
		if err := multierr.Combine(
			e.DynamicFormFieldValueID.Set(idutil.ULIDNow()),
			e.FieldID.Set(l.FieldID),
			e.LessonReportDetailID.Set(lessonReportDetailID),
			e.FieldRenderGuide.Set(l.FieldRenderGuide),
			e.IntValue.Set(l.Value.Int),
			e.StringValue.Set(l.Value.String),
			e.BoolValue.Set(l.Value.Bool),
			e.StringArrayValue.Set(l.Value.StringArray),
			e.IntArrayValue.Set(l.Value.IntArray),
			e.StringSetValue.Set(l.Value.StringSet),
			e.IntSetValue.Set(l.Value.IntSet),
			e.CreatedAt.Set(now),
			e.UpdatedAt.Set(now),
			e.ValueType.Set(l.ValueType),
		); err != nil {
			return nil, err
		}

		res = append(res, e)
	}

	return res, nil
}

func LessonReportFieldsFromDynamicFieldValueGRPC(fields ...*bpb.DynamicFieldValue) (LessonReportFields, error) {
	res := make(LessonReportFields, 0, len(fields))
	for _, field := range fields {
		value := &AttributeValue{}
		switch v := field.Value.(type) {
		case *bpb.DynamicFieldValue_IntValue:
			value.SetInt(int(v.IntValue))
		case *bpb.DynamicFieldValue_StringValue:
			value.SetString(v.StringValue)
		case *bpb.DynamicFieldValue_BoolValue:
			value.SetBool(v.BoolValue)
		case *bpb.DynamicFieldValue_IntArrayValue_:
			intArray := make([]int, 0, len(v.IntArrayValue.ArrayValue))
			for _, item := range v.IntArrayValue.GetArrayValue() {
				intArray = append(intArray, int(item))
			}
			value.SetIntArray(intArray)
		case *bpb.DynamicFieldValue_StringArrayValue_:
			value.SetStringArray(v.StringArrayValue.GetArrayValue())
		case *bpb.DynamicFieldValue_IntSetValue_:
			intSet := make([]int, 0, len(v.IntSetValue.ArrayValue))
			for _, item := range v.IntSetValue.GetArrayValue() {
				intSet = append(intSet, int(item))
			}
			value.SetIntSet(intSet)
		case *bpb.DynamicFieldValue_StringSetValue_:
			value.SetStringSet(v.StringSetValue.GetArrayValue())
		default:
			return nil, fmt.Errorf("unimplement handler for type %T", field.Value)
		}

		res = append(res, &LessonReportField{
			FieldID:          field.DynamicFieldId,
			FieldRenderGuide: field.FieldRenderGuide,
			Value:            value,
			ValueType:        field.ValueType.String(),
		})
	}

	return res, nil
}
