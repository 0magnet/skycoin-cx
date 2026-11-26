package type_checker

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"unicode"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
	"github.com/skycoin/skycoin-cx/t1-cx-tools-packageloader" // thin: only File; no depth gain
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"github.com/skycoin/skycoin-cx/t4-actions"
)

func ParameterDeclarationParse(prgrm *ast.CXProgram, parameterString []byte, pkg *ast.CXPackage, fileName ast.SourceFileGid, lineno int) (*ast.CXArgument, error) {
	var parameterDeclaration *ast.CXArgument
	reParameterDeclaration := regexp.MustCompile(`(\w+)\s+(.+)`)
	parameterDeclarationTokens := reParameterDeclaration.FindSubmatch(parameterString)

	// Check if the tokenized result is empty
	if parameterDeclarationTokens == nil || len(parameterDeclarationTokens[0]) == 0 {
		return nil, fmt.Errorf("%s: %d: parameter declaration error", filepath.Base(prgrm.CXProgramSourceFileText(fileName)), lineno)
	}

	// Set the declarator or the name of the param
	_, declarator := ast.ArgumentMake(prgrm, stringpool.StringGidNil, fileName, lineno)
	declarator.TypeSet(valuetypes.UNDEFINED)
	declarator.Package = ast.CXPackageIndex(pkg.Index)
	declarator.Name = prgrm.CxIrState.StringPool.StringPoolIntern(parameterDeclarationTokens[1])

	//Set the decalaration type
	parameterDeclaration, err := DeclarationSpecifierParse(prgrm, parameterDeclarationTokens[2], fileName, lineno, parameterDeclaration)
	if err != nil {
		prgrm.CXArgBuilderRelease(declarator)
		return nil, err
	}

	// Merge both CXArgs
	details := prgrm.CxIrState.CXArgumentDebugMng.CXArgumentDebugGet(parameterDeclaration.ArgDetails)
	details.FileName, details.FileLine = fileName, lineno
	parameterDeclaration.Name = declarator.Name
	parameterDeclaration.Package = declarator.Package
	prgrm.CXArgBuilderRelease(declarator)

	return parameterDeclaration, nil
}

func DeclarationSpecifierParse(prgrm *ast.CXProgram, declarationSpecifierByte []byte, fileName ast.SourceFileGid, lineno int, declarationSpecifier *ast.CXArgument) (*ast.CXArgument, error) {
	// Base case if all parts are parsed
	if declarationSpecifierByte == nil || len(declarationSpecifierByte) == 0 {
		return declarationSpecifier, nil
	}

	// Checks last byte to determine what to parse
	lastByte := declarationSpecifierByte[len(declarationSpecifierByte)-1]

	if unicode.IsLetter(rune(lastByte)) || unicode.IsNumber(rune(lastByte)) || lastByte == '_' {

		reWords := regexp.MustCompile(`[\w\.]+`)
		words := reWords.FindAll(declarationSpecifierByte, -1)
		wordsIdx := reWords.FindAllIndex(declarationSpecifierByte, -1)
		newLastIdx := wordsIdx[len(wordsIdx)-1][0]

		typeWord := words[len(words)-1]
		typeWordParts := bytes.Split(typeWord, []byte("."))

		newDeclarationSpecifierByte := declarationSpecifierByte[:newLastIdx]

		if len(typeWordParts) == 1 {

			// Types like i32, str, aff, etc...
			if val, ok := TypesMap[string(typeWordParts[0])]; ok {
				newDeclarationSpecifierArg := actions.DeclarationSpecifiersBasic(prgrm, val)
				return DeclarationSpecifierParse(prgrm, newDeclarationSpecifierByte, fileName, lineno, newDeclarationSpecifierArg)
			}

			// Structs
			newDeclarationSpecifierArg := actions.DeclarationSpecifiersStruct(actions.AST, string(typeWordParts[0]), "", false, fileName, lineno, 0)
			return DeclarationSpecifierParse(prgrm, newDeclarationSpecifierByte, fileName, lineno, newDeclarationSpecifierArg)
		}

		// External types
		if val, ok := TypesMap[string(typeWordParts[0])]; ok {
			newDeclarationSpecifierArg := actions.DeclarationSpecifiersStruct(actions.AST, string(typeWordParts[1]), val.Name(), true, fileName, lineno, 0)
			return DeclarationSpecifierParse(prgrm, newDeclarationSpecifierByte, fileName, lineno, newDeclarationSpecifierArg)
		}

		// External structs
		newDeclarationSpecifierArg := actions.DeclarationSpecifiersStruct(actions.AST, string(typeWordParts[1]), string(typeWordParts[0]), true, fileName, lineno, 0)
		return DeclarationSpecifierParse(prgrm, newDeclarationSpecifierByte, fileName, lineno, newDeclarationSpecifierArg)
	}

	if lastByte == ']' {
		reBrackets := regexp.MustCompile(`\[\s*(\d*)\s*\]`)
		brackets := reBrackets.FindAllSubmatch(declarationSpecifierByte, -1)
		bracketsIdx := reBrackets.FindAllIndex(declarationSpecifierByte, -1)
		newLastIdx := bracketsIdx[len(bracketsIdx)-1][0]
		reNumber := regexp.MustCompile(`\d+`)
		number := reNumber.Find(brackets[len(brackets)-1][1])
		newDeclarationSpecifierByte := declarationSpecifierByte[:newLastIdx]

		// Arrays
		if number != nil {
			byteToInt, err := strconv.Atoi(string(number))
			if err != nil {
				return declarationSpecifier, err
			}

			// Parsed right to left, so each length goes in front: outermost first.
			newDeclarationSpecifierArg := actions.DeclarationSpecifiers(declarationSpecifier, []valuetypes.Pointer{valuetypes.Pointer(byteToInt)}, opcodetable.DeclarationArray)
			return DeclarationSpecifierParse(prgrm, newDeclarationSpecifierByte, fileName, lineno, newDeclarationSpecifierArg)
		}

		// Slices
		newDeclarationSpecifierArg := actions.DeclarationSpecifiers(declarationSpecifier, []valuetypes.Pointer{0}, opcodetable.DeclarationSlice)
		return DeclarationSpecifierParse(prgrm, newDeclarationSpecifierByte, fileName, lineno, newDeclarationSpecifierArg)
	}

	// Pointer
	if lastByte == '*' {
		newLastIdx := bytes.LastIndex(declarationSpecifierByte, []byte("*"))
		newDeclarationSpecifierByte := declarationSpecifierByte[:newLastIdx]
		newDeclarationSpecifierArg := actions.DeclarationSpecifiers(declarationSpecifier, []valuetypes.Pointer{0}, opcodetable.DeclarationPointer)
		return DeclarationSpecifierParse(prgrm, newDeclarationSpecifierByte, fileName, lineno, newDeclarationSpecifierArg)

	}

	// If bytes don't match any of the cases
	return nil, fmt.Errorf("%v: %d: declaration specifier error", prgrm.CXProgramSourceFileText(fileName), lineno)
}

// Finds the SourceBytes from the files array
func SourceBytesGet(files []*packageloader.File, fileName string) ([]byte, error) {
	for _, file := range files {
		if file.FileName == fileName {
			return file.Content, nil
		}
	}

	return nil, fmt.Errorf("%s not found", fileName)
}
