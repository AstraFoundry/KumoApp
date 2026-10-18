package application

import (
	"io"
	"os"
	"regexp"
	"strings"
)

// tail returns up to max bytes from the end of a file.
func tail(path string, max int64) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return ""
	}
	start := max64(0, info.Size()-max)
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return ""
	}
	data, _ := io.ReadAll(f)
	return string(data)
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

var logMessage = regexp.MustCompile(`msg="((?:[^"\\]|\\.)*)"`)

// lastErrorLine is the last error or fatal message the core logged since
// Kumo last started it.
func lastErrorLine(path string) string {
	text := tail(path, 64<<10)
	if i := strings.LastIndex(text, "--- Kumo started"); i >= 0 {
		text = text[i:]
	}
	lines := strings.Split(text, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		lower := strings.ToLower(line)
		if !strings.Contains(lower, "level=error") && !strings.Contains(lower, "level=fatal") && !strings.Contains(line, "FATA") && !strings.Contains(line, "ERRO") {
			continue
		}
		if m := logMessage.FindStringSubmatch(line); m != nil {
			return strings.ReplaceAll(m[1], `\"`, `"`)
		}
		return strings.TrimSpace(line)
	}
	return ""
}
