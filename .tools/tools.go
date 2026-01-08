//go:build tools
// +build tools

package drydock

import (
	_ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
	_ "gotest.tools/gotestsum"
)
