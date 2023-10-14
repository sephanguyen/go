package analysisgo

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
	"strings"
)

type ObjFilterOption func(f *ObjectFilter)

func ObjWithSignComments(signs map[string]bool) ObjFilterOption {
	return func(f *ObjectFilter) {
		f.signComments = signs
	}
}

func ObjWithName(matchName func(name string) bool) ObjFilterOption {
	return func(f *ObjectFilter) {
		f.matchName = matchName
	}
}

type ObjectFilter struct {
	signComments map[string]bool
	matchName    func(name string) bool
}

func (o *ObjectIterator) matchFilter(n ast.Node) (matched bool, sign, objName string) {
	if len(o.filter.signComments) != 0 {
		nodePos := o.gf.fs.Position(n.Pos())
		// get comment above this node
		// TODO: persist index to reduce running time
		signComment, _ := o.getSign(nodePos, 0)
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
	if o.filter.matchName != nil && !o.filter.matchName(objName) {
		return false, "", ""
	}

	return true, sign, objName
}

type ObjectIterator struct {
	gf                 *GoFile
	pkg                *types.Package
	info               *types.Info
	filter             *ObjectFilter
	numberVisitedNodes int
	visitedNode        ast.Node
}

func NewObjectIterator(gf *GoFile, ops ...ObjFilterOption) (*ObjectIterator, error) {
	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
	pkg, err := conf.Check("cmd/hello", gf.fs, []*ast.File{gf.f}, info)
	if err != nil {
		return nil, fmt.Errorf("conf.Check: %w", err)
	}

	filter := &ObjectFilter{}
	for _, op := range ops {
		op(filter)
	}

	return &ObjectIterator{
		gf:     gf,
		filter: filter,
		pkg:    pkg,
		info:   info,
	}, nil
}

// TODO: simplify way to checking object which is visited
func (o *ObjectIterator) visited(n ast.Node, nodeIndex int) bool {
	if nodeIndex < o.numberVisitedNodes {
		return true
	}
	if o.visitedNode == nil || n == nil {
		return false
	}

	if o.gf.fs.Position(n.Pos()).Line > o.gf.fs.Position(o.visitedNode.Pos()).Line {
		return false
	}

	return true
}

// GetNext will return an object. If return a nil object, travel ends.
// If return an error, this object maybe don't supported yet, just keep going travel.
// Objects are being supported: a constant declare, a variable declare have a constant value,
//a func declare or a method declare of a struct type.
func (o *ObjectIterator) GetNext() (*Object, error) {
	result := &Object{}

	// object which need find is Ident
	var object *ast.Ident
	// object which need find is a funcDecl (function or method)
	var funcObj *ast.FuncDecl
	stopped := false
	var matchedNode ast.Node
	nodeIndex := 0
	ast.Inspect(o.gf.f, func(n ast.Node) bool {
		if o.visited(n, nodeIndex) {
			nodeIndex++
			return true
		}
		if stopped {
			// TODO: check what the happen when return false
			return false
		}
		defer func() {
			nodeIndex++
		}()
		// TODO: why n is able to nil
		if n == nil {
			return true
		}

		if matchedNode == nil {
			matched, sign, _ := o.matchFilter(n)
			if !matched {
				return true
			}
			matchedNode = n
			result.sign = sign
			// is a funcDecl (function or method), stop travel
			switch node := matchedNode.(type) {
			case *ast.FuncDecl:
				// is method
				funcObj = node
				stopped = true
				return false
			case *ast.Ident:
				object = node
				// checking is a func, if not it is a variable
				if node.Obj != nil {
					if v, ok := node.Obj.Decl.(*ast.FuncDecl); ok {
						funcObj = v
						stopped = true
						return false
					}
				}
			}
		} else if o.gf.fs.Position(n.Pos()).Line > o.gf.fs.Position(matchedNode.Pos()).Line {
			// passed through matched node, stop travel
			return false
		}

		// if matched node is variable, get variable's value
		if obj, ok := n.(ast.Expr); ok {
			if tv, ok := o.info.Types[obj]; ok {
				if tv.Value != nil {
					result.value = tv.Value
				}
			}
		}

		// stop travel
		if object != nil && result.value != nil {
			stopped = true
			return false
		}
		return true
	})
	// visited all node
	if o.numberVisitedNodes == nodeIndex {
		return nil, nil
	}
	o.visitedNode = matchedNode
	o.numberVisitedNodes = nodeIndex
	if !stopped {
		return o.GetNext()
	}

	// Look up that name in the innermost scope to get types.Object
	if funcObj != nil {
		if funcObj.Recv == nil {
			// is a func
			obj, err := o.lookupObject(funcObj.Name.Name, funcObj.Name.Pos(), o.pkg.Scope().Innermost(funcObj.Pos()))
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
				obj, err := o.lookupObjectMethod(str.Name, funcObj.Name.Name, str.Obj.Pos(), o.pkg.Scope().Innermost(str.Pos()))
				if err != nil {
					return nil, fmt.Errorf("lookup method: %w", err)
				}
				structTypeObj, err := o.lookupObject(str.Name, str.Obj.Pos(), o.pkg.Scope().Innermost(str.Pos()))
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
		obj, err := o.lookupObject(object.Name, object.Pos(), o.pkg.Scope().Innermost(object.Pos()))
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

func (o *ObjectIterator) isSignComment(comment string) (bool, string) {
	s := strings.Split(comment, " ")
	if len(s) == 0 {
		return false, ""
	}
	if ok := o.filter.signComments[s[0]]; ok {
		return true, s[0]
	}
	return false, ""
}

// getSign will return sign comment above a target position, begin from comment at index in list comments
func (o *ObjectIterator) getSign(targetPos token.Position, index int) (string, *ast.CommentGroup) {
	if !(index < len(o.gf.f.Comments)) {
		return "", nil
	}

	commentGroup := o.gf.f.Comments[index]
	commentPosition := o.gf.fs.Position(commentGroup.End())
	if commentPosition.Line > targetPos.Line {
		// passed through target
		return "", nil
	}
	if commentPosition.Line != targetPos.Line-1 {
		// go to next comment
		return o.getSign(targetPos, index+1)
	}

	comments := strings.Split(commentGroup.Text(), "\n")
	for _, cmt := range comments {
		cmt = strings.TrimSpace(cmt)
		if ok, sign := o.isSignComment(cmt); ok {
			return sign, commentGroup
		}
	}

	return "", nil
}

// lookupObject will return an object by name and pos
func (o *ObjectIterator) lookupObject(name string, pos token.Pos, inner *types.Scope) (types.Object, error) {
	// Look up that object name in the innermost scope
	_, obj := inner.LookupParent(name, 0)
	if obj == nil || obj.Pos() != pos {
		return nil, fmt.Errorf("could not find object %s", name)
	}

	return obj, nil
}

// lookupObjectMethod will return a method object by its struct type and pos
func (o *ObjectIterator) lookupObjectMethod(structObjName, methodName string, structPos token.Pos, inner *types.Scope) (types.Object, error) {
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
