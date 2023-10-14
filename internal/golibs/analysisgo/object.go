package analysisgo

import (
	"go/constant"
	"go/types"
)

type Object struct {
	types.Object
	value constant.Value
	Owner types.Object // if object is a method, it is struct type which owner this method
	sign  string
}

func (o *Object) Value() constant.Value {
	return o.value
}

func (o *Object) Sign() string {
	return o.sign
}

// IsFunc will return a function, concrete method, or abstract method
func (o *Object) IsFunc() *types.Func {
	if v, ok := o.Object.(*types.Func); ok {
		return v
	}

	return nil
}

// IsVar will return a variable, parameter, result, or struct field
func (o *Object) IsVar() *types.Var {
	if v, ok := o.Object.(*types.Var); ok {
		return v
	}

	return nil
}

// IsConst will return a constant
func (o *Object) IsConst() *types.Const {
	if v, ok := o.Object.(*types.Const); ok {
		return v
	}

	return nil
}

// IsTypeName will return a type name
func (o *Object) IsTypeName() *types.TypeName {
	if v, ok := o.Object.(*types.TypeName); ok {
		return v
	}

	return nil
}

// IsLabel will return a statement label
func (o *Object) IsLabel() *types.Label {
	if v, ok := o.Object.(*types.Label); ok {
		return v
	}

	return nil
}

// IsPkgName will return a package name, e.g. json after import "encoding/json"
func (o *Object) IsPkgName() *types.PkgName {
	if v, ok := o.Object.(*types.PkgName); ok {
		return v
	}

	return nil
}

// IsBuiltin will return a predeclared function such as append or len
func (o *Object) IsBuiltin() *types.Builtin {
	if v, ok := o.Object.(*types.Builtin); ok {
		return v
	}

	return nil
}

// IsNil will return a predeclared nil
func (o *Object) IsNil() *types.Nil {
	if v, ok := o.Object.(*types.Nil); ok {
		return v
	}

	return nil
}
