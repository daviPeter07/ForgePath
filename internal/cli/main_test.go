package cli

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cacheDirectory, err := os.MkdirTemp("", "forgepath-test-cache-")
	if err != nil {
		panic(err)
	}
	if err := os.Setenv("FORGEPATH_CACHE", cacheDirectory); err != nil {
		panic(err)
	}

	code := m.Run()
	_ = os.RemoveAll(cacheDirectory)
	os.Exit(code)
}
