package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimitString(t *testing.T) {
	testCase := []struct {
		name     string
		given    string
		limit    int
		expected string
	}{
		{
			name:     "Given string length is less than the limit",
			given:    "test-string",
			limit:    100,
			expected: "test-string",
		},
		{
			name:     "Given string length is more than the limit",
			given:    "test-string",
			limit:    5,
			expected: "test-",
		},
		{
			name:     "Given string length is equal to the limit",
			given:    "test-string",
			limit:    11,
			expected: "test-string",
		},
		{
			name:     "Given string is a japanese character",
			given:    "ありがとう",
			limit:    3,
			expected: "ありが",
		},
	}

	for _, tc := range testCase {
		actual := LimitString(tc.given, tc.limit)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestAddPrefixString(t *testing.T) {
	testCase := []struct {
		name     string
		given    string
		prefix   string
		count    int
		expected string
	}{
		{
			name:     "Add '0' as the prefix 5 times to a string",
			given:    "test",
			prefix:   "0",
			expected: "00000test",
			count:    5,
		},
		{
			name:     "Add ' ' as the prefix 10 times to a string",
			given:    "test",
			prefix:   " ",
			expected: "          test",
			count:    10,
		},
		{
			name:     "Add '' as the prefix 3 times to a string",
			given:    "test",
			prefix:   "",
			expected: "test",
			count:    3,
		},
		{
			name:     "Add '+' as the prefix 3 times to a japanese character",
			given:    "ありがとう",
			prefix:   "+",
			expected: "+++ありがとう",
			count:    3,
		},
	}

	for _, tc := range testCase {
		actual := AddPrefixString(tc.given, tc.prefix, tc.count)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestAddPrefixStringWithLimit(t *testing.T) {
	testCase := []struct {
		name     string
		given    string
		prefix   string
		limit    int
		expected string
	}{
		{
			name:     "Add '0' to a string until it reaches the limit",
			given:    "test",
			prefix:   "0",
			expected: "000000test",
			limit:    10,
		},
		{
			name:     "Add '0' to a string when the limit is less than the length of given string",
			given:    "test",
			prefix:   "0",
			expected: "te",
			limit:    2,
		},
		{
			name:     "Add '0' to a string when the limit is equal to the length of given string",
			given:    "test",
			prefix:   "0",
			expected: "test",
			limit:    4,
		},
		{
			name:     "Add '+' to a japanese character until it reaches the limit",
			given:    "ありがとう",
			prefix:   "+",
			expected: "+++ありがとう",
			limit:    8,
		},
	}

	for _, tc := range testCase {
		actual := AddPrefixStringWithLimit(tc.given, tc.prefix, tc.limit)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestAddSuffixString(t *testing.T) {
	testCase := []struct {
		name     string
		given    string
		suffix   string
		count    int
		expected string
	}{
		{
			name:     "Add '0' as the suffix 5 times to a string",
			given:    "test",
			suffix:   "0",
			expected: "test00000",
			count:    5,
		},
		{
			name:     "Add ' ' as the suffix 10 times to a string",
			given:    "test",
			suffix:   " ",
			expected: "test          ",
			count:    10,
		},
		{
			name:     "Add '' as the suffix 3 times to a string",
			given:    "test",
			suffix:   "",
			expected: "test",
			count:    3,
		},
		{
			name:     "Add '+' as the suffix 3 times to a japanese character",
			given:    "ありがとう",
			suffix:   "+",
			expected: "ありがとう+++",
			count:    3,
		},
	}

	for _, tc := range testCase {
		actual := AddSuffixString(tc.given, tc.suffix, tc.count)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestAddSuffixStringWithLimit(t *testing.T) {
	testCase := []struct {
		name     string
		given    string
		suffix   string
		limit    int
		expected string
	}{
		{
			name:     "Add '0' to a string until it reaches the limit",
			given:    "test",
			suffix:   "0",
			expected: "test000000",
			limit:    10,
		},
		{
			name:     "Add '0' to a string when the limit is less than the length of given string",
			given:    "test",
			suffix:   "0",
			expected: "te",
			limit:    2,
		},
		{
			name:     "Add '0' to a string when the limit is equal to the length of given string",
			given:    "test",
			suffix:   "0",
			expected: "test",
			limit:    4,
		},
		{
			name:     "Add '+' to a japanese character until it reaches the limit",
			given:    "ありがとう",
			suffix:   "+",
			expected: "ありがとう+++",
			limit:    8,
		},
	}

	for _, tc := range testCase {
		actual := AddSuffixStringWithLimit(tc.given, tc.suffix, tc.limit)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestConvertStringToHalfWidth(t *testing.T) {
	normalizer, err := NewStringNormalizer()
	if err != nil {
		t.Fatal(err)
	}

	testCase := []struct {
		name        string
		given       string
		expected    string
		expectedErr error
	}{
		{
			name:     "Test Katakana char 1",
			given:    "サンプル",
			expected: "ｻﾝﾌﾟﾙ",
		},
		{
			name:     "Test Katakana char 2",
			given:    "アイウエオ",
			expected: "ｱｲｳｴｵ",
		},
		{
			name:     "Test Katakana char 3",
			given:    "カキクケコ",
			expected: "ｶｷｸｹｺ",
		},
		{
			name:     "Test Katakana char 4",
			given:    "サシスセソ",
			expected: "ｻｼｽｾｿ",
		},
		{
			name:     "Test Katakana char 5",
			given:    "スーパーマーケット",
			expected: "ｽｰﾊﾟｰﾏｰｹｯﾄ",
		},
		{
			name:     "Test Latin char",
			given:    "Ｈｅｌｌｏ，　ｗｏｒｌｄ！",
			expected: "Hello, world!",
		},
		{
			name:     "Test numeric char",
			given:    "１２３４５６７８９０",
			expected: "1234567890",
		},
		{
			name:     "Test currency format",
			given:    "１２３，３４５，７８９",
			expected: "123,345,789",
		},
		{
			name:     "Test combined char",
			given:    "キャキュキョ　１２３　ＡＢＣ",
			expected: "ｷｬｷｭｷｮ 123 ABC",
		},
		{
			name:     "Test special character",
			given:    "！＠＃＄％＾＆＊（）",
			expected: "!@#$%^&*()",
		},
		{
			name:     "Test JP special character",
			given:    "。【】",
			expected: "｡【】",
		},
	}

	for _, tc := range testCase {
		actual := normalizer.ToHalfWidth(tc.given)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestConvertStringToFullWidth(t *testing.T) {
	normalizer, err := NewStringNormalizer()
	if err != nil {
		t.Fatal(err)
	}

	testCase := []struct {
		name        string
		given       string
		expected    string
		expectedErr error
	}{
		{
			name:     "Test Katakana char 1",
			given:    "ｻﾝﾌﾟﾙ",
			expected: "サンフﾟル",
		},
		{
			name:     "Test Katakana char 2",
			given:    "ｱｲｳｴｵ",
			expected: "アイウエオ",
		},
		{
			name:     "Test Katakana char 3",
			given:    "ｶｷｸｹｺ",
			expected: "カキクケコ",
		},
		{
			name:     "Test Katakana char 4",
			given:    "ｻｼｽｾｿ",
			expected: "サシスセソ",
		},
		{
			name:     "Test Katakana char 5",
			given:    "ｽｰﾊﾟｰﾏｰｹｯﾄ",
			expected: "スーハﾟーマーケット",
		},
		{
			name:     "Test Latin char",
			given:    "Hello, world!",
			expected: "Ｈｅｌｌｏ，　ｗｏｒｌｄ！",
		},
		{
			name:     "Test numeric char",
			given:    "1234567890",
			expected: "１２３４５６７８９０",
		},
		{
			name:     "Test currency format",
			given:    "123,345,789",
			expected: "１２３，３４５，７８９",
		},
		{
			name:     "Test combined char",
			given:    "ｷｬｷｭｷｮ 123 ABC",
			expected: "キャキュキョ　１２３　ＡＢＣ",
		},
		{
			name:     "Test special character",
			given:    "!@#$%^&*()",
			expected: "！＠＃＄％＾＆＊（）",
		},
		{
			name:     "Test JP special character",
			given:    "｡【】",
			expected: "。【】",
		},
	}

	for _, tc := range testCase {
		actual := normalizer.ToFullWidth(tc.given)
		assert.Equal(t, tc.expected, actual)
	}
}
