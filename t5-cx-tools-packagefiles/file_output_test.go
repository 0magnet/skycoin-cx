package packagefiles_test

import (
	"testing"

	sourceargs "github.com/skycoin/skycoin-cx/t0-cx-compiler-sourceargs"
	"github.com/skycoin/skycoin-cx/t1-cx-tools-packageloader"
	actions "github.com/skycoin/skycoin-cx/t4-actions" // thin: tests use only AST; tests only set actions.AST = nil, could drop it, tier 5 to 2
	"github.com/skycoin/skycoin-cx/t5-cx-tools-packagefiles"
)

func TestFile_Output_GetImportFiles(t *testing.T) {

	tests := []struct {
		scenario    string
		programName string
		testDir     string
		files       []packageloader.File
	}{
		{
			scenario:    "Has Imports",
			programName: "tester",
			testDir:     "./test_files/test_valid_program",
			files: []packageloader.File{
				{
					FileName: "testimport.cx",
				},
				{
					FileName: "testfile.cx",
				},
				{
					FileName: "testfile.cx",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {

			actions.AST = nil

			_, sourceCodes, _, perr := sourceargs.Parse([]string{tc.testDir}, true)
			if perr != nil {
				t.Fatal(perr)
			}
			err := packageloader.CXProgramLoad(tc.programName, sourceCodes)
			if err != nil {
				t.Fatal(err)
			}

			files, err := packagefiles.ImportFilesGet(tc.programName)
			if err != nil {
				t.Fatal(err)
			}

			for _, wantFile := range tc.files {
				var match bool
				var wantFileName string = wantFile.FileName
				var gotFileName string
				for _, gotFile := range files {
					gotFileName = gotFile.FileName
					if gotFileName == wantFileName {
						match = true
						break
					}
				}

				if !match {
					t.Errorf("want file %s, got %s", wantFileName, gotFileName)
				}
			}

		})
	}
}
