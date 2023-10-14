package i18n

import (
	"testing"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/stretchr/testify/assert"
)

func TestConvertStringGradeToInt(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		country pb.Country
		in      []string
		out     []int
	}{
		{
			pb.COUNTRY_VN,
			[]string{
				"G1", "G2", "G3", "G4", "G5", "G6", "G7", "G8", "G9", "G10", "G11", "G12",
				"Lớp 1", "Lớp 2", "Lớp 3", "Lớp 4", "Lớp 5", "Lớp 6", "Lớp 7", "Lớp 8", "Lớp 9", "Lớp 10", "Lớp 11", "Lớp 12",
				"CĐ/ĐH năm 1", "CĐ/ĐH năm 2", "CĐ/ĐH năm 3", "CĐ/ĐH năm 4", "Khác",
			},
			[]int{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
				13, 14, 15, 16, 0,
			},
		},
		{
			pb.COUNTRY_MASTER,
			[]string{"Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5", "Grade 6", "Grade 7", "Grade 8", "Grade 9", "Grade 10", "Grade 11", "Grade 12"},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			pb.COUNTRY_ID,
			[]string{"G1", "G2", "G3", "G4", "G5", "G6", "G7", "G8", "G9", "G10", "G11", "G12"},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			pb.COUNTRY_JP,
			[]string{
				"小学1年生", "小学2年生", "小学3年生", "小学4年生", "小学5年生", "小学6年生", "中学1年生", "中学2年生", "中学3年生", "高校1年生", "高校2年生", "高校3年生",
				"大学1年生", "大学2年生", "大学3年生", "大学4年生", "その他",
			},
			[]int{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
				13, 14, 15, 16, 0,
			},
		},
		{
			pb.COUNTRY_SG,
			[]string{"G1", "G2", "G3", "G4", "G5", "G6", "G7", "G8", "G9", "G10", "G11", "G12"},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
	}
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.country.String(), func(t *testing.T) {
			t.Parallel()
			for i, grade := range tt.in {
				cGrade, err := ConvertStringGradeToInt(tt.country, grade)
				assert.Nil(t, err)
				assert.Equal(t, tt.out[i], cGrade)
			}
		})
	}

	t.Run("invalid country", func(t *testing.T) {
		t.Parallel()
		_, err := ConvertStringGradeToInt(pb.COUNTRY_NONE, "G1")
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = cannot find country grade map")
	})

	t.Run("invalid grade", func(t *testing.T) {
		t.Parallel()
		_, err := ConvertStringGradeToInt(pb.COUNTRY_ID, "INVALID")
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = cannot find grade in map")
	})
}

func TestConvertIntGradeToString(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		country pb.Country
		in      []int
		out     []string
	}{
		{
			pb.COUNTRY_VN,
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 0},
			[]string{
				"Lớp 1", "Lớp 2", "Lớp 3", "Lớp 4", "Lớp 5", "Lớp 6", "Lớp 7", "Lớp 8", "Lớp 9", "Lớp 10", "Lớp 11", "Lớp 12",
				"CĐ/ĐH năm 1", "CĐ/ĐH năm 2", "CĐ/ĐH năm 3", "CĐ/ĐH năm 4", "Khác",
			},
		},
		{
			pb.COUNTRY_MASTER,
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			[]string{"Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5", "Grade 6", "Grade 7", "Grade 8", "Grade 9", "Grade 10", "Grade 11", "Grade 12"},
		},
		{
			pb.COUNTRY_ID,
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			[]string{"G1", "G2", "G3", "G4", "G5", "G6", "G7", "G8", "G9", "G10", "G11", "G12"},
		},
		{
			pb.COUNTRY_SG,
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			[]string{"G1", "G2", "G3", "G4", "G5", "G6", "G7", "G8", "G9", "G10", "G11", "G12"},
		},
		{
			pb.COUNTRY_JP,
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 0},
			[]string{
				"小学1年生", "小学2年生", "小学3年生", "小学4年生", "小学5年生", "小学6年生", "中学1年生", "中学2年生", "中学3年生", "高校1年生", "高校2年生", "高校3年生",
				"大学1年生", "大学2年生", "大学3年生", "大学4年生", "その他",
			},
		},
	}
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.country.String(), func(t *testing.T) {
			t.Parallel()
			for i, grade := range tt.in {
				cGrade, err := ConvertIntGradeToString(tt.country, grade)
				assert.Nil(t, err)
				assert.Equal(t, tt.out[i], cGrade)
			}
		})
	}
	_, err := ConvertIntGradeToString(pb.COUNTRY_NONE, 1)
	assert.Error(t, err)
}

func TestConvertIntGradeToStringV1(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		country cpb.Country
		in      []int
		out     []string
	}{
		{
			cpb.Country_COUNTRY_VN,
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 0},
			[]string{
				"Lớp 1", "Lớp 2", "Lớp 3", "Lớp 4", "Lớp 5", "Lớp 6", "Lớp 7", "Lớp 8", "Lớp 9", "Lớp 10", "Lớp 11", "Lớp 12",
				"CĐ/ĐH năm 1", "CĐ/ĐH năm 2", "CĐ/ĐH năm 3", "CĐ/ĐH năm 4", "Khác",
			},
		},
		{
			cpb.Country_COUNTRY_MASTER,
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			[]string{"Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5", "Grade 6", "Grade 7", "Grade 8", "Grade 9", "Grade 10", "Grade 11", "Grade 12"},
		},
		{
			cpb.Country_COUNTRY_ID,
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			[]string{"G1", "G2", "G3", "G4", "G5", "G6", "G7", "G8", "G9", "G10", "G11", "G12"},
		},
		{
			cpb.Country_COUNTRY_SG,
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			[]string{"G1", "G2", "G3", "G4", "G5", "G6", "G7", "G8", "G9", "G10", "G11", "G12"},
		},
		{
			cpb.Country_COUNTRY_JP,
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 0},
			[]string{
				"小学1年生", "小学2年生", "小学3年生", "小学4年生", "小学5年生", "小学6年生", "中学1年生", "中学2年生", "中学3年生", "高校1年生", "高校2年生", "高校3年生",
				"大学1年生", "大学2年生", "大学3年生", "大学4年生", "その他",
			},
		},
	}
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.country.String(), func(t *testing.T) {
			t.Parallel()
			for i, grade := range tt.in {
				cGrade, err := ConvertIntGradeToStringV1(tt.country, grade)
				assert.Nil(t, err)
				assert.Equal(t, tt.out[i], cGrade)
			}
		})
	}
	_, err := ConvertIntGradeToStringV1(cpb.Country_COUNTRY_NONE, 1)
	assert.Error(t, err)
}

func TestConvertStringToSubject(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		in  string
		out pb.Subject
	}{
		{"SUBJECT_NONE", pb.Subject(0)},
		{"SUBJECT_MATHS", pb.Subject(1)},
		{"SUBJECT_BIOLOGY", pb.Subject(2)},
		{"SUBJECT_PHYSICS", pb.Subject(3)},
		{"SUBJECT_CHEMISTRY", pb.Subject(4)},
		{"SUBJECT_GEOGRAPHY", pb.Subject(5)},
		{"SUBJECT_ENGLISH", pb.Subject(6)},
		{"SUBJECT_ENGLISH_2", pb.Subject(7)},
		{"SUBJECT_JAPANESE", pb.Subject(8)},
		{"SUBJECT_SCIENCE", pb.Subject(9)},
		{"SUBJECT_SOCIAL_STUDIES", pb.Subject(10)},
		{"SUBJECT_LITERATURE", pb.Subject(11)},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			actual := ConvertStringToSubject(tc.in)
			assert.Equal(t, tc.out, actual)
		})
	}
}
