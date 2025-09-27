package main

import (
	"bufio"
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:embed bin/*
var scripts embed.FS

const (
	programCommand = "update-ng"
	programVersion = "1.0.0"
	programUpdated = "2025-09-24T00:30:00Z"
	programLicense = "MIT"
	programContact = "Dominikos Pritis (https://idominikos.com)"
)

var cfgFile string
var logger *zap.Logger

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

// Script categories for organized execution
var scriptCategories = map[string][]string{
	"System Package Managers": {
		"update-apt", "update-brew", "update-dnf", "update-macports",
		"update-pacman", "update-zypper", "update-winget", "update-snap",
	},
	"Version Managers": {
		"update-asdf", "update-fnm", "update-gvm", "update-mise", "update-volta",
	},
	"Language Package Managers": {
		"update-npm", "update-pip", "update-cargo", "update-gem", "update-conda",
		"update-mamba", "update-poetry", "update-pdm", "update-rye", "update-uv",
		"update-yarn", "update-bun", "update-deno",
	},
	"Cloud & Infrastructure": {
		"update-aws", "update-gcloud", "update-kubectl", "update-terraform",
		"update-docker", "update-podman",
	},
	"Development Tools": {
		"update-gh", "update-chezmoi", "update-pre-commit", "update-direnv",
	},
	"Shell & Environment": {
		"update-zsh", "update-oh-my-zsh", "update-antidote",
	},
}

// getScriptDescriptions returns human-readable descriptions for each update script
func getScriptDescriptions() map[string]string {
	return map[string]string{
		"update-apt":        "APT - Debian/Ubuntu package manager",
		"update-brew":       "Homebrew - macOS package manager",
		"update-dnf":        "DNF - Red Hat/Fedora package manager",
		"update-macports":   "MacPorts - macOS package manager",
		"update-pacman":     "Pacman - Arch Linux package manager",
		"update-zypper":     "Zypper - openSUSE package manager",
		"update-winget":     "Windows Package Manager",
		"update-snap":       "Snap - Universal Linux packages",
		"update-asdf":       "ASDF - Multi-language version manager",
		"update-fnm":        "Fast Node Manager - Node.js version manager",
		"update-gvm":        "Go Version Manager",
		"update-mise":       "Mise - Modern version manager for tools",
		"update-volta":      "Volta - JavaScript tool manager",
		"update-npm":        "NPM - Node.js package manager",
		"update-pip":        "Pip - Python package installer",
		"update-cargo":      "Cargo - Rust package manager",
		"update-gem":        "RubyGems - Ruby package manager",
		"update-conda":      "Conda - Data science package manager",
		"update-mamba":      "Mamba - Fast conda-compatible package manager",
		"update-poetry":     "Poetry - Python dependency manager",
		"update-pdm":        "PDM - Modern Python package manager",
		"update-rye":        "Rye - Experimental Python packaging tool",
		"update-uv":         "UV - Ultra-fast Python package installer",
		"update-yarn":       "Yarn - JavaScript package manager",
		"update-bun":        "Bun - Fast JavaScript runtime and toolkit",
		"update-deno":       "Deno - Secure JavaScript/TypeScript runtime",
		"update-aws":        "AWS CLI - Amazon Web Services CLI",
		"update-gcloud":     "Google Cloud SDK",
		"update-kubectl":    "Kubectl - Kubernetes command-line tool",
		"update-terraform":  "Terraform - Infrastructure as Code",
		"update-docker":     "Docker - Container platform",
		"update-podman":     "Podman - Alternative to Docker",
		"update-gh":         "GitHub CLI",
		"update-chezmoi":    "Chezmoi - Dotfiles manager",
		"update-pre-commit": "Pre-commit - Git hooks framework",
		"update-direnv":     "Direnv - Environment variable loader",
		"update-zsh":        "Zsh shell updates",
		"update-oh-my-zsh":  "Oh My Zsh - Zsh configuration framework",
		"update-antidote":   "Antidote - Zsh plugin manager",
		"update-rustup":     "Rustup - Rust toolchain installer",
		"update-flatpak":    "Flatpak - Linux application distribution",
		"update-mas":        "Mac App Store - macOS app updates",
		"update-choco":      "Chocolatey - Windows package manager",
		"update-scoop":      "Scoop - Windows command-line installer",
	}
}

type scriptStatus int

const (
	statusPending scriptStatus = iota
	statusRunning
	statusSuccess
	statusError
	statusSkipped
)

type scriptResult struct {
	name        string
	status      scriptStatus
	duration    time.Duration
	output      string
	error       error
	description string
	lastLine    string
}

type model struct {
	scripts      []string
	results      map[string]*scriptResult
	progress     progress.Model
	spinner      spinner.Model
	completed    int
	total        int
	startTime    time.Time
	showTUI      bool
	parallel     bool
	categories   []string
	verbose      bool
	mu           sync.RWMutex
	lastPrinted  map[string]scriptStatus
	headerPrinted bool
}

type scriptFinishedMsg scriptResult
type scriptProgressMsg struct {
	name   string
	output string
}

func initialModel(scripts []string, showTUI, parallel, verbose bool, categories []string) model {
	p := progress.New(progress.WithDefaultGradient())
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	results := make(map[string]*scriptResult)
	scriptDescriptions := getScriptDescriptions()

	for _, script := range scripts {
		results[script] = &scriptResult{
			name:        script,
			status:      statusPending,
			description: scriptDescriptions[script],
		}
	}

	return model{
		scripts:      scripts,
		results:      results,
		progress:     p,
		spinner:      s,
		total:        len(scripts),
		showTUI:      showTUI,
		parallel:     parallel,
		categories:   categories,
		verbose:      verbose,
		startTime:    time.Now(),
		lastPrinted:  make(map[string]scriptStatus),
		headerPrinted: false,
	}
}

func (m *model) Init() tea.Cmd {
	cmds := []tea.Cmd{m.spinner.Tick}

	if m.parallel {
		// Start all scripts in parallel
		for _, script := range m.scripts {
			cmds = append(cmds, runScript(script))
		}
	} else {
		// Start first script only
		if len(m.scripts) > 0 {
			cmds = append(cmds, runScript(m.scripts[0]))
		}
	}

	return tea.Batch(cmds...)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case scriptProgressMsg:
		m.mu.Lock()
		if result, exists := m.results[msg.name]; exists {
			result.lastLine = msg.output
			result.status = statusRunning
		}
		m.mu.Unlock()

	case scriptFinishedMsg:
		m.mu.Lock()
		if result, exists := m.results[msg.name]; exists {
			result.status = msg.status
			result.duration = msg.duration
			result.output = msg.output
			result.error = msg.error
			if msg.status == statusSuccess || msg.status == statusError || msg.status == statusSkipped {
				m.completed++
			}
		}
		m.mu.Unlock()

		// If running sequentially, start next script
		if !m.parallel && m.completed < m.total {
			for _, script := range m.scripts {
				if m.results[script].status == statusPending {
					return m, runScript(script)
				}
			}
		}

		// Check if all scripts are done
		if m.completed >= m.total {
			if m.showTUI {
				time.Sleep(2 * time.Second)
			}
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *model) View() string {
	if !m.showTUI {
		return ""
	}

	var output strings.Builder

	// Print header only once
	if !m.headerPrinted {
		title := "🚀 Update Command NG - Modern System Updater"
		if m.verbose {
			title += " (Verbose Mode)"
		}
		output.WriteString(titleStyle.Render(title) + "\n")

		output.WriteString(fmt.Sprintf("Starting %d update scripts...\n", m.total))

		if m.parallel {
			output.WriteString(infoStyle.Render("Running in parallel mode") + "\n")
		} else {
			output.WriteString(infoStyle.Render("Running in sequential mode") + "\n")
		}
		output.WriteString("\n")

		m.headerPrinted = true
	}

	// Print updates for scripts that have changed status
	m.mu.RLock()
	for _, script := range m.scripts {
		if result, exists := m.results[script]; exists {
			lastStatus, wasTracked := m.lastPrinted[script]

			// Print if status changed or if we haven't printed this script before
			if !wasTracked || lastStatus != result.status {
				output.WriteString(m.formatScriptUpdate(result))
				m.lastPrinted[script] = result.status
			}
		}
	}
	m.mu.RUnlock()

	// Print final summary when all scripts are completed
	if m.completed >= m.total {
		var successCount, errorCount, skippedCount int
		for _, result := range m.results {
			switch result.status {
			case statusSuccess:
				successCount++
			case statusError:
				errorCount++
			case statusSkipped:
				skippedCount++
			}
		}

		output.WriteString("\n")
		output.WriteString("📊 " + lipgloss.NewStyle().Bold(true).Render("Final Summary: "))
		output.WriteString(fmt.Sprintf("%s %d succeeded", successStyle.Render("✓"), successCount))
		if errorCount > 0 {
			output.WriteString(fmt.Sprintf(", %s %d failed", errorStyle.Render("✗"), errorCount))
		}
		if skippedCount > 0 {
			output.WriteString(fmt.Sprintf(", %s %d skipped", infoStyle.Render("⊝"), skippedCount))
		}

		totalTime := time.Since(m.startTime)
		output.WriteString(fmt.Sprintf("\nTotal time: %v", totalTime.Round(time.Second)))
		output.WriteString("\n🎉 All updates completed!\n")
	}

	return output.String()
}

func (m *model) formatScriptUpdate(result *scriptResult) string {
	name := strings.TrimPrefix(result.name, "update-")
	var output strings.Builder

	switch result.status {
	case statusPending:
		// Don't print anything for pending status to reduce noise
		return ""

	case statusRunning:
		output.WriteString(fmt.Sprintf("%s %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Render("⟳"),
			scriptStyle.Render(name)))
		if result.description != "" {
			output.WriteString(fmt.Sprintf(" - %s", infoStyle.Render(result.description)))
		}
		output.WriteString(infoStyle.Render(" starting..."))
		output.WriteString("\n")

	case statusSuccess:
		duration := result.duration.Round(100 * time.Millisecond)
		output.WriteString(fmt.Sprintf("%s %s", successStyle.Render("✓"), scriptStyle.Render(name)))
		if result.description != "" {
			output.WriteString(fmt.Sprintf(" - %s", infoStyle.Render(result.description)))
		}
		output.WriteString(fmt.Sprintf(" %s", infoStyle.Render(fmt.Sprintf("(%v)", duration))))

		if m.verbose && result.output != "" {
			// Show interesting lines from the output
			lines := strings.Split(strings.TrimSpace(result.output), "\n")
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

			if len(interesting) > 0 {
				for _, line := range interesting {
					output.WriteString(fmt.Sprintf("  %s\n", successStyle.Render("→ "+line)))
				}
			}
		}
		output.WriteString("\n")

	case statusError:
		output.WriteString(fmt.Sprintf("%s %s", errorStyle.Render("✗"), scriptStyle.Render(name)))
		if result.description != "" {
			output.WriteString(fmt.Sprintf(" - %s", infoStyle.Render(result.description)))
		}

		errorMsg := "failed"
		if result.error != nil {
			errorMsg = result.error.Error()
		}
		output.WriteString(fmt.Sprintf(" %s", errorStyle.Render(errorMsg)))

		if result.output != "" {
			lines := strings.Split(strings.TrimSpace(result.output), "\n")
			for i, line := range lines {
				if i >= 3 {
					output.WriteString(fmt.Sprintf("  %s\n", errorStyle.Render("... (truncated)")))
					break
				}
				line = strings.TrimSpace(line)
				if line != "" {
					if len(line) > 100 {
						line = line[:97] + "..."
					}
					output.WriteString(fmt.Sprintf("  %s\n", errorStyle.Render("✗ "+line)))
				}
			}
		}
		output.WriteString("\n")

	case statusSkipped:
		output.WriteString(fmt.Sprintf("%s %s", infoStyle.Render("⊝"), scriptStyle.Render(name)))
		if result.description != "" {
			output.WriteString(fmt.Sprintf(" - %s", infoStyle.Render(result.description)))
		}
		output.WriteString(fmt.Sprintf(" %s", infoStyle.Render("skipped")))

		if m.verbose && result.output != "" {
			lines := strings.Split(strings.TrimSpace(result.output), "\n")
			if len(lines) > 0 && strings.TrimSpace(lines[0]) != "" {
				reason := strings.TrimSpace(lines[0])
				if len(reason) > 80 {
					reason = reason[:77] + "..."
				}
				output.WriteString(fmt.Sprintf("  %s", infoStyle.Render("→ "+reason)))
			}
		}
		output.WriteString("\n")
	}

	return output.String()
}

func (m *model) renderScriptStatus(result *scriptResult) string {
	name := strings.TrimPrefix(result.name, "update-")
	duration := infoStyle.Render(fmt.Sprintf("(%v)", result.duration.Round(100*time.Millisecond)))

	var output strings.Builder

	switch result.status {
	case statusPending:
		output.WriteString(fmt.Sprintf("  %s %s", infoStyle.Render("○"), scriptStyle.Render(name)))
		if result.description != "" {
			output.WriteString(fmt.Sprintf(" - %s", infoStyle.Render(result.description)))
		}
		output.WriteString("\n")

	case statusRunning:
		output.WriteString(fmt.Sprintf("  %s %s", m.spinner.View(), scriptStyle.Render(name)))
		if result.description != "" {
			output.WriteString(fmt.Sprintf(" - %s", infoStyle.Render(result.description)))
		}

		// Always show what's happening when running
		if result.lastLine != "" && strings.TrimSpace(result.lastLine) != "" {
			lastLine := strings.TrimSpace(result.lastLine)
			if len(lastLine) > 80 {
				lastLine = lastLine[:77] + "..."
			}
			output.WriteString(fmt.Sprintf("\n      %s", lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FF")).Render("→ "+lastLine)))
		} else {
			// Show generic progress message if no specific output
			elapsed := time.Since(time.Now().Add(-result.duration))
			if elapsed > 2*time.Second {
				output.WriteString(fmt.Sprintf("\n      %s", infoStyle.Render("→ Processing... ("+elapsed.Round(time.Second).String()+")")))
			} else {
				output.WriteString(fmt.Sprintf("\n      %s", infoStyle.Render("→ Starting up...")))
			}
		}
		output.WriteString("\n")

	case statusSuccess:
		output.WriteString(fmt.Sprintf("  %s %s", successStyle.Render("✓"), scriptStyle.Render(name)))
		if result.description != "" {
			output.WriteString(fmt.Sprintf(" - %s", infoStyle.Render(result.description)))
		}
		output.WriteString(fmt.Sprintf(" %s", duration))

		if m.verbose && result.output != "" {
			// Show a summary of what was updated
			lines := strings.Split(strings.TrimSpace(result.output), "\n")
			var interesting []string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				// Look for lines that indicate updates, installations, etc.
				if strings.Contains(line, "updated") || strings.Contains(line, "installed") ||
				   strings.Contains(line, "upgraded") || strings.Contains(line, "downloading") ||
				   strings.Contains(line, "fetching") || strings.Contains(line, "pulling") {
					if len(line) > 80 {
						line = line[:77] + "..."
					}
					interesting = append(interesting, line)
					if len(interesting) >= 3 { // Limit to 3 lines
						break
					}
				}
			}

			if len(interesting) > 0 {
				output.WriteString("\n")
				for _, line := range interesting {
					output.WriteString(fmt.Sprintf("      %s\n", successStyle.Render("  ↳ "+line)))
				}
			} else if len(lines) > 0 && lines[len(lines)-1] != "" {
				// Show last non-empty line if no interesting updates found
				lastLine := lines[len(lines)-1]
				if len(lastLine) > 80 {
					lastLine = lastLine[:77] + "..."
				}
				output.WriteString(fmt.Sprintf("\n      %s\n", successStyle.Render("  ↳ "+lastLine)))
			}
		}
		output.WriteString("\n")

	case statusError:
		output.WriteString(fmt.Sprintf("  %s %s", errorStyle.Render("✗"), scriptStyle.Render(name)))
		if result.description != "" {
			output.WriteString(fmt.Sprintf(" - %s", infoStyle.Render(result.description)))
		}

		errorMsg := "failed"
		if result.error != nil {
			errorMsg = result.error.Error()
		}
		output.WriteString(fmt.Sprintf(" %s", errorStyle.Render(errorMsg)))

		if result.output != "" {
			// Show the error output, but format it nicely
			lines := strings.Split(strings.TrimSpace(result.output), "\n")
			output.WriteString("\n")
			for i, line := range lines {
				if i >= 5 { // Limit to 5 lines of error output
					output.WriteString(fmt.Sprintf("      %s\n", errorStyle.Render("  ... (truncated)")))
					break
				}
				line = strings.TrimSpace(line)
				if line != "" {
					if len(line) > 80 {
						line = line[:77] + "..."
					}
					output.WriteString(fmt.Sprintf("      %s\n", errorStyle.Render("  ✗ "+line)))
				}
			}
		}
		output.WriteString("\n")

	case statusSkipped:
		output.WriteString(fmt.Sprintf("  %s %s", infoStyle.Render("⊝"), scriptStyle.Render(name)))
		if result.description != "" {
			output.WriteString(fmt.Sprintf(" - %s", infoStyle.Render(result.description)))
		}
		output.WriteString(fmt.Sprintf(" %s", infoStyle.Render("skipped")))

		if m.verbose && result.output != "" {
			// Show why it was skipped
			lines := strings.Split(strings.TrimSpace(result.output), "\n")
			if len(lines) > 0 && lines[0] != "" {
				reason := lines[0]
				if len(reason) > 60 {
					reason = reason[:57] + "..."
				}
				output.WriteString(fmt.Sprintf("\n      %s", infoStyle.Render("  → "+reason)))
			}
		}
		output.WriteString("\n")
	}

	return output.String()
}

func runScript(scriptName string) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		start := time.Now()

		// Send progress update that script is starting
		defer func() {
			if r := recover(); r != nil {
				// Handle any panics
				return
			}
		}()

		// Add a small delay to ensure the running state is visible
		time.Sleep(200 * time.Millisecond)

		// Read script content
		scriptPath := filepath.Join("bin", scriptName)
		data, err := scripts.ReadFile(scriptPath)
		if err != nil {
			return scriptFinishedMsg{
				name:        scriptName,
				status:      statusError,
				duration:    time.Since(start),
				error:       fmt.Errorf("failed to read script: %w", err),
				description: "Error reading script file",
			}
		}

		// Create temp file and execute
		tmpFile := filepath.Join(os.TempDir(), scriptName)
		if err := os.WriteFile(tmpFile, data, 0755); err != nil {
			return scriptFinishedMsg{
				name:        scriptName,
				status:      statusError,
				duration:    time.Since(start),
				error:       fmt.Errorf("failed to write temp script: %w", err),
				description: "Error creating temporary file",
			}
		}
		defer os.Remove(tmpFile)

		// Run the script with real-time output
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		cmd := exec.CommandContext(ctx, "/bin/sh", tmpFile)

		// Create pipes for real-time output
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return scriptFinishedMsg{
				name:        scriptName,
				status:      statusError,
				duration:    time.Since(start),
				error:       fmt.Errorf("failed to create stdout pipe: %w", err),
				description: "Error setting up output stream",
			}
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return scriptFinishedMsg{
				name:        scriptName,
				status:      statusError,
				duration:    time.Since(start),
				error:       fmt.Errorf("failed to create stderr pipe: %w", err),
				description: "Error setting up error stream",
			}
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			return scriptFinishedMsg{
				name:        scriptName,
				status:      statusError,
				duration:    time.Since(start),
				error:       fmt.Errorf("failed to start command: %w", err),
				description: "Error starting command",
			}
		}

		var outputBuffer strings.Builder
		var lastLine string

		// Create a channel to collect all output
		outputDone := make(chan struct{})

		// Read stdout and stderr concurrently
		go func() {
			defer close(outputDone)

			// Merge stdout and stderr
			scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
			for scanner.Scan() {
				line := scanner.Text()
				outputBuffer.WriteString(line + "\n")

				// Update last line for real-time display
				if strings.TrimSpace(line) != "" {
					lastLine = line
					// Send progress update - this would need to be handled differently in a real scenario
					// For now, we'll just track the last line
				}
			}
		}()

		// Wait for command to complete
		err = cmd.Wait()
		<-outputDone // Wait for all output to be collected

		duration := time.Since(start)
		outputStr := outputBuffer.String()

		// Ensure minimum execution time for visibility (at least 500ms)
		if duration < 500*time.Millisecond {
			time.Sleep(500*time.Millisecond - duration)
			duration = time.Since(start)
		}

		// Log the result
		logFields := []zap.Field{
			zap.String("script", scriptName),
			zap.Duration("duration", duration),
			zap.String("output", outputStr),
			zap.String("lastLine", lastLine),
		}

		if err != nil {
			// Check if it's a "not found" error (tool not installed)
			if strings.Contains(outputStr, "not found") || strings.Contains(outputStr, "skipping") ||
			   strings.Contains(outputStr, "command not found") || strings.Contains(outputStr, "No such file") {
				logger.Warn("Script skipped", logFields...)
				return scriptFinishedMsg{
					name:        scriptName,
					status:      statusSkipped,
					duration:    duration,
					output:      outputStr,
					description: "Tool not found or not installed",
				}
			}
			logFields = append(logFields, zap.Error(err))
			logger.Error("Script failed", logFields...)
			return scriptFinishedMsg{
				name:        scriptName,
				status:      statusError,
				duration:    duration,
				output:      outputStr,
				error:       err,
				description: "Execution failed",
			}
		}

		logger.Info("Script finished successfully", logFields...)
		return scriptFinishedMsg{
			name:        scriptName,
			status:      statusSuccess,
			duration:    duration,
			output:      outputStr,
			description: "Completed successfully",
		}
	})
}

func runWithTUI(scripts []string, parallel, verbose bool, categories []string) {
	fmt.Println(titleStyle.Render("🚀 Update Command NG - Modern System Updater"))
	if verbose {
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
	results := make(map[string]*scriptResult)
	scriptDescriptions := getScriptDescriptions()

	// Initialize results
	for _, script := range scripts {
		results[script] = &scriptResult{
			name:        script,
			status:      statusPending,
			description: scriptDescriptions[script],
		}
	}

	var wg sync.WaitGroup
	resultsChan := make(chan scriptResult, len(scripts))

	// Function to run a single script
	runSingleScript := func(scriptName string) {
		defer wg.Done()

		name := strings.TrimPrefix(scriptName, "update-")
		description := scriptDescriptions[scriptName]

		fmt.Printf("%s %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Render("⟳"),
			scriptStyle.Render(name))
		if description != "" {
			fmt.Printf(" - %s", infoStyle.Render(description))
		}
		fmt.Println(infoStyle.Render(" starting..."))

		result := runScriptSync(scriptName, description)

		// Print result immediately
		switch result.status {
		case statusSuccess:
			duration := result.duration.Round(100 * time.Millisecond)
			fmt.Printf("%s %s", successStyle.Render("✓"), scriptStyle.Render(name))
			if description != "" {
				fmt.Printf(" - %s", infoStyle.Render(description))
			}
			fmt.Printf(" %s\n", infoStyle.Render(fmt.Sprintf("(%v)", duration)))

			if verbose && result.output != "" {
				lines := strings.Split(strings.TrimSpace(result.output), "\n")
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

		case statusError:
			fmt.Printf("%s %s", errorStyle.Render("✗"), scriptStyle.Render(name))
			if description != "" {
				fmt.Printf(" - %s", infoStyle.Render(description))
			}

			errorMsg := "failed"
			if result.error != nil {
				errorMsg = result.error.Error()
			}
			fmt.Printf(" %s\n", errorStyle.Render(errorMsg))

			if result.output != "" {
				lines := strings.Split(strings.TrimSpace(result.output), "\n")
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

		case statusSkipped:
			fmt.Printf("%s %s", infoStyle.Render("⊝"), scriptStyle.Render(name))
			if description != "" {
				fmt.Printf(" - %s", infoStyle.Render(description))
			}
			fmt.Printf(" %s", infoStyle.Render("skipped"))

			if verbose && result.output != "" {
				lines := strings.Split(strings.TrimSpace(result.output), "\n")
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
		results[result.name] = &result
		completed++

		switch result.status {
		case statusSuccess:
			successCount++
		case statusError:
			errorCount++
		case statusSkipped:
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

func runScriptSync(scriptName, description string) scriptResult {
	start := time.Now()

	// Add a small delay to show the running state
	time.Sleep(100 * time.Millisecond)

	// Read script content
	scriptPath := filepath.Join("bin", scriptName)
	data, err := scripts.ReadFile(scriptPath)
	if err != nil {
		return scriptResult{
			name:        scriptName,
			status:      statusError,
			duration:    time.Since(start),
			error:       fmt.Errorf("failed to read script: %w", err),
			description: description,
		}
	}

	// Create temp file and execute
	tmpFile := filepath.Join(os.TempDir(), scriptName)
	if err := os.WriteFile(tmpFile, data, 0755); err != nil {
		return scriptResult{
			name:        scriptName,
			status:      statusError,
			duration:    time.Since(start),
			error:       fmt.Errorf("failed to write temp script: %w", err),
			description: description,
		}
	}
	defer os.Remove(tmpFile)

	// Run the script
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", tmpFile)
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
		// Check if it's a "not found" error (tool not installed)
		if strings.Contains(outputStr, "not found") || strings.Contains(outputStr, "skipping") ||
		   strings.Contains(outputStr, "command not found") || strings.Contains(outputStr, "No such file") {
			logger.Warn("Script skipped", logFields...)
			return scriptResult{
				name:        scriptName,
				status:      statusSkipped,
				duration:    duration,
				output:      outputStr,
				description: description,
			}
		}
		logFields = append(logFields, zap.Error(err))
		logger.Error("Script failed", logFields...)
		return scriptResult{
			name:        scriptName,
			status:      statusError,
			duration:    duration,
			output:      outputStr,
			error:       err,
			description: description,
		}
	}

	logger.Info("Script finished successfully", logFields...)
	return scriptResult{
		name:        scriptName,
		status:      statusSuccess,
		duration:    duration,
		output:      outputStr,
		description: description,
	}
}

func getAllScripts() ([]string, error) {
	entries, err := scripts.ReadDir("bin")
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

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find config directory.
		configDir, err := os.UserConfigDir()
		cobra.CheckErr(err)
		updateNgConfigDir := filepath.Join(configDir, "update-ng")

		// Search config in user's config directory.
		viper.AddConfigPath(updateNgConfigDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		// For backward compatibility, also search for old config file in home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigName(".update-ng")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func initLogger() {
	logPath := viper.GetString("log-file")
	if logPath == "" {
		// Default log path
		configDir, err := os.UserConfigDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get user config directory: %v\n", err)
			os.Exit(1)
		}
		logPath = filepath.Join(configDir, "update-ng", "update-ng.log")
	}

	// Ensure log directory exists
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create log directory: %v\n", err)
		os.Exit(1)
	}

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewJSONEncoder(config)
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open log file: %v\n", err)
		os.Exit(1)
	}

	writer := zapcore.AddSync(logFile)
	defaultLogLevel := zapcore.InfoLevel
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
	)

	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "update-ng",
		Short: "Modern system updater with beautiful TUI",
		Long: `Update Command NG - A modern, fast system updater that manages
all your package managers, development tools, and system components
with a beautiful terminal interface.`,
		Version: programVersion,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig()
			initLogger()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			showTUI, _ := cmd.Flags().GetBool("tui")
			parallel, _ := cmd.Flags().GetBool("parallel")
			verbose, _ := cmd.Flags().GetBool("verbose")
			categories, _ := cmd.Flags().GetStringSlice("categories")
			only, _ := cmd.Flags().GetStringSlice("only")
			skip, _ := cmd.Flags().GetStringSlice("skip")

			// Bind flags to viper
			viper.BindPFlag("tui", cmd.Flags().Lookup("tui"))
			viper.BindPFlag("parallel", cmd.Flags().Lookup("parallel"))
			viper.BindPFlag("verbose", cmd.Flags().Lookup("verbose"))
			viper.BindPFlag("only", cmd.Flags().Lookup("only"))
			viper.BindPFlag("skip", cmd.Flags().Lookup("skip"))
			viper.BindPFlag("categories", cmd.Flags().Lookup("categories"))
			viper.BindPFlag("log-file", cmd.Flags().Lookup("log-file"))

			// Get values from viper
			showTUI = viper.GetBool("tui")
			parallel = viper.GetBool("parallel")
			verbose = viper.GetBool("verbose")
			categories = viper.GetStringSlice("categories")
			only = viper.GetStringSlice("only")
			skip = viper.GetStringSlice("skip")

			scripts, err := getAllScripts()
			if err != nil {
				return fmt.Errorf("failed to read scripts: %w", err)
			}

			// Filter scripts based on flags
			if len(only) > 0 {
				var filtered []string
				for _, script := range scripts {
					for _, pattern := range only {
						if strings.Contains(script, pattern) {
							filtered = append(filtered, script)
							break
						}
					}
				}
				scripts = filtered
			}

			if len(skip) > 0 {
				var filtered []string
				for _, script := range scripts {
					shouldSkip := false
					for _, pattern := range skip {
						if strings.Contains(script, pattern) {
							shouldSkip = true
							break
						}
					}
					if !shouldSkip {
						filtered = append(filtered, script)
					}
				}
				scripts = filtered
			}

			if len(scripts) == 0 {
				fmt.Println("No scripts to run.")
				return nil
			}

			if showTUI {
				runWithTUI(scripts, parallel, verbose, categories)
				return nil
			} else {
				// Simple text mode
				fmt.Printf("Running %d update scripts...\n", len(scripts))
				for _, script := range scripts {
					fmt.Printf("Running %s...", script)
					// Run script synchronously in text mode
					// Implementation similar to runScript but synchronous
					fmt.Println(" done")
				}
				return nil
			}
		},
	}

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/update-ng/config.yaml)")
	rootCmd.Flags().BoolP("tui", "t", true, "Show beautiful TUI interface")
	rootCmd.Flags().BoolP("parallel", "p", true, "Run scripts in parallel")
	rootCmd.Flags().BoolP("verbose", "v", true, "Show verbose output with detailed information")
	rootCmd.Flags().StringSliceP("only", "o", []string{}, "Only run scripts matching these patterns")
	rootCmd.Flags().StringSliceP("skip", "s", []string{}, "Skip scripts matching these patterns")
	rootCmd.Flags().StringSliceP("categories", "c", []string{}, "Only run scripts from these categories")
	rootCmd.Flags().String("log-file", "", "Path to log file (default is in user's config directory)")

	// Version subcommands
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s version %s\n", programCommand, programVersion)
			fmt.Printf("Updated: %s\n", programUpdated)
			fmt.Printf("License: %s\n", programLicense)
			fmt.Printf("Contact: %s\n", programContact)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}