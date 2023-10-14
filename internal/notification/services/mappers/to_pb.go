package mappers

import (
	"fmt"
	"sort"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func NotificationTargetToPb(i *entities.InfoNotificationTarget) *cpb.NotificationTargetGroup {
	userGroup := make([]cpb.UserGroup, 0, len(i.UserGroupFilter.UserGroups))
	for _, gr := range i.UserGroupFilter.UserGroups {
		userGroup = append(userGroup, cpb.UserGroup(cpb.UserGroup_value[gr]))
	}
	dst := &cpb.NotificationTargetGroup{
		CourseFilter: &cpb.NotificationTargetGroup_CourseFilter{
			Type:      cpb.NotificationTargetGroupSelect(cpb.NotificationTargetGroupSelect_value[i.CourseFilter.Type]),
			CourseIds: i.CourseFilter.CourseIDs,
			Courses:   NotificationTargetCoursesToPb(i.CourseFilter.Courses),
		},
		GradeFilter: &cpb.NotificationTargetGroup_GradeFilter{
			Type:     cpb.NotificationTargetGroupSelect(cpb.NotificationTargetGroupSelect_value[i.GradeFilter.Type]),
			GradeIds: i.GradeFilter.GradeIDs,
		},
		UserGroupFilter: &cpb.NotificationTargetGroup_UserGroupFilter{
			UserGroups: userGroup,
		},
		LocationFilter: &cpb.NotificationTargetGroup_LocationFilter{
			Type:        cpb.NotificationTargetGroupSelect(cpb.NotificationTargetGroupSelect_value[i.LocationFilter.Type]),
			LocationIds: i.LocationFilter.LocationIDs,
			Locations:   NotificationTargetLocationsToPb(i.LocationFilter.Locations),
		},
		ClassFilter: &cpb.NotificationTargetGroup_ClassFilter{
			Type:     cpb.NotificationTargetGroupSelect(cpb.NotificationTargetGroupSelect_value[i.ClassFilter.Type]),
			ClassIds: i.ClassFilter.ClassIDs,
			Classes:  NotificationTargetClassesToPb(i.ClassFilter.Classes),
		},
		SchoolFilter: &cpb.NotificationTargetGroup_SchoolFilter{
			Type:      cpb.NotificationTargetGroupSelect(cpb.NotificationTargetGroupSelect_value[i.SchoolFilter.Type]),
			SchoolIds: i.SchoolFilter.SchoolIDs,
			Schools:   NotificationTargetSchoolsToPb(i.SchoolFilter.Schools),
		},
	}
	return dst
}

func NotificationTargetCoursesToPb(courses []entities.InfoNotificationTarget_CourseFilter_Course) []*cpb.NotificationTargetGroup_CourseFilter_Course {
	coursesPb := make([]*cpb.NotificationTargetGroup_CourseFilter_Course, 0, len(courses))
	for _, course := range courses {
		coursesPb = append(coursesPb, &cpb.NotificationTargetGroup_CourseFilter_Course{
			CourseId:   course.CourseID,
			CourseName: course.CourseName,
		})
	}
	return coursesPb
}

func NotificationTargetLocationsToPb(locations []entities.InfoNotificationTarget_LocationFilter_Location) []*cpb.NotificationTargetGroup_LocationFilter_Location {
	locationsPb := make([]*cpb.NotificationTargetGroup_LocationFilter_Location, 0, len(locations))
	for _, location := range locations {
		locationsPb = append(locationsPb, &cpb.NotificationTargetGroup_LocationFilter_Location{
			LocationId:   location.LocationID,
			LocationName: location.LocationName,
		})
	}
	return locationsPb
}

func NotificationTargetClassesToPb(classes []entities.InfoNotificationTarget_ClassFilter_Class) []*cpb.NotificationTargetGroup_ClassFilter_Class {
	classesPb := make([]*cpb.NotificationTargetGroup_ClassFilter_Class, 0, len(classes))
	for _, class := range classes {
		classesPb = append(classesPb, &cpb.NotificationTargetGroup_ClassFilter_Class{
			ClassId:   class.ClassID,
			ClassName: class.ClassName,
		})
	}
	return classesPb
}

func NotificationTargetSchoolsToPb(schools []entities.InfoNotificationTarget_SchoolFilter_School) []*cpb.NotificationTargetGroup_SchoolFilter_School {
	schoolsPb := make([]*cpb.NotificationTargetGroup_SchoolFilter_School, 0, len(schools))
	for _, school := range schools {
		schoolsPb = append(schoolsPb, &cpb.NotificationTargetGroup_SchoolFilter_School{
			SchoolId:   school.SchoolID,
			SchoolName: school.SchoolName,
		})
	}
	return schoolsPb
}

func QNQuestionsToPb(questions entities.QuestionnaireQuestions) []*cpb.Question {
	resp := make([]*cpb.Question, 0)
	for _, question := range questions {
		resp = append(resp, &cpb.Question{
			QuestionnaireQuestionId: question.QuestionnaireQuestionID.String,
			Title:                   question.Title.String,
			Type:                    cpb.QuestionType(cpb.QuestionType_value[question.Type.String]),
			Choices:                 database.FromTextArray(question.Choices),
			OrderIndex:              int64(question.OrderIndex.Int),
			Required:                question.IsRequired.Bool,
		})
	}
	return resp
}

func QNUserAnswersToPb(responders []*repositories.QuestionnaireResponder, questionnaireUserAnswers entities.QuestionnaireUserAnswers, questionnaireQuestions entities.QuestionnaireQuestions) []*npb.GetAnswersByFilterResponse_UserAnswer {
	resp := make([]*npb.GetAnswersByFilterResponse_UserAnswer, 0)

	mapUserIDAndUserAnswers := make(map[string]entities.QuestionnaireUserAnswers, 0)
	for _, qNUserAnswer := range questionnaireUserAnswers {
		mapUserIDAndUserAnswers[qNUserAnswer.UserID.String] = append(mapUserIDAndUserAnswers[qNUserAnswer.UserID.String], qNUserAnswer)
	}

	for _, responder := range responders {
		userAnswer := &npb.GetAnswersByFilterResponse_UserAnswer{
			ResponderName:      responder.Name.String,
			UserId:             responder.UserID.String,
			TargetId:           responder.TargetID.String,
			TargetName:         responder.TargetName.String,
			UserNotificationId: responder.UserNotificationID.String,
			IsParent:           responder.IsParent.Bool,
			IsIndividual:       responder.IsIndividual.Bool,
		}

		// Fill submitted_at field
		submittedAt := database.FromTimestamptz(responder.SubmittedAt)
		if submittedAt != nil {
			userAnswer.SubmittedAt = timestamppb.New(*submittedAt)
		}

		userAnswersEnt, ok := mapUserIDAndUserAnswers[responder.UserID.String]
		if ok {
			userAnswerSubmitedAt := timestamppb.Now()
			answers := make([]*cpb.Answer, 0)
			for _, userAnswerEnt := range userAnswersEnt {
				if userAnswerEnt.TargetID.String == userAnswer.TargetId {
					answers = append(answers, &cpb.Answer{
						QuestionnaireQuestionId: userAnswerEnt.QuestionnaireQuestionID.String,
						Answer:                  userAnswerEnt.Answer.String,
					})
					userAnswerSubmitedAt = timestamppb.New(userAnswerEnt.SubmittedAt.Time)
				}
			}

			// userAnswer.SubmittedAt is set above, but for some old submitted, it will be null -> need to get from user answer and assign for it.
			if userAnswer.SubmittedAt == nil {
				userAnswer.SubmittedAt = userAnswerSubmitedAt
			}
			userAnswer.Answers = answers
		}

		// This map support order user answer by question order index asc
		mapQuestionIDAndOrderIndex := make(map[string]int64)

		// Fill empty answer for question that don't have any answers
		mapQuestionIDAndAnswer := make(map[string][]*cpb.Answer)
		for _, answer := range userAnswer.Answers {
			mapQuestionIDAndAnswer[answer.QuestionnaireQuestionId] = append(mapQuestionIDAndAnswer[answer.QuestionnaireQuestionId], answer)
		}
		for _, question := range questionnaireQuestions {
			mapQuestionIDAndOrderIndex[question.QuestionnaireQuestionID.String] = int64(question.OrderIndex.Int)

			_, ok := mapQuestionIDAndAnswer[question.QuestionnaireQuestionID.String]
			if !ok {
				userAnswer.Answers = append(userAnswer.Answers, &cpb.Answer{
					QuestionnaireQuestionId: question.QuestionnaireQuestionID.String,
					Answer:                  "",
				})
			}
		}

		// Sort user answers
		sort.Slice(userAnswer.Answers, func(i, j int) bool {
			return mapQuestionIDAndOrderIndex[userAnswer.Answers[i].QuestionnaireQuestionId] <= mapQuestionIDAndOrderIndex[userAnswer.Answers[j].QuestionnaireQuestionId]
		})

		resp = append(resp, userAnswer)
	}
	return resp
}

func QNUserAnswerToPb(qn *entities.Questionnaire, questions entities.QuestionnaireQuestions, answers entities.QuestionnaireUserAnswers, isSubmitted bool) *cpb.UserQuestionnaire {
	resp := &cpb.UserQuestionnaire{}

	protobufQn := &cpb.Questionnaire{
		QuestionnaireId: qn.QuestionnaireID.String,
		ResubmitAllowed: qn.ResubmitAllowed.Bool,
		ExpirationDate:  timestamppb.New(qn.ExpirationDate.Time),
	}
	for _, item := range questions {
		protobufQn.Questions = append(protobufQn.Questions, &cpb.Question{
			QuestionnaireQuestionId: item.QuestionnaireQuestionID.String,
			Title:                   item.Title.String,
			Type:                    cpb.QuestionType(cpb.QuestionType_value[item.Type.String]),
			Choices:                 database.FromTextArray(item.Choices),
			OrderIndex:              int64(item.OrderIndex.Int),
			Required:                item.IsRequired.Bool,
		})
	}
	for _, item := range answers {
		resp.Answers = append(resp.Answers, &cpb.Answer{
			QuestionnaireQuestionId: item.QuestionnaireQuestionID.String,
			Answer:                  item.Answer.String,
		})
	}
	resp.Questionnaire = protobufQn
	resp.IsSubmitted = isSubmitted
	return resp
}

func NotificationToPb(noti *entities.InfoNotification, notiMsg *entities.InfoNotificationMsg) *cpb.Notification {
	content, _ := notiMsg.GetContent()
	targetGroup, _ := noti.GetTargetGroup()

	notiPb := &cpb.Notification{
		NotificationId: noti.NotificationID.String,
		Data:           string(noti.Data.Bytes),
		EditorId:       noti.EditorID.String,
		CreatedUserId:  noti.CreatedUserID.String,
		ReceiverIds:    database.FromTextArray(noti.ReceiverIDs),
		Message: &cpb.NotificationMessage{
			NotificationMsgId: notiMsg.NotificationMsgID.String,
			Title:             notiMsg.Title.String,
			Content: &cpb.RichText{
				Raw:      content.Raw,
				Rendered: content.RenderedURL,
			},
			MediaIds:  database.FromTextArray(notiMsg.MediaIDs),
			CreatedAt: timestamppb.New(notiMsg.CreatedAt.Time),
			UpdatedAt: timestamppb.New(notiMsg.UpdatedAt.Time),
		},
		Type:        cpb.NotificationType(cpb.NotificationType_value[noti.Type.String]),
		Event:       cpb.NotificationEvent(cpb.NotificationEvent_value[noti.Event.String]),
		Status:      cpb.NotificationStatus(cpb.NotificationStatus_value[noti.Status.String]),
		TargetGroup: NotificationTargetToPb(targetGroup),
		ScheduledAt: timestamppb.New(noti.ScheduledAt.Time),
		CreatedAt:   timestamppb.New(noti.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(noti.UpdatedAt.Time),
		SentAt:      timestamppb.New(noti.SentAt.Time),
	}
	return notiPb
}

func ToUserNotificationPb(srcEnt *entities.UserInfoNotification) *cpb.UserNotification {
	dstPb := &cpb.UserNotification{
		UserId:             srcEnt.UserID.String,
		CourseId:           database.FromTextArray(srcEnt.Courses),
		NotificationId:     srcEnt.NotificationID.String,
		Status:             cpb.UserNotificationStatus(cpb.UserNotificationStatus_value[srcEnt.Status.String]),
		CreatedAt:          timestamppb.New(srcEnt.CreatedAt.Time),
		UpdatedAt:          timestamppb.New(srcEnt.UpdatedAt.Time),
		UserNotificationId: srcEnt.UserNotificationID.String,
	}
	return dstPb
}

func NotificationsFilteredToPb(notifications []*entities.InfoNotification, notiMsgMap map[string]*entities.InfoNotificationMsg, noificationsTags map[string]entities.InfoNotificationsTags) ([]*npb.GetNotificationsByFilterResponse_Notification, error) {
	notificationsPb := make([]*npb.GetNotificationsByFilterResponse_Notification, 0)
	for _, notiEnt := range notifications {
		targetGroupEnt, err := notiEnt.GetTargetGroup()
		if err != nil {
			return nil, fmt.Errorf("an error occurred when find target group of notification id: %v", notiEnt.NotificationID.String)
		}
		targetGroup := NotificationTargetToPb(targetGroupEnt)

		notiMsg, ok := notiMsgMap[notiEnt.NotificationID.String]

		if !ok {
			return nil, fmt.Errorf("expect find message of notification id: %v", notiEnt.NotificationID.String)
		}

		tagIDs := []string{}
		notiTags, ok := noificationsTags[notiEnt.NotificationID.String]
		if ok {
			for _, notiTag := range notiTags {
				tagIDs = append(tagIDs, notiTag.TagID.String)
			}
		}

		notiPb := &npb.GetNotificationsByFilterResponse_Notification{
			NotificationId:    notiEnt.NotificationID.String,
			NotificationMgsId: notiMsg.NotificationMsgID.String,
			Title:             notiMsg.Title.String,
			ComposerId:        notiEnt.EditorID.String,
			Status:            cpb.NotificationStatus(cpb.NotificationStatus_value[notiEnt.Status.String]),
			UserGroupFilter: &npb.GetNotificationsByFilterResponse_UserGroupFilter{
				UserGroups: targetGroup.UserGroupFilter.UserGroups,
			},
			UpdatedAt:   timestamppb.New(notiEnt.UpdatedAt.Time),
			SentAt:      timestamppb.New(notiEnt.SentAt.Time),
			TagIds:      tagIDs,
			TargetGroup: targetGroup,
		}

		notificationsPb = append(notificationsPb, notiPb)
	}
	return notificationsPb, nil
}

func NotificationGroupAudiencesToPb(audiences []*entities.Audience) []*npb.RetrieveGroupAudienceResponse_Audience {
	audiencesPb := make([]*npb.RetrieveGroupAudienceResponse_Audience, 0, len(audiences))
	for _, audience := range audiences {
		audiencesPb = append(audiencesPb, &npb.RetrieveGroupAudienceResponse_Audience{
			UserId:     audience.UserID.String,
			UserName:   audience.Name.String,
			Email:      audience.Email.String,
			Grade:      audience.GradeName.String,
			ChildNames: database.FromTextArray(audience.ChildNames),
		})
	}

	return audiencesPb
}

func NotificationDraftAudiencesToPb(audiences []*entities.Audience) []*npb.RetrieveDraftAudienceResponse_Audience {
	audiencesPb := make([]*npb.RetrieveDraftAudienceResponse_Audience, 0, len(audiences))
	for _, audience := range audiences {
		childName := ""
		childID := ""
		if childNames := database.FromTextArray(audience.ChildNames); len(childNames) > 0 {
			childName = childNames[0] // always 1 element
		}
		if childIDs := database.FromTextArray(audience.ChildIDs); len(childIDs) > 0 {
			childID = childIDs[0] // always 1 element
		}
		audiencesPb = append(audiencesPb, &npb.RetrieveDraftAudienceResponse_Audience{
			UserId:       audience.UserID.String,
			UserName:     audience.Name.String,
			UserGroup:    cpb.UserGroup(cpb.UserGroup_value[audience.UserGroup.String]),
			Email:        audience.Email.String,
			Grade:        audience.GradeName.String,
			IsIndividual: audience.IsIndividual.Bool,
			ChildName:    childName,
			ChildId:      childID,
		})
	}

	return audiencesPb
}
