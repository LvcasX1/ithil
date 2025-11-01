// Package models provides Bubbletea models for the Ithil TUI.
package models

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lvcasx1/ithil/internal/telegram"
	"github.com/lvcasx1/ithil/internal/ui/styles"
	"github.com/lvcasx1/ithil/pkg/types"
)

// AuthModel represents the authentication screen.
type AuthModel struct {
	client       *telegram.Client
	width        int
	height       int
	state        types.AuthState
	input        textinput.Model
	errorMessage string
	loading      bool
}

// NewAuthModel creates a new authentication model.
func NewAuthModel(client *telegram.Client) *AuthModel {
	ti := textinput.New()
	ti.Placeholder = "Phone number (e.g., +1234567890)"
	ti.Focus()
	ti.CharLimit = 20
	ti.Width = 30

	return &AuthModel{
		client: client,
		state:  types.AuthStateWaitPhoneNumber,
		input:  ti,
	}
}

// Init initializes the auth model.
func (m *AuthModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles auth model updates.
func (m *AuthModel) Update(msg tea.Msg) (*AuthModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, m.handleSubmit()
		case "esc", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case authStateMsg:
		m.loading = false
		m.state = msg.state
		m.updateInputForState()
		return m, nil

	case authErrorMsg:
		m.loading = false
		m.errorMessage = msg.error
		return m, nil
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the auth model.
func (m *AuthModel) View() string {
	// Build the view content
	content := m.buildContent()

	// Center the content
	containerStyle := styles.AuthContainerStyle.
		Width(m.width).
		Height(m.height)

	return containerStyle.Render(content)
}

// buildContent builds the auth screen content.
func (m *AuthModel) buildContent() string {
	var content string

	// Title
	title := styles.AuthTitleStyle.Render("Welcome to Ithil")
	content += title + "\n\n"

	// Show loading indicator if loading
	if m.loading {
		loadingMsg := m.getLoadingMessage()
		content += styles.InfoStyle.Render(loadingMsg) + "\n\n"
		content += styles.DimStyle.Render("Please wait...")
		return content
	}

	// Prompt based on state
	prompt := m.getPromptForState()
	content += styles.AuthPromptStyle.Render(prompt) + "\n\n"

	// Input field
	inputView := m.input.View()
	content += styles.AuthInputStyle.Render(inputView) + "\n\n"

	// Error message if any
	if m.errorMessage != "" {
		content += styles.AuthErrorStyle.Render(m.errorMessage) + "\n\n"
	}

	// Help text
	helpText := "Enter: Submit • Esc: Quit"
	content += styles.DimStyle.Render(helpText)

	return content
}

// getPromptForState returns the prompt text for the current auth state.
func (m *AuthModel) getPromptForState() string {
	switch m.state {
	case types.AuthStateWaitPhoneNumber:
		return "Please enter your phone number:"
	case types.AuthStateWaitCode:
		return "Please enter the verification code:"
	case types.AuthStateWaitPassword:
		return "Please enter your 2FA password:"
	case types.AuthStateWaitRegistration:
		return "Please enter your name:"
	default:
		return "Authenticating..."
	}
}

// getLoadingMessage returns the loading message for the current auth state.
func (m *AuthModel) getLoadingMessage() string {
	switch m.state {
	case types.AuthStateWaitPhoneNumber:
		return "Sending verification code..."
	case types.AuthStateWaitCode:
		return "Verifying code..."
	case types.AuthStateWaitPassword:
		return "Verifying password..."
	case types.AuthStateWaitRegistration:
		return "Registering account..."
	default:
		return "Authenticating..."
	}
}

// updateInputForState updates the input field for the current state.
func (m *AuthModel) updateInputForState() {
	switch m.state {
	case types.AuthStateWaitPhoneNumber:
		m.input.Placeholder = "Phone number (e.g., +1234567890)"
		m.input.CharLimit = 20
		m.input.EchoMode = textinput.EchoNormal
	case types.AuthStateWaitCode:
		m.input.Placeholder = "Verification code"
		m.input.CharLimit = 10
		m.input.EchoMode = textinput.EchoNormal
	case types.AuthStateWaitPassword:
		m.input.Placeholder = "2FA password"
		m.input.CharLimit = 100
		m.input.EchoMode = textinput.EchoPassword
	case types.AuthStateWaitRegistration:
		m.input.Placeholder = "First name"
		m.input.CharLimit = 50
		m.input.EchoMode = textinput.EchoNormal
	}

	m.input.SetValue("")
	m.errorMessage = ""
}

// handleSubmit handles the submit action.
func (m *AuthModel) handleSubmit() tea.Cmd {
	value := m.input.Value()
	if value == "" {
		return nil
	}

	// Set loading state
	m.loading = true
	m.errorMessage = ""

	switch m.state {
	case types.AuthStateWaitPhoneNumber:
		return m.sendPhoneNumber(value)
	case types.AuthStateWaitCode:
		return m.sendCode(value)
	case types.AuthStateWaitPassword:
		return m.sendPassword(value)
	case types.AuthStateWaitRegistration:
		return m.sendRegistration(value)
	}

	return nil
}

// sendPhoneNumber sends the phone number.
func (m *AuthModel) sendPhoneNumber(phoneNumber string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.SendPhoneNumber(phoneNumber)
		if err != nil {
			return authErrorMsg{error: err.Error()}
		}
		return authStateMsg{state: types.AuthStateWaitCode}
	}
}

// sendCode sends the verification code.
func (m *AuthModel) sendCode(code string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.SendAuthCode(code)
		if err != nil {
			return authErrorMsg{error: err.Error()}
		}
		return authStateMsg{state: types.AuthStateReady}
	}
}

// sendPassword sends the 2FA password.
func (m *AuthModel) sendPassword(password string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.SendPassword(password)
		if err != nil {
			return authErrorMsg{error: err.Error()}
		}
		return authStateMsg{state: types.AuthStateReady}
	}
}

// sendRegistration sends registration information.
func (m *AuthModel) sendRegistration(name string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.SendRegistration(name, "")
		if err != nil {
			return authErrorMsg{error: err.Error()}
		}
		return authStateMsg{state: types.AuthStateReady}
	}
}

// SetSize sets the size of the auth model.
func (m *AuthModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Messages for auth state changes.
type authStateMsg struct {
	state types.AuthState
}

type authErrorMsg struct {
	error string
}
