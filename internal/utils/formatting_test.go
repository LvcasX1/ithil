package utils

import (
	"strings"
	"testing"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"Hello, World!", 13, "Hello, World!"},
		{"Hello, World!", 10, "Hello, ..."},
		{"Hello, World!", 5, "He..."},
		{"Hello, World!", 3, "..."},
		{"Hello, World!", 0, ""},
		{"Short", 20, "Short"},
		{"Hello", 8, "Hello"},
	}

	for _, tt := range tests {
		result := TruncateString(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("TruncateString(%q, %d) = %q, want %q",
				tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestTruncateString_Unicode(t *testing.T) {
	// Test with unicode characters
	input := "Hello 世界"
	result := TruncateString(input, 10)
	if result != "Hello 世界" {
		t.Errorf("Expected 'Hello 世界', got '%s'", result)
	}

	result = TruncateString(input, 7)
	if result != "Hell..." {
		t.Errorf("Expected 'Hell...', got '%s'", result)
	}
}

func TestWordWrap(t *testing.T) {
	text := "This is a long sentence that needs to be wrapped"
	result := WordWrap(text, 20)
	lines := strings.Split(result, "\n")

	if len(lines) < 2 {
		t.Error("Expected text to be wrapped into multiple lines")
	}

	for _, line := range lines {
		if len(line) > 20 {
			t.Errorf("Line exceeds width: '%s' (length: %d)", line, len(line))
		}
	}
}

func TestWordWrap_ShortText(t *testing.T) {
	text := "Short"
	result := WordWrap(text, 20)
	if result != text {
		t.Errorf("Expected '%s', got '%s'", text, result)
	}
}

func TestWordWrap_ZeroWidth(t *testing.T) {
	text := "Some text"
	result := WordWrap(text, 0)
	if result != text {
		t.Errorf("Expected original text when width is 0, got '%s'", result)
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"Hello", 10, "Hello     "},
		{"Hello", 5, "Hello"},
		{"Hello", 3, "Hello"},
		{"", 5, "     "},
	}

	for _, tt := range tests {
		result := PadRight(tt.input, tt.length)
		if result != tt.expected {
			t.Errorf("PadRight(%q, %d) = %q, want %q",
				tt.input, tt.length, result, tt.expected)
		}
	}
}

func TestPadLeft(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"Hello", 10, "     Hello"},
		{"Hello", 5, "Hello"},
		{"Hello", 3, "Hello"},
		{"", 5, "     "},
	}

	for _, tt := range tests {
		result := PadLeft(tt.input, tt.length)
		if result != tt.expected {
			t.Errorf("PadLeft(%q, %d) = %q, want %q",
				tt.input, tt.length, result, tt.expected)
		}
	}
}

func TestCenter(t *testing.T) {
	tests := []struct {
		input    string
		width    int
		expected string
	}{
		{"Hello", 11, "   Hello   "},
		{"Hello", 10, "  Hello   "},
		{"Hello", 5, "Hello"},
		{"Hello", 3, "Hello"},
	}

	for _, tt := range tests {
		result := Center(tt.input, tt.width)
		if result != tt.expected {
			t.Errorf("Center(%q, %d) = %q, want %q",
				tt.input, tt.width, result, tt.expected)
		}
	}
}

func TestSplitLines(t *testing.T) {
	text := "Line 1\nLine 2\nLine 3"
	lines := SplitLines(text)

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	if lines[0] != "Line 1" || lines[1] != "Line 2" || lines[2] != "Line 3" {
		t.Errorf("Lines don't match expected values: %v", lines)
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1 KB"},
		{1536, "1.5 KB"},
		{1048576, "1 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1 GB"},
		{5242880, "5 MB"},
		{104857600, "100 MB"},
	}

	for _, tt := range tests {
		result := FormatFileSize(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatFileSize(%d) = %q, want %q",
				tt.bytes, result, tt.expected)
		}
	}
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<p>Hello</p>", "Hello"},
		{"<b>Bold</b> text", "Bold text"},
		{"No tags here", "No tags here"},
		{"<a href='link'>Click</a>", "Click"},
		{"Text with <br/> break", "Text with  break"},
		{"", ""},
	}

	for _, tt := range tests {
		result := StripHTML(tt.input)
		if result != tt.expected {
			t.Errorf("StripHTML(%q) = %q, want %q",
				tt.input, result, tt.expected)
		}
	}
}

func TestEscapeMarkdown(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello *world*", "Hello \\*world\\*"},
		{"_italic_", "\\_italic\\_"},
		{"[link](url)", "\\[link\\]\\(url\\)"},
		{"Code: `code`", "Code: \\`code\\`"},
		{"# Header", "\\# Header"},
		{"Normal text", "Normal text"},
	}

	for _, tt := range tests {
		result := EscapeMarkdown(tt.input)
		if result != tt.expected {
			t.Errorf("EscapeMarkdown(%q) = %q, want %q",
				tt.input, result, tt.expected)
		}
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{5, 10, 10},
		{10, 5, 10},
		{5, 5, 5},
		{-5, -10, -5},
		{0, 100, 100},
	}

	for _, tt := range tests {
		result := Max(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("Max(%d, %d) = %d, want %d",
				tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b     int
		expected int
	}{
		{5, 10, 5},
		{10, 5, 5},
		{5, 5, 5},
		{-5, -10, -10},
		{0, 100, 0},
	}

	for _, tt := range tests {
		result := Min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("Min(%d, %d) = %d, want %d",
				tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		value, min, max int
		expected        int
	}{
		{5, 0, 10, 5},
		{-5, 0, 10, 0},
		{15, 0, 10, 10},
		{5, 5, 10, 5},
		{10, 0, 10, 10},
		{50, 0, 100, 50},
	}

	for _, tt := range tests {
		result := Clamp(tt.value, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("Clamp(%d, %d, %d) = %d, want %d",
				tt.value, tt.min, tt.max, result, tt.expected)
		}
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{123, "123"},
		{-456, "-456"},
		{1000, "1000"},
		{-1, "-1"},
	}

	for _, tt := range tests {
		result := formatInt(tt.input)
		if result != tt.expected {
			t.Errorf("formatInt(%d) = %q, want %q",
				tt.input, result, tt.expected)
		}
	}
}
