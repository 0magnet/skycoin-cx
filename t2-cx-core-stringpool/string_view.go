package stringpool

import "unsafe"

// StringView returns text as a Go string that shares its bytes, for
// a Go API that takes a string and does not keep it. Only t6-cx-library-graphics calls it, see #432.
// VIOLATION STRHOT-322: B, a string view of program bytes for the GLFW title and gltext calls, no copy, #412.
func StringView(text []byte) string {
	return unsafe.String(unsafe.SliceData(text), len(text))
}
