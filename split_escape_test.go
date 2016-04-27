package devn

import (
	"bufio"
	"strings"
	"testing"
)

func TestSplitEscape(t *testing.T) {

	lines := `Line 0
Line 1
Line 2 \
should be joined
Line 3
# Line 4 \
Just Join
# Line 5 \
#Drop Hash
Line 6`

	s := bufio.NewScanner(strings.NewReader(lines))
	s.Split(ScanEscapedLines)

	for _, expect := range []string{
		"Line 0",
		"Line 1",
		"Line 2 should be joined",
		"Line 3",
		"# Line 4 Just Join",
		"# Line 5 Drop Hash",
		"Line 6"} {

		if !s.Scan() {
			t.Errorf("Scan ran out of lines")
		}
		got := s.Text()
		t.Logf("Got line: '%s'", got)
		if got != expect {
			t.Errorf("Expected '%s', got '%s'", expect, got)
		}
	}
}
