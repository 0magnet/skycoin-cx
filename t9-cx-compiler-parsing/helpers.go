package parsing

import (
	"bytes"

	"github.com/skycoin/skycoin-cx/t5-cx-compiler-partialparsing" // thin: only CurrentFile, Parse, Program; no depth gain
	"github.com/skycoin/skycoin-cx/t8-cx-compiler-fullparsing"    // thin: only NewLexer, Parse; the pass two parser, keep
)

/*
passone perform cxpartial parsing using partialparsing.y and partialparsing.go
*/
func Passone(source string) int {

	parseErrors := partialparsing.Parse(source)

	return parseErrors
}

/*
passtwo perform cxparsingcompletor parsing using parsingcompletor.y and parsingcompletor.go
*/

func Passtwo(b *bytes.Buffer) int {

	parseErrors := fullparsing.Parse(fullparsing.LexerNew(b))

	return parseErrors
}
