package updater

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C7C7C"))

	scriptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
)

// Runner handles the execution of update scripts
type Runner struct {
	FS      fs.FS
	Logger  *zap.Logger
	Verbose bool
}

// RunScripts executes the provided scripts in parallel or sequential mode
func (r *Runner) RunScripts(scripts []string, parallel bool) {
	fmt.Println(titleStyle.Render("🚀 Update Command NG - Modern System Updater"))
	if r.Verbose {
		fmt.Println(titleStyle.Render("(Verbose Mode)"))
	}
	fmt.Printf("Starting %d update scripts...\n", len(scripts))

	if parallel {
		fmt.Println(infoStyle.Render("Running in parallel mode"))
	} else {
		fmt.Println(infoStyle.Render("Running in sequential mode"))
	}
	fmt.Println()

	start := time.Now()
	results := make(map[string]*ScriptResult)
	scriptDescriptions := GetScriptDescriptions()

	// Initialize results
	for _, script := range scripts {
		results[script] = &ScriptResult{
			Name:        script,
			Status:      StatusPending,
			Description: scriptDescriptions[script],
		}
	}

	var wg sync.WaitGroup
	var outputMutex sync.Mutex
	resultsChan := make(chan ScriptResult, len(scripts))

	// Function to run a single script
	runSingleScript := func(scriptName string) {
		defer wg.Done()

		name := strings.TrimPrefix(scriptName, "update-")
		description := scriptDescriptions[scriptName]

		// Print starting message with mutex protection
		outputMutex.Lock()
		fmt.Printf("%s %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Render("⟳"),
			scriptStyle.Render(name))
		if description != "" {
			fmt.Printf(" - %s", infoStyle.Render(description))
		}
		fmt.Println(infoStyle.Render(" starting..."))
		outputMutex.Unlock()

		result := r.RunScript(scriptName, description)

		// Print result immediately with mutex protection
		outputMutex.Lock()
		switch result.Status {
		case StatusSuccess:
			duration := result.Duration.Round(100 * time.Millisecond)
			fmt.Printf("%s %s", successStyle.Render("✓"), scriptStyle.Render(name))
			if description != "" {
				fmt.Printf(" - %s", infoStyle.Render(description))
			}
			fmt.Printf(" %s\n", infoStyle.Render(fmt.Sprintf("(%v)", duration)))

			if r.Verbose && result.Output != "" {
				lines := strings.Split(strings.TrimSpace(result.Output), "\n")
				var interesting []string
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.Contains(line, "updated") || strings.Contains(line, "installed") ||
						strings.Contains(line, "upgraded") || strings.Contains(line, "downloading") ||
						strings.Contains(line, "fetching") || strings.Contains(line, "pulling") {
						if len(line) > 100 {
							line = line[:97] + "..."
						}
						interesting = append(interesting, line)
						if len(interesting) >= 2 {
							break
						}
					}
				}

				for _, line := range interesting {
					fmt.Printf("  %s\n", successStyle.Render("→ "+line))
				}
			}

		case StatusError:
			fmt.Printf("%s %s", errorStyle.Render("✗"), scriptStyle.Render(name))
			if description != "" {
				fmt.Printf(" - %s", infoStyle.Render(description))
			}

			errorMsg := "failed"
			if result.Error != nil {
				errorMsg = result.Error.Error()
			}
			fmt.Printf(" %s\n", errorStyle.Render(errorMsg))

			if result.Output != "" {
				lines := strings.Split(strings.TrimSpace(result.Output), "\n")
				for i, line := range lines {
					if i >= 3 {
						fmt.Printf("  %s\n", errorStyle.Render("... (truncated)"))
						break
					}
					line = strings.TrimSpace(line)
					if line != "" {
						if len(line) > 100 {
							line = line[:97] + "..."
						}
						fmt.Printf("  %s\n", errorStyle.Render("✗ "+line))
					}
				}
			}

		case StatusSkipped:
			fmt.Printf("%s %s", infoStyle.Render("⊝"), scriptStyle.Render(name))
			if description != "" {
				fmt.Printf(" - %s", infoStyle.Render(description))
			}
			fmt.Printf(" %s", infoStyle.Render("skipped"))

			if r.Verbose && result.Output != "" {
				lines := strings.Split(strings.TrimSpace(result.Output), "\n")
				if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
					reason := strings.TrimSpace(lines[0])
					if len(reason) > 80 {
						reason = reason[:77] + "..."
					}
					fmt.Printf("  %s", infoStyle.Render("→ "+reason))
				}
			}
			fmt.Println()
		}
		outputMutex.Unlock()

		resultsChan <- result
	}

	// Start scripts
	if parallel {
		for _, script := range scripts {
			wg.Add(1)
			go runSingleScript(script)
		}
	} else {
		for _, script := range scripts {
			wg.Add(1)
			runSingleScript(script)
		}
	}

	// Wait for all to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	completed := 0
	var successCount, errorCount, skippedCount int
	for result := range resultsChan {
		results[result.Name] = &result
		completed++

		switch result.Status {
		case StatusSuccess:
			successCount++
		case StatusError:
			errorCount++
		case StatusSkipped:
			skippedCount++
		}
	}

	// Print final summary
	fmt.Println()
	fmt.Printf("📊 %s ", lipgloss.NewStyle().Bold(true).Render("Final Summary:"))
	fmt.Printf("%s %d succeeded", successStyle.Render("✓"), successCount)
	if errorCount > 0 {
		fmt.Printf(", %s %d failed", errorStyle.Render("✗"), errorCount)
	}
	if skippedCount > 0 {
		fmt.Printf(", %s %d skipped", infoStyle.Render("⊝"), skippedCount)
	}

	totalTime := time.Since(start)
	fmt.Printf("\nTotal time: %v\n", totalTime.Round(time.Second))
	fmt.Println("🎉 All updates completed!")
}

func (r *Runner) RunScript(scriptName, description string) ScriptResult {

	start := time.Now()

	// Add a small delay to show the running state
	time.Sleep(100 * time.Millisecond)

	// Read script content
	scriptPath := filepath.Join("bin", scriptName)
	data, err := fs.ReadFile(r.FS, scriptPath)
	if err != nil {
		return ScriptResult{
			Name:        scriptName,
			Status:      StatusError,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to read script: %w", err),
			Description: description,
		}
	}

	// Create a secure temp directory
	tmpDir, err := os.MkdirTemp("", "update-ng-*")
	if err != nil {
		return ScriptResult{
			Name:        scriptName,
			Status:      StatusError,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to create temp dir: %w", err),
			Description: description,
		}
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, scriptName)
	if err := os.WriteFile(tmpFile, data, 0700); err != nil {
		return ScriptResult{
			Name:        scriptName,
			Status:      StatusError,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to write temp script: %w", err),
			Description: description,
		}
	}

	// Run the script
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", tmpFile)
	cmd.Stdin = os.Stdin
	output, err := cmd.CombinedOutput()

	duration := time.Since(start)
	outputStr := string(output)

	// Ensure minimum execution time for visibility (at least 300ms)
	if duration < 300*time.Millisecond {
		time.Sleep(300*time.Millisecond - duration)
		duration = time.Since(start)
	}

	// Log the result
	logFields := []zap.Field{
		zap.String("script", scriptName),
		zap.Duration("duration", duration),
		zap.String("output", outputStr),
	}

	if err != nil {
		// Enhanced error categorization
		if isToolNotAvailable(outputStr) {
			r.Logger.Warn("Script skipped - tool not available", logFields...)
			return ScriptResult{
				Name:        scriptName,
				Status:      StatusSkipped,
				Duration:    duration,
				Output:      outputStr,
				Description: description,
			}
		}

		if isExternallyManaged(outputStr) {
			r.Logger.Info("Script skipped - externally managed", logFields...)
			return ScriptResult{
				Name:        scriptName,
				Status:      StatusSkipped,
				Duration:    duration,
				Output:      outputStr,
				Description: description,
			}
		}

		if isAlreadyUpToDate(outputStr) {
			r.Logger.Info("Script completed - already up to date", logFields...)
			return ScriptResult{
				Name:        scriptName,
				Status:      StatusSuccess,
				Duration:    duration,
				Output:      outputStr,
				Description: description,
			}
		}

		logFields = append(logFields, zap.Error(err))
		r.Logger.Error("Script failed", logFields...)
		return ScriptResult{
			Name:        scriptName,
			Status:      StatusError,
			Duration:    duration,
			Output:      outputStr,
			Error:       err,
			Description: description,
		}
	}

	r.Logger.Info("Script finished successfully", logFields...)
	return ScriptResult{
		Name:        scriptName,
		Status:      StatusSuccess,
		Duration:    duration,
		Output:      outputStr,
		Description: description,
	}
}

func GetAllScripts(fsys fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(fsys, "bin")
	if err != nil {
		return nil, err
	}

	var scriptNames []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "update-") {
			scriptNames = append(scriptNames, entry.Name())
		}
	}
	return scriptNames, nil
}

// FilterScripts applies filtering logic to scripts based on only, skip, and exclude patterns
func FilterScripts(scripts []string, only, skip, exclude []string) []string {
	var filtered []string

	// Apply 'only' filter first (include only matching scripts)
	if len(only) > 0 {
		for _, script := range scripts {
			for _, pattern := range only {
				if matchesPattern(script, pattern) {
					filtered = append(filtered, script)
					break
				}
			}
		}
		scripts = filtered
		filtered = nil
	}

	// Apply 'skip' filter (exclude matching scripts)
	if len(skip) > 0 {
		for _, script := range scripts {
			shouldSkip := false
			for _, pattern := range skip {
				if matchesPattern(script, pattern) {
					shouldSkip = true
					break
				}
			}
			if !shouldSkip {
				filtered = append(filtered, script)
			}
		}
		scripts = filtered
		filtered = nil
	}

	// Apply 'exclude' filter (exact name exclusions)
	if len(exclude) > 0 {
		for _, script := range scripts {
			shouldExclude := false
			for _, exact := range exclude {
				// Support both with and without 'update-' prefix
				exactName := exact
				if !strings.HasPrefix(exact, "update-") {
					exactName = "update-" + exact
				}
				if script == exactName || script == exact {
					shouldExclude = true
					break
				}
			}
			if !shouldExclude {
				filtered = append(filtered, script)
			}
		}
		scripts = filtered
	}

	return scripts
}

// matchesPattern checks if a script matches a pattern (supports both exact and partial matching)
func matchesPattern(script, pattern string) bool {
	// Support patterns both with and without 'update-' prefix
	normalizedScript := script
	normalizedPattern := pattern

	// If pattern doesn't start with 'update-', also check against script without prefix
	if !strings.HasPrefix(pattern, "update-") {
		scriptWithoutPrefix := strings.TrimPrefix(script, "update-")
		if strings.Contains(scriptWithoutPrefix, pattern) || scriptWithoutPrefix == pattern {
			return true
		}
		// Also check with 'update-' prefix added to pattern
		normalizedPattern = "update-" + pattern
	}

	// Exact match or contains check
	return normalizedScript == normalizedPattern || strings.Contains(normalizedScript, normalizedPattern)
}

// isToolNotAvailable checks if the error indicates the tool is not installed
func isToolNotAvailable(output string) bool {
	notAvailablePatterns := []string{
		"not found",
		"command not found",
		"No such file",
		"skipping",
		"which: no",
		"executable file not found",
		"is not recognized as an internal or external command",
	}

	lowerOutput := strings.ToLower(output)
	for _, pattern := range notAvailablePatterns {
		if strings.Contains(lowerOutput, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// isExternallyManaged checks if the tool is managed by external package manager
func isExternallyManaged(output string) bool {
	externallyManagedPatterns := []string{
		"installed via a package manager",
		"installed through an external package manager",
		"self-update is not available",
		"cannot update",
		"managed externally",
		"externally-managed-environment",
	}

	lowerOutput := strings.ToLower(output)
	for _, pattern := range externallyManagedPatterns {
		if strings.Contains(lowerOutput, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// isAlreadyUpToDate checks if the tool is already up to date
func isAlreadyUpToDate(output string) bool {
	upToDatePatterns := []string{
		"requirement already satisfied",
		"already up to date",
		"already at the latest version",
		"no updates available",
		"nothing to update",
		"already current",
	}

	lowerOutput := strings.ToLower(output)
	for _, pattern := range upToDatePatterns {
		if strings.Contains(lowerOutput, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}
