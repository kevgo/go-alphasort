package main_test

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
	"github.com/kevgo/go-alphasort/test"
)

func TestFeatures(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "go-alphasort")
	build := exec.Command("go", "build", "-o", binPath, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("cannot compile go-alphasort: %v\n%s", err, output)
	}
	suite := godog.TestSuite{
		Name:                "go-alphasort",
		ScenarioInitializer: test.InitializeScenario(binPath),
		Options: &godog.Options{
			Format:   "progress",
			Paths:    []string{"features"},
			Strict:   true,
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("end-to-end tests failed")
	}
}
