//go:build cxfx

package graphics

import (
	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"golang.org/x/mobile/exp/audio/al"
)

func toBuffers(in interface{}) []al.Buffer { // REFACTOR : ??
	var out []al.Buffer
	var buffers []int32 = in.([]int32)
	for _, b := range buffers {
		out = append(out, al.Buffer(b))
	}
	return out
}

func toSources(in interface{}) []al.Source { // REFACTOR : ??
	var out []al.Source
	var sources []int32 = in.([]int32)
	for _, s := range sources {
		out = append(out, al.Source(s))
	}
	return out
}

func OpAlCloseDevice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	al.CloseDevice()
}

func OpAlDeleteBuffers(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	buffers := toBuffers(inputs[0].SliceI32Get(prgrm))
	al.DeleteBuffers(buffers...)
}

func OpAlDeleteSources(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	sources := toSources(inputs[0].SliceI32Get(prgrm))
	al.DeleteSources(sources...)
}

func OpAlDeviceError(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	err := al.DeviceError()
	outputs[0].I32Set(prgrm, err)
}

func OpAlError(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	err := al.Error()
	outputs[0].I32Set(prgrm, err)
}

// VIOLATION STRHOT-192: B, OpenAL returns text, #432.
func OpAlExtensions(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	extensions := al.Extensions()
	outputs[0].StrSet(prgrm, extensions)
}

func OpAlOpenDevice(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	if err := al.OpenDevice(); err != nil {
		panic(err)
	}
}

func OpAlPauseSources(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	sources := toSources(inputs[0].SliceI32Get(prgrm))
	al.PauseSources(sources...)
}

func OpAlPlaySources(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	sources := toSources(inputs[0].SliceI32Get(prgrm))
	al.PlaySources(sources...)
}

// VIOLATION STRHOT-193: B, OpenAL returns text, #432.
func OpAlRenderer(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	renderer := al.Renderer()
	outputs[0].StrSet(prgrm, renderer)
}

func OpAlRewindSources(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	sources := toSources(inputs[0].SliceI32Get(prgrm))
	al.RewindSources(sources...)
}

func OpAlStopSources(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	sources := toSources(inputs[0].SliceI32Get(prgrm))
	al.StopSources(sources...)
}

// VIOLATION STRHOT-194: B, OpenAL returns text, #432.
func OpAlVendor(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	vendor := al.Vendor()
	outputs[0].StrSet(prgrm, vendor)
}

// VIOLATION STRHOT-195: B, OpenAL returns text, #432.
func OpAlVersion(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	version := al.Version()
	outputs[0].StrSet(prgrm, version)
}

func OpAlGenBuffers(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	buffers := al.GenBuffers(int(inputs[0].I32Get(prgrm)))
	outputSlicePointer := outputs[0].Offset
	outputSliceOffset := valuetypes.PtrRead(prgrm.Memory, outputSlicePointer)
	for _, b := range buffers { // REFACTOR append with copy ?
		var obj [4]byte
		valuetypes.I32Write(obj[:], 0, int32(b))
		outputSliceOffset = ast.SliceWrite(prgrm, outputSliceOffset, obj[:])
	}
	outputs[0].PtrSet(prgrm, outputSliceOffset)
}

func OpAlBufferData(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	buffer := al.Buffer(inputs[0].I32Get(prgrm))
	format := inputs[1].I32Get(prgrm)
	samples := toBytes(inputs[2].SliceUI8Get(prgrm))
	frequency := inputs[3].I32Get(prgrm)
	buffer.BufferData(uint32(format), samples, frequency)
}

func OpAlGenSources(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	sources := al.GenSources(int(inputs[0].I32Get(prgrm)))
	outputSlicePointer := outputs[0].Offset
	outputSliceOffset := valuetypes.PtrRead(prgrm.Memory, outputSlicePointer)
	for _, s := range sources { // REFACTOR append with copy ?
		var obj [4]byte
		valuetypes.I32Write(obj[:], 0, int32(s))
		outputSliceOffset = ast.SliceWrite(prgrm, outputSliceOffset, obj[:])
	}
	outputs[0].PtrSet(prgrm, outputSliceOffset)
}

func OpAlSourceBuffersProcessed(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	source := al.Source(inputs[0].I32Get(prgrm))
	outputs[0].I32Set(prgrm, source.BuffersProcessed())
}

func OpAlSourceBuffersQueued(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	source := al.Source(inputs[0].I32Get(prgrm))
	outputs[0].I32Set(prgrm, source.BuffersQueued())
}

func OpAlSourceQueueBuffers(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	source := al.Source(inputs[0].I32Get(prgrm))
	buffers := toBuffers(inputs[1].SliceI32Get(prgrm))
	source.QueueBuffers(buffers...)
}

func OpAlSourceState(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	source := al.Source(inputs[0].I32Get(prgrm))
	outputs[0].I32Set(prgrm, source.State())
}

func OpAlSourceUnqueueBuffers(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	source := al.Source(inputs[0].I32Get(prgrm))
	buffers := toBuffers(inputs[1].SliceI32Get(prgrm))
	source.UnqueueBuffers(buffers...)
}
