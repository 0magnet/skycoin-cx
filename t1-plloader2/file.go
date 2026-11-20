package loader

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	"github.com/skycoin/skycoin-cx/t0-plbolt" // thin: only Add; no depth gain
	"golang.org/x/crypto/blake2b"
)

type File struct {
	FileName   string
	Length     uint32
	Content    []byte
	Blake2Hash string
}

func (f File) MarshalBinary() ([]byte, error) {
	return json.Marshal(f)
}

func (f *File) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, f)
}

func (newFile *File) getHash() ([64]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(newFile)
	if err != nil {
		return [64]byte{}, err
	}
	return blake2b.Sum512(buffer.Bytes()), err
}

func (newFile *File) saveToDatabase(hash [64]byte) error {
	key := fmt.Sprintf("%x", hash[:])
	value, err := newFile.MarshalBinary()
	if err != nil {
		return fmt.Errorf("encode file %s: %w", newFile.FileName, err)
	}
	if err := bolt.Add(key, value); err != nil {
		return fmt.Errorf("save file %s: %w", newFile.FileName, err)
	}
	FileHashMap[newFile.FileName] = key
	return nil
}
