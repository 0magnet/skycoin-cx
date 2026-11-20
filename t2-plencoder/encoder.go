package encoder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/skycoin/skycoin-cx/t0-plbolt"   // thin: only Get; no depth gain
	"github.com/skycoin/skycoin-cx/t1-plloader" // thin: only File, Package, PackageList; types only, could move them to t0, tier 2 to 1
)

func SavePackagesToDisk(programName string, outDir string) error {
	if programName == "" {
		return errors.New("save packages: empty program name")
	}
	if outDir == "" {
		return fmt.Errorf("save packages %s: empty output path", programName)
	}
	err := os.Mkdir(outDir, 0755)
	if err != nil {
		return err
	}

	var packageList loader.PackageList
	listBytes, err := bolt.Get(programName)
	if err != nil {
		return err
	}
	if err := packageList.UnmarshalBinary(listBytes); err != nil {
		return fmt.Errorf("decode package list %s: %w", programName, err)
	}

	for _, packageKey := range packageList.Packages {
		var packageStruct loader.Package
		packageBytes, err := bolt.Get(packageKey)
		if err != nil {
			return err
		}
		if err := packageStruct.UnmarshalBinary(packageBytes); err != nil {
			return fmt.Errorf("decode package %s: %w", packageKey, err)
		}

		dir := filepath.Join(outDir, "src")
		if packageStruct.PackageName != "main" {
			if err := checkPathElement(packageStruct.PackageName); err != nil {
				return fmt.Errorf("package %s: %w", packageKey, err)
			}
			dir = filepath.Join(dir, packageStruct.PackageName)
		}

		err = SaveFilesToDisk(packageStruct, dir)
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveFilesToDisk(packageStruct loader.Package, dir string) error {
	if dir == "" {
		return fmt.Errorf("save files of package %s: empty output path", packageStruct.PackageName)
	}
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	for _, fileKey := range packageStruct.Files {
		var fileStruct loader.File
		fileBytes, err := bolt.Get(fileKey)
		if err != nil {
			return err
		}
		if err := fileStruct.UnmarshalBinary(fileBytes); err != nil {
			return fmt.Errorf("decode file %s of package %s: %w", fileKey, packageStruct.PackageName, err)
		}

		if err := checkPathElement(fileStruct.FileName); err != nil {
			return fmt.Errorf("file %s of package %s: %w", fileKey, packageStruct.PackageName, err)
		}
		err = os.WriteFile(filepath.Join(dir, fileStruct.FileName), fileStruct.Content, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

// checkPathElement rejects a name read from the database that is not a single
// path element, so stored data cannot write outside the output directory.
func checkPathElement(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) || filepath.Base(name) != name || filepath.IsAbs(name) || filepath.VolumeName(name) != "" {
		return fmt.Errorf("unsafe name %q", name)
	}
	return nil
}
