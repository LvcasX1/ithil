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

	// Track pin order - Telegram returns pinned chats in their correct order
	pinOrderCounter := 1

	// Process each dialog
	for _, dialog := range dialogs {
		d, ok := dialog.(*tg.Dialog)
		if !ok {
			continue
		}

		chat := &types.Chat{
			UnreadCount: d.UnreadCount,
			IsPinned:    d.Pinned,
			PinOrder:    0, // Will be set below if pinned
			IsMuted:     false, // Will be set based on notify settings
		}

		// Assign pin order to pinned chats (maintains Telegram's original order)
		if d.Pinned {
			chat.PinOrder = pinOrderCounter
			pinOrderCounter++
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
				ID:     int64(msg.ID),
				ChatID: chat.ID,
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

// PinChat pins or unpins a chat.
func (c *Client) PinChat(chat *types.Chat, pin bool) error {
	c.logger.Info("Pinning chat", "chatID", chat.ID, "pin", pin)

	// Convert chat to InputPeer
	inputPeer, err := c.chatToInputPeer(chat)
	if err != nil {
		c.logger.Error("Failed to convert chat to InputPeer", "error", err)
		return fmt.Errorf("failed to convert chat to InputPeer: %w", err)
	}

	// Create InputDialogPeer
	dialogPeer := &tg.InputDialogPeer{
		Peer: inputPeer,
	}

	// Toggle dialog pin
	_, err = c.api.MessagesToggleDialogPin(c.ctx, &tg.MessagesToggleDialogPinRequest{
		Pinned: pin,
		Peer:   dialogPeer,
	})

	if err != nil {
		c.logger.Error("Failed to toggle dialog pin", "error", err)
		return fmt.Errorf("failed to toggle dialog pin: %w", err)
	}

	c.logger.Info("Successfully toggled dialog pin", "chatID", chat.ID, "pinned", pin)
	return nil
}

// MuteChat mutes or unmutes a chat.
func (c *Client) MuteChat(chat *types.Chat, mute bool) error {
	c.logger.Info("Muting chat", "chatID", chat.ID, "mute", mute)

	// Convert chat to InputPeer
	inputPeer, err := c.chatToInputPeer(chat)
	if err != nil {
		c.logger.Error("Failed to convert chat to InputPeer", "error", err)
		return fmt.Errorf("failed to convert chat to InputPeer: %w", err)
	}

	// Create InputNotifyPeer
	notifyPeer := &tg.InputNotifyPeer{
		Peer: inputPeer,
	}

	// Create notification settings
	settings := tg.InputPeerNotifySettings{}
	if mute {
		// Mute indefinitely by setting MuteUntil to max int32 value
		settings.SetMuteUntil(2147483647) // Unix timestamp far in the future
	} else {
		// Unmute by setting MuteUntil to 0
		settings.SetMuteUntil(0)
	}

	// Update notification settings
	_, err = c.api.AccountUpdateNotifySettings(c.ctx, &tg.AccountUpdateNotifySettingsRequest{
		Peer:     notifyPeer,
		Settings: settings,
	})

	if err != nil {
		c.logger.Error("Failed to update notification settings", "error", err)
		return fmt.Errorf("failed to update notification settings: %w", err)
	}

	c.logger.Info("Successfully updated notification settings", "chatID", chat.ID, "muted", mute)
	return nil
}

// ArchiveChat archives or unarchives a chat.
func (c *Client) ArchiveChat(chat *types.Chat, archive bool) error {
	c.logger.Info("Archiving chat", "chatID", chat.ID, "archive", archive)

	// Convert chat to InputPeer
	inputPeer, err := c.chatToInputPeer(chat)
	if err != nil {
		c.logger.Error("Failed to convert chat to InputPeer", "error", err)
		return fmt.Errorf("failed to convert chat to InputPeer: %w", err)
	}

	// Create InputFolderPeer
	// Folder ID 1 is the archive folder, 0 means no folder (unarchive)
	folderID := 0
	if archive {
		folderID = 1
	}

	folderPeer := tg.InputFolderPeer{
		Peer:     inputPeer,
		FolderID: folderID,
	}

	// Edit peer folders
	_, err = c.api.FoldersEditPeerFolders(c.ctx, []tg.InputFolderPeer{folderPeer})

	if err != nil {
		c.logger.Error("Failed to edit peer folders", "error", err)
		return fmt.Errorf("failed to edit peer folders: %w", err)
	}

	c.logger.Info("Successfully edited peer folders", "chatID", chat.ID, "archived", archive)
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
