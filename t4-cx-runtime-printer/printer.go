// Package printer produces the text of runtime errors once a run has
// returned, from the data the old loop and vm record, see #413.
package printer

import (
	"fmt"
	"strconv"

	"github.com/skycoin/skycoin-cx/t0-types"
	"github.com/skycoin/skycoin-cx/t3-ast"
)

// ErrorHeader returns "error: file:line", the start of an error.
func ErrorHeader(fileName string, lineNumber int) string {
	return "error: " + fileName + ":" + strconv.FormatInt(int64(lineNumber), 10)
}

// RuntimeErrorPrint prints the runtime error recorded in
// prgrm.RuntimeError and returns its exit code. The record must be raised.
func RuntimeErrorPrint(prgrm *ast.CXProgram) int {
	record := &prgrm.RuntimeError
	call := prgrm.CallStack[record.RuntimeErrorRecordCall]
	cxLine, _ := prgrm.PreviousCXLineGet(prgrm.CxIrState.CXFunctionMng.CXFunctionGet(call.Operator).Expressions, record.RuntimeErrorRecordLine)
	fileName := prgrm.CXProgramSourceFileText(cxLine.FileName)

	if record.RuntimeErrorRecordAssertLine {
		fmt.Printf("%s : %d, %s\n", fileName, cxLine.LineNumber, printerHeapBytes(prgrm, record.RuntimeErrorRecordAssertMessage))
	}

	fmt.Printf("%s, %s, %v", ErrorHeader(fileName, cxLine.LineNumber), ast.ErrorString(record.RuntimeErrorRecordCode), record.RuntimeErrorRecordPanicValue)

	if record.RuntimeErrorRecordStackPrint {
		StackPrint(prgrm)
	}
	return record.RuntimeErrorRecordCode
}

// printerHeapBytes returns the bytes of the CX string at heapOffset, nil
// for the empty string, as types.ObjectRead reads them.
func printerHeapBytes(prgrm *ast.CXProgram, heapOffset types.Pointer) []byte {
	if heapOffset != 0 && heapOffset.Valid() {
		return types.ObjectPayloadRead(prgrm.Memory, heapOffset)
	}
	return nil
}

func OpDebugPrintStack(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	StackPrint(prgrm)
}
