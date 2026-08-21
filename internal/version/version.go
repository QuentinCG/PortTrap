package version

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var raw string

// Version is the release version read from the VERSION file (the single source of truth).
var Version = "v" + strings.TrimSpace(raw)
