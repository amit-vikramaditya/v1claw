package tools

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/amit-vikramaditya/v1claw/pkg/bus"
	"github.com/amit-vikramaditya/v1claw/pkg/permissions"
)

type ExecTool struct {
	workingDir          string
	timeout             time.Duration
	denyPatterns        []*regexp.Regexp
	allowPatterns       []*regexp.Regexp
	restrictToWorkspace bool
	securityMiddleware  SecurityMiddleware
	bus                 *bus.MessageBus
}

func NewExecTool(workingDir string, restrict bool, msgBus *bus.MessageBus) *ExecTool {
	denyPatterns := []*regexp.Regexp{
		// Destructive filesystem wipes
		regexp.MustCompile(`\brm\s+(-[a-z]*r[a-z]*f|-[a-z]*f[a-z]*r)\s+/`), // rm -rf / or rm -fr /
		regexp.MustCompile(`\brmdir\s+/`),
		// Disk-level wipes / partition tools
		regexp.MustCompile(`\bdd\b.*\bof=/dev/`),   // dd of=/dev/sda
		regexp.MustCompile(`\bmkfs\b`),             // any filesystem formatting
		regexp.MustCompile(`\bfdisk\b`),
		regexp.MustCompile(`\bparted\b`),
		regexp.MustCompile(`\bwipefs\b`),
		// Privilege escalation
		regexp.MustCompile(`\bsudo\b`),
		regexp.MustCompile(`\bsu\s`),               // su <user> (not "sudo")
		regexp.MustCompile(`\bdoas\b`),
		// Pipe-to-shell download execution (curl|bash etc.)
		regexp.MustCompile(`\|\s*(bash|sh|zsh|ksh|csh|fish)\b`),
		// Kernel module manipulation
		regexp.MustCompile(`\bmodprobe\b`),
		regexp.MustCompile(`\binsmod\b`),
		regexp.MustCompile(`\brmmod\b`),
		// Fork-bomb pattern:  :() { :|:& };:
		regexp.MustCompile(`:\s*\(\)\s*\{`),
		// Overwrite critical system files
		regexp.MustCompile(`>\s*/etc/(passwd|shadow|sudoers|crontab)\b`),
		// chmod setuid/setgid bits
		regexp.MustCompile(`\bchmod\s+[ugoa]*[+]s\b`),
		regexp.MustCompile(`\bchmod\s+[0-7]*[4-7][0-7]{3}\b`), // e.g. chmod 4755
	}

	return &ExecTool{
		workingDir:          workingDir,
		timeout:             60 * time.Second,
		denyPatterns:        denyPatterns,
		allowPatterns:       nil,
		restrictToWorkspace: restrict,
		securityMiddleware:  &AllowlistMiddleware{Allowed: DefaultAllowlist},
		bus:                 msgBus,
	}
}

func (t *ExecTool) SetSecurityMiddleware(sm SecurityMiddleware) {
	t.securityMiddleware = sm
}

func (t *ExecTool) Name() string {
	return "exec"
}

func (t *ExecTool) Description() string {
	return "Execute a shell command and return its output. Use with caution."
}

func (t *ExecTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to execute",
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Optional working directory for the command",
			},
		},
		"required": []string{"command"},
	}
}

func (t *ExecTool) Execute(ctx context.Context, tc ToolContext, args map[string]interface{}) *ToolResult {
	command, ok := args["command"].(string)
	if !ok {
		return ErrorResult("command is required")
	}

	cwd := t.workingDir
	if wd, ok := args["working_dir"].(string); ok && wd != "" {
		cwd = wd
	}

	// Validate working_dir stays within workspace when restricted
	if t.restrictToWorkspace && cwd != t.workingDir {
		absWD, err1 := filepath.Abs(t.workingDir)
		absCWD, err2 := filepath.Abs(cwd)
		if err1 == nil && err2 == nil {
			// Validate symlinks in cwd
			cwdReal := absCWD
			if resolved, err := filepath.EvalSymlinks(absCWD); err == nil {
				cwdReal = resolved
			} else if !os.IsNotExist(err) && !os.IsPermission(err) {
				return ErrorResult(fmt.Sprintf("failed to resolve working_dir symlink: %v", err))
			}

			if !strings.HasPrefix(cwdReal, absWD+string(filepath.Separator)) && cwdReal != absWD {
				if t.bus != nil && tc.Channel != "" {
					t.bus.PublishOutbound(bus.OutboundMessage{
						Channel: tc.Channel,
						ChatID:  tc.ChatID,
						Content: fmt.Sprintf("⚠️ **Security Alert**: I attempted to execute a command in `%s` but was blocked by your Strict Sandbox configuration. I am restricted to `%s`.", cwdReal, t.workingDir),
					})
				}
				return ErrorResult("working_dir must be within the workspace directory")
			}
		}
	}

	if cwd == "" {
		wd, err := os.Getwd()
		if err == nil {
			cwd = wd
		}
	}

	// 1. Security Middleware Vetting
	vettedCommand := command
	if t.securityMiddleware != nil {
		var err error
		vettedCommand, err = t.securityMiddleware.VerifyCommand(command)
		if err != nil {
			return ErrorResult(fmt.Sprintf("Security violation: %v", err))
		}
	}

	// 2. Legacy Pattern Guarding (still here for a second layer of defense/specific hardware checks)
	if guardError := t.guardCommand(vettedCommand, cwd); guardError != "" {
		return ErrorResult(guardError)
	}

	// ... continue with execution using vettedCommand
	var cmdCtx context.Context
	var cancel context.CancelFunc
	if t.timeout > 0 {
		cmdCtx, cancel = context.WithTimeout(ctx, t.timeout)
	} else {
		cmdCtx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	// Determine command and arguments for direct execution.
	// This entirely bypasses 'sh -c' to prevent shell injection.
	program, argsSlice := parseCommandForDirectExecution(vettedCommand)
	if program == "" {
		return ErrorResult("Failed to parse command for direct execution or empty command")
	}

	// 3. Program Path Validation (Security against arbitrary binary execution)
	if t.restrictToWorkspace {
		// Attempt to resolve the program in the PATH
		resolvedProgram, err := exec.LookPath(program)
		if err != nil {
			return ErrorResult(fmt.Sprintf("Program '%s' not found in PATH or not executable", program))
		}
		program = resolvedProgram
	}

	var cmd *exec.Cmd
	cmd = exec.CommandContext(cmdCtx, program, argsSlice...)
	if cwd != "" {
		cmd.Dir = cwd
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\nSTDERR:\n" + stderr.String()
	}

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			msg := fmt.Sprintf("Command timed out after %v", t.timeout)
			return &ToolResult{
				ForLLM:  msg,
				ForUser: msg,
				IsError: true,
			}
		}
		output += fmt.Sprintf("\nExit code: %v", err)
	}

	if output == "" {
		output = "(no output)"
	}

	maxLen := 10000
	if len(output) > maxLen {
		output = output[:maxLen] + fmt.Sprintf("\n... (truncated, %d more chars)", len(output)-maxLen)
	}

	if err != nil {
		return &ToolResult{
			ForLLM:  output,
			ForUser: output,
			IsError: true,
		}
	}

	return &ToolResult{
		ForLLM:  output,
		ForUser: output,
		IsError: false,
	}
}

func (t *ExecTool) guardCommand(command, cwd string) string {
	cmd := strings.TrimSpace(command)
	lower := strings.ToLower(cmd)

	// Block interpreter bypass patterns for dynamic languages that could escape the sandbox.
	// Use filepath.Base so that full-path invocations (/usr/bin/python3, /bin/sh …) are caught
	// in addition to bare names.
	interpreters := []string{"python", "python3", "sh", "bash", "zsh", "ruby", "perl", "node"}
	fields := strings.Fields(lower)
	if len(fields) > 0 {
		baseCmd := filepath.Base(fields[0])
		for _, interp := range interpreters {
			if baseCmd == interp {
				// Catch -c flag, eval, OS module imports, /etc/ access — common sandbox escapes.
				if strings.Contains(lower, "import os") ||
					strings.Contains(lower, "/etc/") ||
					strings.Contains(lower, "system(") ||
					strings.Contains(lower, " -c ") ||
					strings.Contains(lower, "\t-c ") ||
					strings.Contains(lower, "eval") {
					return "Command blocked by safety guard (interpreter bypass patterns detected)"
				}
				break
			}
		}
	}

	for _, pattern := range t.denyPatterns {
		if pattern.MatchString(lower) {
			return "Command blocked by safety guard (dangerous pattern detected)"
		}
	}

	// Block hardware commands unless the corresponding permission is enabled.
	if msg := guardHardwareCommands(lower); msg != "" {
		return msg
	}

	if len(t.allowPatterns) > 0 {
		allowed := false
		for _, pattern := range t.allowPatterns {
			if pattern.MatchString(lower) {
				allowed = true
				break
			}
		}
		if !allowed {
			return "Command blocked by safety guard (not in allowlist)"
		}
	}

	if t.restrictToWorkspace {
		if strings.Contains(cmd, "..\\") || strings.Contains(cmd, "../") {
			return "Command blocked by safety guard (path traversal detected)"
		}

		cwdPath, err := filepath.Abs(cwd)
		if err != nil {
			return ""
		}

		pathPattern := regexp.MustCompile(`[A-Za-z]:\\[^\\\"'\s]+|/[^\s\"']+`)
		matches := pathPattern.FindAllString(cmd, -1)

		for _, raw := range matches {
			p, err := filepath.Abs(raw)
			if err != nil {
				continue
			}

			pReal := p
			if resolved, err := filepath.EvalSymlinks(p); err == nil {
				pReal = resolved
			}

			cwdReal := cwdPath
			if resolved, err := filepath.EvalSymlinks(cwdPath); err == nil {
				cwdReal = resolved
			}

			rel, err := filepath.Rel(cwdReal, pReal)
			if err != nil {
				continue
			}

			if strings.HasPrefix(rel, "..") {
				return "Command blocked by safety guard (path outside working dir)"
			}
		}
	}

	return ""
}

func (t *ExecTool) SetTimeout(timeout time.Duration) {
	t.timeout = timeout
}

func (t *ExecTool) SetRestrictToWorkspace(restrict bool) {
	t.restrictToWorkspace = restrict
}

func (t *ExecTool) SetAllowPatterns(patterns []string) error {
	t.allowPatterns = make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("invalid allow pattern %q: %w", p, err)
		}
		t.allowPatterns = append(t.allowPatterns, re)
	}
	return nil
}

// hardwareGuard maps command prefixes to the permission feature they require.
var hardwareGuards = []struct {
	pattern *regexp.Regexp
	feature permissions.Feature
	desc    string
}{
	{regexp.MustCompile(`\btermux-camera-photo\b`), permissions.Camera, "camera capture"},
	{regexp.MustCompile(`\btermux-microphone-record\b`), permissions.Microphone, "microphone recording"},
	{regexp.MustCompile(`\btermux-sms-(send|list)\b`), permissions.SMS, "SMS access"},
	{regexp.MustCompile(`\btermux-telephony-call\b`), permissions.PhoneCalls, "phone call"},
	{regexp.MustCompile(`\btermux-location\b`), permissions.Location, "location access"},
	{regexp.MustCompile(`\btermux-clipboard-(get|set)\b`), permissions.Clipboard, "clipboard access"},
	{regexp.MustCompile(`\btermux-sensor\b`), permissions.Sensors, "sensor access"},
	// Catch-all for any termux-* hardware command not explicitly listed.
	{regexp.MustCompile(`\btermux-(vibrate|torch|volume|notification|toast|wifi|battery|usb|fingerprint|wallpaper|dialog|download|media-player|media-scan|storage-get|tts-speak)\b`), permissions.ShellHardware, "hardware command"},
}

// guardHardwareCommands checks if a shell command invokes hardware APIs
// and verifies the corresponding permission is enabled.
func guardHardwareCommands(lower string) string {
	reg := permissions.Global()
	for _, g := range hardwareGuards {
		if g.pattern.MatchString(lower) {
			if !reg.IsAllowed(g.feature) {
				return fmt.Sprintf("Command blocked: %s requires permissions.%s=true in config", g.desc, g.feature)
			}
		}
	}
	// Catch-all: block ANY termux-* command not already allowed above
	if strings.Contains(lower, "termux-") {
		if !reg.IsAllowed(permissions.ShellHardware) {
			return "Command blocked: termux hardware commands require permissions.shell_hardware=true in config"
		}
	}
	return ""
}

// parseCommandForDirectExecution parses a command string into a program and arguments.
// This is used for direct execution (bypassing 'sh -c') to prevent shell injection.
func parseCommandForDirectExecution(command string) (string, []string) {
	if command == "" {
		return "", nil
	}

	var parts []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escapeNext := false

	for _, r := range command {
		if escapeNext {
			current.WriteRune(r)
			escapeNext = false
			continue
		}

		if r == '\\' && !inSingleQuote {
			escapeNext = true
			continue
		}

		if r == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			continue
		}

		if r == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			continue
		}

		if (r == ' ' || r == '\t') && !inSingleQuote && !inDoubleQuote {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(r)
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	if len(parts) == 0 {
		return "", nil
	}
	program := parts[0]
	args := parts[1:]
	return program, args
}
