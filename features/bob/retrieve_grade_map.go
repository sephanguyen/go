package bob

import (
	"context"
	"fmt"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) studentRetrievesGradeMap(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.Conn).RetrieveGradeMap(s.signedCtx(ctx), &pb.RetrieveGradeMapRequest{})
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobReturnAllGradeMap(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveGradeMapResponse)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}
	gradeMap := rsp.GradeMap
	var testCases = []struct {
		country pb.Country
		in      []string
		out     []int
	}{
		{
			pb.COUNTRY_VN,
			[]string{"Lớp 1", "Lớp 2", "Lớp 3", "Lớp 4", "Lớp 5", "Lớp 6", "Lớp 7", "Lớp 8", "Lớp 9", "Lớp 10", "Lớp 11", "Lớp 12"},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			pb.COUNTRY_MASTER,
			[]string{"Grade 1", "Grade 2", "Grade 3", "Grade 4", "Grade 5", "Grade 6", "Grade 7", "Grade 8", "Grade 9", "Grade 10", "Grade 11", "Grade 12"},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
		{
			pb.COUNTRY_JP,
			[]string{"小学1年生", "小学2年生", "小学3年生", "小学4年生", "小学5年生", "小学6年生", "中学1年生", "中学2年生", "中学3年生", "高校1年生", "高校2年生", "高校3年生"},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		},
	}
	for _, tt := range testCases {
		for i, grade := range tt.in {
			localGrade := gradeMap[tt.country.String()]
			localGradeMap := localGrade.LocalGrade
			cGrade := localGradeMap[grade]
			if cGrade != int32(tt.out[i]) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("%d is not equal %d for country %s", cGrade, tt.out[i], tt.country.String())
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
