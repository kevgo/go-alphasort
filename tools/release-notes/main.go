// Prints the section matching the given tage from CHANGELOG.md.
//
// Usage: go run ./tools/release-notes <tag>
//
// Tag "v1.2.3" selects section "## 1.2.3".
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const changelogPath = "CHANGELOG.md"

// extract returns the body of the changelog section for the given tag.
func extract(changelog, tag string) (string, error) {
	heading := "## " + strings.TrimPrefix(tag, "v")
	var section []string
	found := false
	for line := range strings.Lines(changelog) {
		line = strings.TrimRight(line, " \t\r\n")
		if !found {
			found = line == heading
			continue
		}
		if strings.HasPrefix(line, "## ") {
			break
		}
		section = append(section, line)
	}
	if !found {
		return "", fmt.Errorf("%s has no section %q", changelogPath, heading)
	}
	body := strings.TrimSpace(strings.Join(section, "\n"))
	if body == "" {
		return "", fmt.Errorf("section %q in %s is empty", heading, changelogPath)
	}
	return body + "\n", nil
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes this tool with the given CLI arguments and returns the exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: release-notes <tag>")
		return 2
	}
	content, err := os.ReadFile(changelogPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	notes, err := extract(string(content), args[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if _, err := io.WriteString(stdout, notes); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
