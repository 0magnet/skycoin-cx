package ast

/*
 * CXOperationType enum contains CX operations types for CXOperation struct
 */
type CXOperationType int

const (
	UNUSED CXOperationType = iota
	CXOperationTypeAtomicOperator
	CXOperationTypeArgument
	CXOperationTypeLine
)
