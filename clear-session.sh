#!/bin/bash
# Helper script to clear Ithil session data

SESSION_DIR="$HOME/.config/ithil"
SESSION_FILE="$SESSION_DIR/session.json"
AUTH_FILE="$SESSION_DIR/session.json.auth"

echo "Clearing Ithil session data..."

if [ -f "$SESSION_FILE" ]; then
    rm "$SESSION_FILE"
    echo "✓ Removed session file: $SESSION_FILE"
else
    echo "  No session file found"
fi

if [ -f "$AUTH_FILE" ]; then
    rm "$AUTH_FILE"
    echo "✓ Removed auth file: $AUTH_FILE"
else
    echo "  No auth file found"
fi

echo ""
echo "Session cleared! You can now restart Ithil to authenticate again."
