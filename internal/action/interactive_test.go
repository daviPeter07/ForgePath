package action

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestIsInteractiveRejectsNonTerminalInputs(t *testing.T) {
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer devNull.Close()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()

	file, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	inputs := map[string]io.Reader{
		"null device":    devNull,
		"pipe":           reader,
		"regular file":   file,
		"in-memory data": strings.NewReader("input"),
		"buffer":         &bytes.Buffer{},
	}
	for name, input := range inputs {
		if isInteractive(input) {
			t.Errorf("isInteractive(%s) = true, want false", name)
		}
	}
}
