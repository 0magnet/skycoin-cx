package encoder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/skycoin/skycoin-cx/t0-plbolt"   // thin: only Get; no depth gain
	"github.com/skycoin/skycoin-cx/t0-plredis"  // thin: only Get; no depth gain
	"github.com/skycoin/skycoin-cx/t1-plloader" // thin: only File, Package, PackageList; types only, could move them to t0, tier 2 to 1
)

func SavePackagesToDisk(packageName string, path string, database string) error {
	err := os.Mkdir(path, 0755)
	if err != nil {
		return err
	}

	var packageList loader.PackageList
	listBytes, err := getBytes(packageName, database)
	if err != nil {
		return err
	}
	if err := packageList.UnmarshalBinary(listBytes); err != nil {
		return fmt.Errorf("decode package list %s: %w", packageName, err)
	}

	for _, pack := range packageList.Packages {
		var packageStruct loader.Package
		packageBytes, err := getBytes(pack, database)
		if err != nil {
			return err
		}
		if err := packageStruct.UnmarshalBinary(packageBytes); err != nil {
			return fmt.Errorf("decode package %s: %w", pack, err)
		}

		filePath := filepath.Join(path, "src")
		if packageStruct.PackageName != "main" {
			if err := checkPathElement(packageStruct.PackageName); err != nil {
				return fmt.Errorf("package %s: %w", pack, err)
			}
			filePath = filepath.Join(filePath, packageStruct.PackageName)
		}

		err = SaveFilesToDisk(packageStruct, filePath, database)
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveFilesToDisk(packageStruct loader.Package, filePath string, database string) error {
	err := os.MkdirAll(filePath, 0755)
	if err != nil {
		return err
	}
	for _, file := range packageStruct.Files {
		var fileStruct loader.File
		fileBytes, err := getBytes(file, database)
		if err != nil {
			return err
		}
		if err := fileStruct.UnmarshalBinary(fileBytes); err != nil {
			return fmt.Errorf("decode file %s of package %s: %w", file, packageStruct.PackageName, err)
		}

		if err := checkPathElement(fileStruct.FileName); err != nil {
			return fmt.Errorf("file %s of package %s: %w", file, packageStruct.PackageName, err)
		}
		err = os.WriteFile(filepath.Join(filePath, fileStruct.FileName), fileStruct.Content, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func getBytes(key string, database string) ([]byte, error) {
	switch database {
	case "redis":
		value, err := redis.Get(key)
		if err != nil {
			return nil, err
		}
		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("redis: value of %q is %T, want string", key, value)
		}
		return []byte(s), nil
	case "bolt":
		return bolt.Get(key)
	}
	return nil, fmt.Errorf("unknown database %q", database)
}

// checkPathElement rejects a name read from the database that is not a single
// path element, so stored data cannot write outside the output directory.
func checkPathElement(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) || filepath.Base(name) != name || filepath.IsAbs(name) || filepath.VolumeName(name) != "" {
		return fmt.Errorf("unsafe name %q", name)
	}
	return nil
}
