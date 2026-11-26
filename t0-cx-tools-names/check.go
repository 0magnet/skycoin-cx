package main

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// NameCheckVagueWords are the words of the AGENTS.md Naming section that do
// not say what a name is. A Go name with one of them as a word is a finding.
var NameCheckVagueWords = map[string]bool{
	"data": true, "info": true, "tmp": true, "temp": true, "item": true,
	"thing": true, "stuff": true, "misc": true, "do": true, "handle": true,
	"handler": true, "process": true, "manager": true, "helper": true,
	"util": true,
}

// NameWords splits a name into its lowercase words at underscores, at a
// lower to upper case step and at the end of an upper case run.
func NameWords(name string) []string {
	var words []string
	runes := []rune(name)
	start := 0
	flush := func(end int) {
		if end > start {
			words = append(words, strings.ToLower(string(runes[start:end])))
		}
	}
	for i, r := range runes {
		if r == '_' {
			flush(i)
			start = i + 1
			continue
		}
		if i == start || !unicode.IsUpper(r) {
			continue
		}
		previous := runes[i-1]
		nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
		if unicode.IsLower(previous) || unicode.IsDigit(previous) || (unicode.IsUpper(previous) && nextLower) {
			flush(i)
			start = i
		}
	}
	flush(len(runes))
	return words
}

// NameFinding returns why a Go name breaks the Naming rules, or "" when it
// does not.
func NameFinding(row NameRow) string {
	words := NameWords(row.Name)
	if len(words) > 1 && (words[0] == "is" || words[0] == "has") {
		return "is/has prefix"
	}
	for _, word := range words {
		if NameCheckVagueWords[word] {
			return "vague word " + word
		}
	}
	if !NameCheckPackageLevel(row) {
		return ""
	}
	if len([]rune(row.Name)) == 1 {
		return "single letter"
	}
	if strings.Contains(strings.Trim(row.Name, "_"), "_") {
		return "underscore, not Java type name style"
	}
	if row.Exported && row.Kind != "field" && row.Kind != "method" && row.Kind != "interface_method" &&
		len(words) > 1 && words[0] == strings.ToLower(row.Package) {
		return "repeats the package name"
	}
	if row.Exported && (row.Kind == "function" || row.Kind == "method") && len(words) > 1 &&
		NameCheckVerbs[words[0]] && !NameCheckInterfaceMethods[row.Name] {
		return "verb first, the subject leads"
	}
	for _, word := range words {
		if NameCheckAbbreviations[word] {
			return "abbreviation " + word
		}
	}
	return ""
}

// NameCheckKey names a finding without its line, so an allowlist entry
// survives edits elsewhere in the file. A package is keyed by its directory.
func NameCheckKey(row NameRow) string {
	if row.Kind == "package" {
		return path.Dir(row.File) + "|package||" + row.Name
	}
	return row.File + "|" + row.Kind + "|" + row.Parent + "|" + row.Name
}

// NameCheckFindings returns the sorted keys of the Go findings outside test
// files, each with its reason. Imports are skipped, their names are not ours.
func NameCheckFindings(rows []NameRow) map[string]string {
	findings := map[string]string{}
	for _, row := range rows {
		if row.Language != "go" || row.Kind == "import" {
			continue
		}
		if directory := path.Dir(row.File); row.Kind == "package" && !strings.Contains(directory, "/") && strings.HasPrefix(directory, "t") {
			if finding := NameCheckDirectoryFinding(directory, row.Name); finding != "" {
				findings[directory+"|directory||"+directory] = finding
			}
		}
		if row.TestFile {
			continue
		}
		if finding := NameFinding(row); finding != "" {
			findings[NameCheckKey(row)] = finding
		}
	}
	return findings
}

// NameCheckAllowlistRead reads the allowlist keys, one per line. Blank lines
// and lines starting with # are skipped, so a comment can give the reason.
func NameCheckAllowlistRead(file string) (map[string]bool, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	allowed := map[string]bool{}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		allowed[line] = true
	}
	return allowed, nil
}

// NameCheckCompare returns the findings missing from the allowlist and the
// allowlist entries that are no longer findings, both sorted.
func NameCheckCompare(findings map[string]string, allowed map[string]bool) (added, fixed []string) {
	for key, finding := range findings {
		if !allowed[key] {
			added = append(added, key+"  ("+finding+")")
		}
	}
	for key := range allowed {
		if _, found := findings[key]; !found {
			fixed = append(fixed, key)
		}
	}
	sort.Strings(added)
	sort.Strings(fixed)
	return added, fixed
}

// NameCheckRun prints the new and the fixed findings and returns the exit
// code, 1 when either list is not empty.
func NameCheckRun(rows []NameRow, allowlistPath string) int {
	allowed, err := NameCheckAllowlistRead(allowlistPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	added, fixed := NameCheckCompare(NameCheckFindings(rows), allowed)
	for _, key := range added {
		fmt.Println("name breaks the Naming section of AGENTS.md, rename it:", key)
	}
	for _, key := range fixed {
		fmt.Printf("fixed, remove from %s: %s\n", allowlistPath, key)
	}
	fmt.Println(len(allowed), "allowlisted names")
	if len(added) > 0 || len(fixed) > 0 {
		return 1
	}
	return 0
}

// NameCheckVerbs are the verbs that may not lead an exported function or
// method name. The subject leads, so related names sort together:
// RouteGroupClose, not CloseRouteGroup. Run and compile are not listed, they
// are subjects as well, as in RunBuffersPrepare and CompileErrorList.
var NameCheckVerbs = map[string]bool{
	"add": true, "alloc": true, "allocate": true, "append": true, "apply": true,
	"assign": true, "build": true, "calculate": true, "cast": true,
	"check": true, "clear": true, "close": true, "collect": true,
	"compute": true, "convert": true, "copy": true, "create": true, "declare": true,
	"decode": true, "delete": true, "deserialize": true, "emit": true, "encode": true,
	"execute": true, "extract": true, "find": true, "format": true, "free": true,
	"generate": true, "get": true, "init": true, "insert": true, "load": true,
	"lookup": true, "make": true, "mark": true, "new": true, "open": true,
	"parse": true, "pop": true, "print": true, "push": true, "read": true,
	"register": true, "release": true, "remove": true, "reset": true,
	"resize": true, "resolve": true, "save": true, "scan": true,
	"serialize": true, "set": true, "start": true, "stop": true, "update": true,
	"validate": true, "verify": true, "write": true,
}

// NameCheckInterfaceMethods are method names that a standard interface or a
// generated parser calls, which keep their Go name.
var NameCheckInterfaceMethods = map[string]bool{
	"String": true, "Error": true, "Unwrap": true, "Is": true, "As": true,
	"Read": true, "Write": true, "Close": true, "Seek": true, "Format": true,
	"MarshalBinary": true, "UnmarshalBinary": true, "MarshalJSON": true,
	"UnmarshalJSON": true, "MarshalText": true, "UnmarshalText": true,
	"Len": true, "Less": true, "Swap": true, "Lex": true, "ReadByte": true,
	"WriteByte": true, "WriteString": true, "ReadRune": true, "Reset": true,
}

// NameCheckAbbreviations are shortened words that are not the usual name of
// what they stand for. CX type and opcode words such as str, ptr and jmp are
// CX names, and the compiler words arg, instr and spec are usual, so none of
// them is listed.
var NameCheckAbbreviations = map[string]bool{
	"prgrm": true, "expr": true, "exprs": true, "fn": true, "fns": true,
	"glbl": true, "glbls": true, "inp": true, "inps": true, "idx": true,
	"sym": true, "syms": true, "decl": true, "decls": true, "obj": true,
	"objs": true, "val": true, "vals": true, "num": true, "cnt": true,
	"sig": true, "typ": true, "cur": true, "curr": true, "prev": true,
	"res": true, "ret": true, "elt": true, "elem": true, "cxe": true,
	"cxf": true, "strct": true,
}

// NameCheckPackageLevel reports whether a row is declared at package level,
// where a name is read far from its declaration.
func NameCheckPackageLevel(row NameRow) bool {
	switch row.Kind {
	case "function", "method", "struct", "type", "enum", "enum_value", "constant",
		"global_variable", "field", "interface", "interface_method", "function_type":
		return true
	}
	return false
}

// NameCheckDirectoryPattern is the form of a package directory,
// t<N>-cx-<layer>-<subject>, see docs/NamingPlan.md.
var NameCheckDirectoryPattern = regexp.MustCompile(`^t[0-9]+-cx-(core|compiler|runtime|library|tools|command)-([a-z0-9]+)$`)

// NameCheckDirectoryFinding returns why a package directory breaks the
// form, or why its package clause is not its subject, or "".
func NameCheckDirectoryFinding(directory string, packageName string) string {
	match := NameCheckDirectoryPattern.FindStringSubmatch(directory)
	if match == nil {
		return "directory not t<N>-cx-<layer>-<subject>"
	}
	if packageName != "main" && packageName != match[2] {
		return "package clause is not the directory subject " + match[2]
	}
	return ""
}
