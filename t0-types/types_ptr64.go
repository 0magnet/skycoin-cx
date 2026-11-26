//go:build ptr64

package types

type Pointer uint64

const PointerSize = UI64Size
const InvalidPointer = Pointer(MaxUint64)

func PtrToIntCast(value Pointer) int {
	panicIf(uint64(value) > uint64(MaxInt), CXRuntimeInvalidCast)
	return int(value)
}

func PtrToI32Cast(value Pointer) int32 {
	panicIf(uint64(value) > uint64(MaxInt32), CXRuntimeInvalidCast)
	return int32(value)
}

func PtrToUI32Cast(value Pointer) uint32 {
	panicIf(uint64(value) > uint64(MaxUint32), CXRuntimeInvalidCast)
	return uint32(value)
}

func PtrToUI64Cast(value Pointer) uint64 {
	return uint64(value)
}

func IntToPtrCast(value int) Pointer {
	panicIf(value < 0 || uint64(value) > uint64(InvalidPointer), CXRuntimeInvalidCast)
	return Pointer(value)
}

func I32ToPtrCast(value int32) Pointer {
	panicIf(value < 0, CXRuntimeInvalidCast)
	return Pointer(value)
}

func I64ToPtrCast(value int64) Pointer {
	panicIf(value < 0, CXRuntimeInvalidCast)
	return Pointer(value)
}

func UI64ToPtrCast(value uint64) Pointer {
	return Pointer(value)
}

func F32ToPtrCast(value float32) Pointer {
	panicIf(value < 0 || uint64(value) > uint64(InvalidPointer), CXRuntimeInvalidCast)
	return Pointer(value)
}

func F64ToPtrCast(value float64) Pointer {
	panicIf(value < 0 || uint64(value) > uint64(InvalidPointer), CXRuntimeInvalidCast)
	return Pointer(value)
}

func PtrRead(memory []byte, offset Pointer) Pointer {
	return Pointer(UI64Read(memory, offset))
}

func PtrWrite(memory []byte, offset Pointer, value Pointer) {
	UI64Write(memory, offset, PtrToUI64Cast(value))
}
