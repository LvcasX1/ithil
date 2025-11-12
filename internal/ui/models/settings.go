package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/ithil/internal/app"
	"github.com/lvcasx1/ithil/internal/ui/components"
	"github.com/lvcasx1/ithil/internal/ui/styles"
)

// SettingsTab represents different settings tabs
type SettingsTab int

const (
	SettingsTabAccount SettingsTab = iota
	// Future tabs can be added here:
	// SettingsTabAppearance
	// SettingsTabPrivacy
	// SettingsTabAdvanced
)

// SettingsModel represents the settings screen.
type SettingsModel struct {
	config          *app.Config
	width           int
	height          int
	currentTab      SettingsTab
	useDefaultCreds bool
	apiIDInput      textinput.Model
	apiHashInput    textinput.Model
	focusedInput    int // 0 = toggle, 1 = apiID, 2 = apiHash
	saved           bool
	saveError       string
}

// NewSettingsModel creates a new settings model.
func NewSettingsModel(config *app.Config) *SettingsModel {
	// API ID input
	apiIDInput := textinput.New()
	apiIDInput.Placeholder = "Your API ID (7-9 digits)"
	apiIDInput.CharLimit = 10
	apiIDInput.Width = 30
	if config.Telegram.APIID != "" {
		apiIDInput.SetValue(config.Telegram.APIID)
	}

	// API Hash input
	apiHashInput := textinput.New()
	apiHashInput.Placeholder = "Your API Hash (32 characters)"
	apiHashInput.CharLimit = 32
	apiHashInput.Width = 40
	apiHashInput.EchoMode = textinput.EchoPassword
	apiHashInput.EchoCharacter = '*'
	if config.Telegram.APIHash != "" {
		apiHashInput.SetValue(config.Telegram.APIHash)
	}

	return &SettingsModel{
		config:          config,
		currentTab:      SettingsTabAccount,
		useDefaultCreds: config.Telegram.UseDefaultCredentials,
		apiIDInput:      apiIDInput,
		apiHashInput:    apiHashInput,
		focusedInput:    0,
	}
}

// Init initializes the settings model.
func (m *SettingsModel) Init() tea.Cmd {
	return nil
}

// Update handles settings model updates.
func (m *SettingsModel) Update(msg tea.Msg) (*SettingsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear saved/error messages on any key
		if m.saved || m.saveError != "" {
			m.saved = false
			m.saveError = ""
		}

		switch msg.String() {
		case "tab":
			// Cycle through inputs
			m.focusedInput = (m.focusedInput + 1) % 3
			m.updateInputFocus()
			return m, nil

		case "shift+tab":
			// Cycle backwards through inputs
			m.focusedInput = (m.focusedInput - 1 + 3) % 3
			m.updateInputFocus()
			return m, nil

		case " ", "enter":
			if m.focusedInput == 0 {
				// Toggle credentials mode
				m.useDefaultCreds = !m.useDefaultCreds
				if m.useDefaultCreds {
					// Blur inputs when switching to default
					m.apiIDInput.Blur()
					m.apiHashInput.Blur()
				}
				return m, nil
			}

		case "ctrl+s":
			// Save settings
			return m, m.saveSettings()

		case "esc":
			// Close settings (handled by parent)
			return m, nil
		}

		// Update focused input
		if m.focusedInput == 1 && !m.useDefaultCreds {
			m.apiIDInput, cmd = m.apiIDInput.Update(msg)
			return m, cmd
		} else if m.focusedInput == 2 && !m.useDefaultCreds {
			m.apiHashInput, cmd = m.apiHashInput.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the settings screen.
func (m *SettingsModel) View() string {
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(styles.AccentCyan)).
		Render("Settings")

	closeHelp := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Render("[Esc to close | Ctrl+S to save]")

	headerLine := lipgloss.JoinHorizontal(lipgloss.Left,
		header,
		strings.Repeat(" ", m.width-lipgloss.Width(header)-lipgloss.Width(closeHelp)-2),
		closeHelp,
	)

	b.WriteString(headerLine)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n\n")

	// Tab header (for future expansion)
	tabStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(styles.AccentCyan))

	b.WriteString(tabStyle.Render("Account"))
	b.WriteString("  ")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Render("Appearance  Privacy  Advanced"))
	b.WriteString("\n\n")

	// Content
	b.WriteString(m.renderAccountTab())

	return b.String()
}

// renderAccountTab renders the account settings tab.
func (m *SettingsModel) renderAccountTab() string {
	var b strings.Builder

	// Section title
	sectionTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(styles.TextBright)).
		Render("API Credentials")

	b.WriteString(sectionTitle)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 20))
	b.WriteString("\n\n")

	// Current status
	status := "● Using Default Credentials"
	statusColor := styles.AccentGreen
	if !m.useDefaultCreds && !app.IsUsingDefaultCredentials(m.config) {
		status = "● Using Custom Credentials"
		statusColor = styles.AccentCyan
	}

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(statusColor)).
		Bold(true)

	b.WriteString("Status: ")
	b.WriteString(statusStyle.Render(status))
	b.WriteString("\n\n")

	// Info box about default credentials
	infoBox := components.InfoMessage(
		"Default credentials are shared across all users.\nFor enhanced privacy, use your own API credentials from https://my.telegram.org",
		m.width-4,
	)
	b.WriteString(infoBox)
	b.WriteString("\n\n")

	// Toggle
	toggleIndicator := "○"
	if m.useDefaultCreds {
		toggleIndicator = "●"
	}

	toggleStyle := lipgloss.NewStyle()
	if m.focusedInput == 0 {
		toggleStyle = toggleStyle.
			Foreground(lipgloss.Color(styles.AccentCyan)).
			Bold(true)
	}

	b.WriteString(toggleStyle.Render(toggleIndicator + " Use default credentials (zero setup)"))
	b.WriteString("\n")

	toggleIndicator2 := "○"
	if !m.useDefaultCreds {
		toggleIndicator2 = "●"
	}

	b.WriteString(toggleStyle.Render(toggleIndicator2 + " Use custom credentials (enhanced privacy)"))
	b.WriteString("\n\n")

	// Custom credentials inputs (only shown if not using defaults)
	if !m.useDefaultCreds {
		b.WriteString("API ID:   ")
		b.WriteString(m.apiIDInput.View())
		b.WriteString("\n")

		b.WriteString("API Hash: ")
		b.WriteString(m.apiHashInput.View())
		b.WriteString("\n\n")

		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color(styles.TextSecondary)).
			Render("Get credentials from: https://my.telegram.org"))
		b.WriteString("\n\n")
	} else {
		b.WriteString("\n")
	}

	// Warning box
	warningBox := components.WarningMessage(
		"Changing credentials requires restart and re-login\nYour session will be cleared for privacy",
		m.width-4,
	)
	b.WriteString(warningBox)
	b.WriteString("\n\n")

	// Save status
	if m.saved {
		successMsg := components.SuccessMessage(
			"Settings saved! Please restart Ithil to apply changes.",
			m.width-4,
		)
		b.WriteString(successMsg)
		b.WriteString("\n")
	} else if m.saveError != "" {
		errorMsg := components.ErrorMessage(
			"Save failed: "+m.saveError,
			m.width-4,
		)
		b.WriteString(errorMsg)
		b.WriteString("\n")
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TextSecondary)).
		Render("Tab: Next field | Space/Enter: Toggle | Ctrl+S: Save | Esc: Close"))

	return b.String()
}

// updateInputFocus updates which input is focused.
func (m *SettingsModel) updateInputFocus() {
	m.apiIDInput.Blur()
	m.apiHashInput.Blur()

	if !m.useDefaultCreds {
		if m.focusedInput == 1 {
			m.apiIDInput.Focus()
		} else if m.focusedInput == 2 {
			m.apiHashInput.Focus()
		}
	}
}

// saveSettings saves the settings to config file.
func (m *SettingsModel) saveSettings() tea.Cmd {
	return func() tea.Msg {
		// Update config
		m.config.Telegram.UseDefaultCredentials = m.useDefaultCreds

		if !m.useDefaultCreds {
			// Validate custom credentials
			apiID := strings.TrimSpace(m.apiIDInput.Value())
			apiHash := strings.TrimSpace(m.apiHashInput.Value())

			if apiID == "" || apiHash == "" {
				m.saveError = "API ID and API Hash are required for custom credentials"
				return nil
			}

			// Validate API ID is numeric
			if _, err := strconv.Atoi(apiID); err != nil {
				m.saveError = "API ID must be a number"
				return nil
			}

			// Validate API Hash length
			if len(apiHash) != 32 {
				m.saveError = "API Hash must be 32 characters"
				return nil
			}

			m.config.Telegram.APIID = apiID
			m.config.Telegram.APIHash = apiHash
		} else {
			// Clear custom credentials when using defaults
			m.config.Telegram.APIID = ""
			m.config.Telegram.APIHash = ""
		}

		// Determine save path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			m.saveError = "Failed to get home directory"
			return nil
		}

		configPath := filepath.Join(homeDir, ".config", "ithil", "config.yaml")

		// Save config
		if err := m.config.SaveConfig(configPath); err != nil {
			m.saveError = err.Error()
			return nil
		}

		// Create marker file for session clearing
		markerPath := filepath.Join(homeDir, ".config", "ithil", ".credential-change")
		if err := os.WriteFile(markerPath, []byte("1"), 0600); err != nil {
			// Not critical, just log
			fmt.Fprintf(os.Stderr, "Warning: Could not create credential change marker: %v\n", err)
		}

		m.saved = true
		m.saveError = ""

		return nil
	}
}

// SetSize updates the model size.
func (m *SettingsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
