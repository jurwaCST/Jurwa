//go:build tools

// Package tools keeps libraries used by services/_template in go.mod.
// Go skips folders starting with "_", so without this file "go mod tidy"
// would remove them and every new service's tests would fail to build.
package tools

import (
	_ "github.com/stretchr/testify/assert"
)
