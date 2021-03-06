package ptrls

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
)

type Program struct {
	SSA       *ssa.Program
	Main      *ssa.Package
	SrcFuncs  []*ssa.Function
	Fset      *token.FileSet
	TypesInfo *types.Info
	Files     []*ast.File
}

func buildSSA(fset *token.FileSet, pkgs []*packages.Package) (*Program, error) {
	mode := ssa.GlobalDebug// | ssa.NaiveForm
	prog := ssa.NewProgram(fset, mode)

	// Create SSA packages for all imports.
	// Order is not significant.
	created := make(map[*types.Package]bool)
	var createAll func(pkgs []*types.Package)
	createAll = func(pkgs []*types.Package) {
		for _, p := range pkgs {
			if !created[p] {
				created[p] = true
				prog.CreatePackage(p, nil, nil, true)
				createAll(p.Imports())
			}
		}
	}

	var mainPkg *packages.Package
	var files []*ast.File
	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
		InitOrder:  []*types.Initializer{},
	}
	for _, pkg := range pkgs {
		createAll(pkg.Types.Imports())
		mergeTypesInfo(info, pkg.TypesInfo)
		files = append(files, pkg.Syntax...)
		if pkg.Module != nil && pkg.Module.Main {
			mainPkg = pkg
		}
	}

	if mainPkg == nil {
		return nil, errors.New("cannot find main module")
	}

	ssapkg := prog.CreatePackage(mainPkg.Types, files, info, true)
	ssapkg.Build()

	var funcs []*ssa.Function
	for _, f := range files {
		for _, decl := range f.Decls {
			if fdecl, ok := decl.(*ast.FuncDecl); ok {

				if fdecl.Name.Name == "_" {
					continue
				}

				fn := info.Defs[fdecl.Name].(*types.Func)
				if fn == nil {
					return nil, fmt.Errorf("cannot get an object: %s", fdecl.Name.Name)
				}

				f := ssapkg.Prog.FuncValue(fn)
				if f == nil {
					return nil, fmt.Errorf("cannot get a ssa function: %s", fdecl.Name.Name)
				}

				var addAnons func(f *ssa.Function)
				addAnons = func(f *ssa.Function) {
					funcs = append(funcs, f)
					for _, anon := range f.AnonFuncs {
						addAnons(anon)
					}
				}
				addAnons(f)
			}
		}
	}

	return &Program{
		SSA:       ssapkg.Prog,
		Main:      ssapkg,
		SrcFuncs:  funcs,
		Fset:      fset,
		TypesInfo: info,
		Files:     files,
	}, nil
}

func mergeTypesInfo(x, y *types.Info) {
	for k, v := range y.Types {
		x.Types[k] = v
	}
	for k, v := range y.Defs {
		x.Defs[k] = v
	}
	for k, v := range y.Uses {
		x.Uses[k] = v
	}
	for k, v := range y.Implicits {
		x.Implicits[k] = v
	}
	for k, v := range y.Selections {
		x.Selections[k] = v
	}
	x.InitOrder = append(x.InitOrder, y.InitOrder...)
}

func (prog *Program) Pos(filename string, offset int) token.Pos {
	var pos token.Pos
	prog.Fset.Iterate(func(f *token.File) bool {
		if f.Name() == filename {
			pos = f.Pos(offset)
			return false
		}
		return true
	})
	return pos
}

func (prog *Program) Path(pos token.Pos) (path []ast.Node, exact bool) {
	for _, f := range prog.Files {
		if f.Pos() <= pos && pos <= f.End() {
			return astutil.PathEnclosingInterval(f, pos, pos)
		}
	}
	return nil, false
}
