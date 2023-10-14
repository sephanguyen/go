package helpers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
)

func (helper *CommunicationHelper) ParseQuestionTemplateFromString(str string) []*npb.QuestionnaireTemplateQuestion {
	parts := strings.Split(str, ",")
	ret := []*npb.QuestionnaireTemplateQuestion{}
	for _, item := range parts {
		trimmed := strings.TrimSpace(item)
		regex := regexp.MustCompile(`^([0-9]+)\.([^\.]*)(.*)$`)
		matches := regex.FindStringSubmatch(trimmed)
		isrequired := false
		id := matches[1]
		qttype := matches[2]
		var qType cpb.QuestionType
		var choices []string
		switch qttype {
		case "free_text":
			qType = cpb.QuestionType_QUESTION_TYPE_FREE_TEXT
		case "check_box":
			for i := 0; i < 3; i++ {
				choices = append(choices, idutil.ULIDNow())
			}
			qType = cpb.QuestionType_QUESTION_TYPE_CHECK_BOX
		case "multiple_choice":
			qType = cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE
			for i := 0; i < 3; i++ {
				choices = append(choices, idutil.ULIDNow())
			}
		default:
			panic(fmt.Errorf("unknown question type %s", qttype))
		}
		if len(matches) == 4 && matches[3] == ".required" {
			isrequired = true
		}
		order, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			panic(fmt.Errorf("failed to parse order index from %s", trimmed))
		}
		question := &npb.QuestionnaireTemplateQuestion{
			OrderIndex: order,
			Title:      idutil.ULIDNow(),
			Type:       qType,
			Required:   isrequired,
			Choices:    choices,
		}

		ret = append(ret, question)
	}
	return ret
}

func (helper *CommunicationHelper) MapQuestionnaireQuestionToQuestionnaireTemplateQuestion(questions []*cpb.Question) []*npb.QuestionnaireTemplateQuestion {
	templateQuestions := []*npb.QuestionnaireTemplateQuestion{}

	for _, question := range questions {
		templateQuestion := &npb.QuestionnaireTemplateQuestion{
			OrderIndex: question.OrderIndex,
			Title:      question.Title,
			Type:       question.Type,
			Required:   question.Required,
			Choices:    question.Choices,
		}

		templateQuestions = append(templateQuestions, templateQuestion)
	}

	return templateQuestions
}
