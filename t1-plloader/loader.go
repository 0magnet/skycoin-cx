package loader

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/skycoin/skycoin-cx/t0-plbolt" // thin: only Add; no depth gain
	"github.com/skycoin/skycoin-cx/t0-util"   // thin: only CXOpenFile, CXStatFile; no depth gain
	"golang.org/x/crypto/blake2b"
)

// NOTE: t1-plloader and t1-plloader2 are near copies of each other. See #91.
// NOTE: unused here, and a hand kept copy of the default packages. See #95.
var builtinPackages = []string{"al", "gl", "glfw", "time", "os", "gltext", "cx", "json", "cipher", "tcp"}

// VIOLATION: state outside State. Owner: t1-plloader; moves to PackageFileMng in phase 1.
// NOTE: written but never read, and keyed by base file name, so files in different
// packages overwrite each other. See #94.
var FileHashMap = make(map[string]string)

// VIOLATION: state outside State. Owner: t1-plloader; moves to PackageFileMng in phase 1.
// NOTE: written but never read. See #94.
var PackageHashMap = make(map[string]string)

func Contains(list []string, element string) bool {
	for _, elem := range list {
		if elem == element {
			return true
		}
	}
	return false
}

// NOTE: drops a file whose package does not match its directory without an error,
// unlike t1-plloader2. The tests depend on it. See #96.
func createFileMap(files []*os.File) (fileMap map[string][]*os.File, err error) {
	fileMap = make(map[string][]*os.File)
	for _, file := range files {
		if file == nil {
			return fileMap, errors.New("create file map: nil source file")
		}
		absPath, err := filepath.Abs(file.Name())
		if err != nil {
			return fileMap, fmt.Errorf("%s: resolve path: %w", file.Name(), err)
		}
		packageName := filepath.Base(filepath.Dir(absPath))
		filePackageName, err := getPackageName(file)
		if err != nil {
			return fileMap, err
		}
		if packageName == "src" {
			if filePackageName == "main" {
				fileMap["main"] = append(fileMap["main"], file)
				continue
			}
		}
		if filePackageName == packageName {
			fileMap[packageName] = append(fileMap[packageName], file)
		}
	}
	return fileMap, nil
}

func getPackageName(file *os.File) (string, error) {
	if file == nil {
		return "", errors.New("get package name: nil file")
	}
	openFile, err := os.Open(file.Name())
	if err != nil {
		return "", err
	}
	defer openFile.Close()

	scanner := bufio.NewScanner(openFile)

	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		if len(line) > 0 && line[0] == "package" {
			if len(line) < 2 {
				return "", fmt.Errorf("%s: package clause without a name", file.Name())
			}
			return line[1], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("%s: read: %w", file.Name(), err)
	}
	return "", errors.New("file doesn't contain a package name")
}

func createImportMap(fileMap map[string][]*os.File) (importMap map[string][]string, err error) {
	importMap = make(map[string][]string)
	for packageName := range fileMap {
		packageImports := []string{}
		for _, file := range fileMap[packageName] {
			newImports, err := getImports(file)
			if err != nil {
				return importMap, err
			}
			packageImports = append(packageImports, newImports...)
		}
		importMap[packageName] = packageImports
	}
	return importMap, nil
}

func getImports(file *os.File) (imports []string, err error) {
	if file == nil {
		return nil, errors.New("get imports: nil file")
	}
	openFile, err := os.Open(file.Name())
	if err != nil {
		return imports, err
	}
	defer openFile.Close()

	scanner := bufio.NewScanner(openFile)

	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		if len(line) == 0 || line[0] != "import" {
			continue
		}
		if len(line) < 2 || len(line[1]) < 2 || line[1][0] != '"' || line[1][len(line[1])-1] != '"' {
			return nil, fmt.Errorf("%s: malformed import line %q", file.Name(), scanner.Text())
		}
		imports = append(imports, line[1][1:len(line[1])-1])
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: read: %w", file.Name(), err)
	}
	return imports, nil
}

// NOTE: finds only self imports and loops of two packages, a -> b -> c -> a passes.
// See #97.
func checkForDependencyLoop(importMap map[string][]string) (err error) {
	for packageName := range importMap {
		for _, importName := range importMap[packageName] {
			if importName == packageName {
				return errors.New("Module " + packageName + " imports itself")
			}
			if Contains(importMap[importName], packageName) {
				return errors.New("Dependency loop between modules " + packageName + " and " + importName)
			}
		}
	}
	return nil
}

func LoadCXProgram(programName string, sourceCode []*os.File) (err error) {
	if programName == "" {
		return errors.New("load program: empty program name")
	}
	if len(sourceCode) == 0 {
		return fmt.Errorf("load program %s: no source files", programName)
	}
	fileMap, err := createFileMap(sourceCode)
	if err != nil {
		return err
	}
	importMap, err := createImportMap(fileMap)
	if err != nil {
		return err
	}
	err = checkForDependencyLoop(importMap)
	if err != nil {
		return
	}

	packageNames := make([]string, 0, len(fileMap))
	for key := range fileMap {
		packageNames = append(packageNames, key)
	}
	sort.Strings(packageNames)

	packageListStruct := PackageList{}
	for _, key := range packageNames {
		err = addNewPackage(&packageListStruct, key, fileMap[key])
		if err != nil {
			return err
		}
	}

	value, err := packageListStruct.MarshalBinary()
	if err != nil {
		return err
	}
	return bolt.Add(programName, value)
}

func addNewPackage(packageListStruct *PackageList, packageName string, files []*os.File) error {
	if packageListStruct == nil {
		return errors.New("add package: nil package list")
	}
	if packageName == "" {
		return errors.New("add package: empty package name")
	}
	packageStruct := Package{
		PackageName: packageName,
	}

	for _, file := range files {
		fileStruct, err := fileStructFromFile(file)
		if err != nil {
			return err
		}
		err = packageStruct.appendFile(&fileStruct)
		if err != nil {
			return err
		}
	}

	return packageListStruct.appendPackage(&packageStruct)
}

func fileStructFromFile(file *os.File) (File, error) {
	if file == nil {
		return File{}, errors.New("file struct: nil file")
	}
	fileName := filepath.Base(file.Name())
	fileBytes, err := ioutil.ReadFile(file.Name())
	if err != nil {
		return File{}, fmt.Errorf("%s: read: %w", file.Name(), err)
	}
	if uint64(len(fileBytes)) > math.MaxUint32 {
		return File{}, fmt.Errorf("%s: %d bytes does not fit the uint32 Length field", file.Name(), len(fileBytes))
	}
	fileHash := blake2b.Sum512(fileBytes)

	fileStruct := File{
		FileName:   fileName,
		Length:     uint32(len(fileBytes)),
		Content:    fileBytes,
		Blake2Hash: string(fileHash[:]),
	}

	return fileStruct, nil
}

// ParseArgsForCX splits args into "++" cx flags and .cx source files. A
// directory argument adds every .cx file in it, and in its subdirectories
// when alsoSubdirs is set. On error every file it opened is closed again.
func ParseArgsForCX(args []string, alsoSubdirs bool) (cxArgs []string, sourceCode []*os.File, fileNames []string, err error) {
	defer func() {
		if err != nil {
			for _, file := range sourceCode {
				file.Close()
			}
			cxArgs, sourceCode, fileNames = nil, nil, nil
		}
	}()

	skip := false // flag for skipping arg

	for _, arg := range args {

		// skip arg if skip flag is specified
		if skip {
			skip = false
			continue
		}

		// cli flags are either "--key=value" or "-key value"
		// we have to skip both cases
		// NOTE: assumes every flag without = takes a value, so a bool flag swallows the
		// next argument, which may be a source file. See #101.
		if len(arg) > 1 && arg[0] == '-' {
			if !strings.Contains(arg, "=") {
				skip = true
			}
			continue
		}

		// cli cx flags are prefixed with "++"
		if len(arg) > 2 && arg[:2] == "++" {
			cxArgs = append(cxArgs, arg)
			continue
		}

		arg = filepath.Clean(arg)
		// NOTE: CXStatFile, the directory walk and CXOpenFile apply the working directory
		// in different ways. See #100.
		fi, err := util.CXStatFile(arg)
		if err != nil {
			return cxArgs, sourceCode, fileNames, fmt.Errorf("%s: source file or library not found: %w", arg, err)
		}

		switch mode := fi.Mode(); {
		case mode.IsDir():
			var fileList []string

			// Checking if we want to check all subdirectories.
			if alsoSubdirs {
				fileList, err = listDirRecursive(arg)
			} else {
				fileList, err = listDir(arg)
			}
			if err != nil {
				return cxArgs, sourceCode, fileNames, fmt.Errorf("%s: list source files: %w", arg, err)
			}

			for _, path := range fileList {
				// only loading .cx files
				if filepath.Ext(path) != ".cx" {
					continue
				}
				file, err := util.CXOpenFile(path)
				if err != nil {
					return cxArgs, sourceCode, fileNames, fmt.Errorf("%s: open source file: %w", path, err)
				}
				sourceCode = append(sourceCode, file)
				fileNames = append(fileNames, file.Name())
			}
		case mode.IsRegular():
			file, err := util.CXOpenFile(arg)
			if err != nil {
				return cxArgs, sourceCode, fileNames, fmt.Errorf("%s: open source file: %w", arg, err)
			}
			fileNames = append(fileNames, file.Name())
			sourceCode = append(sourceCode, file)
		default:
			return cxArgs, sourceCode, fileNames, fmt.Errorf("%s: not a regular file or directory", arg)
		}
	}

	return cxArgs, sourceCode, fileNames, nil
}

func listDirRecursive(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func listDir(root string) ([]string, error) {
	var files []string
	fileInfo, err := ioutil.ReadDir(root)
	if err != nil {
		return files, err
	}

	for _, file := range fileInfo {
		if file.IsDir() {
			continue
		}
		files = append(files, filepath.Join(root, file.Name()))
	}
	return files, nil
}
