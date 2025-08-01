package advanced

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nix-ai-help/internal/ai"
	"nix-ai-help/pkg/logger"
)

// Task represents a single task in a multi-step plan
type Task struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Command     string   `json:"command,omitempty"`
	Status      string   `json:"status"` // pending, in-progress, completed, failed
	Prerequisites []string `json:"prerequisites,omitempty"`
	DependsOn   []string `json:"depends_on,omitempty"`
	EstimatedTime string  `json:"estimated_time,omitempty"`
	ActualTime  string   `json:"actual_time,omitempty"`
	StartTime   string   `json:"start_time,omitempty"`
	EndTime     string   `json:"end_time,omitempty"`
	Result      string   `json:"result,omitempty"`
	Error       string   `json:"error,omitempty"`
}

// TaskPlan represents a complete multi-step task plan
type TaskPlan struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tasks       []Task   `json:"tasks"`
	Status      string   `json:"status"` // pending, in-progress, completed, failed
	StartTime   string   `json:"start_time,omitempty"`
	EndTime     string   `json:"end_time,omitempty"`
	Progress    float64  `json:"progress"`
	EstimatedTotalTime string `json:"estimated_total_time,omitempty"`
	ActualTotalTime    string `json:"actual_total_time,omitempty"`
}

// TaskPlanner implements multi-step task planning for complex operations
type TaskPlanner struct {
	provider ai.Provider
	logger   *logger.Logger
}

// NewTaskPlanner creates a new task planner
func NewTaskPlanner(provider ai.Provider, log *logger.Logger) *TaskPlanner {
	return &TaskPlanner{
		provider: provider,
		logger:   log,
	}
}

// CreateTaskPlan creates a multi-step plan for a complex task
func (tp *TaskPlanner) CreateTaskPlan(ctx context.Context, task string) (*TaskPlan, error) {
	// Create a prompt asking the AI to break down the task into steps
	planningPrompt := fmt.Sprintf(`You are tasked with creating a detailed plan for the following NixOS-related task:

"%s"

Break this task down into a series of discrete, actionable steps. Each step should be something that can be accomplished with a single command or small set of related commands.

For each step, provide:
1. A clear title
2. A detailed description
3. The specific command(s) to execute (if applicable)
4. Any prerequisites or dependencies
5. Estimated time to complete

Structure your response as a JSON object with these fields:
- id: a unique identifier for the plan
- title: a descriptive title for the overall plan
- description: a detailed description of what the plan accomplishes
- tasks: an array of task objects with:
  - id: unique identifier for the task
  - title: brief title of the task
  - description: detailed description of what the task does
  - command: the command(s) to execute (optional)
  - prerequisites: array of prerequisite tasks
  - depends_on: array of task IDs this task depends on
  - estimated_time: estimated time to complete (e.g., "5m", "10m")

Example response format:
{
  "id": "plan-12345",
  "title": "Setting up a Development Environment",
  "description": "Complete plan to set up a development environment for a specific programming language",
  "tasks": [
    {
      "id": "task-1",
      "title": "Install Language Runtime",
      "description": "Install the runtime for the programming language",
      "command": "nix-env -iA nixpkgs.python3",
      "prerequisites": [],
      "depends_on": [],
      "estimated_time": "2m"
    },
    {
      "id": "task-2",
      "title": "Set Up Project Directory",
      "description": "Create and initialize the project directory structure",
      "command": "mkdir myproject && cd myproject && git init",
      "prerequisites": ["task-1"],
      "depends_on": ["task-1"],
      "estimated_time": "1m"
    }
  ]
}`, task)

	response, err := tp.provider.GenerateResponse(ctx, planningPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to query AI for task planning: %w", err)
	}

	// Parse the response
	var plan TaskPlan
	if err := json.Unmarshal([]byte(response), &plan); err != nil {
		// If JSON parsing fails, create a simple plan
		plan = tp.createSimplePlan(task)
	}

	// Initialize plan status
	plan.Status = "pending"
	plan.StartTime = time.Now().Format("2006-01-02 15:04:05")
	plan.Progress = 0.0

	// Initialize task statuses
	for i := range plan.Tasks {
		plan.Tasks[i].Status = "pending"
	}

	return &plan, nil
}

// createSimplePlan creates a simple plan when AI planning fails
func (tp *TaskPlanner) createSimplePlan(task string) TaskPlan {
	return TaskPlan{
		ID:          fmt.Sprintf("plan-%d", time.Now().Unix()),
		Title:       fmt.Sprintf("Plan for: %s", task),
		Description: fmt.Sprintf("Automatically generated plan for task: %s", task),
		Tasks: []Task{
			{
				ID:          "task-1",
				Title:       "Execute Main Task",
				Description: fmt.Sprintf("Execute the main task: %s", task),
				Command:     "# Task execution would happen here",
				Status:      "pending",
				EstimatedTime: "5m",
			},
		},
		Status: "pending",
	}
}

// UpdateTaskStatus updates the status of a task in the plan
func (tp *TaskPlanner) UpdateTaskStatus(plan *TaskPlan, taskID, status, result string) {
	for i := range plan.Tasks {
		if plan.Tasks[i].ID == taskID {
			plan.Tasks[i].Status = status
			plan.Tasks[i].Result = result
			
			// Update timestamps
			now := time.Now().Format("2006-01-02 15:04:05")
			if status == "in-progress" {
				plan.Tasks[i].StartTime = now
			} else if status == "completed" || status == "failed" {
				plan.Tasks[i].EndTime = now
				if plan.Tasks[i].StartTime != "" {
					start, _ := time.Parse("2006-01-02 15:04:05", plan.Tasks[i].StartTime)
					end, _ := time.Parse("2006-01-02 15:04:05", now)
					duration := end.Sub(start)
					plan.Tasks[i].ActualTime = duration.String()
				}
			}
			
			// Update overall plan progress
			tp.updatePlanProgress(plan)
			break
		}
	}
}

// updatePlanProgress calculates and updates the overall plan progress
func (tp *TaskPlanner) updatePlanProgress(plan *TaskPlan) {
	if len(plan.Tasks) == 0 {
		plan.Progress = 0
		return
	}

	completed := 0
	for _, task := range plan.Tasks {
		if task.Status == "completed" {
			completed++
		}
	}

	plan.Progress = float64(completed) / float64(len(plan.Tasks))
	
	// If all tasks are completed, mark plan as completed
	if completed == len(plan.Tasks) {
		plan.Status = "completed"
		plan.EndTime = time.Now().Format("2006-01-02 15:04:05")
		
		// Calculate total actual time
		if plan.StartTime != "" && plan.EndTime != "" {
			start, _ := time.Parse("2006-01-02 15:04:05", plan.StartTime)
			end, _ := time.Parse("2006-01-02 15:04:05", plan.EndTime)
			duration := end.Sub(start)
			plan.ActualTotalTime = duration.String()
		}
	}
}

// FormatTaskPlan formats a task plan for display
func (tp *TaskPlanner) FormatTaskPlan(plan *TaskPlan) string {
	var output strings.Builder
	
	output.WriteString(fmt.Sprintf("# 📋 Task Plan: %s\n\n", plan.Title))
	output.WriteString(fmt.Sprintf("**Description:** %s\n\n", plan.Description))
	output.WriteString(fmt.Sprintf("**Status:** %s  \n", plan.Status))
	output.WriteString(fmt.Sprintf("**Progress:** %.1f%%  \n", plan.Progress*100))
	
	if plan.StartTime != "" {
		output.WriteString(fmt.Sprintf("**Started:** %s  \n", plan.StartTime))
	}
	
	if plan.EndTime != "" {
		output.WriteString(fmt.Sprintf("**Completed:** %s  \n", plan.EndTime))
	}
	
	if plan.EstimatedTotalTime != "" {
		output.WriteString(fmt.Sprintf("**Estimated Time:** %s  \n", plan.EstimatedTotalTime))
	}
	
	if plan.ActualTotalTime != "" {
		output.WriteString(fmt.Sprintf("**Actual Time:** %s  \n", plan.ActualTotalTime))
	}
	
	output.WriteString("\n## 📝 Tasks\n\n")
	
	for i, task := range plan.Tasks {
		statusEmoji := "⏳"
		switch task.Status {
		case "in-progress":
			statusEmoji = "🔄"
		case "completed":
			statusEmoji = "✅"
		case "failed":
			statusEmoji = "❌"
		}
		
		output.WriteString(fmt.Sprintf("### %d. %s %s\n", i+1, task.Title, statusEmoji))
		output.WriteString(fmt.Sprintf("**Description:** %s\n", task.Description))
		
		if task.Command != "" {
			output.WriteString(fmt.Sprintf("**Command:** `%s`\n", task.Command))
		}
		
		if len(task.Prerequisites) > 0 {
			output.WriteString(fmt.Sprintf("**Prerequisites:** %s\n", strings.Join(task.Prerequisites, ", ")))
		}
		
		if len(task.DependsOn) > 0 {
			output.WriteString(fmt.Sprintf("**Depends On:** %s\n", strings.Join(task.DependsOn, ", ")))
		}
		
		if task.EstimatedTime != "" {
			output.WriteString(fmt.Sprintf("**Estimated Time:** %s  \n", task.EstimatedTime))
		}
		
		if task.ActualTime != "" {
			output.WriteString(fmt.Sprintf("**Actual Time:** %s  \n", task.ActualTime))
		}
		
		if task.Result != "" {
			output.WriteString(fmt.Sprintf("**Result:** %s\n", task.Result))
		}
		
		if task.Error != "" {
			output.WriteString(fmt.Sprintf("**Error:** %s\n", task.Error))
		}
		
		output.WriteString("\n")
	}
	
	return output.String()
}

// GetNextPendingTask returns the next task that can be executed
func (tp *TaskPlanner) GetNextPendingTask(plan *TaskPlan) *Task {
	for i := range plan.Tasks {
		task := &plan.Tasks[i]
		if task.Status == "pending" && tp.areDependenciesMet(plan, task) {
			return task
		}
	}
	return nil
}

// areDependenciesMet checks if all dependencies for a task are met
func (tp *TaskPlanner) areDependenciesMet(plan *TaskPlan, task *Task) bool {
	// Check if all tasks this task depends on are completed
	for _, depID := range task.DependsOn {
		depFound := false
		for _, planTask := range plan.Tasks {
			if planTask.ID == depID {
				depFound = true
				if planTask.Status != "completed" {
					return false // Dependency not completed
				}
				break
			}
		}
		if !depFound {
			return false // Dependency not found
		}
	}
	return true // All dependencies met
}