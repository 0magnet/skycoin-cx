package tree

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/skycoin/skycoin-cx/t0-plbolt"  // thin: only Get; no depth gain
	"github.com/skycoin/skycoin-cx/t0-plredis" // thin: only Get; no depth gain
	"github.com/skycoin/skycoin-cx/t1-plloader"
)

func GetImportTree(packageName string, database string) (output string, err error) {
	var packageList loader.PackageList
	listBytes, err := GetStructBytes(packageName, database)
	if err != nil {
		return output, err
	}
	if err := packageList.UnmarshalBinary(listBytes); err != nil {
		return output, fmt.Errorf("decode package list %s: %w", packageName, err)
	}

	var importMap = make(map[string][]string)
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
		importMap[packageStruct.PackageName] = imports
	}
	output += "main\n"
	var layers = make(map[int]bool)
	for i, mainImport := range importMap["main"] {
		if i == len(importMap["main"])-1 {
			output += "`--" + mainImport + "\n"
			layers[0] = false
		} else {
			output += "|--" + mainImport + "\n"
			layers[0] = true
		}
		output += AddDependenciesRecur(importMap, importMap[mainImport], 1, layers)
	}
	return output, nil
}

func AddDependenciesRecur(importMap map[string][]string, dependencies []string, depth int, layers map[int]bool) (output string) {
	if len(dependencies) == 0 {
		return output
	}
	for i, dependencyImport := range dependencies {
		for l := 0; l < depth; l++ {
			if layers[l] {
				output += "|  "
			} else {
				output += "   "
			}
		}
		if i == len(dependencies)-1 {
			output += "`--" + dependencyImport + "\n"
			layers[depth] = false
		} else {
			output += "|--" + dependencyImport + "\n"
			layers[depth] = true
		}
		output += AddDependenciesRecur(importMap, importMap[dependencyImport], depth+1, layers)
	}
	return output
}

func GetStructBytes(structName string, database string) ([]byte, error) {
	switch database {
	case "redis":
		value, err := redis.Get(structName)
		if err != nil {
			return nil, err
		}
		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("redis: value of %q is %T, want string", structName, value)
		}
		return []byte(s), nil
	case "bolt":
		return bolt.Get(structName)
	}
	return nil, fmt.Errorf("unknown database %q", database)
}
