package ptrls

import (
	"fmt"
	"go/token"

	"golang.org/x/tools/go/packages"
)

func Load(patterns ...string) (*Program, error) {
	cfg := &packages.Config{
		Fset: token.NewFileSet(),
		Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedTypesInfo |
			packages.NeedTypes | packages.NeedDeps | packages.NeedModule,
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("ptrls.Load: %w", err)
	}

	prog, err := buildSSA(cfg.Fset, pkgs)
	if err != nil {
		return nil, fmt.Errorf("ptrls.Load: %w", err)
	}

	return prog, nil
}
