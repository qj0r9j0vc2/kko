package output

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

type Alignment int

const (
	AlignLeft Alignment = iota
	AlignRight
	AlignCenter
)

type Column struct {
	Header string
	Width  int
	Align  Alignment
}

type TableWriter struct {
	w       io.Writer
	columns []Column
	rows    [][]string
}

func NewTable(columns ...Column) *TableWriter {
	return &TableWriter{
		w:       os.Stdout,
		columns: columns,
	}
}

func (t *TableWriter) SetWriter(w io.Writer) {
	t.w = w
}

func (t *TableWriter) AddRow(values ...string) {
	row := make([]string, len(t.columns))
	for i, v := range values {
		if i < len(t.columns) {
			row[i] = v
		}
	}
	t.rows = append(t.rows, row)
}

func (t *TableWriter) Render() {
	for i := range t.columns {
		if t.columns[i].Width == 0 {
			t.columns[i].Width = utf8.RuneCountInString(t.columns[i].Header)
			for _, row := range t.rows {
				if i < len(row) {
					w := utf8.RuneCountInString(row[i])
					if w > t.columns[i].Width {
						t.columns[i].Width = w
					}
				}
			}
		}
	}

	t.printHeader()
	for _, row := range t.rows {
		t.printRow(row)
	}
}

func (t *TableWriter) printHeader() {
	parts := make([]string, len(t.columns))
	for i, col := range t.columns {
		parts[i] = t.pad(Header(col.Header), col.Header, col.Width, col.Align)
	}
	fmt.Fprintln(t.w, "  "+strings.Join(parts, "  "))
}

func (t *TableWriter) printRow(row []string) {
	parts := make([]string, len(t.columns))
	for i, col := range t.columns {
		val := ""
		if i < len(row) {
			val = row[i]
		}
		parts[i] = t.pad(val, val, col.Width, col.Align)
	}
	fmt.Fprintln(t.w, "  "+strings.Join(parts, "  "))
}

func (t *TableWriter) pad(display, raw string, width int, align Alignment) string {
	rawWidth := utf8.RuneCountInString(raw)
	padding := width - rawWidth
	if padding < 0 {
		padding = 0
	}

	switch align {
	case AlignRight:
		return strings.Repeat(" ", padding) + display
	case AlignCenter:
		left := padding / 2
		right := padding - left
		return strings.Repeat(" ", left) + display + strings.Repeat(" ", right)
	default:
		return display + strings.Repeat(" ", padding)
	}
}

func PrintDivider(w io.Writer, width int) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintln(w, "  "+Muted(strings.Repeat("─", width)))
}
