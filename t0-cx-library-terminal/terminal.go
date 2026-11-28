// Package terminal holds the input side of the terminal library, the
// program arguments and the lines of standard input, which the old loop and
// the vm both read (decision D142). It uses only the standard library.
package terminal

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// VIOLATION: state outside State. Owner: t0-cx-library-terminal; moves to TerminalMng in phase 1.
var terminalArgs [][]byte

// VIOLATION: state outside State. Owner: t0-cx-library-terminal; moves to TerminalMng in phase 1.
var terminalInput *bufio.Reader

// ArgPrefix starts the arguments cx passes to a program. Every form drops
// it, so cx program.cx ++5, ./program ++5 and ./program 5 all read 5.
const ArgPrefix = "++"

// ArgsSet makes args the arguments the program reads, without the name of
// the program or of cx.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func ArgsSet(args []string) {
	terminalArgs = make([][]byte, len(args))
	for index, arg := range args {
		terminalArgs[index] = []byte(strings.TrimPrefix(arg, ArgPrefix))
	}
}

// ArgCount returns the number of program arguments.
func ArgCount() int32 {
	return int32(len(terminalArgs))
}

// ArgRead copies argument index into buffer and returns the bytes copied,
// or -1 when there is no such argument. A longer argument is cut to buffer.
func ArgRead(index int32, buffer []byte) int32 {
	if index < 0 || int(index) >= len(terminalArgs) {
		return -1
	}
	return int32(copy(buffer, terminalArgs[index]))
}

// InputSet makes reader the standard input of the program and returns the
// buffered reader that read() of the old loop shares, so both read one
// stream.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func InputSet(reader io.Reader) *bufio.Reader {
	terminalInput = bufio.NewReader(reader)
	return terminalInput
}

// InputLineRead copies the next line of standard input into buffer without
// its new line or carriage return and returns the bytes copied, or -1 at the
// end of the input. The part of a line longer than buffer is dropped.
// VIOLATION VMALLOC-018: a system library call allocates in Go and the OS, #407.
func InputLineRead(buffer []byte) int32 {
	if terminalInput == nil {
		terminalInput = bufio.NewReader(os.Stdin)
	}
	copied := 0
	read := false
	for {
		chunk, err := terminalInput.ReadSlice('\n')
		read = read || len(chunk) > 0
		if err == nil {
			chunk = chunk[:len(chunk)-1]
		}
		for _, character := range chunk {
			if character != '\r' && copied < len(buffer) {
				buffer[copied] = character
				copied++
			}
		}
		if err != bufio.ErrBufferFull {
			if err != nil && !read {
				return -1
			}
			return int32(copied)
		}
	}
}
