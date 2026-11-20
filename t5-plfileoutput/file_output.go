package file_output

import (
	"errors"
	"fmt"

	"github.com/skycoin/skycoin-cx/t0-plbolt"  // thin: only Get; no depth gain
	"github.com/skycoin/skycoin-cx/t0-plredis" // thin: only Get; no depth gain
	"github.com/skycoin/skycoin-cx/t1-plloader"
)

// - Adds Imports to AST
// - Returns Import Files
// - Packages must be added to AST first or an error will occur
// - Call AddPkgsToAST before calling this function
func GetImportFiles(packageName string, database string) (files []*loader.File, err error) {
	if packageName == "" {
		return files, errors.New("GetImportFiles: empty program name")
	}

	// Get package list
	var packageList loader.PackageList
	listBytes, err := GetStructBytes(packageName, database)
	if err != nil {
		return files, err
	}
	if err := packageList.UnmarshalBinary(listBytes); err != nil {
		return files, fmt.Errorf("decode package list %s: %w", packageName, err)
	}

	for _, packageString := range packageList.Packages {

		//  Get package struct
		var packageStruct loader.Package
		packageBytes, err := GetStructBytes(packageString, database)
		if err != nil {
			return files, err
		}
		if err := packageStruct.UnmarshalBinary(packageBytes); err != nil {
			return files, fmt.Errorf("decode package %s: %w", packageString, err)
		}

		for _, fileString := range packageStruct.Files {

			// Get file struct
			var fileStruct loader.File
			fileBytes, err := GetStructBytes(fileString, database)
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

func GetStructBytes(structName string, database string) ([]byte, error) {
	if structName == "" {
		return nil, errors.New("get struct bytes: empty key")
	}
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
