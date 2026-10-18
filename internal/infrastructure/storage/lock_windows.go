package storage

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// FileLock is an exclusive lock on a file, held across processes.
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
	var overlapped windows.Overlapped
	if err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, &overlapped); err != nil {
		f.Close()
		return nil, err
	}
	return &FileLock{file: f}, nil
}

// Unlock releases the lock.
func (l *FileLock) Unlock() error {
	defer l.file.Close()
	var overlapped windows.Overlapped
	return windows.UnlockFileEx(windows.Handle(l.file.Fd()), 0, 1, 0, &overlapped)
}
