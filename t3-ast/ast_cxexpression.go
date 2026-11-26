package ast

import (
	"github.com/skycoin/skycoin-cx/t1-constants"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
)

/*
 * CXEXPR_TYPE enum contains CX expressions types for CXExpression struct
 */
type CXEXPR_TYPE int

const (
	CXEXPR_UNUSED CXEXPR_TYPE = iota
	CXEXPR_METHOD_CALL
	CXEXPR_STRUCT_LITERAL
	CXEXPR_ARRAY_LITERAL
	// CXEXPR_FUNCTION_CALL is a resolved call, which cannot be called again.
	CXEXPR_FUNCTION_CALL
	// CXEXPR_STRUCT_LITERAL_FIELD writes one field of a struct literal that is not
	// yet given its variable.
	CXEXPR_STRUCT_LITERAL_FIELD
	// CXEXPR_SHORT_DECLARATION_SHARED declares a name of a := with more than
	// one name, which assigns a variable of the same scope instead.
	CXEXPR_SHORT_DECLARATION_SHARED
	// CXEXPR_SHORT_DECLARATION_SHARED_LAST is the last of them.
	CXEXPR_SHORT_DECLARATION_SHARED_LAST
)

// String returns alias for constants defined for cx edpression type
func (cxet CXEXPR_TYPE) String() string {
	return [...]string{"Unused", "MethodCall", "StructLiteral", "ArrayLiteral", "FunctionCall", "StructLiteralField", "ShortDeclarationShared", "ShortDeclarationSharedLast"}[int(cxet)]
}

// TypeMethodCall checks if expression type is method call
func (cxe CXExpression) TypeMethodCall() bool {
	return cxe.ExpressionType == CXEXPR_METHOD_CALL
}

// TypeStructLiteral checks if expression type is struct literal
func (cxe CXExpression) TypeStructLiteral() bool {
	return cxe.ExpressionType == CXEXPR_STRUCT_LITERAL
}

// TypeArrayLiteral checks if expression type is array literal
func (cxe CXExpression) TypeArrayLiteral() bool {
	return cxe.ExpressionType == CXEXPR_ARRAY_LITERAL
}

// BreakOperator checks if expression type is break
func (cxe CXExpression) BreakOperator(prgrm *CXProgram) bool {
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S0120.
	expression, err := prgrm.GetCXAtomicOp(cxe.Index)
	if err != nil {
		panic(err)
	}

	expressionOperator := prgrm.CxIrState.CxIrStateFunctionGetOrNil(expression.Operator)
	return expressionOperator != nil && expressionOperator.AtomicOPCode == constants.OP_BREAK
}

// ContinueOperator checks if expression type is continue
func (cxe CXExpression) ContinueOperator(prgrm *CXProgram) bool {
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S0122.
	expression, err := prgrm.GetCXAtomicOp(cxe.Index)
	if err != nil {
		panic(err)
	}

	expressionOperator := prgrm.CxIrState.CxIrStateFunctionGetOrNil(expression.Operator)
	return expressionOperator != nil && expressionOperator.AtomicOPCode == constants.OP_CONTINUE
}

// UndTyped checks if expression type is und type
func (cxe CXExpression) UndTyped(prgrm *CXProgram) bool {
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S0124.
	expression, err := prgrm.GetCXAtomicOp(cxe.Index)
	if err != nil {
		panic(err)
	}

	expressionOperator := prgrm.CxIrState.CxIrStateFunctionGetOrNil(expression.Operator)
	return expressionOperator != nil && OpcodeOperator(expressionOperator.AtomicOPCode)

}

// ----------------------------------------------------------------
//                             `CXExpression` Getters

func (cxe CXExpression) GetLabel(prgrm *CXProgram) stringpool.StringGid {
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S0126.
	expression, err := prgrm.GetCXAtomicOp(cxe.Index)
	if err != nil {
		panic(err)
	}

	return expression.Label
}
