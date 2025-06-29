package workflow

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ConditionEvaluator handles evaluation of workflow conditions
type ConditionEvaluator struct {
	context map[string]interface{}
}

// NewConditionEvaluator creates a new condition evaluator
func NewConditionEvaluator() *ConditionEvaluator {
	return &ConditionEvaluator{
		context: make(map[string]interface{}),
	}
}

// SetContext sets the evaluation context
func (ce *ConditionEvaluator) SetContext(ctx map[string]interface{}) {
	ce.context = ctx
}

// EvaluateCondition evaluates a single condition
func (ce *ConditionEvaluator) EvaluateCondition(condition Condition) (bool, error) {
	switch condition.Type {
	case "file_exists":
		return ce.evaluateFileExists(condition)
	case "file_contains":
		return ce.evaluateFileContains(condition)
	case "command_success":
		return ce.evaluateCommandSuccess(condition)
	case "variable_equals":
		return ce.evaluateVariableEquals(condition)
	case "variable_contains":
		return ce.evaluateVariableContains(condition)
	case "time_condition":
		return ce.evaluateTimeCondition(condition)
	case "system_condition":
		return ce.evaluateSystemCondition(condition)
	case "expression":
		return ce.evaluateExpression(condition)
	default:
		return false, fmt.Errorf("unknown condition type: %s", condition.Type)
	}
}

// EvaluateConditions evaluates multiple conditions with logic operators
func (ce *ConditionEvaluator) EvaluateConditions(conditions []Condition, operator string) (bool, error) {
	if len(conditions) == 0 {
		return true, nil
	}

	if len(conditions) == 1 {
		return ce.EvaluateCondition(conditions[0])
	}

	switch strings.ToLower(operator) {
	case "and", "":
		return ce.evaluateAndConditions(conditions)
	case "or":
		return ce.evaluateOrConditions(conditions)
	default:
		return false, fmt.Errorf("unknown operator: %s", operator)
	}
}

// evaluateAndConditions evaluates conditions with AND logic
func (ce *ConditionEvaluator) evaluateAndConditions(conditions []Condition) (bool, error) {
	for _, condition := range conditions {
		result, err := ce.EvaluateCondition(condition)
		if err != nil {
			return false, err
		}
		if !result {
			return false, nil
		}
	}
	return true, nil
}

// evaluateOrConditions evaluates conditions with OR logic
func (ce *ConditionEvaluator) evaluateOrConditions(conditions []Condition) (bool, error) {
	for _, condition := range conditions {
		result, err := ce.EvaluateCondition(condition)
		if err != nil {
			return false, err
		}
		if result {
			return true, nil
		}
	}
	return false, nil
}

// evaluateFileExists checks if a file exists
func (ce *ConditionEvaluator) evaluateFileExists(condition Condition) (bool, error) {
	path, ok := condition.Parameters["path"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'path' parameter for file_exists condition")
	}

	path = ce.expandVariables(path)
	_, err := os.Stat(path)
	return err == nil, nil
}

// evaluateFileContains checks if a file contains specific content
func (ce *ConditionEvaluator) evaluateFileContains(condition Condition) (bool, error) {
	path, ok := condition.Parameters["path"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'path' parameter for file_contains condition")
	}

	content, ok := condition.Parameters["content"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'content' parameter for file_contains condition")
	}

	path = ce.expandVariables(path)
	content = ce.expandVariables(content)

	data, err := os.ReadFile(path)
	if err != nil {
		return false, nil // File doesn't exist or can't be read
	}

	useRegex, _ := condition.Parameters["regex"].(bool)
	if useRegex {
		match, err := regexp.MatchString(content, string(data))
		if err != nil {
			return false, fmt.Errorf("invalid regex pattern: %v", err)
		}
		return match, nil
	}

	return strings.Contains(string(data), content), nil
}

// evaluateCommandSuccess checks if a command executes successfully
func (ce *ConditionEvaluator) evaluateCommandSuccess(condition Condition) (bool, error) {
	command, ok := condition.Parameters["command"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'command' parameter for command_success condition")
	}

	command = ce.expandVariables(command)

	// For now, we'll simulate command execution
	// In a real implementation, you would execute the command
	// and check its exit code
	return true, nil
}

// evaluateVariableEquals checks if a variable equals a specific value
func (ce *ConditionEvaluator) evaluateVariableEquals(condition Condition) (bool, error) {
	variable, ok := condition.Parameters["variable"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'variable' parameter for variable_equals condition")
	}

	expected, ok := condition.Parameters["value"]
	if !ok {
		return false, fmt.Errorf("missing 'value' parameter for variable_equals condition")
	}

	actual, exists := ce.context[variable]
	if !exists {
		return false, nil
	}

	return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected), nil
}

// evaluateVariableContains checks if a variable contains a specific value
func (ce *ConditionEvaluator) evaluateVariableContains(condition Condition) (bool, error) {
	variable, ok := condition.Parameters["variable"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'variable' parameter for variable_contains condition")
	}

	substring, ok := condition.Parameters["substring"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'substring' parameter for variable_contains condition")
	}

	actual, exists := ce.context[variable]
	if !exists {
		return false, nil
	}

	actualStr := fmt.Sprintf("%v", actual)
	return strings.Contains(actualStr, substring), nil
}

// evaluateTimeCondition checks time-based conditions
func (ce *ConditionEvaluator) evaluateTimeCondition(condition Condition) (bool, error) {
	conditionType, ok := condition.Parameters["condition"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'condition' parameter for time_condition")
	}

	now := time.Now()

	switch conditionType {
	case "before":
		timeStr, ok := condition.Parameters["time"].(string)
		if !ok {
			return false, fmt.Errorf("missing or invalid 'time' parameter for time_condition")
		}

		targetTime, err := time.Parse("15:04", timeStr)
		if err != nil {
			return false, fmt.Errorf("invalid time format: %v", err)
		}

		// Compare only time, not date
		nowTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), now.Second(), 0, time.UTC)
		targetTime = time.Date(0, 1, 1, targetTime.Hour(), targetTime.Minute(), 0, 0, time.UTC)

		return nowTime.Before(targetTime), nil

	case "after":
		timeStr, ok := condition.Parameters["time"].(string)
		if !ok {
			return false, fmt.Errorf("missing or invalid 'time' parameter for time_condition")
		}

		targetTime, err := time.Parse("15:04", timeStr)
		if err != nil {
			return false, fmt.Errorf("invalid time format: %v", err)
		}

		// Compare only time, not date
		nowTime := time.Date(0, 1, 1, now.Hour(), now.Minute(), now.Second(), 0, time.UTC)
		targetTime = time.Date(0, 1, 1, targetTime.Hour(), targetTime.Minute(), 0, 0, time.UTC)

		return nowTime.After(targetTime), nil

	case "weekday":
		weekday, ok := condition.Parameters["weekday"].(string)
		if !ok {
			return false, fmt.Errorf("missing or invalid 'weekday' parameter for time_condition")
		}

		targetWeekday, err := parseWeekday(weekday)
		if err != nil {
			return false, err
		}

		return now.Weekday() == targetWeekday, nil

	case "day_of_month":
		day, ok := condition.Parameters["day"].(int)
		if !ok {
			return false, fmt.Errorf("missing or invalid 'day' parameter for time_condition")
		}

		return now.Day() == day, nil

	default:
		return false, fmt.Errorf("unknown time condition: %s", conditionType)
	}
}

// evaluateSystemCondition checks system-based conditions
func (ce *ConditionEvaluator) evaluateSystemCondition(condition Condition) (bool, error) {
	conditionType, ok := condition.Parameters["condition"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'condition' parameter for system_condition")
	}

	switch conditionType {
	case "load_average":
		threshold, ok := condition.Parameters["threshold"].(float64)
		if !ok {
			return false, fmt.Errorf("missing or invalid 'threshold' parameter for load_average condition")
		}

		// For now, we'll simulate load average check
		// In a real implementation, you would read from /proc/loadavg
		_ = threshold
		return true, nil

	case "disk_space":
		path, ok := condition.Parameters["path"].(string)
		if !ok {
			return false, fmt.Errorf("missing or invalid 'path' parameter for disk_space condition")
		}

		threshold, ok := condition.Parameters["threshold"].(float64)
		if !ok {
			return false, fmt.Errorf("missing or invalid 'threshold' parameter for disk_space condition")
		}

		// For now, we'll simulate disk space check
		// In a real implementation, you would use syscall.Statfs
		_ = path
		_ = threshold
		return true, nil

	case "memory_usage":
		threshold, ok := condition.Parameters["threshold"].(float64)
		if !ok {
			return false, fmt.Errorf("missing or invalid 'threshold' parameter for memory_usage condition")
		}

		// For now, we'll simulate memory usage check
		// In a real implementation, you would read from /proc/meminfo
		_ = threshold
		return true, nil

	default:
		return false, fmt.Errorf("unknown system condition: %s", conditionType)
	}
}

// evaluateExpression evaluates a simple expression
func (ce *ConditionEvaluator) evaluateExpression(condition Condition) (bool, error) {
	expression, ok := condition.Parameters["expression"].(string)
	if !ok {
		return false, fmt.Errorf("missing or invalid 'expression' parameter for expression condition")
	}

	expression = ce.expandVariables(expression)

	// Simple expression evaluation
	// For now, we'll handle basic comparisons
	return ce.evaluateSimpleExpression(expression)
}

// evaluateSimpleExpression evaluates basic expressions like "x > 5", "y == 'test'"
func (ce *ConditionEvaluator) evaluateSimpleExpression(expression string) (bool, error) {
	// Handle equality
	if strings.Contains(expression, "==") {
		parts := strings.SplitN(expression, "==", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid expression format: %s", expression)
		}

		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])

		return left == right, nil
	}

	// Handle inequality
	if strings.Contains(expression, "!=") {
		parts := strings.SplitN(expression, "!=", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid expression format: %s", expression)
		}

		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])

		return left != right, nil
	}

	// Handle greater than
	if strings.Contains(expression, ">") {
		parts := strings.SplitN(expression, ">", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid expression format: %s", expression)
		}

		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])

		leftNum, err1 := strconv.ParseFloat(left, 64)
		rightNum, err2 := strconv.ParseFloat(right, 64)

		if err1 != nil || err2 != nil {
			return false, fmt.Errorf("cannot compare non-numeric values: %s", expression)
		}

		return leftNum > rightNum, nil
	}

	// Handle less than
	if strings.Contains(expression, "<") {
		parts := strings.SplitN(expression, "<", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid expression format: %s", expression)
		}

		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])

		leftNum, err1 := strconv.ParseFloat(left, 64)
		rightNum, err2 := strconv.ParseFloat(right, 64)

		if err1 != nil || err2 != nil {
			return false, fmt.Errorf("cannot compare non-numeric values: %s", expression)
		}

		return leftNum < rightNum, nil
	}

	return false, fmt.Errorf("unsupported expression format: %s", expression)
}

// expandVariables expands variables in a string using the context
func (ce *ConditionEvaluator) expandVariables(input string) string {
	result := input

	// Find all variable references like ${var} or {{var}}
	varPattern := regexp.MustCompile(`\$\{([^}]+)\}|\{\{([^}]+)\}\}`)

	matches := varPattern.FindAllStringSubmatch(result, -1)
	for _, match := range matches {
		var varName string
		if match[1] != "" {
			varName = match[1]
		} else {
			varName = match[2]
		}

		if value, exists := ce.context[varName]; exists {
			result = strings.ReplaceAll(result, match[0], fmt.Sprintf("%v", value))
		}
	}

	return result
}

// parseWeekday parses weekday string to time.Weekday
func parseWeekday(weekday string) (time.Weekday, error) {
	switch strings.ToLower(weekday) {
	case "sunday", "sun":
		return time.Sunday, nil
	case "monday", "mon":
		return time.Monday, nil
	case "tuesday", "tue":
		return time.Tuesday, nil
	case "wednesday", "wed":
		return time.Wednesday, nil
	case "thursday", "thu":
		return time.Thursday, nil
	case "friday", "fri":
		return time.Friday, nil
	case "saturday", "sat":
		return time.Saturday, nil
	default:
		return time.Sunday, fmt.Errorf("invalid weekday: %s", weekday)
	}
}
