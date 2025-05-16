// A generated module for Pokedexcli functions

package main

import (
	"context"
	"runtime"
	"dagger/pokedexcli/internal/dagger"
)

type Pokedexcli struct{
	// +private
	Source *dagger.Directory
	// +private
	Repo string
}


func New(
	// +defaultPath="/"
	// +ignore=[".git"]
	source *dagger.Directory,
	// +optional
	// +default="github.com/luanmtruong/pokedexcli"
	repo string,
) *Pokedexcli {
	return &Pokedexcli{
		Source:   source,
		Repo:     repo,
	}
}

// Run Test
func (p *Pokedexcli) Test(ctx context.Context) (string, error) {
	return dag.Golang().
		WithSource(p.Source).
		Test(ctx)
}

// Run Lint
func (p *Pokedexcli) Lint(ctx context.Context) (string, error) {
	return dag.Golang().
		WithSource(p.Source).
		GolangciLint(ctx)
}

// Build the app
func (p *Pokedexcli) Build(
	// +optional
	arch string,
) *dagger.Directory {
	if arch == "" {
		arch = runtime.GOARCH
	}
	return dag.
		Golang().
		WithSource(p.Source).
		Build([]string{}, dagger.GolangBuildOpts{Arch: arch})
}

// Stateless formatter
func (p *Pokedexcli) FormatFile(
	// Directory with go module
	source *dagger.Directory,
	// File path to format
	path string,
) *dagger.Directory {
	return dag.
		Container().
		From("golang:1.24").
		WithExec([]string{"go", "install", "golang.org/x/tools/gopls@latest"}).
		WithWorkdir("/app").
		WithDirectory("/app", source).
		WithExec([]string{"gopls", "format", "-w", path}).
		WithExec([]string{"gopls", "imports", "-w", path}).
		Directory("/app")
}

// Formatter
func (p *Pokedexcli) Format() *dagger.Directory {
	return dag.
		Golang().
		WithSource(p.Source).
		Fmt().
		GolangciLintFix()
}

// Stateless checker
func (p *Pokedexcli) CheckDirectory(
	ctx context.Context,
	// Directory to run checks on
	source *dagger.Directory,
) (string, error) {
	p.Source = source
	return p.Check(ctx)
}

// Checker
func (p *Pokedexcli) Check(ctx context.Context) (string, error) {
	lint, err := p.Lint(ctx)
	if err != nil {
		return "", err
	}
	test, err := p.Test(ctx)
	if err != nil {
		return "", err
	}
	return lint + "\n" + test, nil
}
