package file_output

import (
	"errors"

	"github.com/skycoin/skycoin-cx/t1-plloader"
)

// ImportFilesGet returns every file of the program stored under programName.
func ImportFilesGet(programName string) (files []*loader.File, err error) {
	if programName == "" {
		return files, errors.New("GetImportFiles: empty program name")
	}
	stored, err := loader.ProgramRead(programName)
	if err != nil {
		return files, err
	}
	for _, pkg := range stored {
		for i := range pkg.Files {
			files = append(files, &pkg.Files[i])
		}
	}
	return files, nil
}
