//go:build tools
// +build tools

package drydock

import (
	_ "gotest.tools/gotestsum"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
