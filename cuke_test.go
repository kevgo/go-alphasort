package alphasort_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
	"github.com/kevgo/go-alphasort/test"
	"github.com/shoenig/test/must"
)

func TestFeatures(t *testing.T) {
	// "make cuke" runs "make install" first, which installs the binary to test into ~/go/bin.
	home, err := os.UserHomeDir()
	must.NoError(t, err, must.Sprint("cannot determine the home directory"))
	binPath := filepath.Join(home, "go", "bin", "alphasort")
	_, err = os.Stat(binPath)
	must.NoError(t, err, must.Sprintf("cannot find the alphasort binary at %s, please run \"make install\"", binPath))
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
	must.Zero(t, suite.Run(), must.Sprint("end-to-end tests failed"))
}
