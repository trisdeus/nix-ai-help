package execution

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// CommandValidatorImpl provides comprehensive command validation
type CommandValidatorImpl struct {
	allowedCommands    map[string]bool
	forbiddenCommands  map[string]bool
	allowedDirectories []string
	forbiddenPaths     []string
	dangerousPatterns  []*regexp.Regexp
}

// NewCommandValidator creates a new command validator
func NewCommandValidator(
	allowedCommands []string,
	forbiddenCommands []string,
	allowedDirectories []string,
	forbiddenPaths []string,
) *CommandValidatorImpl {
	validator := &CommandValidatorImpl{
		allowedCommands:    make(map[string]bool),
		forbiddenCommands:  make(map[string]bool),
		allowedDirectories: allowedDirectories,
		forbiddenPaths:     forbiddenPaths,
		dangerousPatterns:  compileDangerousPatterns(),
	}
	
	// Build allowed commands map
	for _, cmd := range allowedCommands {
		validator.allowedCommands[cmd] = true
		// Also allow commands with wildcards (e.g., "nix*" allows "nix-env", "nixos-rebuild")
		if strings.HasSuffix(cmd, "*") {
			prefix := strings.TrimSuffix(cmd, "*")
			validator.allowedCommands[prefix] = true
		}
	}
	
	// Build forbidden commands map
	for _, cmd := range forbiddenCommands {
		validator.forbiddenCommands[cmd] = true
	}
	
	return validator
}

// ValidateCommand performs comprehensive command validation
func (cv *CommandValidatorImpl) ValidateCommand(req CommandRequest) error {
	// Check if command is explicitly forbidden
	if cv.forbiddenCommands[req.Command] {
		return &SecurityError{
			Command: req.Command,
			Reason:  "command is explicitly forbidden",
			Code:    "FORBIDDEN_COMMAND",
		}
	}
	
	// Check if command is allowed
	if !cv.isCommandAllowed(req.Command) {
		return &SecurityError{
			Command: req.Command,
			Reason:  "command not in allowed list",
			Code:    "COMMAND_NOT_ALLOWED",
		}
	}
	
	// Validate arguments for dangerous patterns
	if err := cv.validateArguments(req.Args); err != nil {
		return err
	}
	
	// Validate working directory
	if req.WorkingDir != "" {
		if err := cv.validateWorkingDirectory(req.WorkingDir); err != nil {
			return err
		}
	}
	
	// Check for command-specific restrictions
	if err := cv.validateCommandSpecific(req); err != nil {
		return err
	}
	
	return nil
}

// IsCommandAllowed checks if a command is in the allowed list
func (cv *CommandValidatorImpl) IsCommandAllowed(command string, args []string) bool {
	// Check exact match
	if cv.allowedCommands[command] {
		return true
	}
	
	// Check wildcard matches
	for allowedCmd := range cv.allowedCommands {
		if strings.HasSuffix(allowedCmd, "*") {
			prefix := strings.TrimSuffix(allowedCmd, "*")
			if strings.HasPrefix(command, prefix) {
				return true
			}
		}
	}
	
	return false
}

// RequiresConfirmation determines if a command requires user confirmation
func (cv *CommandValidatorImpl) RequiresConfirmation(req CommandRequest) bool {
	// Always require confirmation for sudo commands
	if req.RequiresSudo {
		return true
	}
	
	// Require confirmation for system-level commands
	if req.Category == string(CategorySystem) {
		return true
	}
	
	// Require confirmation for potentially destructive operations
	destructiveCommands := []string{"rm", "mv", "cp", "nixos-rebuild"}
	for _, cmd := range destructiveCommands {
		if req.Command == cmd || strings.HasPrefix(req.Command, cmd) {
			return true
		}
	}
	
	// Check for destructive arguments
	for _, arg := range req.Args {
		if strings.Contains(arg, "--force") || 
		   strings.Contains(arg, "--delete") ||
		   strings.Contains(arg, "-rf") ||
		   strings.Contains(arg, "--remove") {
			return true
		}
	}
	
	return false
}

// isCommandAllowed checks if a command is allowed (internal implementation)
func (cv *CommandValidatorImpl) isCommandAllowed(command string) bool {
	return cv.IsCommandAllowed(command, nil)
}

// validateArguments checks command arguments for dangerous patterns
func (cv *CommandValidatorImpl) validateArguments(args []string) error {
	for i, arg := range args {
		// Check against dangerous patterns
		for _, pattern := range cv.dangerousPatterns {
			if pattern.MatchString(arg) {
				return &SecurityError{
					Reason: fmt.Sprintf("argument %d matches dangerous pattern: %s", i, arg),
					Code:   "DANGEROUS_ARGUMENT",
				}
			}
		}
		
		// Check for shell injection attempts
		if err := cv.checkShellInjection(arg); err != nil {
			return err
		}
		
		// Check for path traversal attempts
		if err := cv.checkPathTraversal(arg); err != nil {
			return err
		}
		
		// Check against forbidden paths
		if err := cv.checkForbiddenPaths(arg); err != nil {
			return err
		}
	}
	
	return nil
}

// validateWorkingDirectory validates the working directory
func (cv *CommandValidatorImpl) validateWorkingDirectory(workingDir string) error {
	// Convert to absolute path
	absPath, err := filepath.Abs(workingDir)
	if err != nil {
		return &ValidationError{
			Reason: fmt.Sprintf("invalid working directory path: %s", workingDir),
			Code:   "INVALID_PATH",
		}
	}
	
	// Check if directory is in allowed list
	allowed := false
	for _, allowedDir := range cv.allowedDirectories {
		allowedAbs, err := filepath.Abs(allowedDir)
		if err != nil {
			continue
		}
		
		// Check if working directory is within allowed directory
		if strings.HasPrefix(absPath, allowedAbs) {
			allowed = true
			break
		}
	}
	
	if !allowed {
		return &SecurityError{
			Reason: fmt.Sprintf("working directory not allowed: %s", workingDir),
			Code:   "DIRECTORY_NOT_ALLOWED",
		}
	}
	
	// Check against forbidden paths
	for _, forbiddenPath := range cv.forbiddenPaths {
		forbiddenAbs, err := filepath.Abs(forbiddenPath)
		if err != nil {
			continue
		}
		
		if strings.HasPrefix(absPath, forbiddenAbs) {
			return &SecurityError{
				Reason: fmt.Sprintf("working directory in forbidden path: %s", workingDir),
				Code:   "FORBIDDEN_PATH",
			}
		}
	}
	
	return nil
}

// validateCommandSpecific performs command-specific validation
func (cv *CommandValidatorImpl) validateCommandSpecific(req CommandRequest) error {
	switch req.Command {
	case "rm":
		return cv.validateRmCommand(req.Args)
	case "mv", "cp":
		return cv.validateFileOperations(req.Args)
	case "nixos-rebuild":
		return cv.validateNixOSRebuild(req.Args)
	case "systemctl":
		return cv.validateSystemctl(req.Args)
	case "nix-env":
		return cv.validateNixEnv(req.Args)
	default:
		return nil
	}
}

// validateRmCommand validates rm command arguments
func (cv *CommandValidatorImpl) validateRmCommand(args []string) error {
	for _, arg := range args {
		// Prevent dangerous rm operations
		if arg == "-rf" || arg == "-r" || arg == "--recursive" {
			return &SecurityError{
				Reason: "recursive rm operations are not allowed",
				Code:   "DANGEROUS_RM",
			}
		}
		
		// Check for system directories
		systemDirs := []string{"/", "/boot", "/etc", "/usr", "/var", "/sys", "/proc"}
		for _, sysDir := range systemDirs {
			if strings.HasPrefix(arg, sysDir) {
				return &SecurityError{
					Reason: fmt.Sprintf("cannot rm system directory: %s", arg),
					Code:   "SYSTEM_DIR_RM",
				}
			}
		}
	}
	
	return nil
}

// validateFileOperations validates mv/cp command arguments
func (cv *CommandValidatorImpl) validateFileOperations(args []string) error {
	if len(args) < 2 {
		return &ValidationError{
			Reason: "insufficient arguments for file operation",
			Code:   "INSUFFICIENT_ARGS",
		}
	}
	
	// Check source and destination paths
	for _, path := range args {
		if strings.HasPrefix(path, "/etc/") && !strings.HasPrefix(path, "/etc/nixos/") {
			return &SecurityError{
				Reason: fmt.Sprintf("file operation on restricted path: %s", path),
				Code:   "RESTRICTED_PATH",
			}
		}
	}
	
	return nil
}

// validateNixOSRebuild validates nixos-rebuild command arguments
func (cv *CommandValidatorImpl) validateNixOSRebuild(args []string) error {
	validActions := []string{"switch", "test", "build", "dry-run", "boot"}
	
	if len(args) == 0 {
		return &ValidationError{
			Reason: "nixos-rebuild requires an action argument",
			Code:   "MISSING_ACTION",
		}
	}
	
	action := args[0]
	for _, validAction := range validActions {
		if action == validAction {
			return nil
		}
	}
	
	return &ValidationError{
		Reason: fmt.Sprintf("invalid nixos-rebuild action: %s", action),
		Code:   "INVALID_ACTION",
	}
}

// validateSystemctl validates systemctl command arguments
func (cv *CommandValidatorImpl) validateSystemctl(args []string) error {
	if len(args) == 0 {
		return &ValidationError{
			Reason: "systemctl requires action and service arguments",
			Code:   "MISSING_ARGS",
		}
	}
	
	action := args[0]
	allowedActions := []string{"start", "stop", "restart", "status", "enable", "disable", "reload"}
	
	for _, allowed := range allowedActions {
		if action == allowed {
			return nil
		}
	}
	
	return &ValidationError{
		Reason: fmt.Sprintf("systemctl action not allowed: %s", action),
		Code:   "SYSTEMCTL_ACTION_NOT_ALLOWED",
	}
}

// validateNixEnv validates nix-env command arguments
func (cv *CommandValidatorImpl) validateNixEnv(args []string) error {
	// Allow common nix-env operations
	allowedFlags := []string{"-i", "-iA", "-e", "-q", "-u", "--install", "--uninstall", "--query", "--upgrade"}
	
	if len(args) == 0 {
		return nil // nix-env without args is safe (shows help)
	}
	
	firstArg := args[0]
	for _, allowed := range allowedFlags {
		if firstArg == allowed {
			return nil
		}
	}
	
	return &ValidationError{
		Reason: fmt.Sprintf("nix-env flag not allowed: %s", firstArg),
		Code:   "NIX_ENV_FLAG_NOT_ALLOWED",
	}
}

// checkShellInjection checks for shell injection attempts
func (cv *CommandValidatorImpl) checkShellInjection(arg string) error {
	// Check for command separators
	injectionPatterns := []string{";", "&&", "||", "|", "`", "$(", "${"}
	
	for _, pattern := range injectionPatterns {
		if strings.Contains(arg, pattern) {
			return &SecurityError{
				Reason: fmt.Sprintf("potential shell injection detected: %s", pattern),
				Code:   "SHELL_INJECTION",
			}
		}
	}
	
	return nil
}

// checkPathTraversal checks for path traversal attempts
func (cv *CommandValidatorImpl) checkPathTraversal(arg string) error {
	if strings.Contains(arg, "../") || strings.Contains(arg, "..\\") {
		return &SecurityError{
			Reason: "path traversal attempt detected",
			Code:   "PATH_TRAVERSAL",
		}
	}
	
	return nil
}

// checkForbiddenPaths checks if argument contains forbidden paths
func (cv *CommandValidatorImpl) checkForbiddenPaths(arg string) error {
	for _, forbiddenPath := range cv.forbiddenPaths {
		if strings.HasPrefix(arg, forbiddenPath) {
			return &SecurityError{
				Reason: fmt.Sprintf("argument references forbidden path: %s", forbiddenPath),
				Code:   "FORBIDDEN_PATH",
			}
		}
	}
	
	return nil
}

// compileDangerousPatterns compiles regex patterns for dangerous operations
func compileDangerousPatterns() []*regexp.Regexp {
	patterns := []string{
		`rm\s+-rf\s+/`,           // rm -rf /
		`dd\s+if=.*of=/dev/`,     // dd to block devices
		`mkfs\..*`,               // filesystem creation
		`fdisk\s+/dev/`,          // disk partitioning
		`parted\s+/dev/`,         // disk partitioning
		`wipefs\s+`,              // filesystem signature wiping
		`shred\s+`,               // secure file deletion
		`:(){ :|:& };:`,          // fork bomb
	}
	
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, regex)
		}
	}
	
	return compiled
}