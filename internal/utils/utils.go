package utils

import (
    "fmt"
    "regexp"
)

// ansiRegexp matches ANSI escape sequences.
var ansiRegexp = regexp.MustCompile(`\033\[[0-9;]*m`)

// StripANSI removes ANSI escape sequences from a string.
func StripANSI(s string) string {
    return ansiRegexp.ReplaceAllString(s, "")
}

// TruncateLine truncates a string to maxLen characters, appending "..." if truncated.
func TruncateLine(line string, maxLen int) string {
    if len(line) > maxLen {
        return line[:maxLen] + "..."
    }
    return line
}

// Bold wraps text in ANSI codes for bold formatting.
func Bold(text string) string {
    return "\033[1m" + text + "\033[0m"
}

// Color wraps text in ANSI codes for the specified foreground color.
// Supported colors: "black", "red", "green", "yellow", "blue", "magenta", "cyan", "white".
func Color(text, color string) string {
    var code string
    switch color {
    case "black":
        code = "30"
    case "red":
        code = "31"
    case "green":
        code = "32"
    case "yellow":
        code = "33"
    case "blue":
        code = "34"
    case "magenta":
        code = "35"
    case "cyan":
        code = "36"
    case "white":
        code = "37"
    case "gray", "grey":
        code = "90"
    default:
        code = "37"
    }
    return fmt.Sprintf("\033[%sm%s\033[0m", code, text)
}

// Highlight wraps text in ANSI codes for bold formatting with both foreground and background colors.
// Example: Highlight("feat", "white", "green")
func Highlight(text, fg, bg string) string {
    fgCodes := map[string]string{
        "black":         "30",
        "red":           "31",
        "green":         "32",
        "yellow":        "33",
        "blue":          "34",
        "magenta":       "35",
        "cyan":          "36",
        "white":         "37",
        "gray":          "90",
        "grey":          "90",
   }
    bgCodes := map[string]string{
        "black":         "40",
        "red":           "41",
        "green":         "42",
        "yellow":        "43",
        "blue":          "44",
        "magenta":       "45",
        "cyan":          "46",
        "white":         "47",
        "gray":          "100",
        "grey":          "100",
   }
    fgCode, ok := fgCodes[fg]
    if !ok {
        fgCode = fgCodes["white"]
    }
    bgCode, ok := bgCodes[bg]
    if !ok {
        bgCode = bgCodes["black"]
    }
    // Bold with the specified foreground and background colors.
    return fmt.Sprintf("\033[1;%s;%sm%s\033[0m", fgCode, bgCode, text)
}

// ColorReset returns the ANSI code to reset terminal styles.
func ColorReset() string {
    return "\033[0m"
}

