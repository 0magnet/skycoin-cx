package types

//"fmt"

// TODO: rewrite api
// TODO: cleanup object size mess: object size == object data size + OBJECT_HEADER_SIZE.

const MARK_SIZE = Pointer(1)
const FORWARDING_ADDRESS_SIZE = POINTER_SIZE
const OBJECT_GC_HEADER_SIZE = MARK_SIZE + FORWARDING_ADDRESS_SIZE

const OBJECT_SIZE = POINTER_SIZE
const OBJECT_HEADER_SIZE = OBJECT_GC_HEADER_SIZE + OBJECT_SIZE

func Get_obj_header(memory []byte, offset Pointer) []byte {
	return memory[offset : offset+OBJECT_HEADER_SIZE]
}

func ObjectPayloadGet(memory []byte, offset Pointer, size Pointer) []byte {
	offset += OBJECT_HEADER_SIZE
	return memory[offset : offset+size]
}

func Compute_obj_size(payload []byte) Pointer {
	return Cast_int_to_ptr(len(payload)) + OBJECT_HEADER_SIZE
}

func Read_obj_forwarding_address(memory []byte, offset Pointer) Pointer {
	return Read_ptr(memory, offset+MARK_SIZE)
}

func Read_obj_size(memory []byte, offset Pointer) Pointer {
	return Read_ptr(memory, offset+OBJECT_HEADER_SIZE-OBJECT_SIZE)
}

func Write_obj_size(memory []byte, offset Pointer, size Pointer) {
	Write_ptr(memory, offset+OBJECT_HEADER_SIZE-OBJECT_SIZE, size)
}

func Write_obj_mark(memory []byte, offset Pointer, mark byte) {
	Write_ui8(memory, offset, mark)
}

func Write_obj_forwarding_address(memory []byte, offset Pointer, address Pointer) {
	Write_ptr(memory, offset+MARK_SIZE, address)
}

func Make_obj(payload []byte) []byte {
	size := Compute_obj_size(payload)
	obj := make([]byte, size)
	//fmt.Printf("MAKE_OBJ SIZE %d, LEN_DATA %d, LEN_OBJ %d\n", size, len(data), len(obj))
	ObjectPayloadWrite(obj, 0, size, payload)
	return obj
}

func ObjectPayloadWrite(memory []byte, offset Pointer, size Pointer, obj []byte) {
	//size := Cast_int_to_ptr(len(obj))
	Write_obj_size(memory, offset, size)
	WriteSlice_byte(memory, offset+OBJECT_HEADER_SIZE, obj)
}

func ObjectPayloadRead(memory []byte, offset Pointer) []byte {
	size := Read_obj_size(memory, offset)
	size -= OBJECT_HEADER_SIZE
	obj := GetSlice_byte(memory, offset+OBJECT_HEADER_SIZE, size)
	return obj
}

func Read_obj(memory []byte, offset Pointer) []byte {
	heapOffset := Read_ptr(memory, offset)
	if heapOffset != 0 && heapOffset.Valid() {
		obj := ObjectPayloadRead(memory, heapOffset)
		return obj
	}
	return nil
}
