// Package cache provides caching functionality for messages and media.
package cache

import (
	"sync"

	"github.com/lvcasx1/ithil/pkg/types"
)

// Cache manages cached data for the application.
type Cache struct {
	mu               sync.RWMutex
	messages         map[int64]map[int64]*types.Message // chatID -> messageID -> Message
	chats            map[int64]*types.Chat              // chatID -> Chat
	users            map[int64]*types.User              // userID -> User
	maxMessagesPerChat int
}

// New creates a new cache instance.
func New(maxMessagesPerChat int) *Cache {
	return &Cache{
		messages:         make(map[int64]map[int64]*types.Message),
		chats:            make(map[int64]*types.Chat),
		users:            make(map[int64]*types.User),
		maxMessagesPerChat: maxMessagesPerChat,
	}
}

// GetMessage retrieves a message from the cache.
func (c *Cache) GetMessage(chatID, messageID int64) (*types.Message, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	chatMessages, exists := c.messages[chatID]
	if !exists {
		return nil, false
	}

	msg, exists := chatMessages[messageID]
	return msg, exists
}

// SetMessage stores a message in the cache.
func (c *Cache) SetMessage(chatID int64, message *types.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.messages[chatID]; !exists {
		c.messages[chatID] = make(map[int64]*types.Message)
	}

	c.messages[chatID][message.ID] = message

	// Trim old messages if limit exceeded
	if len(c.messages[chatID]) > c.maxMessagesPerChat {
		c.trimMessages(chatID)
	}
}

// AddMessage is an alias for SetMessage for convenience.
func (c *Cache) AddMessage(chatID int64, message *types.Message) {
	c.SetMessage(chatID, message)
}

// GetMessages retrieves all messages for a chat.
func (c *Cache) GetMessages(chatID int64) []*types.Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	chatMessages, exists := c.messages[chatID]
	if !exists {
		return []*types.Message{}
	}

	messages := make([]*types.Message, 0, len(chatMessages))
	for _, msg := range chatMessages {
		messages = append(messages, msg)
	}

	return messages
}

// DeleteMessage removes a message from the cache.
func (c *Cache) DeleteMessage(chatID, messageID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if chatMessages, exists := c.messages[chatID]; exists {
		delete(chatMessages, messageID)
	}
}

// GetChat retrieves a chat from the cache.
func (c *Cache) GetChat(chatID int64) (*types.Chat, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	chat, exists := c.chats[chatID]
	return chat, exists
}

// SetChat stores a chat in the cache.
func (c *Cache) SetChat(chat *types.Chat) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.chats[chat.ID] = chat
}

// GetChats retrieves all chats from the cache.
func (c *Cache) GetChats() []*types.Chat {
	c.mu.RLock()
	defer c.mu.RUnlock()

	chats := make([]*types.Chat, 0, len(c.chats))
	for _, chat := range c.chats {
		chats = append(chats, chat)
	}

	return chats
}

// DeleteChat removes a chat from the cache.
func (c *Cache) DeleteChat(chatID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.chats, chatID)
	delete(c.messages, chatID)
}

// GetUser retrieves a user from the cache.
func (c *Cache) GetUser(userID int64) (*types.User, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	user, exists := c.users[userID]
	return user, exists
}

// SetUser stores a user in the cache.
func (c *Cache) SetUser(user *types.User) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.users[user.ID] = user
}

// GetUsers retrieves all users from the cache.
func (c *Cache) GetUsers() []*types.User {
	c.mu.RLock()
	defer c.mu.RUnlock()

	users := make([]*types.User, 0, len(c.users))
	for _, user := range c.users {
		users = append(users, user)
	}

	return users
}

// Clear clears all cached data.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.messages = make(map[int64]map[int64]*types.Message)
	c.chats = make(map[int64]*types.Chat)
	c.users = make(map[int64]*types.User)
}

// ClearChat clears all messages for a specific chat.
func (c *Cache) ClearChat(chatID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.messages, chatID)
}

// trimMessages removes the oldest messages when the limit is exceeded.
// This is a simple implementation that removes messages randomly.
// A better implementation would remove the oldest messages.
func (c *Cache) trimMessages(chatID int64) {
	chatMessages := c.messages[chatID]
	if len(chatMessages) <= c.maxMessagesPerChat {
		return
	}

	// Remove excess messages (simple approach - remove oldest by ID)
	// In a real implementation, we'd want to keep track of message order
	toRemove := len(chatMessages) - c.maxMessagesPerChat

	for msgID := range chatMessages {
		if toRemove <= 0 {
			break
		}
		delete(chatMessages, msgID)
		toRemove--
	}
}

// GetMessageCount returns the number of cached messages for a chat.
func (c *Cache) GetMessageCount(chatID int64) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if chatMessages, exists := c.messages[chatID]; exists {
		return len(chatMessages)
	}
	return 0
}

// GetTotalMessageCount returns the total number of cached messages.
func (c *Cache) GetTotalMessageCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, chatMessages := range c.messages {
		count += len(chatMessages)
	}
	return count
}

// GetChatCount returns the number of cached chats.
func (c *Cache) GetChatCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.chats)
}

// GetUserCount returns the number of cached users.
func (c *Cache) GetUserCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.users)
}
