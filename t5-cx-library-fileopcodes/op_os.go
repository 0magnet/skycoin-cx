//go:build !nocxos

package fileopcodes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t0-cx-library-filetable"
	filelog "github.com/skycoin/skycoin-cx/t0-cx-runtime-filelog"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

const (
	OsSeekSet = iota
	OsSeekCurrent
	OsSeekEnd
)

// ValidFile returns the open file of handle, nil for one that is not open.
func ValidFile(handle int32) *os.File {
	return filetable.FileHandleOf(handle)
}

func OpOsLogFile(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	filelog.Enable(inputs[0].BoolGet(prgrm))
}

// VIOLATION STRHOT-131: B, hands a path or text to Go or the OS, #412.
func OpOsReadAllText(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false

	if byts, fileError := ast.ProgramSandboxFileRead(prgrm, &inputs[0]); fileError == nil {
		outputs[0].StrSet(prgrm, string(byts))
		success = true
	}

	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsOpen(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	handle := int32(-1)
	if file, fileError := ast.ProgramSandboxFileOpen(prgrm, &inputs[0]); fileError == nil {
		handle = filetable.FileHandleTake(file)
	}

	outputs[0].I32Set(prgrm, int32(handle))
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsCreate(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	handle := int32(-1)
	if file, fileError := ast.ProgramSandboxFileCreate(prgrm, &inputs[0]); fileError == nil {
		handle = filetable.FileHandleTake(file)
	}

	outputs[0].I32Set(prgrm, int32(handle))
}

// OpOsOpenReadWrite is os.OpenReadWrite(path str) (handle i32): the file at
// path opened for reading and writing, created when it does not exist, or -1.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsOpenReadWrite(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	handle := int32(-1)
	if file, fileError := ast.ProgramSandboxFileOpenReadWrite(prgrm, &inputs[0]); fileError == nil {
		handle = filetable.FileHandleTake(file)
	}
	outputs[0].I32Set(prgrm, handle)
}

// OpOsRemove is os.Remove(path str) (removed bool): removes the file or empty
// directory at path inside the sandbox.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsRemove(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].BoolSet(prgrm, ast.ProgramSandboxFileRemove(prgrm, &inputs[0]) == nil)
}

func OpOsClose(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].BoolSet(prgrm, filetable.FileHandleClose(inputs[0].I32Get(prgrm)))
}

// OpOsSync is os.Sync(handle i32) (synced bool): commits what was written to
// the file of handle to stable storage.
func OpOsSync(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].BoolSet(prgrm, filetable.FileHandleSync(inputs[0].I32Get(prgrm)))
}

func OpOsSeek(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	outputs[0].I64Set(prgrm, filetable.FileHandleSeek(inputs[0].I32Get(prgrm), inputs[1].I64Get(prgrm), inputs[2].I32Get(prgrm)))
}

// VIOLATION STRHOT-134: B, hands a path or text to Go or the OS, #412.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadStr(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var len uint64
	var value string
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Read(file, binary.LittleEndian, &len); err == nil {
			bytes := make([]byte, len)
			if err := binary.Read(file, binary.LittleEndian, &bytes); err == nil {
				value = string(bytes)
				success = true
			}
		}
	}
	outputs[0].StrSet(prgrm, value)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadF64(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var value float64
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Read(file, binary.LittleEndian, &value); err == nil {
			success = true
		}
	}

	outputs[0].F64Set(prgrm, value)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadF32(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var value float32
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Read(file, binary.LittleEndian, &value); err == nil {
			success = true
		}
	}

	outputs[0].F32Set(prgrm, value)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadUI64(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var value uint64
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Read(file, binary.LittleEndian, &value); err == nil {
			success = true
		}
	}

	outputs[0].UI64Set(prgrm, value)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadUI32(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var value uint32
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Read(file, binary.LittleEndian, &value); err == nil {
			success = true
		}
	}

	outputs[0].UI32Set(prgrm, value)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadUI16(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var value uint16
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Read(file, binary.LittleEndian, &value); err == nil {
			success = true
		}
	}

	outputs[0].UI16Set(prgrm, value)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadUI8(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var value uint8
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Read(file, binary.LittleEndian, &value); err == nil {
			success = true
		}
	}

	outputs[0].UI8Set(prgrm, value)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadI64(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var value int64
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Read(file, binary.LittleEndian, &value); err == nil {
			success = true
		}
	}

	outputs[0].I64Set(prgrm, value)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadI32(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var value int32
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Read(file, binary.LittleEndian, &value); err == nil {
			success = true
		}
	}

	outputs[0].I32Set(prgrm, value)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadI16(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var value int16
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Read(file, binary.LittleEndian, &value); err == nil {
			success = true
		}
	}

	outputs[0].I16Set(prgrm, value)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadI8(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var value int8
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Read(file, binary.LittleEndian, &value); err == nil {
			success = true
		}
	}

	outputs[0].I8Set(prgrm, value)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadBOOL(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var value bool
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Read(file, binary.LittleEndian, &value); err == nil {
			success = true
		}
	}

	outputs[0].BoolSet(prgrm, value)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION STRHOT-135: B, hands a path or text to Go or the OS, #412.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteStr(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		value := inputs[1].StrGet(prgrm)
		len := len(value)
		if err := binary.Write(file, binary.LittleEndian, uint64(len)); err == nil {
			if err := binary.Write(file, binary.LittleEndian, []byte(value)); err == nil {
				success = true
			}
		}
	}

	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteF64(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Write(file, binary.LittleEndian, inputs[1].F64Get(prgrm)); err == nil {
			success = true
		}
	}

	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteF32(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Write(file, binary.LittleEndian, inputs[1].F32Get(prgrm)); err == nil {
			success = true
		}
	}

	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteUI64(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Write(file, binary.LittleEndian, inputs[1].UI64Get(prgrm)); err == nil {
			success = true
		}
	}

	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteUI32(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Write(file, binary.LittleEndian, inputs[1].I32Get(prgrm)); err == nil {
			success = true
		}
	}

	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteUI16(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Write(file, binary.LittleEndian, inputs[1].UI16Get(prgrm)); err == nil {
			success = true
		}
	}

	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteUI8(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Write(file, binary.LittleEndian, inputs[1].UI8Get(prgrm)); err == nil {
			success = true
		}
	}

	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteI64(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Write(file, binary.LittleEndian, inputs[1].I64Get(prgrm)); err == nil {
			success = true
		}
	}

	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteI32(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Write(file, binary.LittleEndian, inputs[1].I32Get(prgrm)); err == nil {
			success = true
		}
	}

	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteI16(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Write(file, binary.LittleEndian, inputs[1].I16Get(prgrm)); err == nil {
			success = true
		}
	}

	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteI8(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Write(file, binary.LittleEndian, inputs[1].I8Get(prgrm)); err == nil {
			success = true
		}
	}

	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteBOOL(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if err := binary.Write(file, binary.LittleEndian, inputs[1].BoolGet(prgrm)); err == nil {
			success = true
		}
	}

	outputs[0].BoolSet(prgrm, success)
}

func getSlice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) (outputSlicePointer valuetypes.Pointer, outputSliceOffset valuetypes.Pointer, sizeofElement valuetypes.Pointer, count valuetypes.Pointer) {
	var inp1, out0 *ast.CXArgument

	var inpType, outType valuetypes.Code
	var inpIsSlice, outIsSlice bool
	var inpSize valuetypes.Pointer
	if inputs[1].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1388.
		inp1 = prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(inputs[1].TypeSignature.Meta))

		inpType = inp1.Type
		inpIsSlice = (inp1.AssignmentElementGet(prgrm)).SliceTyped()
		inpSize = (inp1.AssignmentElementGet(prgrm)).Size
	} else if inputs[1].TypeSignature.Type == ast.TypeSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S1389.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(inputs[1].TypeSignature.Meta)
		inpType = valuetypes.Code(sliceDetails.Meta)
		inpIsSlice = true
		inpSize = inpType.Size()
	} else {
		panic("type is not type cx argument deprecate\n\n")
	}

	if outputs[0].TypeSignature.Type == ast.TypeCXArgumentDeprecate {
		// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
		// Now an arena getter, site S1390.
		out0 = prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(outputs[0].TypeSignature.Meta))

		outType = out0.Type
		outIsSlice = (out0.AssignmentElementGet(prgrm)).SliceTyped()
	} else if outputs[0].TypeSignature.Type == ast.TypeSliceAtomic {
		// OBJMODEL: state.CXTypeSignature_ArrayMng.CXTypeSignature_ArrayGet(gid), phase 5 #317.
		// Now an arena getter, site S1391.
		sliceDetails := prgrm.CXTypeSignatureArrayFromArrayGet(outputs[0].TypeSignature.Meta)

		outType = valuetypes.Code(sliceDetails.Meta)
		outIsSlice = true
	} else {
		panic("type is not type cx argument deprecate\n\n")
	}

	if inpType != outType || !inpIsSlice || !outIsSlice {
		panic(valuetypes.CXRuntimeInvalidArgument)
	}

	outputSlicePointer = outputs[0].Offset
	sizeofElement = inpSize
	// The count is a ui64. One past what the heap can hold reads nothing.
	count64 := inputs[2].UI64Get(prgrm)
	if count64 > uint64(ast.SliceElementLimit(prgrm, sizeofElement)) {
		outputSliceOffset = ast.SliceOffsetGet(prgrm, inputs[1].FramePointer, inputs[1].TypeSignature)
		return
	}
	count = valuetypes.Pointer(count64)
	outputSliceOffset = ast.SliceResize(prgrm, outputs[0].FramePointer, outputs[0].TypeSignature, inputs[1].TypeSignature, count, sizeofElement)
	return
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadF64Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	_, outputSliceOffset, sizeofElement, count := getSlice(prgrm, inputs, outputs)
	if count > 0 {
		if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
			values := make([]float64, count)
			if err := binary.Read(file, binary.LittleEndian, values); err == nil {
				success = true
				outputSliceElements := ast.SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
				for i := valuetypes.Pointer(0); i < count; i++ {
					valuetypes.F64Write(outputSliceElements, i*sizeofElement, values[i])
				}
			}
		}
	}

	outputs[0].PtrSet(prgrm, outputSliceOffset)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadF32Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	_, outputSliceOffset, sizeofElement, count := getSlice(prgrm, inputs, outputs)
	if count > 0 {
		if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
			values := make([]float32, count)
			if err := binary.Read(file, binary.LittleEndian, values); err == nil {
				success = true
				outputSliceElements := ast.SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
				for i := valuetypes.Pointer(0); i < count; i++ {
					valuetypes.F32Write(outputSliceElements, i*sizeofElement, values[i])
				}
			}
		}
	}

	outputs[0].PtrSet(prgrm, outputSliceOffset)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadUI64Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	_, outputSliceOffset, sizeofElement, count := getSlice(prgrm, inputs, outputs)
	if count > 0 {
		if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
			values := make([]uint64, count)
			if err := binary.Read(file, binary.LittleEndian, values); err == nil {
				success = true
				outputSliceElements := ast.SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
				for i := valuetypes.Pointer(0); i < count; i++ {
					valuetypes.UI64Write(outputSliceElements, i*sizeofElement, values[i])
				}
			}
		}
	}

	outputs[0].PtrSet(prgrm, outputSliceOffset)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadUI32Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	_, outputSliceOffset, sizeofElement, count := getSlice(prgrm, inputs, outputs)
	if count > 0 {
		if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
			values := make([]uint32, count)
			if err := binary.Read(file, binary.LittleEndian, values); err == nil {
				success = true
				outputSliceElements := ast.SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
				for i := valuetypes.Pointer(0); i < count; i++ {
					valuetypes.UI32Write(outputSliceElements, i*sizeofElement, values[i])
				}
			}
		}
	}

	outputs[0].PtrSet(prgrm, outputSliceOffset)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadUI16Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	_, outputSliceOffset, sizeofElement, count := getSlice(prgrm, inputs, outputs)
	if count > 0 {
		if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
			values := make([]uint16, count)
			if err := binary.Read(file, binary.LittleEndian, values); err == nil {
				success = true
				outputSliceElements := ast.SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
				for i := valuetypes.Pointer(0); i < count; i++ {
					valuetypes.UI16Write(outputSliceElements, i*sizeofElement, values[i])
				}
			}
		}
	}

	outputs[0].PtrSet(prgrm, outputSliceOffset)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadUI8Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	_, outputSliceOffset, sizeofElement, count := getSlice(prgrm, inputs, outputs)
	if count > 0 {
		if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
			values := make([]uint8, count)
			if err := binary.Read(file, binary.LittleEndian, values); err == nil {
				success = true
				outputSliceElements := ast.SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
				for i := valuetypes.Pointer(0); i < count; i++ {
					valuetypes.UI8Write(outputSliceElements, i*sizeofElement, values[i])
				}
			}
		}
	}

	outputs[0].PtrSet(prgrm, outputSliceOffset)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadI64Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	_, outputSliceOffset, sizeofElement, count := getSlice(prgrm, inputs, outputs)
	if count > 0 {
		if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
			values := make([]int64, count)
			if err := binary.Read(file, binary.LittleEndian, values); err == nil {
				success = true
				outputSliceElements := ast.SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
				for i := valuetypes.Pointer(0); i < count; i++ {
					valuetypes.I64Write(outputSliceElements, i*sizeofElement, values[i])
				}
			}
		}
	}

	outputs[0].PtrSet(prgrm, outputSliceOffset)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadI32Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	_, outputSliceOffset, sizeofElement, count := getSlice(prgrm, inputs, outputs)
	if count > 0 {
		if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
			values := make([]int32, count)
			if err := binary.Read(file, binary.LittleEndian, values); err == nil {
				success = true
				outputSliceElements := ast.SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
				for i := valuetypes.Pointer(0); i < count; i++ {
					valuetypes.I32Write(outputSliceElements, i*sizeofElement, values[i])
				}
			}
		}
	}

	outputs[0].PtrSet(prgrm, outputSliceOffset)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadI16Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	_, outputSliceOffset, sizeofElement, count := getSlice(prgrm, inputs, outputs)
	if count > 0 {
		if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
			values := make([]int16, count)
			if err := binary.Read(file, binary.LittleEndian, values); err == nil {
				success = true
				outputSliceElements := ast.SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
				for i := valuetypes.Pointer(0); i < count; i++ {
					valuetypes.I16Write(outputSliceElements, i*sizeofElement, values[i])
				}
			}
		}
	}

	outputs[0].PtrSet(prgrm, outputSliceOffset)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsReadI8Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	_, outputSliceOffset, sizeofElement, count := getSlice(prgrm, inputs, outputs)
	if count > 0 {
		if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
			values := make([]int8, count)
			if err := binary.Read(file, binary.LittleEndian, values); err == nil {
				success = true
				outputSliceElements := ast.SliceElementsGet(prgrm, outputSliceOffset, sizeofElement)
				for i := valuetypes.Pointer(0); i < count; i++ {
					valuetypes.I8Write(outputSliceElements, i*sizeofElement, values[i])
				}
			}
		}
	}

	outputs[0].PtrSet(prgrm, outputSliceOffset)
	outputs[1].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteF64Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if elements := inputs[1].SliceF64Get(prgrm); elements != nil {
			if err := binary.Write(file, binary.LittleEndian, elements); err == nil {
				success = true
			}
		}
	}
	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteF32Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if elements := inputs[1].SliceF32Get(prgrm); elements != nil {
			if err := binary.Write(file, binary.LittleEndian, elements); err == nil {
				success = true
			}
		}
	}
	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteUI64Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if elements := inputs[1].SliceF64Get(prgrm); elements != nil {
			if err := binary.Write(file, binary.LittleEndian, elements); err == nil {
				success = true
			}
		}
	}
	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteUI32Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if elements := inputs[1].SliceUI32Get(prgrm); elements != nil {
			if err := binary.Write(file, binary.LittleEndian, elements); err == nil {
				success = true
			}
		}
	}
	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteUI16Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if elements := inputs[1].SliceUI16Get(prgrm); elements != nil {
			if err := binary.Write(file, binary.LittleEndian, elements); err == nil {
				success = true
			}
		}
	}
	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteUI8Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if elements := inputs[1].SliceUI8Get(prgrm); elements != nil {
			if err := binary.Write(file, binary.LittleEndian, elements); err == nil {
				success = true
			}
		}
	}
	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteI64Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if elements := inputs[1].SliceI64Get(prgrm); elements != nil {
			if err := binary.Write(file, binary.LittleEndian, elements); err == nil {
				success = true
			}
		}
	}
	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteI32Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if elements := inputs[1].SliceI32Get(prgrm); elements != nil {
			if err := binary.Write(file, binary.LittleEndian, elements); err == nil {
				success = true
			}
		}
	}
	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteI16Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if elements := inputs[1].SliceI16Get(prgrm); elements != nil {
			if err := binary.Write(file, binary.LittleEndian, elements); err == nil {
				success = true
			}
		}
	}
	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsWriteI8Slice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	success := false
	if file := ValidFile(inputs[0].I32Get(prgrm)); file != nil {
		if elements := inputs[1].SliceI8Get(prgrm); elements != nil {
			if err := binary.Write(file, binary.LittleEndian, elements); err == nil {
				success = true
			}
		}
	}
	outputs[0].BoolSet(prgrm, success)
}

// VIOLATION STRHOT-136: B, hands a path or text to Go or the OS, #412.
func OpOsGetWorkingDirectory(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	wd, err := os.Getwd()
	if err == nil && !strings.HasSuffix(wd, string(os.PathSeparator)) {
		wd += string(os.PathSeparator)
	}
	outputs[0].StrSet(prgrm, wd)
}

// VIOLATION STRHOT-137: B, hands a path or text to Go or the OS, #412.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func OpOsRun(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	var runError int32 = OsRunSuccess

	command := inputs[0].StrGet(prgrm)
	dir := inputs[3].StrGet(prgrm)
	args, err := splitCommand(command)
	if err != nil || len(args) == 0 {
		runError = OsRunEmptyCommand
		if err != nil {
			runError = OsRunStartFailed
		}
		outputs[0].I32Set(prgrm, runError)
		outputs[1].I32Set(prgrm, 0)
		outputs[2].StrSet(prgrm, "")
		return
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	var cmdError int32 = 0

	timeoutMs := inputs[2].I32Get(prgrm)
	timeout := time.Duration(math.MaxInt64)
	if timeoutMs > 0 {
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}

	if err := cmd.Start(); err != nil {
		runError = OsRunStartFailed
	} else {
		done := make(chan error)
		go func() { done <- cmd.Wait() }()

		select {
		case <-time.After(timeout):
			cmd.Process.Kill()
			runError = OsRunTimeout
		case err := <-done:
			if err != nil {
				if exiterr, ok := err.(*exec.ExitError); ok {
					// from stackoverflow
					// The program has exited with an exit code != 0
					// This works on both Unix and Windows. Although package
					// syscall is generally platform dependent, WaitStatus is
					// defined for both Unix and Windows and in both cases has
					// an ExitStatus() method with the same signature.
					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
						cmdError = int32(status.ExitStatus())
					}
				} else {
					runError = OsRunWaitFailed
				}
			}
		}
	}

	stdOutBytes := out.Bytes()
	maxSize := inputs[1].I32Get(prgrm)
	if (maxSize > 0) && (len(stdOutBytes) > int(maxSize)) {
		stdOutBytes = stdOutBytes[0:maxSize]
	}

	outputs[0].I32Set(prgrm, runError)
	outputs[1].I32Set(prgrm, cmdError)
	outputs[2].StrSet(prgrm, string(stdOutBytes))
}

// splitCommand splits an os.Run command into words at spaces and tabs.
// Single or double quotes group a word. Backslashes are literal, so
// Windows paths need no escaping.
// VIOLATION STRHOT-138: B, hands a path or text to Go or the OS, #412.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func splitCommand(command string) ([]string, error) {
	var args []string
	var word strings.Builder
	inWord := false
	var quote rune
	for _, c := range command {
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			} else {
				word.WriteRune(c)
			}
		case c == '\'' || c == '"':
			quote = c
			inWord = true
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			if inWord {
				args = append(args, word.String())
				word.Reset()
				inWord = false
			}
		default:
			word.WriteRune(c)
			inWord = true
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated %c quote", quote)
	}
	if inWord {
		args = append(args, word.String())
	}
	return args, nil
}
