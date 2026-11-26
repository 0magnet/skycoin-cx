package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t1-cx-core-opcodetable"
)

// llvmFile is the generated opcode table of the LLVM target, decision D10.
const llvmFile = "t3-cx-compiler-llvm/llvm_opcodes_gen.go"

// llvmScalarIdents spells each scalar type as the ir.Opcode names do.
var llvmScalarIdents = map[valuetypes.Code]string{
	valuetypes.BOOL: "Bool", valuetypes.I8: "I8", valuetypes.I16: "I16", valuetypes.I32: "I32", valuetypes.I64: "I64",
	valuetypes.UI8: "UI8", valuetypes.UI16: "UI16", valuetypes.UI32: "UI32", valuetypes.UI64: "UI64", valuetypes.F32: "F32", valuetypes.F64: "F64",
}

// llvmOperationIdents spells each operation of the scalar opcode families of
// the image as the ir.Opcode names do. A row whose operation is not here and
// is not a conversion has no image opcode.
var llvmOperationIdents = map[string]string{
	"add": "Add", "sub": "Sub", "mul": "Mul", "div": "Div", "mod": "Mod", "neg": "Neg", "abs": "Abs",
	"max": "Max", "min": "Min", "eq": "Eq", "uneq": "Uneq", "lt": "Lt", "lteq": "Lteq", "gt": "Gt",
	"gteq": "Gteq", "bitand": "BitAnd", "bitor": "BitOr", "bitxor": "BitXor", "bitclear": "BitClear",
	"bitshl": "BitShl", "bitshr": "BitShr", "print": "Print", "not": "Not", "and": "And", "or": "Or",
}

// llvmRow is one scalar image opcode and the definition row it lowers from.
type llvmRow struct {
	ident     string
	row       opcodetable.OpcodeDefinition
	operation string
	operand   valuetypes.Code
	result    valuetypes.Code
	// conversion is a row named from.to.
	conversion bool
}

// llvmRows finds the definition rows named type.operation or type.type
// whose image opcode the LLVM target compiles, in table order.
func llvmRows() ([]llvmRow, error) {
	codes := typeCodes()
	var rows []llvmRow
	for _, row := range opcodetable.OpcodeDefinitionTable {
		typeName, operation, found := strings.Cut(row.Name, ".")
		operand := codes[typeName]
		typeIdent, scalar := llvmScalarIdents[operand]
		if !found || !scalar {
			continue
		}
		ident, conversion := "", false
		if operationIdent, known := llvmOperationIdents[operation]; known {
			ident = "Opcode" + typeIdent + operationIdent
		} else if target, scalar := llvmScalarIdents[codes[operation]]; scalar && operand != valuetypes.BOOL && codes[operation] != valuetypes.BOOL && codes[operation] != operand {
			ident = "Opcode" + typeIdent + "To" + target
			conversion = true
		} else {
			continue
		}
		inputs := strings.Fields(row.Inputs)
		outputs := strings.Fields(row.Outputs)
		if len(inputs) < 1 || len(inputs) > 2 || len(outputs) > 1 || inputs[0] != typeName {
			return nil, fmt.Errorf("definition %q: a scalar image opcode takes one or two %s inputs and gives at most one output", row.Name, typeName)
		}
		result := valuetypes.UNDEFINED
		if len(outputs) == 1 {
			result = codes[outputs[0]]
			if _, scalar := llvmScalarIdents[result]; !scalar {
				return nil, fmt.Errorf("definition %q: output %q is not a scalar type", row.Name, outputs[0])
			}
		}
		rows = append(rows, llvmRow{ident: ident, row: row, operation: operation, operand: operand, result: result, conversion: conversion})
	}
	return rows, nil
}

// llvmFileSource writes the opcode table of the LLVM target.
func llvmFileSource(rows []llvmRow) []byte {
	var b bytes.Buffer
	b.WriteString(header)
	b.WriteString("package llvm\n\nimport (\n\t\"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes\"\n\t\"github.com/skycoin/skycoin-cx/t2-cx-runtime-ir\"\n)\n\n")
	b.WriteString("// opcodeFacts describes each scalar image opcode by the row of\n// OpcodeDefinitionTable it lowers from. Read only.\n")
	b.WriteString("var opcodeFacts = [ir.OpcodeCount]OpcodeFact{\n")
	for _, row := range rows {
		fmt.Fprintf(&b, "\tir.%s: {%q, %q, %s, %d, %s, %q},\n", row.ident, row.row.Name, row.operation, typeConstant(row.operand),
			len(strings.Fields(row.row.Inputs)), typeConstant(row.result), llvmTemplate(row))
	}
	b.WriteString("}\n")
	return b.Bytes()
}
