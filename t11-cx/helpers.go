package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	filelog "github.com/skycoin/skycoin-cx/t0-cx-runtime-filelog"
	types "github.com/skycoin/skycoin-cx/t0-types"
	opcodes "github.com/skycoin/skycoin-cx/t4-opcodes" // thin: only AssertFailed; no depth gain

	constants "github.com/skycoin/skycoin-cx/t1-constants"
	profiling "github.com/skycoin/skycoin-cx/t1-profiling"
	repl "github.com/skycoin/skycoin-cx/t10-cxrepl" // thin: only Repl, TargetFunction; only user of cxrepl, could merge it here, max tier 11 to 10
	globals "github.com/skycoin/skycoin-cx/t2-globals"
	ast "github.com/skycoin/skycoin-cx/t3-ast"
	actions "github.com/skycoin/skycoin-cx/t4-actions"
	printer "github.com/skycoin/skycoin-cx/t4-cx-runtime-printer"
	cxinit "github.com/skycoin/skycoin-cx/t7-init"             // thin: only MakeProgram, RegisterPackages; no depth gain
	cxgo "github.com/skycoin/skycoin-cx/t8-cxparsingcompletor" // thin: only Tokenize; no depth gain
	cxparsing "github.com/skycoin/skycoin-cx/t9-cxparsing"     // thin: only AddInitFunction, ErrCompilation, ParseSourceStrings; no depth gain
)

func parseProgram(options cxCmdFlags, fileNames []string, sourceCode []*os.File) bool {

	stopCPUProfile, err := profiling.CPUProfileStart("parse", profiling.DebugProfileRate)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		stopCPUProfile = func() error { return nil }
	}
	defer func() {
		if err := stopCPUProfile(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	if profiling.DebugProfile {
		defer func() {
			if err := profiling.DumpMemProfile("parse"); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}()
	}

	profiling.ProfileStart("parse")
	defer profiling.ProfileStop("parse")
	defer func() { printer.CompileErrorsPrint(actions.AST) }()

	if actions.AST == nil {
		// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
		// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S0056 S0057.
		actions.AST = cxinit.ProgramMake()
		cxinit.PackagesRegister(actions.AST)
	}

	sources := make([]string, len(sourceCode))
	for i, file := range sourceCode {
		source, err := io.ReadAll(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			profiling.CleanupAndExit(types.CXCompilationError)
		}
		sources[i] = string(source)
	}

	// Setting what function to start in if using the REPL.
	repl.TargetFunction = constants.MainFunction

	if err := compileSourceStrings(fileNames, sources); err != nil {
		if err != cxparsing.ErrCompilation {
			fmt.Fprintln(os.Stdout, "error:", err)
		}
		profiling.CleanupAndExit(types.CXCompilationError)
	}

	return true
}

// compileSourceStrings compiles sources, named by fileNames, into actions.AST
// with its *init. The cx command and the tests compile through it (#516).
func compileSourceStrings(fileNames []string, sources []string) error {
	if err := cxparsing.SourceStringsParse(sources, fileNames); err != nil {
		return err
	}

	// Checking if a main package exists. If not, create and add it to `AST`.
	if _, err := actions.AST.FunctionGet(actions.AST.CxIrState.CxIrStateNameMain, actions.AST.CxIrState.CxIrStateNameMain); err != nil {
		initMainPkg(actions.AST)
	}

	// Adding *init function that initializes all the global variables.
	if err := cxparsing.SysInitFunctionAdd(actions.AST); err != nil {
		return fmt.Errorf("initializing *init: %w", err)
	}

	actions.LineNo = 0

	if globals.FoundCompileErrors {
		printer.CompileErrorsPrint(actions.AST)
		return cxparsing.ErrCompilation
	}
	return nil
}

// initMainPkg adds a `main` package with an empty `main` function to `prgrm`.
func initMainPkg(prgrm *ast.CXProgram) {
	// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
	// Now a constructor call, site S0058.
	mod := ast.PackageMake(prgrm.CxIrState.CxIrStateNameMain)
	// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
	// Now an arena insert, site S0059.
	modIdx := prgrm.PackageAdd(mod)
	// OBJMODEL: state.CXPackageMng.CXPackageGet(gid), phase 5 #362.
	// Now an arena getter, site S0060.
	modPkg, err := prgrm.PackageFromArrayGet(modIdx)
	if err != nil {
		panic(err)
	}
	_, fn := ast.FunctionMake(prgrm, prgrm.CxIrState.CxIrStateNameMain, actions.CurrentFile, actions.LineNo)
	modPkg.FunctionAdd(prgrm, fn)
}

// optionTokenize checks if the user wants to use CX to generate the lexer tokens
func printTokenize(options cxCmdFlags, fileNames []string) {
	var r *os.File
	var w *os.File
	var err error
	var lines []string
	var sourceFilename string

	if len(fileNames) == 0 {
		r = os.Stdin
	} else {
		sourceFilename = fileNames[0]
		if len(fileNames) > 1 {
			fmt.Fprintln(os.Stderr, "Multiple source files detected. Ignoring all except", sourceFilename)
		}
		r, err = filelog.Open(sourceFilename)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ProgramError reading:", sourceFilename, err)
			return
		}
		defer r.Close()

		bytesRead, err := os.ReadFile(sourceFilename)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ProgramError reading:", sourceFilename, err)
			return
		}

		file_content := string(bytesRead)
		lines = strings.Split(file_content, "\n")
	}

	if options.lexerOutput == "" {
		w = os.Stdout
	} else {
		tokenFilename := options.lexerOutput
		w, err = filelog.Create(tokenFilename)
		if err != nil {
			fmt.Fprintln(os.Stderr, "ProgramError writing:", tokenFilename, err)
			return
		}
		defer w.Close()
	}

	cxgo.Tokenize(r, w, lines, sourceFilename)
}

func printProgramAST(options cxCmdFlags, cxArgs []string, sourceCode []*os.File) {
	profiling.ProfileStart("run")
	defer profiling.ProfileStop("run")

	if options.replMode || len(sourceCode) == 0 {
		repl.Repl()
		return
	}

	// Print CX program.
	actions.AST.ProgramPrint()

	if opcodes.AssertFailed() {
		os.Exit(types.CXAssert)
	}
}

// Used for the -heap-initial, -heap-max and -stack-size flags.
// This function parses, for example, "1M" to 1048576 (the corresponding number of bytes)
// Possible suffixes are: G or g (gigabytes), M or m (megabytes), K or k (kilobytes)
func parseMemoryString(s string) types.Pointer {
	suffix := s[len(s)-1]
	_, notSuffix := strconv.ParseFloat(string(suffix), 64)

	if notSuffix == nil {
		// then we don't have a suffix
		num, err := strconv.ParseInt(s, 10, 64)

		if err != nil {
			// malformed size
			return types.InvalidPointer
		}

		return types.I64ToPtrCast(num)
	} else {
		// then we have a suffix
		num, err := strconv.ParseFloat(s[:len(s)-1], 64)

		if err != nil {
			// malformed size
			return types.InvalidPointer
		}

		// The user can use suffixes to give as input gigabytes, megabytes or kilobytes.
		switch suffix {
		case 'G', 'g':
			return types.F64ToPtrCast(num * 1073741824)
		case 'M', 'm':
			return types.F64ToPtrCast(num * 1048576)
		case 'K', 'k':
			return types.F64ToPtrCast(num * 1024)
		default:
			return types.InvalidPointer
		}
	}
}

type SourceCode struct {
	Code string //Unused?
}

// CXPathSet sets the bin, pkg and src paths of globals from --cxpath, else
// CXPATH, else cx/ in the home directory. It returns an error when none is
// known, as on js/wasm without HOME, and then leaves the paths unset.
func CXPathSet(options cxCmdFlags) error {
	cxPath := os.Getenv("CXPATH")
	if options.cxpath != "" {
		absolute, err := filepath.Abs(options.cxpath)
		if err != nil {
			return fmt.Errorf("--cxpath %s: %w", options.cxpath, err)
		}
		cxPath = absolute
	}
	if cxPath == "" {
		home, err := cxPathHomeDirectory()
		if err != nil {
			return fmt.Errorf("CXPATH is not set and there is no home directory: %w", err)
		}
		cxPath = filepath.Join(home, "cx")
	}
	globals.BINPATH = filepath.Join(cxPath, "bin/")
	globals.PKGPATH = filepath.Join(cxPath, "pkg/")
	globals.SRCPATH = filepath.Join(cxPath, "src/")
	return nil
}

// cxPathHomeDirectory is the home directory of the user, from os/user and
// else from $HOME, since user.Current fails on js/wasm.
var cxPathHomeDirectory = func() (string, error) {
	if current, err := user.Current(); err == nil && current.HomeDir != "" {
		return current.HomeDir, nil
	}
	return os.UserHomeDir()
}
