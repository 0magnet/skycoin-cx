//go:build cximageembed

package main

import (
	_ "embed"
	"os"
	"unsafe"
)

// coreImageEmbedded is the image make build-mcu copies to program.cximage
// before it builds this package. A string stays in flash on a
// microcontroller, where a []byte would be copied to RAM.
//
//go:embed program.cximage
var coreImageEmbedded string

// coreImageFileRead returns the embedded image. Nothing writes the bytes:
// ImageFileDecode only reads them and Machine.Reset copies the data. Every
// argument is a program argument.
func coreImageFileRead() ([]byte, []string, string) {
	return unsafe.Slice(unsafe.StringData(coreImageEmbedded), len(coreImageEmbedded)), os.Args[1:], ""
}
