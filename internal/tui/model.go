package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/paperalt/goclean/internal/cleaner"
)

// Styles
var (
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("63"))

	redStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	greenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	orangeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))

	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)

	// Layout
	checkMark = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).SetString("✓")
	crossMark = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).SetString("✗")

	// Container - Dynamic Border
	baseDocStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63"))

	// Modal Style
	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Padding(1, 4).
			Align(lipgloss.Center)
)

type state int

const (
	stateScanning state = iota
	stateReview
	stateLargeFileSelection
	stateConfirm // New Confirmation State
	stateCleaning
	stateDone
)

type item struct {
	cleaner        cleaner.Cleaner
	selected       bool
	size           int64
	scanned        bool
	cleaned        bool
	err            error
	skip           bool
	statusOverride string
}

type largeFileItem struct {
	path     string
	size     int64
	selected bool
}

type model struct {
	state  state
	items  []*item
	cursor int

	// Large File Sub-menu
	largeFiles      []*largeFileItem
	lfCursor        int
	lfCleanerIndex  int
	lfSelectedCount int
	lfSelectedSize  int64

	spinner   spinner.Model
	totalSize int64
	width     int
	height    int
	quitting  bool
	isRoot    bool
}

func InitialModel(cleaners []cleaner.Cleaner) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	isRoot := os.Geteuid() == 0

	items := make([]*item, len(cleaners))
	lfIndex := -1

	for i, c := range cleaners {
		skip := c.RequiresRoot() && !isRoot
		items[i] = &item{
			cleaner:  c,
			selected: !skip,
			skip:     skip,
		}

		if _, ok := c.(*cleaner.LargeFileCleaner); ok {
			lfIndex = i
			items[i].selected = false
		}
	}

	return model{
		state:          stateScanning,
		items:          items,
		lfCleanerIndex: lfIndex,
		spinner:        s,
		isRoot:         isRoot,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, scanCmd(m.items, m.isRoot))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		// Global quit (unless in submenu or confirm)
		case "q":
			if m.state == stateLargeFileSelection {
				m.state = stateReview
				return m, nil
			}
			if m.state == stateConfirm {
				m.state = stateReview
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		}

		if m.state == stateReview {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				// Allow cursor to go one past items (to the Clean button)
				if m.cursor < len(m.items) {
					m.cursor++
				}
			case " ":
				// Handle Space on Clean Button
				if m.cursor == len(m.items) {
					// Trigger Clean Logic
					hasSelection := false
					for _, it := range m.items {
						if it.selected {
							hasSelection = true
							break
						}
					}
					if hasSelection {
						m.state = stateConfirm
						if m.lfCleanerIndex >= 0 {
							var selectedPaths []string
							for _, lf := range m.largeFiles {
								if lf.selected {
									selectedPaths = append(selectedPaths, lf.path)
								}
							}
							if lfc, ok := m.items[m.lfCleanerIndex].cleaner.(*cleaner.LargeFileCleaner); ok {
								lfc.SetFilesToClean(selectedPaths)
							}
						}
						return m, nil
					}
					return m, nil
				}

				// Handle Space on Large File Cleaner -> Drill Down
				if m.lfCleanerIndex >= 0 && m.cursor == m.lfCleanerIndex && !m.items[m.cursor].skip {
					m.state = stateLargeFileSelection
					if len(m.largeFiles) == 0 {
						if lfc, ok := m.items[m.lfCleanerIndex].cleaner.(*cleaner.LargeFileCleaner); ok {
							for _, f := range lfc.FoundFiles {
								m.largeFiles = append(m.largeFiles, &largeFileItem{
									path:     f.Path,
									size:     f.Size,
									selected: false,
								})
							}
						}
					}
					return m, nil
				}

				// Normal Toggle Logic for other items
				if m.cursor < len(m.items) {
					it := m.items[m.cursor]
					if !it.skip {
						it.selected = !it.selected
					}
				}

			case "enter":
				// Handle Enter on Clean Button
				if m.cursor == len(m.items) {
					// Trigger Clean Logic
					hasSelection := false
					for _, it := range m.items {
						if it.selected {
							hasSelection = true
							break
						}
					}
					if hasSelection {
						m.state = stateConfirm
						if m.lfCleanerIndex >= 0 {
							var selectedPaths []string
							for _, lf := range m.largeFiles {
								if lf.selected {
									selectedPaths = append(selectedPaths, lf.path)
								}
							}
							if lfc, ok := m.items[m.lfCleanerIndex].cleaner.(*cleaner.LargeFileCleaner); ok {
								lfc.SetFilesToClean(selectedPaths)
							}
						}
						return m, nil
					}
					return m, nil
				}

				// Drill down for Large Files
				if m.cursor == m.lfCleanerIndex && !m.items[m.cursor].skip {
					m.state = stateLargeFileSelection
					if len(m.largeFiles) == 0 {
						if lfc, ok := m.items[m.cursor].cleaner.(*cleaner.LargeFileCleaner); ok {
							for _, f := range lfc.FoundFiles {
								m.largeFiles = append(m.largeFiles, &largeFileItem{
									path:     f.Path,
									size:     f.Size,
									selected: false,
								})
							}
						}
					}
					return m, nil
				}

				// Normal Item -> Toggle
				if m.cursor < len(m.items) {
					it := m.items[m.cursor]
					if !it.skip {
						it.selected = !it.selected
					}
					return m, nil
				}

			case "c": // Hotkey Trigger
				// Trigger Clean Logic
				hasSelection := false
				for _, it := range m.items {
					if it.selected {
						hasSelection = true
						break
					}
				}
				if hasSelection {
					m.state = stateConfirm
					if m.lfCleanerIndex >= 0 {
						var selectedPaths []string
						for _, lf := range m.largeFiles {
							if lf.selected {
								selectedPaths = append(selectedPaths, lf.path)
							}
						}
						if lfc, ok := m.items[m.lfCleanerIndex].cleaner.(*cleaner.LargeFileCleaner); ok {
							lfc.SetFilesToClean(selectedPaths)
						}
					}
					return m, nil
				}
			}

		} else if m.state == stateConfirm {
			switch msg.String() {
			case "y", "Y", "enter": // Confirm
				m.state = stateCleaning
				return m, cleanCmd(m.items)

			case "n", "N", "esc", "backspace": // Cancel
				m.state = stateReview
			}

		} else if m.state == stateLargeFileSelection {
			switch msg.String() {
			case "esc", "backspace", "left", "h":
				m.state = stateReview
			case "up", "k":
				if m.lfCursor > 0 {
					m.lfCursor--
				}
			case "down", "j":
				if m.lfCursor < len(m.largeFiles)-1 {
					m.lfCursor++
				}
			case " ", "enter":
				if len(m.largeFiles) > 0 {
					m.largeFiles[m.lfCursor].selected = !m.largeFiles[m.lfCursor].selected

					// Re-calc summary
					m.lfSelectedCount = 0
					m.lfSelectedSize = 0
					for _, lf := range m.largeFiles {
						if lf.selected {
							m.lfSelectedCount++
							m.lfSelectedSize += lf.size
						}
					}

					m.items[m.lfCleanerIndex].selected = m.lfSelectedCount > 0
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case scanResultMsg:
		for _, it := range m.items {
			if it.cleaner == msg.cleaner {
				it.size = msg.size
				it.err = msg.err
				it.scanned = true
			}
		}

		allScanned := true
		var total int64
		for _, it := range m.items {
			if !it.scanned {
				allScanned = false
			}
			total += it.size
		}
		m.totalSize = total

		if allScanned {
			m.state = stateReview
		}
		return m, nil

	case cleanResultMsg:
		for _, it := range m.items {
			if it.cleaner == msg.cleaner {
				it.cleaned = true
				it.err = msg.err
			}
		}

		allDone := true
		for _, it := range m.items {
			if it.selected && !it.cleaned {
				allDone = false
			}
		}

		if allDone {
			m.state = stateDone
			return m, tea.Quit
		}
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Bye!\n"
	}

	doc := baseDocStyle.
		Width(m.width - 4).
		Height(m.height - 2)

	var s strings.Builder

	s.WriteString(titleStyle.Render("Linux System Cleaner") + "\n\n")

	switch m.state {
	case stateScanning:
		s.WriteString(fmt.Sprintf(" %s Scanning system...\n\n", m.spinner.View()))
		for _, it := range m.items {
			status := "..."
			if it.skip {
				status = orangeStyle.Render("Requires sudo")
			} else if it.scanned {
				if it.err != nil {
					status = redStyle.Render("ERROR")
				} else {
					status = formatBytes(it.size)
				}
			}
			s.WriteString(fmt.Sprintf("  %-35s %s\n", it.cleaner.Name(), status))
		}

	case stateReview:
		s.WriteString(" Select items to clean:\n\n")
		header := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
			fmt.Sprintf("   %-3s %-35s %12s", "   ", "Name", "Size"),
		)
		s.WriteString(header + "\n")

		// Calculate pagination bounds for the main items list
		// Reserve space for headers, footers, titles, etc. (roughly 12 lines)
		availItemRows := m.height - 12
		if availItemRows < 5 {
			availItemRows = 5
		}

		startItem := 0
		endItem := len(m.items)

		// Total items + 1 (for the Clean button)
		totalRenderItems := len(m.items) + 1

		if totalRenderItems > availItemRows {
			startItem = m.cursor - (availItemRows / 2)
			if startItem < 0 {
				startItem = 0
			}
			endItem = startItem + availItemRows - 1 // -1 to leave room for the Clean button if it's visible

			if endItem > len(m.items) {
				endItem = len(m.items)
				startItem = endItem - (availItemRows - 1)
				if startItem < 0 {
					startItem = 0
				}
			}
		}

		for i := startItem; i < endItem; i++ {
			it := m.items[i]
			checked := "[ ]"
			if it.selected {
				checked = "[x]"
			} else if it.skip {
				checked = "[-]"
			}

			cursor := "   "
			if m.cursor == i {
				cursor = " > "
			}

			itemStyle := lipgloss.NewStyle()
			if m.cursor == i {
				itemStyle = selectedItemStyle
			}

			if it.skip {
				itemStyle = itemStyle.Foreground(lipgloss.Color("241"))
			} else if !it.selected && m.cursor != i {
				itemStyle = itemStyle.Foreground(lipgloss.Color("241"))
			}

			name := it.cleaner.Name()
			sizeStr := formatBytes(it.size)

			extras := ""
			if i == m.lfCleanerIndex {
				if m.lfSelectedCount > 0 {
					sizeStr = fmt.Sprintf("%s / %s", formatBytes(m.lfSelectedSize), formatBytes(it.size))
					extras = greenStyle.Render(fmt.Sprintf(" (%d files selected)", m.lfSelectedCount))
				} else if it.size > 0 && !it.skip {
					extras = subtleStyle.Render(" (Enter/Space to detail)")
				}
			}

			if it.skip {
				sizeStr = "Sudo Req."
			}

			if len(name) > 35 {
				name = name[:32] + "..."
			}

			if m.cursor == i {
				s.WriteString(fmt.Sprintf("%s %s %-35s %12s%s\n",
					cursor,
					itemStyle.Render(checked),
					itemStyle.Render(name),
					itemStyle.Render(sizeStr),
					extras,
				))
			} else {
				s.WriteString(itemStyle.Render(fmt.Sprintf("%s %s %-35s %12s", cursor, checked, name, sizeStr)) + extras + "\n")
			}
		}

		// Calculate if there are more items above/below to show an indicator
		if startItem > 0 {
			s.WriteString(subtleStyle.Render("   ↑ more items above\n"))
		} else if endItem < len(m.items) {
			s.WriteString(subtleStyle.Render("   ↓ more items below\n"))
		}

		// Re-calculate Total Selected
		var totalSelectedSize int64
		for i, it := range m.items {
			if i == m.lfCleanerIndex {
				totalSelectedSize += m.lfSelectedSize
			} else if it.selected {
				totalSelectedSize += it.size
			}
		}

		s.WriteString("\n " + greenStyle.Render(fmt.Sprintf("Total Selected to Clean: %s", formatBytes(totalSelectedSize))) + "\n")
		s.WriteString(" " + subtleStyle.Render(fmt.Sprintf("Total Reclaimable: %s", formatBytes(m.totalSize))) + "\n\n")

		btnCursor := "   "
		btnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		if m.cursor == len(m.items) {
			btnCursor = " > "
			btnStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
		}

		// Only render the clean button if it's within the scroll view or we are at the end
		if totalRenderItems <= availItemRows || m.cursor >= len(m.items)-(availItemRows/2) {
			s.WriteString(fmt.Sprintf("%s %s\n", btnCursor, btnStyle.Render("[ CLEAN SELECTED ITEMS ]")))
		} else if endItem == len(m.items) {
			s.WriteString(fmt.Sprintf("%s %s\n", btnCursor, btnStyle.Render("[ CLEAN SELECTED ITEMS ]")))
		} else {
			s.WriteString("\n") // keep layout stable
		}

		s.WriteString(subtleStyle.Render("\n ↑/↓: Navigate • Space: Toggle • Enter: Clean/Details • q: Quit"))

	case stateConfirm:
		// Modal Overlay
		s.Reset() // Clear buffer

		var totalSelectedSize int64
		itemCount := 0
		for i, it := range m.items {
			if i == m.lfCleanerIndex {
				if m.lfSelectedCount > 0 {
					totalSelectedSize += m.lfSelectedSize
					itemCount++
				}
			} else if it.selected {
				totalSelectedSize += it.size
				itemCount++
			}
		}

		modalContent := fmt.Sprintf(
			"Confirmation Required\n\n"+
				"Ready to clean %d categories.\n"+
				"Total size to delete: %s\n\n"+
				"Proceed? (y/N)",
			itemCount,
			formatBytes(totalSelectedSize),
		)

		modal := modalStyle.Render(modalContent)

		availH := m.height - 4
		topPad := availH/2 - 3
		if topPad < 0 {
			topPad = 0
		}

		s.WriteString(strings.Repeat("\n", topPad))
		s.WriteString(lipgloss.PlaceHorizontal(m.width-6, lipgloss.Center, modal))

	case stateLargeFileSelection:
		s.WriteString(" Select Large Files to Delete:\n\n")

		if len(m.largeFiles) == 0 {
			s.WriteString(subtleStyle.Render("  No large unused files found.\n"))
		} else {
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("   %-3s %-50s %10s", "   ", "Path", "Size")) + "\n")

			start, end := m.getPaginatorBounds(m.height - 10)

			for i := start; i < end; i++ {
				lf := m.largeFiles[i]

				checked := "[ ]"
				if lf.selected {
					checked = "[x]"
				}

				cursor := "   "
				if m.lfCursor == i {
					cursor = " > "
				}

				style := lipgloss.NewStyle()
				if m.lfCursor == i {
					style = selectedItemStyle
					checked = style.Render(checked)
				} else if !lf.selected {
					style = style.Foreground(lipgloss.Color("241"))
				}

				path := lf.path
				availWidth := m.width - 25
				if availWidth < 20 {
					availWidth = 20
				}

				if len(path) > availWidth {
					path = "..." + path[len(path)-(availWidth-3):]
				}

				line := fmt.Sprintf("%s %s %-*s %10s", cursor, checked, availWidth, path, formatBytes(lf.size))
				if m.lfCursor == i {
					line = fmt.Sprintf("%s %s %-*s %10s", cursor, checked, availWidth, style.Render(path), style.Render(formatBytes(lf.size)))
				}

				s.WriteString(line + "\n")
			}
		}

		s.WriteString("\n" + subtleStyle.Render(" ↑/↓: Navigate • Space/Enter: Toggle • Esc/Back: Save & Return"))

	case stateCleaning:
		s.WriteString(fmt.Sprintf(" %s Cleaning selected items...\n\n", m.spinner.View()))
		for _, it := range m.items {
			if !it.selected {
				continue
			}
			icon := "•"
			status := "Waiting..."
			if it.cleaned {
				if it.err != nil {
					icon = crossMark.String()
					status = redStyle.Render(fmt.Sprintf("FAILED: %v", it.err))
				} else {
					icon = checkMark.String()
					status = greenStyle.Render("Done")
				}
			}
			s.WriteString(fmt.Sprintf("  %s %-35s %s\n", icon, it.cleaner.Name(), status))
		}

	case stateDone:
		s.WriteString("\n " + greenStyle.Render("Cleanup Complete!") + "\n")
		s.WriteString(subtleStyle.Render("\n Press q to quit."))
	}

	return doc.Render(s.String())
}

func (m model) getPaginatorBounds(maxRows int) (int, int) {
	if maxRows < 5 {
		maxRows = 5
	}
	if len(m.largeFiles) <= maxRows {
		return 0, len(m.largeFiles)
	}
	start := m.lfCursor - (maxRows / 2)
	if start < 0 {
		start = 0
	}
	end := start + maxRows
	if end > len(m.largeFiles) {
		end = len(m.largeFiles)
		start = end - maxRows
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

// Helpers
type scanResultMsg struct {
	cleaner cleaner.Cleaner
	size    int64
	err     error
}

type cleanResultMsg struct {
	cleaner cleaner.Cleaner
	err     error
}

func scanCmd(items []*item, isRoot bool) tea.Cmd {
	var cmds []tea.Cmd
	for _, it := range items {
		c := it.cleaner
		itCopy := it
		cmds = append(cmds, func() tea.Msg {
			if itCopy.skip {
				return scanResultMsg{cleaner: c, size: 0, err: nil}
			}
			size, err := c.Scan()
			return scanResultMsg{cleaner: c, size: size, err: err}
		})
	}
	return tea.Batch(cmds...)
}

func cleanCmd(items []*item) tea.Cmd {
	var cmds []tea.Cmd
	for _, it := range items {
		if it.selected {
			c := it.cleaner
			cmds = append(cmds, func() tea.Msg {
				err := c.Clean()
				return cleanResultMsg{cleaner: c, err: err}
			})
		}
	}
	return tea.Batch(cmds...)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
