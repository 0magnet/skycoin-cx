package sandbox

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unsafe"

	filelog "github.com/skycoin/skycoin-cx/t0-cx-runtime-filelog"
)

// PathLengthMaximum is the longest path a program may pass, in bytes.
const PathLengthMaximum = 4096

// Path is a path a running CX program passes to a file opcode, validated to
// stay inside the sandbox root, see #414 and #433. It views the cleaned bytes
// that PathClean appended, which the calling library keeps in its pool (D139).
type Path struct {
	pathText []byte
}

// Text returns the cleaned, slash separated path. The caller must not write
// to it.
func (programPath Path) Text() []byte {
	return programPath.pathText
}

// PathFromBytes validates a path read from program memory and returns it
// cleaned in new bytes.
func PathFromBytes(pathBytes []byte) (Path, *FileError) {
	return PathClean(nil, pathBytes)
}

// PathClean validates a path read from program memory and appends it cleaned
// to dst, returning a Path that views the appended bytes. It must be
// relative, use only '/' and not start with ".." once cleaned.
func PathClean(dst []byte, pathBytes []byte) (Path, *FileError) {
	switch {
	case len(pathBytes) == 0:
		return pathReject(FileErrorCodePathInvalid, FilePathReasonEmpty, pathBytes)
	case len(pathBytes) > PathLengthMaximum:
		return pathReject(FileErrorCodePathInvalid, FilePathReasonTooLong, pathBytes)
	case bytes.IndexByte(pathBytes, 0) >= 0:
		return pathReject(FileErrorCodePathInvalid, FilePathReasonNul, pathBytes)
	case bytes.IndexByte(pathBytes, '\\') >= 0:
		return pathReject(FileErrorCodePathInvalid, FilePathReasonBackslash, pathBytes)
	case pathBytes[0] == '/':
		return pathReject(FileErrorCodePathOutsideSandbox, FilePathReasonAbsolute, pathBytes)
	case len(pathBytes) >= 2 && pathBytes[1] == ':' && pathDriveLetter(pathBytes[0]):
		return pathReject(FileErrorCodePathOutsideSandbox, FilePathReasonVolume, pathBytes)
	}
	start := len(dst)
	dst = pathCleanAppend(dst, pathBytes)
	cleaned := dst[start:len(dst):len(dst)]
	if len(cleaned) == 2 && cleaned[0] == '.' && cleaned[1] == '.' || len(cleaned) >= 3 && cleaned[0] == '.' && cleaned[1] == '.' && cleaned[2] == '/' {
		return pathReject(FileErrorCodePathOutsideSandbox, FilePathReasonLeavesSandbox, pathBytes)
	}
	return Path{pathText: cleaned}, nil
}

// pathCleanAppend appends the shortest path equal to the relative path in
// name, the result of path.Clean in go/src/path/path.go, without a Go string.
func pathCleanAppend(dst []byte, name []byte) []byte {
	start := len(dst)
	dotdot := start
	for r := 0; r < len(name); {
		switch {
		case name[r] == '/':
			r++
		case name[r] == '.' && (r+1 == len(name) || name[r+1] == '/'):
			r++
		case name[r] == '.' && name[r+1] == '.' && (r+2 == len(name) || name[r+2] == '/'):
			r += 2
			switch {
			case len(dst) > dotdot:
				w := len(dst) - 1
				for w > dotdot && dst[w] != '/' {
					w--
				}
				dst = dst[:w]
			default:
				if len(dst) > start {
					dst = append(dst, '/')
				}
				dst = append(dst, '.', '.')
				dotdot = len(dst)
			}
		default:
			if len(dst) > start {
				dst = append(dst, '/')
			}
			for ; r < len(name) && name[r] != '/'; r++ {
				dst = append(dst, name[r])
			}
		}
	}
	if len(dst) == start {
		dst = append(dst, '.')
	}
	return dst
}

func pathDriveLetter(letter byte) bool {
	return letter >= 'A' && letter <= 'Z' || letter >= 'a' && letter <= 'z'
}

func pathReject(code FileErrorCode, reason FilePathReason, pathBytes []byte) (Path, *FileError) {
	fileError := &FileError{
		FileErrorCode:      code,
		FileErrorOperation: FileOperationPathValidate,
		FileErrorReason:    reason,
	}
	if filelog.Enabled() {
		fileLogPrint(FileOperationPathValidate, nil, pathBytes, fileError)
	}
	return Path{}, fileError
}

// RootOpen opens the directory every program path resolves in.
// An empty dir means ".".
func RootOpen(dir string) (*Root, *FileError) {
	if dir == "" {
		dir = "."
	}
	root, err := rootDirectoryOpen(dir)
	if err != nil {
		return nil, fileErrorFromFileSystemError(FileOperationSandboxRootOpen, []byte(dir), err)
	}
	return root, nil
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func rootCheck(root *Root, operation FileOperationCode) *FileError {
	if root != nil {
		return nil
	}
	return &FileError{
		FileErrorCode:      FileErrorCodeSandboxRootMissing,
		FileErrorOperation: operation,
		FileErrorReason:    FilePathReasonRootMissing,
	}
}

// FileOpen opens a program path for reading inside root.
// VIOLATION STRHOT-272: B, the path to os.Root.Open as a view of pool bytes, #412.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func FileOpen(root *Root, programPath Path) (*os.File, *FileError) {
	if fileError := rootCheck(root, FileOperationOpen); fileError != nil {
		return nil, fileError
	}
	if filelog.Enabled() {
		fileLogPrint(FileOperationOpen, root, programPath.pathText, nil)
	}
	file, err := root.Open(unsafe.String(unsafe.SliceData(programPath.pathText), len(programPath.pathText)))
	if err != nil {
		return nil, pathFileError(FileOperationOpen, programPath, err)
	}
	return file, nil
}

// FileCreate creates or truncates a program path inside root.
// VIOLATION STRHOT-270: B, the path to os.Root.Create as a view of pool bytes, #412.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func FileCreate(root *Root, programPath Path) (*os.File, *FileError) {
	if fileError := rootCheck(root, FileOperationCreate); fileError != nil {
		return nil, fileError
	}
	if filelog.Enabled() {
		fileLogPrint(FileOperationCreate, root, programPath.pathText, nil)
	}
	file, err := root.Create(unsafe.String(unsafe.SliceData(programPath.pathText), len(programPath.pathText)))
	if err != nil {
		return nil, pathFileError(FileOperationCreate, programPath, err)
	}
	return file, nil
}

// FileOpenReadWrite opens a program path inside root for
// reading and writing at any offset, creating it when it does not exist and
// keeping its content when it does.
// VIOLATION STRHOT-349: B, the path to os.Root.OpenFile as a view of pool bytes, #412.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func FileOpenReadWrite(root *Root, programPath Path) (*os.File, *FileError) {
	if fileError := rootCheck(root, FileOperationOpenReadWrite); fileError != nil {
		return nil, fileError
	}
	if filelog.Enabled() {
		fileLogPrint(FileOperationOpenReadWrite, root, programPath.pathText, nil)
	}
	file, err := rootFileOpen(root, unsafe.String(unsafe.SliceData(programPath.pathText), len(programPath.pathText)), os.O_RDWR|os.O_CREATE, 0o666)
	if err != nil {
		return nil, pathFileError(FileOperationOpenReadWrite, programPath, err)
	}
	return file, nil
}

// FileRemove removes a program path inside root, a file or an
// empty directory. A symlink is removed itself, never what it points to.
// VIOLATION STRHOT-350: B, the path to os.Root.Remove as a view of pool bytes, #412.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func FileRemove(root *Root, programPath Path) *FileError {
	if fileError := rootCheck(root, FileOperationRemove); fileError != nil {
		return fileError
	}
	if filelog.Enabled() {
		fileLogPrint(FileOperationRemove, root, programPath.pathText, nil)
	}
	if err := root.Remove(unsafe.String(unsafe.SliceData(programPath.pathText), len(programPath.pathText))); err != nil {
		return pathFileError(FileOperationRemove, programPath, err)
	}
	return nil
}

// FileRead reads a whole program path inside root.
func FileRead(root *Root, programPath Path) ([]byte, *FileError) {
	file, fileError := FileOpen(root, programPath)
	if fileError != nil {
		fileError.FileErrorOperation = FileOperationRead
		return nil, fileError
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, pathFileError(FileOperationRead, programPath, err)
	}
	return content, nil
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func pathFileError(operation FileOperationCode, programPath Path, err error) *FileError {
	fileError := fileErrorFromFileSystemError(operation, programPath.pathText, err)
	if filelog.Enabled() {
		fileLogPrint(operation, nil, programPath.pathText, fileError)
	}
	return fileError
}

// fileLogPrint prints one line of the file log that os.LogFile turns on: the
// operation on pathText in root, or its failure when fileError is not nil.
// VIOLATION STRHOT-357: B, the file log text to stdout, only while os.LogFile has turned it on, #412.
func fileLogPrint(operation FileOperationCode, root *Root, pathText []byte, fileError *FileError) {
	switch {
	case fileError != nil && operation == FileOperationPathValidate:
		fmt.Printf("program path rejected: %s\n", fileError.FileErrorTextOf(pathText))
	case fileError != nil:
		fmt.Printf("program file operation %d failed: %s\n", operation, fileError.FileErrorTextOf(pathText))
	default:
		fmt.Printf("%s '%s' in '%s'\n", fileOperationName(operation), pathText, root.Name())
	}
}

func fileOperationName(operation FileOperationCode) string {
	switch operation {
	case FileOperationOpen:
		return "FileOpen: Opening"
	case FileOperationCreate:
		return "FileCreate: Creating"
	case FileOperationOpenReadWrite:
		return "FileOpenReadWrite: Opening"
	case FileOperationRemove:
		return "FileRemove: Removing"
	}
	return "file operation"
}

// PathMove returns programPath viewing text, which holds the same bytes, as
// when the caller has interned them in its pool. Other bytes give the empty
// Path, which no file operation accepts.
func PathMove(programPath Path, text []byte) Path {
	if !bytes.Equal(programPath.pathText, text) {
		return Path{}
	}
	return Path{pathText: text}
}
