# Focus Management Bug Fix

## Issue Description

**Bug**: Navigating in the chats pane (chat list) was also causing the chat content (conversation/messages) to scroll/navigate at the same time. This was a focus management issue where keyboard events were being processed by multiple components simultaneously.

## Root Cause

The root cause was in the viewport update logic in three model files:

1. `/home/lvcasx1/Work/personal/ithil/internal/ui/models/chatlist.go`
2. `/home/lvcasx1/Work/personal/ithil/internal/ui/models/conversation.go`
3. `/home/lvcasx1/Work/personal/ithil/internal/ui/models/sidebar.go`

### The Problem

In the main model's `Update()` method (`/home/lvcasx1/Work/personal/ithil/internal/ui/models/main.go` lines 314-325), all sub-models were being updated regardless of their focus state:

```go
} else {
    // Update all panes
    var cmd tea.Cmd

    m.chatList, cmd = m.chatList.Update(msg)
    cmds = append(cmds, cmd)

    m.conversation, cmd = m.conversation.Update(msg)
    cmds = append(cmds, cmd)

    m.sidebar, cmd = m.sidebar.Update(msg)
    cmds = append(cmds, cmd)
}
```

While each model had focus checks for handling navigation keys (e.g., `if !m.focused { return m, nil }`), the **viewport.Update()** call was happening **unconditionally** at the end of each model's Update() method.

This meant that even though a component correctly ignored keyboard events when not focused, its underlying viewport would still process those same keyboard events, causing unintended scrolling.

## The Fix

Added focus guards before calling `viewport.Update()` in all three component models:

### 1. ChatListModel (`chatlist.go` lines 100-105)

**Before:**
```go
// Update viewport
m.viewport, cmd = m.viewport.Update(msg)
return m, cmd
```

**After:**
```go
// CRITICAL: Only update viewport when this pane is focused
// This prevents navigation keys from affecting the viewport when the pane is not focused
if m.focused {
    m.viewport, cmd = m.viewport.Update(msg)
}
return m, cmd
```

### 2. ConversationModel (`conversation.go` lines 162-166)

**Before:**
```go
// Update viewport
m.viewport, cmd = m.viewport.Update(msg)
```

**After:**
```go
// CRITICAL: Only update viewport when this pane is focused
// This prevents navigation keys from affecting the viewport when the pane is not focused
if m.focused {
    m.viewport, cmd = m.viewport.Update(msg)
}
```

### 3. SidebarModel (`sidebar.go` lines 80-84)

**Before:**
```go
// Update viewport
m.viewport, cmd = m.viewport.Update(msg)
return m, cmd
```

**After:**
```go
// CRITICAL: Only update viewport when this pane is focused
// This prevents navigation keys from affecting the viewport when the pane is not focused
if m.focused {
    m.viewport, cmd = m.viewport.Update(msg)
}
return m, cmd
```

## How It Works

The focus management system works as follows:

1. **Main Model** (`main.go`) tracks which pane is currently focused via `m.focusPane` (FocusChatList, FocusConversation, or FocusSidebar)

2. **Focus Switching** happens via:
   - `Tab`: cycles to next pane
   - `Shift+Tab`: cycles to previous pane
   - `Ctrl+1`: focus chat list
   - `Ctrl+2`: focus conversation
   - `Ctrl+3`: focus sidebar

3. **updatePaneFocus()** method (lines 448-452) sets the focused state on each component:
   ```go
   func (m *MainModel) updatePaneFocus() {
       m.chatList.SetFocused(m.focusPane == FocusChatList)
       m.conversation.SetFocused(m.focusPane == FocusConversation)
       m.sidebar.SetFocused(m.focusPane == FocusSidebar)
   }
   ```

4. **Each component** now properly guards both:
   - Navigation key handling (already existed)
   - Viewport updates (newly added)

## Testing

To verify the fix works correctly:

1. **Chat List Navigation**:
   - Focus the chat list (Ctrl+1)
   - Use j/k or up/down arrows
   - Verify: Only the chat list scrolls, conversation stays still

2. **Conversation Navigation**:
   - Focus the conversation (Ctrl+2)
   - Use j/k or up/down arrows
   - Verify: Only the conversation scrolls, chat list stays still

3. **Sidebar Navigation**:
   - Toggle sidebar visible (Ctrl+S)
   - Focus the sidebar (Ctrl+3)
   - Use j/k or up/down arrows
   - Verify: Only the sidebar scrolls, other panes stay still

4. **Tab Switching**:
   - Use Tab/Shift+Tab to cycle through panes
   - Verify: Visual focus indicator changes (border colors)
   - Verify: Navigation only affects the focused pane

5. **All Other Features**:
   - Input field focusing (i/a keys in conversation)
   - Message sending (Enter when input is focused)
   - Reply/Edit functionality (r/e keys)
   - Help toggle (?)
   - All other keyboard shortcuts should remain functional

## Key Learnings

1. **Event Propagation**: In Bubbletea, when the main model calls `Update()` on sub-models, messages propagate to ALL components. It's the responsibility of each component to filter messages based on its state (focused, visible, etc.)

2. **Viewport Components**: Bubble Tea's viewport component responds to keyboard events automatically. When using viewports, you must guard the `viewport.Update()` call with focus checks to prevent unintended interactions.

3. **Focus Guards**: Both explicit key handling AND child component updates need focus guards:
   - Explicit: `if !m.focused { return m, nil }`
   - Child components: `if m.focused { childComponent.Update(msg) }`

4. **Consistent Pattern**: Apply the same focus isolation pattern to all components that can handle input to ensure predictable behavior.

## Impact

This fix ensures proper focus isolation between panes, preventing the confusing behavior where navigation in one pane affected another pane. The application now behaves as expected with clear, isolated component interaction.

## Files Modified

1. `/home/lvcasx1/Work/personal/ithil/internal/ui/models/chatlist.go`
2. `/home/lvcasx1/Work/personal/ithil/internal/ui/models/conversation.go`
3. `/home/lvcasx1/Work/personal/ithil/internal/ui/models/sidebar.go`

## Build Status

Build successful with no compilation errors.
