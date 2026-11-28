package ast

import (
	"errors"
	"testing"
)

// The failure codes give the text the format strings gave before #412.
func TestSerializeFailureText(t *testing.T) {
	var decoder serializeDecoder
	decoder.serializeDecoderFailNamed(serializeFailureMngLive, serializeNameCXTypeSignatureArrayMng, uint64(3), uint64(4))
	section := serializeDecoder{}
	section.serializeDecoderFail(serializeSectionError{serializeNameSectionTypeSignatures, decoder.serializeDecoderError})
	want := "type signatures section: serialized program is malformed: CXTypeSignature_ArrayMng has 3 live slots, Len 4"
	if got := section.serializeDecoderError.Error(); got != want {
		t.Errorf("text %q, want %q", got, want)
	}
	var truncated serializeDecoder
	truncated.serializeDecoderFail(serializeSectionError{serializeNameSectionMemory, errSerializeTruncated})
	if !errors.Is(truncated.serializeDecoderError, errSerializeTruncated) {
		t.Error("a section error does not unwrap to its cause")
	}
	var gid serializeDecoder
	gid.serializeDecoderFailWith(serializeFailureStringGid, int64(0x0f00000001))
	if got := gid.serializeDecoderError.Error(); got != "serialized program is malformed: string Gid 0xf00000001" {
		t.Errorf("text %q", got)
	}
	if err := SerializedHeaderCheck([]byte("not a program at all, long enough")); err == nil || err.Error() != "not a serialized CX program" {
		t.Errorf("header check gave %v", err)
	}
}
