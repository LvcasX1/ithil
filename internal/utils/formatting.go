// Package utils provides utility functions for the Ithil application.
package utils

import (
	"strings"
	"unicode/utf8"
)

// TruncateString truncates a string to the specified length, adding "..." if truncated.
func TruncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}

	if maxLen <= 3 {
		return strings.Repeat(".", maxLen)
	}

	// Count runes, not bytes
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}

	return string(runes[:maxLen-3]) + "..."
}

// WordWrap wraps text to fit within the specified width.
func WordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else {
			testLine := currentLine + " " + word
			if utf8.RuneCountInString(testLine) <= width {
				currentLine = testLine
			} else {
				lines = append(lines, currentLine)
				currentLine = word
			}
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

// PadRight pads a string to the right with spaces.
func PadRight(s string, length int) string {
	count := utf8.RuneCountInString(s)
	if count >= length {
		return s
	}
	return s + strings.Repeat(" ", length-count)
}

// PadLeft pads a string to the left with spaces.
func PadLeft(s string, length int) string {
	count := utf8.RuneCountInString(s)
	if count >= length {
		return s
	}
	return strings.Repeat(" ", length-count) + s
}

// Center centers a string within the specified width.
func Center(s string, width int) string {
	count := utf8.RuneCountInString(s)
	if count >= width {
		return s
	}

	padding := width - count
	leftPadding := padding / 2
	rightPadding := padding - leftPadding

	return strings.Repeat(" ", leftPadding) + s + strings.Repeat(" ", rightPadding)
}

// SplitLines splits text into lines, respecting newlines.
func SplitLines(text string) []string {
	return strings.Split(text, "\n")
}

// FormatFileSize formats a file size in bytes to a human-readable string.
func FormatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return formatFloat(float64(bytes)/float64(GB)) + " GB"
	case bytes >= MB:
		return formatFloat(float64(bytes)/float64(MB)) + " MB"
	case bytes >= KB:
		return formatFloat(float64(bytes)/float64(KB)) + " KB"
	default:
		return formatInt(bytes) + " B"
	}
}

// formatFloat formats a float to 2 decimal places.
func formatFloat(f float64) string {
	if f == float64(int64(f)) {
		return formatInt(int64(f))
	}
	return strings.TrimRight(strings.TrimRight(formatFloatPrecision(f, 2), "0"), ".")
}

// formatFloatPrecision formats a float with specified precision.
func formatFloatPrecision(f float64, precision int) string {
	result := ""
	// Simple sprintf replacement for float
	intPart := int64(f)
	fracPart := f - float64(intPart)
	result = formatInt(intPart)
	if precision > 0 {
		result += "."
		for i := 0; i < precision; i++ {
			fracPart *= 10
			digit := int64(fracPart) % 10
			result += formatInt(digit)
		}
	}
	return result
}

// formatInt converts an int64 to string.
func formatInt(i int64) string {
	if i == 0 {
		return "0"
	}

	negative := i < 0
	if negative {
		i = -i
	}

	var result []byte
	for i > 0 {
		digit := i % 10
		result = append([]byte{byte('0' + digit)}, result...)
		i /= 10
	}

	if negative {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}

// StripHTML removes HTML tags from text (basic implementation).
func StripHTML(s string) string {
	// Very basic HTML tag removal
	inTag := false
	var result strings.Builder

	for _, char := range s {
		if char == '<' {
			inTag = true
			continue
		}
		if char == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// EscapeMarkdown escapes markdown special characters.
func EscapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"*", "\\*",
		"_", "\\_",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"`", "\\`",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(s)
}

// Max returns the maximum of two integers.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min returns the minimum of two integers.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Clamp clamps a value between min and max.
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
