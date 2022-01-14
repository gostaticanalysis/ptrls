package ptrls

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/pointer"
	"golang.org/x/tools/go/ssa"
)

func Analyze(prog *Program, poses ...token.Pos) (map[ast.Node]pointer.Pointer, error) {
	config := &pointer.Config{
		Mains: []*ssa.Package{prog.Main},
	}

	ptrs := make(map[ast.Node]pointer.Pointer)
	value2node := make(map[ssa.Value]ast.Node)
	for _, pos := range poses {
		path, exact := prog.Path(pos)
		if !exact {
			p := prog.Fset.Position(pos)
			return nil, fmt.Errorf("cannot AST node from %s", p)
		}

		expr, _ := path[0].(ast.Expr)
		typ := prog.TypesInfo.TypeOf(expr)
		if !pointer.CanPoint(typ) {
			// skip
			continue
		}

		v := getValue(prog, path, expr)
		if v == nil {
			continue
		}

		value2node[v] = expr
		config.AddQuery(v)
	}

	result, err := pointer.Analyze(config)
	if err != nil {
		return nil, err
	}

	for v, p := range result.Queries {
		n := value2node[v]
		ptrs[n] = p
	}

	return ptrs, nil
}

func getValue(prog *Program, path []ast.Node, expr ast.Expr) ssa.Value {
	f := ssa.EnclosingFunction(prog.Main, path)
	if f != nil {
		v, _ := f.ValueForExpr(expr)
		if v != nil {
			return v
		}
	}

	var id *ast.Ident
	switch expr := expr.(type) {
	case *ast.Ident:
		id = expr
	case *ast.SelectorExpr:
		id = expr.Sel
	}

	obj, _ := prog.TypesInfo.ObjectOf(id).(*types.Var)
	if obj != nil {
		v, _ := prog.SSA.VarValue(obj, prog.Main, path)
		if v != nil {
			return v
		}
	}

	return nil
}
