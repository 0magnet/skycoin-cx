// Command cx-core is the minimum interpreter core of cx build. It runs the
// serialized program appended to its own file and passes its command line
// arguments to the program as os.Args. A core with no program appended, as on
// js/wasm, runs cx-core --program FILE args. It has no parser and loads no
// source.
package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	types "github.com/skycoin/skycoin-cx/t0-types"
	constants "github.com/skycoin/skycoin-cx/t1-constants"
	sandbox "github.com/skycoin/skycoin-cx/t1-cx-runtime-sandbox"
	ast "github.com/skycoin/skycoin-cx/t3-ast"
	printer "github.com/skycoin/skycoin-cx/t4-cx-runtime-printer" // thin: only PrinterCompileErrorsRegister; no depth gain
	core "github.com/skycoin/skycoin-cx/t5-cx-runtime-core"
)

// coreLibraryRegisters register the system libraries linked into this core.
// The library_*.go files add them by build tag, cxcorelibraryfile,
// cxcorelibrarynet, cxcorelibraryprocess and cxcorelibrarycipher.
var coreLibraryRegisters []func(prgrm *ast.CXProgram)

// coreLibraryOpcodesLinked marks the libraries whose functions are opcodes
// that this core links, by the same build tags (decision D119).
var coreLibraryOpcodesLinked = [constants.LibraryCount]bool{
	constants.LibraryNumberCipher: coreLibraryCipherLinked,
	constants.LibraryNumberNet:    coreLibraryNetLinked,
	constants.LibraryNumberFile:   coreLibraryFileLinked,
}

// coreOpcodeDispatch is the opcode switch of this core, the switch of
// t5-cx-runtime-core and then the switches of the libraries it links.
func coreOpcodeDispatch(prgrm *ast.CXProgram, code int, inputs []ast.CXValue, outputs []ast.CXValue) bool {
	return core.CoreOpcodeDispatch(prgrm, code, inputs, outputs) || coreOpcodeDispatchCipher(prgrm, code, inputs, outputs) || coreOpcodeDispatchNet(prgrm, code, inputs, outputs) || coreOpcodeDispatchFile(prgrm, code, inputs, outputs)
}

func main() {
	os.Exit(int(coreRun(os.Args[1:])))
}

// coreRun loads the appended program and runs it with args as os.Args. It
// returns the exit code.
func coreRun(args []string) int32 {
	runtime.LockOSThread()
	runtime.GOMAXPROCS(2)
	core.ExecutorMemoryLimitsClamp()

	program, args, problem := coreProgramLoad(args)
	if problem != "" {
		fmt.Fprintln(os.Stderr, "cx-core:", problem)
		return types.CX_COMPILATION_ERROR
	}

	core.CoreRegister(core.CoreProgramMake(), coreLibraryRegisters...)
	ast.OpcodeDispatch = coreOpcodeDispatch
	prgrm, err := ast.Deserialize(program, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cx-core: %s\n", err)
		return types.CX_COMPILATION_ERROR
	}
	if missing := core.CoreOpcodesMissing(prgrm); len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "cx-core: the program uses opcodes this core does not include: %s\n", strings.Join(missing, ", "))
		return types.CX_COMPILATION_ERROR
	}
	core.CoreLibrariesEnabledResolve(prgrm)
	if missing := core.CoreLibrariesMissing(prgrm, coreLibraryOpcodesLinked); len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "cx-core: the program enables system libraries this core does not link: %s\n", strings.Join(missing, ", "))
		return types.CX_COMPILATION_ERROR
	}
	if core.CoreLibrariesActive(prgrm)[constants.LibraryNumberFile] {
		// The file library resolves paths in the working directory, as cx
		// does without --sandbox. The core has no flags to name another.
		root, fileError := sandbox.RootOpen("")
		if fileError != nil {
			fmt.Fprintf(os.Stderr, "cx-core: cannot open the sandbox: %s\n", fileError.FileErrorMessage)
			return types.CX_COMPILATION_ERROR
		}
		defer root.Close()
		prgrm.ProgramSandboxRoot = root
	}
	printer.PrinterCompileErrorsRegister(prgrm)
	prgrm.ProgramStdinInit(os.Stdin)
	return core.ExecutorRun(prgrm, core.ExecutorAuto, args)
}

// coreProgramFileFlag names a program file for a core with no program
// appended, cx-core --program FILE args, for hosts without os.Executable
// such as js/wasm. FILE is a program made with cx build.
const coreProgramFileFlag = "--program"

// coreProgramLoad returns the serialized program and the arguments of the
// program. The program appended to this executable comes first and gets
// every argument. Without one, --program FILE names it. problem says why
// there is no program.
func coreProgramLoad(args []string) (program []byte, programArgs []string, problem string) {
	path, err := os.Executable()
	if err == nil {
		appended, fileError := core.CoreTrailerRead(path)
		if fileError == nil && appended != nil {
			return appended, args, ""
		}
		if fileError != nil && (len(args) < 2 || args[0] != coreProgramFileFlag) {
			return nil, nil, fileError.FileErrorPath + ": " + fileError.FileErrorMessage
		}
	}
	if len(args) < 2 || args[0] != coreProgramFileFlag {
		if err != nil {
			return nil, nil, "cannot find this executable: " + err.Error() + ", name the program with " + coreProgramFileFlag + " FILE"
		}
		return nil, nil, "no CX program is appended to this file, make one with cx build or name one with " + coreProgramFileFlag + " FILE"
	}
	built, fileError := core.CoreTrailerRead(args[1])
	if fileError != nil {
		return nil, nil, fileError.FileErrorPath + ": " + fileError.FileErrorMessage
	}
	if built == nil {
		return nil, nil, args[1] + " holds no CX program, make one with cx build"
	}
	return built, args[2:], ""
}
