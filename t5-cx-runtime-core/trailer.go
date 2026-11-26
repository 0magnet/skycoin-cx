package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"

	sandbox "github.com/skycoin/skycoin-cx/t1-cx-runtime-sandbox"
)

// The trailer ends a file that cx build made: the core binary, the serialized
// program, then TrailerSize bytes. The layout is in "Built programs" in
// docs/IR.md. All integers are little endian.
const (
	TrailerSize          = 32
	TrailerMagic         = "CXBUNDLE"
	TrailerFormatVersion = 1
)

// Trailer describes the program a trailer marks.
type Trailer struct {
	TrailerOffset   uint64 // offset of the program from the start of the file
	TrailerLength   uint64 // length of the program in bytes
	TrailerChecksum uint32 // CRC-32 (IEEE) of the program
}

// TrailerAppend returns core followed by program and its trailer.
func TrailerAppend(core []byte, program []byte) []byte {
	out := make([]byte, 0, len(core)+len(program)+TrailerSize)
	out = append(out, core...)
	out = append(out, program...)
	out = binary.LittleEndian.AppendUint64(out, uint64(len(core)))
	out = binary.LittleEndian.AppendUint64(out, uint64(len(program)))
	out = binary.LittleEndian.AppendUint32(out, crc32.ChecksumIEEE(program))
	out = binary.LittleEndian.AppendUint32(out, TrailerFormatVersion)
	return append(out, TrailerMagic...)
}

// TrailerDecode reads the trailer from tail, the last TrailerSize
// bytes of a file of fileSize bytes. found is false when tail is no trailer.
// problem is not empty when it is one that does not fit the file.
func TrailerDecode(tail []byte, fileSize uint64) (trailer Trailer, found bool, problem string) {
	if len(tail) != TrailerSize || !bytes.Equal(tail[24:], []byte(TrailerMagic)) {
		return trailer, false, ""
	}
	trailer.TrailerOffset = binary.LittleEndian.Uint64(tail[0:])
	trailer.TrailerLength = binary.LittleEndian.Uint64(tail[8:])
	trailer.TrailerChecksum = binary.LittleEndian.Uint32(tail[16:])
	if version := binary.LittleEndian.Uint32(tail[20:]); version != TrailerFormatVersion {
		return trailer, true, fmt.Sprintf("trailer format %d, this core reads format %d", version, TrailerFormatVersion)
	}
	end := trailer.TrailerOffset + trailer.TrailerLength
	if end < trailer.TrailerOffset || end+TrailerSize != fileSize {
		return trailer, true, fmt.Sprintf("trailer gives a program of %d bytes at %d, which does not end at the trailer of a %d byte file", trailer.TrailerLength, trailer.TrailerOffset, fileSize)
	}
	return trailer, true, ""
}

// TrailerRead returns the program appended to the file at path, nil
// when the file ends in no trailer.
func TrailerRead(path string) ([]byte, *sandbox.FileError) {
	file, err := os.Open(path)
	if err != nil {
		return nil, sandbox.FileErrorWrap(sandbox.FileOperationOpen, path, err)
	}
	defer file.Close()
	fileStat, err := file.Stat()
	if err != nil {
		return nil, sandbox.FileErrorWrap(sandbox.FileOperationRead, path, err)
	}
	size := fileStat.Size()
	if size < TrailerSize {
		return nil, nil
	}
	tail := make([]byte, TrailerSize)
	if _, err := file.ReadAt(tail, size-TrailerSize); err != nil {
		return nil, sandbox.FileErrorWrap(sandbox.FileOperationRead, path, err)
	}
	trailer, found, problem := TrailerDecode(tail, uint64(size))
	if !found {
		return nil, nil
	}
	if problem != "" {
		return nil, coreTrailerContentInvalid(path, problem)
	}
	program := make([]byte, trailer.TrailerLength)
	if _, err := file.ReadAt(program, int64(trailer.TrailerOffset)); err != nil && err != io.EOF {
		return nil, sandbox.FileErrorWrap(sandbox.FileOperationRead, path, err)
	}
	if crc32.ChecksumIEEE(program) != trailer.TrailerChecksum {
		return nil, coreTrailerContentInvalid(path, "the appended program does not match its checksum")
	}
	return program, nil
}

func coreTrailerContentInvalid(path string, problem string) *sandbox.FileError {
	return &sandbox.FileError{
		FileErrorCode:      sandbox.FileErrorCodeContentInvalid,
		FileErrorOperation: sandbox.FileOperationRead,
		FileErrorMessage:   problem,
		FileErrorPath:      path,
	}
}
