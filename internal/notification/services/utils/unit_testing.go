package utils

//nolint
import (
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func GenSampleNotification() *cpb.Notification {
	userID := idutil.ULIDNow()
	infoNotification := &cpb.Notification{
		Data:        `{"promote": [{"code": "ABC123", "type": "Prime", "amount": "30", "expired_at": "2020-02-13 05:35:44.657508"}, {"code": "ABC123", "type": "Basic", "amount": "20", "expired_at": "2020-03-23 05:35:44.657508"}], "image_url": "https://manabie.com/f50cffe1a8068b04a1b05d1a13b60642.png"}`,
		EditorId:    userID,
		ReceiverIds: []string{},
		Message: &cpb.NotificationMessage{
			Title: "🎁 Em được gửi tặng 1 món quà đặc biệt! 🎁",
			Content: &cpb.RichText{
				Raw:      `{"blocks":[{"key":"74q28","text":"2 sau khi update prepare for deployment","type":"unstyled","depth":0,"inlineStyleRanges":[],"entityRanges":[],"data":{}}],"entityMap":{}}`,
				Rendered: `https://storage.googleapis.com/stag-manabie-backend/content/b4e13f4263ebab160895abcb982d8cf8.html`,
			},
		},
		Type:   cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED,
		Event:  cpb.NotificationEvent_NOTIFICATION_EVENT_NONE,
		Status: cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT,
		TargetGroup: &cpb.NotificationTargetGroup{
			CourseFilter: &cpb.NotificationTargetGroup_CourseFilter{
				Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
				CourseIds: []string{"course_id_1"},
			},
			GradeFilter: &cpb.NotificationTargetGroup_GradeFilter{
				Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
				GradeIds: []string{"grade-id"},
			},
			UserGroupFilter: &cpb.NotificationTargetGroup_UserGroupFilter{
				UserGroups: []cpb.UserGroup{cpb.UserGroup_USER_GROUP_STUDENT, cpb.UserGroup_USER_GROUP_PARENT},
			},
			LocationFilter: &cpb.NotificationTargetGroup_LocationFilter{
				Type:        cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
				LocationIds: []string{"location_id_1"},
			},
			ClassFilter: &cpb.NotificationTargetGroup_ClassFilter{
				Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
				ClassIds: []string{},
			},
			SchoolFilter: &cpb.NotificationTargetGroup_SchoolFilter{
				Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
				SchoolIds: []string{"school_id"},
			},
		},
		SchoolId:    constant.ManabieSchool,
		IsImportant: false,
	}

	return infoNotification
}

// nolint
func GenQuestionaire() entities.Questionnaire {
	var ret entities.Questionnaire
	database.AllRandomEntity(&ret)
	ret.QuestionnaireID.Set(idutil.ULIDNow())
	return ret
}

func GenQNUserAnswers(userID string, targetID string) entities.QuestionnaireUserAnswers {
	ret := make([]*entities.QuestionnaireUserAnswer, 0, 5)
	for i := 0; i < 5; i++ {
		newEnt := GenQNUserAnswer(userID, targetID, idutil.ULIDNow())
		ret = append(ret, &newEnt)
	}
	return ret
}

func GenQNUserAnswer(userID string, targetID string, questionID string) entities.QuestionnaireUserAnswer {
	return entities.QuestionnaireUserAnswer{
		AnswerID:                database.Text(idutil.ULIDNow()),
		UserNotificationID:      database.Text(idutil.ULIDNow()),
		UserID:                  database.Text(userID),
		TargetID:                database.Text(targetID),
		Answer:                  database.Text(idutil.ULIDNow()),
		QuestionnaireQuestionID: database.Text(questionID),
		SubmittedAt:             database.Timestamptz(time.Now()),
	}
}

func GenQNQuestions(qnID string) (ret entities.QuestionnaireQuestions) {
	for i := 0; i < 3; i++ {
		var temp entities.QuestionnaireQuestion
		database.AllRandomEntity(&temp)
		// nolint
		temp.QuestionnaireID = database.Text(qnID)
		temp.OrderIndex = database.Int4(int32(i))
		temp.Choices = database.TextArray([]string{idutil.ULIDNow(), idutil.ULIDNow(), idutil.ULIDNow()})
		switch i {
		case 0:
			temp.Type = database.Text(cpb.QuestionType_QUESTION_TYPE_CHECK_BOX.String())
		case 1:
			temp.Type = database.Text(cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE.String())
		case 2:
			temp.Type = database.Text(cpb.QuestionType_QUESTION_TYPE_FREE_TEXT.String())
			temp.Choices = database.TextArray([]string{})
		}
		ret = append(ret, &temp)
	}
	return
}

// nolint
func GenSampleNotificationWithMsg() (*entities.InfoNotification, *entities.InfoNotificationMsg) {
	notiMsgEnt := GenNotificationMsgEntity()
	notiEnt := GenNotificationEntity()
	notiEnt.NotificationMsgID.Set(notiMsgEnt.NotificationMsgID)

	return &notiEnt, &notiMsgEnt
}

func GenNotificationEntity() entities.InfoNotification {
	noti := entities.InfoNotification{
		NotificationID:    database.Text("notification_id_1"),
		NotificationMsgID: database.Text("notification_msg_id_1"),
		Type:              database.Text(cpb.NotificationType_NOTIFICATION_TYPE_COMPOSED.String()),
		Data:              database.JSONB("hello world"),
		EditorID:          database.Text("editor_id"),
		TargetGroups: database.JSONB(entities.InfoNotificationTarget{
			CourseFilter: entities.InfoNotificationTarget_CourseFilter{
				Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST.String(),
				CourseIDs: []string{"course_id_1", "course_id_2"},
			},
			GradeFilter: entities.InfoNotificationTarget_GradeFilter{
				Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST.String(),
				GradeIDs: []string{idutil.ULIDNow(), idutil.ULIDNow()},
			},
			UserGroupFilter: entities.InfoNotificationTarget_UserGroupFilter{
				UserGroups: []string{cpb.UserGroup_USER_GROUP_PARENT.String(), cpb.UserGroup_USER_GROUP_STUDENT.String()},
			},
			ClassFilter: entities.InfoNotificationTarget_ClassFilter{
				Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST.String(),
				ClassIDs: []string{"class_id_1", "class_id_2"},
			},
			LocationFilter: entities.InfoNotificationTarget_LocationFilter{
				Type:        cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST.String(),
				LocationIDs: []string{"location_id_1", "location_id_2"},
			},
			SchoolFilter: entities.InfoNotificationTarget_SchoolFilter{
				Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST.String(),
				SchoolIDs: []string{"school_id_1", "school_id_2"},
			},
		}),
		ReceiverIDs: database.TextArray([]string{"receiver_id_1", "receiver_id_2"}),
		Event:       database.Text(cpb.NotificationEvent_NOTIFICATION_EVENT_NONE.String()),
		Status:      database.Text(cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String()),
		ScheduledAt: database.Timestamptz(time.Now()),
		Owner:       database.Int4(constants.ManabieSchool),
		CreatedAt:   database.Timestamptz(time.Now().Add(-2 * time.Hour)),
		DeletedAt:   database.Timestamptz(time.Now().Add(-2 * time.Hour)),
	}

	return noti
}

func GenNotificationMsgEntity() entities.InfoNotificationMsg {
	notiMsg := entities.InfoNotificationMsg{
		NotificationMsgID: database.Text("notification_msg_id_1"),
		Title:             database.Text("title of notification"),
		Content: database.JSONB(entities.RichText{
			Raw:         `raw rich text`,
			RenderedURL: `rendered url html`,
		}),
		MediaIDs:  database.TextArray([]string{"media_id_1", "media_id_2"}),
		CreatedAt: database.Timestamptz(time.Now().Add(-2 * time.Hour)),
		DeletedAt: database.Timestamptz(time.Now().Add(-2 * time.Hour)),
	}

	return notiMsg
}

func GenUserNotificationEntity() entities.UserInfoNotification {
	notiMsg := entities.UserInfoNotification{
		NotificationID: database.Text("notification_id_1"),
		UserID:         database.Text("user_id_1"),
		Status:         database.Text(cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NEW.String()),
		Courses:        database.TextArray([]string{"course_id_1", "course_id_2"}),
		CurrentGrade:   database.Int2(12),
		CreatedAt:      database.Timestamptz(time.Now().Add(-2 * time.Hour)),
		DeletedAt:      database.Timestamptz(time.Now().Add(-2 * time.Hour)),
	}

	return notiMsg
}
func GenSampleQuestionnaire() *cpb.Questionnaire {
	questionnaire := &cpb.Questionnaire{
		Questions: []*cpb.Question{
			{
				QuestionnaireQuestionId: idutil.ULIDNow(),
				Title:                   idutil.ULIDNow(),
				Type:                    cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE,
				Choices:                 []string{idutil.ULIDNow(), idutil.ULIDNow()},
				OrderIndex:              1,
				Required:                true,
			},
			{
				QuestionnaireQuestionId: idutil.ULIDNow(),
				Title:                   idutil.ULIDNow(),
				Type:                    cpb.QuestionType_QUESTION_TYPE_CHECK_BOX,
				Choices:                 []string{idutil.ULIDNow(), idutil.ULIDNow()},
				OrderIndex:              2,
				Required:                true,
			},
			{
				QuestionnaireQuestionId: idutil.ULIDNow(),
				Title:                   idutil.ULIDNow(),
				Type:                    cpb.QuestionType_QUESTION_TYPE_FREE_TEXT,
				Choices:                 []string{},
				OrderIndex:              3,
				Required:                true,
			},
		},
		ExpirationDate:  timestamppb.New(time.Now().Add(3 * time.Minute)),
		QuestionnaireId: idutil.ULIDNow(),
		ResubmitAllowed: true,
	}

	return questionnaire
}

func GenSampleAnswersForQuestionnaire(qn *cpb.Questionnaire) []*cpb.Answer {
	answers := make([]*cpb.Answer, 0, len(qn.Questions))
	for _, ques := range qn.Questions {
		switch ques.Type {
		case cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE:
			answers = append(answers, &cpb.Answer{
				QuestionnaireQuestionId: ques.QuestionnaireQuestionId,
				Answer:                  ques.Choices[0],
			})
		case cpb.QuestionType_QUESTION_TYPE_CHECK_BOX:
			answers = append(answers, &cpb.Answer{
				QuestionnaireQuestionId: ques.QuestionnaireQuestionId,
				Answer:                  ques.Choices[0],
			})
			answers = append(answers, &cpb.Answer{
				QuestionnaireQuestionId: ques.QuestionnaireQuestionId,
				Answer:                  ques.Choices[1],
			})
		case cpb.QuestionType_QUESTION_TYPE_FREE_TEXT:
			answers = append(answers, &cpb.Answer{
				QuestionnaireQuestionId: ques.QuestionnaireQuestionId,
				Answer:                  idutil.ULIDNow(),
			})
		}
	}
	return answers
}

// nolint
func GetMD5String(s string) string {
	h := md5.New()
	_, err := io.WriteString(h, s)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func GenInfoNotificationTagBulkInsert() []*entities.InfoNotificationTag {
	n := rand.Intn(10-1) + 1
	records := make([]*entities.InfoNotificationTag, 0, n)

	for i := 0; i < n; i++ {
		record := &entities.InfoNotificationTag{
			NotificationTagID: database.Text(idutil.ULIDNow()),
			NotificationID:    database.Text(fmt.Sprintf("lo_id_%d", i)),
			TagID:             database.Text(fmt.Sprintf("study_plan_item_id_%d", i)),
		}
		records = append(records, record)
	}
	return records
}

func GetSampleQuestionnaireTemplate() *npb.QuestionnaireTemplate {
	questionnaireTemplate := &npb.QuestionnaireTemplate{
		QuestionnaireTemplateId: idutil.ULIDNow(),
		Name:                    "Questionnaire Template 1",
		ResubmitAllowed:         true,
		ExpirationDate:          timestamppb.New(time.Now().Add(3 * time.Minute)),
		Questions: []*npb.QuestionnaireTemplateQuestion{
			{
				QuestionnaireTemplateQuestionId: idutil.ULIDNow(),
				Title:                           idutil.ULIDNow(),
				Type:                            cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE,
				Choices:                         []string{idutil.ULIDNow(), idutil.ULIDNow()},
				OrderIndex:                      1,
				Required:                        true,
			},
			{
				QuestionnaireTemplateQuestionId: idutil.ULIDNow(),
				Title:                           idutil.ULIDNow(),
				Type:                            cpb.QuestionType_QUESTION_TYPE_CHECK_BOX,
				Choices:                         []string{idutil.ULIDNow(), idutil.ULIDNow()},
				OrderIndex:                      2,
				Required:                        true,
			},
			{
				QuestionnaireTemplateQuestionId: idutil.ULIDNow(),
				Title:                           idutil.ULIDNow(),
				Type:                            cpb.QuestionType_QUESTION_TYPE_FREE_TEXT,
				Choices:                         []string{},
				OrderIndex:                      3,
				Required:                        true,
			},
		},
	}

	return questionnaireTemplate
}
