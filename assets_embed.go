package v1clawassets

import "embed"

// Workspace contains the built-in workspace templates shipped with v1claw.
//
//go:embed workspace
var Workspace embed.FS
