package sandbox

import (
	"errors"
	"fmt"
	"io/fs"
)

// FileErrorCode is a file failure code, 1001 to 1499, see docs/ErrorHandling.md.
type FileErrorCode uint32

// FileOperationCode is a file operation code, 1500 to 1999, see docs/ErrorHandling.md.
type FileOperationCode uint32

const (
	FileErrorCodeNotFound              FileErrorCode = 1001
	FileErrorCodePermissionDenied      FileErrorCode = 1002
	FileErrorCodePathInvalid           FileErrorCode = 1003
	FileErrorCodePathOutsideSandbox    FileErrorCode = 1004
	FileErrorCodeSandboxRootMissing    FileErrorCode = 1005
	FileErrorCodeFileSystemCallFailure FileErrorCode = 1006
	FileErrorCodeContentInvalid        FileErrorCode = 1007
)

const (
	FileOperationOpen            FileOperationCode = 1501
	FileOperationCreate          FileOperationCode = 1502
	FileOperationRead            FileOperationCode = 1503
	FileOperationPathValidate    FileOperationCode = 1504
	FileOperationSandboxRootOpen FileOperationCode = 1505
	FileOperationWrite           FileOperationCode = 1506
	FileOperationOpenReadWrite   FileOperationCode = 1507
	FileOperationRemove          FileOperationCode = 1508
)

// FilePathReason says why a program path was refused, with the text
// FileErrorTextOf gives it, 0 for a failure the os package reported.
type FilePathReason uint32

const (
	FilePathReasonNone FilePathReason = iota
	FilePathReasonEmpty
	FilePathReasonTooLong
	FilePathReasonNul
	FilePathReasonBackslash
	FilePathReasonAbsolute
	FilePathReasonVolume
	FilePathReasonLeavesSandbox
	FilePathReasonRootMissing
)

// FileError is a class A failure of a file operation, nil on success. A
// refused program path has a FileErrorReason and no text until
// FileErrorTextOf builds it, so the run path builds no string (#412).
type FileError struct {
	FileErrorCode      FileErrorCode
	FileErrorOperation FileOperationCode
	FileErrorReason    FilePathReason
	FileErrorMessage   string
	FileErrorPath      string
}

// FileErrorTextOf returns the text of the failure of the path in pathText.
// VIOLATION STRHOT-358: B, the failure text for the file log and for callers that print it, #412.
func (fileError *FileError) FileErrorTextOf(pathText []byte) string {
	switch fileError.FileErrorReason {
	case FilePathReasonNone:
		return fileError.FileErrorMessage
	case FilePathReasonRootMissing:
		return "no sandbox root is open for this program"
	}
	return fmt.Sprintf("program path %q: %s", pathText, filePathReasonText[fileError.FileErrorReason])
}

var filePathReasonText = [...]string{
	FilePathReasonEmpty:         "empty",
	FilePathReasonTooLong:       "longer than 4096 bytes",
	FilePathReasonNul:           "contains a NUL byte",
	FilePathReasonBackslash:     "contains '\\', use '/'",
	FilePathReasonAbsolute:      "absolute",
	FilePathReasonVolume:        "has a volume name",
	FilePathReasonLeavesSandbox: "leaves the sandbox",
}

// fileSystemPathEscapeMessage is the text of os.errPathEscapes, which os.Root
// returns for a path or symlink that leaves the root (go/src/os/file.go).
const fileSystemPathEscapeMessage = "path escapes from parent"

// fileErrorFromFileSystemError wraps a Go error from the os package.
// VIOLATION STRHOT-268: B, a failed os call, its class told by the text of an unexported os error, #412.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func fileErrorFromFileSystemError(operation FileOperationCode, pathText []byte, fileSystemError error) *FileError {
	code := FileErrorCodeFileSystemCallFailure
	var pathError *fs.PathError
	switch {
	case errors.As(fileSystemError, &pathError) && pathError.Err.Error() == fileSystemPathEscapeMessage:
		code = FileErrorCodePathOutsideSandbox
	case errors.Is(fileSystemError, fs.ErrNotExist):
		code = FileErrorCodeNotFound
	case errors.Is(fileSystemError, fs.ErrPermission):
		code = FileErrorCodePermissionDenied
	}
	return &FileError{
		FileErrorCode:      code,
		FileErrorOperation: operation,
		FileErrorMessage:   fileSystemError.Error(),
		FileErrorPath:      string(pathText),
	}
}

// FileErrorWrap wraps a Go error of a file operation of another package,
// at its module boundary.
func FileErrorWrap(operation FileOperationCode, path string, fileSystemError error) *FileError {
	return fileErrorFromFileSystemError(operation, []byte(path), fileSystemError)
}
