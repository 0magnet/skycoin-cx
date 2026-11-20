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
func GetImportFiles(programName string) (files []*loader.File, err error) {
	if programName == "" {
		return files, errors.New("GetImportFiles: empty program name")
	}

	// Get package list
	var packageList loader.PackageList
	listBytes, err := bolt.Get(programName)
	if err != nil {
		return files, err
	}
	if err := packageList.UnmarshalBinary(listBytes); err != nil {
		return files, fmt.Errorf("decode package list %s: %w", programName, err)
	}

	for _, packageKey := range packageList.Packages {

		//  Get package struct
		var packageStruct loader.Package
		packageBytes, err := bolt.Get(packageKey)
		if err != nil {
			return files, err
		}
		if err := packageStruct.UnmarshalBinary(packageBytes); err != nil {
			return files, fmt.Errorf("decode package %s: %w", packageKey, err)
		}

		for _, fileKey := range packageStruct.Files {

			// Get file struct
			var fileStruct loader.File
			fileBytes, err := bolt.Get(fileKey)
			if err != nil {
				return files, err
			}
			if err := fileStruct.UnmarshalBinary(fileBytes); err != nil {
				return files, fmt.Errorf("decode file %s: %w", fileKey, err)
			}

			// Add file struct to array
			files = append(files, &fileStruct)
		}
	}

	return files, nil
}
