package typechecker

import "github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"

// Types Map
var TypesMap map[string]valuetypes.Code = map[string]valuetypes.Code{

	"bool": valuetypes.BOOL,

	"i8":  valuetypes.I8,
	"i16": valuetypes.I16,
	"i32": valuetypes.I32,
	"i64": valuetypes.I64,

	"ui8":  valuetypes.UI8,
	"ui16": valuetypes.UI16,
	"ui32": valuetypes.UI32,
	"ui64": valuetypes.UI64,

	"f32": valuetypes.F32,
	"f64": valuetypes.F64,

	"str": valuetypes.STR,
	"aff": valuetypes.AFF,
}
