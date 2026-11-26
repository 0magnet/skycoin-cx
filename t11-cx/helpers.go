package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	filelog "github.com/skycoin/skycoin-cx/t0-cx-runtime-filelog"
	"github.com/skycoin/skycoin-cx/t4-cx-runtime-opcodes" // thin: only AssertFailed; no depth gain

	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
	"github.com/skycoin/skycoin-cx/t1-cx-runtime-profiling"
	"github.com/skycoin/skycoin-cx/t10-cx-tools-repl" // thin: only Repl, TargetFunction; only user of cxrepl, could merge it here, max tier 11 to 10
	"github.com/skycoin/skycoin-cx/t2-cx-core-environment"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"github.com/skycoin/skycoin-cx/t4-cx-compiler-astbuild"
	printer "github.com/skycoin/skycoin-cx/t4-cx-runtime-printer"
	"github.com/skycoin/skycoin-cx/t7-cx-runtime-libraries"    // thin: only MakeProgram, RegisterPackages; no depth gain
	"github.com/skycoin/skycoin-cx/t8-cx-compiler-fullparsing" // thin: only Tokenize; no depth gain
	"github.com/skycoin/skycoin-cx/t9-cx-compiler-parsing"     // thin: only AddInitFunction, ErrCompilation, ParseSourceStrings; no depth gain
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
	defer func() { printer.CompileErrorsPrint(astbuild.AST) }()

	if astbuild.AST == nil {
		// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
		// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S0056 S0057.
		astbuild.AST = libraries.ProgramMake()
		libraries.PackagesRegister(astbuild.AST)
	}

	sources := make([]string, len(sourceCode))
	for i, file := range sourceCode {
		source, err := io.ReadAll(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			profiling.CleanupAndExit(valuetypes.CXCompilationError)
		}
		sources[i] = string(source)
	}

	// Setting what function to start in if using the REPL.
	repl.TargetFunction = opcodetable.MainFunction

	if err := compileSourceStrings(fileNames, sources); err != nil {
		if err != parsing.ErrCompilation {
			fmt.Fprintln(os.Stdout, "error:", err)
		}
		profiling.CleanupAndExit(valuetypes.CXCompilationError)
	}

	return true
}

// compileSourceStrings compiles sources, named by fileNames, into actions.AST
// with its *init. The cx command and the tests compile through it (#516).
func compileSourceStrings(fileNames []string, sources []string) error {
	if err := parsing.SourceStringsParse(sources, fileNames); err != nil {
		return err
	}

	// Checking if a main package exists. If not, create and add it to `AST`.
	if _, err := astbuild.AST.FunctionGet(astbuild.AST.CxIrState.CxIrStateNameMain, astbuild.AST.CxIrState.CxIrStateNameMain); err != nil {
		initMainPkg(astbuild.AST)
	}

	// Adding *init function that initializes all the global variables.
	if err := parsing.SysInitFunctionAdd(astbuild.AST); err != nil {
		return fmt.Errorf("initializing *init: %w", err)
	}

	astbuild.LineNo = 0

	if environment.FoundCompileErrors {
		printer.CompileErrorsPrint(astbuild.AST)
		return parsing.ErrCompilation
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
	_, fn := ast.FunctionMake(prgrm, prgrm.CxIrState.CxIrStateNameMain, astbuild.CurrentFile, astbuild.LineNo)
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

	fullparsing.Tokenize(r, w, lines, sourceFilename)
}

func printProgramAST(options cxCmdFlags, cxArgs []string, sourceCode []*os.File) {
	profiling.ProfileStart("run")
	defer profiling.ProfileStop("run")

	if options.replMode || len(sourceCode) == 0 {
		repl.Repl()
		return
	}

	// Print CX program.
	astbuild.AST.ProgramPrint()

	if opcodes.AssertFailed() {
		os.Exit(valuetypes.CXAssert)
	}
}

// Used for the -heap-initial, -heap-max and -stack-size flags.
// This function parses, for example, "1M" to 1048576 (the corresponding number of bytes)
// Possible suffixes are: G or g (gigabytes), M or m (megabytes), K or k (kilobytes)
func parseMemoryString(s string) valuetypes.Pointer {
	suffix := s[len(s)-1]
	_, notSuffix := strconv.ParseFloat(string(suffix), 64)

	if notSuffix == nil {
		// then we don't have a suffix
		num, err := strconv.ParseInt(s, 10, 64)

		if err != nil {
			// malformed size
			return valuetypes.InvalidPointer
		}

		return valuetypes.I64ToPtrCast(num)
	} else {
		// then we have a suffix
		num, err := strconv.ParseFloat(s[:len(s)-1], 64)

		if err != nil {
			// malformed size
			return valuetypes.InvalidPointer
		}

		// The user can use suffixes to give as input gigabytes, megabytes or kilobytes.
		switch suffix {
		case 'G', 'g':
			return valuetypes.F64ToPtrCast(num * 1073741824)
		case 'M', 'm':
			return valuetypes.F64ToPtrCast(num * 1048576)
		case 'K', 'k':
			return valuetypes.F64ToPtrCast(num * 1024)
		default:
			return valuetypes.InvalidPointer
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
	environment.BINPATH = filepath.Join(cxPath, "bin/")
	environment.PKGPATH = filepath.Join(cxPath, "pkg/")
	environment.SRCPATH = filepath.Join(cxPath, "src/")
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
