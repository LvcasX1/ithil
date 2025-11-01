// Package components provides reusable UI components for the Ithil TUI.
package components

import (
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/ui/styles"
)

// FilePickerComponent represents a simple file picker dialog.
type FilePickerComponent struct {
	currentPath string
	entries     []os.DirEntry
	selected    int
	width       int
	height      int
	isFile      bool
	err         error
}

// NewFilePickerComponent creates a new file picker component.
func NewFilePickerComponent(initialPath string) *FilePickerComponent {
	if initialPath == "" {
		// Default to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			initialPath = "."
		} else {
			initialPath = home
		}
	}

	fp := &FilePickerComponent{
		currentPath: initialPath,
		width:       80,
		height:      20,
	}

	fp.loadDirectory()
	return fp
}

// loadDirectory loads the current directory contents.
func (f *FilePickerComponent) loadDirectory() {
	entries, err := os.ReadDir(f.currentPath)
	if err != nil {
		f.err = err
		return
	}
	f.entries = entries
	f.selected = 0
	f.err = nil
}

// Init initializes the file picker component.
func (f *FilePickerComponent) Init() tea.Cmd {
	return nil
}

// Update handles file picker component updates.
func (f *FilePickerComponent) Update(msg tea.Msg) (*FilePickerComponent, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if f.selected > 0 {
				f.selected--
			}
		case "down", "j":
			if f.selected < len(f.entries)-1 {
				f.selected++
			}
		case "enter":
			// Select current entry
			if f.selected < len(f.entries) {
				entry := f.entries[f.selected]
				newPath := filepath.Join(f.currentPath, entry.Name())

				if entry.IsDir() {
					// Navigate into directory
					f.currentPath = newPath
					f.loadDirectory()
				} else {
					// Select file
					f.isFile = true
					return f, func() tea.Msg {
						return FileSelectedMsg{Path: newPath}
					}
				}
			}
		case "backspace", "h":
			// Go to parent directory
			parent := filepath.Dir(f.currentPath)
			if parent != f.currentPath {
				f.currentPath = parent
				f.loadDirectory()
			}
		case "esc", "q":
			// Cancel file picker
			return f, func() tea.Msg {
				return FilePickerCancelledMsg{}
			}
		}
	}

	return f, nil
}

// View renders the file picker component.
func (f *FilePickerComponent) View() string {
	var sb strings.Builder

	// Title
	title := " Select a file to attach "
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.AccentCyan)).
		Bold(true).
		Align(lipgloss.Center)
	sb.WriteString(titleStyle.Width(f.width).Render(title))
	sb.WriteString("\n")

	// Current path
	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Italic(true)
	sb.WriteString(pathStyle.Render("Path: " + f.currentPath))
	sb.WriteString("\n\n")

	// Error if any
	if f.err != nil {
		errorText := "Error: " + f.err.Error()
		sb.WriteString(styles.ErrorStyle.Render(errorText))
		sb.WriteString("\n\n")
	}

	// File list
	visibleHeight := f.height - 8 // Reserve space for title, path, help
	startIdx := 0
	if f.selected > visibleHeight/2 && len(f.entries) > visibleHeight {
		startIdx = f.selected - visibleHeight/2
		if startIdx+visibleHeight > len(f.entries) {
			startIdx = len(f.entries) - visibleHeight
		}
	}

	endIdx := startIdx + visibleHeight
	if endIdx > len(f.entries) {
		endIdx = len(f.entries)
	}

	for i := startIdx; i < endIdx; i++ {
		entry := f.entries[i]
		name := entry.Name()

		// Add directory indicator
		if entry.IsDir() {
			name = "📁 " + name + "/"
		} else {
			name = "📄 " + name
		}

		// Highlight selected
		if i == f.selected {
			sb.WriteString(styles.SelectedStyle.Render("> " + name))
		} else {
			sb.WriteString("  " + name)
		}
		sb.WriteString("\n")
	}

	// Help text
	sb.WriteString("\n")
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary))
	helpText := "↑/↓: Navigate • Enter: Select • Backspace: Parent dir • Esc: Cancel"
	sb.WriteString(helpStyle.Render(helpText))

	// Container with border
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(styles.AccentCyan)).
		Padding(1, 2).
		Width(f.width).
		Height(f.height)

	return containerStyle.Render(sb.String())
}

// SetSize sets the size of the file picker.
func (f *FilePickerComponent) SetSize(width, height int) {
	f.width = width
	f.height = height
}

// GetCurrentPath returns the current directory path.
func (f *FilePickerComponent) GetCurrentPath() string {
	return f.currentPath
}

// FileSelectedMsg is sent when a file is selected.
type FileSelectedMsg struct {
	Path string
}

// FilePickerCancelledMsg is sent when the file picker is cancelled.
type FilePickerCancelledMsg struct{}
