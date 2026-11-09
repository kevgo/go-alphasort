// Package test provides the Cucumber step definitions for the end-to-end tests of alphasort.
package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

// InitializeScenario provides a godog scenario initializer
// that runs the alphasort executable at the given path.
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

		sc.Step(`^the exit code is (\d+)$`, func(want int) error {
			if state.exitCode != want {
				return fmt.Errorf("expected exit code %d but got %d\nSTDOUT:\n%s\nSTDERR:\n%s", want, state.exitCode, state.stdout, state.stderr)
			}
			return nil
		})

		sc.Step(`^the file "([^"]+)" now has content:$`, func(name string, want *godog.DocString) error {
			have, err := os.ReadFile(filepath.Join(state.dir, name))
			if err != nil {
				return err
			}
			return compare("file "+name, string(have), want.Content+"\n")
		})

		sc.Step(`^the file "([^"]+)" is unchanged$`, func(name string) error {
			have, err := os.ReadFile(filepath.Join(state.dir, name))
			if err != nil {
				return err
			}
			return compare("file "+name, string(have), state.originals[name])
		})

		sc.Step(`^the output contains "([^"]+)"$`, func(want string) error {
			if !strings.Contains(state.stdout, want) {
				return fmt.Errorf("expected STDOUT to contain %q but it was:\n%s", want, state.stdout)
			}
			return nil
		})

		sc.Step(`^the output is:$`, func(want *godog.DocString) error {
			return compare("STDOUT", state.stdout, want.Content+"\n")
		})

		sc.Step(`^the output is empty$`, func() error {
			return compare("STDOUT", state.stdout, "")
		})

		sc.Step(`^the error output contains "([^"]+)"$`, func(want string) error {
			if !strings.Contains(state.stderr, want) {
				return fmt.Errorf("expected STDERR to contain %q but it was:\n%s", want, state.stderr)
			}
			return nil
		})
	}
}

// compare provides an error describing the difference if the given texts differ.
func compare(what, have, want string) error {
	if have == want {
		return nil
	}
	return fmt.Errorf("unexpected %s\n--- HAVE ---\n%s\n--- WANT ---\n%s", what, have, want)
}
