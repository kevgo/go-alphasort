package test

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// scenarioState contains the state of a single Cucumber scenario.
type scenarioState struct {
	binPath   string            // the location of the installed alphasort executable
	dir       string            // the temporary directory in which the scenario runs
	originals map[string]string // the content of the files created by the scenario
	stdout    string
	stderr    string
	exitCode  int
}

// run executes the given alphasort command in the scenario directory.
func (s *scenarioState) run(command string) error {
	args := strings.Fields(command)
	if len(args) == 0 || args[0] != "alphasort" {
		return fmt.Errorf("can only run alphasort commands, got %q", command)
	}
	cmd := exec.Command(s.binPath, args[1:]...)
	cmd.Dir = s.dir
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
