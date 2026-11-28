//go:build !cximageembed

package main

import "os"

// coreImageFileRead reads the image file named by the first argument. A
// build with the cximageembed tag embeds the image instead. The arguments
// after the file are the program arguments.
func coreImageFileRead() ([]byte, []string, string) {
	if len(os.Args) < 2 {
		return nil, nil, "give the image file that cx build --target=image wrote"
	}
	file, err := os.ReadFile(os.Args[1])
	if err != nil {
		return nil, nil, err.Error()
	}
	return file, os.Args[2:], ""
}
