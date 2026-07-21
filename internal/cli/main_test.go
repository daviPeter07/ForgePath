package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	directory, err := os.MkdirTemp("", "forgepath-tests-")
	if err != nil {
		panic(err)
	}
	variables := map[string]string{
		"FORGEPATH_CACHE":  directory,
		"FORGEPATH_CONFIG": filepath.Join(directory, "config.json"),
		"FORGEPATH_STATE":  filepath.Join(directory, "state.json"),
	}
	for name, value := range variables {
		if err := os.Setenv(name, value); err != nil {
			panic(err)
		}
	}

	code := m.Run()
	_ = os.RemoveAll(directory)
	os.Exit(code)
}
