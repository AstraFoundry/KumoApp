//go:build !windows

package storage

import (
	"os"
	"path/filepath"
	"syscall"
)

// FileLock is an advisory lock on a file, held across processes.
type FileLock struct{ file *os.File }

// Lock blocks until the lock at path is held.
func Lock(path string) (*FileLock, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, err
	}
	return &FileLock{file: f}, nil
}

// Unlock releases the lock.
func (l *FileLock) Unlock() error {
	defer l.file.Close()
	return syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
}
