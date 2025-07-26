package main

import (
	"io/fs"
	"os"
	"os/exec"
)

// WorkingDirProvider abstracts os.Getwd for testing
type WorkingDirProvider interface {
	Getwd() (string, error)
}

// CommandExecutor abstracts exec.Command for testing
type CommandExecutor interface {
	Command(name string, arg ...string) *exec.Cmd
}

// FileSystemProvider abstracts file system operations for testing
type FileSystemProvider interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm fs.FileMode) error
	Stat(name string) (fs.FileInfo, error)
}

// OSWorkingDirProvider implements WorkingDirProvider using os.Getwd
type OSWorkingDirProvider struct{}

// Getwd returns the rooted path name corresponding to the current directory.
func (o *OSWorkingDirProvider) Getwd() (string, error) {
	return os.Getwd()
}

// OSCommandExecutor implements CommandExecutor using exec.Command
type OSCommandExecutor struct{}

// Command returns the Cmd struct to execute the named program with the given arguments.
func (o *OSCommandExecutor) Command(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}

// OSFileSystemProvider implements FileSystemProvider using os functions
type OSFileSystemProvider struct{}

// ReadFile reads the named file and returns the contents.
// #nosec G304 -- filename is controlled by application logic, not user input
func (o *OSFileSystemProvider) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// WriteFile writes data to the named file, creating it if necessary.
func (o *OSFileSystemProvider) WriteFile(filename string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

// Stat returns a FileInfo describing the named file.
func (o *OSFileSystemProvider) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}
