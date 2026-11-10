// Package test provides the Cucumber step definitions for the end-to-end tests of alphasort.
package test

import (
	"context"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cucumber/godog"
	"github.com/shoenig/test/must"
)

func InitializeScenario(binPath string) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		state := &scenarioState{binPath: binPath}

		sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
			dir, err := os.MkdirTemp("", "alphasort-e2e-")
			state.dir = dir
			state.originals = map[string]string{}
			return ctx, err
		})

		sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
			return ctx, os.RemoveAll(state.dir)
		})

		sc.Step(`^a file "([^"]+)" with content:$`, func(name string, content *godog.DocString) error {
			path := filepath.Join(state.dir, name)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			text := content.Content + "\n"
			state.originals[name] = text
			return os.WriteFile(path, []byte(text), 0o644)
		})

		sc.Step(`^I run "([^"]+)"$`, state.run)

		sc.Step(`^the exit code is (\d+)$`, func(ctx context.Context, want int) {
			must.Eq(t(ctx), want, state.exitCode, must.Sprintf("STDOUT:\n%s\nSTDERR:\n%s", state.stdout, state.stderr))
		})

		sc.Step(`^the file "([^"]+)" now has content:$`, func(ctx context.Context, name string, want *godog.DocString) {
			have, err := os.ReadFile(filepath.Join(state.dir, name))
			must.NoError(t(ctx), err)
			must.Eq(t(ctx), want.Content+"\n", string(have), must.Sprint("file "+name))
		})

		sc.Step(`^the file "([^"]+)" is unchanged$`, func(ctx context.Context, name string) {
			have, err := os.ReadFile(filepath.Join(state.dir, name))
			must.NoError(t(ctx), err)
			must.Eq(t(ctx), state.originals[name], string(have), must.Sprint("file "+name))
		})

		sc.Step(`^the output contains "([^"]+)"$`, func(ctx context.Context, want string) {
			must.StrContains(t(ctx), state.stdout, want, must.Sprint("STDOUT"))
		})

		sc.Step(`^the output matches "([^"]+)"$`, func(ctx context.Context, pattern string) {
			must.RegexMatch(t(ctx), regexp.MustCompile(pattern), state.stdout, must.Sprint("STDOUT"))
		})

		sc.Step(`^the output is:$`, func(ctx context.Context, want *godog.DocString) {
			must.Eq(t(ctx), want.Content+"\n", state.stdout, must.Sprint("STDOUT"))
		})

		sc.Step(`^the output is empty$`, func(ctx context.Context) {
			must.Eq(t(ctx), "", state.stdout, must.Sprint("STDOUT"))
		})

		sc.Step(`^the error output contains "([^"]+)"$`, func(ctx context.Context, want string) {
			must.StrContains(t(ctx), state.stderr, want, must.Sprint("STDERR"))
		})

		sc.Step(`^the error output is empty$`, func(ctx context.Context) {
			must.Eq(t(ctx), "", state.stderr, must.Sprint("STDERR"))
		})
	}
}

// t provides the testing.T of the current scenario in a form usable by the must library.
func t(ctx context.Context) must.T {
	return mustT{godog.T(ctx)}
}

// mustT adapts godog.TestingT to the must.T interface.
type mustT struct {
	godog.TestingT
}

// Fatalf records the failure message so that godog displays it as the reason why the step failed.
// godog's own Fatalf only logs the message and reports a generic "fail called on TestingT" error.
func (t mustT) Fatalf(format string, args ...any) {
	t.Errorf(format, args...)
	t.FailNow()
}

func (mustT) Helper() {}
