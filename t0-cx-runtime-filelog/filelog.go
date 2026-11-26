// Package filelog opens, creates and stats files, and prints each call and
// its failure while logging is enabled, by cx --log-file or os.LogFile.
package filelog

import (
	"fmt"
	"os"
)

// VIOLATION: state outside State. Owner: t0-cx-runtime-filelog; moves to ConfigMng in phase 2.
var enabled bool

// Enable turns the logging of file calls on or off.
func Enable(enable bool) {
	enabled = enable
}

// Enabled reports whether file calls are logged.
func Enabled() bool {
	return enabled
}

// Open opens filename for reading.
func Open(filename string) (*os.File, error) {
	if enabled {
		fmt.Printf("CXOpenFile: Opening '%s'\n", filename)
	}
	file, err := os.Open(filename)
	if enabled && err != nil {
		fmt.Printf("CXOpenFile: Failed to open '%s': %v\n", filename, err)
	}
	return file, err
}

// Create creates or truncates filename.
func Create(filename string) (*os.File, error) {
	if enabled {
		fmt.Printf("Creating file : '', '%s'\n", filename)
	}
	file, err := os.Create(filename)
	if enabled && err != nil {
		fmt.Printf("Failed to create file : '', '%s', err '%v'\n", filename, err)
	}
	return file, err
}

// Stat returns the file information of path.
func Stat(path string) (os.FileInfo, error) {
	if enabled {
		fmt.Printf("Stating file : '', '%s'\n", path)
	}
	fileStat, err := os.Stat(path)
	if enabled && err != nil {
		fmt.Printf("Failed to stat file : '', '%s', err '%v'\n", path, err)
	}
	return fileStat, err
}
