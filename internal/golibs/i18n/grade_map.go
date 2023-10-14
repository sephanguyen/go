package i18n

import (
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	TotalAllowedGrades = []int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	InGradeMap = map[pb.Country]map[string]int{
		pb.COUNTRY_ID: {
			"G1":     1,
			"G2":     2,
			"G3":     3,
			"G4":     4,
			"G5":     5,
			"G6":     6,
			"G7":     7,
			"G8":     8,
			"G9":     9,
			"G10":    10,
			"G11":    11,
			"G12":    12,
			"U1":     13,
			"U2":     14,
			"U3":     15,
			"U4":     16,
			"Others": 0,
		},
		pb.COUNTRY_SG: {
			"G1":     1,
			"G2":     2,
			"G3":     3,
			"G4":     4,
			"G5":     5,
			"G6":     6,
			"G7":     7,
			"G8":     8,
			"G9":     9,
			"G10":    10,
			"G11":    11,
			"G12":    12,
			"U1":     13,
			"U2":     14,
			"U3":     15,
			"U4":     16,
			"Others": 0,
		},
		pb.COUNTRY_JP: {
			"小学1年生": 1,
			"小学2年生": 2,
			"小学3年生": 3,
			"小学4年生": 4,
			"小学5年生": 5,
			"小学6年生": 6,
			"中学1年生": 7,
			"中学2年生": 8,
			"中学3年生": 9,
			"高校1年生": 10,
			"高校2年生": 11,
			"高校3年生": 12,
			"大学1年生": 13,
			"大学2年生": 14,
			"大学3年生": 15,
			"大学4年生": 16,
			"その他":   0,
		},
		pb.COUNTRY_VN: {
			"G1":          1,
			"G2":          2,
			"G3":          3,
			"G4":          4,
			"G5":          5,
			"G6":          6,
			"G7":          7,
			"G8":          8,
			"G9":          9,
			"G10":         10,
			"G11":         11,
			"G12":         12,
			"Lớp 1":       1,
			"Lớp 2":       2,
			"Lớp 3":       3,
			"Lớp 4":       4,
			"Lớp 5":       5,
			"Lớp 6":       6,
			"Lớp 7":       7,
			"Lớp 8":       8,
			"Lớp 9":       9,
			"Lớp 10":      10,
			"Lớp 11":      11,
			"Lớp 12":      12,
			"CĐ/ĐH năm 1": 13,
			"CĐ/ĐH năm 2": 14,
			"CĐ/ĐH năm 3": 15,
			"CĐ/ĐH năm 4": 16,
			"Khác":        0,
		},
		pb.COUNTRY_MASTER: {
			"Grade 1":      1,
			"Grade 2":      2,
			"Grade 3":      3,
			"Grade 4":      4,
			"Grade 5":      5,
			"Grade 6":      6,
			"Grade 7":      7,
			"Grade 8":      8,
			"Grade 9":      9,
			"Grade 10":     10,
			"Grade 11":     11,
			"Grade 12":     12,
			"University 1": 13,
			"University 2": 14,
			"University 3": 15,
			"University 4": 16,
			"Others":       0,
		},
	}
	OutGradeMap = map[pb.Country]map[int]string{
		pb.COUNTRY_SG: {
			1:  "G1",
			2:  "G2",
			3:  "G3",
			4:  "G4",
			5:  "G5",
			6:  "G6",
			7:  "G7",
			8:  "G8",
			9:  "G9",
			10: "G10",
			11: "G11",
			12: "G12",
			13: "U1",
			14: "U2",
			15: "U3",
			16: "U4",
			0:  "Others",
		},
		pb.COUNTRY_JP: {
			1:  "小学1年生",
			2:  "小学2年生",
			3:  "小学3年生",
			4:  "小学4年生",
			5:  "小学5年生",
			6:  "小学6年生",
			7:  "中学1年生",
			8:  "中学2年生",
			9:  "中学3年生",
			10: "高校1年生",
			11: "高校2年生",
			12: "高校3年生",
			13: "大学1年生",
			14: "大学2年生",
			15: "大学3年生",
			16: "大学4年生",
			0:  "その他",
		},
		pb.COUNTRY_ID: {
			1:  "G1",
			2:  "G2",
			3:  "G3",
			4:  "G4",
			5:  "G5",
			6:  "G6",
			7:  "G7",
			8:  "G8",
			9:  "G9",
			10: "G10",
			11: "G11",
			12: "G12",
			13: "U1",
			14: "U2",
			15: "U3",
			16: "U4",
			0:  "Others",
		},
		pb.COUNTRY_VN: {
			1:  "Lớp 1",
			2:  "Lớp 2",
			3:  "Lớp 3",
			4:  "Lớp 4",
			5:  "Lớp 5",
			6:  "Lớp 6",
			7:  "Lớp 7",
			8:  "Lớp 8",
			9:  "Lớp 9",
			10: "Lớp 10",
			11: "Lớp 11",
			12: "Lớp 12",
			13: "CĐ/ĐH năm 1",
			14: "CĐ/ĐH năm 2",
			15: "CĐ/ĐH năm 3",
			16: "CĐ/ĐH năm 4",
			0:  "Khác",
		},
		pb.COUNTRY_MASTER: {
			1:  "Grade 1",
			2:  "Grade 2",
			3:  "Grade 3",
			4:  "Grade 4",
			5:  "Grade 5",
			6:  "Grade 6",
			7:  "Grade 7",
			8:  "Grade 8",
			9:  "Grade 9",
			10: "Grade 10",
			11: "Grade 11",
			12: "Grade 12",
			13: "University 1",
			14: "University 2",
			15: "University 3",
			16: "University 4",
			0:  "Others",
		},
	}

	// For proto v1
	OutGradeMapV1 = map[cpb.Country]map[int]string{
		cpb.Country_COUNTRY_SG: {
			1:  "G1",
			2:  "G2",
			3:  "G3",
			4:  "G4",
			5:  "G5",
			6:  "G6",
			7:  "G7",
			8:  "G8",
			9:  "G9",
			10: "G10",
			11: "G11",
			12: "G12",
			13: "U1",
			14: "U2",
			15: "U3",
			16: "U4",
			0:  "Others",
		},
		cpb.Country_COUNTRY_JP: {
			1:  "小学1年生",
			2:  "小学2年生",
			3:  "小学3年生",
			4:  "小学4年生",
			5:  "小学5年生",
			6:  "小学6年生",
			7:  "中学1年生",
			8:  "中学2年生",
			9:  "中学3年生",
			10: "高校1年生",
			11: "高校2年生",
			12: "高校3年生",
			13: "大学1年生",
			14: "大学2年生",
			15: "大学3年生",
			16: "大学4年生",
			0:  "その他",
		},
		cpb.Country_COUNTRY_ID: {
			1:  "G1",
			2:  "G2",
			3:  "G3",
			4:  "G4",
			5:  "G5",
			6:  "G6",
			7:  "G7",
			8:  "G8",
			9:  "G9",
			10: "G10",
			11: "G11",
			12: "G12",
			13: "U1",
			14: "U2",
			15: "U3",
			16: "U4",
			0:  "Others",
		},
		cpb.Country_COUNTRY_VN: {
			1:  "Lớp 1",
			2:  "Lớp 2",
			3:  "Lớp 3",
			4:  "Lớp 4",
			5:  "Lớp 5",
			6:  "Lớp 6",
			7:  "Lớp 7",
			8:  "Lớp 8",
			9:  "Lớp 9",
			10: "Lớp 10",
			11: "Lớp 11",
			12: "Lớp 12",
			13: "CĐ/ĐH năm 1",
			14: "CĐ/ĐH năm 2",
			15: "CĐ/ĐH năm 3",
			16: "CĐ/ĐH năm 4",
			0:  "Khác",
		},
		cpb.Country_COUNTRY_MASTER: {
			1:  "Grade 1",
			2:  "Grade 2",
			3:  "Grade 3",
			4:  "Grade 4",
			5:  "Grade 5",
			6:  "Grade 6",
			7:  "Grade 7",
			8:  "Grade 8",
			9:  "Grade 9",
			10: "Grade 10",
			11: "Grade 11",
			12: "Grade 12",
			13: "University 1",
			14: "University 2",
			15: "University 3",
			16: "University 4",
			0:  "Others",
		},
	}
)

// ConvertStringGradeToInt converts gradeStr to the corresponding integer code.
func ConvertStringGradeToInt(country pb.Country, gradeStr string) (int, error) {
	localMap := InGradeMap[country]
	if len(localMap) == 0 {
		return -1, status.Error(codes.InvalidArgument, "cannot find country grade map")
	}

	grade, ok := localMap[gradeStr]
	if !ok {
		return -1, status.Error(codes.InvalidArgument, "cannot find grade in map")
	}

	return grade, nil
}

// ConvertIntGradeToString converts grade to the corresponding string representation.
func ConvertIntGradeToString(country pb.Country, grade int) (string, error) {
	localMap := OutGradeMap[country]

	if len(localMap) == 0 {
		return "", status.Error(codes.InvalidArgument, "cannot find country grade map")
	}
	return localMap[grade], nil
}

// ConvertIntGradeToStringV1 converts grade to the corresponding string representation.
func ConvertIntGradeToStringV1(country cpb.Country, grade int) (string, error) {
	localMap := OutGradeMapV1[country]

	if len(localMap) == 0 {
		return "", status.Error(codes.InvalidArgument, "cannot find country grade map")
	}
	return localMap[grade], nil
}

// ConvertStringToSubject converts a subject's string code to pb.Subject.
func ConvertStringToSubject(s string) pb.Subject {
	subject := pb.Subject(pb.Subject_value[s])
	return subject
}

func ConvertTextArrayToPlanPrivileges(p pgtype.TextArray) []pb.PlanPrivilege {
	value := make([]pb.PlanPrivilege, len(p.Elements))

	for _, v := range p.Elements {
		var a pb.PlanPrivilege
		v.AssignTo(&a)
		value = append(value, a)
	}
	return value
}
