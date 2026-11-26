package astbuild

import (
	"testing"

	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
)

func TestConstantsHaveTheirOwnCodes(t *testing.T) {
	ConstantsInit()
	names := map[int]string{}
	for _, entry := range ConstNames.NativeNameIndexAll() {
		name, code := entry.NativeNameEntryText, entry.NativeNameEntryCode
		if other, ok := names[code]; ok {
			t.Errorf("%s and %s share code %d", name, other, code)
		}
		names[code] = name
	}
	want := map[string]int32{
		"cx.SUCCESS":                  valuetypes.CXSuccess,
		"cx.ASSERT":                   valuetypes.CXAssert,
		"cx.RUNTIME_INVALID_ARGUMENT": valuetypes.CXRuntimeInvalidArgument,
		"cx.RUNTIME_NOT_IMPLEMENTED":  valuetypes.CXRuntimeNotImplemented,
		"cx.POINTER_SIZE":             valuetypes.PtrToI32Cast(valuetypes.PointerSize),
	}
	for name, value := range want {
		entry, ok := ConstNames.NativeNameIndexText(name)
		code := entry.NativeNameEntryCode
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if got := valuetypes.I32Read(Constants[code].Value, 0); got != value {
			t.Errorf("%s = %d, want %d", name, got, value)
		}
	}
}
