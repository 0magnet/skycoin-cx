package ast

import (
	"os"

	sandbox "github.com/skycoin/skycoin-cx/t1-cx-runtime-sandbox"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
)

// ProgramStringBytesRead returns the bytes of the str in value. They alias
// program memory, so the caller copies them before the program runs on.
func ProgramStringBytesRead(prgrm *CXProgram, value *CXValue) []byte {
	return value.StrBytesGet(prgrm)
}

// ProgramFileSystemPathRead validates the str in value as a program path and
// interns it cleaned in pool, the pool of the library that calls, see #433
// and D139 on #412. The Path views the pool bytes of the Gid.
func ProgramFileSystemPathRead(prgrm *CXProgram, value *CXValue, pool *stringpool.StringPool) (sandbox.Path, stringpool.StringGid, *sandbox.FileError) {
	cleaned, fileError := sandbox.PathClean(pool.StringPoolArenaFree(), ProgramStringBytesRead(prgrm, value))
	if fileError != nil {
		return sandbox.Path{}, stringpool.StringGidNil, fileError
	}
	gid := pool.StringPoolIntern(cleaned.Text())
	return sandbox.PathMove(cleaned, pool.StringPoolBytes(gid)), gid, nil
}

// ProgramSandboxFileOpen opens the program path in value for reading inside
// the program's sandbox root, cleaning the path into pool.
func ProgramSandboxFileOpen(prgrm *CXProgram, value *CXValue, pool *stringpool.StringPool) (*os.File, *sandbox.FileError) {
	programPath, _, fileError := ProgramFileSystemPathRead(prgrm, value, pool)
	if fileError != nil {
		return nil, fileError
	}
	return sandbox.FileOpen(prgrm.ProgramSandboxRoot, programPath)
}

// ProgramSandboxFileCreate creates the program path in value inside the
// program's sandbox root, cleaning the path into pool.
func ProgramSandboxFileCreate(prgrm *CXProgram, value *CXValue, pool *stringpool.StringPool) (*os.File, *sandbox.FileError) {
	programPath, _, fileError := ProgramFileSystemPathRead(prgrm, value, pool)
	if fileError != nil {
		return nil, fileError
	}
	return sandbox.FileCreate(prgrm.ProgramSandboxRoot, programPath)
}

// ProgramSandboxFileRead reads the whole program path in value inside the
// program's sandbox root, cleaning the path into pool.
func ProgramSandboxFileRead(prgrm *CXProgram, value *CXValue, pool *stringpool.StringPool) ([]byte, *sandbox.FileError) {
	programPath, _, fileError := ProgramFileSystemPathRead(prgrm, value, pool)
	if fileError != nil {
		return nil, fileError
	}
	return sandbox.FileRead(prgrm.ProgramSandboxRoot, programPath)
}

// ProgramSandboxFileOpenReadWrite opens the program path in value for
// reading and writing inside the program's sandbox root, creating it when
// it does not exist, cleaning the path into pool.
func ProgramSandboxFileOpenReadWrite(prgrm *CXProgram, value *CXValue, pool *stringpool.StringPool) (*os.File, *sandbox.FileError) {
	programPath, _, fileError := ProgramFileSystemPathRead(prgrm, value, pool)
	if fileError != nil {
		return nil, fileError
	}
	return sandbox.FileOpenReadWrite(prgrm.ProgramSandboxRoot, programPath)
}

// ProgramSandboxFileRemove removes the program path in value inside the
// program's sandbox root, cleaning the path into pool.
func ProgramSandboxFileRemove(prgrm *CXProgram, value *CXValue, pool *stringpool.StringPool) *sandbox.FileError {
	programPath, _, fileError := ProgramFileSystemPathRead(prgrm, value, pool)
	if fileError != nil {
		return fileError
	}
	return sandbox.FileRemove(prgrm.ProgramSandboxRoot, programPath)
}
