package ast

import (
	"os"

	"github.com/skycoin/skycoin-cx/t0-programfiles" // thin: only the program path and sandbox calls; no depth gain
)

// ProgramStringBytesRead returns the bytes of the str in value. They alias
// program memory, so the caller copies them before the program runs on.
func ProgramStringBytesRead(prgrm *CXProgram, value *CXValue) []byte {
	return value.Get_str_bytes(prgrm)
}

// ProgramFileSystemPathRead validates the str in value as a program path,
// reading its bytes from program memory without a Go string, see #433.
func ProgramFileSystemPathRead(prgrm *CXProgram, value *CXValue) (programfiles.ProgramFileSystemPath, *programfiles.FileError) {
	return programfiles.ProgramFileSystemPathFromBytes(ProgramStringBytesRead(prgrm, value))
}

// ProgramSandboxFileOpen opens the program path in value for reading inside
// the program's sandbox root.
func ProgramSandboxFileOpen(prgrm *CXProgram, value *CXValue) (*os.File, *programfiles.FileError) {
	programPath, fileError := ProgramFileSystemPathRead(prgrm, value)
	if fileError != nil {
		return nil, fileError
	}
	return programfiles.ProgramSandboxFileOpen(prgrm.ProgramSandboxRoot, programPath)
}

// ProgramSandboxFileCreate creates the program path in value inside the
// program's sandbox root.
func ProgramSandboxFileCreate(prgrm *CXProgram, value *CXValue) (*os.File, *programfiles.FileError) {
	programPath, fileError := ProgramFileSystemPathRead(prgrm, value)
	if fileError != nil {
		return nil, fileError
	}
	return programfiles.ProgramSandboxFileCreate(prgrm.ProgramSandboxRoot, programPath)
}

// ProgramSandboxFileRead reads the whole program path in value inside the
// program's sandbox root.
func ProgramSandboxFileRead(prgrm *CXProgram, value *CXValue) ([]byte, *programfiles.FileError) {
	programPath, fileError := ProgramFileSystemPathRead(prgrm, value)
	if fileError != nil {
		return nil, fileError
	}
	return programfiles.ProgramSandboxFileRead(prgrm.ProgramSandboxRoot, programPath)
}

// ProgramSandboxFileOpenReadWrite opens the program path in value for
// reading and writing inside the program's sandbox root, creating it when
// it does not exist.
func ProgramSandboxFileOpenReadWrite(prgrm *CXProgram, value *CXValue) (*os.File, *programfiles.FileError) {
	programPath, fileError := ProgramFileSystemPathRead(prgrm, value)
	if fileError != nil {
		return nil, fileError
	}
	return programfiles.ProgramSandboxFileOpenReadWrite(prgrm.ProgramSandboxRoot, programPath)
}

// ProgramSandboxFileRemove removes the program path in value inside the
// program's sandbox root.
func ProgramSandboxFileRemove(prgrm *CXProgram, value *CXValue) *programfiles.FileError {
	programPath, fileError := ProgramFileSystemPathRead(prgrm, value)
	if fileError != nil {
		return fileError
	}
	return programfiles.ProgramSandboxFileRemove(prgrm.ProgramSandboxRoot, programPath)
}
