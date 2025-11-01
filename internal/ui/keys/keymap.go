// Package keys defines keyboard shortcuts and key bindings for the Ithil TUI.
package keys

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keyboard shortcuts for the application.
type KeyMap struct {
	// Global navigation
	Quit    key.Binding
	Help    key.Binding
	Back    key.Binding
	Refresh key.Binding
	Search  key.Binding

	// Pane navigation
	NextPane          key.Binding
	PrevPane          key.Binding
	FocusChatList     key.Binding
	FocusConversation key.Binding
	FocusSidebar      key.Binding

	// List navigation (vim-style)
	Up           key.Binding
	Down         key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	GotoTop      key.Binding
	GotoBottom   key.Binding

	// Chat list actions
	OpenChat    key.Binding
	PinChat     key.Binding
	MuteChat    key.Binding
	ArchiveChat key.Binding
	MarkRead    key.Binding
	DeleteChat  key.Binding

	// Conversation actions
	Reply    key.Binding
	Edit     key.Binding
	Delete   key.Binding
	Forward  key.Binding
	Copy     key.Binding
	React    key.Binding
	Pin      key.Binding
	Download key.Binding

	// Message input
	Send        key.Binding
	SendAlt     key.Binding // Ctrl+Enter if Enter sends, or vice versa
	NewLine     key.Binding
	Attach      key.Binding
	Emoji       key.Binding
	CancelReply key.Binding
	CancelEdit  key.Binding

	// Media actions
	ViewMedia key.Binding
	OpenLink  key.Binding

	// Sidebar
	ToggleSidebar key.Binding
	UserInfo      key.Binding
	ChatInfo      key.Binding

	// Settings
	Settings      key.Binding
	ToggleTheme   key.Binding
	ToggleStealth key.Binding

	// Filters/Folders
	ShowFilters key.Binding
	NextFilter  key.Binding
	PrevFilter  key.Binding
}

// DefaultKeyMap returns the default keyboard shortcuts.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Global navigation
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "ctrl+q"),
			key.WithHelp("ctrl+c/q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back/cancel"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh"),
		),
		Search: key.NewBinding(
			key.WithKeys("/", "ctrl+f"),
			key.WithHelp("/ or ctrl+f", "search"),
		),

		// Pane navigation
		NextPane: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next pane"),
		),
		PrevPane: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous pane"),
		),
		FocusChatList: key.NewBinding(
			key.WithKeys("ctrl+1"),
			key.WithHelp("ctrl+1", "focus chat list"),
		),
		FocusConversation: key.NewBinding(
			key.WithKeys("ctrl+2"),
			key.WithHelp("ctrl+2", "focus conversation"),
		),
		FocusSidebar: key.NewBinding(
			key.WithKeys("ctrl+3"),
			key.WithHelp("ctrl+3", "focus sidebar"),
		),

		// List navigation (vim-style)
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+b"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+f"),
			key.WithHelp("pgdown", "page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "half page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "half page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g/home", "go to top"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G/end", "go to bottom"),
		),

		// Chat list actions
		OpenChat: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open chat"),
		),
		PinChat: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pin/unpin chat"),
		),
		MuteChat: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "mute/unmute chat"),
		),
		ArchiveChat: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "archive chat"),
		),
		MarkRead: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "mark as read"),
		),
		DeleteChat: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete chat"),
		),

		// Conversation actions
		Reply: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reply to message"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit message"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete message"),
		),
		Forward: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "forward message"),
		),
		Copy: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy message"),
		),
		React: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "react to message"),
		),
		Pin: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pin message"),
		),
		Download: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "save/download"),
		),

		// Message input
		Send: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "send message"),
		),
		SendAlt: key.NewBinding(
			key.WithKeys("ctrl+enter"),
			key.WithHelp("ctrl+enter", "send message"),
		),
		NewLine: key.NewBinding(
			key.WithKeys("shift+enter"),
			key.WithHelp("shift+enter", "new line"),
		),
		Attach: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "attach file"),
		),
		Emoji: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("ctrl+e", "insert emoji"),
		),
		CancelReply: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel reply"),
		),
		CancelEdit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel edit"),
		),

		// Media actions
		ViewMedia: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "view media"),
		),
		OpenLink: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open link"),
		),

		// Sidebar
		ToggleSidebar: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "toggle sidebar"),
		),
		UserInfo: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "user info"),
		),
		ChatInfo: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "chat info"),
		),

		// Settings
		Settings: key.NewBinding(
			key.WithKeys("ctrl+,"),
			key.WithHelp("ctrl+,", "settings"),
		),
		ToggleTheme: key.NewBinding(
			key.WithKeys("ctrl+t"),
			key.WithHelp("ctrl+t", "toggle theme"),
		),
		ToggleStealth: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle stealth mode"),
		),

		// Filters/Folders
		ShowFilters: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "show filters"),
		),
		NextFilter: key.NewBinding(
			key.WithKeys("ctrl+]"),
			key.WithHelp("ctrl+]", "next filter"),
		),
		PrevFilter: key.NewBinding(
			key.WithKeys("ctrl+["),
			key.WithHelp("ctrl+[", "previous filter"),
		),
	}
}

// VimKeyMap returns a keymap with additional vim-style bindings.
func VimKeyMap() KeyMap {
	km := DefaultKeyMap()

	// Add additional vim-style bindings
	km.Quit.SetKeys("ctrl+c", "ctrl+q", ":q")
	km.Search.SetKeys("/", "ctrl+f", ":s")

	return km
}

// ShortHelp returns a slice of key bindings to show in short help.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Help,
		k.Quit,
		k.Up,
		k.Down,
		k.OpenChat,
	}
}

// FullHelp returns all key bindings organized by category.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Navigation
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.GotoTop, k.GotoBottom, k.NextPane, k.PrevPane},

		// Actions
		{k.OpenChat, k.Send, k.Reply, k.Edit},
		{k.Delete, k.Forward, k.Copy, k.Pin},

		// Chat management
		{k.PinChat, k.MuteChat, k.MarkRead, k.ArchiveChat},

		// Global
		{k.Search, k.Refresh, k.Settings, k.ToggleStealth, k.Help, k.Quit},
	}
}

// ChatListKeys returns key bindings relevant to the chat list.
func (k KeyMap) ChatListKeys() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.OpenChat,
		k.PinChat,
		k.MuteChat,
		k.MarkRead,
		k.ArchiveChat,
		k.Search,
	}
}

// ConversationKeys returns key bindings relevant to the conversation view.
func (k KeyMap) ConversationKeys() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Reply,
		k.Edit,
		k.Delete,
		k.Forward,
		k.Copy,
		k.React,
		k.Pin,
		k.Download,
	}
}

// InputKeys returns key bindings relevant to message input.
func (k KeyMap) InputKeys() []key.Binding {
	return []key.Binding{
		k.Send,
		k.SendAlt,
		k.NewLine,
		k.Attach,
		k.Emoji,
		k.CancelReply,
		k.CancelEdit,
	}
}
