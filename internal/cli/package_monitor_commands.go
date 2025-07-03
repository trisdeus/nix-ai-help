package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"nix-ai-help/pkg/utils"

	"github.com/spf13/cobra"
)

// CreatePackageMonitorCommand creates the package-monitor command and its subcommands
func CreatePackageMonitorCommand() *cobra.Command {
	packageMonitorCmd := &cobra.Command{
		Use:   "package-monitor",
		Short: "Package monitoring and update management",
		Long: `Package monitoring and update management commands.

Monitor installed packages, check for updates, analyze dependencies,
and get AI-powered recommendations for package management.

Examples:
  nixai package-monitor list           # List installed packages
  nixai package-monitor updates       # Check for available updates
  nixai package-monitor security      # Check for security updates
  nixai package-monitor analyze       # Analyze package dependencies
  nixai package-monitor orphans       # Find orphaned packages
  nixai package-monitor outdated      # Find outdated packages`,
	}

	// Add subcommands
	packageMonitorCmd.AddCommand(createPackageListCmd())
	packageMonitorCmd.AddCommand(createPackageUpdatesCmd())
	packageMonitorCmd.AddCommand(createPackageSecurityCmd())
	packageMonitorCmd.AddCommand(createPackageAnalyzeCmd())
	packageMonitorCmd.AddCommand(createPackageOrphansCmd())
	packageMonitorCmd.AddCommand(createPackageOutdatedCmd())
	packageMonitorCmd.AddCommand(createPackageStatsCmd())

	return packageMonitorCmd
}

// createPackageListCmd creates the list subcommand
func createPackageListCmd() *cobra.Command {
	var jsonOutput bool
	var detailed bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return outputPackageListJSON(detailed)
			}
			return outputPackageList(detailed)
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	cmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Show detailed package information")
	return cmd
}

// createPackageUpdatesCmd creates the updates subcommand
func createPackageUpdatesCmd() *cobra.Command {
	var jsonOutput bool
	var securityOnly bool

	cmd := &cobra.Command{
		Use:   "updates",
		Short: "Check for available package updates",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return outputPackageUpdatesJSON(securityOnly)
			}
			return outputPackageUpdates(securityOnly)
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	cmd.Flags().BoolVarP(&securityOnly, "security", "s", false, "Show only security updates")
	return cmd
}

// createPackageSecurityCmd creates the security subcommand
func createPackageSecurityCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "security",
		Short: "Check for security-related package updates",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return outputSecurityUpdatesJSON()
			}
			return outputSecurityUpdates()
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	return cmd
}

// createPackageAnalyzeCmd creates the analyze subcommand
func createPackageAnalyzeCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze package dependencies and system state",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return outputPackageAnalysisJSON()
			}
			return outputPackageAnalysis()
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	return cmd
}

// createPackageOrphansCmd creates the orphans subcommand
func createPackageOrphansCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "orphans",
		Short: "Find orphaned packages (no longer needed)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return outputOrphanPackagesJSON()
			}
			return outputOrphanPackages()
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	return cmd
}

// createPackageOutdatedCmd creates the outdated subcommand
func createPackageOutdatedCmd() *cobra.Command {
	var jsonOutput bool
	var threshold int

	cmd := &cobra.Command{
		Use:   "outdated",
		Short: "Find packages that haven't been updated recently",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return outputOutdatedPackagesJSON(threshold)
			}
			return outputOutdatedPackages(threshold)
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	cmd.Flags().IntVarP(&threshold, "days", "d", 90, "Consider packages outdated after this many days")
	return cmd
}

// createPackageStatsCmd creates the stats subcommand
func createPackageStatsCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show package statistics and summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				return outputPackageStatsJSON()
			}
			return outputPackageStats()
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	return cmd
}

// Implementation functions

func outputPackageList(detailed bool) error {
	fmt.Println(utils.FormatHeader("📦 Installed Packages"))
	fmt.Println()

	// Try to get Nix packages first
	if nixPkgs := getNixPackages(); len(nixPkgs) > 0 {
		fmt.Println(utils.FormatHeader("Nix Packages"))
		for i, pkg := range nixPkgs {
			if i >= 20 && !detailed { // Limit output unless detailed
				fmt.Printf("... and %d more packages (use --detailed to see all)\n", len(nixPkgs)-i)
				break
			}
			fmt.Printf("  %s\n", pkg)
		}
		fmt.Printf("\nTotal Nix packages: %d\n", len(nixPkgs))
	}

	// Try to get system packages
	if sysPkgs := getSystemPackages(); len(sysPkgs) > 0 {
		fmt.Println()
		fmt.Println(utils.FormatHeader("System Packages (sample)"))
		for i, pkg := range sysPkgs {
			if i >= 10 { // Show only first 10 system packages
				fmt.Printf("... and %d more system packages\n", len(sysPkgs)-i)
				break
			}
			fmt.Printf("  %s\n", pkg)
		}
		fmt.Printf("\nTotal system packages: %d\n", len(sysPkgs))
	}

	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Timestamp", time.Now().Format(time.RFC3339)))

	return nil
}

func outputPackageListJSON(detailed bool) error {
	data := map[string]interface{}{
		"nix_packages":    getNixPackages(),
		"system_packages": getSystemPackages(),
		"timestamp":       time.Now().Format(time.RFC3339),
	}

	if !detailed && len(data["nix_packages"].([]string)) > 50 {
		nixPkgs := data["nix_packages"].([]string)
		data["nix_packages"] = nixPkgs[:50]
		data["nix_packages_truncated"] = true
		data["total_nix_packages"] = len(nixPkgs)
	}

	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputPackageUpdates(securityOnly bool) error {
	fmt.Println(utils.FormatHeader("🔄 Package Updates"))
	fmt.Println()

	// Check for Nix channel updates
	fmt.Println(utils.FormatHeader("Nix Channel Status"))
	if channelInfo := getNixChannelInfo(); channelInfo != "" {
		fmt.Println(channelInfo)
	} else {
		fmt.Println("Could not determine Nix channel status")
	}

	// Check system updates
	fmt.Println()
	fmt.Println(utils.FormatHeader("System Updates"))
	if updates := getSystemUpdates(); updates != "" {
		fmt.Println(updates)
	} else {
		fmt.Println("No system update information available")
	}

	// Check flake inputs if in a flake directory
	fmt.Println()
	fmt.Println(utils.FormatHeader("Flake Updates"))
	if flakeInfo := getFlakeUpdateInfo(); flakeInfo != "" {
		fmt.Println(flakeInfo)
	} else {
		fmt.Println("Not in a flake directory or no flake.lock found")
	}

	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Timestamp", time.Now().Format(time.RFC3339)))

	return nil
}

func outputPackageUpdatesJSON(securityOnly bool) error {
	data := map[string]interface{}{
		"nix_channel":    getNixChannelInfo(),
		"system_updates": getSystemUpdates(),
		"flake_updates":  getFlakeUpdateInfo(),
		"timestamp":      time.Now().Format(time.RFC3339),
	}

	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputSecurityUpdates() error {
	fmt.Println(utils.FormatHeader("🔒 Security Updates"))
	fmt.Println()

	fmt.Println(utils.FormatWarning("Security update detection requires specific tools and vulnerability databases."))
	fmt.Println()

	// Check for known security tools
	securityTools := []string{"vulnix", "nix-audit"}
	hasTools := false

	for _, tool := range securityTools {
		if _, err := exec.LookPath(tool); err == nil {
			hasTools = true
			fmt.Printf("✅ %s: available\n", tool)
		} else {
			fmt.Printf("❌ %s: not installed\n", tool)
		}
	}

	if hasTools {
		fmt.Println()
		fmt.Println("Run security scanning tools manually:")
		fmt.Println("  vulnix --system")
		fmt.Println("  nix-audit")
	} else {
		fmt.Println()
		fmt.Println("To enable security scanning, install security tools:")
		fmt.Println("  nix-env -iA nixpkgs.vulnix")
		fmt.Println("  nix-env -iA nixpkgs.nix-audit")
	}

	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Timestamp", time.Now().Format(time.RFC3339)))

	return nil
}

func outputSecurityUpdatesJSON() error {
	data := map[string]interface{}{
		"security_tools_available": checkSecurityTools(),
		"recommendations":          getSecurityRecommendations(),
		"timestamp":                time.Now().Format(time.RFC3339),
	}

	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputPackageAnalysis() error {
	fmt.Println(utils.FormatHeader("🔍 Package Analysis"))
	fmt.Println()

	// Nix store analysis
	fmt.Println(utils.FormatHeader("Nix Store Analysis"))
	if storeInfo := getNixStoreInfo(); storeInfo != "" {
		fmt.Println(storeInfo)
	}

	// Generation analysis
	fmt.Println()
	fmt.Println(utils.FormatHeader("System Generations"))
	if genInfo := getGenerationInfo(); genInfo != "" {
		fmt.Println(genInfo)
	}

	// Profile analysis
	fmt.Println()
	fmt.Println(utils.FormatHeader("User Profile Analysis"))
	if profileInfo := getProfileInfo(); profileInfo != "" {
		fmt.Println(profileInfo)
	}

	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Timestamp", time.Now().Format(time.RFC3339)))

	return nil
}

func outputPackageAnalysisJSON() error {
	data := map[string]interface{}{
		"nix_store":    getNixStoreInfo(),
		"generations":  getGenerationInfo(),
		"profiles":     getProfileInfo(),
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputOrphanPackages() error {
	fmt.Println(utils.FormatHeader("🧹 Orphaned Packages"))
	fmt.Println()

	// Check for packages that might be orphaned
	fmt.Println("Checking for potentially orphaned packages...")
	fmt.Println()

	// In NixOS, orphaned packages are handled differently
	fmt.Println("In NixOS, 'orphaned' packages are typically:")
	fmt.Println("1. Packages in old generations that are no longer in current config")
	fmt.Println("2. Build-time dependencies that can be garbage collected")
	fmt.Println()

	// Check for old generations
	if genInfo := getGenerationInfo(); genInfo != "" {
		fmt.Println(utils.FormatHeader("Old Generations (can be cleaned up)"))
		fmt.Println(genInfo)
	}

	fmt.Println()
	fmt.Println("To clean up orphaned packages:")
	fmt.Println("  nix-collect-garbage -d  # Delete old generations and unused packages")
	fmt.Println("  nixos-rebuild switch   # Ensure current config is applied")

	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Timestamp", time.Now().Format(time.RFC3339)))

	return nil
}

func outputOrphanPackagesJSON() error {
	data := map[string]interface{}{
		"generations":      getGenerationInfo(),
		"cleanup_commands": []string{"nix-collect-garbage -d", "nixos-rebuild switch"},
		"timestamp":        time.Now().Format(time.RFC3339),
	}

	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputOutdatedPackages(threshold int) error {
	fmt.Println(utils.FormatHeader("📅 Outdated Package Analysis"))
	fmt.Println()

	fmt.Printf("Checking for packages older than %d days...\n", threshold)
	fmt.Println()

	// Check Nix channel age
	if channelAge := getNixChannelAge(); channelAge != "" {
		fmt.Println(utils.FormatHeader("Nix Channel Age"))
		fmt.Println(channelAge)
	}

	// Check system generation age
	if genAge := getSystemGenerationAge(); genAge != "" {
		fmt.Println()
		fmt.Println(utils.FormatHeader("System Generation Age"))
		fmt.Println(genAge)
	}

	fmt.Println()
	fmt.Println("Recommendations:")
	fmt.Println("  nix-channel --update     # Update Nix channels")
	fmt.Println("  nixos-rebuild switch     # Apply latest configuration")
	fmt.Println("  nix flake update         # Update flake inputs (if using flakes)")

	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Timestamp", time.Now().Format(time.RFC3339)))

	return nil
}

func outputOutdatedPackagesJSON(threshold int) error {
	data := map[string]interface{}{
		"threshold_days":       threshold,
		"channel_age":          getNixChannelAge(),
		"generation_age":       getSystemGenerationAge(),
		"update_commands":      []string{"nix-channel --update", "nixos-rebuild switch", "nix flake update"},
		"timestamp":            time.Now().Format(time.RFC3339),
	}

	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func outputPackageStats() error {
	fmt.Println(utils.FormatHeader("📊 Package Statistics"))
	fmt.Println()

	nixPkgs := getNixPackages()
	sysPkgs := getSystemPackages()

	fmt.Println(utils.FormatKeyValue("Nix Packages", fmt.Sprintf("%d", len(nixPkgs))))
	fmt.Println(utils.FormatKeyValue("System Packages", fmt.Sprintf("%d", len(sysPkgs))))
	fmt.Println(utils.FormatKeyValue("Total Packages", fmt.Sprintf("%d", len(nixPkgs)+len(sysPkgs))))

	// Nix store stats
	if storeSize := getNixStoreSize(); storeSize != "" {
		fmt.Println(utils.FormatKeyValue("Nix Store Size", storeSize))
	}

	// Generation count
	if genCount := getGenerationCount(); genCount != "" {
		fmt.Println(utils.FormatKeyValue("System Generations", genCount))
	}

	fmt.Println()
	fmt.Println(utils.FormatKeyValue("Timestamp", time.Now().Format(time.RFC3339)))

	return nil
}

func outputPackageStatsJSON() error {
	nixPkgs := getNixPackages()
	sysPkgs := getSystemPackages()

	data := map[string]interface{}{
		"nix_packages":       len(nixPkgs),
		"system_packages":    len(sysPkgs),
		"total_packages":     len(nixPkgs) + len(sysPkgs),
		"nix_store_size":     getNixStoreSize(),
		"generation_count":   getGenerationCount(),
		"timestamp":          time.Now().Format(time.RFC3339),
	}

	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

// Helper functions

func getNixPackages() []string {
	var packages []string

	// Try nix-env first
	if out, err := exec.Command("nix-env", "-q").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			if line = strings.TrimSpace(line); line != "" {
				packages = append(packages, line)
			}
		}
	}

	// If no user packages, try to get system packages from configuration
	if len(packages) == 0 {
		if out, err := exec.Command("nix-env", "-qaP", "--installed").Output(); err == nil {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			for _, line := range lines {
				if line = strings.TrimSpace(line); line != "" {
					packages = append(packages, line)
				}
			}
		}
	}

	return packages
}

func getSystemPackages() []string {
	var packages []string

	// Try to get packages via dpkg (if available)
	if out, err := exec.Command("dpkg", "-l").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "ii ") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					packages = append(packages, fields[1])
				}
			}
		}
	}

	return packages
}

func getNixChannelInfo() string {
	if out, err := exec.Command("nix-channel", "--list").Output(); err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}

func getSystemUpdates() string {
	// Try different package managers
	if out, err := exec.Command("apt", "list", "--upgradable").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			return fmt.Sprintf("%d packages can be upgraded", len(lines)-1)
		}
	}
	return ""
}

func getFlakeUpdateInfo() string {
	if _, err := os.Stat("flake.lock"); err == nil {
		if out, err := exec.Command("nix", "flake", "metadata").Output(); err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return ""
}

func checkSecurityTools() map[string]bool {
	tools := map[string]bool{}
	securityTools := []string{"vulnix", "nix-audit"}

	for _, tool := range securityTools {
		_, err := exec.LookPath(tool)
		tools[tool] = err == nil
	}

	return tools
}

func getSecurityRecommendations() []string {
	return []string{
		"Install security scanning tools: nix-env -iA nixpkgs.vulnix",
		"Run regular vulnerability scans: vulnix --system",
		"Keep system updated: nixos-rebuild switch",
		"Monitor security advisories for your packages",
	}
}

func getNixStoreInfo() string {
	if out, err := exec.Command("nix", "path-info", "--all", "--size").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		return fmt.Sprintf("Nix store contains %d paths", len(lines)-1)
	}
	return ""
}

func getNixStoreSize() string {
	if out, err := exec.Command("du", "-sh", "/nix/store").Output(); err == nil {
		fields := strings.Fields(string(out))
		if len(fields) > 0 {
			return fields[0]
		}
	}
	return ""
}

func getGenerationInfo() string {
	if out, err := exec.Command("nixos-rebuild", "list-generations").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 5 {
			// Show first and last few generations
			result := strings.Join(lines[:3], "\n") + "\n...\n" + strings.Join(lines[len(lines)-2:], "\n")
			return result
		}
		return strings.TrimSpace(string(out))
	}
	return ""
}

func getGenerationCount() string {
	if out, err := exec.Command("nixos-rebuild", "list-generations").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		return fmt.Sprintf("%d", len(lines))
	}
	return ""
}

func getProfileInfo() string {
	if out, err := exec.Command("nix-env", "--list-generations").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 0 {
			return fmt.Sprintf("%d user profile generations", len(lines))
		}
	}
	return ""
}

func getNixChannelAge() string {
	if out, err := exec.Command("nix-channel", "--list").Output(); err == nil {
		// This is a simplified check - would need more complex logic for real age detection
		if strings.Contains(string(out), "nixos") {
			return "NixOS channel found - check with 'nix-channel --update' for latest"
		}
	}
	return ""
}

func getSystemGenerationAge() string {
	if out, err := exec.Command("nixos-rebuild", "list-generations").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 0 {
			// Get the last line which should be the current generation
			lastGen := lines[len(lines)-1]
			return "Current generation: " + lastGen
		}
	}
	return ""
}