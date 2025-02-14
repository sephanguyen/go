/*
  ____            _   _           _
 / ___|   _   _  | | | |   __ _  | |__    _   _   ___
 \___ \  | | | | | | | |  / _` | | '_ \  | | | | / __|
  ___) | | |_| | | | | | | (_| | | |_) | | |_| | \__ \
 |____/   \__, | |_| |_|  \__,_| |_.__/   \__,_| |___/
          |___/

*/
// * Code generated by protoc-gen-syllabus. DO NOT EDIT.
// * versions:
// * - protoc-gen-syllabus (devel)
// * - protoc                   v3.14.0
// * source: common

package scpb

func (x NotificationType) FromString(str string) NotificationType {
	switch str {
	case NotificationType_NOTIFICATION_TYPE_NONE.String():
		return NotificationType_NOTIFICATION_TYPE_NONE
	case NotificationType_NOTIFICATION_TYPE_TEXT.String():
		return NotificationType_NOTIFICATION_TYPE_TEXT
	case NotificationType_NOTIFICATION_TYPE_PROMO_CODE.String():
		return NotificationType_NOTIFICATION_TYPE_PROMO_CODE
	case NotificationType_NOTIFICATION_TYPE_ASSIGNMENT.String():
		return NotificationType_NOTIFICATION_TYPE_ASSIGNMENT
	case NotificationType_NOTIFICATION_TYPE_COMPOSED.String():
		return NotificationType_NOTIFICATION_TYPE_COMPOSED
	case NotificationType_NOTIFICATION_TYPE_NATS_ASYNC.String():
		return NotificationType_NOTIFICATION_TYPE_NATS_ASYNC
	}
	return NotificationType(0)
}

func (x NotificationStatus) FromString(str string) NotificationStatus {
	switch str {
	case NotificationStatus_NOTIFICATION_STATUS_NONE.String():
		return NotificationStatus_NOTIFICATION_STATUS_NONE
	case NotificationStatus_NOTIFICATION_STATUS_DRAFT.String():
		return NotificationStatus_NOTIFICATION_STATUS_DRAFT
	case NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String():
		return NotificationStatus_NOTIFICATION_STATUS_SCHEDULED
	case NotificationStatus_NOTIFICATION_STATUS_SENT.String():
		return NotificationStatus_NOTIFICATION_STATUS_SENT
	case NotificationStatus_NOTIFICATION_STATUS_DISCARD.String():
		return NotificationStatus_NOTIFICATION_STATUS_DISCARD
	}
	return NotificationStatus(0)
}

func (x NotificationEvent) FromString(str string) NotificationEvent {
	switch str {
	case NotificationEvent_NOTIFICATION_EVENT_NONE.String():
		return NotificationEvent_NOTIFICATION_EVENT_NONE
	case NotificationEvent_NOTIFICATION_EVENT_X_LO_COMPLETED.String():
		return NotificationEvent_NOTIFICATION_EVENT_X_LO_COMPLETED
	case NotificationEvent_NOTIFICATION_EVENT_TEACHER_GIVE_ASSIGNMENT.String():
		return NotificationEvent_NOTIFICATION_EVENT_TEACHER_GIVE_ASSIGNMENT
	case NotificationEvent_NOTIFICATION_EVENT_TEACHER_RETURN_ASSIGNMENT.String():
		return NotificationEvent_NOTIFICATION_EVENT_TEACHER_RETURN_ASSIGNMENT
	case NotificationEvent_NOTIFICATION_EVENT_STUDENT_SUBMIT_ASSIGNMENT.String():
		return NotificationEvent_NOTIFICATION_EVENT_STUDENT_SUBMIT_ASSIGNMENT
	case NotificationEvent_NOTIFICATION_EVENT_ASSIGNMENT_UPDATED.String():
		return NotificationEvent_NOTIFICATION_EVENT_ASSIGNMENT_UPDATED
	}
	return NotificationEvent(0)
}

func (x UserNotificationStatus) FromString(str string) UserNotificationStatus {
	switch str {
	case UserNotificationStatus_USER_NOTIFICATION_STATUS_NONE.String():
		return UserNotificationStatus_USER_NOTIFICATION_STATUS_NONE
	case UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String():
		return UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW
	case UserNotificationStatus_USER_NOTIFICATION_STATUS_SEEN.String():
		return UserNotificationStatus_USER_NOTIFICATION_STATUS_SEEN
	case UserNotificationStatus_USER_NOTIFICATION_STATUS_READ.String():
		return UserNotificationStatus_USER_NOTIFICATION_STATUS_READ
	case UserNotificationStatus_USER_NOTIFICATION_STATUS_FAILED.String():
		return UserNotificationStatus_USER_NOTIFICATION_STATUS_FAILED
	}
	return UserNotificationStatus(0)
}

func (x NotificationTargetGroupSelect) FromString(str string) NotificationTargetGroupSelect {
	switch str {
	case NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE.String():
		return NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE
	case NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL.String():
		return NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL
	case NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST.String():
		return NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST
	}
	return NotificationTargetGroupSelect(0)
}

func (x QuestionType) FromString(str string) QuestionType {
	switch str {
	case QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE.String():
		return QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE
	case QuestionType_QUESTION_TYPE_CHECK_BOX.String():
		return QuestionType_QUESTION_TYPE_CHECK_BOX
	case QuestionType_QUESTION_TYPE_FREE_TEXT.String():
		return QuestionType_QUESTION_TYPE_FREE_TEXT
	}
	return QuestionType(0)
}

func (x UserNotificationQuestionnaireStatus) FromString(str string) UserNotificationQuestionnaireStatus {
	switch str {
	case UserNotificationQuestionnaireStatus_USER_NOTIFICATION_QUESTIONNAIRE_STATUS_ANSWERED.String():
		return UserNotificationQuestionnaireStatus_USER_NOTIFICATION_QUESTIONNAIRE_STATUS_ANSWERED
	case UserNotificationQuestionnaireStatus_USER_NOTIFICATION_QUESTIONNAIRE_STATUS_UNANSWERED.String():
		return UserNotificationQuestionnaireStatus_USER_NOTIFICATION_QUESTIONNAIRE_STATUS_UNANSWERED
	}
	return UserNotificationQuestionnaireStatus(0)
}
