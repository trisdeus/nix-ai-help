package types

// ExecutionRequest represents a request to execute a command
type ExecutionRequest struct {
	UserQuery     string            `json:"user_query"`
	Command       string            `json:"command,omitempty"`
	Args          []string          `json:"args,omitempty"`
	Description   string            `json:"description,omitempty"`
	Category      string            `json:"category,omitempty"`
	RequiresSudo  bool              `json:"requires_sudo,omitempty"`
	WorkingDir    string            `json:"working_dir,omitempty"`
	Environment   map[string]string `json:"environment,omitempty"`
	DryRun        bool              `json:"dry_run,omitempty"`
	Timeout       string            `json:"timeout,omitempty"`
	Confirmation  bool              `json:"confirmation,omitempty"`
}

// ExecutionResponse represents the result of a command execution
type ExecutionResponse struct {
	Success           bool                   `json:"success"`
	Command           string                 `json:"command"`
	Output            string                 `json:"output,omitempty"`
	Error             string                 `json:"error,omitempty"`
	ExitCode          int                    `json:"exit_code"`
	Duration          string                 `json:"duration"`
	DryRun            bool                   `json:"dry_run"`
	Suggestions       []string               `json:"suggestions,omitempty"`
	RelatedCommands   []string               `json:"related_commands,omitempty"`
	Documentation     []string               `json:"documentation,omitempty"`
	SecurityWarnings  []string               `json:"security_warnings,omitempty"`
	ExecutionCapabilities map[string]interface{} `json:"execution_capabilities,omitempty"`
}