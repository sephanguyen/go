package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/y-bash/go-gaga"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RemoveStrFromSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func LimitString(s string, limit int) string {
	// If the length of string is greater than limit, return the indexed string
	if len([]rune(s)) > limit {
		return string([]rune(s)[:limit])
	}

	return s
}

func AddPrefixString(s string, char string, count int) string {
	if count == 0 {
		return s
	}

	var builder strings.Builder

	for i := 0; i < count; i++ {
		builder.WriteString(char)
	}
	builder.WriteString(s)

	return builder.String()
}

func AddPrefixStringWithLimit(s string, char string, limit int) string {
	limitedString := LimitString(s, limit)

	// Determine how many prefix should be added to reach the limit
	charToPrefix := limit - len([]rune(s))

	newS := AddPrefixString(limitedString, char, charToPrefix)
	return newS
}

func AddSuffixString(s string, char string, count int) string {
	if count == 0 {
		return s
	}

	var builder strings.Builder
	builder.WriteString(s)

	for i := 0; i < count; i++ {
		builder.WriteString(char)
	}

	return builder.String()
}

func AddSuffixStringWithLimit(s string, char string, limit int) string {
	limitedString := LimitString(s, limit)

	// Determine how many suffix should be added to reach the limit
	charToSuffix := limit - len([]rune(s))
	newS := AddSuffixString(limitedString, char, charToSuffix)

	return newS
}

var (
	RegexHalfWidthKanaValidationNumbers          = "0123456789"
	RegexHalfWidthKanaValidationCapitalAlphabets = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	HalfWidthKanaValidationHalfWidthKatakana     = "ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃﾄﾅﾆﾇﾈﾉﾊﾋﾌﾍﾎﾏﾐﾑﾒﾓﾔﾕﾖﾗﾘﾙﾚﾛﾜﾝ"

	// RegexHalfWidthKanaValidationSymbols rule: ﾞﾟ()｢｣/-.\
	// but there is an escape for - and \
	RegexHalfWidthKanaValidationSymbols = `ﾞﾟ()｢｣/\-.\\ `
	InvalidHalfWidthKanaBankHolder      = status.Error(codes.InvalidArgument, "bank holder does not follow Half-width Kana rule")
)

func ValidateBankHolder(bankHolder string) error {
	halfWidthKanaValidation := fmt.Sprintf(
		"^[%s%s%s%s]+$",
		RegexHalfWidthKanaValidationNumbers,
		RegexHalfWidthKanaValidationCapitalAlphabets,
		HalfWidthKanaValidationHalfWidthKatakana,
		RegexHalfWidthKanaValidationSymbols,
	)

	regex := regexp.MustCompile(halfWidthKanaValidation)
	if !regex.MatchString(bankHolder) {
		return InvalidHalfWidthKanaBankHolder
	}
	return nil
}

type StringNormalizer struct {
	halfWidthNormalizer interface {
		String(s string) string
	}
	fullWidthNormalizer interface {
		String(s string) string
	}
}

func NewStringNormalizer() (*StringNormalizer, error) {
	halfWidthNormalizer, err := gaga.Norm(gaga.LatinToNarrow | gaga.DigitToNarrow | gaga.KatakanaToNarrow | gaga.SymbolToNarrow | gaga.KanaSymbolToNarrow)
	if err != nil {
		return nil, err
	}

	fullWidthNormalizer, err := gaga.Norm(gaga.LatinToWide | gaga.DigitToWide | gaga.KatakanaToWide | gaga.SymbolToWide | gaga.KanaSymbolToWide)
	if err != nil {
		return nil, err
	}

	return &StringNormalizer{
		halfWidthNormalizer: halfWidthNormalizer,
		fullWidthNormalizer: fullWidthNormalizer,
	}, nil
}

func (n *StringNormalizer) ToHalfWidth(s string) string {
	return n.halfWidthNormalizer.String(s)
}

func (n *StringNormalizer) ToFullWidth(s string) string {
	return n.fullWidthNormalizer.String(s)
}
