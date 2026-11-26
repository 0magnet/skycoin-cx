package types

//"fmt"

func StrPayloadWrite(memory []byte, offset Pointer, value string) {
	bytes := []byte(value)
	size := ObjectSizeCompute(bytes)
	ObjectPayloadWrite(memory, offset, size, bytes)
}

// VIOLATION STRHOT-011: B, only the boundary opcodes use Go strings, #412.
// VIOLATION VMALLOC-003: a Go string copy of a str, for the boundary opcodes only, #407.
func StrRead(memory []byte, offset Pointer) string {
	str := string(ObjectRead(memory, offset))
	return str
}

// StrBytesRead returns the bytes of the str whose pointer is at offset, a
// view into memory without a copy. The empty str is nil.
func StrBytesRead(memory []byte, offset Pointer) []byte {
	return ObjectRead(memory, offset)
}

func StrSizeRead(memory []byte, offset Pointer) Pointer {
	heapOffset := PtrRead(memory, offset)
	if heapOffset > 0 && heapOffset.Valid() {
		return ObjectSizeRead(memory, heapOffset) - ObjectHeaderSize
	}
	return 0
}
