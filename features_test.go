package main_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

// binPath is the location of the compiled go-alphasort executable used by the end-to-end tests.
var binPath string

func TestFeatures(t *testing.T) {
	binPath = filepath.Join(t.TempDir(), "go-alphasort")
	build := exec.Command("go", "build", "-o", binPath, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("cannot compile go-alphasort: %v\n%s", err, output)
	}
	suite := godog.TestSuite{
		Name:                "go-alphasort",
		ScenarioInitializer: initializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("end-to-end tests failed")
	}
}

// compare provides an error describing the difference if the given texts differ.
func compare(what, have, want string) error {
	if have == want {
		return nil
	}
	return fmt.Errorf("unexpected %s\n--- HAVE ---\n%s\n--- WANT ---\n%s", what, have, want)
}

func initializeScenario(sc *godog.ScenarioContext) {
	state := &scenarioState{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		dir, err := os.MkdirTemp("", "go-alphasort-e2e-")
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

	sc.Step(`^I run "([^"]+)"$`, func(command string) error {
		return state.run(command, "")
	})

	sc.Step(`^I run "([^"]+)" with input:$`, func(command string, input *godog.DocString) error {
		return state.run(command, input.Content+"\n")
	})

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

// scenarioState contains the state of a single Cucumber scenario.
type scenarioState struct {
	dir       string            // the temporary directory in which the scenario runs
	originals map[string]string // the content of the files created by the scenario
	stdout    string
	stderr    string
	exitCode  int
}

// run executes the given go-alphasort command in the scenario directory.
func (s *scenarioState) run(command, input string) error {
	args := strings.Fields(command)
	if len(args) == 0 || args[0] != "go-alphasort" {
		return fmt.Errorf("can only run go-alphasort commands, got %q", command)
	}
	cmd := exec.Command(binPath, args[1:]...)
	cmd.Dir = s.dir
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	s.stdout = stdout.String()
	s.stderr = stderr.String()
	s.exitCode = 0
	if exitErr, isExitErr := errors.AsType[*exec.ExitError](err); isExitErr {
		s.exitCode = exitErr.ExitCode()
	} else if err != nil {
		return err
	}
	return nil
}
