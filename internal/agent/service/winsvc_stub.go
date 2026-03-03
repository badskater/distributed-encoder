//go:build !windows

package service

import (
	"context"
	"fmt"
)

// runAsWindowsService is a stub for non-Windows platforms.
func runAsWindowsService(_ string, _ func(ctx context.Context) error) error {
	return fmt.Errorf("windows service not supported on this platform")
}

// isWindowsService always returns false on non-Windows platforms.
func isWindowsService() bool {
	return false
}
