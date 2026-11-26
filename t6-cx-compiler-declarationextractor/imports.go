package declarationextractor

import (
	"bufio"
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
)

type ImportDeclaration struct {
	PackageID  string // name of package
	FileID     string // name of file
	LineNumber int    // line number of declaration
	ImportName string // name of import being declared
}

func ImportsExtract(source []byte, fileName string) ([]ImportDeclaration, error) {

	var ImportDeclarationsArray []ImportDeclaration
	var pkg string

	//Regexs
	rePkg := regexp.MustCompile(`^(?:.+\s+|\s*)package(?:\s+[\S\s]+|\s*)$`)
	rePkgName := regexp.MustCompile(`^\s*package\s+([_a-zA-Z][_a-zA-Z0-9]*)`)

	reader := bytes.NewReader(source)
	scanner := bufio.NewScanner(reader)

	var lineno int // line number

	for scanner.Scan() {

		line := scanner.Bytes()
		lineno++

		// Package declaration extraction
		if rePkg.FindIndex(line) != nil {

			matchPkg := rePkgName.FindSubmatch(line)

			if matchPkg == nil || !bytes.Equal(matchPkg[0], bytes.TrimSpace(line)) {
				return ImportDeclarationsArray, fmt.Errorf("%v:%v: syntax error: package declaration", filepath.Base(fileName), lineno)
			}

			pkg = string(matchPkg[1])

		}

		// Extract Import
		checkLine := bytes.Split(line, []byte(" "))
		if bytes.Equal(checkLine[0], []byte("import")) {

			var importDeclaration ImportDeclaration

			importDeclaration.PackageID = pkg
			importDeclaration.FileID = fileName
			importDeclaration.LineNumber = lineno

			importDeclaration.ImportName = string(checkLine[1][1 : len(checkLine[1])-1])

			ImportDeclarationsArray = append(ImportDeclarationsArray, importDeclaration)
		}
	}

	return ImportDeclarationsArray, nil
}
