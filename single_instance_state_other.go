//go:build !windows

package main

import (
	"lianxiang-browser/backend"
	"path/filepath"
	"strings"
)

func singleInstanceStateRoot(appRoot string) string {
	if root := strings.TrimSpace(backend.RuntimeStateRoot(appRoot)); root != "" {
		return root
	}
	return filepath.Join(appRoot, "data")
}
