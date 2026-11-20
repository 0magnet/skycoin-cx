package graph

import (
	"bufio"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/skycoin/skycoin-cx/t0-plbolt" // thin: only Get; no depth gain
	// "github.com/skycoin/skycoin-cx/t0-plredis" // redis: unrelated to CX, commented out
	"github.com/skycoin/skycoin-cx/t1-plloader"
)

func GetImportGraph(packageName string, database string) (output string, err error) {
	if packageName == "" {
		return output, errors.New("GetImportGraph: empty program name")
	}
	var packageList loader.PackageList
	listBytes, err := GetStructBytes(packageName, database)
	if err != nil {
		return output, err
	}
	if err := packageList.UnmarshalBinary(listBytes); err != nil {
		return output, fmt.Errorf("decode package list %s: %w", packageName, err)
	}

	var importMap = make(map[string][]string)
	var indexList = make([]string, 0)
	var importIndexMap = make(map[string]int)
	for _, packageString := range packageList.Packages {
		var imports []string
		var packageStruct loader.Package
		packageBytes, err := GetStructBytes(packageString, database)
		if err != nil {
			return output, err
		}
		if err := packageStruct.UnmarshalBinary(packageBytes); err != nil {
			return output, fmt.Errorf("decode package %s: %w", packageString, err)
		}

		for _, fileString := range packageStruct.Files {
			var fileStruct loader.File
			fileBytes, err := GetStructBytes(fileString, database)
			if err != nil {
				return output, err
			}
			if err := fileStruct.UnmarshalBinary(fileBytes); err != nil {
				return output, fmt.Errorf("decode file %s: %w", fileString, err)
			}

			scanner := bufio.NewScanner(strings.NewReader(string(fileStruct.Content)))
			scanner.Split(bufio.ScanWords)

			wordBefore := ""
			for scanner.Scan() {
				if scanner.Text() != "import" {
					wordBefore = scanner.Text()
					continue
				}
				if wordBefore == "//" {
					wordBefore = scanner.Text()
					continue
				}
				if scanner.Text() == "var" || scanner.Text() == "const" || scanner.Text() == "type" || scanner.Text() == "func" {
					break
				}
				if !scanner.Scan() {
					break
				}
				tok := scanner.Text()
				if len(tok) < 2 || tok[0] != '"' || tok[len(tok)-1] != '"' {
					// "import" not followed by a quoted path, so not an import statement.
					wordBefore = tok
					continue
				}
				importString := tok[1 : len(tok)-1]
				if !loader.Contains(imports, importString) {
					imports = append(imports, importString)
				}
				wordBefore = scanner.Text()
			}
			if err := scanner.Err(); err != nil {
				return output, fmt.Errorf("file %s: scan: %w", fileStruct.FileName, err)
			}
		}
		indexList = append(indexList, packageStruct.PackageName)
		importIndexMap[packageStruct.PackageName] = len(indexList) - 1
		importMap[packageStruct.PackageName] = imports
	}
	var otherImports []string
	var otherImportIndex = len(indexList)
	for i, module := range indexList {
		output += "id=" + strconv.Itoa(i) + ", "
		output += "module=" + module + ", "
		var tmp = "imports="
		imports := importMap[module]
		for _, importString := range imports {
			importIndex, ok := importIndexMap[importString]
			if !ok {
				tmp += strconv.Itoa(otherImportIndex) + ","
				otherImports = append(otherImports, "id="+strconv.Itoa(otherImportIndex)+", module="+importString+", imports=,\n")
				otherImportIndex++
				continue
			}
			tmp += strconv.Itoa(importIndex) + ","
		}
		output += tmp + "\n"
	}
	for _, otherImport := range otherImports {
		output += otherImport
	}
	return output, nil
}

func GetStructBytes(structName string, database string) ([]byte, error) {
	if structName == "" {
		return nil, errors.New("get struct bytes: empty key")
	}
	switch database {
	// redis: unrelated to CX, commented out
	// case "redis":
	// 	value, err := redis.Get(structName)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	s, ok := value.(string)
	// 	if !ok {
	// 		return nil, fmt.Errorf("redis: value of %q is %T, want string", structName, value)
	// 	}
	// 	return []byte(s), nil
	case "bolt":
		return bolt.Get(structName)
	}
	return nil, fmt.Errorf("unknown database %q", database)
}
