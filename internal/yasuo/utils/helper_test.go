package utils

import (
	"strings"
	"testing"

	bobpb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"

	"github.com/stretchr/testify/assert"
)

func TestPbSubject2String(t *testing.T) {
	t.Parallel()
	in := []bobpb.Subject{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 10}
	expected := []string{
		"SUBJECT_NONE",
		"SUBJECT_MATHS",
		"SUBJECT_BIOLOGY",
		"SUBJECT_PHYSICS",
		"SUBJECT_CHEMISTRY",
		"SUBJECT_GEOGRAPHY",
		"SUBJECT_ENGLISH",
		"SUBJECT_ENGLISH_2",
		"SUBJECT_JAPANESE",
		"SUBJECT_SCIENCE",
		"SUBJECT_LITERATURE",
		"SUBJECT_SOCIAL_STUDIES",
	}
	actual := PbSubject2String(in)
	assert.Equal(t, expected, actual)
}

func TestCustomUserClaims(t *testing.T) {
	t.Parallel()
	userGroup := "abcdef"
	userID := "qwerty"
	t.Run("schoolID=0", func(t *testing.T) {
		t.Parallel()
		actual := CustomUserClaims(userGroup, userID, 0)
		expected := map[string]interface{}{
			"https://hasura.io/jwt/claims": map[string]interface{}{
				"x-hasura-allowed-roles": []string{userGroup},
				"x-hasura-default-role":  userGroup,
				"x-hasura-user-id":       userID,
			},
		}
		assert.Equal(t, expected, actual)
	})

	t.Run("schoolID!=0", func(t *testing.T) {
		t.Parallel()
		actual := CustomUserClaims(userGroup, userID, 123)
		expected := map[string]interface{}{
			"https://hasura.io/jwt/claims": map[string]interface{}{
				"x-hasura-allowed-roles": []string{userGroup},
				"x-hasura-default-role":  userGroup,
				"x-hasura-user-id":       userID,
				"x-hasura-school-id":     "123",
			},
		}
		assert.Equal(t, expected, actual)
	})
}

func TestPbOrderStatus2String(t *testing.T) {
	t.Parallel()
	in := []pb.OrderStatus{0, 1, 2, 3, 4, 5, 6, 8, 7}
	expected := []string{
		"ORDER_STATUS_NONE",
		"ORDER_STATUS_WAITING_FOR_PAYMENT",
		"ORDER_STATUS_SUCCESSFULLY",
		"ORDER_STATUS_FAILED",
		"ORDER_STATUS_CANCELED",
		"ORDER_STATUS_ENDED",
		"ORDER_STATUS_PROCESSING_PAYMENT",
		"ORDER_STATUS_DISABLED",
		"ORDER_STATUS_DELETED",
	}
	actual := PbOrderStatus2String(in)
	assert.Equal(t, expected, actual)
}

func TestGenerateQuestionURL(t *testing.T) {
	t.Parallel()
	url1 := GenerateQuestionURL("testEvn", "testBucket", "testID")
	url2 := GenerateQuestionURL("testEvn", "testBucket", "testID")
	assert.True(t, strings.HasPrefix(url1, "testBucket/testEvn/testID/"), "expecting generated url have correct prefix")
	assert.True(t, strings.HasPrefix(url2, "testBucket/testEvn/testID/"), "expecting generated url have correct prefix")

	assert.True(t, strings.HasSuffix(url1, ".html"))
	assert.True(t, strings.HasSuffix(url2, ".html"))

	assert.NotEqual(t, url1, url2, "url must always generated randomly")
}

func TestReverseString(t *testing.T) {
	t.Parallel()
	in := "bròwn"
	expected := "nwòrb"
	actual := reserveString(in)
	assert.Equal(t, expected, actual)
}

func TestIsContain(t *testing.T) {
	t.Parallel()
	in := []string{"a", "b"}
	assert.True(t, IsContain(in, "a"))
	assert.True(t, IsContain(in, "b"))
	assert.False(t, IsContain(in, "c"))
}

func TestEncodeDecodeNumber(t *testing.T) {
	t.Parallel()
	var in uint = 123123123
	encoded := EncodeNumber2String(in)
	assert.NotEqual(t, in, encoded)
	decoded, err := DecodeString2Number(encoded)
	assert.NoError(t, err)
	assert.Equal(t, in, decoded)
}

func TestEncodeDecodeOrderID(t *testing.T) {
	t.Parallel()
	var in uint = 123123123
	encoded := EncodeOrderID2String(in)
	assert.NotEqual(t, in, encoded)
	decoded, err := DecodeString2OrderID(encoded)
	assert.NoError(t, err)
	assert.Equal(t, in, decoded)
}

func TestIsArrayMatch(t *testing.T) {
	t.Parallel()
	a := [6]int{1, 2, 0, 3, 4, 1}
	assert.True(t, IsArrayMatch(6, func(i int) bool {
		return a[i] != 9
	}))

	assert.False(t, IsArrayMatch(6, func(i int) bool {
		return a[i] == 5
	}))
}
