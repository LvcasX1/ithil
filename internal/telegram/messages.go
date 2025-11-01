// Package telegram provides a wrapper around the gotd Telegram client.
package telegram

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/gotd/td/tg"
	"github.com/lvcasx1/ithil/pkg/types"
)

// GetMessages retrieves message history for a chat.
func (c *Client) GetMessages(chat *types.Chat, limit int, offset int) ([]*types.Message, error) {
	c.logger.Info("Getting messages", "chatID", chat.ID, "limit", limit, "offset", offset)

	if limit <= 0 {
		limit = 50
	}

	// Convert chat to InputPeer
	inputPeer, err := c.chatToInputPeer(chat)
	if err != nil {
		c.logger.Error("Failed to convert chat to InputPeer", "error", err)
		return nil, fmt.Errorf("failed to convert chat to InputPeer: %w", err)
	}

	// Get message history from Telegram
	history, err := c.api.MessagesGetHistory(c.ctx, &tg.MessagesGetHistoryRequest{
		Peer:       inputPeer,
		OffsetID:   offset,
		OffsetDate: 0,
		AddOffset:  0,
		Limit:      limit,
		MaxID:      0,
		MinID:      0,
		Hash:       0,
	})

	if err != nil {
		c.logger.Error("Failed to get message history", "error", err)
		return nil, fmt.Errorf("failed to get message history: %w", err)
	}

	// Process the response based on type
	var messages []*types.Message
	switch h := history.(type) {
	case *tg.MessagesMessages:
		messages = c.convertMessages(h.Messages, h.Users, chat.ID)
	case *tg.MessagesMessagesSlice:
		messages = c.convertMessages(h.Messages, h.Users, chat.ID)
	case *tg.MessagesChannelMessages:
		messages = c.convertMessages(h.Messages, h.Users, chat.ID)
	case *tg.MessagesMessagesNotModified:
		c.logger.Info("Messages not modified")
		return []*types.Message{}, nil
	default:
		return nil, fmt.Errorf("unexpected message history type: %T", history)
	}

	c.logger.Info("Retrieved messages", "count", len(messages))
	return messages, nil
}

// convertMessages converts Telegram messages to internal Message type
func (c *Client) convertMessages(messages []tg.MessageClass, users []tg.UserClass, chatID int64) []*types.Message {
	result := make([]*types.Message, 0, len(messages))

	// Create user map for quick lookup
	userMap := make(map[int64]*tg.User)
	for _, user := range users {
		if u, ok := user.(*tg.User); ok {
			userMap[u.ID] = u
		}
	}

	// Cache all users from the message history
	// This ensures that when rendering messages, we have user info available
	if c.updateHandler != nil && c.updateHandler.cache != nil {
		for _, user := range users {
			if tgUser, ok := user.(*tg.User); ok {
				convertedUser := c.updateHandler.convertUser(tgUser)
				if convertedUser != nil {
					c.updateHandler.cache.SetUser(convertedUser)
					c.logger.Debug("Cached user from message history", "id", convertedUser.ID, "username", convertedUser.Username, "firstName", convertedUser.FirstName)
				}
			}
		}
	}

	// Convert each message
	for _, msg := range messages {
		if m, ok := msg.(*tg.Message); ok {
			convertedMsg := c.convertMessage(m, userMap, chatID)
			if convertedMsg != nil {
				result = append(result, convertedMsg)
			}
		}
	}

	// Reverse to get chronological order (oldest first)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// convertMessage converts a single Telegram message to internal Message type
func (c *Client) convertMessage(msg *tg.Message, userMap map[int64]*tg.User, chatID int64) *types.Message {
	if msg == nil {
		return nil
	}

	message := &types.Message{
		ID:            int64(msg.ID),
		ChatID:        chatID,
		Content:       c.convertMessageContent(msg),
		Date:          c.convertDate(msg.Date),
		IsOutgoing:    msg.Out,
		IsChannelPost: msg.Post,
		IsPinned:      msg.Pinned,
		IsEdited:      msg.EditDate != 0,
		Views:         msg.Views,
		MediaAlbumID:  msg.GroupedID,
	}

	// Set edit date if edited
	if msg.EditDate != 0 {
		message.EditDate = c.convertDate(msg.EditDate)
	}

	// Extract sender ID from peer
	if msg.FromID != nil {
		switch fromPeer := msg.FromID.(type) {
		case *tg.PeerUser:
			message.SenderID = fromPeer.UserID
		case *tg.PeerChannel:
			message.SenderID = fromPeer.ChannelID
		case *tg.PeerChat:
			message.SenderID = int64(fromPeer.ChatID)
		}
	}

	// Handle reply information
	if msg.ReplyTo != nil {
		if replyToMsg, ok := msg.ReplyTo.(*tg.MessageReplyHeader); ok {
			if replyToMsg.ReplyToMsgID != 0 {
				message.ReplyToMessageID = int64(replyToMsg.ReplyToMsgID)
			}
		}
	}

	// Handle forward information
	if fwd, ok := msg.GetFwdFrom(); ok {
		message.IsForwarded = true
		message.ForwardInfo = c.convertForwardInfo(&fwd)
	}

	return message
}

// convertMessageContent converts message content to internal format
func (c *Client) convertMessageContent(msg *tg.Message) types.MessageContent {
	content := types.MessageContent{
		Type: types.MessageTypeText,
		Text: msg.Message,
	}

	// Convert entities
	if len(msg.Entities) > 0 {
		content.Entities = c.convertMessageEntities(msg.Entities)
	}

	// Handle media
	if msg.Media != nil {
		switch media := msg.Media.(type) {
		case *tg.MessageMediaPhoto:
			content.Type = types.MessageTypePhoto
			if photo, ok := media.Photo.(*tg.Photo); ok {
				content.Media = c.convertPhoto(photo)
			}
			// Caption is not directly available in MessageMediaPhoto in gotd

		case *tg.MessageMediaDocument:
			doc, ok := media.Document.(*tg.Document)
			if !ok {
				break
			}

			// Determine document type from attributes
			isVideo := false
			isVoice := false
			isVideoNote := false
			isAudio := false
			isSticker := false
			isAnimation := false

			for _, attr := range doc.Attributes {
				switch attr.(type) {
				case *tg.DocumentAttributeVideo:
					if vAttr, ok := attr.(*tg.DocumentAttributeVideo); ok {
						if vAttr.RoundMessage {
							isVideoNote = true
						} else {
							isVideo = true
						}
					}
				case *tg.DocumentAttributeAudio:
					if aAttr, ok := attr.(*tg.DocumentAttributeAudio); ok {
						if aAttr.Voice {
							isVoice = true
						} else {
							isAudio = true
						}
					}
				case *tg.DocumentAttributeSticker:
					isSticker = true
				case *tg.DocumentAttributeAnimated:
					isAnimation = true
				}
			}

			// Set content type based on attributes
			switch {
			case isSticker:
				content.Type = types.MessageTypeSticker
				content.Sticker = c.convertSticker(doc)
			case isVoice:
				content.Type = types.MessageTypeVoice
				content.Media = c.convertDocument(doc)
			case isVideoNote:
				content.Type = types.MessageTypeVideoNote
				content.Media = c.convertDocument(doc)
			case isVideo:
				content.Type = types.MessageTypeVideo
				content.Media = c.convertDocument(doc)
			case isAudio:
				content.Type = types.MessageTypeAudio
				content.Media = c.convertDocument(doc)
			case isAnimation:
				content.Type = types.MessageTypeAnimation
				content.Animation = c.convertAnimation(doc)
			default:
				content.Type = types.MessageTypeDocument
				content.Document = c.convertDocumentFile(doc)
			}

		case *tg.MessageMediaGeo:
			content.Type = types.MessageTypeLocation
			if geoPoint, ok := media.Geo.(*tg.GeoPoint); ok {
				content.Location = &types.Location{
					Latitude:  geoPoint.Lat,
					Longitude: geoPoint.Long,
				}
			}

		case *tg.MessageMediaContact:
			content.Type = types.MessageTypeContact
			content.Contact = &types.Contact{
				PhoneNumber: media.PhoneNumber,
				FirstName:   media.FirstName,
				LastName:    media.LastName,
				UserID:      media.UserID,
				VCard:       media.Vcard,
			}

		case *tg.MessageMediaPoll:
			content.Type = types.MessageTypePoll
			content.Poll = c.convertPoll(media.Poll)

		case *tg.MessageMediaVenue:
			content.Type = types.MessageTypeVenue
			if geoPoint, ok := media.Geo.(*tg.GeoPoint); ok {
				content.Location = &types.Location{
					Latitude:  geoPoint.Lat,
					Longitude: geoPoint.Long,
				}
			}

		case *tg.MessageMediaGame:
			content.Type = types.MessageTypeGame
			// Game details can be added if needed
		}
	}

	return content
}

// convertMessageEntities converts Telegram entities to internal format
func (c *Client) convertMessageEntities(entities []tg.MessageEntityClass) []types.MessageEntity {
	result := make([]types.MessageEntity, 0, len(entities))

	for _, entity := range entities {
		switch e := entity.(type) {
		case *tg.MessageEntityBold:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypeBold,
				Offset: e.Offset,
				Length: e.Length,
			})
		case *tg.MessageEntityItalic:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypeItalic,
				Offset: e.Offset,
				Length: e.Length,
			})
		case *tg.MessageEntityCode:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypeCode,
				Offset: e.Offset,
				Length: e.Length,
			})
		case *tg.MessageEntityPre:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypePre,
				Offset: e.Offset,
				Length: e.Length,
			})
		case *tg.MessageEntityTextURL:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypeTextURL,
				Offset: e.Offset,
				Length: e.Length,
				URL:    e.URL,
			})
		case *tg.MessageEntityMention:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypeMention,
				Offset: e.Offset,
				Length: e.Length,
			})
		case *tg.MessageEntityHashtag:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypeHashtag,
				Offset: e.Offset,
				Length: e.Length,
			})
		case *tg.MessageEntityCashtag:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypeCashtag,
				Offset: e.Offset,
				Length: e.Length,
			})
		case *tg.MessageEntityBotCommand:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypeBotCommand,
				Offset: e.Offset,
				Length: e.Length,
			})
		case *tg.MessageEntityURL:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypeURL,
				Offset: e.Offset,
				Length: e.Length,
			})
		case *tg.MessageEntityEmail:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypeEmail,
				Offset: e.Offset,
				Length: e.Length,
			})
		case *tg.MessageEntityPhone:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypePhoneNumber,
				Offset: e.Offset,
				Length: e.Length,
			})
		case *tg.MessageEntitySpoiler:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypeSpoiler,
				Offset: e.Offset,
				Length: e.Length,
			})
		case *tg.MessageEntityStrike:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypeStrikethrough,
				Offset: e.Offset,
				Length: e.Length,
			})
		case *tg.MessageEntityUnderline:
			result = append(result, types.MessageEntity{
				Type:   types.EntityTypeUnderline,
				Offset: e.Offset,
				Length: e.Length,
			})
		}
	}

	return result
}

// chatToInputPeer converts a Chat to InputPeerClass for API calls
func (c *Client) chatToInputPeer(chat *types.Chat) (tg.InputPeerClass, error) {
	switch chat.Type {
	case types.ChatTypePrivate:
		return &tg.InputPeerUser{
			UserID:     chat.ID,
			AccessHash: chat.AccessHash,
		}, nil
	case types.ChatTypeGroup:
		return &tg.InputPeerChat{
			ChatID: chat.ID,
		}, nil
	case types.ChatTypeSupergroup, types.ChatTypeChannel:
		return &tg.InputPeerChannel{
			ChannelID:  chat.ID,
			AccessHash: chat.AccessHash,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported chat type: %v", chat.Type)
	}
}

// convertForwardInfo converts Telegram forward info to internal format
func (c *Client) convertForwardInfo(fwd *tg.MessageFwdHeader) *types.ForwardInfo {
	info := &types.ForwardInfo{
		Date:            c.convertDate(fwd.Date),
		AuthorSignature: fwd.PostAuthor,
	}

	if fwd.FromID != nil {
		switch from := fwd.FromID.(type) {
		case *tg.PeerUser:
			info.Origin = types.ForwardOriginUser
			info.FromUserID = from.UserID
		case *tg.PeerChannel:
			info.Origin = types.ForwardOriginChannel
			info.FromChatID = from.ChannelID
		case *tg.PeerChat:
			info.Origin = types.ForwardOriginChat
			info.FromChatID = int64(from.ChatID)
		}
	}

	if fwd.FromName != "" {
		info.Origin = types.ForwardOriginHiddenUser
	}

	return info
}

// convertPhoto converts Telegram photo to Media
func (c *Client) convertPhoto(photo *tg.Photo) *types.Media {
	if photo == nil {
		return nil
	}

	media := &types.Media{
		ID: fmt.Sprintf("%d", photo.ID),
	}

	// Get the largest photo size
	if len(photo.Sizes) > 0 {
		lastSize := photo.Sizes[len(photo.Sizes)-1]
		if photoSize, ok := lastSize.(*tg.PhotoSize); ok {
			media.Width = photoSize.W
			media.Height = photoSize.H
			media.Size = int64(photoSize.Size)
		}
	}

	return media
}

// convertDocument converts Telegram document to Media
func (c *Client) convertDocument(doc *tg.Document) *types.Media {
	if doc == nil {
		return nil
	}

	media := &types.Media{
		ID:       fmt.Sprintf("%d", doc.ID),
		Size:     doc.Size,
		MimeType: doc.MimeType,
	}

	// Extract dimensions and duration from attributes
	for _, attr := range doc.Attributes {
		switch a := attr.(type) {
		case *tg.DocumentAttributeVideo:
			media.Width = a.W
			media.Height = a.H
			media.Duration = int(a.Duration)
		case *tg.DocumentAttributeAudio:
			media.Duration = int(a.Duration)
		}
	}

	return media
}

// convertSticker converts Telegram document to Sticker
func (c *Client) convertSticker(doc *tg.Document) *types.Sticker {
	if doc == nil {
		return nil
	}

	sticker := &types.Sticker{
		File: c.convertDocument(doc),
	}

	// Extract sticker attributes
	for _, attr := range doc.Attributes {
		switch a := attr.(type) {
		case *tg.DocumentAttributeSticker:
			sticker.Emoji = a.Alt
		case *tg.DocumentAttributeImageSize:
			sticker.Width = a.W
			sticker.Height = a.H
		case *tg.DocumentAttributeVideo:
			sticker.IsVideo = true
		}
	}

	return sticker
}

// convertAnimation converts Telegram document to Animation
func (c *Client) convertAnimation(doc *tg.Document) *types.Animation {
	if doc == nil {
		return nil
	}

	animation := &types.Animation{
		MimeType: doc.MimeType,
		File:     c.convertDocument(doc),
	}

	// Extract animation attributes
	for _, attr := range doc.Attributes {
		switch a := attr.(type) {
		case *tg.DocumentAttributeVideo:
			animation.Width = a.W
			animation.Height = a.H
			animation.Duration = int(a.Duration)
		case *tg.DocumentAttributeFilename:
			animation.FileName = a.FileName
		}
	}

	return animation
}

// convertDocumentFile converts Telegram document to Document
func (c *Client) convertDocumentFile(doc *tg.Document) *types.Document {
	if doc == nil {
		return nil
	}

	document := &types.Document{
		MimeType: doc.MimeType,
		File:     c.convertDocument(doc),
	}

	// Extract filename
	for _, attr := range doc.Attributes {
		if a, ok := attr.(*tg.DocumentAttributeFilename); ok {
			document.FileName = a.FileName
		}
	}

	return document
}

// convertPoll converts Telegram poll to internal format
func (c *Client) convertPoll(poll tg.Poll) *types.Poll {
	// Extract question text from TextWithEntities
	questionText := poll.Question.Text

	result := &types.Poll{
		ID:              fmt.Sprintf("%d", poll.ID),
		Question:        questionText,
		TotalVoterCount: 0, // Would need PollResults to get actual counts
		IsClosed:        poll.Closed,
		Options:         make([]types.PollOption, 0, len(poll.Answers)),
	}

	for _, answer := range poll.Answers {
		result.Options = append(result.Options, types.PollOption{
			Text:       answer.Text.Text,
			VoterCount: 0, // Would need PollResults to get actual counts
		})
	}

	if poll.Quiz {
		result.Type = types.PollTypeQuiz
	} else {
		result.Type = types.PollTypeRegular
	}

	if poll.CloseDate != 0 {
		result.CloseDate = c.convertDate(poll.CloseDate)
	}

	if poll.ClosePeriod != 0 {
		result.OpenPeriod = poll.ClosePeriod
	}

	return result
}

// SendMessage sends a message to a chat.
func (c *Client) SendMessage(chat *types.Chat, text string, replyToMessageID int64) (*types.Message, error) {
	c.logger.Info("Sending message", "chatID", chat.ID, "textLength", len(text))

	// Convert chat to InputPeer
	inputPeer, err := c.chatToInputPeer(chat)
	if err != nil {
		c.logger.Error("Failed to convert chat to InputPeer", "error", err)
		return nil, fmt.Errorf("failed to convert chat to InputPeer: %w", err)
	}

	// Generate random ID for message deduplication
	randomID := c.generateRandomID()

	// Build request
	request := &tg.MessagesSendMessageRequest{
		Peer:      inputPeer,
		Message:   text,
		RandomID:  randomID,
		NoWebpage: false,
	}

	// Add reply-to information if provided
	if replyToMessageID > 0 {
		request.ReplyTo = &tg.InputReplyToMessage{
			ReplyToMsgID: int(replyToMessageID),
		}
	}

	// Send the message
	updates, err := c.api.MessagesSendMessage(c.ctx, request)
	if err != nil {
		c.logger.Error("Failed to send message", "error", err)
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Extract the sent message from updates
	message, err := c.extractMessageFromUpdates(updates, chat.ID, text)
	if err != nil {
		c.logger.Error("Failed to extract message from updates", "error", err)
		// Return a basic message even if extraction fails
		return &types.Message{
			ID:         0, // Will be updated by server
			ChatID:     chat.ID,
			Content:    types.MessageContent{Type: types.MessageTypeText, Text: text},
			Date:       time.Now(),
			IsOutgoing: true,
		}, nil
	}

	c.logger.Info("Message sent successfully", "messageID", message.ID)
	return message, nil
}

// EditMessage edits an existing message.
func (c *Client) EditMessage(chat *types.Chat, messageID int64, newText string) error {
	c.logger.Info("Editing message", "chatID", chat.ID, "messageID", messageID)

	// Convert chat to InputPeer
	inputPeer, err := c.chatToInputPeer(chat)
	if err != nil {
		c.logger.Error("Failed to convert chat to InputPeer", "error", err)
		return fmt.Errorf("failed to convert chat to InputPeer: %w", err)
	}

	// Build edit request
	request := &tg.MessagesEditMessageRequest{
		Peer:      inputPeer,
		ID:        int(messageID),
		Message:   newText,
		NoWebpage: false,
	}

	// Edit the message
	updates, err := c.api.MessagesEditMessage(c.ctx, request)
	if err != nil {
		c.logger.Error("Failed to edit message", "error", err)
		return fmt.Errorf("failed to edit message: %w", err)
	}

	// Log success
	c.logger.Info("Message edited successfully", "messageID", messageID, "updates", updates)
	return nil
}

// DeleteMessage deletes a message.
func (c *Client) DeleteMessage(chatID, messageID int64) error {
	c.logger.Info("Deleting message", "chatID", chatID, "messageID", messageID)

	// TODO: Implement message deletion with gotd API
	// Use c.api.MessagesDeleteMessages() with appropriate parameters

	return nil
}

// ForwardMessage forwards a message to another chat.
func (c *Client) ForwardMessage(fromChatID, toChatID, messageID int64) error {
	c.logger.Info("Forwarding message", "fromChatID", fromChatID, "toChatID", toChatID, "messageID", messageID)

	// TODO: Implement message forwarding with gotd API
	// Use c.api.MessagesForwardMessages() with appropriate parameters

	return nil
}

// GetMessage retrieves a specific message.
func (c *Client) GetMessage(chatID, messageID int64) (*types.Message, error) {
	c.logger.Info("Getting message", "chatID", chatID, "messageID", messageID)

	// TODO: Implement message retrieval with gotd API
	// Use c.api.MessagesGetMessages() with appropriate parameters

	return nil, nil
}

// GetChatHistory retrieves chat history.
func (c *Client) GetChatHistory(chatID int64, fromMessageID int64, limit int) ([]*types.Message, error) {
	c.logger.Info("Getting chat history", "chatID", chatID, "limit", limit)

	// TODO: Implement chat history retrieval with gotd API
	// Use c.api.MessagesGetHistory() with appropriate parameters

	return []*types.Message{}, nil
}

// MarkChatAsRead marks a chat as read.
func (c *Client) MarkChatAsRead(chat *types.Chat) error {
	c.logger.Info("Marking chat as read", "chatID", chat.ID, "title", chat.Title)

	// Convert chat to InputPeer
	inputPeer, err := c.chatToInputPeer(chat)
	if err != nil {
		c.logger.Error("Failed to convert chat to InputPeer", "error", err)
		return fmt.Errorf("failed to convert chat to InputPeer: %w", err)
	}

	// Mark messages as read using MessagesReadHistory
	// This works for all chat types (private, group, channel, supergroup)
	_, err = c.api.MessagesReadHistory(c.ctx, &tg.MessagesReadHistoryRequest{
		Peer:  inputPeer,
		MaxID: 0, // 0 means mark all as read
	})

	if err != nil {
		c.logger.Error("Failed to mark chat as read", "chatID", chat.ID, "error", err)
		return fmt.Errorf("failed to mark chat as read: %w", err)
	}

	c.logger.Info("Successfully marked chat as read", "chatID", chat.ID)
	return nil
}

// SendTyping sends a typing indicator.
func (c *Client) SendTyping(chatID int64) error {
	// TODO: Implement with gotd API
	// Use c.api.MessagesSetTyping() with appropriate parameters
	return nil
}

// generateRandomID generates a random ID for message deduplication.
func (c *Client) generateRandomID() int64 {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return time.Now().UnixNano()
	}
	return int64(binary.BigEndian.Uint64(b))
}

// extractMessageFromUpdates extracts a message from Telegram updates.
func (c *Client) extractMessageFromUpdates(updates tg.UpdatesClass, chatID int64, originalText string) (*types.Message, error) {
	switch u := updates.(type) {
	case *tg.Updates:
		// Look for the new message in updates
		for _, update := range u.Updates {
			if msg, ok := update.(*tg.UpdateNewMessage); ok {
				if m, ok := msg.Message.(*tg.Message); ok {
					// Create user map from users in the update
					userMap := make(map[int64]*tg.User)
					for _, user := range u.Users {
						if usr, ok := user.(*tg.User); ok {
							userMap[usr.ID] = usr
						}
					}
					return c.convertMessage(m, userMap, chatID), nil
				}
			}
		}
		return nil, fmt.Errorf("no message found in updates")

	case *tg.UpdateShortSentMessage:
		// For UpdateShortSentMessage, we construct a basic message
		// This update type doesn't include full message details, so we use the original text
		return &types.Message{
			ID:         int64(u.ID),
			ChatID:     chatID,
			Date:       c.convertDate(u.Date),
			IsOutgoing: true,
			Content: types.MessageContent{
				Type: types.MessageTypeText,
				Text: originalText, // Use the original text we sent
			},
			IsEdited: false,
		}, nil

	case *tg.UpdateShort:
		// Handle UpdateShort type
		if msg, ok := u.Update.(*tg.UpdateNewMessage); ok {
			if m, ok := msg.Message.(*tg.Message); ok {
				userMap := make(map[int64]*tg.User)
				return c.convertMessage(m, userMap, chatID), nil
			}
		}
		return nil, fmt.Errorf("no message found in short update")

	default:
		return nil, fmt.Errorf("unexpected update type: %T", updates)
	}
}
