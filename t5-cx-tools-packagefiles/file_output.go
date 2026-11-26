package packagefiles

import (
	"errors"

	"github.com/skycoin/skycoin-cx/t1-cx-tools-packageloader"
)

// ImportFilesGet returns every file of the program stored under programName.
func ImportFilesGet(programName string) (files []*packageloader.File, err error) {
	if programName == "" {
		return files, errors.New("GetImportFiles: empty program name")
	}
	stored, err := packageloader.ProgramRead(programName)
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
