package mock

import (
	"context"
	"fmt"
)

// TODO: autogen mock from port

type LessonPortMockIsLessonMediumOnlineType func(ctx context.Context, lessonID string) (bool, error)
type LessonPortMockCheckLessonMemberIDsType func(ctx context.Context, lessonID string, userIDs []string) (memberIDs []string, err error)

type LessonPortMock struct {
	isLessonMediumOnline []LessonPortMockIsLessonMediumOnlineType
	checkLessonMemberIDs []LessonPortMockCheckLessonMemberIDsType
}

func (u *LessonPortMock) IsLessonMediumOnline(ctx context.Context, lessonID string) (bool, error) {
	fn := u.isLessonMediumOnline[0]
	u.isLessonMediumOnline = u.isLessonMediumOnline[1:]
	return fn(ctx, lessonID)
}

func (u *LessonPortMock) SetIsLessonMediumOnline(fn LessonPortMockIsLessonMediumOnlineType, num int) {
	if num < 1 {
		num = 1
	}
	if u.isLessonMediumOnline == nil {
		u.isLessonMediumOnline = make([]LessonPortMockIsLessonMediumOnlineType, 0, num)
	}

	for i := 0; i < num; i++ {
		u.isLessonMediumOnline = append(u.isLessonMediumOnline, fn)
	}
}

func (u *LessonPortMock) CheckLessonMemberIDs(ctx context.Context, lessonID string, userIDs []string) (memberIDs []string, err error) {
	fn := u.checkLessonMemberIDs[0]
	u.checkLessonMemberIDs = u.checkLessonMemberIDs[1:]
	return fn(ctx, lessonID, userIDs)
}

func (u *LessonPortMock) SetCheckLessonMemberIDs(fn LessonPortMockCheckLessonMemberIDsType, num int) {
	if num < 1 {
		num = 1
	}
	if u.checkLessonMemberIDs == nil {
		u.checkLessonMemberIDs = make([]LessonPortMockCheckLessonMemberIDsType, 0, num)
	}

	for i := 0; i < num; i++ {
		u.checkLessonMemberIDs = append(u.checkLessonMemberIDs, fn)
	}
}

func (u *LessonPortMock) AllFuncBeCalledAsExpected() error {
	if num := len(u.isLessonMediumOnline); num > 0 {
		return fmt.Errorf("IsLessonMediumOnline func still have %d time called", num)
	}

	if num := len(u.checkLessonMemberIDs); num > 0 {
		return fmt.Errorf("CheckLessonMemberIDs func still have %d time called", num)
	}

	return nil
}
