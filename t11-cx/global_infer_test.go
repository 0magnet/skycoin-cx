package main

import (
	"fmt"
	"slices"
	"testing"

	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"github.com/skycoin/skycoin-cx/t4-cx-compiler-astbuild"
	"github.com/skycoin/skycoin-cx/t9-cx-compiler-parsing"
)

// A global that takes its type from a value, var inferred = f(), has the
// global of var declared T, where f returns T.
func TestGlobalInferredMatchesDeclared(t *testing.T) {
	for _, typ := range []string{"i32", "f64", "bool", "str", "*i32", "[3]i32", "[2][3]f32", "[]i32", "[][]i32", "*[2]i32", "Point", "*Point", "[]Point", "[2]Point", "[]str"} {
		t.Run(typ, func(t *testing.T) {
			initProgram()
			src := fmt.Sprintf("package main\ntype Point struct {\n\tx i32\n\ty f64\n}\nfunc value() (r %s) {\n\treturn r\n}\nvar inferred = value()\nvar declared %s\nfunc main() {\n}\n", typ, typ)
			if err := parsing.SourceStringsParse([]string{src}, []string{"infer.cx"}); err != nil {
				t.Fatal(err)
			}
			pkg, err := astbuild.AST.PackageGet(astbuild.AST.CxIrState.CxIrStateNameMain)
			if err != nil {
				t.Fatal(err)
			}
			inferred, err := pkg.GlobalGet(astbuild.AST, astbuild.AST.CXProgramNameIntern("inferred"))
			if err != nil {
				t.Fatal(err)
			}
			declared, err := pkg.GlobalGet(astbuild.AST, astbuild.AST.CXProgramNameIntern("declared"))
			if err != nil {
				t.Fatal(err)
			}
			if got, want := globalTypeDescription(astbuild.AST, inferred), globalTypeDescription(astbuild.AST, declared); got != want {
				t.Errorf("inferred global\n%s\nwant the declared global\n%s", got, want)
			}
		})
	}
}

// globalTypeDescription prints what the type of the global glbl is made of.
func globalTypeDescription(prgrm *ast.CXProgram, glbl *ast.CXTypeSignature) string {
	text := fmt.Sprintf("signature type %d pass by %d deref %v type %s size %d", glbl.Type, glbl.PassBy, glbl.Deref, ast.FormattedTypeGet(prgrm, glbl), glbl.SizeGet(prgrm, false))
	switch glbl.Type {
	case ast.TypeAtomic, ast.TypePointerAtomic:
		text += fmt.Sprintf(" meta %d", glbl.Meta)
	case ast.TypeArrayAtomic, ast.TypePointerArrayAtomic, ast.TypeSliceAtomic, ast.TypePointerSliceAtomic:
		array := prgrm.CXTypeSignatureArrayFromArrayGet(glbl.Meta)
		text += fmt.Sprintf(" element %d lengths %v indexes %d", array.Meta, array.Lengths, len(array.Indexes))
	case ast.TypeCXArgumentDeprecate:
		arg := prgrm.CXArgGet(ast.CXArgumentIndex(glbl.Meta))
		structName := ""
		if arg.StructType != nil {
			structName = prgrm.CXProgramNameText(arg.StructType.Name)
		}
		text += fmt.Sprintf(" argument type %d target %d struct %q size %d lengths %v specifiers %v pass by %d", arg.Type, arg.PointerTargetType, structName, arg.Size, arg.Lengths, slices.Clone(arg.DeclarationSpecifiers), arg.PassBy)
	}
	return text
}
