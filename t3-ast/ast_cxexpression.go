package ast

import (
	"github.com/skycoin/skycoin-cx/t1-constants"
	stringpool "github.com/skycoin/skycoin-cx/t2-cx-core-stringpool"
)

/*
 * CXExpressionType enum contains CX expressions types for CXExpression struct
 */
type CXExpressionType int

const (
	CXExpressionTypeUnused CXExpressionType = iota
	CXExpressionTypeMethodCall
	CXExpressionTypeStructLiteral
	CXExpressionTypeArrayLiteral
	// CXExpressionTypeFunctionCall is a resolved call, which cannot be called again.
	CXExpressionTypeFunctionCall
	// CXExpressionTypeStructLiteralField writes one field of a struct literal that is not
	// yet given its variable.
	CXExpressionTypeStructLiteralField
	// CXExpressionTypeShortDeclarationShared declares a name of a := with more than
	// one name, which assigns a variable of the same scope instead.
	CXExpressionTypeShortDeclarationShared
	// CXExpressionTypeShortDeclarationSharedLast is the last of them.
	CXExpressionTypeShortDeclarationSharedLast
)

// String returns alias for constants defined for cx edpression type
func (cxet CXExpressionType) String() string {
	return [...]string{"Unused", "MethodCall", "StructLiteral", "ArrayLiteral", "FunctionCall", "StructLiteralField", "ShortDeclarationShared", "ShortDeclarationSharedLast"}[int(cxet)]
}

// TypeMethodCall checks if expression type is method call
func (cxe CXExpression) TypeMethodCall() bool {
	return cxe.ExpressionType == CXExpressionTypeMethodCall
}

// TypeStructLiteral checks if expression type is struct literal
func (cxe CXExpression) TypeStructLiteral() bool {
	return cxe.ExpressionType == CXExpressionTypeStructLiteral
}

// TypeArrayLiteral checks if expression type is array literal
func (cxe CXExpression) TypeArrayLiteral() bool {
	return cxe.ExpressionType == CXExpressionTypeArrayLiteral
}

// BreakOperator checks if expression type is break
func (cxe CXExpression) BreakOperator(prgrm *CXProgram) bool {
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S0120.
	expression, err := prgrm.CXAtomicOpGet(cxe.Index)
	if err != nil {
		panic(err)
	}

	expressionOperator := prgrm.CxIrState.CxIrStateFunctionGetOrNil(expression.Operator)
	return expressionOperator != nil && expressionOperator.AtomicOPCode == constants.OpBreak
}

// ContinueOperator checks if expression type is continue
func (cxe CXExpression) ContinueOperator(prgrm *CXProgram) bool {
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S0122.
	expression, err := prgrm.CXAtomicOpGet(cxe.Index)
	if err != nil {
		panic(err)
	}

	expressionOperator := prgrm.CxIrState.CxIrStateFunctionGetOrNil(expression.Operator)
	return expressionOperator != nil && expressionOperator.AtomicOPCode == constants.OpContinue
}

// UndTyped checks if expression type is und type
func (cxe CXExpression) UndTyped(prgrm *CXProgram) bool {
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S0124.
	expression, err := prgrm.CXAtomicOpGet(cxe.Index)
	if err != nil {
		panic(err)
	}

	expressionOperator := prgrm.CxIrState.CxIrStateFunctionGetOrNil(expression.Operator)
	return expressionOperator != nil && OpcodeOperator(expressionOperator.AtomicOPCode)

}

// ----------------------------------------------------------------
//                             `CXExpression` Getters

func (cxe CXExpression) LabelGet(prgrm *CXProgram) stringpool.StringGid {
	// OBJMODEL: state.CXAtomicOperatorMng.CXAtomicOperatorGet(gid), phase 5 #342.
	// Now an arena getter, site S0126.
	expression, err := prgrm.CXAtomicOpGet(cxe.Index)
	if err != nil {
		panic(err)
	}

	return expression.Label
}
