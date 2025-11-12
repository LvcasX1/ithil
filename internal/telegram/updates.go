// Package telegram provides a wrapper around the gotd Telegram client.
package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gotd/td/tg"

	"github.com/lvcasx1/ithil/pkg/types"
)

// UpdateHandler handles incoming updates from Telegram.
type UpdateHandler struct {
	client *Client
	cache  CacheInterface
	logger *slog.Logger
}

// CacheInterface defines the cache methods needed by UpdateHandler.
type CacheInterface interface {
	SetUser(user *types.User)
	AddMessage(chatID int64, message *types.Message)
}

// NewUpdateHandler creates a new update handler.
// cache can be nil initially and set later via SetCache.
func NewUpdateHandler(client *Client, cache CacheInterface, logger *slog.Logger) *UpdateHandler {
	return &UpdateHandler{
		client: client,
		cache:  cache,
		logger: logger,
	}
}

// SetCache sets the cache for the update handler.
// This is useful when the cache is created after the client.
func (h *UpdateHandler) SetCache(cache CacheInterface) {
	h.cache = cache
}

// Handle processes incoming updates from gotd.
func (h *UpdateHandler) Handle(ctx context.Context, updates tg.UpdatesClass) error {
	h.logger.Info("UPDATE RECEIVED", "type", fmt.Sprintf("%T", updates))

	switch u := updates.(type) {
	case *tg.Updates:
		// Cache users from this update batch
		h.cacheUsers(u.Users)

		for _, update := range u.Updates {
			h.processUpdate(ctx, update)
		}
	case *tg.UpdatesCombined:
		// Cache users from this update batch
		h.cacheUsers(u.Users)

		for _, update := range u.Updates {
			h.processUpdate(ctx, update)
		}
	case *tg.UpdateShort:
		h.processUpdate(ctx, u.Update)
	case *tg.UpdateShortMessage:
		h.handleShortMessage(ctx, u)
	case *tg.UpdateShortChatMessage:
		h.handleShortChatMessage(ctx, u)
	case *tg.UpdateShortSentMessage:
		// Message we sent was delivered
		h.logger.Debug("Message sent", "id", u.ID)
	case *tg.UpdatesTooLong:
		h.logger.Warn("Updates too long, need to refetch state")
	default:
		h.logger.Debug("Unhandled update type", "type", updates.TypeName())
	}

	return nil
}

// processUpdate processes a single update.
func (h *UpdateHandler) processUpdate(ctx context.Context, update tg.UpdateClass) {
	switch u := update.(type) {
	case *tg.UpdateNewMessage:
		h.handleNewMessage(ctx, u.Message)
	case *tg.UpdateEditMessage:
		h.handleEditMessage(ctx, u.Message)
	case *tg.UpdateDeleteMessages:
		h.handleDeleteMessages(ctx, u.Messages)
	case *tg.UpdateReadHistoryInbox:
		h.handleReadHistoryInbox(ctx, u)
	case *tg.UpdateReadHistoryOutbox:
		h.handleReadHistoryOutbox(ctx, u)
	case *tg.UpdateUserStatus:
		h.handleUserStatus(ctx, u)
	case *tg.UpdateUserTyping:
		h.handleUserTyping(ctx, u)
	case *tg.UpdateChatUserTyping:
		h.handleChatUserTyping(ctx, u)
	case *tg.UpdateNewChannelMessage:
		h.handleNewMessage(ctx, u.Message)
	case *tg.UpdateEditChannelMessage:
		h.handleEditMessage(ctx, u.Message)
	default:
		h.logger.Debug("Unhandled update", "type", update.TypeName())
	}
}

// handleNewMessage handles new message updates.
func (h *UpdateHandler) handleNewMessage(ctx context.Context, msg tg.MessageClass) {
	message := h.convertMessage(msg)
	if message == nil {
		return
	}

	h.logger.Info("New message update received", "chatID", message.ChatID, "text", message.Content.Text[:min(50, len(message.Content.Text))])

	h.sendUpdate(&types.Update{
		Type:    types.UpdateTypeNewMessage,
		ChatID:  message.ChatID,
		Message: message,
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// handleEditMessage handles message edit updates.
func (h *UpdateHandler) handleEditMessage(ctx context.Context, msg tg.MessageClass) {
	message := h.convertMessage(msg)
	if message == nil {
		return
	}

	h.sendUpdate(&types.Update{
		Type:    types.UpdateTypeMessageEdited,
		ChatID:  message.ChatID,
		Message: message,
	})
}

// handleDeleteMessages handles message deletion updates.
func (h *UpdateHandler) handleDeleteMessages(ctx context.Context, messageIDs []int) {
	for _, msgID := range messageIDs {
		h.sendUpdate(&types.Update{
			Type: types.UpdateTypeMessageDeleted,
			Data: int64(msgID),
		})
	}
}

// handleShortMessage handles short message updates (direct messages).
func (h *UpdateHandler) handleShortMessage(ctx context.Context, u *tg.UpdateShortMessage) {
	message := &types.Message{
		ID:         int64(u.ID),
		ChatID:     u.UserID,
		SenderID:   u.UserID,
		Date:       time.Unix(int64(u.Date), 0),
		IsOutgoing: u.Out,
		Content: types.MessageContent{
			Type: types.MessageTypeText,
			Text: u.Message,
		},
	}

	h.sendUpdate(&types.Update{
		Type:    types.UpdateTypeNewMessage,
		ChatID:  u.UserID,
		Message: message,
	})
}

// handleShortChatMessage handles short chat message updates.
func (h *UpdateHandler) handleShortChatMessage(ctx context.Context, u *tg.UpdateShortChatMessage) {
	chatID := int64(u.ChatID)

	message := &types.Message{
		ID:         int64(u.ID),
		ChatID:     chatID,
		SenderID:   u.FromID,
		Date:       time.Unix(int64(u.Date), 0),
		IsOutgoing: u.Out,
		Content: types.MessageContent{
			Type: types.MessageTypeText,
			Text: u.Message,
		},
	}

	h.sendUpdate(&types.Update{
		Type:    types.UpdateTypeNewMessage,
		ChatID:  chatID,
		Message: message,
	})
}

// handleReadHistoryInbox handles read inbox history updates.
func (h *UpdateHandler) handleReadHistoryInbox(ctx context.Context, u *tg.UpdateReadHistoryInbox) {
	var chatID int64
	switch peer := u.Peer.(type) {
	case *tg.PeerUser:
		chatID = peer.UserID
	case *tg.PeerChat:
		chatID = peer.ChatID
	case *tg.PeerChannel:
		chatID = peer.ChannelID
	}

	h.sendUpdate(&types.Update{
		Type:   types.UpdateTypeChatReadInbox,
		ChatID: chatID,
		Data:   int64(u.MaxID),
	})
}

// handleReadHistoryOutbox handles read outbox history updates.
func (h *UpdateHandler) handleReadHistoryOutbox(ctx context.Context, u *tg.UpdateReadHistoryOutbox) {
	var chatID int64
	switch peer := u.Peer.(type) {
	case *tg.PeerUser:
		chatID = peer.UserID
	case *tg.PeerChat:
		chatID = peer.ChatID
	case *tg.PeerChannel:
		chatID = peer.ChannelID
	}

	h.sendUpdate(&types.Update{
		Type:   types.UpdateTypeChatReadOutbox,
		ChatID: chatID,
		Data:   int64(u.MaxID),
	})
}

// handleUserStatus handles user status updates.
func (h *UpdateHandler) handleUserStatus(ctx context.Context, u *tg.UpdateUserStatus) {
	h.sendUpdate(&types.Update{
		Type: types.UpdateTypeUserStatus,
		Data: map[string]interface{}{
			"user_id": u.UserID,
			"status":  u.Status,
		},
	})
}

// handleUserTyping handles user typing updates.
func (h *UpdateHandler) handleUserTyping(ctx context.Context, u *tg.UpdateUserTyping) {
	h.sendUpdate(&types.Update{
		Type:   types.UpdateTypeUserStatus,
		ChatID: u.UserID,
		Data: map[string]interface{}{
			"user_id": u.UserID,
			"typing":  true,
		},
	})
}

// handleChatUserTyping handles chat user typing updates.
func (h *UpdateHandler) handleChatUserTyping(ctx context.Context, u *tg.UpdateChatUserTyping) {
	chatID := int64(u.ChatID)

	// Extract user ID from peer
	var userID int64
	switch peer := u.FromID.(type) {
	case *tg.PeerUser:
		userID = peer.UserID
	case *tg.PeerChat:
		userID = peer.ChatID
	case *tg.PeerChannel:
		userID = peer.ChannelID
	}

	h.sendUpdate(&types.Update{
		Type:   types.UpdateTypeUserStatus,
		ChatID: chatID,
		Data: map[string]interface{}{
			"user_id": userID,
			"typing":  true,
		},
	})
}

// convertMessage converts a gotd message to our internal message type.
func (h *UpdateHandler) convertMessage(msg tg.MessageClass) *types.Message {
	switch m := msg.(type) {
	case *tg.Message:
		message := &types.Message{
			ID:         int64(m.ID),
			Date:       time.Unix(int64(m.Date), 0),
			IsOutgoing: m.Out,
			IsPinned:   m.Pinned,
		}

		// Extract chat ID
		switch peer := m.PeerID.(type) {
		case *tg.PeerUser:
			message.ChatID = peer.UserID
		case *tg.PeerChat:
			message.ChatID = peer.ChatID
		case *tg.PeerChannel:
			message.ChatID = peer.ChannelID
		}

		// Extract sender ID
		if m.FromID != nil {
			switch from := m.FromID.(type) {
			case *tg.PeerUser:
				message.SenderID = from.UserID
			case *tg.PeerChat:
				message.SenderID = from.ChatID
			case *tg.PeerChannel:
				message.SenderID = from.ChannelID
			}
		}

		// Extract message content
		message.Content = h.convertMessageContent(m)

		// Extract reply info
		if m.ReplyTo != nil {
			if replyTo, ok := m.ReplyTo.(*tg.MessageReplyHeader); ok {
				if replyTo.ReplyToMsgID != 0 {
					message.ReplyToMessageID = int64(replyTo.ReplyToMsgID)
				}
			}
		}

		// Extract edit date
		if m.EditDate != 0 {
			message.EditDate = time.Unix(int64(m.EditDate), 0)
			message.IsEdited = true
		}

		// Extract views
		if m.Views != 0 {
			message.Views = m.Views
		}

		return message

	case *tg.MessageService:
		// Service messages (user joined, left, etc.)
		// We could handle these specially, but for now just log
		h.logger.Debug("Service message", "action", m.Action.TypeName())
		return nil

	case *tg.MessageEmpty:
		return nil

	default:
		h.logger.Debug("Unknown message type", "type", msg.TypeName())
		return nil
	}
}

// convertMessageContent converts gotd message media to our internal content type.
func (h *UpdateHandler) convertMessageContent(msg *tg.Message) types.MessageContent {
	content := types.MessageContent{
		Type: types.MessageTypeText,
		Text: msg.Message,
	}

	// Handle media
	if msg.Media != nil {
		switch media := msg.Media.(type) {
		case *tg.MessageMediaPhoto:
			content.Type = types.MessageTypePhoto
			// TODO: Extract photo details
		case *tg.MessageMediaDocument:
			content.Type = types.MessageTypeDocument
			// TODO: Extract document details
		case *tg.MessageMediaGeo:
			content.Type = types.MessageTypeLocation
			// TODO: Extract location details
		case *tg.MessageMediaContact:
			content.Type = types.MessageTypeContact
			// TODO: Extract contact details
		case *tg.MessageMediaPoll:
			content.Type = types.MessageTypePoll
			// TODO: Extract poll details
		default:
			h.logger.Debug("Unhandled media type", "type", media.TypeName())
		}
	}

	// TODO: Handle entities (bold, italic, etc.)

	return content
}

// sendUpdate sends an update to the updates channel.
func (h *UpdateHandler) sendUpdate(update *types.Update) {
	select {
	case h.client.updates <- update:
		// Update sent successfully
	default:
		h.logger.Warn("Update channel is full, dropping update", "type", update.Type)
	}
}

// cacheUsers caches users from an Updates message.
func (h *UpdateHandler) cacheUsers(users []tg.UserClass) {
	if h.cache == nil {
		return
	}

	for _, userClass := range users {
		if tgUser, ok := userClass.(*tg.User); ok {
			user := h.convertUser(tgUser)
			if user != nil {
				h.cache.SetUser(user)
				h.logger.Debug("Cached user", "id", user.ID, "username", user.Username, "firstName", user.FirstName)
			}
		}
	}
}

// convertUser converts a gotd User to our internal User type.
func (h *UpdateHandler) convertUser(tgUser *tg.User) *types.User {
	if tgUser == nil {
		return nil
	}

	user := &types.User{
		ID:         tgUser.ID,
		FirstName:  tgUser.FirstName,
		LastName:   tgUser.LastName,
		Username:   tgUser.Username,
		IsBot:      tgUser.Bot,
		IsVerified: tgUser.Verified,
		IsPremium:  tgUser.Premium,
	}

	// Convert user status
	if tgUser.Status != nil {
		switch tgUser.Status.(type) {
		case *tg.UserStatusOnline:
			user.Status = types.UserStatusOnline
		case *tg.UserStatusOffline:
			user.Status = types.UserStatusOffline
		case *tg.UserStatusRecently:
			user.Status = types.UserStatusRecently
		case *tg.UserStatusLastWeek:
			user.Status = types.UserStatusLastWeek
		case *tg.UserStatusLastMonth:
			user.Status = types.UserStatusLastMonth
		default:
			user.Status = types.UserStatusOffline
		}
	}

	return user
}
