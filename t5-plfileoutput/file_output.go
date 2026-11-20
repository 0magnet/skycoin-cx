package file_output

import (
	"errors"
	"fmt"

	"github.com/skycoin/skycoin-cx/t0-plbolt" // thin: only Get; no depth gain
	"github.com/skycoin/skycoin-cx/t1-plloader"
)

// - Adds Imports to AST
// - Returns Import Files
// - Packages must be added to AST first or an error will occur
// - Call AddPkgsToAST before calling this function
func GetImportFiles(packageName string) (files []*loader.File, err error) {
	if packageName == "" {
		return files, errors.New("GetImportFiles: empty program name")
	}

	// Get package list
	var packageList loader.PackageList
	listBytes, err := bolt.Get(packageName)
	if err != nil {
		return files, err
	}
	if err := packageList.UnmarshalBinary(listBytes); err != nil {
		return files, fmt.Errorf("decode package list %s: %w", packageName, err)
	}

	for _, packageString := range packageList.Packages {

		//  Get package struct
		var packageStruct loader.Package
		packageBytes, err := bolt.Get(packageString)
		if err != nil {
			return files, err
		}
		if err := packageStruct.UnmarshalBinary(packageBytes); err != nil {
			return files, fmt.Errorf("decode package %s: %w", packageString, err)
		}

		for _, fileString := range packageStruct.Files {

			// Get file struct
			var fileStruct loader.File
			fileBytes, err := bolt.Get(fileString)
			if err != nil {
				return files, err
			}
			if err := fileStruct.UnmarshalBinary(fileBytes); err != nil {
				return files, fmt.Errorf("decode file %s: %w", fileString, err)
			}

			// Add file struct to array
			files = append(files, &fileStruct)
		}
	}

	return files, nil
}
