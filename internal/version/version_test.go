package version

import (
	"regexp"
	"strings"
	"testing"
)

func TestVersionFormat(t *testing.T) {
	semver := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	v := strings.TrimSpace(raw)
	if !semver.MatchString(v) {
		t.Fatalf("VERSION must be in x.x.x format, got %q", v)
	}
}
