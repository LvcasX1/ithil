package telegram

import (
	"testing"

	"github.com/lvcasx1/ithil/pkg/types"
)

// Basic smoke tests for chat-related functionality

func TestChatTypeConstants(t *testing.T) {
	// Verify chat type constants are properly defined
	chatTypes := []types.ChatType{
		types.ChatTypePrivate,
		types.ChatTypeGroup,
		types.ChatTypeSupergroup,
		types.ChatTypeChannel,
	}

	// Just verify they exist and are distinct
	seen := make(map[types.ChatType]bool)
	for _, ct := range chatTypes {
		if seen[ct] {
			t.Errorf("Duplicate chat type value: %d", ct)
		}
		seen[ct] = true
	}
}

func TestUserStatusConstants(t *testing.T) {
	// Verify user status constants are properly defined
	statuses := []types.UserStatus{
		types.UserStatusOnline,
		types.UserStatusOffline,
		types.UserStatusRecently,
		types.UserStatusLastWeek,
		types.UserStatusLastMonth,
	}

	// Just verify they exist and are distinct
	seen := make(map[types.UserStatus]bool)
	for _, status := range statuses {
		if seen[status] {
			t.Errorf("Duplicate status value: %d", status)
		}
		seen[status] = true
	}
}
