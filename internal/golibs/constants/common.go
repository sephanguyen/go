package constants

import (
	"math"
	"time"
)

const (
	ManabieCity     = math.MinInt32
	ManabieDistrict = math.MinInt32

	NatsMaxRedeliveryTimes   = uint32(10)
	JPREPMaxMessageInflight  = 3
	MaxRecordProcessPertime  = 100
	NatsJSMaxRedeliveryTimes = 10

	SubjectClassEventNats           = "class_event"
	QueueGroupSubjectClassEventNats = "queue_class_event"

	SubjectLessonEventNats                    = "lesson_event"
	QueueGroupSubjectLessonEventNats          = "queue_lesson_event"
	DurableGroupSubjectSubjectLessonEventNats = "durable_lesson_event"

	SubjectSyncLessonChat    = "sync_lesson_chat"
	QueueGroupSyncLessonChat = "queue_sync_lesson_chat"
	DurableSyncLessonChat    = "durable_sync_lesson_chat"

	StreamStudentPackageEventNats    = "studentpackage"
	StreamStudentPackageEventNatsV2  = "studentpackagev2"
	SubjectStudentPackageEventNats   = "StudentPackage.Upserted"
	SubjectStudentPackageV2EventNats = "StudentPackageV2.Upserted"
	QueueStudentPackageEventNats     = "queue-student-package"
	QueueStudentPackageEventNatsV2   = "queue-student-package-v2"
	DurableStudentPackageEventNats   = "durable-student-package"
	DurableStudentPackageEventNatsV2 = "durable-student-package-v2"
	DeliverStudentPackageEventNats   = "deliver.student-package"
	DeliverStudentPackageEventNatsV2 = "deliver.student-package-v2"

	SubjectUserEventNats = "subject_user_event"
	QueueUserEventNats   = "queue_user_event"
	DurableUserEventNats = "durable_user_event"

	QueueSyncStudentSubscriptionEventNats   = "queue-sync-student-subscription"
	DurableSyncStudentSubscriptionEventNats = "durable-sync-student-subscription"
	DeliverStudentSubscriptionEventNats     = "deliver.sync-student-subscription"

	QueueNotificationSyncStudentPackageEventNatsV2   = "queue-notification-sync-student-package-v2"
	DurableNotificationSyncStudentPackageEventNatsV2 = "durable-notification-sync-student-package-v2"
	DeliverNotificationSyncStudentPackageEventNatsV2 = "deliver.notification-sync-student-package-v2"

	StreamSyncJprepStudentPackageEventNats              = "syncjprepstudentpackage"
	SubjectSyncJprepStudentPackageEventNats             = "SyncJprepStudentPackage.Synced"
	QueueNotificationSyncJprepStudentPackageEventNats   = "queue-notification-sync-jprep-student-package"
	DurableNotificationSyncJprepStudentPackageEventNats = "durable-notification-sync-jprep-student-package"
	DeliverNotificationSyncJprepStudentPackageEventNats = "deliver.notification-sync-jprep-student-package"

	QueueStudentSubscriptionClassEventNats   = "queue-student-subscription-class"
	DurableStudentSubscriptionClassEventNats = "durable-student-subscription-class"
	DeliverStudentSubscriptionClassEventNats = "deliver.student-subscription-class"

	QueueStudentSubscriptionLessonMemberEventNats   = "queue-student-subscription-lesson-member"
	DurableStudentSubscriptionLessonMemberEventNats = "durable-student-subscription-lesson-member"
	DeliverStudentSubscriptionLessonMemberEventNats = "deliver.student-subscription-lesson-member"

	QueueStudentCourseDurationNats        = "queue-lesson-student-course-duration"
	DurableStudentCourseDurationEventNats = "durable-lesson-student-course-duration"
	DeliverStudentCourseDurationEventNats = "deliver.lesson-student-course-duration"

	// Deprecated
	SubjectConversationEventNats = "subject_conversation_event"
	// Deprecated
	QueueConversationEventNats = "queue_conversation_event"
	// Deprecated
	DurableConversationEventNats = "durable_conversation_event"

	StreamESConversation          = "esconversation"
	SubjectESConversation         = "ESConversation.>"
	SubjectCourseStudentEventNats = "ESConversation.CourseStudent.Elastic"
	QueueCourseStudentEventNats   = "queue-esconversation-course-student-elastic"
	DurableCourseStudentEventNats = "durable-esconversation-course-student-elastic"
	DeliverCourseStudentEventNats = "deliver.esconversation-course-student-elastic"

	// Deprecated
	SubjectMessageEventNats = "subject_message_event"
	// Deprecated
	QueueMessageEventNats = "queue_message_event"
	// Deprecated
	DurableMessageEventNats = "durable_message_event"

	StreamChatMigration                = "chat_migrate"
	SubjectChatMigrateResourcePath     = "chat_migrate.resource_path"
	ChatMigrateResourcePathConsumerKey = "chat_migrate.tom"

	// replace SubjectConversationEventNats
	StreamChat                = "chat"
	SubjectChat               = "chat.chat.>"
	SubjectChatCreated        = "chat.chat.created"
	SubjectChatUpdated        = "chat.chat.updated"
	SubjectChatMessageCreated = "chat.chat.message.created"
	SubjectChatMembersUpdated = "chat.chat.members.updated"

	// TODO: consider merging those
	ConsumerElasticChatMessageCreated = "chat.chat.message.created.elastic"
	ConsumerElasticChatMembersUpdated = "chat.chat.members.updated.elastic"
	ConsumerElasticChatUpdated        = "chat.chat.updated.elasticc"
	ConsumerElasticChatCreated        = "chat.chat.created.elastic"
	QueueElasticChatMessageCreated    = "group.chat.chat.message.created.elastic"
	QueueElasticChatMembersUpdated    = "group.chat.chat.members.updated.elastic"
	QueueElasticChatCreated           = "group.chat.chat.created.elastic"
	QueueElasticChatUpdated           = "group.chat.chat.updated.elastic"
	DurableElasticChatMessageCreated  = "durable_chat_chat_message_created_elastic"
	DurableElasticChatMembersUpdated  = "durable_chat_chat_members_updated_elastic"
	DurableElasticChatCreated         = "durable_chat_chat_created_elastic"
	DurableElasticChatUpdated         = "durable_chat_chat_updated_elastic"
	DeliverElasticChatMessageCreated  = "deliver.chat.chat.created.created.elastic"
	DeliverElasticChatMembersUpdated  = "deliver.chat.chat.members.updated.elastic"
	DeliverElasticChatUpdated         = "deliver.chat.chat.updated.elastic"
	DeliverElasticChatCreated         = "deliver.chat.chat.created.elastic"

	// TODO: conflicting with SubjectConversationInternal
	SubjectESConversationByMessageEventCreated = "ESConversationByMessageEvent.Created"
	QueueESConversationByMessageEventCreated   = "queue_es_conversation_by_message_event_created"
	DurableESConversationByMessageEventCreated = "durable_es_conversation_by_message_event_created"

	StreamLearningObjectives            = "learningobjectives"
	SubjectLearningObjectivesCreated    = "LearningObjective.Created"
	QueueGroupLearningObjectivesCreated = "queue-learning-objectives-created"
	DurableLearningObjectivesCreated    = "durable-learning-objectives-created"
	DeliverLearningObjectivesCreated    = "deliver.learning-objectives"

	// import study plan by csv
	StreamStudyPlanItems                    = "studyplanitems"
	SubjectStudyPlanItemsImported           = "StudyPlanItems.Imported"
	QueueGroupSubjectStudyPlanItemsImported = "queue-study-plan-items-imported"
	DurableStudyPlanItemsImported           = "durable-study-plan-items-imported"
	DeliverStudyPlanItemsImported           = "deliver.study-plan-items"

	// SubjectAssignmentsCreated = "subject_assignments_created"

	StreamUserDeviceToken            = "userdevicetoken"
	SubjectUserDeviceTokenUpdated    = "UserDeviceToken.Updated"
	QueueGroupUserDeviceTokenUpdated = "queue-user-device-token-updated"   //nolint:gosec
	DurableUserDeviceTokenUpdated    = "durable-user-device-token-updated" //nolint:gosec
	DeliverUserDeviceTokenUpdated    = "deliver.user-device-token-updated" //nolint:gosec

	StreamActivityLog             = "activitylog"
	SubjectActivityLogCreated     = "ActivityLog.Created"
	DurableActivityLogCreated     = "durable-activity-log-created"
	DurableActivityLogCreatedPull = "durable-activity-log-created-pull"
	QueueActivityLogCreated       = "queue-activity-log-created"
	DeliverActivityLogCreated     = "deliver.ActivityLogCreated"

	// Subject for nats-jetstream
	SubjectStudentPackageEventNatsJS            = "StudentPackage.Upserted"
	StreamChatMessage                           = "chatmessage"
	SubjectSendChatMessageCreated               = "ChatMessage.Created"
	QueueChatMessageCreated                     = "queue-chat-message-created"
	DeliverChatMessageCreated                   = "deliver.chat-message-created"
	DurableChatMessageCreated                   = "durable-chat-message-created"
	SubjectChatMessageDeleted                   = "ChatMessage.Deleted"
	QueueChatMessageDeleted                     = "queue-chat-message-deleted"
	DeliverChatMessageDeleted                   = "deliver.chat-message-deleted"
	DurableChatMessageDeleted                   = "durable-chat-message-deleted"
	StreamSyncUserCourse                        = "syncusercourse"
	SubjectJPREPSyncUserCourseNatsJS            = "SyncUserCourse.Synced"
	DurableJPREPSyncUserCourseNatsJS            = "durable-sync-user-course"
	DeliverSyncUserCourse                       = "deliver.sync-user-course"
	QueueJPREPSyncUserCourseNatsJS              = "queue-sync-user-course"
	DurableJPREPLogNatsJS                       = "durable-log-payload"
	DeliverSyncUserRegistrationLogPayload       = "deliver.sync-user-registration-log-payload"
	QueueJPREPLogNatsJS                         = "queue-log-payload"
	DeliverSyncUserCourseLogPayload             = "deliver.sync-user-course-log-payload"
	StreamCloudConvertJobEvent                  = "cloudconvertjobevent"
	SubjectCloudConvertJobEventNatsJS           = "CloudConvertJobEvent.Updated"
	DurableCloudConvertJobEventNatsJS           = "durable-cloud-convert"
	DeliverCloudConvertJobEvent                 = "deliver.cloud-convert"
	QueueCloudConvertJobEventNatsJS             = "queue-cloud-convert"
	StreamSyncUserRegistration                  = "syncuserregistration"
	SubjectUserRegistrationNatsJS               = "SyncUserRegistration.Synced"
	DurableSyncStaff                            = "durable-sync-staff"
	DeliverSyncUserRegistrationStaff            = "deliver.sync-user-registration-staff"
	QueueSyncStaff                              = "queue-sync-staff"
	DurableSyncStudentPackage                   = "durable-sync-student-package"
	DeliverSyncUserRegistrationStudentPackage   = "deliver.sync-user-registration-student-package"
	QueueSyncStudentPackage                     = "queue-sync-student-package"
	DurableSyncClassMember                      = "durable-sync-class-member"
	DeliverSyncUserRegistrationClassMember      = "deliver.sync-user-registration-class-member"
	QueueSyncClassMember                        = "queue-sync-class-member"
	DurableSyncStudent                          = "durable-sync-student"
	DeliverSyncUserRegistrationStudent          = "deliver.sync-user-registration-student"
	QueueSyncStudent                            = "queue-sync-student"
	StreamSyncMasterRegistration                = "syncmasterregistration"
	SubjectSyncMasterRegistration               = "SyncMasterRegistration.Synced"
	DeliverSyncMasterRegistrationLogPayload     = "deliver.sync-master-registration-log-payload"
	DurableSyncLiveLesson                       = "durable-sync-live-lesson"
	DeliverSyncMasterRegistrationLiveLesson     = "deliver.sync-master-registration-live-lesson"
	QueueSyncLiveLesson                         = "queue-sync-live-lesson"
	DurableSyncClass                            = "durable-sync-class"
	DeliverSyncMasterRegistrationClass          = "deliver.sync-master-registration-class"
	QueueSyncClass                              = "queue-sync-class"
	DurableSyncCourseAcademic                   = "durable-sync-course-academic"
	DeliverSyncMasterRegistrationCourseAcademic = "deliver.sync-master-registration-course-academic"
	QueueSyncCourseAcademic                     = "queue-sync-course-academic"
	DurableSyncCourseClass                      = "durable-sync-course-class"
	DeliverSyncMasterRegistrationCourseClass    = "deliver.sync-master-registration-course-class"
	QueueSyncCourseClass                        = "queue-sync-course-class"
	DurableSyncCourse                           = "durable-sync-course"
	DeliverSyncMasterRegistrationCourse         = "deliver.sync-master-registration-course"
	QueueSyncCourse                             = "queue-sync-course"
	DurableSyncAcademicYear                     = "durable-sync-academic-year"
	DeliverSyncMasterRegistrationAcademicYear   = "deliver.sync-master-registration-academic-year"
	QueueSyncAcademicYear                       = "queue-sync-academic-year"

	// class
	SubjectClass                 = "Class.*"
	StreamClass                  = "class"
	DeliverClassEvent            = "deliver.class-event"
	SubjectClassUpserted         = "Class.Upserted"
	DurableClassUpserted         = "durable-class-upserted"
	QueueClassUpserted           = "queue-class-upserted"
	DurableInternalLessonCreated = "durable-internal-lesson-created"
	QueueInternalLessonCreated   = "queue-internal-lesson-created"

	// class - notification subscribe
	DeliverNotificationClassUpserted = "deliver.notification-class-upserted"
	DurableNotificationClassUpserted = "durable-notification-class-upserted"
	QueueNotificationClassUpserted   = "queue-notification-class-upserted"

	// entryexitmgmt
	DurableEntryExitUserCreated     = "durable-entryexit-user-created"
	QueueEntryExitUserCreated       = "queue-entryexit-user-created"
	DeliverEntryExitUserCreated     = "deliver.user-entryexit-created"
	DurableEntryExitUserCreatedPull = "durable-entryexit-user-created-pull"

	// invoicemgmt - payment details creation
	DurablePaymentDetailsUserCreated = "durable-paymentdetails-user-created"
	QueuePaymentDetailsUserCreated   = "queue-paymentdetails-user-created"
	DeliverPaymentDetailsUserCreated = "deliver.user-paymentdetails-created"

	// invoicemgmt - payment details updated
	DurablePaymentDetailsUserUpdated = "durable-paymentdetails-user-updated"
	QueuePaymentDetailsUserUpdated   = "queue-paymentdetails-user-updated"
	DeliverPaymentDetailsUserUpdated = "deliver.user-paymentdetails-updated"

	// lesson chat
	SubjectLessonChatSynced = "LessonChat.Synced"
	DeliverSyncLessonChat   = "deliver.sync-lesson-chat"
	StreamLessonChat        = "lessonchat"
	DurableLessonChat       = "durable-sync-lesson-chat"
	QueueLessonChat         = "queue-sync-lesson-chat"

	StreamSyncStudentPackage        = "syncstudentpackage"
	SubjectSyncStudentPackage       = "SyncStudentPackage.Synced"
	DurableSyncStudentPackageFatima = "durable-sync-student-package-fatima"
	QueueSyncStudentPackageFatima   = "queue-sync-student-package-fatima"
	DeliverSyncStudentPackageFatima = "deliver.sync-student-package-fatima"
	DurableSyncStudentPackageEureka = "durable-sync-student-package-eureka"
	DeliverSyncStudentPackageEureka = "deliver.sync-student-package-eureka"
	QueueSyncStudentPackageEureka   = "queue-sync-student-package-eureka"

	SubjectAssignmentsCreated = "Assignments.Created"
	QueueAssignmentsCreated   = "queue-assignments-created"
	DurableAssignmentsCreated = "durable-assignments-created"
	StreamAssignments         = "assignments"
	DeliverAssignmentCreated  = "deliver.assignments-created"

	// lesson
	SubjectLesson        = "Lesson.*"
	SubjectLessonCreated = "Lesson.Created"
	DeliverLessonCreated = "deliver.internal-lesson-created"
	StreamLesson         = "lesson"
	DeliverLessonEvent   = "deliver.lesson-event"
	DurableLesson        = "durable-lesson"
	QueueLesson          = "queue-lesson"

	// lesson - deleted
	SubjectLessonDeleted = "Lesson.Deleted"
	QueueLessonDeleted   = "queue-lesson-deleted"
	DurableLessonDeleted = "durable-lesson-deleted"
	DeliverLessonDeleted = "deliver.lesson-deleted"

	// lesson - updated
	SubjectLessonUpdated = "Lesson.Updated"
	QueueLessonUpdated   = "queue-lesson-updated"
	DurableLessonUpdated = "durable-lesson-updated"
	DeliverLessonUpdated = "deliver.lesson-updated"

	SubjectEnrollmentStatusAssignmentCreated = "EnrollmentStatusAssignment.Created"
	StreamEnrollmentStatusAssignment         = "enrollmentstatusassignment"
	DurableEnrollmentStatusAssignment        = "durable-enrollment-status-assignment"
	DeliverEnrollmentStatusAssignment        = "deliver.enrollment-status-assignment"
	QueueEnrollmentStatusAssignment          = "queue-enrollment-status-assignment"

	// lesson - upcoming lesson notification
	SubjectUpcomingLiveLessonNotification = "UpcomingLiveLessonNotification.Created"
	QueueUpcomingLiveLessonNotification   = "queue-upcoming-live-lesson-notification"
	DurableUpcomingLiveLessonNotification = "durable-upcoming-live-lesson-notification"
	DeliverUpcomingLiveLessonNotification = "deliver.upcoming-live-lesson-notification"
	StreamUpcomingLiveLessonNotification  = "upcominglivelessonnotification"

	// sync student lesson
	SubjectSyncStudentLessons              = "SyncStudentLessonsConversations.Synced"
	QueueSyncStudentLessonsConversations   = "queue-sync-student-lessons-conversations"
	DurableSyncStudentLessonsConversations = "durable-sync-student-lessons-conversations"
	DeliverSyncStudentLessonsConversations = "deliver.sync-student-lessons-conversations"
	StreamSyncStudentLessons               = "syncstudentlessonsconversations"

	// sync grade
	StreamSyncGradeEvent   = "syncgrade"
	SubjectSyncGradeUpsert = "SyncGrade.upsert"
	QueueSyncGradeUpsert   = "queue-sync-grade-upsert"
	DeliverSyncGradeUpsert = "deliver.sync-grade-upsert"
	DurableGradeUpsert     = "durable-sync-grade-upsert" //nolint:gosec

	// user
	SubjectUser        = "User.*"
	SubjectUserCreated = "User.Created"
	SubjectUserUpdated = "User.Updated"

	DeliverUserCreated = "deliver.user-created"
	QueueUserCreated   = "queue-user-created"
	DurableUserCreated = "durable-user-created"

	DeliverUserUpdatedTom = "deliver.user-updated-tom"
	QueueUserUpdatedTom   = "queue-user-updated-tom"
	DurableUserUpdatedTom = "durable-user-updated-tom"

	DeliverConversationMgmtUserCreated = "deliver.conversationmgmt-user-created"
	QueueConversationMgmtUserCreated   = "queue-conversationmgmt-user-created"
	DurableConversationMgmtUserCreated = "durable-conversationmgmt-user-created"

	StreamUser = "user"

	// staff
	SubjectStaffUpsertTimesheetConfig = "Staff.UpsertTimesheetConfig"
	SubjectUpsertStaff                = "Staff.Upserted"
	StreamStaff                       = "staff"

	DeliverUpsertStaffTom = "deliver.upsert-staff-tom"
	QueueUpsertStaffTom   = "queue-upsert-staff-tom"
	DurableUpsertStaffTom = "durable-upsert-staff-tom"

	DeliverConversationMgmtUpsertStaff = "deliver.conversationmgmt-upsert-staff"
	QueueConversationMgmtUpsertStaff   = "queue-conversationmgmt-upsert-staff"
	DurableConversationMgmtUpsertStaff = "durable-conversationmgmt-upsert-staff"

	// master data
	StreamSyncLocationTypeUpserted             = "synclocationtype"
	SubjectSyncLocationTypeUpserted            = "SyncLocationType.Upserted"
	DurableSyncLocationTypeUpsertedOrgCreation = "durable-synclocationtype-org-creation"
	QueueSyncLocationTypeUpsertedOrgCreation   = "queue-synclocationtype-org-creation"
	DeliverSyncLocationTypeUpsertedOrgCreation = "deliver.synclocationtype-org-creation"

	DurableSyncLocationTypeUpsertedImporter = "durable-synclocationtype-importer"
	QueueSyncLocationTypeUpsertedImporter   = "queue-synclocationtype-importer"
	DeliverSyncLocationTypeUpsertedImporter = "deliver.synclocationtype-importer"

	// sync location
	StreamSyncLocationUpserted             = "synclocation"
	SubjectSyncLocationUpserted            = "SyncLocation.Upserted"
	DurableSyncLocationUpsertedOrgCreation = "durable-synclocation-org-creation"
	QueueSyncLocationUpsertedOrgCreation   = "queue-synclocation-org-creation"
	DeliverSyncLocationUpsertedOrgCreation = "deliver.synclocation-org-creation"

	DurableSyncLocationUpsertedImporter = "durable-synclocation-importer"
	QueueSyncLocationUpsertedImporter   = "queue-synclocation-importer"
	DeliverSyncLocationUpsertedImporter = "deliver.synclocation-importer"

	// Organization
	SubjectOrganization        = "Organization.*"
	SubjectOrganizationCreated = "Organization.Created"
	StreamOrganization         = "organization"
	QueueOrganizationCreated   = "queue-organization-created"
	DurableOrganizationCreated = "durable-organization-created"
	DeliverOrganizationCreated = "deliver.Organization-created"

	// class
	SubjectMasterMgmtClass         = "MasterMgmt.Class.*"
	SubjectMasterMgmtClassUpserted = "MasterMgmt.Class.Upserted"
	StreamMasterMgmtClass          = "mastermgmtclass"
	DurableMasterMgmtClassUpserted = "durable-master-mgmt-class-upserted"
	DeliverMasterMgmtClassEvent    = "deliver.master-mgmt-class-event"
	QueueMasterMgmtClassUpserted   = "queue-master-mgmt-class-upserted"

	// reserve class
	SubjectMasterMgmtReserveClass         = "MasterMgmt.ReserveClass.*"
	SubjectMasterMgmtReserveClassUpserted = "MasterMgmt.ReserveClass.Upserted"
	StreamMasterMgmtReserveClass          = "mastermgmtreserveclass"
	DurableMasterMgmtReserveClassUpserted = "durable-master-mgmt-reserve-class-upserted"
	DeliverMasterMgmtReserveClassEvent    = "deliver.master-mgmt-reserve-class-event"
	QueueMasterMgmtReserveClassUpserted   = "queue-master-mgmt-reserve-class-upserted"

	// student change class
	QueueStudentSubscriptionChangeClassEventNats   = "queue-student-subscription-change-class"
	DurableStudentSubscriptionChangeClassEventNats = "durable-student-subscription-change-class"
	DeliverStudentSubscriptionChangeClassEventNats = "deliver.student-subscription-change-class"

	// import student
	StreamImportStudentEvent  = "importstudent"
	SubjectImportStudentEvent = "ImportStudent.Upserted"
	QueueImportStudentEvent   = "queue-import-student"
	DurableImportStudentEvent = "durable-import-student"
	DeliverImportStudentEvent = "deliver.import-student"

	// import parent
	StreamImportParentEvent  = "importparent"
	SubjectImportParentEvent = "ImportParent.Upserted"
	QueueImportParentEvent   = "queue-import-parent"
	DurableImportParentEvent = "durable-import-parent"
	DeliverImportParentEvent = "deliver.import-parent"

	// timesheet
	QueueTimesheetLesson        = "queue-timesheet-lesson"
	DurableTimesheetLesson      = "durable-timesheet-lesson"
	DeliverTimesheetLessonEvent = "deliver.timesheet-lesson-event"
	StreamTimesheetLesson       = "timesheetlesson"
	SubjectTimesheetLesson      = "TimesheetLesson.Locked"

	// timesheet action log
	QueueTimesheetActionLog        = "queue-timesheet-action-log"
	DurableTimesheetActionLog      = "durable-timesheet-action-log"
	DeliverTimesheetActionLogEvent = "deliver.timesheet-action-log-event"
	StreamTimesheetActionLog       = "timesheetactionlog"
	SubjectTimesheetActionLog      = "TimesheetActionLog.Created"

	// timesheet auto create flag
	QueueTimesheetAutoCreateFlag   = "queue-timesheet-auto-create-flag"
	DurableTimesheetAutoCreateFlag = "durable-timesheet-auto-create-flag"
	DeliverTimesheetAutoCreateFlag = "deliver.timesheet-auto-create-flag-event"
	StreamTimesheetAutoCreateFlag  = "timesheetautocreateflag"
	SubjectTimesheetAutoCreateFlag = "TimesheetAutoCreateFlag.Updated"

	QueueLockLesson   = "queue-lock-lesson"
	DurableLockLesson = "durable-lock-lesson"
	DeliverLockLesson = "deliver.lock-lesson"

	StreamStudentEventLog          = "studenteventlogs"
	SubjectStudentEventLogsCreated = "StudentEventLogs.Created"
	DeliverStudentEventLog         = "deliver.student-event-logs"
	QueueStudentEventLogsCreated   = "queue-student-event-logs-created"
	DurableStudentEventLogsCreated = "durable-student-event-logs-created"

	SubjectNotificationCreated = "Notification.Created"
	StreamNotification         = "notification"
	DurableNotification        = "durable-notification"
	QueueNotification          = "queue-notification"
	DeliverNotification        = "deliver.notification"

	// order management
	StreamOrderEventLog         = "ordereventlog"
	SubjectOrderEventLogCreated = "OrderEventLog.Created"
	DeliverOrderEventLogCreated = "deliver.order-event-log-created"
	QueueOrderEventLogCreated   = "queue-order-event-log-created"
	DurableOrderEventLogCreated = "durable-order-event-log-created"

	StreamStudentCourseEventSync  = "studentcourseeventsync"
	SubjectStudentCourseEventSync = "StudentCourseEvent.Sync"
	DeliverStudentCourseEventSync = "deliver.student-course-event-sync"
	QueueStudentCourseEventSync   = "queue-student-course-event-sync"
	DurableStudentCourseEventSync = "durable-student-course-event-sync"

	QueueLessonSyncStudentCourseSlotInfo   = "queue-lesson-sync-student-course-slot-info"
	DeliverLessonSyncStudentCourseSlotInfo = "deliver.lesson-sync-student-course-slot-info"
	DurableLessonSyncStudentCourseSlotInfo = "durable-lesson-sync-student-course-slot-info"

	StreamOrderWithProductInfoLog         = "orderwithproductinfolog"
	SubjectOrderWithProductInfoLogCreated = "OrderWithProductInfoLog.Created"
	DeliverOrderWithProductInfoLogCreated = "deliver.order-with-product-info-log-created"
	QueueOrderWithProductInfoLogCreated   = "queue-order-with-product-info-log-created"
	DurableOrderWithProductInfoLogCreated = "durable-order-with-product-info-log-created"

	StreamUpdateStudentProduct         = "updatestudentproduct"
	SubjectUpdateStudentProductCreated = "UpdateStudentProduct.Created"
	DeliverUpdateStudentProductCreated = "deliver.update-student-product-created"
	QueueUpdateStudentProductCreated   = "queue-update-student-product-created"
	DurableUpdateStudentProductCreated = "durable-update-student-product-created"

	// hephaestus
	StreamDebeziumIncrementalSnapshot      = "debeziumincrementalsnapshot"
	SubjectDebeziumIncrementalSnapshotSend = "DebeziumIncrementalSnapshot.Send"

	DurableDebeziumIncrementalSnapshotSend = "durable-debezium-incremental-snapshot"

	QueueBobDebeziumIncrementalSnapshotSend   = "queue-bob-debezium-incremental-snapshot"
	DurableBobDebeziumIncrementalSnapshotSend = "durable-bob-debezium-incremental-snapshot"

	QueueCalendarDebeziumIncrementalSnapshotSend   = "queue-calendar-debezium-incremental-snapshot"
	DurableCalendarDebeziumIncrementalSnapshotSend = "durable-calendar-debezium-incremental-snapshot"

	QueueFatimaDebeziumIncrementalSnapshotSend   = "queue-fatima-debezium-incremental-snapshot"
	DurableFatimaDebeziumIncrementalSnapshotSend = "durable-fatima-debezium-incremental-snapshot"

	QueueMastermgmtDebeziumIncrementalSnapshotSend   = "queue-mastermgmt-debezium-incremental-snapshot"
	DurableMastermgmtDebeziumIncrementalSnapshotSend = "durable-mastermgmt-debezium-incremental-snapshot"

	DeliverDebeziumIncrementalSnapshotSend = "deliver.debezium-incremental-snapshot"

	StreamCleanDataTestEventNats     = "clean_data"
	SubjectCleanDataTestEventNats    = "CleanData.Created"
	QueueArchitectureCleanDataTes    = "queue-architecture-clean-data"
	DeliverArchitectureCleanDataTes  = "deliver.durable-architecture-clean-data"
	DurableArchitectureCleanDataTest = "durable-architecture-clean-data"

	// user group
	StreamUserGroup        = "usergroup"
	SubjectUpsertUserGroup = "UserGroup.Upserted"

	DeliverUpsertUserGroupTom = "deliver.upsert-usergroup-tom"
	QueueUpsertUserGroupTom   = "queue-upsert-usergroup-tom"
	DurableUpserUserGroupTom  = "durable-upsert-usergroup-tom"

	// virtual classroom
	QueueLessonDefaultChatState   = "queue-lesson-default-chat-state"
	DeliverLessonDefaultChatState = "deliver.lesson-default-chat-state"
	DurableLessonDefaultChatState = "durable-lesson-default-chat-state"

	QueueCreateLiveLessonRoom   = "queue-create-live-lesson-room"
	DeliverCreateLiveLessonRoom = "deliver.create-live-lesson-room"
	DurableCreateLiveLessonRoom = "durable-create-live-lesson-room"

	// virtual classroom - live room
	StreamLiveRoom         = "liveroom"
	SubjectLiveRoomUpdated = "LiveRoom.Updated"
	QueueLiveRoom          = "queue-live-room"
	DeliverLiveRoom        = "deliver.live-room"
	DurableLiveRoom        = "durable-live-room"

	// KAFKA TOPICS
	// Kafka topic name should be: {{team-name}}.{{topic-name}}

	// email sending
	EmailSendingTopic = "communication.email-sending"

	// system notifications
	SystemNotificationUpsertingTopic = "notification.system-notification-upserting"
)

const (
	ManabieSchool = math.MinInt32 + iota
	JPREPSchool
	SynersiaSchool
	RenseikaiSchool
	TestingSchool
	GASchool
	KECSchool
	AICSchool
	NSGSchool
	E2ETokyo
	E2EHCM
	_
	_
	KECDemo
	_
	_
	_
	_
	ManagaraBase
	ManagaraHighSchool

	UsermgmtSF = 100013
)

// Default locations of current partner
const (
	ManabieOrgLocation      = "01FR4M51XJY9E77GSN4QZ1Q9N1"
	JPREPOrgLocation        = "01FR4M51XJY9E77GSN4QZ1Q9N2"
	E2EOrgLocation          = "01FR4M51XJY9E77GSN4QZ1Q9N5"
	KECDemoOrgLocation      = "01FR4M51XJY9E77GSN4QZ1Q8N4"
	ManagaraBaseOrgLocation = "01GFMMFRXC6SKTTT44HWR3BRY8"
	ManagaraHSOrgLocation   = "01GFMNHQ1WHGRC8AW6K913AM3G"
	SyncAccount             = "01GMA8R9VGZGS6W5FSRRHHAN97"
)

const (
	Japanese = "Japanese"
	English  = "English"
	Mandarin = "Mandarin" // Kenji in japansese
)

const (
	ESConversationIndexName = "conversations"
)

// fields
const (
	Name            = "name"
	Country         = "country"
	SchoolID        = "schoolID"
	Subject         = "subject"
	Grade           = "grade"
	CountryAndGrade = "country and grade"
	All             = "all"
	None            = "none"
)

const UserMgmtTask = "usermgmt-task"

const (
	DetectTextFromImageRetryInitial    = 100 * time.Millisecond
	DetectTextFromImageRetryMax        = 15 * time.Second
	DetectTextFromImageRetryMultiplier = 1.30
)

const (
	BrightcoveConfigKey = "mastermgmt.brightcoveconfig"
	BrightcoveAccountID = "mastermgmt.brightcoveconfig.accountid"
)
