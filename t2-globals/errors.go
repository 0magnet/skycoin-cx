package globals

// DebugGolangStackTrace prints the Go stack of a CX runtime error. cx
// --debug-go-stack sets it.
// VIOLATION: state outside State. Owner: t2-globals; moves to ParseMng in phase 2.
var DebugGolangStackTrace bool

// VIOLATION: state outside State. Owner: t2-globals; moves to ParseMng in phase 2.
var FoundCompileErrors bool
