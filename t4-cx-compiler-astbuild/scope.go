package astbuild

import (
	"math"

	assert "github.com/skycoin/skycoin-cx/t0-cx-core-assert"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
)

// DefineNewScope marks the first and last non CXOperationTypeLine expressions of a block
// as the start and end of a scope. Nested blocks may share an expression.
func DefineNewScope(exprs []ast.CXExpression) {
	first, last := -1, -1
	for i := range exprs {
		if exprs[i].Type != ast.CXOperationTypeLine {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	if first < 0 {
		return
	}
	assert.InvariantTrue(exprs[first].ExpressionScopeOpenCount < math.MaxUint8, "DefineNewScope: too many scopes open at one expression")
	assert.InvariantTrue(exprs[last].ExpressionScopeCloseCount < math.MaxUint8, "DefineNewScope: too many scopes close at one expression")
	exprs[first].ExpressionScopeOpenCount++
	exprs[last].ExpressionScopeCloseCount++
}
