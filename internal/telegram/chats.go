// Package telegram provides a wrapper around the gotd Telegram client.
package telegram

import (
	"fmt"
	"time"

	"github.com/gotd/td/tg"
	"github.com/lvcasx1/ithil/pkg/types"
)

// GetChats retrieves the list of chats.
func (c *Client) GetChats(limit int) ([]*types.Chat, error) {
	c.logger.Info("Getting chats", "limit", limit)

	if limit <= 0 {
		limit = 100
	}

	// Get dialogs (chats) from Telegram
	result, err := c.api.MessagesGetDialogs(c.ctx, &tg.MessagesGetDialogsRequest{
		OffsetDate: 0,
		OffsetID:   0,
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      limit,
		Hash:       0,
	})

	if err != nil {
		c.logger.Error("Failed to get chats", "error", err)
		return nil, fmt.Errorf("failed to get chats: %w", err)
	}

	var chats []*types.Chat

	// Process the response
	switch dialogs := result.(type) {
	case *tg.MessagesDialogs:
		chats = c.convertDialogs(dialogs.Dialogs, dialogs.Chats, dialogs.Users, dialogs.Messages)
	case *tg.MessagesDialogsSlice:
		chats = c.convertDialogs(dialogs.Dialogs, dialogs.Chats, dialogs.Users, dialogs.Messages)
	case *tg.MessagesDialogsNotModified:
		c.logger.Info("Dialogs not modified")
		return []*types.Chat{}, nil
	default:
		return nil, fmt.Errorf("unexpected dialogs type: %T", result)
	}

	c.logger.Info("Retrieved chats", "count", len(chats))
	return chats, nil
}

// convertDialogs converts Telegram dialogs to internal Chat type
func (c *Client) convertDialogs(dialogs []tg.DialogClass, chats []tg.ChatClass, users []tg.UserClass, messages []tg.MessageClass) []*types.Chat {
	result := make([]*types.Chat, 0, len(dialogs))

	// Create maps for quick lookup
	chatMap := make(map[int64]tg.ChatClass)
	userMap := make(map[int64]tg.UserClass)
	messageMap := make(map[int64]tg.MessageClass)

	for _, chat := range chats {
		switch c := chat.(type) {
		case *tg.Chat:
			chatMap[int64(c.ID)] = c
		case *tg.ChatForbidden:
			chatMap[int64(c.ID)] = c
		case *tg.Channel:
			chatMap[c.ID] = c
		case *tg.ChannelForbidden:
			chatMap[c.ID] = c
		}
	}

	for _, user := range users {
		if u, ok := user.(*tg.User); ok {
			userMap[u.ID] = u

			// Cache the user for later lookup
			if c.updateHandler != nil && c.updateHandler.cache != nil {
				convertedUser := c.updateHandler.convertUser(u)
				if convertedUser != nil {
					c.updateHandler.cache.SetUser(convertedUser)
					c.logger.Debug("Cached user from dialogs", "id", convertedUser.ID, "username", convertedUser.Username, "firstName", convertedUser.FirstName)
				}
			}
		}
	}

	for _, msg := range messages {
		if m, ok := msg.(*tg.Message); ok {
			messageMap[int64(m.ID)] = m
		}
	}

	// Process each dialog
	for _, dialog := range dialogs {
		d, ok := dialog.(*tg.Dialog)
		if !ok {
			continue
		}

		chat := &types.Chat{
			UnreadCount: d.UnreadCount,
			IsPinned:    d.Pinned,
			IsMuted:     false, // Will be set based on notify settings
		}

		// Get peer information
		switch peer := d.Peer.(type) {
		case *tg.PeerUser:
			if user, ok := userMap[peer.UserID].(*tg.User); ok {
				chat.ID = user.ID
				chat.Title = user.FirstName
				if user.LastName != "" {
					chat.Title += " " + user.LastName
				}
				chat.Type = types.ChatTypePrivate
				chat.Username = user.Username
				chat.AccessHash = user.AccessHash
			}
		case *tg.PeerChat:
			if chatData, ok := chatMap[int64(peer.ChatID)].(*tg.Chat); ok {
				chat.ID = int64(chatData.ID)
				chat.Title = chatData.Title
				chat.Type = types.ChatTypeGroup
				// Regular chats don't have access hash
				chat.AccessHash = 0
			}
		case *tg.PeerChannel:
			if channel, ok := chatMap[peer.ChannelID].(*tg.Channel); ok {
				chat.ID = channel.ID
				chat.Title = channel.Title
				chat.Username = channel.Username
				chat.AccessHash = channel.AccessHash
				if channel.Broadcast {
					chat.Type = types.ChatTypeChannel
				} else {
					chat.Type = types.ChatTypeSupergroup
				}
			}
		}

		// Get last message info
		if msg, ok := messageMap[int64(d.TopMessage)].(*tg.Message); ok {
			// Create a simple Message object for the last message
			chat.LastMessage = &types.Message{
				ID:      int64(msg.ID),
				ChatID:  chat.ID,
				Content: types.MessageContent{
					Type: types.MessageTypeText,
					Text: msg.Message,
				},
				Date: c.convertDate(msg.Date),
			}
		}

		result = append(result, chat)
	}

	return result
}

// convertDate converts Unix timestamp to time.Time
func (c *Client) convertDate(unixTime int) time.Time {
	return time.Unix(int64(unixTime), 0)
}

// GetChat retrieves a specific chat.
func (c *Client) GetChat(chatID int64) (*types.Chat, error) {
	c.logger.Info("Getting chat", "chatID", chatID)

	// TODO: Implement chat retrieval with gotd API
	// Use c.api.MessagesGetChats() or c.api.ChannelsGetChannels() with appropriate parameters

	return nil, nil
}

// SearchChats searches for chats.
func (c *Client) SearchChats(query string, limit int) ([]*types.Chat, error) {
	c.logger.Info("Searching chats", "query", query, "limit", limit)

	// TODO: Implement chat search with gotd API
	// Use c.api.ContactsSearch() with appropriate parameters

	return []*types.Chat{}, nil
}

// PinChat pins a chat.
func (c *Client) PinChat(chatID int64, pin bool) error {
	c.logger.Info("Pinning chat", "chatID", chatID, "pin", pin)

	// TODO: Implement with gotd API
	// Use c.api.MessagesToggleDialogPin() with appropriate parameters

	return nil
}

// MuteChat mutes or unmutes a chat.
func (c *Client) MuteChat(chatID int64, mute bool) error {
	c.logger.Info("Muting chat", "chatID", chatID, "mute", mute)

	// TODO: Implement with gotd API
	// Use c.api.AccountUpdateNotifySettings() with appropriate parameters

	return nil
}

// ArchiveChat archives or unarchives a chat.
func (c *Client) ArchiveChat(chatID int64, archive bool) error {
	c.logger.Info("Archiving chat", "chatID", chatID, "archive", archive)

	// TODO: Implement with gotd API
	// Use c.api.FoldersEditPeerFolders() with appropriate parameters

	return nil
}

// DeleteChat deletes a chat.
func (c *Client) DeleteChat(chatID int64) error {
	c.logger.Info("Deleting chat", "chatID", chatID)

	// TODO: Implement with gotd API
	// Use c.api.MessagesDeleteHistory() or c.api.ChannelsDeleteChannel() with appropriate parameters

	return nil
}

// GetUser retrieves user information.
func (c *Client) GetUser(userID int64) (*types.User, error) {
	c.logger.Info("Getting user", "userID", userID)

	// TODO: Implement user retrieval with gotd API
	// Use c.api.UsersGetUsers() with appropriate parameters

	return nil, nil
}
