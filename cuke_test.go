package alphasort_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
	"github.com/kevgo/go-alphasort/test"
)

func TestFeatures(t *testing.T) {
	// "make cuke" runs "make install" first, which installs the binary to test into ~/go/bin.
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot determine the home directory: %v", err)
	}
	binPath := filepath.Join(home, "go", "bin", "alphasort")
	if _, err := os.Stat(binPath); err != nil {
		t.Fatalf("cannot find the alphasort binary at %s, please run \"make install\": %v", binPath, err)
	}
	suite := godog.TestSuite{
		Name:                "alphasort",
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
