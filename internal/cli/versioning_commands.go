package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"nix-ai-help/internal/collaboration/team"
	"nix-ai-help/internal/versioning/repository"
	"nix-ai-help/internal/web"
	"nix-ai-help/pkg/logger"
	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// AddVersioningCommands adds versioning-related commands to the CLI
func AddVersioningCommands(rootCmd *cobra.Command, logger *logger.Logger) {
	// Configuration versioning command group
	versionCmd := &cobra.Command{
		Use:   "version-control",
		Short: "Git-like configuration version control",
		Long:  "Manage NixOS configurations with Git-like version control, branching, and collaboration features",
	}

	// Initialize repository command
	initCmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a new configuration repository",
		Long:  "Initialize a new Git-like repository for NixOS configuration management",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("failed to get absolute path: %w", err)
			}

			repo, err := repository.NewConfigRepository(absPath, logger)
			if err != nil {
				return fmt.Errorf("failed to initialize repository: %w", err)
			}

			fmt.Printf("✅ Initialized configuration repository in %s\n", absPath)
			logger.Info(fmt.Sprintf("Initialized repository at %s", absPath))
			return nil
		},
	}

	// Commit command
	commitCmd := &cobra.Command{
		Use:   "commit",
		Short: "Create a new configuration commit",
		Long:  "Create a new commit with the current configuration files",
		RunE: func(cmd *cobra.Command, args []string) error {
			message, _ := cmd.Flags().GetString("message")
			if message == "" {
				return fmt.Errorf("commit message is required (use -m flag)")
			}

			repoPath, _ := cmd.Flags().GetString("repo")
			if repoPath == "" {
				repoPath = "."
			}

			repo, err := repository.NewConfigRepository(repoPath, logger)
			if err != nil {
				return fmt.Errorf("failed to open repository: %w", err)
			}

			// Get files from current directory
			files, err := getConfigFiles(repoPath)
			if err != nil {
				return fmt.Errorf("failed to get config files: %w", err)
			}

			metadata := make(map[string]string)
			if author, _ := cmd.Flags().GetString("author"); author != "" {
				metadata["author"] = author
			}

			snapshot, err := repo.Commit(context.Background(), message, files, metadata)
			if err != nil {
				return fmt.Errorf("failed to create commit: %w", err)
			}

			fmt.Printf("✅ Created commit %s\n", snapshot.ID[:8])
			fmt.Printf("📝 Message: %s\n", message)
			fmt.Printf("📁 Files: %d\n", len(files))
			return nil
		},
	}
	commitCmd.Flags().StringP("message", "m", "", "Commit message")
	commitCmd.Flags().StringP("author", "a", "", "Commit author")
	commitCmd.Flags().StringP("repo", "r", "", "Repository path")

	// Branch command
	branchCmd := &cobra.Command{
		Use:   "branch",
		Short: "Manage configuration branches",
		Long:  "Create, list, switch, and delete configuration branches",
	}

	// List branches
	branchListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all branches",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := cmd.Flags().GetString("repo")
			if repoPath == "" {
				repoPath = "."
			}

			repo, err := repository.NewConfigRepository(repoPath, logger)
			if err != nil {
				return fmt.Errorf("failed to open repository: %w", err)
			}

			branchManager := repository.NewBranchManager(repo, logger)
			branches, err := branchManager.ListBranchesWithInfo(context.Background())
			if err != nil {
				return fmt.Errorf("failed to list branches: %w", err)
			}

			fmt.Printf("\n%s\n", utils.FormatHeader("Configuration Branches"))
			for _, branch := range branches {
				currentMarker := ""
				if branch.Current {
					currentMarker = " (current)"
				}
				fmt.Printf("🌿 %s%s - %s\n", branch.Name, currentMarker, branch.Description)
				fmt.Printf("   Environment: %s | Protected: %v\n", branch.Environment, branch.Protected)
			}
			fmt.Println()
			return nil
		},
	}
	branchListCmd.Flags().StringP("repo", "r", "", "Repository path")

	// Create branch
	branchCreateCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]
			description, _ := cmd.Flags().GetString("description")

			repoPath, _ := cmd.Flags().GetString("repo")
			if repoPath == "" {
				repoPath = "."
			}

			repo, err := repository.NewConfigRepository(repoPath, logger)
			if err != nil {
				return fmt.Errorf("failed to open repository: %w", err)
			}

			branchManager := repository.NewBranchManager(repo, logger)

			// Check if this should be a feature branch
			if feature, _ := cmd.Flags().GetBool("feature"); feature {
				_, err = branchManager.CreateFeatureBranch(context.Background(), branchName, description)
			} else {
				err = repo.CreateBranch(context.Background(), branchName, "")
			}

			if err != nil {
				return fmt.Errorf("failed to create branch: %w", err)
			}

			fmt.Printf("✅ Created branch: %s\n", branchName)
			return nil
		},
	}
	branchCreateCmd.Flags().StringP("description", "d", "", "Branch description")
	branchCreateCmd.Flags().BoolP("feature", "f", false, "Create as feature branch")
	branchCreateCmd.Flags().StringP("repo", "r", "", "Repository path")

	// Switch branch
	branchSwitchCmd := &cobra.Command{
		Use:   "switch <name>",
		Short: "Switch to a different branch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branchName := args[0]

			repoPath, _ := cmd.Flags().GetString("repo")
			if repoPath == "" {
				repoPath = "."
			}

			repo, err := repository.NewConfigRepository(repoPath, logger)
			if err != nil {
				return fmt.Errorf("failed to open repository: %w", err)
			}

			err = repo.SwitchBranch(context.Background(), branchName)
			if err != nil {
				return fmt.Errorf("failed to switch branch: %w", err)
			}

			fmt.Printf("✅ Switched to branch: %s\n", branchName)
			return nil
		},
	}
	branchSwitchCmd.Flags().StringP("repo", "r", "", "Repository path")

	branchCmd.AddCommand(branchListCmd, branchCreateCmd, branchSwitchCmd)

	// History command
	historyCmd := &cobra.Command{
		Use:   "history",
		Short: "View configuration change history",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath, _ := cmd.Flags().GetString("repo")
			if repoPath == "" {
				repoPath = "."
			}

			repo, err := repository.NewConfigRepository(repoPath, logger)
			if err != nil {
				return fmt.Errorf("failed to open repository: %w", err)
			}

			snapshots, err := repo.ListSnapshots(context.Background())
			if err != nil {
				return fmt.Errorf("failed to get history: %w", err)
			}

			limit, _ := cmd.Flags().GetInt("limit")
			if limit > 0 && limit < len(snapshots) {
				snapshots = snapshots[:limit]
			}

			fmt.Printf("\n%s\n", utils.FormatHeader("Configuration History"))
			for _, snapshot := range snapshots {
				fmt.Printf("📄 %s - %s\n", snapshot.ID[:8], snapshot.Message)
				fmt.Printf("   Author: %s | Date: %s\n", snapshot.Author, snapshot.Timestamp.Format("2006-01-02 15:04:05"))
				fmt.Printf("   Files: %d | Hash: %s\n", len(snapshot.Files), snapshot.Hash[:16])
				fmt.Println()
			}
			return nil
		},
	}
	historyCmd.Flags().IntP("limit", "l", 10, "Limit number of commits to show")
	historyCmd.Flags().StringP("repo", "r", "", "Repository path")

	versionCmd.AddCommand(initCmd, commitCmd, branchCmd, historyCmd)
	rootCmd.AddCommand(versionCmd)
}

// AddCollaborationCommands adds collaboration-related commands
func AddCollaborationCommands(rootCmd *cobra.Command, logger *logger.Logger) {
	// Team management command group
	teamCmd := &cobra.Command{
		Use:   "team",
		Short: "Manage configuration teams",
		Long:  "Create and manage teams for collaborative configuration development",
	}

	// Create team
	teamCreateCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new team",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			teamName := args[0]
			description, _ := cmd.Flags().GetString("description")

			teamManager := team.NewTeamManager(logger)

			// TODO: Get actual user ID from config or authentication
			userID := "current_user"

			newTeam, err := teamManager.CreateTeam(context.Background(), teamName, description, userID)
			if err != nil {
				return fmt.Errorf("failed to create team: %w", err)
			}

			fmt.Printf("✅ Created team: %s\n", newTeam.Name)
			fmt.Printf("📝 Description: %s\n", description)
			fmt.Printf("🆔 Team ID: %s\n", newTeam.ID)
			return nil
		},
	}
	teamCreateCmd.Flags().StringP("description", "d", "", "Team description")

	// List teams
	teamListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all teams",
		RunE: func(cmd *cobra.Command, args []string) error {
			teamManager := team.NewTeamManager(logger)
			teams, err := teamManager.ListTeams(context.Background())
			if err != nil {
				return fmt.Errorf("failed to list teams: %w", err)
			}

			fmt.Printf("\n%s\n", utils.FormatHeader("Configuration Teams"))
			for _, t := range teams {
				fmt.Printf("👥 %s (%s)\n", t.Name, t.ID)
				fmt.Printf("   Description: %s\n", t.Description)
				fmt.Printf("   Members: %d | Created: %s\n", len(t.Members), t.CreatedAt.Format("2006-01-02"))
				fmt.Println()
			}
			return nil
		},
	}

	teamCmd.AddCommand(teamCreateCmd, teamListCmd)
	rootCmd.AddCommand(teamCmd)
}

// AddWebInterfaceCommands adds web interface commands
func AddWebInterfaceCommands(rootCmd *cobra.Command, logger *logger.Logger) {
	// Web interface command
	webCmd := &cobra.Command{
		Use:   "web",
		Short: "Start web interface",
		Long:  "Start the web-based configuration management interface",
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetInt("port")
			repoPath, _ := cmd.Flags().GetString("repo")
			if repoPath == "" {
				repoPath = "."
			}

			// Initialize components
			repo, err := repository.NewConfigRepository(repoPath, logger)
			if err != nil {
				return fmt.Errorf("failed to open repository: %w", err)
			}

			teamManager := team.NewTeamManager(logger)
			server := web.NewServer(port, teamManager, repo, logger)

			fmt.Printf("🌐 Starting web interface on port %d\n", port)
			fmt.Printf("📂 Repository: %s\n", repoPath)
			fmt.Printf("🔗 Open: http://localhost:%d\n", port)

			return server.Start(context.Background())
		},
	}
	webCmd.Flags().IntP("port", "p", 8080, "Port to serve on")
	webCmd.Flags().StringP("repo", "r", "", "Repository path")

	rootCmd.AddCommand(webCmd)
}

// Helper functions

func getConfigFiles(repoPath string) (map[string]string, error) {
	files := make(map[string]string)

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and hidden files
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Only include configuration files
		if isConfigFile(path) {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Use relative path as key
			relPath, err := filepath.Rel(repoPath, path)
			if err != nil {
				return err
			}

			files[relPath] = string(content)
		}

		return nil
	})

	return files, err
}

func isConfigFile(path string) bool {
	configExtensions := []string{".nix", ".yaml", ".yml", ".json", ".toml", ".conf"}
	ext := filepath.Ext(path)

	for _, configExt := range configExtensions {
		if ext == configExt {
			return true
		}
	}

	// Also include common config file names
	baseName := filepath.Base(path)
	configNames := []string{"configuration.nix", "hardware-configuration.nix", "flake.nix", "home.nix"}

	for _, configName := range configNames {
		if baseName == configName {
			return true
		}
	}

	return false
}
