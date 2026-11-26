package repl

//put the repo command stuff here

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/skycoin/skycoin-cx/t2-cx-core-environment"    // thin: only FoundCompileErrors; no depth gain
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"            // thin: only InREPL, ProgramMemoryRelease, SourceFileGidNil; no depth gain
	"github.com/skycoin/skycoin-cx/t4-cx-runtime-interpreter" // thin: only RunCompiled; no depth gain
	"github.com/skycoin/skycoin-cx/t4-cx-runtime-printer"     // thin: only CompileErrorsPrint, RuntimeErrorPrint; no depth gain
	"github.com/skycoin/skycoin-cx/t7-cx-runtime-libraries"   // thin: only MakeProgram, RegisterPackages; no depth gain

	"github.com/skycoin/skycoin-cx/t4-cx-compiler-astbuild"       // thin: only AST, LineNo; no depth gain
	"github.com/skycoin/skycoin-cx/t5-cx-compiler-partialparsing" // thin: only CurrentFile, Parse, Program; no depth gain
	"github.com/skycoin/skycoin-cx/t8-cx-compiler-fullparsing"    // thin: only Lexer, NewLexer, Parse; no depth gain
	"github.com/skycoin/skycoin-cx/t9-cx-compiler-parsing"        // thin: only AddInitFunction, ErrCompilation; could move AddInitFunction to t4-cx-compiler-astbuild
)

const VERSION = "0.8.0"

// VIOLATION: state outside State. Owner: t10-cx-tools-repl; moves to ReplMng in phase 2.
var TargetFunction string = ""

// VIOLATION: state outside State. Owner: t10-cx-tools-repl; moves to ReplMng in phase 2.
var TargetStruct string = ""

// VIOLATION: state outside State. Owner: t10-cx-tools-repl; moves to ReplMng in phase 2.
var TargetModule string = ""

func unsafeEval(code string) (out string) {
	var lexer *fullparsing.Lexer

	// storing strings sent to standard output
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return fmt.Sprintf("%s", err)
	}
	// Drain the pipe while the program runs, or output larger than the
	// pipe buffer blocks it.
	outC := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		r.Close()
		outC <- buf.String()
	}()
	os.Stdout = w

	restored := false
	restore := func() string {
		if restored {
			return ""
		}
		restored = true
		w.Close()
		os.Stdout = old // restoring the real stdout
		return <-outC
	}

	defer func() {
		if r := recover(); r != nil {
			printer.CompileErrorsPrint(astbuild.AST)
			restore()
			out = fmt.Sprintf("%v", r)
			// lexer.Stop()
		}
	}()

	astbuild.LineNo = 0
	// Eval code has no file, and a file Gid of an earlier program is stale.
	astbuild.CurrentFile = ast.SourceFileGidNil
	partialparsing.CurrentFile = ast.SourceFileGidNil
	environment.FoundCompileErrors = false

	// VIOLATION: init chaining. Init 7-8 again, a fresh program on REPL reset.
	// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
	// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S0001 S0002.
	astbuild.AST = libraries.ProgramMake()
	libraries.PackagesRegister(astbuild.AST)
	partialparsing.Program = astbuild.AST

	parseErrors := partialparsing.Parse(code)

	astbuild.AST = partialparsing.Program

	lexer = fullparsing.LexerNew(bytes.NewBufferString(code))
	parseErrors += fullparsing.Parse(lexer)
	printer.CompileErrorsPrint(astbuild.AST)
	if parseErrors > 0 || environment.FoundCompileErrors {
		restore()
		// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
		// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S0003 S0004.
		astbuild.AST = libraries.ProgramMake()
		libraries.PackagesRegister(astbuild.AST)
		return parsing.ErrCompilation.Error()
	}
	err = parsing.SysInitFunctionAdd(astbuild.AST)
	if err != nil {
		restore()
		return fmt.Sprintf("%s", err)
	}
	// SysInitFunctionAdd records initialization cycles and global type errors (#519).
	if environment.FoundCompileErrors {
		printer.CompileErrorsPrint(astbuild.AST)
		restore()
		astbuild.AST = libraries.ProgramMake()
		libraries.PackagesRegister(astbuild.AST)
		return parsing.ErrCompilation.Error()
	}

	astbuild.AST.CxIrState.CxIrStateManagerGrowBoundary()
	astbuild.AST.CxIrState.StringPool.StringPoolFreeze()
	//err = actions.AST.RunCompiled(0, nil);
	err = interpreter.RunCompiled(astbuild.AST, 0, nil)
	if astbuild.AST.RuntimeError.RuntimeErrorRecordRaised {
		os.Exit(printer.RuntimeErrorPrint(astbuild.AST))
	}
	if err != nil {
		restore()
		// VIOLATION: init chaining. Init 7-8 again, a fresh program on REPL reset.
		// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
		// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S0005 S0006.
		ast.ProgramMemoryRelease(astbuild.AST)
		astbuild.AST = libraries.ProgramMake()
		libraries.PackagesRegister(astbuild.AST)
		return fmt.Sprintf("%s", err)
	}
	//Tod: If error equals nill?

	out = restore()

	// VIOLATION: init chaining. Init 7-8 again, a fresh program on REPL reset.
	// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
	// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S0007 S0008.
	ast.ProgramMemoryRelease(astbuild.AST)
	astbuild.AST = libraries.ProgramMake()
	libraries.PackagesRegister(astbuild.AST)
	return out
}

// Eval runs code as a CX program and returns its output. The caller must
// first run the init in t11-cx Run: ConstantsInit, ProgramMake, PackagesRegister.
func Eval(code string) string {
	runtime.GOMAXPROCS(2)
	ch := make(chan string, 1)

	var result string

	go func() {
		result = unsafeEval(code)
		ch <- result
	}()

	timer := time.NewTimer(20 * time.Second)
	defer timer.Stop()

	select {
	case <-ch:
		return result
	case <-timer.C:
		// VIOLATION: init chaining. Init 7-8 again, a fresh program on REPL reset.
		// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
		// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S0009 S0010.
		astbuild.AST = libraries.ProgramMake()
		libraries.PackagesRegister(astbuild.AST)
		return "Timed out."
	}
}

func Repl() {
	fmt.Println("CX", VERSION)
	fmt.Println("More information about CX is available at https://github.com/skycoin/skycoin-cx")

	ast.InREPL = true

	// fi := bufio.NewReader(os.NewFile(0, "stdin"))
	fi := bufio.NewReader(os.Stdin)
	astbuild.AST.ProgramStdinInit(fi)
	// scanner := bufio.NewScanner(os.Stdin)

	for {
		astbuild.AST.CxIrState.CxIrStateManagerGrowBoundary()
		var inp string
		var ok bool

		printPrompt()

		if inp, ok = readline(fi); ok {
			if TargetFunction != "" {
				inp = fmt.Sprintf(":func %s {\n%s\n}\n", TargetFunction, inp)
			}
			if TargetModule != "" {
				inp = fmt.Sprintf("%s", inp)
			}
			if TargetStruct != "" {
				inp = fmt.Sprintf(":struct %s {\n%s\n}\n", TargetStruct, inp)
			}

			b := bytes.NewBufferString(inp)

			fullparsing.Parse(fullparsing.LexerNew(b))
			printer.CompileErrorsPrint(astbuild.AST)
			//yyParse(LexerNew(b))
		} else {
			if TargetFunction != "" {
				TargetFunction = ""
				continue
			}

			if TargetStruct != "" {
				TargetStruct = ""
				continue
			}

			if TargetModule != "" {
				TargetModule = ""
				continue
			}

			fmt.Printf("\nBye!\n")
			break
		}
	}
}

func printPrompt() {
	if TargetModule != "" {
		fmt.Println(fmt.Sprintf(":package %s ...", TargetModule))
		fmt.Printf("* ")
	} else if TargetFunction != "" {
		fmt.Println(fmt.Sprintf(":func %s {...", TargetFunction))
		fmt.Printf("\t* ")
	} else if TargetStruct != "" {
		fmt.Println(fmt.Sprintf(":struct %s {...", TargetStruct))
		fmt.Printf("\t* ")
	} else {
		fmt.Printf("* ")
	}
}

// ----------------------------------------------------------------
//                     Utility functions

func readline(fi *bufio.Reader) (string, bool) {
	s, err := fi.ReadString('\n')

	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\r", "", -1)

	for _, ch := range s {
		if ch == rune(4) {
			err = io.EOF
			break
		}
	}

	if err != nil {
		return "", false
	}

	return s, true
}
