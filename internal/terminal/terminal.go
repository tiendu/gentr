package terminal

import (
	"fmt"
	"regexp"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func StripANSI(value string) string {
	return ansiRegexp.ReplaceAllString(value, "")
}

func TruncateLine(value string, maxLen int) string {
	if maxLen < 0 {
		return value
	}

	runes := []rune(value)
	if len(runes) <= maxLen {
		return value
	}

	return string(runes[:maxLen]) + "..."
}

func Bold(text string) string {
	return "\x1b[1m" + text + Reset()
}

func Color(text, color string) string {
	codes := map[string]string{
		"black": "30", "red": "31", "green": "32", "yellow": "33",
		"blue": "34", "magenta": "35", "cyan": "36", "white": "37",
		"gray": "90", "grey": "90",
	}

	code, ok := codes[color]
	if !ok {
		code = codes["white"]
	}

	return fmt.Sprintf("\x1b[%sm%s%s", code, text, Reset())
}

func Highlight(text, foreground, background string) string {
	foregroundCodes := map[string]string{
		"black": "30", "red": "31", "green": "32", "yellow": "33",
		"blue": "34", "magenta": "35", "cyan": "36", "white": "37",
		"gray": "90", "grey": "90",
	}
	backgroundCodes := map[string]string{
		"black": "40", "red": "41", "green": "42", "yellow": "43",
		"blue": "44", "magenta": "45", "cyan": "46", "white": "47",
		"gray": "100", "grey": "100",
	}

	foregroundCode, ok := foregroundCodes[foreground]
	if !ok {
		foregroundCode = foregroundCodes["white"]
	}
	backgroundCode, ok := backgroundCodes[background]
	if !ok {
		backgroundCode = backgroundCodes["black"]
	}

	return fmt.Sprintf("\x1b[1;%s;%sm%s%s", foregroundCode, backgroundCode, text, Reset())
}

func Reset() string {
	return "\x1b[0m"
}
