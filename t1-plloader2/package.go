package loader

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/skycoin/skycoin-cx/t0-plbolt" // thin: only Add; no depth gain
	"golang.org/x/crypto/blake2b"
)

type Package struct {
	PackageName string
	Blake2Hash  string
	Files       []string
}

func (p Package) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Package) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}

// Encode a file and put it in the specified package
func (newPackage *Package) appendFile(newFile *File) error {
	if newFile == nil {
		return errors.New("append file: nil file")
	}
	hash, err := newFile.getHash()
	if err != nil {
		return err
	}
	newPackage.Files = append(newPackage.Files, fmt.Sprintf("%x", hash[:]))
	err = newFile.saveToDatabase(hash)
	if err != nil {
		return err
	}
	return nil
}

func (newPackage *Package) getHash() ([64]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(newPackage)
	if err != nil {
		return [64]byte{}, err
	}
	return blake2b.Sum512(buffer.Bytes()), err
}

func (newPackage *Package) saveToDatabase(hash [64]byte) error {
	key := fmt.Sprintf("%x", hash[:])
	value, err := newPackage.MarshalBinary()
	if err != nil {
		return fmt.Errorf("encode package %s: %w", newPackage.PackageName, err)
	}
	if err := bolt.Add(key, value); err != nil {
		return fmt.Errorf("save package %s: %w", newPackage.PackageName, err)
	}
	PackageHashMap[newPackage.PackageName] = key
	return nil
}
