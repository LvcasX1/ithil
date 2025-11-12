package types

import (
	"testing"
	"time"
)

func TestUser_GetDisplayName_FullName(t *testing.T) {
	user := &User{
		ID:        12345,
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
	}

	result := user.GetDisplayName()
	expected := "John Doe"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestUser_GetDisplayName_FirstNameOnly(t *testing.T) {
	user := &User{
		ID:        12345,
		FirstName: "Alice",
		Username:  "alice123",
	}

	result := user.GetDisplayName()
	expected := "Alice"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestUser_GetDisplayName_UsernameOnly(t *testing.T) {
	user := &User{
		ID:       12345,
		Username: "bob_user",
	}

	result := user.GetDisplayName()
	expected := "bob_user"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestUser_GetDisplayName_Empty(t *testing.T) {
	user := &User{
		ID: 12345,
	}

	result := user.GetDisplayName()
	expected := ""
	if result != expected {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

func TestUser_GetDisplayName_Priority(t *testing.T) {
	// First name should be prioritized over username
	user := &User{
		ID:        12345,
		FirstName: "Charlie",
		Username:  "charlie99",
	}

	result := user.GetDisplayName()
	if result != "Charlie" {
		t.Errorf("Expected 'Charlie', got '%s'", result)
	}
}

func TestChatType_Constants(t *testing.T) {
	// Verify chat type constants are defined
	types := []ChatType{
		ChatTypePrivate,
		ChatTypeGroup,
		ChatTypeSupergroup,
		ChatTypeChannel,
		ChatTypeSecret,
	}

	for i, ct := range types {
		if int(ct) != i {
			t.Errorf("Expected chat type %d to have value %d, got %d", i, i, int(ct))
		}
	}
}

func TestUserStatus_Constants(t *testing.T) {
	// Verify user status constants are defined
	statuses := []UserStatus{
		UserStatusOnline,
		UserStatusOffline,
		UserStatusRecently,
		UserStatusLastWeek,
		UserStatusLastMonth,
	}

	for i, status := range statuses {
		if int(status) != i {
			t.Errorf("Expected status %d to have value %d, got %d", i, i, int(status))
		}
	}
}

func TestMessage_Creation(t *testing.T) {
	now := time.Now()
	msg := &Message{
		ID:         123,
		ChatID:     456,
		SenderID:   789,
		Date:       now,
		IsOutgoing: true,
		Content: MessageContent{
			Type: MessageTypeText,
			Text: "Hello, World!",
		},
	}

	if msg.ID != 123 {
		t.Errorf("Expected message ID 123, got %d", msg.ID)
	}
	if msg.ChatID != 456 {
		t.Errorf("Expected chat ID 456, got %d", msg.ChatID)
	}
	if msg.SenderID != 789 {
		t.Errorf("Expected sender ID 789, got %d", msg.SenderID)
	}
	if !msg.IsOutgoing {
		t.Error("Expected message to be outgoing")
	}
	if msg.Content.Text != "Hello, World!" {
		t.Errorf("Expected text 'Hello, World!', got '%s'", msg.Content.Text)
	}
}

func TestChat_Creation(t *testing.T) {
	chat := &Chat{
		ID:          12345,
		Type:        ChatTypePrivate,
		Title:       "Test Chat",
		Username:    "testchat",
		UnreadCount: 5,
		IsPinned:    true,
		PinOrder:    1,
	}

	if chat.ID != 12345 {
		t.Errorf("Expected chat ID 12345, got %d", chat.ID)
	}
	if chat.Type != ChatTypePrivate {
		t.Errorf("Expected ChatTypePrivate, got %d", chat.Type)
	}
	if chat.Title != "Test Chat" {
		t.Errorf("Expected title 'Test Chat', got '%s'", chat.Title)
	}
	if chat.UnreadCount != 5 {
		t.Errorf("Expected unread count 5, got %d", chat.UnreadCount)
	}
	if !chat.IsPinned {
		t.Error("Expected chat to be pinned")
	}
	if chat.PinOrder != 1 {
		t.Errorf("Expected pin order 1, got %d", chat.PinOrder)
	}
}

func TestUser_BooleanFields(t *testing.T) {
	user := &User{
		ID:              12345,
		FirstName:       "Bot",
		IsBot:           true,
		IsVerified:      true,
		IsPremium:       false,
		IsContact:       true,
		IsMutualContact: false,
	}

	if !user.IsBot {
		t.Error("Expected user to be a bot")
	}
	if !user.IsVerified {
		t.Error("Expected user to be verified")
	}
	if user.IsPremium {
		t.Error("Expected user not to be premium")
	}
	if !user.IsContact {
		t.Error("Expected user to be a contact")
	}
	if user.IsMutualContact {
		t.Error("Expected user not to be a mutual contact")
	}
}

func TestMessage_BooleanFields(t *testing.T) {
	msg := &Message{
		ID:            123,
		ChatID:        456,
		IsOutgoing:    true,
		IsChannelPost: false,
		IsPinned:      true,
		IsEdited:      true,
		IsForwarded:   false,
	}

	if !msg.IsOutgoing {
		t.Error("Expected message to be outgoing")
	}
	if msg.IsChannelPost {
		t.Error("Expected message not to be a channel post")
	}
	if !msg.IsPinned {
		t.Error("Expected message to be pinned")
	}
	if !msg.IsEdited {
		t.Error("Expected message to be edited")
	}
	if msg.IsForwarded {
		t.Error("Expected message not to be forwarded")
	}
}

func TestMessageContent_TextMessage(t *testing.T) {
	content := MessageContent{
		Type: MessageTypeText,
		Text: "This is a test message",
	}

	if content.Type != MessageTypeText {
		t.Errorf("Expected MessageTypeText, got %d", content.Type)
	}
	if content.Text != "This is a test message" {
		t.Errorf("Expected specific text, got '%s'", content.Text)
	}
}
