//go:build !nocxos

package cxos

const (
	// os
	CONST_OS_RUN_SUCCESS = iota + 0xFFFF
	CONST_OS_RUN_EMPTY_CMD
	CONST_OS_RUN_PANIC
	CONST_OS_RUN_START_FAILED
	CONST_OS_RUN_WAIT_FAILED
	CONST_OS_RUN_TIMEOUT

	// os.FileModes (FIXME: these are uint32 in Go, with _DIR & _TYPE exceeding math.MaxInt32)
	//CONST_OS_FILEMODE_DIR
	//CONST_OS_FILEMODE_TYPE
	CONST_OS_FILEMODE_APPEND
	CONST_OS_FILEMODE_EXCLUSIVE
	CONST_OS_FILEMODE_TEMPORARY
	CONST_OS_FILEMODE_SYMLINK
	CONST_OS_FILEMODE_DEVICE
	CONST_OS_FILEMODE_NAMED_PIPE
	CONST_OS_FILEMODE_SOCKET
	CONST_OS_FILEMODE_SETUID
	CONST_OS_FILEMODE_SETGID
	CONST_OS_FILEMODE_CHAR_DEVICE
	CONST_OS_FILEMODE_STICKY
	CONST_OS_FILEMODE_IRREGULAR
	CONST_OS_FILEMODE_PERM

	CONST_OS_SEEK_SET
	CONST_OS_SEEK_CUR
	CONST_OS_SEEK_END
)

const (
	OsRunSuccess = iota
	OsRunEmptyCommand
	OsRunPanic // 2
	OsRunStartFailed
	OsRunWaitFailed
	OsRunTimeout
)

// ConstantsInit registers the os.* constants through constantAdd,
// actions.ConstI32Add in cx. Taking it as an argument keeps the compiler out
// of this package, so the cx build core can link the file library.
func ConstantsInit(constantAdd func(code int, name string, value int32)) {
	// os
	constantAdd(CONST_OS_RUN_SUCCESS, "os.RUN_SUCCESS", OsRunSuccess)
	constantAdd(CONST_OS_RUN_EMPTY_CMD, "os.RUN_EMPTY_CMD", OsRunEmptyCommand)
	constantAdd(CONST_OS_RUN_PANIC, "os.RUN_PANIC", OsRunPanic)
	constantAdd(CONST_OS_RUN_START_FAILED, "os.RUN_START_FAILED", OsRunStartFailed)
	constantAdd(CONST_OS_RUN_WAIT_FAILED, "os.RUN_WAIT_FAILED", OsRunWaitFailed)
	constantAdd(CONST_OS_RUN_TIMEOUT, "os.RUN_TIMEOUT", OsRunTimeout)

	// os.FileModes (FIXME: these are uint32 in Go, with _DIR & _TYPE exceeding math.MaxInt32)
	//cxcore.ConstI32Add(CONST_OS_FILEMODE_DIR, "os.ModeDir", 2147483648)
	//cxcore.ConstI32Add(CONST_OS_FILEMODE_TYPE, "os.ModeType", 2399666176)
	constantAdd(CONST_OS_FILEMODE_APPEND, "os.ModeAppend", 1073741824)
	constantAdd(CONST_OS_FILEMODE_EXCLUSIVE, "os.ModeExclusive", 536870912)
	constantAdd(CONST_OS_FILEMODE_TEMPORARY, "os.ModeTemporary", 268435456)
	constantAdd(CONST_OS_FILEMODE_SYMLINK, "os.ModeSymlink", 134217728)
	constantAdd(CONST_OS_FILEMODE_DEVICE, "os.ModeDevice", 67108864)
	constantAdd(CONST_OS_FILEMODE_NAMED_PIPE, "os.ModeNamedPipe", 33554432)
	constantAdd(CONST_OS_FILEMODE_SOCKET, "os.ModeSocket", 16777216)
	constantAdd(CONST_OS_FILEMODE_SETUID, "os.ModeSetuid", 8388608)
	constantAdd(CONST_OS_FILEMODE_SETGID, "os.ModeSetgid", 4194304)
	constantAdd(CONST_OS_FILEMODE_CHAR_DEVICE, "os.ModeCharDevice", 2097152)
	constantAdd(CONST_OS_FILEMODE_STICKY, "os.ModeSticky", 1048576)
	constantAdd(CONST_OS_FILEMODE_IRREGULAR, "os.ModeIrregular", 524288)
	constantAdd(CONST_OS_FILEMODE_PERM, "os.ModePerm", 511)

	constantAdd(CONST_OS_SEEK_SET, "os.SEEK_SET", OsSeekSet)
	constantAdd(CONST_OS_SEEK_CUR, "os.SEEK_CUR", OsSeekCurrent)
	constantAdd(CONST_OS_SEEK_END, "os.SEEK_END", OsSeekEnd)

}
