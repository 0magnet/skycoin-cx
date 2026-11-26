package ast

import (
	"os"

	sandbox "github.com/skycoin/skycoin-cx/t1-cx-runtime-sandbox"
)

// ProgramStringBytesRead returns the bytes of the str in value. They alias
// program memory, so the caller copies them before the program runs on.
func ProgramStringBytesRead(prgrm *CXProgram, value *CXValue) []byte {
	return value.StrBytesGet(prgrm)
}

// ProgramFileSystemPathRead validates the str in value as a program path,
// reading its bytes from program memory without a Go string, see #433.
func ProgramFileSystemPathRead(prgrm *CXProgram, value *CXValue) (sandbox.Path, *sandbox.FileError) {
	return sandbox.PathFromBytes(ProgramStringBytesRead(prgrm, value))
}

// ProgramSandboxFileOpen opens the program path in value for reading inside
// the program's sandbox root.
func ProgramSandboxFileOpen(prgrm *CXProgram, value *CXValue) (*os.File, *sandbox.FileError) {
	programPath, fileError := ProgramFileSystemPathRead(prgrm, value)
	if fileError != nil {
		return nil, fileError
	}
	return sandbox.FileOpen(prgrm.ProgramSandboxRoot, programPath)
}

// ProgramSandboxFileCreate creates the program path in value inside the
// program's sandbox root.
func ProgramSandboxFileCreate(prgrm *CXProgram, value *CXValue) (*os.File, *sandbox.FileError) {
	programPath, fileError := ProgramFileSystemPathRead(prgrm, value)
	if fileError != nil {
		return nil, fileError
	}
	return sandbox.FileCreate(prgrm.ProgramSandboxRoot, programPath)
}

// ProgramSandboxFileRead reads the whole program path in value inside the
// program's sandbox root.
func ProgramSandboxFileRead(prgrm *CXProgram, value *CXValue) ([]byte, *sandbox.FileError) {
	programPath, fileError := ProgramFileSystemPathRead(prgrm, value)
	if fileError != nil {
		return nil, fileError
	}
	return sandbox.FileRead(prgrm.ProgramSandboxRoot, programPath)
}

// ProgramSandboxFileOpenReadWrite opens the program path in value for
// reading and writing inside the program's sandbox root, creating it when
// it does not exist.
func ProgramSandboxFileOpenReadWrite(prgrm *CXProgram, value *CXValue) (*os.File, *sandbox.FileError) {
	programPath, fileError := ProgramFileSystemPathRead(prgrm, value)
	if fileError != nil {
		return nil, fileError
	}
	return sandbox.FileOpenReadWrite(prgrm.ProgramSandboxRoot, programPath)
}

// ProgramSandboxFileRemove removes the program path in value inside the
// program's sandbox root.
func ProgramSandboxFileRemove(prgrm *CXProgram, value *CXValue) *sandbox.FileError {
	programPath, fileError := ProgramFileSystemPathRead(prgrm, value)
	if fileError != nil {
		return fileError
	}
	return sandbox.FileRemove(prgrm.ProgramSandboxRoot, programPath)
}
