package mock

import (
	"context"
	"fmt"
)

// TODO: autogen mock from port

type UserModulePortMock struct {
	checkExistedUserIDs []func(ctx context.Context, ids []string) (existed []string, err error)
}

func (u *UserModulePortMock) CheckExistedUserIDs(ctx context.Context, ids []string) (existed []string, err error) {
	fn := u.checkExistedUserIDs[0]
	u.checkExistedUserIDs = u.checkExistedUserIDs[1:]
	return fn(ctx, ids)
}

func (u *UserModulePortMock) SetCheckExistedUserIDs(fn func(ctx context.Context, ids []string) (existed []string, err error), num int) {
	if num < 1 {
		num = 1
	}
	if u.checkExistedUserIDs == nil {
		u.checkExistedUserIDs = make([]func(ctx context.Context, ids []string) (existed []string, err error), 0, num)
	}

	for i := 0; i < num; i++ {
		u.checkExistedUserIDs = append(u.checkExistedUserIDs, fn)
	}
}

func (u *UserModulePortMock) AllFuncBeCalledAsExpected() error {
	if num := len(u.checkExistedUserIDs); num > 0 {
		return fmt.Errorf("CheckExistedUserIDs func still have %d time called", num)
	}

	return nil
}
