package speeches

import cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

// Languagues
var (
	EN = "en-US"
	JP = "ja-JP"
)

var mapper = map[string]string{
	cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG.String(): EN,
	cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_JP.String():  JP,
	// new config
	cpb.QuizItemAttributeConfig_LANGUAGE_CONFIG_ENG.String():            EN,
	cpb.QuizItemAttributeConfig_LANGUAGE_CONFIG_JP.String():             JP,
	cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_NONE.String(): "",
}

// WhiteList for generating audio files
var WhiteList = []string{EN}

func GetLanguage(input []string) string {
	for _, each := range input {
		if val, ok := mapper[each]; ok {
			return val
		}
	}
	return ""
}
