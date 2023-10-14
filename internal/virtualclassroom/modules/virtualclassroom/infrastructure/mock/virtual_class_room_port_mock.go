package mock

import (
	"context"
	"fmt"
)

// TODO: autogen mock from port

type VirtualClassRoomPortPortMock struct {
	isExistent []func(ctx context.Context, id string) (bool, error)
}

func (v *VirtualClassRoomPortPortMock) IsExistent(ctx context.Context, id string) (bool, error) {
	fn := v.isExistent[0]
	v.isExistent = v.isExistent[1:]
	return fn(ctx, id)
}

func (v *VirtualClassRoomPortPortMock) SetIsExistent(fn func(ctx context.Context, id string) (bool, error), num int) {
	if num < 1 {
		num = 1
	}
	if v.isExistent == nil {
		v.isExistent = make([]func(ctx context.Context, id string) (bool, error), 0, num)
	}

	for i := 0; i < num; i++ {
		v.isExistent = append(v.isExistent, fn)
	}
}

func (v *VirtualClassRoomPortPortMock) AllFuncBeCalledAsExpected() error {
	if num := len(v.isExistent); num > 0 {
		return fmt.Errorf("IsExistent func still have %d time called", num)
	}

	return nil
}
