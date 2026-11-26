package main

import (
	"strings"
	"testing"
)

func TestNameWords(test *testing.T) {
	cases := map[string]string{
		"CXFunctionMng":      "cx function mng",
		"outputSliceData":    "output slice data",
		"Set_i8":             "set i8",
		"HTTPServerID":       "http server id",
		"isThereDeclPointer": "is there decl pointer",
		"CONST_GL_INFO_LOG":  "const gl info log",
		"CxHashOfBytes":      "cx hash of bytes",
		"netUDPOpen":         "net udp open",
		"ImageDataStart":     "image data start",
		"Is":                 "is",
	}
	for name, want := range cases {
		if got := strings.Join(NameWords(name), " "); got != want {
			test.Errorf("NameWords(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestNameFinding(test *testing.T) {
	cases := map[string]string{
		"outputSliceData": "vague word data",
		"IsTypeStruct":    "is/has prefix",
		"hasSlash":        "is/has prefix",
		"Is":              "",
		"island":          "",
		"Hash":            "",
		"metadata":        "",
		"SymbolTable":     "",
	}
	for name, want := range cases {
		if got := NameFinding(NameRow{Name: name}); got != want {
			test.Errorf("NameFinding(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestNameCheckCompare(test *testing.T) {
	rows := []NameRow{
		{Language: "go", File: "a/a.go", Kind: "local_variable", Parent: "Run", Name: "data"},
		{Language: "go", File: "a/a.go", Kind: "function", Name: "IsOpen"},
		{Language: "go", File: "a/a_test.go", Kind: "local_variable", Parent: "TestRun", Name: "tmp", TestFile: true},
		{Language: "cx", File: "a/a.cx", Kind: "local_variable", Parent: "main", Name: "data"},
	}
	allowed := map[string]bool{"a/a.go|function||IsOpen": true, "a/a.go|function||isGone": true}
	added, fixed := NameCheckCompare(NameCheckFindings(rows), allowed)
	if strings.Join(added, ";") != "a/a.go|local_variable|Run|data  (vague word data)" {
		test.Errorf("added = %q", added)
	}
	if strings.Join(fixed, ";") != "a/a.go|function||isGone" {
		test.Errorf("fixed = %q", fixed)
	}
}

func TestNameFindingPackageLevel(test *testing.T) {
	cases := []struct {
		row  NameRow
		want string
	}{
		{NameRow{Package: "gc", Kind: "struct", Name: "GCState", Exported: true}, "repeats the package name"},
		{NameRow{Package: "ast", Kind: "method", Parent: "CXProgram", Name: "GetCXAtomicOp", Exported: true}, "verb first, the subject leads"},
		{NameRow{Package: "ast", Kind: "method", Parent: "CXProgram", Name: "CXAtomicOpGet", Exported: true}, ""},
		{NameRow{Package: "ast", Kind: "method", Parent: "Lexer", Name: "Error", Exported: true}, ""},
		{NameRow{Package: "types", Kind: "constant", Name: "OBJECT_HEADER_SIZE", Exported: true}, "underscore, not Java type name style"},
		{NameRow{Package: "ast", Kind: "field", Parent: "Lexer", Name: "b"}, "single letter"},
		{NameRow{Package: "ast", Kind: "local_variable", Parent: "Run", Name: "b"}, ""},
		{NameRow{Package: "ast", Kind: "function", Name: "exprLine"}, "abbreviation expr"},
		{NameRow{Package: "ast", Kind: "parameter", Parent: "Run", Name: "prgrm"}, ""},
	}
	for _, c := range cases {
		if got := NameFinding(c.row); got != c.want {
			test.Errorf("NameFinding(%+v) = %q, want %q", c.row, got, c.want)
		}
	}
}

func TestNameCheckDirectoryFinding(test *testing.T) {
	cases := []struct{ directory, packageName, want string }{
		{"t3-cx-runtime-vm", "vm", ""},
		{"t11-cx-command-cx", "main", ""},
		{"t3-ast", "ast", "directory not t<N>-cx-<layer>-<subject>"},
		{"t2-cx-core-stringpool", "pool", "package clause is not the directory subject stringpool"},
		{"t4-cx-misc-thing", "thing", "directory not t<N>-cx-<layer>-<subject>"},
	}
	for _, c := range cases {
		if got := NameCheckDirectoryFinding(c.directory, c.packageName); got != c.want {
			test.Errorf("NameCheckDirectoryFinding(%q, %q) = %q, want %q", c.directory, c.packageName, got, c.want)
		}
	}
}
