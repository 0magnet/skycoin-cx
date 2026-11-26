package printer

import (
	"fmt"
	"runtime/debug"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t2-cx-core-environment"
	"github.com/skycoin/skycoin-cx/t2-cx-runtime-ir"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	vm "github.com/skycoin/skycoin-cx/t3-cx-runtime-vm"
)

// VmStopMessage is what vm prints after the error string for stop.
func VmStopMessage(stop vm.Stop) string {
	return vm.StopMessages[stop]
}

// printerVmStopLine returns the file and line of the instruction machine
// stopped at, "" and 0 when it has none, as PreviousCXLineGet gives them.
func printerVmStopLine(machine *vm.Machine) (string, int) {
	image := machine.MachineImage
	line, found := ir.LineAt(image, machine.MachineFunction, machine.MachinePc)
	if !found {
		return "", 0
	}
	return image.ImageFileNames[line.LineFile], int(line.LineNumber)
}

// VmTestFailurePrint prints a failed test as assert in t4-opcodes
// does.
func VmTestFailurePrint(machine *vm.Machine) {
	fileName, lineNumber := printerVmStopLine(machine)
	message := machine.TestMessage()
	if len(message) > 0 {
		fmt.Printf("%s: %d: result was not equal to the expected value; %s\n", fileName, lineNumber, message)
	} else {
		fmt.Printf("%s: %d: result was not equal to the expected value\n", fileName, lineNumber)
	}
}

// VmErrorPrint prints the runtime error vm stopped with as the old
// loop prints it, the call stack included. StackPrint reads the old
// program state, so the vm memory and frames are copied into prgrm first.
func VmErrorPrint(prgrm *ast.CXProgram, machine *vm.Machine, stop vm.Stop) {
	image := machine.MachineImage
	fileName, lineNumber := printerVmStopLine(machine)
	code := int(vm.StopExitCodes[stop])
	message := vm.StopMessages[stop]
	fmt.Printf("%s, %s, %v", ErrorHeader(fileName, lineNumber), ast.ErrorString(code), message)

	copy(prgrm.Memory, machine.MachineMemory[:ir.ImageDataEnd(image)])
	depth := machine.MachineDepth
	for i := uint32(0); i < depth; i++ {
		frame := &machine.MachineFrames[i]
		prgrm.CallStack[i].Operator = prgrm.CXProgramFunctionSlotGid(frame.FrameFunction)
		prgrm.CallStack[i].FramePointer = valuetypes.Pointer(frame.FramePointer)
	}
	prgrm.CallStack[depth].Operator = prgrm.CXProgramFunctionSlotGid(machine.MachineFunction)
	prgrm.CallStack[depth].FramePointer = valuetypes.Pointer(machine.MachineFramePointer)
	prgrm.CallCounter = valuetypes.Pointer(depth)
	prgrm.Stack.Pointer = valuetypes.Pointer(machine.MachineStackPointer)
	StackPrint(prgrm)

	if environment.DebugGolangStackTrace {
		debug.PrintStack()
	}
}
