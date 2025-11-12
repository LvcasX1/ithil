package telegram

import (
	"testing"

	"github.com/lvcasx1/ithil/pkg/types"
)

func TestChatToInputPeer_Private(t *testing.T) {
	client := &Client{}

	chat := &types.Chat{
		ID:         12345,
		Type:       types.ChatTypePrivate,
		AccessHash: 98765,
	}

	peer, err := client.chatToInputPeer(chat)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if peer == nil {
		t.Fatal("Expected peer, got nil")
	}
}

func TestChatToInputPeer_Group(t *testing.T) {
	client := &Client{}

	chat := &types.Chat{
		ID:   12345,
		Type: types.ChatTypeGroup,
	}

	peer, err := client.chatToInputPeer(chat)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if peer == nil {
		t.Fatal("Expected peer, got nil")
	}
}

func TestChatToInputPeer_Channel(t *testing.T) {
	client := &Client{}

	chat := &types.Chat{
		ID:         12345,
		Type:       types.ChatTypeChannel,
		AccessHash: 98765,
	}

	peer, err := client.chatToInputPeer(chat)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if peer == nil {
		t.Fatal("Expected peer, got nil")
	}
}

func TestChatToInputChannel_Success(t *testing.T) {
	client := &Client{}

	chat := &types.Chat{
		ID:         12345,
		Type:       types.ChatTypeChannel,
		AccessHash: 98765,
	}

	channel, err := client.chatToInputChannel(chat)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if channel == nil {
		t.Fatal("Expected channel, got nil")
	}
}

func TestChatToInputChannel_InvalidType(t *testing.T) {
	client := &Client{}

	chat := &types.Chat{
		ID:   12345,
		Type: types.ChatTypePrivate,
	}

	_, err := client.chatToInputChannel(chat)
	if err == nil {
		t.Fatal("Expected error for non-channel chat, got nil")
	}
}

func TestMessageEntityTypes(t *testing.T) {
	// Verify message entity types are properly defined
	entityTypes := []types.EntityType{
		types.EntityTypeBold,
		types.EntityTypeItalic,
		types.EntityTypeCode,
		types.EntityTypePre,
	}

	// Just verify they exist and are distinct
	seen := make(map[types.EntityType]bool)
	for _, et := range entityTypes {
		if seen[et] {
			t.Errorf("Duplicate entity type value: %d", et)
		}
		seen[et] = true
	}
}
