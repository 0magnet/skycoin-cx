//go:build cxfx

package graphics

import (
	"bufio"
	"github.com/madelynnblue/go-dsp/wav"
	"github.com/skycoin/skycoin-cx/t0-cx-core-valuetypes"
	"github.com/skycoin/skycoin-cx/t3-cx-core-ast"
	"github.com/skycoin/skycoin/src/cipher/encoder"
	//"golang.org/x/mobile/exp/audio/al"
)

// GCREACH: allocates on the heap, which adds to the GC pressure,
// see docs/GarbageCollector.md.
// VIOLATION STRHOT-196: B, reads the file at a program path, #432.
func OpAlLoadWav(prgrm *ast.CXProgram, inputs []ast.CXValue, outputs []ast.CXValue) {
	file, fileError := ast.ProgramSandboxFileOpen(prgrm, &inputs[0])
	if fileError != nil {
		panic(fileError.FileErrorMessage)
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	wav, err := wav.New(reader)
	if err != nil {
		panic(err)
	}

	samples, err := wav.ReadSamples(wav.Samples)
	if err != nil {
		panic(err)
	}

	encoded := encoder.Serialize(samples)

	outputs[0].I32Set(prgrm, int32(wav.Header.AudioFormat))
	outputs[1].I32Set(prgrm, int32(wav.Header.NumChannels))
	outputs[2].I32Set(prgrm, int32(wav.Header.SampleRate))
	outputs[3].I32Set(prgrm, int32(wav.Header.ByteRate))
	outputs[4].I32Set(prgrm, int32(wav.Header.BlockAlign))
	outputs[5].I32Set(prgrm, int32(wav.Header.BitsPerSample))
	outputs[6].I32Set(prgrm, int32(wav.Samples))
	outputs[7].I64Set(prgrm, int64(wav.Duration))

	outputSlicePointer := outputs[8].Offset
	outputSliceOffset := valuetypes.PtrRead(prgrm.Memory, outputSlicePointer)
	outputSliceOffset = ast.SliceResizeEx(prgrm, outputSliceOffset, valuetypes.IntToPtrCast(len(encoded)), 1)
	copy(ast.SliceElementsGet(prgrm, outputSliceOffset, 1), encoded)
	outputs[8].PtrSet(prgrm, outputSliceOffset)
}

func toBytes(in interface{}) []byte { // REFACTOR : ??
	if in != nil {
		return in.([]byte)
	}
	return nil
}
