package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/entro314-labs/update-ng/internal/updater"
)

var (
	// Color palette
	purple    = lipgloss.Color("#7D56F4")
	green     = lipgloss.Color("#04B575")
	red       = lipgloss.Color("#FF5F87")
	orange    = lipgloss.Color("#FFA500")
	blue      = lipgloss.Color("#61AFEF")
	yellow    = lipgloss.Color("#E5C07B")
	cyan      = lipgloss.Color("#56B6C2")
	gray      = lipgloss.Color("#888888")
	darkGray  = lipgloss.Color("#444444")
	lightGray = lipgloss.Color("#CCCCCC")

	// Title styles
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(purple).
			Padding(0, 2).
			Bold(true)

	// Panel styles
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purple).
			Padding(1, 2).
			MarginRight(1)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(cyan).
				Padding(1, 2).
				MarginRight(1)

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true).
			Underline(true)

	subHeaderStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	// Status styles
	successStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	runningStyle = lipgloss.NewStyle().
			Foreground(orange).
			Bold(true)

	pendingStyle = lipgloss.NewStyle().
			Foreground(gray)

	// Info styles
	infoStyle = lipgloss.NewStyle().
			Foreground(blue)

	dimStyle = lipgloss.NewStyle().
			Foreground(darkGray)

	highlightStyle = lipgloss.NewStyle().
			Foreground(yellow).
			Bold(true)

	// Output styles
	outputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(darkGray).
			Padding(0, 1).
			Foreground(lightGray)

	outputLineStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	// Stat box style
	statBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purple).
			Padding(0, 1).
			Align(lipgloss.Center).
			Width(12)

	// Badge styles
	successBadge = lipgloss.NewStyle().
			Background(green).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)

	errorBadge = lipgloss.NewStyle().
			Background(red).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	runningBadge = lipgloss.NewStyle().
			Background(orange).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)

	skippedBadge = lipgloss.NewStyle().
			Background(gray).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)
)

type ScriptItem struct {
	Name         string
	Description  string
	Status       updater.ScriptStatus
	Result       *updater.ScriptResult
	OutputLines  []string
	StartTime    time.Time
	LastActivity time.Time
}

type Model struct {
	runner    *updater.Runner
	scripts   []*ScriptItem
	scriptMap map[string]*ScriptItem

	width    int
	height   int
	spinner  spinner.Model
	progress progress.Model
	viewport viewport.Model

	doneCount    int
	successCount int
	errorCount   int
	skippedCount int
	totalCount   int
	parallel     bool
	quitting     bool
	startTime    time.Time

	showDetailedOutput bool
	focusedPanel       int // 0: stats, 1: running, 2: output
}

func NewModel(runner *updater.Runner, scripts []string, parallel bool) Model {
	// Multiple spinners for variety
	spinners := []spinner.Spinner{
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}

	s := spinner.New()
	s.Spinner = spinners[0]
	s.Style = runningStyle

	prog := progress.New(
		progress.WithGradient("#7D56F4", "#04B575"),
		progress.WithWidth(50),
		progress.WithoutPercentage(),
	)

	items := make([]*ScriptItem, len(scripts))
	sMap := make(map[string]*ScriptItem)
	descriptions := updater.GetScriptDescriptions()

	for i, name := range scripts {
		desc := descriptions[name]
		if desc == "" {
			desc = strings.TrimPrefix(name, "update-")
		}
		item := &ScriptItem{
			Name:        name,
			Description: desc,
			Status:      updater.StatusPending,
			OutputLines: []string{},
		}
		items[i] = item
		sMap[name] = item
	}

	vp := viewport.New(80, 20)

	return Model{
		runner:             runner,
		scripts:            items,
		scriptMap:          sMap,
		spinner:            s,
		progress:           prog,
		viewport:           vp,
		totalCount:         len(items),
		parallel:           parallel,
		startTime:          time.Now(),
		showDetailedOutput: true,
		focusedPanel:       1,
	}
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, m.spinner.Tick)
	cmds = append(cmds, m.tickCmd())

	if m.parallel {
		for _, item := range m.scripts {
			cmds = append(cmds, m.runScriptCmd(item.Name, item.Description))
		}
	} else {
		if len(m.scripts) > 0 {
			cmds = append(cmds, m.runScriptCmd(m.scripts[0].Name, m.scripts[0].Description))
		}
	}

	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "d":
			m.showDetailedOutput = !m.showDetailedOutput
		case "tab":
			m.focusedPanel = (m.focusedPanel + 1) % 3
		case "up", "k":
			m.viewport.ScrollUp(1)
		case "down", "j":
			m.viewport.ScrollDown(1)
		case "pgup":
			m.viewport.PageUp()
		case "pgdown":
			m.viewport.PageDown()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = (msg.Width / 2) - 6
		m.viewport.Height = max(10, msg.Height-15)
		m.progress.Width = min(50, msg.Width-30)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tickMsg:
		return m, m.tickCmd()

	case updater.ScriptStarted:
		item, ok := m.scriptMap[msg.Name]
		if ok {
			item.Status = updater.StatusRunning
			item.StartTime = time.Now()
			item.LastActivity = time.Now()
		}
		return m, m.executeScriptCmd(msg.Name, msg.Description)

	case updater.ScriptOutput:
		item, ok := m.scriptMap[msg.Name]
		if ok {
			item.OutputLines = append(item.OutputLines, msg.Line)
			item.LastActivity = time.Now()
			if len(item.OutputLines) > 100 {
				item.OutputLines = item.OutputLines[len(item.OutputLines)-100:]
			}
		}
		return m, nil

	case updater.ScriptResult:
		item, ok := m.scriptMap[msg.Name]
		if ok {
			item.Status = msg.Status
			item.Result = &msg
		}
		m.doneCount++

		switch msg.Status {
		case updater.StatusSuccess:
			m.successCount++
		case updater.StatusError:
			m.errorCount++
		case updater.StatusSkipped:
			m.skippedCount++
		}

		var cmd tea.Cmd
		if !m.parallel && m.doneCount < m.totalCount {
			for _, s := range m.scripts {
				if s.Status == updater.StatusPending {
					cmd = m.runScriptCmd(s.Name, s.Description)
					break
				}
			}
		}

		if m.doneCount >= m.totalCount {
			return m, tea.Quit
		}
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return m.renderSummary()
	}

	// Header
	header := titleStyle.Render("🚀 UPDATE COMMAND NG") + "\n"

	mode := "PARALLEL"
	if !m.parallel {
		mode = "SEQUENTIAL"
	}
	header += dimStyle.Render(fmt.Sprintf("Mode: %s", mode)) + "\n\n"

	// Stats row (horizontal boxes)
	statsRow := m.renderStatsRow()

	// Progress bar with percentage
	percent := float64(m.doneCount) / float64(m.totalCount)
	progressBar := lipgloss.NewStyle().Width(m.width - 4).Render(
		headerStyle.Render("PROGRESS") + "\n" +
			m.progress.ViewAs(percent) + " " +
			highlightStyle.Render(fmt.Sprintf("%.0f%%", percent*100)) + " " +
			dimStyle.Render(fmt.Sprintf("(%d/%d)", m.doneCount, m.totalCount)),
	)

	// Two-column layout
	leftPanel := m.renderLeftPanel()
	rightPanel := m.renderRightPanel()

	// Combine panels horizontally
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		rightPanel,
	)

	// Footer
	elapsed := time.Since(m.startTime).Round(time.Second)
	footer := dimStyle.Render(fmt.Sprintf(
		"⏱  %v | Tab: switch panel | D: toggle output | Q: quit",
		elapsed,
	))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		statsRow,
		"",
		progressBar,
		"",
		mainContent,
		"",
		footer,
	)
}

func (m Model) renderStatsRow() string {
	elapsed := time.Since(m.startTime).Round(time.Second)

	timeBox := statBoxStyle.Render(
		dimStyle.Render("TIME") + "\n" +
			highlightStyle.Render(elapsed.String()),
	)

	successBox := statBoxStyle.Copy().BorderForeground(green).Render(
		successStyle.Render("✓ SUCCESS") + "\n" +
			successStyle.Render(fmt.Sprintf("%d", m.successCount)),
	)

	errorBox := statBoxStyle.Copy().BorderForeground(red).Render(
		errorStyle.Render("✗ FAILED") + "\n" +
			errorStyle.Render(fmt.Sprintf("%d", m.errorCount)),
	)

	skippedBox := statBoxStyle.Copy().BorderForeground(gray).Render(
		dimStyle.Render("⊝ SKIPPED") + "\n" +
			dimStyle.Render(fmt.Sprintf("%d", m.skippedCount)),
	)

	pendingCount := m.totalCount - m.doneCount
	pendingBox := statBoxStyle.Copy().BorderForeground(orange).Render(
		runningStyle.Render("• PENDING") + "\n" +
			runningStyle.Render(fmt.Sprintf("%d", pendingCount)),
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		timeBox,
		" ",
		successBox,
		" ",
		errorBox,
		" ",
		skippedBox,
		" ",
		pendingBox,
	)
}

func (m Model) renderLeftPanel() string {
	style := panelStyle
	if m.focusedPanel == 1 {
		style = activePanelStyle
	}

	width := (m.width / 2) - 6
	if width < 30 {
		width = 30
	}

	var content strings.Builder
	content.WriteString(headerStyle.Render("🔄 ACTIVE TASKS") + "\n\n")

	runningScripts := m.getRunningScripts()
	if len(runningScripts) == 0 {
		content.WriteString(dimStyle.Render("No tasks currently running...") + "\n")
	} else {
		for i, item := range runningScripts {
			if i > 0 {
				content.WriteString("\n")
			}
			content.WriteString(m.renderRunningScriptCard(item, width-4))
		}
	}

	// Add completed scripts below
	content.WriteString("\n\n" + headerStyle.Render("📋 COMPLETED") + "\n\n")
	completed := m.getCompletedScripts()
	if len(completed) == 0 {
		content.WriteString(dimStyle.Render("No completed tasks yet...") + "\n")
	} else {
		maxShow := 5
		for i, item := range completed {
			if i >= maxShow {
				remaining := len(completed) - maxShow
				content.WriteString(dimStyle.Render(fmt.Sprintf("...and %d more", remaining)) + "\n")
				break
			}
			content.WriteString(m.renderCompactScript(item) + "\n")
		}
	}

	return style.Width(width).Height(m.height - 12).Render(content.String())
}

func (m Model) renderRightPanel() string {
	style := panelStyle
	if m.focusedPanel == 2 {
		style = activePanelStyle
	}

	width := (m.width / 2) - 6
	if width < 30 {
		width = 30
	}

	var content strings.Builder
	content.WriteString(headerStyle.Render("📝 LIVE OUTPUT") + "\n\n")

	runningScripts := m.getRunningScripts()
	if len(runningScripts) == 0 {
		content.WriteString(dimStyle.Render("No output to display...") + "\n")
	} else {
		for i, item := range runningScripts {
			if i > 0 {
				content.WriteString("\n" + strings.Repeat("─", width-4) + "\n\n")
			}

			name := strings.TrimPrefix(item.Name, "update-")
			content.WriteString(
				runningStyle.Render("▶ "+strings.ToUpper(name)) + " " +
					dimStyle.Render(item.Description) + "\n\n",
			)

			if len(item.OutputLines) == 0 {
				content.WriteString(dimStyle.Render("  Initializing...") + "\n")
			} else {
				start := 0
				maxLines := 10
				if len(item.OutputLines) > maxLines {
					start = len(item.OutputLines) - maxLines
				}

				for _, line := range item.OutputLines[start:] {
					trimmed := strings.TrimSpace(line)
					if trimmed != "" {
						if len(trimmed) > width-8 {
							trimmed = trimmed[:width-11] + "..."
						}
						content.WriteString(outputLineStyle.Render("  "+trimmed) + "\n")
					}
				}
			}
		}
	}

	m.viewport.SetContent(content.String())
	m.viewport.Width = width - 4
	m.viewport.Height = m.height - 15

	return style.Width(width).Height(m.height - 12).Render(m.viewport.View())
}

func (m Model) renderRunningScriptCard(item *ScriptItem, width int) string {
	name := strings.TrimPrefix(item.Name, "update-")
	elapsed := time.Since(item.StartTime).Round(time.Second)

	// Card header
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.spinner.View()+" ",
		runningBadge.Render("RUNNING"),
		" ",
		highlightStyle.Render(strings.ToUpper(name)),
	)

	// Description
	desc := dimStyle.Render(item.Description)

	// Last output preview
	preview := ""
	if len(item.OutputLines) > 0 {
		lastLine := item.OutputLines[len(item.OutputLines)-1]
		trimmed := strings.TrimSpace(lastLine)
		if trimmed != "" {
			if len(trimmed) > width-6 {
				trimmed = trimmed[:width-9] + "..."
			}
			preview = "\n" + infoStyle.Render("→ "+trimmed)
		}
	}

	// Timer
	timer := dimStyle.Render(fmt.Sprintf("⏱  %v", elapsed))

	card := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		desc,
		preview,
		"",
		timer,
	)

	return outputBoxStyle.Width(width).Render(card)
}

func (m Model) renderCompactScript(item *ScriptItem) string {
	name := strings.TrimPrefix(item.Name, "update-")

	var badge string
	switch item.Status {
	case updater.StatusSuccess:
		badge = successBadge.Render(" ✓ ")
	case updater.StatusError:
		badge = errorBadge.Render(" ✗ ")
	case updater.StatusSkipped:
		badge = skippedBadge.Render(" ⊝ ")
	default:
		badge = dimStyle.Render(" • ")
	}

	timing := ""
	if item.Result != nil && item.Result.Duration > 0 {
		timing = dimStyle.Render(fmt.Sprintf(" (%v)", item.Result.Duration.Round(10*time.Millisecond)))
	}

	return badge + " " + name + timing
}

func (m Model) renderSummary() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(titleStyle.Render("✨ UPDATE COMPLETE") + "\n\n")

	totalTime := time.Since(m.startTime).Round(time.Second)

	// Summary boxes
	timeBox := statBoxStyle.Copy().Width(20).Render(
		headerStyle.Render("TOTAL TIME") + "\n" +
			highlightStyle.Render(totalTime.String()),
	)

	scriptsBox := statBoxStyle.Copy().Width(20).Render(
		headerStyle.Render("SCRIPTS RUN") + "\n" +
			infoStyle.Render(fmt.Sprintf("%d", m.totalCount)),
	)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, timeBox, "  ", scriptsBox) + "\n\n")

	// Results
	b.WriteString(headerStyle.Render("RESULTS") + "\n\n")
	b.WriteString(fmt.Sprintf("  %s %d successful\n", successStyle.Render("✓"), m.successCount))
	if m.errorCount > 0 {
		b.WriteString(fmt.Sprintf("  %s %d failed\n", errorStyle.Render("✗"), m.errorCount))
	}
	if m.skippedCount > 0 {
		b.WriteString(fmt.Sprintf("  %s %d skipped\n", dimStyle.Render("⊝"), m.skippedCount))
	}

	b.WriteString("\n")
	return b.String()
}

func (m Model) getRunningScripts() []*ScriptItem {
	var running []*ScriptItem
	for _, item := range m.scripts {
		if item.Status == updater.StatusRunning {
			running = append(running, item)
		}
	}
	return running
}

func (m Model) getCompletedScripts() []*ScriptItem {
	var completed []*ScriptItem
	for _, item := range m.scripts {
		if item.Status != updater.StatusRunning && item.Status != updater.StatusPending {
			completed = append(completed, item)
		}
	}
	return completed
}

func (m *Model) runScriptCmd(name, description string) tea.Cmd {
	return func() tea.Msg {
		return updater.ScriptStarted{
			Name:        name,
			Description: description,
		}
	}
}

func (m *Model) executeScriptCmd(name, description string) tea.Cmd {
	return func() tea.Msg {
		return m.runner.RunScript(name, description)
	}
}

type tickMsg time.Time

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
