package constant

// Lesson
type (
	LessonStatus           string
	LessonType             string
	LessonTeachingMethod   string
	LessonTeachingMedium   string
	LessonSchedulingStatus string
)

const (
	LessonStatusNone       LessonStatus = "LESSON_STATUS_NONE"
	LessonStatusCompleted  LessonStatus = "LESSON_STATUS_COMPLETED"
	LessonStatusInProgress LessonStatus = "LESSON_STATUS_IN_PROGRESS"
	LessonStatusNotStarted LessonStatus = "LESSON_STATUS_NOT_STARTED"
	LessonStatusDraft      LessonStatus = "LESSON_STATUS_DRAFT"

	LessonTypeNone    LessonType = "LESSON_TYPE_NONE"
	LessonTypeOnline  LessonType = "LESSON_TYPE_ONLINE"
	LessonTypeOffline LessonType = "LESSON_TYPE_OFFLINE"
	LessonTypeHybrid  LessonType = "LESSON_TYPE_HYBRID"

	LessonTeachingMethodIndividual LessonTeachingMethod = "LESSON_TEACHING_METHOD_INDIVIDUAL"
	LessonTeachingMethodGroup      LessonTeachingMethod = "LESSON_TEACHING_METHOD_GROUP"

	LessonTeachingMediumOffline LessonTeachingMedium = "LESSON_TEACHING_MEDIUM_OFFLINE"
	LessonTeachingMediumOnline  LessonTeachingMedium = "LESSON_TEACHING_MEDIUM_ONLINE"

	LessonSchedulingStatusPublished LessonSchedulingStatus = "LESSON_SCHEDULING_STATUS_PUBLISHED"
	LessonSchedulingStatusDraft     LessonSchedulingStatus = "LESSON_SCHEDULING_STATUS_DRAFT"
	LessonSchedulingStatusCompleted LessonSchedulingStatus = "LESSON_SCHEDULING_STATUS_COMPLETED"
	LessonSchedulingStatusCanceled  LessonSchedulingStatus = "LESSON_SCHEDULING_STATUS_CANCELED"
)

// Lesson Report
type (
	ReportType              string
	ReportSubmittingStatus  string
	StudentAttendStatus     string
	DomainType              string
	StudentAttendanceNotice string
	StudentAttendanceReason string
)

const (
	ReportTypeIndividual ReportType = "LESSON_REPORT_INDIVIDUAL"
	ReportTypeGroup      ReportType = "LESSON_REPORT_GROUP"

	ReportSubmittingStatusSaved     ReportSubmittingStatus = "LESSON_REPORT_SUBMITTING_STATUS_SAVED"
	ReportSubmittingStatusSubmitted ReportSubmittingStatus = "LESSON_REPORT_SUBMITTING_STATUS_SUBMITTED"
	ReportSubmittingStatusApproved  ReportSubmittingStatus = "LESSON_REPORT_SUBMITTING_STATUS_APPROVED"

	StudentAttendStatusEmpty      StudentAttendStatus = "STUDENT_ATTEND_STATUS_EMPTY"
	StudentAttendStatusAttend     StudentAttendStatus = "STUDENT_ATTEND_STATUS_ATTEND"
	StudentAttendStatusAbsent     StudentAttendStatus = "STUDENT_ATTEND_STATUS_ABSENT"
	StudentAttendStatusLate       StudentAttendStatus = "STUDENT_ATTEND_STATUS_LATE"
	StudentAttendStatusLeaveEarly StudentAttendStatus = "STUDENT_ATTEND_STATUS_LEAVE_EARLY"
	StudentAttendStatusReallocate StudentAttendStatus = "STUDENT_ATTEND_STATUS_REALLOCATE"

	DomainTypeBo      DomainType = "DOMAIN_TYPE_BO"
	DomainTypeTeacher DomainType = "DOMAIN_TYPE_TEACHER"
	DomainTypeLearner DomainType = "DOMAIN_TYPE_LEARNER"

	StudentAttendanceNoticeEmpty     StudentAttendanceNotice = "NOTICE_EMPTY"
	StudentAttendanceNoticeInAdvance StudentAttendanceNotice = "IN_ADVANCE"
	StudentAttendanceNoticeOnTheDay  StudentAttendanceNotice = "ON_THE_DAY"
	StudentAttendanceNoticeNoContact StudentAttendanceNotice = "NO_CONTACT"

	StudentAttendanceReasonEmpty             StudentAttendanceReason = "REASON_EMPTY"
	StudentAttendanceReasonPhysicalCondition StudentAttendanceReason = "PHYSICAL_CONDITION"
	StudentAttendanceReasonSchoolEvent       StudentAttendanceReason = "SCHOOL_EVENT"
	StudentAttendanceReasonFamilyReason      StudentAttendanceReason = "FAMILY_REASON"
	StudentAttendanceReasonOther             StudentAttendanceReason = "REASON_OTHER"

	ReportReviewPermission = "lesson.report.review"
	LessonWritePermission  = "lesson.lesson.write"
)

// unleash name
const (
	PermissionToSubmitReport      string = "Lesson_LessonManagement_BackOffice_ReportReviewPermission"
	OptimisticLockingLessonReport string = "Lesson_LessonManagement_BackOffice_OptimisticLockingLessonReport"
)

// Dynamic Field
type FieldValueType string
type SystemDefinedField string
type FeatureName string

const (
	FieldValueTypeInt         FieldValueType = "VALUE_TYPE_INT"
	FieldValueTypeString      FieldValueType = "VALUE_TYPE_STRING"
	FieldValueTypeBool        FieldValueType = "VALUE_TYPE_BOOL"
	FieldValueTypeIntArray    FieldValueType = "VALUE_TYPE_INT_ARRAY"
	FieldValueTypeStringArray FieldValueType = "VALUE_TYPE_STRING_ARRAY"
	FieldValueTypeIntSet      FieldValueType = "VALUE_TYPE_INT_SET"
	FieldValueTypeStringSet   FieldValueType = "VALUE_TYPE_STRING_SET"
	FieldValueTypeNull        FieldValueType = "VALUE_TYPE_NULL"

	SystemDefinedFieldAttendanceStatus SystemDefinedField = "attendance_status"
	SystemDefinedFieldAttendanceRemark SystemDefinedField = "attendance_remark"
	SystemDefinedFiledAttendanceReason SystemDefinedField = "attendance_reason"
	SystemDefinedFiledAttendanceNotice SystemDefinedField = "attendance_notice"
	SystemDefinedFiledAttendanceNote   SystemDefinedField = "attendance_note"

	FeatureNameIndividualLessonReport FeatureName = "FEATURE_NAME_INDIVIDUAL_LESSON_REPORT"
	FeatureNameGroupLessonReport      FeatureName = "FEATURE_NAME_GROUP_LESSON_REPORT"
	// new value for individual lesson report
	FeatureNameIndividualUpdateLessonReport FeatureName = "FORM_CONFIG_LESSON_REPORT_IND_UPDATE"
)
