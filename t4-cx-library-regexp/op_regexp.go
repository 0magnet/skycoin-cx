//go:build !noregexp

package regexp

import (
	"fmt"
	"os"
	"regexp"
	"unsafe"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes" // thin: only StrBytesRead, STR, CXRuntimeError; no depth gain
	objmodel "github.com/skycoin/skycoin-cx/t1-cx-core-objmodel"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// LibraryState is the state of the regexp library, its one global,
// see D139 on #412. LibraryPool interns every expression a program
// compiles, and LibraryCompiledMng holds the compiled form of each.
type LibraryState struct {
	LibraryPool        stringpool.StringPool
	LibraryCompiledMng CompiledMng
}

// Library is the state of the regexp library.
var Library LibraryState

// CompiledMng holds the compiled expressions in a backing array, by the slot
// of the Gid of their expression plus one, and the empty expression at 0. A
// slot is nil for an expression that did not compile.
type CompiledMng struct {
	CompiledMngSlots []*regexp.Regexp
}

func compiledMngIndex(expression stringpool.StringGid) int {
	if expression == stringpool.StringGidNil {
		return 0
	}
	return int(objmodel.GidIndex(int64(expression))) + 1
}

// CompiledGet returns the compiled form of expression, nil when it has none.
func (regexps *CompiledMng) CompiledGet(expression stringpool.StringGid) *regexp.Regexp {
	if index := compiledMngIndex(expression); index < len(regexps.CompiledMngSlots) {
		return regexps.CompiledMngSlots[index]
	}
	return nil
}

// CompiledSet records compiled as the compiled form of expression.
func (regexps *CompiledMng) CompiledSet(expression stringpool.StringGid, compiled *regexp.Regexp) {
	index := compiledMngIndex(expression)
	for len(regexps.CompiledMngSlots) <= index {
		regexps.CompiledMngSlots = append(regexps.CompiledMngSlots, nil)
	}
	regexps.CompiledMngSlots[index] = compiled
}

// expOffset returns the offset of the `exp` field of the CX `regexp.Regexp`
// held by value. The field holds the expression, which finds its compiled
// form. exp is the only field PackageRegister gives Regexp, so it starts the
// struct.
func expOffset(value ast.CXValue) valuetypes.Pointer {
	return value.Offset
}

// regexpCompile is a helper function for `OpRegexpMustCompile` and
// `OpRegexpCompile`. It writes the CX structure `regexp.Regexp` and compiles
// its expression once, interned in LibraryPool.
// GCREACH: allocates the expression str, which adds to the GC pressure,
// see docs/GarbageCollector.md.
// VIOLATION STRHOT-142: B, regexp.Compile takes the expression as a string, a view of pool bytes, #412.
// VIOLATION VMALLOC-017: compiles and matches with the Go regexp package, #407.
func regexpCompile(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) error {
	exp := inputs[0].StrBytesGet(prgrm)

	// The CX Regexp only holds the expression. The same expression in two
	// parts of a program shares one compiled instance.
	ast.StrBytesAllocateAt(prgrm, expOffset(outputs[0]), exp, nil)

	expression := Library.LibraryPool.StringPoolIntern(exp)
	if Library.LibraryCompiledMng.CompiledGet(expression) != nil {
		return nil
	}
	text := Library.LibraryPool.StringPoolBytes(expression)
	r, err := regexp.Compile(unsafe.String(unsafe.SliceData(text), len(text)))
	if err != nil {
		return err
	}
	Library.LibraryCompiledMng.CompiledSet(expression, r)
	return nil
}

// OpRegexpMustCompile is a wrapper for golang's `regexp`'s `MustCompile`.
func OpRegexpMustCompile(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	err := regexpCompile(prgrm, inputs, outputs)

	if err != nil {
		regexpMustCompileFailureReport(err)
	}
}

// regexpMustCompileFailureReport writes the error of a failed MustCompile to
// standard error and stops the program with CXRuntimeError.
// STRHOT REPORTER: builds the text of the Go regexp error, #412.
func regexpMustCompileFailureReport(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	panic(valuetypes.CXRuntimeError)
}

// OpRegexpCompile is a wrapper for golang's `regexp`'s `Compile`. Its second
// output is the text of the error, empty on success.
// VIOLATION STRHOT-143: B, Go gives the regexp error only as text, copied into the str, #412.
func OpRegexpCompile(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	err := regexpCompile(prgrm, inputs, outputs)

	var errText []byte
	if err != nil {
		message := err.Error()
		errText = unsafe.Slice(unsafe.StringData(message), len(message))
	}
	outputs[1].StrBytesSet(prgrm, errText)
}

// OpRegexpFind is a wrapper for golang's `regexp`'s `Find`.
// VIOLATION VMALLOC-017: compiles and matches with the Go regexp package, #407.
func OpRegexpFind(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	exp := valuetypes.StrBytesRead(prgrm.Memory, expOffset(inputs[0]))
	expression, found := Library.LibraryPool.StringPoolFind(exp)
	r := Library.LibraryCompiledMng.CompiledGet(expression)
	if !found || r == nil {
		regexpFindFailureReport(exp)
	}
	outputs[0].StrBytesSet(prgrm, r.Find(inputs[1].StrBytesGet(prgrm)))
}

// regexpFindFailureReport writes that Find ran on a Regexp whose expression
// exp did not compile and stops the program with CXRuntimeError.
// STRHOT REPORTER: builds the text from the expression bytes, #412.
func regexpFindFailureReport(exp []byte) {
	fmt.Fprintln(os.Stderr, "regexp: Find on a Regexp that did not compile: "+string(exp))
	panic(valuetypes.CXRuntimeError)
}
