package main

import (
	"context"
	"fmt"
	"strings"
	"dagger/pokedexcli/internal/dagger"
)

// Debug broken tests. Returns a unified diff of the test fixes
func (g *Pokedexcli) DebugTests(
	ctx context.Context,
	// The model to use to debug debug tests
	// +optional
	// +default = "gemini-2.0-flash"
	model string,
) (string, error) {
	prompt := dag.CurrentModule().Source().File("prompts/fix_tests.md")

	// Check if backend is broken
	if _, err := g.CheckDirectory(ctx, g.Source); err != nil {
		ws := dag.Workspace(
			g.Source,
		)
		env := dag.Env().
			WithWorkspaceInput("workspace", ws, "workspace to read, write, and test code").
			WithWorkspaceOutput("fixed", "workspace with fixed tests")
		return dag.LLM(dagger.LLMOpts{Model: model}).
			WithEnv(env).
			WithPromptFile(prompt).
			Env().
			Output("fixed").
			AsWorkspace().
			Diff(ctx)
	}

	return "", fmt.Errorf("no broken tests found")
}

// Debug broken tests on a pull request and comment fix suggestions
func (g *Pokedexcli) DebugBrokenTestsPr(
	ctx context.Context,
	// Github token with permissions to comment on the pull request
	githubToken *dagger.Secret,
	// Git commit in Github
	commit string,
	// The model to use to debug debug tests
	// +optional
	// +default = "gemini-2.0-flash"
	model string,
) error {
	gh := dag.GithubIssue(dagger.GithubIssueOpts{Token: githubToken})
	// Determine PR head
	gitRef := dag.Git(g.Repo).Commit(commit)
	gitSource := gitRef.Tree()
	pr, err := gh.GetPrForCommit(ctx, g.Repo, commit)
	if err != nil {
		return err
	}

	// Set source to PR head
	g = New(gitSource, g.Repo)

	// Suggest fix
	suggestionDiff, err := g.DebugTests(ctx, model)
	if err != nil {
		return err
	}
	if suggestionDiff == "" {
		return fmt.Errorf("no suggestions found")
	}

	// Convert the diff to CodeSuggestions
	codeSuggestions := parseDiff(suggestionDiff)

	// For each suggestion, comment on PR
	for _, suggestion := range codeSuggestions {
		markupSuggestion := "```suggestion\n" + strings.Join(suggestion.Suggestion, "\n") + "\n```"
		err := gh.WritePullRequestCodeComment(
			ctx,
			g.Repo,
			pr,
			commit,
			markupSuggestion,
			suggestion.File,
			"RIGHT",
			suggestion.Line)
		if err != nil {
			return err
		}
	}
	return nil
}
