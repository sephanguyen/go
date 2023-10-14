package analysisgo

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

type PackageIterator struct {
	pkg          *Package
	filter       *ObjectFilter
	visitedIndex int
}

func NewPackageIterator(pkg *Package, ops ...ObjFilterOption) (*PackageIterator, error) {
	filter := &ObjectFilter{}
	for _, op := range ops {
		op(filter)
	}

	if pkg.Syntax == nil || pkg.Types == nil || pkg.Fset == nil || pkg.TypesInfo == nil {
		return nil, fmt.Errorf("package have some nil fields of these Syntax, Types, Fset, TypesInfo")
	}

	return &PackageIterator{
		filter:       filter,
		pkg:          pkg,
		visitedIndex: -1,
	}, nil
}

func (pi *PackageIterator) GetNext() *FileIterator {
	pi.visitedIndex++
	if !(pi.visitedIndex < len(pi.pkg.Syntax)) {
		return nil
	}

	return &FileIterator{
		fs:                 pi.pkg.Fset,
		f:                  pi.pkg.Syntax[pi.visitedIndex],
		scope:              pi.pkg.Types.Scope(),
		info:               pi.pkg.TypesInfo,
		filter:             pi.filter,
		numberVisitedNodes: 0,
		visitedNodes:       make(map[ast.Node]bool),
	}
}

type FileIterator struct {
	fs    *token.FileSet
	f     *ast.File
	scope *types.Scope
	info  *types.Info

	filter             *ObjectFilter
	numberVisitedNodes int
	visitedNodes       map[ast.Node]bool
}

func NewFileIterator(fs *token.FileSet, f *ast.File, scope *types.Scope, info *types.Info, ops ...ObjFilterOption) *FileIterator {
	filter := &ObjectFilter{}
	for _, op := range ops {
		op(filter)
	}

	return &FileIterator{
		fs:           fs,
		f:            f,
		scope:        scope,
		info:         info,
		filter:       filter,
		visitedNodes: make(map[ast.Node]bool),
	}
}

// GetNext will return an object. If return a nil object, travel ends.
// If return an error, this object maybe don't supported yet, just keep going travel.
// Objects are being supported: a constant declare, a variable declare have a constant value,
//a func declare or a method declare of a struct type.
func (fi *FileIterator) GetNext() (*Object, error) {
	result := &Object{}

	// object which need find is a var
	var object *ast.Ident
	// object which need find is a function or method
	var funcObj *ast.FuncDecl
	stopped := false
	nodeIndex := 0
	ast.Inspect(fi.f, func(n ast.Node) bool {
		if stopped {
			return false
		}
		defer func() {
			nodeIndex++
		}()
		if fi.visited(n) {
			return true
		}
		if n == nil {
			return true
		}

		matched, sign, _ := fi.matchFilter(n)
		if !matched {
			return true
		}
		result.sign = sign
		stopped = true
		switch node := n.(type) {
		case *ast.FuncDecl:
			// get a method
			funcObj = node
			return false
		case *ast.Ident:
			if node.Obj != nil {
				switch v := node.Obj.Decl.(type) {
				case *ast.FuncDecl:
					// get a func
					funcObj = v
					return false
				case *ast.ValueSpec:
					// get a variable
					object = node
					if len(v.Values) == 0 {
						break
					}
					if tv, ok := fi.info.Types[v.Values[0]]; ok {
						if tv.Value != nil {
							result.value = tv.Value
						}
					}
					return false
				case *ast.AssignStmt:
					// get a constant
					object = node
					if len(v.Rhs) == 0 {
						break
					}
					if tv, ok := fi.info.Types[v.Rhs[0]]; ok {
						if tv.Value != nil {
							result.value = tv.Value
						}
					}
					return false
				}
			}
		}
		stopped = false
		result.sign = ""

		return true
	})
	// TODO: simplify way to checking visited all node
	// visited all node
	if fi.numberVisitedNodes == nodeIndex {
		return nil, nil
	}
	fi.numberVisitedNodes = nodeIndex

	if !stopped {
		return fi.GetNext()
	}

	// Look up that name in the innermost scope to get types.Object
	if funcObj != nil {
		if funcObj.Recv == nil {
			// is a func
			obj, err := fi.lookupObject(funcObj.Name.Name, funcObj.Name.Pos(), fi.scope.Innermost(funcObj.Pos()))
			if err != nil {
				return nil, fmt.Errorf("lookup func: %w", err)
			}
			result.Object = obj
		} else {
			// is a method
			for _, recv := range funcObj.Recv.List {
				// get struct type which owner this method
				str, ok := recv.Type.(*ast.Ident)
				if !ok {
					continue
				}

				// lookup method object by struct type
				obj, err := fi.lookupObjectMethod(str.Name, funcObj.Name.Name, str.Obj.Pos(), fi.scope.Innermost(str.Pos()))
				if err != nil {
					return nil, fmt.Errorf("lookup method: %w", err)
				}
				structTypeObj, err := fi.lookupObject(str.Name, str.Obj.Pos(), fi.scope.Innermost(str.Pos()))
				if err != nil {
					return nil, fmt.Errorf("lookup owner of method: %w", err)
				}
				result.Object = obj
				result.Owner = structTypeObj
				break
			}
			if result.Object == nil {
				return nil, fmt.Errorf("could not find struct type of method %s", funcObj.Name.Name)
			}
		}
	} else {
		obj, err := fi.lookupObject(object.Name, object.Pos(), fi.scope.Innermost(object.Pos()))
		if err != nil {
			return nil, fmt.Errorf("lookup object: %w", err)
		}
		result.Object = obj
	}

	if result.Object == nil {
		return nil, fmt.Errorf("could not look up object by name %s", object.Name)
	}

	return result, nil
}

func (fi *FileIterator) visited(n ast.Node) bool {
	if n == nil {
		return false
	}

	if _, ok := fi.visitedNodes[n]; ok {
		return true
	}
	fi.visitedNodes[n] = true

	return false
}

func (fi *FileIterator) matchFilter(n ast.Node) (matched bool, sign, objName string) {
	if len(fi.filter.signComments) != 0 {
		nodePos := fi.fs.Position(n.Pos())
		// get comment above this node
		// TODO: persist index to reduce running time
		signComment, _ := fi.getSign(nodePos, 0)
		if len(signComment) == 0 {
			return false, "", ""
		}
		sign = signComment
	}

	switch id := n.(type) {
	case *ast.Ident:
		// get variable name
		objName = id.Name
	case *ast.FuncDecl:
		// get func/method name
		objName = id.Name.Name
	default:
		return false, "", ""
	}
	if fi.filter.matchName != nil && !fi.filter.matchName(objName) {
		return false, "", ""
	}

	return true, sign, objName
}

// getSign will return sign comment above a target position, begin from comment at index in list comments
func (fi *FileIterator) getSign(targetPos token.Position, index int) (string, *ast.CommentGroup) {
	if !(index < len(fi.f.Comments)) {
		return "", nil
	}

	commentGroup := fi.f.Comments[index]
	commentPosition := fi.fs.Position(commentGroup.End())
	if commentPosition.Line > targetPos.Line {
		// passed through target
		return "", nil
	}
	if commentPosition.Line != targetPos.Line-1 {
		// go to next comment
		return fi.getSign(targetPos, index+1)
	}

	comments := strings.Split(commentGroup.Text(), "\n")
	for _, cmt := range comments {
		cmt = strings.TrimSpace(cmt)
		if ok, sign := fi.isSignComment(cmt); ok {
			return sign, commentGroup
		}
	}

	return "", nil
}

func (fi *FileIterator) isSignComment(comment string) (bool, string) {
	s := strings.Split(comment, " ")
	if len(s) == 0 {
		return false, ""
	}
	if ok := fi.filter.signComments[s[0]]; ok {
		return true, s[0]
	}
	return false, ""
}

// lookupObject will return an object by name and pos
func (fi *FileIterator) lookupObject(name string, pos token.Pos, inner *types.Scope) (types.Object, error) {
	// Look up that object name in the innermost scope
	_, obj := inner.LookupParent(name, 0)
	if obj == nil || obj.Pos() != pos {
		return nil, fmt.Errorf("could not find object %s", name)
	}

	return obj, nil
}

// lookupObjectMethod will return a method object by its struct type and pos
func (fi *FileIterator) lookupObjectMethod(structObjName, methodName string, structPos token.Pos, inner *types.Scope) (types.Object, error) {
	// Look up that method of struct type object in the innermost scope
	_, obj := inner.LookupParent(structObjName, 0)
	if obj == nil || obj.Pos() != structPos {
		return nil, fmt.Errorf("could not find struct type %s", structObjName)
	}

	if str, ok := obj.Type().(*types.Named); ok {
		for i := 0; i < str.NumMethods(); i++ {
			method := str.Method(i)
			if method.Name() == methodName {
				return method, nil
			}
		}
	} else {
		return nil, fmt.Errorf("object %s is not a struct type", structObjName)
	}

	return nil, fmt.Errorf("struct type %s have no method %s", structObjName, methodName)
}
