package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
)

// Writer prints text output, styled only on a terminal. Status labels are
// ASCII so color is never the only signal.
type Writer struct {
	out   io.Writer
	color bool
}

func (w *Writer) style(code, text string) string {
	if !w.color {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

// Line prints one line.
func (w *Writer) Line(text string) { fmt.Fprintln(w.out, text) }

// Status prints "[ok] message" and its siblings.
func (w *Writer) Status(kind, message string) {
	codes := map[string]string{"ok": "32", "warn": "33", "error": "31", "dry-run": "36", "info": "2"}
	w.Line(w.style(codes[kind], "["+kind+"]") + " " + message)
}

// Field prints an aligned "label  value" line.
func (w *Writer) Field(label string, value any) {
	fmt.Fprintf(w.out, "%s %v\n", w.style("2", fmt.Sprintf("%-14s", label)), value)
}

// Heading prints a bold line.
func (w *Writer) Heading(text string) { w.Line(w.style("1", text)) }

// Dim styles secondary text.
func (w *Writer) Dim(text string) string { return w.style("2", text) }

// Table prints rows in aligned columns.
func (w *Writer) Table(header []string, rows [][]string) {
	tw := tabwriter.NewWriter(w.out, 0, 0, 2, ' ', 0)
	if len(header) > 0 {
		for i, h := range header {
			header[i] = w.style("2", h)
		}
		fmt.Fprintln(tw, strings.Join(header, "\t"))
	}
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	tw.Flush()
}

// Value prints any data as indented JSON, for commands without a text
// rendering.
func (w *Writer) Value(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintln(w.out, v)
		return
	}
	fmt.Fprintln(w.out, string(data))
}

// resourcesDir is Contents/Resources when exe is inside a macOS app, else
// the executable's own directory.
func resourcesDir(exe string) string {
	dir := filepath.Dir(exe)
	if filepath.Base(dir) == "MacOS" {
		resources := filepath.Join(filepath.Dir(dir), "Resources")
		if _, err := os.Stat(resources); err == nil {
			return resources
		}
	}
	return dir
}
