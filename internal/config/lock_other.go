//go:build !windows && !linux && !darwin

package config

import "os"

func tryLockFile(_ *os.File) (bool, error) {
	return true, nil
}

func unlockFile(_ *os.File) error {
	return nil
}
