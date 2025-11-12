package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/lvcasx1/ithil/pkg/types"
)

func TestNew(t *testing.T) {
	cache := New(100)
	if cache == nil {
		t.Fatal("New() returned nil")
	}
	if cache.maxMessagesPerChat != 100 {
		t.Errorf("Expected maxMessagesPerChat 100, got %d", cache.maxMessagesPerChat)
	}
}

func TestSetAndGetMessage(t *testing.T) {
	cache := New(100)
	chatID := int64(12345)
	message := &types.Message{
		ID:      1,
		ChatID:  chatID,
		Content: types.MessageContent{Text: "Hello, World!"},
		Date:    time.Now(),
	}

	// Set message
	cache.SetMessage(chatID, message)

	// Get message
	retrieved, exists := cache.GetMessage(chatID, message.ID)
	if !exists {
		t.Fatal("Message not found in cache")
	}
	if retrieved.ID != message.ID {
		t.Errorf("Expected message ID %d, got %d", message.ID, retrieved.ID)
	}
	if retrieved.Content.Text != "Hello, World!" {
		t.Errorf("Expected text 'Hello, World!', got '%s'", retrieved.Content.Text)
	}
}

func TestGetMessage_NotFound(t *testing.T) {
	cache := New(100)
	_, exists := cache.GetMessage(12345, 999)
	if exists {
		t.Error("Expected message not to exist")
	}
}

func TestAddMessage(t *testing.T) {
	cache := New(100)
	chatID := int64(12345)
	message := &types.Message{
		ID:      1,
		ChatID:  chatID,
		Content: types.MessageContent{Text: "Test"},
		Date:    time.Now(),
	}

	cache.AddMessage(chatID, message)

	retrieved, exists := cache.GetMessage(chatID, message.ID)
	if !exists {
		t.Fatal("Message not found after AddMessage")
	}
	if retrieved.ID != message.ID {
		t.Errorf("Expected message ID %d, got %d", message.ID, retrieved.ID)
	}
}

func TestGetMessages(t *testing.T) {
	cache := New(100)
	chatID := int64(12345)

	// Add multiple messages
	for i := int64(1); i <= 5; i++ {
		message := &types.Message{
			ID:      i,
			ChatID:  chatID,
			Content: types.MessageContent{Text: "Message"},
			Date:    time.Now(),
		}
		cache.SetMessage(chatID, message)
	}

	messages := cache.GetMessages(chatID)
	if len(messages) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(messages))
	}
}

func TestGetMessages_EmptyChat(t *testing.T) {
	cache := New(100)
	messages := cache.GetMessages(99999)
	if messages == nil {
		t.Error("Expected empty slice, got nil")
	}
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(messages))
	}
}

func TestDeleteMessage(t *testing.T) {
	cache := New(100)
	chatID := int64(12345)
	message := &types.Message{
		ID:      1,
		ChatID:  chatID,
		Content: types.MessageContent{Text: "Test"},
		Date:    time.Now(),
	}

	cache.SetMessage(chatID, message)
	cache.DeleteMessage(chatID, message.ID)

	_, exists := cache.GetMessage(chatID, message.ID)
	if exists {
		t.Error("Message should have been deleted")
	}
}

func TestSetAndGetChat(t *testing.T) {
	cache := New(100)
	chat := &types.Chat{
		ID:    12345,
		Title: "Test Chat",
		Type:  types.ChatTypePrivate,
	}

	cache.SetChat(chat)

	retrieved, exists := cache.GetChat(chat.ID)
	if !exists {
		t.Fatal("Chat not found in cache")
	}
	if retrieved.Title != "Test Chat" {
		t.Errorf("Expected title 'Test Chat', got '%s'", retrieved.Title)
	}
}

func TestGetChats(t *testing.T) {
	cache := New(100)

	// Add multiple chats
	for i := int64(1); i <= 3; i++ {
		chat := &types.Chat{
			ID:    i,
			Title: "Chat",
			Type:  types.ChatTypePrivate,
		}
		cache.SetChat(chat)
	}

	chats := cache.GetChats()
	if len(chats) != 3 {
		t.Errorf("Expected 3 chats, got %d", len(chats))
	}
}

func TestDeleteChat(t *testing.T) {
	cache := New(100)
	chatID := int64(12345)

	chat := &types.Chat{
		ID:    chatID,
		Title: "Test",
		Type:  types.ChatTypePrivate,
	}
	cache.SetChat(chat)

	message := &types.Message{
		ID:      1,
		ChatID:  chatID,
		Content: types.MessageContent{Text: "Test"},
		Date:    time.Now(),
	}
	cache.SetMessage(chatID, message)

	// Delete chat should delete both chat and messages
	cache.DeleteChat(chatID)

	_, exists := cache.GetChat(chatID)
	if exists {
		t.Error("Chat should have been deleted")
	}

	messages := cache.GetMessages(chatID)
	if len(messages) != 0 {
		t.Error("Messages should have been deleted with chat")
	}
}

func TestSetAndGetUser(t *testing.T) {
	cache := New(100)
	user := &types.User{
		ID:        12345,
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
	}

	cache.SetUser(user)

	retrieved, exists := cache.GetUser(user.ID)
	if !exists {
		t.Fatal("User not found in cache")
	}
	if retrieved.FirstName != "John" {
		t.Errorf("Expected first name 'John', got '%s'", retrieved.FirstName)
	}
	if retrieved.Username != "johndoe" {
		t.Errorf("Expected username 'johndoe', got '%s'", retrieved.Username)
	}
}

func TestGetUsers(t *testing.T) {
	cache := New(100)

	// Add multiple users
	for i := int64(1); i <= 4; i++ {
		user := &types.User{
			ID:        i,
			FirstName: "User",
			Username:  "user",
		}
		cache.SetUser(user)
	}

	users := cache.GetUsers()
	if len(users) != 4 {
		t.Errorf("Expected 4 users, got %d", len(users))
	}
}

func TestClear(t *testing.T) {
	cache := New(100)

	// Add some data
	cache.SetChat(&types.Chat{ID: 1, Title: "Chat", Type: types.ChatTypePrivate})
	cache.SetUser(&types.User{ID: 1, FirstName: "User", Username: "user"})
	cache.SetMessage(1, &types.Message{ID: 1, ChatID: 1, Content: types.MessageContent{Text: "Msg"}, Date: time.Now()})

	cache.Clear()

	if cache.GetChatCount() != 0 {
		t.Error("Chats should be cleared")
	}
	if cache.GetUserCount() != 0 {
		t.Error("Users should be cleared")
	}
	if cache.GetTotalMessageCount() != 0 {
		t.Error("Messages should be cleared")
	}
}

func TestClearChat(t *testing.T) {
	cache := New(100)
	chatID := int64(12345)

	// Add messages to chat
	for i := int64(1); i <= 3; i++ {
		cache.SetMessage(chatID, &types.Message{
			ID:      i,
			ChatID:  chatID,
			Content: types.MessageContent{Text: "Test"},
			Date:    time.Now(),
		})
	}

	cache.ClearChat(chatID)

	messages := cache.GetMessages(chatID)
	if len(messages) != 0 {
		t.Errorf("Expected 0 messages after clear, got %d", len(messages))
	}
}

func TestGetMessageCount(t *testing.T) {
	cache := New(100)
	chatID := int64(12345)

	// Add 5 messages
	for i := int64(1); i <= 5; i++ {
		cache.SetMessage(chatID, &types.Message{
			ID:      i,
			ChatID:  chatID,
			Content: types.MessageContent{Text: "Test"},
			Date:    time.Now(),
		})
	}

	count := cache.GetMessageCount(chatID)
	if count != 5 {
		t.Errorf("Expected 5 messages, got %d", count)
	}
}

func TestGetTotalMessageCount(t *testing.T) {
	cache := New(100)

	// Add messages to multiple chats
	for chatID := int64(1); chatID <= 3; chatID++ {
		for msgID := int64(1); msgID <= 2; msgID++ {
			cache.SetMessage(chatID, &types.Message{
				ID:      msgID,
				ChatID:  chatID,
				Content: types.MessageContent{Text: "Test"},
				Date:    time.Now(),
			})
		}
	}

	total := cache.GetTotalMessageCount()
	if total != 6 {
		t.Errorf("Expected 6 total messages, got %d", total)
	}
}

func TestGetChatCount(t *testing.T) {
	cache := New(100)

	for i := int64(1); i <= 5; i++ {
		cache.SetChat(&types.Chat{
			ID:    i,
			Title: "Chat",
			Type:  types.ChatTypePrivate,
		})
	}

	count := cache.GetChatCount()
	if count != 5 {
		t.Errorf("Expected 5 chats, got %d", count)
	}
}

func TestGetUserCount(t *testing.T) {
	cache := New(100)

	for i := int64(1); i <= 3; i++ {
		cache.SetUser(&types.User{
			ID:        i,
			FirstName: "User",
			Username:  "user",
		})
	}

	count := cache.GetUserCount()
	if count != 3 {
		t.Errorf("Expected 3 users, got %d", count)
	}
}

func TestMessageLimit(t *testing.T) {
	limit := 10
	cache := New(limit)
	chatID := int64(12345)

	// Add more messages than the limit
	for i := int64(1); i <= 15; i++ {
		cache.SetMessage(chatID, &types.Message{
			ID:      i,
			ChatID:  chatID,
			Content: types.MessageContent{Text: "Test"},
			Date:    time.Now(),
		})
	}

	count := cache.GetMessageCount(chatID)
	if count > limit {
		t.Errorf("Expected at most %d messages due to limit, got %d", limit, count)
	}
}

func TestConcurrentAccess(t *testing.T) {
	cache := New(100)
	chatID := int64(12345)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			cache.SetMessage(chatID, &types.Message{
				ID:      id,
				ChatID:  chatID,
				Content: types.MessageContent{Text: "Test"},
				Date:    time.Now(),
			})
		}(int64(i))
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			cache.GetMessage(chatID, id)
		}(int64(i))
	}

	wg.Wait()

	// Just verify no race conditions occurred
	// (if there were race conditions, the test would fail with -race flag)
	count := cache.GetMessageCount(chatID)
	if count == 0 {
		t.Error("Expected some messages to be cached")
	}
}
