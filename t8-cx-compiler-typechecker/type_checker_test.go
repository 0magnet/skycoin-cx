package typechecker_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	sourceargs "github.com/skycoin/skycoin-cx/t0-cx-compiler-sourceargs"
	"github.com/skycoin/skycoin-cx/t0-cx-core-syspackages" // thin: tests use only IsDefaultPackage; no depth gain
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
	"github.com/skycoin/skycoin-cx/t1-cx-tools-packageloader" // thin: tests use only LoadCXProgram; no depth gain
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"github.com/skycoin/skycoin-cx/t4-cx-compiler-astbuild"
	"github.com/skycoin/skycoin-cx/t5-cx-tools-packagefiles" // thin: tests use only GetImportFiles; no depth gain
	"github.com/skycoin/skycoin-cx/t6-cx-compiler-declarationextractor"
	cxinit "github.com/skycoin/skycoin-cx/t7-init" // thin: tests use only MakeProgram, RegisterPackages; could take the program as an argument, tier 8 to 7
	"github.com/skycoin/skycoin-cx/t8-cx-compiler-typechecker"
)

func TestTypeChecker_ParseDeclarationSpecifier(t *testing.T) {
	tests := []struct {
		scenario                                string
		testString                              string
		fileName                                string
		lineno                                  int
		strctName                               string
		pkgName                                 string
		wantDeclarationSpecifierFormattedString string
		wantLengths                             []int
		wantErr                                 error
	}{
		{
			scenario:                                "Has Type Specifier",
			testString:                              "str",
			fileName:                                "./myFile",
			lineno:                                  4,
			wantDeclarationSpecifierFormattedString: "str",
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has Indentifier",
			testString:                              "Animal",
			fileName:                                "./testFile",
			lineno:                                  10,
			strctName:                               "Animal",
			pkgName:                                 "",
			wantDeclarationSpecifierFormattedString: "Animal",
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has External Indentifier",
			testString:                              "tester.Direction",
			fileName:                                "./myFile",
			lineno:                                  15,
			strctName:                               "Direction",
			pkgName:                                 "tester",
			wantDeclarationSpecifierFormattedString: "Direction",
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has External Type Specifier",
			testString:                              "i32.counter",
			fileName:                                "./myFile",
			lineno:                                  23,
			strctName:                               "counter",
			pkgName:                                 "i32",
			wantDeclarationSpecifierFormattedString: "counter",
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has Array ",
			testString:                              "[5]i64",
			fileName:                                "./testFile",
			lineno:                                  67,
			wantDeclarationSpecifierFormattedString: "[5]i64",
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has Slice",
			testString:                              "[]str",
			fileName:                                "./myFile",
			lineno:                                  45,
			wantDeclarationSpecifierFormattedString: "[]str",
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has Pointer",
			testString:                              "*ui8",
			fileName:                                "./testFile",
			lineno:                                  23,
			wantDeclarationSpecifierFormattedString: "*ui8",
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has Pointer Array",
			testString:                              "[5]*ui8",
			fileName:                                "./testFile",
			lineno:                                  6,
			wantDeclarationSpecifierFormattedString: "[5]*ui8",
			wantLengths:                             []int{5},
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has Array Pointer",
			testString:                              "*[12]str",
			fileName:                                "./testFile",
			lineno:                                  68,
			wantDeclarationSpecifierFormattedString: "*[12]str",
			wantErr:                                 nil,
		},
		// Lengths are outermost first, as the grammar stores them. This test used to
		// print the same string from lengths stored innermost first, {10, 5}.
		{
			scenario:                                "Has 2D Array",
			testString:                              "[5][10]i64",
			fileName:                                "./testFile",
			lineno:                                  21,
			wantDeclarationSpecifierFormattedString: "[5][10]i64",
			wantLengths:                             []int{5, 10},
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has 3D Array",
			testString:                              "[5][1][23]bool",
			fileName:                                "./testFile",
			lineno:                                  34,
			wantDeclarationSpecifierFormattedString: "[5][1][23]bool",
			wantLengths:                             []int{5, 1, 23},
			wantErr:                                 nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {
			// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
			// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S1528 S1529.
			astbuild.AST = cxinit.ProgramMake()
			cxinit.PackagesRegister(astbuild.AST)

			if tc.strctName != "" {

				// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
				// Now a constructor call, site S1530.
				pkg := ast.PackageMake(astbuild.AST.CXProgramNameIntern(tc.pkgName))
				// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
				// Now an arena insert, site S1531.
				pkgIdx := astbuild.AST.PackageAdd(pkg)
				// OBJMODEL: state.CXStructMng.CXStructAlloc(), phase 4 #336.
				// Now a constructor call, site S1532.
				strct := ast.StructMake(astbuild.AST.CXProgramNameIntern(tc.strctName))
				strct.Package = ast.CXPackageIndex(pkgIdx)
				pkg = pkg.StructAdd(astbuild.AST, strct)

			}

			if tc.pkgName != "" {
				astbuild.ImportDeclare(astbuild.AST, tc.pkgName, astbuild.AST.CXProgramSourceFileIntern(tc.fileName), tc.lineno)
			}

			var gotDeclarationSpecifier *ast.CXArgument
			gotDeclarationSpecifier, gotErr := typechecker.DeclarationSpecifierParse(astbuild.AST, []byte(tc.testString), astbuild.AST.CXProgramSourceFileIntern(tc.fileName), tc.lineno, gotDeclarationSpecifier)
			gotTypeSignature := ast.CXTypeSignatureRepresentationOfCXArgGet(astbuild.AST, gotDeclarationSpecifier)
			gotDeclarationSpecifierFormattedString := ast.FormattedTypeGet(astbuild.AST, gotTypeSignature)

			if gotDeclarationSpecifierFormattedString != tc.wantDeclarationSpecifierFormattedString {
				t.Errorf("want %s, got %s", tc.wantDeclarationSpecifierFormattedString, gotDeclarationSpecifierFormattedString)
			}
			if tc.wantLengths != nil && fmt.Sprint(gotDeclarationSpecifier.Lengths) != fmt.Sprint(tc.wantLengths) {
				t.Errorf("want lengths %v, got %v", tc.wantLengths, gotDeclarationSpecifier.Lengths)
			}

			if (gotErr != nil && tc.wantErr == nil) ||
				(gotErr == nil && tc.wantErr != nil) {
				t.Errorf("want error %v, got %v", tc.wantErr, gotErr)
			}

			if gotErr != nil && tc.wantErr != nil {
				if gotErr.Error() != tc.wantErr.Error() {
					t.Errorf("want error %v, got %v", tc.wantErr, gotErr)
				}
			}
		})
	}
}

func TestTypeChecker_ParseParameterDeclaration(t *testing.T) {
	tests := []struct {
		scenario                                string
		testString                              string
		fileName                                string
		lineno                                  int
		strctName                               string
		pkgName                                 string
		wantParameterDeclarationFormattedString string
		wantErr                                 error
	}{
		{
			scenario:                                "Has Type Specifier",
			testString:                              "name str",
			fileName:                                "./myFile",
			lineno:                                  4,
			wantParameterDeclarationFormattedString: "name str",
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has Indentifier",
			testString:                              "cat Animal",
			fileName:                                "./testFile",
			lineno:                                  10,
			strctName:                               "Animal",
			pkgName:                                 "",
			wantParameterDeclarationFormattedString: "cat Animal",
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has External Indentifier",
			testString:                              "North tester.Direction",
			fileName:                                "./myFile",
			lineno:                                  15,
			strctName:                               "Direction",
			pkgName:                                 "tester",
			wantParameterDeclarationFormattedString: "North Direction",
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has External Type Specifier",
			testString:                              "clock i32.counter",
			fileName:                                "./myFile",
			lineno:                                  23,
			strctName:                               "counter",
			pkgName:                                 "i32",
			wantParameterDeclarationFormattedString: "clock counter",
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has Array ",
			testString:                              "lottery [5]i64",
			fileName:                                "./testFile",
			lineno:                                  67,
			wantParameterDeclarationFormattedString: "lottery [5]i64",
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has Slice",
			testString:                              "months []str",
			fileName:                                "./myFile",
			lineno:                                  45,
			wantParameterDeclarationFormattedString: "months []str",
			wantErr:                                 nil,
		},
		{
			scenario:                                "Has Pointer",
			testString:                              "link *ui8",
			fileName:                                "./testFile",
			lineno:                                  23,
			wantParameterDeclarationFormattedString: "link *ui8",
			wantErr:                                 nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {
			// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
			// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S1533 S1534.
			astbuild.AST = cxinit.ProgramMake()
			cxinit.PackagesRegister(astbuild.AST)

			var pkg *ast.CXPackage
			// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
			// Now a constructor call, site S1535.
			pkg = ast.PackageMake(astbuild.AST.CXProgramNameIntern(tc.pkgName))
			// OBJMODEL: state.CXPackageMng.CXPackageAlloc(), phase 4 #361.
			// Now an arena insert, site S1536.
			pkgIdx := astbuild.AST.PackageAdd(pkg)

			if tc.strctName != "" {
				// OBJMODEL: state.CXStructMng.CXStructAlloc(), phase 4 #336.
				// Now a constructor call, site S1537.
				strct := ast.StructMake(astbuild.AST.CXProgramNameIntern(tc.strctName))
				strct.Package = ast.CXPackageIndex(pkgIdx)
				pkg = pkg.StructAdd(astbuild.AST, strct)
			}

			if tc.pkgName != "" {
				astbuild.ImportDeclare(astbuild.AST, tc.pkgName, astbuild.AST.CXProgramSourceFileIntern(tc.fileName), tc.lineno)
			}

			var gotParameterDeclaration *ast.CXArgument
			gotParameterDeclaration, gotErr := typechecker.ParameterDeclarationParse(astbuild.AST, []byte(tc.testString), pkg, astbuild.AST.CXProgramSourceFileIntern(tc.fileName), tc.lineno)
			gotParameterDeclarationFormattedName := ast.FormattedNameGet(astbuild.AST, gotParameterDeclaration, false, pkg)
			gotTypeSignature := ast.CXTypeSignatureRepresentationOfCXArgGet(astbuild.AST, gotParameterDeclaration)
			gotParameterDeclarationFormattedType := ast.FormattedTypeGet(astbuild.AST, gotTypeSignature)
			gotParameterDeclarationFormattedString := gotParameterDeclarationFormattedName + " " + gotParameterDeclarationFormattedType

			if gotParameterDeclarationFormattedString != tc.wantParameterDeclarationFormattedString {
				t.Errorf("want %s, got %s", tc.wantParameterDeclarationFormattedString, gotParameterDeclarationFormattedString)
			}

			if (gotErr != nil && tc.wantErr == nil) ||
				(gotErr == nil && tc.wantErr != nil) {
				t.Errorf("want error %v, got %v", tc.wantErr, gotErr)
			}

			if gotErr != nil && tc.wantErr != nil {
				if gotErr.Error() != tc.wantErr.Error() {
					t.Errorf("want error %v, got %v", tc.wantErr, gotErr)
				}
			}
		})
	}
}

func TestTypeChecker_ParseImports(t *testing.T) {

	tests := []struct {
		scenario string
		testDir  string
		imports  []string
	}{
		{
			scenario: "Has Imports",
			testDir:  "./test_files/ParseImports/HasImports.cx",
			imports: []string{
				"myImport",
				"anotherImport",
			},
		},
		{
			scenario: "Has Imports 2",
			testDir:  "./test_files/ParseImports/HasImports2.cx",
			imports: []string{
				"os",
				"testimport2",
				"testimport1",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {

			// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
			// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S1538 S1539.
			astbuild.AST = cxinit.ProgramMake()
			cxinit.PackagesRegister(astbuild.AST)

			srcBytes, err := os.ReadFile(tc.testDir)
			if err != nil {
				t.Error(err)
			}

			ReplaceCommentsWithWhitespaces := declarationextractor.ReplaceCommentsWithWhitespaces(srcBytes)

			Imports, err := declarationextractor.ImportsExtract(ReplaceCommentsWithWhitespaces, tc.testDir)
			if err != nil {
				t.Fatal(err)
			}

			err = typechecker.ImportsParse(Imports)

			program := astbuild.AST

			for _, wantImport := range tc.imports {

				var match bool
				var gotImport string
				for _, pkgIdx := range program.Packages {
					// OBJMODEL: state.CXPackageMng.CXPackageGet(gid), phase 5 #362.
					// Now an arena getter, site S1540.
					pkg, err := program.PackageFromArrayGet(pkgIdx)
					if err != nil {
						panic(err)
					}

					if syspackages.PackageDefault(astbuild.AST.CXProgramNameText(pkg.Name)) {
						continue
					}

					for _, impIdx := range pkg.Imports {
						// OBJMODEL: state.CXPackageMng.CXPackageGet(gid), phase 5 #362.
						// Now an arena getter, site S1541.
						impPkg, err := program.PackageFromArrayGet(impIdx)
						if err != nil {
							panic(err)
						}

						gotImport = astbuild.AST.CXProgramNameText(impPkg.Name)

						if gotImport == wantImport {
							match = true
							break
						}
					}
				}

				if !match {
					t.Errorf("want import %s, got %s", wantImport, gotImport)
				}
			}

		})
	}
}

func TestTypeChecker_ParseGlobals(t *testing.T) {

	type GlobalTypeSignature struct {
		Package string
		Index   int
		Name    string
		Type    string
	}

	tests := []struct {
		scenario             string
		testDir              string
		globalTypeSignatures []GlobalTypeSignature
	}{
		{
			scenario: "Has globals",
			testDir:  "./test_files/ParseGlobals/HasGlobals",
			globalTypeSignatures: []GlobalTypeSignature{
				{
					Name:    "Bool",
					Index:   0,
					Package: "main",
					Type:    "bool",
				},
				{
					Name:    "Byte",
					Index:   1,
					Package: "main",
					Type:    "i8",
				},
				{
					Name:    "I16",
					Index:   2,
					Package: "main",
					Type:    "i16",
				},
				{
					Name:    "I32",
					Index:   3,
					Package: "main",
					Type:    "i32",
				},
				{
					Name:    "I64",
					Index:   4,
					Package: "main",
					Type:    "i64",
				},
				{
					Name:    "UByte",
					Index:   5,
					Package: "main",
					Type:    "ui8",
				},
				{
					Name:    "UI16",
					Index:   6,
					Package: "main",
					Type:    "ui16",
				},
				{
					Name:    "UI32",
					Index:   7,
					Package: "main",
					Type:    "ui32",
				},
				{
					Name:    "UI64",
					Index:   8,
					Package: "main",
					Type:    "ui64",
				},
				{
					Name:    "F32",
					Index:   9,
					Package: "main",
					Type:    "f32",
				},
				{
					Name:    "F64",
					Index:   10,
					Package: "main",
					Type:    "f64",
				},
				{
					Name:    "string",
					Index:   11,
					Package: "main",
					Type:    "str",
				},
				{
					Name:    "Affordance",
					Index:   12,
					Package: "main",
					Type:    "aff",
				},
				{
					Name:    "intArray",
					Index:   13,
					Package: "main",
					Type:    "[5]i32",
				},
			},
		},
		{
			scenario: "Has globals 2",
			testDir:  "./test_files/ParseGlobals/HasGlobals2",
			globalTypeSignatures: []GlobalTypeSignature{
				{
					Name:    "number",
					Index:   0,
					Package: "main",
					Type:    "i32",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {

			// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
			// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S1542 S1543.
			astbuild.AST = cxinit.ProgramMake()
			cxinit.PackagesRegister(astbuild.AST)

			_, sourceCode, _, perr := sourceargs.Parse([]string{tc.testDir}, true)
			if perr != nil {
				t.Fatal(perr)
			}

			err := packageloader.CXProgramLoad("test", sourceCode)
			if err != nil {
				t.Fatal(err)
			}

			files, err := packagefiles.ImportFilesGet("test")
			if err != nil {
				t.Fatal(err)
			}

			ReplaceCommentsWithWhitespaces := declarationextractor.ReplaceCommentsWithWhitespaces(files[0].Content)
			ReplaceStringContentsWithWhitespaces, err := declarationextractor.ReplaceStringContentsWithWhitespaces(ReplaceCommentsWithWhitespaces)
			if err != nil {
				t.Fatal(err)
			}

			Globals, err := declarationextractor.GlobalsExtract(ReplaceStringContentsWithWhitespaces, files[0].FileName)
			if err != nil {
				t.Fatal(err)
			}

			err = typechecker.GlobalsParse(files, Globals)
			if err != nil {
				t.Fatal(err)
			}

			program := astbuild.AST

			// Index is the global's position in its package, since the
			// native packages register type signatures before main.
			for _, wantGlobal := range tc.globalTypeSignatures {
				pkg, err := program.PackageGet(program.CXProgramNameIntern(wantGlobal.Package))
				if err != nil {
					t.Fatal(err)
				}

				gotIndex := -1
				var gotGlobalType string
				for i, globalIdx := range pkg.Globals.Fields {
					// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
					// Now an arena getter, site S1544.
					gotGlobal := program.CXTypeSignatureFromArrayGet(globalIdx)
					if program.CXProgramNameText(gotGlobal.Name) == wantGlobal.Name {
						gotIndex = i
						gotGlobalType = ast.FormattedTypeGet(astbuild.AST, gotGlobal)
						break
					}
				}

				if gotIndex != wantGlobal.Index || gotGlobalType != wantGlobal.Type {
					t.Errorf("want global %v %v %v %v, got index %v type %q", wantGlobal.Name, wantGlobal.Index, wantGlobal.Package, wantGlobal.Type, gotIndex, gotGlobalType)
				}
			}
		})
	}
}

// func TestTypeChecks_ParseEnums(t *testing.T) {

// 	tests := []struct {
// 		scenario string
// 		testDir  string
// 	}{}

// 	for _, tc := range tests {
// 		t.Run(tc.scenario, func(t *testing.T) {

// 		})
// 	}

// }

func TestTypeChecker_ParseStructs(t *testing.T) {

	type StructFieldTypeSignature struct {
		Index int
		Name  string
		Type  string
	}

	type StructTypeSignature struct {
		Package string
		Index   int
		Name    string
		Type    string
		Fields  []StructFieldTypeSignature
	}

	tests := []struct {
		scenario             string
		testDir              string
		structTypeSignatures []StructTypeSignature
	}{
		{
			scenario: "Has Structs",
			testDir:  "./test_files/ParseStructs/HasStructs",
			structTypeSignatures: []StructTypeSignature{
				{
					Name:    "CustomType",
					Index:   0,
					Package: "main",
					Fields: []StructFieldTypeSignature{
						{
							Index: 0,
							Name:  "fieldA",
							Type:  "str",
						},
						{
							Index: 1,
							Name:  "fieldB",
							Type:  "i32",
						},
					},
				},
				{
					Name:    "AnotherType",
					Index:   1,
					Package: "main",
					Fields: []StructFieldTypeSignature{
						{
							Index: 0,
							Name:  "name",
							Type:  "str",
						},
					},
				},
			},
		},
		{
			scenario: "Has Structs 2",
			testDir:  "./test_files/ParseStructs/HasStructs2",
			structTypeSignatures: []StructTypeSignature{
				{
					Name:    "animal",
					Index:   0,
					Package: "main",
					Fields: []StructFieldTypeSignature{
						{
							Index: 0,
							Name:  "name",
							Type:  "str",
						},
						{
							Index: 1,
							Name:  "legs",
							Type:  "i32",
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {

			// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
			// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S1545 S1546.
			astbuild.AST = cxinit.ProgramMake()
			cxinit.PackagesRegister(astbuild.AST)

			_, sourceCode, _, perr := sourceargs.Parse([]string{tc.testDir}, true)
			if perr != nil {
				t.Fatal(perr)
			}

			err := packageloader.CXProgramLoad("test", sourceCode)
			if err != nil {
				t.Fatal(err)
			}

			files, err := packagefiles.ImportFilesGet("test")
			if err != nil {
				t.Fatal(err)
			}

			ReplaceCommentsWithWhitespaces := declarationextractor.ReplaceCommentsWithWhitespaces(files[0].Content)

			structs, err := declarationextractor.StructsExtract(ReplaceCommentsWithWhitespaces, files[0].FileName)

			// The native packages register structs first, so count from the first new one.
			firstStruct := len(astbuild.AST.CXStructs)
			typechecker.StructsParse(files, structs)

			program := astbuild.AST

			for _, wantStruct := range tc.structTypeSignatures {

				var match bool
				var ast3 string
				// OBJMODEL: state.CXStructMng.CXStructAlloc(), phase 4 #336.
				// Now a local object, site S1547.
				var gotStruct ast.CXStruct

				for _, pkgIdx := range program.Packages {

					// OBJMODEL: state.CXPackageMng.CXPackageGet(gid), phase 5 #362.
					// Now an arena getter, site S1548.
					pkg, err := program.PackageFromArrayGet(pkgIdx)
					gotPkgName := pkg.Name

					if err != nil {
						t.Log(err)
					}

					if syspackages.PackageDefault(astbuild.AST.CXProgramNameText(pkg.Name)) {
						continue
					}

					for _, structIdx := range pkg.Structs {
						gotStruct = program.CXStructs[structIdx]

						if astbuild.AST.CXProgramNameText(gotStruct.Name) == wantStruct.Name &&
							astbuild.AST.CXProgramNameText(gotPkgName) == wantStruct.Package {
							ast3 += fmt.Sprintf("want %s %d. %s, got %s %d. %s\n", wantStruct.Package, wantStruct.Index, wantStruct.Name, astbuild.AST.CXProgramNameText(gotPkgName), gotStruct.Index-firstStruct, astbuild.AST.CXProgramNameText(gotStruct.Name))

							if gotStruct.Index-firstStruct != wantStruct.Index || len(gotStruct.Fields) != len(wantStruct.Fields) {
								ast3 += fmt.Sprintf("want %d fields, got %d fields", len(wantStruct.Fields), len(gotStruct.Fields))
								break
							}

							for _, wantField := range wantStruct.Fields {

								for _, gotFieldIdx := range gotStruct.Fields {

									// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
									// Now an arena getter, site S1549.
									gotField := program.CXTypeSignatureFromArrayGet(gotFieldIdx)
									var gotFieldType string
									var gotFieldIndex int

									if program.CXProgramNameText(gotField.Name) == wantField.Name {

										gotFieldType = ast.FormattedTypeGet(astbuild.AST, gotField)

										ast3 += fmt.Sprintf("want field %d. %s %s, got %d. %s %s\n", wantField.Index, wantField.Name, wantField.Type, gotFieldIndex, program.CXProgramNameText(gotField.Name), gotFieldType)

										if gotFieldIndex == wantField.Index &&
											gotFieldType == wantField.Type {
											match = true
											break
										}
									}

								}

							}

							break
						}

					}

				}

				if !match {
					t.Error(ast3)
				}

			}

		})
	}

}

func TestTypeChecker_ParseFuncHeaders(t *testing.T) {

	type CXFunc struct {
		Name    string
		Index   int
		Package string
		Inputs  string
		Outputs string
	}

	tests := []struct {
		scenario    string
		testDir     string
		functionCXs []CXFunc
	}{
		{
			scenario: "Has funcs",
			testDir:  "./test_files/ParseFuncs/HasFuncs",
			functionCXs: []CXFunc{
				{
					Name:    "main",
					Index:   0,
					Package: "main",
				},
				{
					Name:    "add",
					Index:   1,
					Package: "main",
					Inputs:  "a i32, b i32",
					Outputs: "answer i32",
				},
				{
					Name:    "divide",
					Index:   2,
					Package: "main",
					Inputs:  "c i32, d i32",
					Outputs: "quotient i32, remainder f32",
				},
				{
					Name:    "printer",
					Index:   3,
					Package: "main",
					Inputs:  "message str",
				},
			},
		},
		{
			scenario: "Has funcs 2",
			testDir:  "./test_files/ParseFuncs/HasFuncs2",
			functionCXs: []CXFunc{
				{
					Name:    "main",
					Index:   0,
					Package: "main",
				},
				{
					Name:    "is_divisible",
					Index:   1,
					Package: "main",
					Inputs:  "x i32, y i32",
					Outputs: "result bool",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {

			// OBJMODEL: state.CXProgramMng.CXProgramAlloc(), phase 4 #371.
			// Also state.CXProgramMng.CXProgramRelease(gid), phase 4 #371. Site S1550 S1551.
			astbuild.AST = cxinit.ProgramMake()
			cxinit.PackagesRegister(astbuild.AST)

			_, sourceCode, _, perr := sourceargs.Parse([]string{tc.testDir}, true)
			if perr != nil {
				t.Fatal(perr)
			}

			err := packageloader.CXProgramLoad("test", sourceCode)
			if err != nil {
				t.Fatal(err)
			}

			files, err := packagefiles.ImportFilesGet("test")
			if err != nil {
				t.Fatal(err)
			}

			ReplaceCommentsWithWhitespaces := declarationextractor.ReplaceCommentsWithWhitespaces(files[0].Content)
			ReplaceStringContentsWithWhitespaces, err := declarationextractor.ReplaceStringContentsWithWhitespaces(ReplaceCommentsWithWhitespaces)
			if err != nil {
				t.Fatal(err)
			}

			funcs, err := declarationextractor.FuncsExtract(ReplaceStringContentsWithWhitespaces, files[0].FileName)

			typechecker.FuncHeadersParse(files, funcs)

			program := astbuild.AST

			for _, wantFunc := range tc.functionCXs {

				var match bool

				var wantFuncName string = wantFunc.Name
				var wantFuncInputs string = wantFunc.Inputs
				var wantFuncOutputs string = wantFunc.Outputs

				var gotFunc *ast.CXFunction
				var gotFuncName string
				var gotFuncInputs string
				var gotFuncOutputs string

				for _, pkgIdx := range program.Packages {

					// OBJMODEL: state.CXPackageMng.CXPackageGet(gid), phase 5 #362.
					// Now an arena getter, site S1552.
					pkg, err := program.PackageFromArrayGet(pkgIdx)

					if err != nil {
						t.Log(err)
					}

					for _, gotFuncIdx := range pkg.Functions {
						gotFunc = program.CxIrState.CXFunctionMng.CXFunctionGet(gotFuncIdx)
						gotFuncName = astbuild.AST.CXProgramNameText(gotFunc.Name)

						var inps bytes.Buffer
						var outs bytes.Buffer

						getFormattedParam(program, gotFunc.InputsGet(program), pkg, &inps)
						getFormattedParam(program, gotFunc.OutputsGet(program), pkg, &outs)

						gotFuncInputs = inps.String()
						gotFuncOutputs = outs.String()

						if gotFuncName == wantFuncName &&
							gotFunc.Index == wantFunc.Index &&
							astbuild.AST.CXProgramNameText(pkg.Name) == wantFunc.Package {
							if gotFuncInputs == wantFuncInputs && gotFuncOutputs == wantFuncOutputs {
								match = true
							}
							break
						}

					}

				}

				if !match {
					t.Errorf("want func %v %d. %s (%s) (%s), got %v %d. %s (%s) (%s)", wantFunc.Package, wantFunc.Index, wantFuncName, wantFuncInputs, wantFuncOutputs, gotFunc.Package, gotFunc.Index, gotFuncName, gotFuncInputs, gotFuncOutputs)
				}
			}

		})
	}

}

func TestTypeChecker_ParseAllDeclarations(t *testing.T) {

	type Global struct {
		Name string
		Type string
	}

	type StructField struct {
		Name string
		Type string
	}

	type Struct struct {
		Name   string
		Fields []StructField
	}

	type Func struct {
		Name    string
		Inputs  string
		Outputs string
	}

	type Package struct {
		Name    string
		Imports []string
		Globals []Global
		Structs []Struct
		Funcs   []Func
	}

	tests := []struct {
		scenario    string
		testDir     string
		wantProgram []Package
	}{
		{
			scenario: "Has Declarations",
			testDir:  "./test_files/ParseAllDeclarations/HasDeclarations",
			wantProgram: []Package{
				{
					Name: "main",
					Imports: []string{
						"helper",
						"config",
					},
					Globals: []Global{
						{
							Name: "Bool",
							Type: "bool",
						},
						{
							Name: "Byte",
							Type: "i8",
						},
						{
							Name: "I16",
							Type: "i16",
						},
						{
							Name: "I32",
							Type: "i32",
						},
						{
							Name: "I64",
							Type: "i64",
						},
						{
							Name: "UByte",
							Type: "ui8",
						},
						{
							Name: "UI16",
							Type: "ui16",
						},
						{
							Name: "UI32",
							Type: "ui32",
						},
						{
							Name: "UI64",
							Type: "ui64",
						},
						{
							Name: "F32",
							Type: "f32",
						},
						{
							Name: "F64",
							Type: "f64",
						},
						{
							Name: "string",
							Type: "str",
						},
						{
							Name: "Affordance",
							Type: "aff",
						},
						{
							Name: "intArray",
							Type: "[5]i32",
						},
					},
					Structs: []Struct{
						{
							Name: "CustomType",
							Fields: []StructField{
								{
									Name: "fieldA",
									Type: "str",
								},
								{
									Name: "fieldB",
									Type: "i32",
								},
							},
						},
						{
							Name: "AnotherType",
							Fields: []StructField{
								{
									Name: "name",
									Type: "str",
								},
							},
						},
					},
					Funcs: []Func{
						{
							Name: "main",
						},
						{
							Name:    "add",
							Inputs:  "a i32, b i32",
							Outputs: "answer i32",
						},
						{
							Name:    "divide",
							Inputs:  "c i32, d i32",
							Outputs: "quotient i32, remainder f32",
						},
						{
							Name:   "printer",
							Inputs: "message str",
						},
						{
							Name:   "CustomType.setFieldA",
							Inputs: "main.customType *CustomType, stringParam str",
						},
					},
				},
			},
		},
		{
			scenario: "Has Multiple Files",
			testDir:  "./test_files/ParseAllDeclarations/HasMultipleFiles",
			wantProgram: []Package{
				{
					Name: "main",
					Imports: []string{
						"math",
					},
					Globals: []Global{},
					Structs: []Struct{},
					Funcs: []Func{
						{
							Name: "main",
						},
					},
				},
				{
					Name:    "math",
					Globals: []Global{},
					Structs: []Struct{},
					Funcs: []Func{
						{
							Name:    "double",
							Outputs: "out i32",
						},
					},
				},
			},
		},
		{
			scenario: "Has Imports",
			testDir:  "./test_files/ParseAllDeclarations/HasImports",
			wantProgram: []Package{
				{
					Name: "main",
					Globals: []Global{
						{
							Name: "dog",
							Type: "Animal",
						},
					},
					Structs: []Struct{},
					Funcs: []Func{
						{
							Name: "main",
						},
					},
				},
				{
					Name:    "helper",
					Globals: []Global{},
					Structs: []Struct{
						{
							Name: "Animal",
							Fields: []StructField{
								{
									Name: "sound",
									Type: "str",
								},
							},
						},
					},
					Funcs: []Func{
						{
							Name:   "Animal.Speak",
							Inputs: "helper.a *Animal",
						},
					},
				},
				{
					Name: "config",
					Globals: []Global{
						{
							Name: "Name",
							Type: "str",
						},
						{
							Name: "Apple",
							Type: "i32",
						},
					},
					Structs: []Struct{},
					Funcs:   []Func{},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {

			astbuild.AST = nil

			_, sourceCode, _, perr := sourceargs.Parse([]string{tc.testDir}, true)
			if perr != nil {
				t.Fatal(perr)
			}

			err := packageloader.CXProgramLoad("test", sourceCode)
			if err != nil {
				t.Fatal(err)
			}

			files, err := packagefiles.ImportFilesGet("test")
			if err != nil {
				t.Fatal(err)
			}

			Imports, Globals, Enums, TypeDefinitions, Structs, Funcs, gotErr := declarationextractor.AllDeclarationsExtract(files)
			if (Enums != nil && TypeDefinitions != nil) || gotErr != nil {
				t.Fatal(gotErr)
			}

			err = typechecker.AllDeclarationsParse(files, Imports, Globals, Structs, Funcs)
			if err != nil {
				t.Fatal(err)
			}

			program := astbuild.AST

			for _, wantPkg := range tc.wantProgram {
				gotPkg, err := program.PackageGet(program.CXProgramNameIntern(wantPkg.Name))
				if err != nil {
					t.Fatal(err)
				}

				if syspackages.PackageDefault(astbuild.AST.CXProgramNameText(gotPkg.Name)) {
					continue
				}

				// Globals
				for _, wantGlobal := range wantPkg.Globals {
					var gotGlobalName string
					var gotGlobalType string
					var match bool
					for _, globalIdx := range gotPkg.Globals.Fields {
						// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
						// Now an arena getter, site S1554.
						gotGlobal := program.CXTypeSignatureFromArrayGet(globalIdx)
						gotGlobalName = program.CXProgramNameText(gotGlobal.Name)

						if gotGlobalName == wantGlobal.Name {

							gotGlobalType = ast.FormattedTypeGet(program, gotGlobal)

							if gotGlobalType == wantGlobal.Type {
								match = true
							}

							break
						}

					}

					if !match {
						t.Errorf("want global %s %s %s, got %s %s %s", wantPkg.Name, wantGlobal.Name, wantGlobal.Type, astbuild.AST.CXProgramNameText(gotPkg.Name), gotGlobalName, gotGlobalType)
					}
				}

				// Structs
				for _, wantStruct := range wantPkg.Structs {
					var match bool
					var fields string
					// OBJMODEL: state.CXStructMng.CXStructAlloc(), phase 4 #336.
					// Now a local object, site S1555.
					var gotStruct ast.CXStruct
					for _, structIdx := range gotPkg.Structs {
						gotStruct = program.CXStructs[structIdx]

						if astbuild.AST.CXProgramNameText(gotStruct.Name) == wantStruct.Name {

							var fieldMatch int
							for _, wantField := range wantStruct.Fields {

								for _, gotFieldIdx := range gotStruct.Fields {

									// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
									// Now an arena getter, site S1556.
									gotField := program.CXTypeSignatureFromArrayGet(gotFieldIdx)
									var gotFieldType string

									gotFieldType = ast.FormattedTypeGet(program, gotField)

									if program.CXProgramNameText(gotField.Name) == wantField.Name {

										if gotFieldType == wantField.Type {
											fieldMatch++
											break
										}
									}

									fields += fmt.Sprintf("want field %s %s, got %s %s\n", wantField.Name, wantField.Type, program.CXProgramNameText(gotField.Name), gotFieldType)

								}

								if fieldMatch == len(wantStruct.Fields) {
									match = true
								}

							}

							break
						}

					}
					if !match {
						t.Errorf("want %s %s, got %s %s\n%s", wantPkg.Name, wantStruct.Name, astbuild.AST.CXProgramNameText(gotPkg.Name), astbuild.AST.CXProgramNameText(gotStruct.Name), fields)
					}

				}

				// Funcs
				for _, wantFunc := range wantPkg.Funcs {

					var match bool

					var wantFuncName string = wantFunc.Name
					var wantFuncInputs string = wantFunc.Inputs
					var wantFuncOutputs string = wantFunc.Outputs

					var gotFuncName string
					var gotFuncInputs string
					var gotFuncOutputs string
					for _, gotFuncIdx := range gotPkg.Functions {
						gotFunc := program.CxIrState.CXFunctionMng.CXFunctionGet(gotFuncIdx)
						gotFuncName = astbuild.AST.CXProgramNameText(gotFunc.Name)

						var inps bytes.Buffer
						var outs bytes.Buffer

						getFormattedParam(program, gotFunc.InputsGet(program), gotPkg, &inps)
						getFormattedParam(program, gotFunc.OutputsGet(program), gotPkg, &outs)

						gotFuncInputs = inps.String()
						gotFuncOutputs = outs.String()

						if gotFuncName == wantFuncName {
							if gotFuncInputs == wantFuncInputs && gotFuncOutputs == wantFuncOutputs {
								match = true
							}
							break
						}

					}

					if !match {
						t.Errorf("want func %s (%s) (%s), got %s (%s) (%s)", wantFuncName, wantFuncInputs, wantFuncOutputs, gotFuncName, gotFuncInputs, gotFuncOutputs)

					}
				}

			}

		})
	}
}

func getFormattedParam(prgrm *ast.CXProgram, paramTypeSigIdxs []ast.CXTypeSignatureIndex, pkg *ast.CXPackage, buf *bytes.Buffer) {
	for i, paramTypeSigIdx := range paramTypeSigIdxs {
		// OBJMODEL: state.CXTypeSignatureMng.CXTypeSignatureGet(gid), phase 5 #332.
		// Now an arena getter, site S1558.
		paramTypeSig := prgrm.CXTypeSignatureFromArrayGet(paramTypeSigIdx)

		// Checking if this argument comes from an imported package.
		externalPkg := false
		if ast.CXPackageIndex(pkg.Index) != paramTypeSig.Package {
			externalPkg = true
		}

		if paramTypeSig.Type == ast.TypeCXArgumentDeprecate {
			// OBJMODEL: state.CXArgumentMng.CXArgumentGet(gid), phase 5 #327.
			// Now an arena getter, site S1559.
			param := prgrm.CXArgFromArrayGet(ast.CXArgumentIndex(paramTypeSig.Meta))

			buf.WriteString(fmt.Sprintf("%s %s", ast.FormattedNameGet(prgrm, param, externalPkg, pkg), ast.FormattedTypeGet(prgrm, paramTypeSig)))
		} else {
			name := prgrm.CXProgramNameText(paramTypeSig.Name)

			// If it's a literal, just override the name with LiteralPlaceholder.
			if paramTypeSig.Name == stringpool.StringGidNil {
				name = opcodetable.LiteralPlaceholder
			}

			// TODO: Check if external pkg and pkg name shown are correct
			if externalPkg {
				name = fmt.Sprintf("%s.%s", prgrm.CXProgramNameText(pkg.Name), name)
			}

			buf.WriteString(fmt.Sprintf("%s %s", name, ast.FormattedTypeGet(prgrm, paramTypeSig)))
		}

		if i != len(paramTypeSigIdxs)-1 {
			buf.WriteString(", ")
		}

	}
}
