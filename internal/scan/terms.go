package scan

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// TermSet holds compiled regex for matching references.
type TermSet struct {
	Patterns []*regexp.Regexp
}

// LoadTermsFromFile reads a newline-separated list of regex patterns.
// Lines starting with '#' and blank lines are ignored.
func LoadTermsFromFile(path string) (TermSet, error) {
	f, err := os.Open(path)
	if err != nil {
		return TermSet{}, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	var patterns []*regexp.Regexp
	ln := 0
	for s.Scan() {
		ln++
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rx, err := regexp.Compile(line)
		if err != nil {
			return TermSet{}, fmt.Errorf("invalid regex at %s:%d: %w", path, ln, err)
		}
		patterns = append(patterns, rx)
	}
	if err := s.Err(); err != nil {
		return TermSet{}, err
	}
	return TermSet{Patterns: patterns}, nil
}

// WithAddedRegex returns a copy with additional compiled regex patterns.
func (t TermSet) WithAddedRegex(extra ...string) TermSet {
	out := make([]*regexp.Regexp, 0, len(t.Patterns)+len(extra))
	out = append(out, t.Patterns...)
	for _, e := range extra {
		out = append(out, regexp.MustCompile(e))
	}
	return TermSet{Patterns: out}
}

