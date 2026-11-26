package ast

import (
	"runtime/debug"
	"strconv"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
	"github.com/skycoin/skycoin-cx/t2-cx-core-environment" // thin: only DebugGolangStackTrace, HeapOffset; core globals, keep
)

// RuntimeErrorRecord is a runtime error of the old loop as data. The loop
// fills it without building text and the printer package prints it after
// the loop has returned, see #413.
type RuntimeErrorRecord struct {
	// RuntimeErrorRecordRaised is set once a runtime error ended the loop.
	RuntimeErrorRecordRaised bool
	// RuntimeErrorRecordCode is the CX exit code, a CX_RUNTIME_* or CXAssert.
	RuntimeErrorRecordCode int
	// RuntimeErrorRecordPanicValue is the value the panic carried. A Go
	// runtime error formats its own message only when it is printed.
	RuntimeErrorRecordPanicValue interface{}
	// RuntimeErrorRecordCall is the call stack position of the failure.
	RuntimeErrorRecordCall valuetypes.Pointer
	// RuntimeErrorRecordLine is the expression index in that call.
	RuntimeErrorRecordLine int
	// RuntimeErrorRecordStackPrint is false for a stack overflow at the
	// entry point, which prints no call stack.
	RuntimeErrorRecordStackPrint bool
	// RuntimeErrorRecordAssertLine is set by panicIf, which prints the line
	// and the message in RuntimeErrorRecordAssertMessage before the error.
	RuntimeErrorRecordAssertLine bool
	// RuntimeErrorRecordAssertMessage is the heap offset of the CX string.
	RuntimeErrorRecordAssertMessage valuetypes.Pointer
}

// ErrorString ...
func ErrorString(code int) string {
	if str, found := opcodetable.ErrorStrings[code]; found {
		return str
	}
	return opcodetable.ErrorStrings[valuetypes.CXRuntimeError]
}

// RuntimeOutOfMemory is the panic value of an allocation of
// RuntimeOutOfMemoryRequest bytes that does not fit the heap cap. Its
// message is what the runtime error prints.
type RuntimeOutOfMemory struct {
	RuntimeOutOfMemoryRequest uint64
	// RuntimeOutOfMemoryTaken is the heap pointer, garbage included, as the
	// collector runs only between two steps.
	RuntimeOutOfMemoryTaken valuetypes.Pointer
	RuntimeOutOfMemoryCap   valuetypes.Pointer
}

func (outOfMemory RuntimeOutOfMemory) Error() string {
	return "out of memory: " + strconv.FormatUint(outOfMemory.RuntimeOutOfMemoryRequest, 10) + " bytes requested with " +
		strconv.FormatUint(uint64(outOfMemory.RuntimeOutOfMemoryTaken), 10) + " taken do not fit the heap cap of " +
		strconv.FormatUint(uint64(outOfMemory.RuntimeOutOfMemoryCap), 10) + " bytes, raise it with -heap-max"
}

func errorCode(r interface{}) int {
	switch v := r.(type) {
	case int:
		return int(v)
	case RuntimeOutOfMemory:
		return valuetypes.CXRuntimeHeapExhaustedError
	case RuntimeLibraryDisabled:
		return valuetypes.CXRuntimeLibraryDisabled
	case RuntimeOpcodeUnlinked:
		return valuetypes.CXRuntimeNotImplemented
	default:
		return valuetypes.CXRuntimeError
	}
}

// runtimeErrorRecordSet records the panic value r with the call the program
// stopped in. stackPrint is false for an overflow at the entry point.
func runtimeErrorRecordSet(prgrm *CXProgram, r interface{}, stackPrint bool, defaultError int) {
	call := prgrm.CallStack[prgrm.CallCounter]
	code := errorCode(r)
	if code == valuetypes.CXRuntimeError {
		code = defaultError
	}

	record := &prgrm.RuntimeError
	record.RuntimeErrorRecordRaised = true
	record.RuntimeErrorRecordCode = code
	record.RuntimeErrorRecordPanicValue = r
	record.RuntimeErrorRecordCall = prgrm.CallCounter
	record.RuntimeErrorRecordLine = call.Line
	record.RuntimeErrorRecordStackPrint = stackPrint

	// The Go stack of the panic exists only here, so it is printed here.
	if environment.DebugGolangStackTrace {
		debug.PrintStack()
	}
}

// RuntimeError is deferred by the run loop. It turns a panic into a
// RuntimeErrorRecord on prgrm and lets the loop return.
func RuntimeError(prgrm *CXProgram) {
	if r := recover(); r != nil {
		switch r {
		case opcodetable.StackOverflowError:
			// callPush never pushes a frame that does not fit, so the caller is
			// the current call. Only the entry frame leaves the pointer past the stack.
			stackPrint := prgrm.Stack.Pointer <= opcodetable.StackSize
			runtimeErrorRecordSet(prgrm, r, stackPrint, valuetypes.CXRuntimeStackOverflowError)
		default:
			runtimeErrorRecordSet(prgrm, r, true, valuetypes.CXRuntimeError)
		}
	}
}
