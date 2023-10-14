package communication

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

func parseQuestionFromString(str string) []*cpb.Question {
	parts := strings.Split(str, ",")
	ret := []*cpb.Question{}
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
		question := &cpb.Question{
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
