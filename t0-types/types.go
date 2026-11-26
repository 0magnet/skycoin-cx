package types

import (
	"math"
	//"fmt"
)

const (
	CXSuccess = iota //zero can be success
	CXCompilationError
	CXPanic // 2
	CXInternalError
	CXAssert
	CXRuntimeError
	CXRuntimeStackOverflowError
	CXRuntimeHeapExhaustedError
	CXRuntimeInvalidArgument
	CXRuntimeSliceIndexOutOfRange
	CXRuntimeNotImplemented
	CXRuntimeInvalidCast
	CXRuntimeLibraryDisabled
)

const (
	MaxInt   = int(MaxUint >> 1)
	MaxInt32 = int(MaxUint32 >> 1)
	MinInt32 = -MaxInt32 - 1

	MaxUint   = ^uint(0)
	MaxUint8  = ^uint8(0)
	MaxUint16 = ^uint16(0)
	MaxUint32 = ^uint32(0)
	MaxUint64 = ^uint64(0)
)

const (
	BoolSize = Pointer(1)

	I8Size  = Pointer(1)
	I16Size = Pointer(2)
	I32Size = Pointer(4)
	I64Size = Pointer(8)

	UI8Size  = Pointer(1)
	UI16Size = Pointer(2)
	UI32Size = Pointer(4)
	UI64Size = Pointer(8)

	F32Size = Pointer(4)
	F64Size = Pointer(8)

	ArraySize = PointerSize
	SliceSize = PointerSize

	StrSize = PointerSize

	// AffSize is an aff value, four i32 the compiler resolves from an aff
	// literal, see docs/IR.md.
	AffSize = Pointer(16)
)

type Code int

const (
	UNUSED Code = iota //reserve zero value, if this value appears, program should crash; should be assert

	BOOL

	I8
	I16
	I32
	I64

	UI8
	UI16
	UI32
	UI64

	F32
	F64

	POINTER
	STR

	ARRAY
	SLICE
	STRUCT

	FUNC
	AFF

	UNDEFINED
	IDENTIFIER
	// INVALID types a placeholder for a value whose compile error is recorded.
	INVALID

	COUNT
)

func (t Code) Name() string {
	return definitions[t].name
}

// NameBytes returns the bytes of Name without building a string, for
// printing on the run path (#412). The caller must not write to them.
func (t Code) NameBytes() []byte {
	return definitionNames[t]
}

func (t Code) Size() Pointer {
	//	panicIf(!definitions[t].size.Valid(), CXRuntimeInvalidArgument)
	return definitions[t].size
}

func (t Code) Primitive() bool {
	return definitions[t].primitive
}

type Type struct {
	name      string
	size      Pointer
	primitive bool
}

var definitions []Type = []Type{
	{"UNUSED", InvalidPointer, false},

	{"bool", BoolSize, true},

	{"i8", I8Size, true},
	{"i16", I16Size, true},
	{"i32", I32Size, true},
	{"i64", I64Size, true},

	{"ui8", UI8Size, true},
	{"ui16", UI16Size, true},
	{"ui32", UI32Size, true},
	{"ui64", UI64Size, true},

	{"f32", F32Size, true},
	{"f64", F64Size, true},

	{"ptr", PointerSize, false},
	{"str", StrSize, false},
	{"array", ArraySize, false},
	{"slice", SliceSize, false},
	{"struct", InvalidPointer, false},
	{"func", InvalidPointer, false},
	{"aff", AffSize, false},
	{"und", InvalidPointer, false},
	{"ident", PointerSize, false}, // TODO:PTR use InvalidPointer to track addressing issues.
	{"invalid type", 0, false},
}

func panicIf(condition bool, error int) {
	if condition {
		panic(error)
	}
}

func (pointer Pointer) Valid() bool {
	return pointer != InvalidPointer
}

func (pointer *Pointer) Add(value Pointer) bool {
	if pointer.Valid() && value.Valid() {
		*pointer += value
		return true
	}

	*pointer = InvalidPointer
	return false
}

func SintToSptrCast(value []int) []Pointer {
	l := len(value)
	if l == 0 {
		return nil
	}

	sptr := make([]Pointer, l)
	for i, k := range value {
		sptr[i] = IntToPtrCast(k)
	}
	return sptr
}

func BoolRead(memory []byte, offset Pointer) bool {
	memory = memory[offset:]
	return memory[0] != 0
}

func I8Read(memory []byte, offset Pointer) int8 {
	memory = memory[offset:]
	return int8(memory[0])
}

func I16Read(memory []byte, offset Pointer) int16 {
	memory = memory[offset:]
	return int16(memory[0]) | int16(memory[1])<<8
}

func I32Read(memory []byte, offset Pointer) int32 {
	memory = memory[offset:]
	return int32(memory[0]) | int32(memory[1])<<8 | int32(memory[2])<<16 | int32(memory[3])<<24
}

func I64Read(memory []byte, offset Pointer) int64 {
	memory = memory[offset:]
	return int64(memory[0]) | int64(memory[1])<<8 | int64(memory[2])<<16 | int64(memory[3])<<24 |
		int64(memory[4])<<32 | int64(memory[5])<<40 | int64(memory[6])<<48 | int64(memory[7])<<56
}

func UI8Read(memory []byte, offset Pointer) uint8 {
	memory = memory[offset:]
	return uint8(memory[0])
}

func UI16Read(memory []byte, offset Pointer) uint16 {
	memory = memory[offset:]
	return uint16(memory[0]) | uint16(memory[1])<<8
}

func UI32Read(memory []byte, offset Pointer) uint32 {
	memory = memory[offset:]
	return uint32(memory[0]) | uint32(memory[1])<<8 | uint32(memory[2])<<16 | uint32(memory[3])<<24
}

func UI64Read(memory []byte, offset Pointer) uint64 {
	memory = memory[offset:]
	return uint64(memory[0]) | uint64(memory[1])<<8 | uint64(memory[2])<<16 | uint64(memory[3])<<24 |
		uint64(memory[4])<<32 | uint64(memory[5])<<40 | uint64(memory[6])<<48 | uint64(memory[7])<<56
}

func F32Read(memory []byte, offset Pointer) float32 {
	return math.Float32frombits(UI32Read(memory, offset))
}

func F64Read(memory []byte, offset Pointer) float64 {
	return math.Float64frombits(UI64Read(memory, offset))
}

// VIOLATION VMALLOC-012: copies a slice into a new Go slice, for the boundary opcodes only, #407.
func SliceI8Read(memory []byte, offset Pointer) (out []int8) {
	count := IntToPtrCast(len(memory))
	if count > 0 {
		out = make([]int8, count)
		for i := Pointer(0); i < count; i++ {
			out[i] = I8Read(memory, i)
		}
	}
	return
}

// VIOLATION VMALLOC-012: copies a slice into a new Go slice, for the boundary opcodes only, #407.
func SliceI16Read(memory []byte, offset Pointer) (out []int16) {
	count := IntToPtrCast(len(memory) / 2)
	if count > 0 {
		out = make([]int16, count)
		for i := Pointer(0); i < count; i++ {
			out[i] = I16Read(memory, i*2)
		}
	}
	return
}

// VIOLATION VMALLOC-012: copies a slice into a new Go slice, for the boundary opcodes only, #407.
func SliceI32Read(memory []byte, offset Pointer) (out []int32) {
	count := IntToPtrCast(len(memory) / 4)
	if count > 0 {
		out = make([]int32, count)
		for i := Pointer(0); i < count; i++ {
			out[i] = I32Read(memory, i*4)
		}
	}
	return
}

// VIOLATION VMALLOC-012: copies a slice into a new Go slice, for the boundary opcodes only, #407.
func SliceI64Read(memory []byte, offset Pointer) (out []int64) {
	count := IntToPtrCast(len(memory) / 8)
	if count > 0 {
		out = make([]int64, count)
		for i := Pointer(0); i < count; i++ {
			out[i] = I64Read(memory, i*8)
		}
	}
	return
}

// VIOLATION VMALLOC-012: copies a slice into a new Go slice, for the boundary opcodes only, #407.
func SliceUI8Read(memory []byte, offset Pointer) (out []uint8) {
	count := IntToPtrCast(len(memory))
	if count > 0 {
		out = make([]uint8, count)
		for i := Pointer(0); i < count; i++ {
			out[i] = UI8Read(memory, i)
		}
	}
	return
}

// VIOLATION VMALLOC-012: copies a slice into a new Go slice, for the boundary opcodes only, #407.
func SliceUI16Read(memory []byte, offset Pointer) (out []uint16) {
	count := IntToPtrCast(len(memory) / 2)
	if count > 0 {
		out = make([]uint16, count)
		for i := Pointer(0); i < count; i++ {
			out[i] = UI16Read(memory, i*2)
		}
	}
	return
}

// VIOLATION VMALLOC-012: copies a slice into a new Go slice, for the boundary opcodes only, #407.
func SliceUI32Read(memory []byte, offset Pointer) (out []uint32) {
	count := IntToPtrCast(len(memory) / 4)
	if count > 0 {
		out = make([]uint32, count)
		for i := Pointer(0); i < count; i++ {
			out[i] = UI32Read(memory, i*4)
		}
	}
	return
}

func SliceUI64Read(memory []byte, offset Pointer) (out []uint64) {
	count := IntToPtrCast(len(memory) / 8)
	if count > 0 {
		out = make([]uint64, count)
		for i := Pointer(0); i < count; i++ {
			out[i] = UI64Read(memory, i*8)
		}
	}
	return
}

// VIOLATION VMALLOC-012: copies a slice into a new Go slice, for the boundary opcodes only, #407.
func SliceF32Read(memory []byte, offset Pointer) (out []float32) {
	count := IntToPtrCast(len(memory) / 4)
	if count > 0 {
		out = make([]float32, count)
		for i := Pointer(0); i < count; i++ {
			out[i] = F32Read(memory, i*4)
		}
	}
	return
}

// VIOLATION VMALLOC-012: copies a slice into a new Go slice, for the boundary opcodes only, #407.
func SliceF64Read(memory []byte, offset Pointer) (out []float64) {
	count := IntToPtrCast(len(memory) / 8)
	if count > 0 {
		out = make([]float64, count)
		for i := Pointer(0); i < count; i++ {
			out[i] = F64Read(memory, i*8)
		}
	}
	return
}

func SliceByteGet(memory []byte, offset Pointer, size Pointer) []byte {
	return memory[offset : offset+size]
}

func BoolWrite(memory []byte, offset Pointer, value bool) {
	if value {
		memory[offset] = 1
	} else {
		memory[offset] = 0
	}
}

func I8Write(mem []byte, offset Pointer, v int8) {
	mem[offset] = byte(v)
}

func I16Write(mem []byte, offset Pointer, v int16) {
	mem[offset] = byte(v)
	mem[offset+1] = byte(v >> 8)
}

func I32Write(mem []byte, offset Pointer, v int32) {
	mem[offset] = byte(v)
	mem[offset+1] = byte(v >> 8)
	mem[offset+2] = byte(v >> 16)
	mem[offset+3] = byte(v >> 24)
}

func I64Write(mem []byte, offset Pointer, v int64) {
	mem[offset] = byte(v)
	mem[offset+1] = byte(v >> 8)
	mem[offset+2] = byte(v >> 16)
	mem[offset+3] = byte(v >> 24)
	mem[offset+4] = byte(v >> 32)
	mem[offset+5] = byte(v >> 40)
	mem[offset+6] = byte(v >> 48)
	mem[offset+7] = byte(v >> 56)
}

func UI8Write(mem []byte, offset Pointer, v uint8) {
	mem[offset] = v
}

func UI16Write(mem []byte, offset Pointer, v uint16) {
	mem[offset] = byte(v)
	mem[offset+1] = byte(v >> 8)
}

func UI32Write(memory []byte, offset Pointer, value uint32) {
	memory[offset] = byte(value)
	memory[offset+1] = byte(value >> 8)
	memory[offset+2] = byte(value >> 16)
	memory[offset+3] = byte(value >> 24)
}

func UI64Write(memory []byte, offset Pointer, value uint64) {
	memory[offset] = byte(value)
	memory[offset+1] = byte(value >> 8)
	memory[offset+2] = byte(value >> 16)
	memory[offset+3] = byte(value >> 24)
	memory[offset+4] = byte(value >> 32)
	memory[offset+5] = byte(value >> 40)
	memory[offset+6] = byte(value >> 48)
	memory[offset+7] = byte(value >> 56)
}

func F32Write(mem []byte, offset Pointer, f float32) {
	v := math.Float32bits(f)
	mem[offset] = byte(v)
	mem[offset+1] = byte(v >> 8)
	mem[offset+2] = byte(v >> 16)
	mem[offset+3] = byte(v >> 24)
}

func F64Write(mem []byte, offset Pointer, f float64) {
	v := math.Float64bits(f)
	mem[offset] = byte(v)
	mem[offset+1] = byte(v >> 8)
	mem[offset+2] = byte(v >> 16)
	mem[offset+3] = byte(v >> 24)
	mem[offset+4] = byte(v >> 32)
	mem[offset+5] = byte(v >> 40)
	mem[offset+6] = byte(v >> 48)
	mem[offset+7] = byte(v >> 56)
}

func SliceByteWrite(memory []byte, offset Pointer, byts []byte) {
	// TODO:PTR use copy()
	count := IntToPtrCast(len(byts))
	for c := Pointer(0); c < count; c++ {
		memory[offset+c] = byts[c]
	}
}

// TotalLength returns the number of elements in an array with the given lengths.
func TotalLength(lengths []Pointer) Pointer {
	total := Pointer(1)
	for _, i := range lengths {
		total *= i
	}

	return total
}

// definitionNames holds the name of each definition as bytes. It is built
// while the package initializes, so AppendName reads no string.
var definitionNames = definitionNamesMake()

func definitionNamesMake() [][]byte {
	names := make([][]byte, len(definitions))
	for i, definition := range definitions {
		names[i] = []byte(definition.name)
		names[i] = names[i][:len(names[i]):len(names[i])]
	}
	return names
}
