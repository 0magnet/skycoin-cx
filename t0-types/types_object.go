package types

//"fmt"

// TODO: rewrite api
// TODO: cleanup object size mess: object size == object data size + ObjectHeaderSize.

const MarkSize = Pointer(1)
const ForwardingAddressSize = PointerSize
const ObjectGCHeaderSize = MarkSize + ForwardingAddressSize

const ObjectSize = PointerSize
const ObjectHeaderSize = ObjectGCHeaderSize + ObjectSize

func ObjectHeaderGet(memory []byte, offset Pointer) []byte {
	return memory[offset : offset+ObjectHeaderSize]
}

func ObjectPayloadGet(memory []byte, offset Pointer, size Pointer) []byte {
	offset += ObjectHeaderSize
	return memory[offset : offset+size]
}

func ObjectSizeCompute(payload []byte) Pointer {
	return IntToPtrCast(len(payload)) + ObjectHeaderSize
}

func ObjectForwardingAddressRead(memory []byte, offset Pointer) Pointer {
	return PtrRead(memory, offset+MarkSize)
}

func ObjectSizeRead(memory []byte, offset Pointer) Pointer {
	return PtrRead(memory, offset+ObjectHeaderSize-ObjectSize)
}

func ObjectSizeWrite(memory []byte, offset Pointer, size Pointer) {
	PtrWrite(memory, offset+ObjectHeaderSize-ObjectSize, size)
}

func ObjectMarkWrite(memory []byte, offset Pointer, mark byte) {
	UI8Write(memory, offset, mark)
}

func ObjectForwardingAddressWrite(memory []byte, offset Pointer, address Pointer) {
	PtrWrite(memory, offset+MarkSize, address)
}

func ObjectMake(payload []byte) []byte {
	size := ObjectSizeCompute(payload)
	obj := make([]byte, size)
	//fmt.Printf("MAKE_OBJ SIZE %d, LEN_DATA %d, LEN_OBJ %d\n", size, len(data), len(obj))
	ObjectPayloadWrite(obj, 0, size, payload)
	return obj
}

func ObjectPayloadWrite(memory []byte, offset Pointer, size Pointer, obj []byte) {
	//size := IntToPtrCast(len(obj))
	ObjectSizeWrite(memory, offset, size)
	SliceByteWrite(memory, offset+ObjectHeaderSize, obj)
}

func ObjectPayloadRead(memory []byte, offset Pointer) []byte {
	size := ObjectSizeRead(memory, offset)
	size -= ObjectHeaderSize
	obj := SliceByteGet(memory, offset+ObjectHeaderSize, size)
	return obj
}

func ObjectRead(memory []byte, offset Pointer) []byte {
	heapOffset := PtrRead(memory, offset)
	if heapOffset != 0 && heapOffset.Valid() {
		obj := ObjectPayloadRead(memory, heapOffset)
		return obj
	}
	return nil
}
