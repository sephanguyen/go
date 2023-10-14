package godogutil

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"

	"go.uber.org/multierr"
)

// chained is like func1, func1 arg1, func1 arg2, func2, func2 arg1, func3(no arg), func4(no arg)
// it detect if a func receive context as first argument, then it will pass a context which is renew
// after func call
// DOES NOT SUPPORT FUNCTION WITH VARIADIC PARAMS
func MultiErrChain(ctx context.Context, chained ...interface{}) (context.Context, error) {
	var retErrs = []error{}
	i := 0
	for i < len(chained) {
		currentI := chained[i]
		if reflect.TypeOf(currentI).Kind() == reflect.Func {
			args := []interface{}{ctx}
			j := i + 1
			for j < len(chained) {
				currentJ := chained[j]
				if reflect.TypeOf(currentJ).Kind() == reflect.Func {
					break
				}
				args = append(args, currentJ)
				j++
			}
			var ret []reflect.Value
			var er error
			if len(args) == 0 {
				ret, er = Call(currentI)
			} else {
				ret, er = Call(currentI, args...)
			}
			if er != nil {
				panic(er)
			}
			// return err
			if len(ret) == 1 {
				// if not nil then append error
				if !ret[0].IsNil() {
					err := ret[0].Interface().(error)
					retErrs = append(retErrs, err)
					if err != nil {
						return ctx, fmt.Errorf("%s %w", GetFunctionName(currentI), err)
					}
				}
			} else if len(ret) == 2 { //return ctx, err
				ctx = ret[0].Interface().(context.Context)
				if !ret[1].IsNil() {
					err := ret[1].Interface().(error)
					retErrs = append(retErrs, err)
					if err != nil {
						return ctx, fmt.Errorf("%s %w", GetFunctionName(currentI), err)
					}
				}
			}

			i = j
		} else {
			i++
		}
	}

	return ctx, multierr.Combine(retErrs...)
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// param[0] always is a chained context, some func does not support context, check if its NumIn == len(param) -1
func Call(someFunc interface{}, params ...interface{}) (result []reflect.Value, err error) {
	f := reflect.ValueOf(someFunc)
	numInput := f.Type().NumIn()
	switch {
	case len(params)-1 == numInput:
		params = params[1:]
	case len(params) == numInput:
	default:
		err = errors.New("The number of params is not adapted.")
		return
	}

	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return
}
