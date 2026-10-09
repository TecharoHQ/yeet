// Package filter decides which package builds run based on a CEL expression.
package filter

import (
	"fmt"

	"cel.dev/cel-go/cel"
)

// Filter matches package builds against a compiled CEL expression.
type Filter struct {
	prg cel.Program
}

// New compiles expr into a Filter. The expression can use the string
// variables method, goos, and goarch. An empty expression matches every
// build.
func New(expr string) (*Filter, error) {
	if expr == "" {
		return &Filter{}, nil
	}

	env, err := cel.NewEnv(
		cel.Variable("method", cel.StringType),
		cel.Variable("goos", cel.StringType),
		cel.Variable("goarch", cel.StringType),
	)
	if err != nil {
		return nil, fmt.Errorf("can't create CEL environment: %w", err)
	}

	ast, iss := env.Compile(expr)
	if err := iss.Err(); err != nil {
		return nil, fmt.Errorf("invalid expression: %w", err)
	}

	if !ast.OutputType().IsExactType(cel.BoolType) {
		return nil, fmt.Errorf("expression must evaluate to a bool, got: %s", ast.OutputType())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("can't create CEL program: %w", err)
	}

	return &Filter{prg: prg}, nil
}

// Match reports whether a build with the given method, goos, and goarch
// passes the filter.
func (f *Filter) Match(method, goos, goarch string) (bool, error) {
	if f.prg == nil {
		return true, nil
	}

	out, _, err := f.prg.Eval(map[string]any{
		"method": method,
		"goos":   goos,
		"goarch": goarch,
	})
	if err != nil {
		return false, fmt.Errorf("can't evaluate filter expression: %w", err)
	}

	return out.Value() == true, nil
}
