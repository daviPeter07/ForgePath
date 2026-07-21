package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReportsErrorsOnStderr(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	missing := filepath.Join(t.TempDir(), "missing")

	code := run(context.Background(), []string{"list", missing}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "missing") {
		t.Fatalf("stderr = %q, want path error", stderr.String())
	}
}
