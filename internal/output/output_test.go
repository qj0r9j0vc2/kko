package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestTableWriter_Render(t *testing.T) {
	SetColor(false)
	defer SetColor(true)

	var buf bytes.Buffer
	table := NewTable(
		Column{Header: "#", Width: 3, Align: AlignRight},
		Column{Header: "NAME"},
	)
	table.SetWriter(&buf)
	table.AddRow("1", "Starbucks Hakdong")
	table.AddRow("2", "Starbucks Sinsa")
	table.Render()

	out := buf.String()
	if !strings.Contains(out, "Starbucks Hakdong") {
		t.Errorf("expected output to contain 'Starbucks Hakdong', got:\n%s", out)
	}
	if !strings.Contains(out, "NAME") {
		t.Errorf("expected header row with NAME")
	}
}

func TestTableWriter_EmptyRows(t *testing.T) {
	SetColor(false)
	defer SetColor(true)

	var buf bytes.Buffer
	table := NewTable(
		Column{Header: "A"},
		Column{Header: "B"},
	)
	table.SetWriter(&buf)
	table.Render()

	out := buf.String()
	if !strings.Contains(out, "A") {
		t.Error("expected header even with no rows")
	}
}

func TestTableWriter_Alignment(t *testing.T) {
	SetColor(false)
	defer SetColor(true)

	var buf bytes.Buffer
	table := NewTable(
		Column{Header: "NUM", Width: 6, Align: AlignRight},
	)
	table.SetWriter(&buf)
	table.AddRow("42")
	table.Render()

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Fatal("expected at least 2 lines (header + row)")
	}
	dataLine := lines[1]
	if !strings.Contains(dataLine, "    42") {
		t.Errorf("expected right-aligned '42' with padding, got %q", dataLine)
	}
}

func TestColorize_Disabled(t *testing.T) {
	SetColor(false)
	defer SetColor(true)

	got := Header("hello")
	if strings.Contains(got, "\033[") {
		t.Error("expected no ANSI codes when color disabled")
	}
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestColorFunctions_Disabled(t *testing.T) {
	SetColor(false)
	defer SetColor(true)

	funcs := map[string]func(string) string{
		"Header":  Header,
		"Label":   Label,
		"Muted":   Muted,
		"Success": Success,
		"Warning": Warning,
		"Error":   Error,
		"Link":    Link,
	}
	for name, fn := range funcs {
		got := fn("test")
		if strings.Contains(got, "\033[") {
			t.Errorf("%s() should not contain ANSI codes when color disabled", name)
		}
	}
}

func TestMarshalJSON(t *testing.T) {
	data := map[string]string{"key": "value"}
	s, err := MarshalJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, `"key": "value"`) {
		t.Errorf("unexpected JSON output: %s", s)
	}
}

func TestMarshalJSON_Indented(t *testing.T) {
	data := map[string]int{"a": 1}
	s, err := MarshalJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s, "  ") {
		t.Error("expected indented JSON output")
	}
}
