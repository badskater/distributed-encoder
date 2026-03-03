// Package webui provides the embedded static web UI assets.
package webui

import "embed"

//go:embed dist
var Files embed.FS
