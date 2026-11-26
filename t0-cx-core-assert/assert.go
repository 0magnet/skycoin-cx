// Package assert holds the class B checks of docs/ErrorHandling.md. A failed
// check prints the violation, its caller's file:line and the goroutine stack
// to stderr, then ends the process with os.Exit(ExitCode).
//
// os.Exit is used instead of panic so that no deferred function and no
// recover anywhere on the stack, vendored code included, can catch or hide a
// violation. Deferred functions do not run, which is intended, since nothing
// recovers from a class B failure.
package assert

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
)

// ExitCode is the exit status of a failed assert. It equals
// CXInternalError in t0-types, which this tier 0 package cannot import.
const ExitCode = 3

// PreconditionTrue crashes when a function's input breaks its contract.
// VIOLATION STRHOT-223: E, error or assert text, built on a failure branch only, #412.
func PreconditionTrue(condition bool, message string) {
	if !condition {
		assertCrash("precondition", message)
	}
}

// InvariantTrue crashes when a state that must always hold does not.
// VIOLATION STRHOT-224: E, error or assert text, built on a failure branch only, #412.
func InvariantTrue(condition bool, message string) {
	if !condition {
		assertCrash("invariant", message)
	}
}

// PostconditionTrue crashes when a function's result breaks its contract.
func PostconditionTrue(condition bool, message string) {
	if !condition {
		assertCrash("postcondition", message)
	}
}

// Unreachable crashes unconditionally. Use it for impossible states,
// such as the default case of a switch over every known value.
func Unreachable(message string) {
	assertCrash("unreachable", message)
}

// assertCrash must be called directly by an exported Assert function, so that
// runtime.Caller(2) is the caller of that function.
// VIOLATION STRHOT-225: E, error or assert text, built on a failure branch only, #412.
func assertCrash(kind string, message string) {
	location := "unknown location"
	if _, file, line, ok := runtime.Caller(2); ok {
		location = fmt.Sprintf("%s:%d", file, line)
	}
	fmt.Fprintf(os.Stderr, "assert failed: %s violated at %s: %s\n\n%s", kind, location, message, debug.Stack())
	os.Exit(ExitCode)
}
