package bolt

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	bolt "go.etcd.io/bbolt"
)

// VIOLATION: state outside State. Owner: t0-plbolt; moves to PackageDBMng in phase 2.
var DBPath string

// VIOLATION: init chaining. Init 1 of the package loader tools: opens
// program_list.db next to the executable, called from Go package init.
func init() {
	path, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	DBPath = filepath.Dir(path)
	err = withDB(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte("program"))
			return err
		})
	})
	if err != nil {
		log.Fatal(err)
	}
}

// withDB opens the database, runs fn and closes it again. A close error is
// returned when fn itself succeeded.
func withDB(fn func(db *bolt.DB) error) (err error) {
	file := DBPath + "/program_list.db"
	db, err := bolt.Open(file, 0644, nil)
	if err != nil {
		return fmt.Errorf("bolt: open %s: %w", file, err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("bolt: close %s: %w", file, cerr)
		}
	}()
	return fn(db)
}

// Add stores value under key in the program bucket.
func Add(key string, value []byte) error {
	return withDB(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("program"))
			if bucket == nil {
				return errors.New("bolt: bucket program was not found")
			}
			if err := bucket.Put([]byte(key), value); err != nil {
				return fmt.Errorf("bolt: put %q: %w", key, err)
			}
			return nil
		})
	})
}

// Get returns a copy of the value stored under key. A missing key is an error.
func Get(key string) ([]byte, error) {
	var ret []byte
	err := withDB(func(db *bolt.DB) error {
		return db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("program"))
			if bucket == nil {
				return errors.New("bolt: bucket program was not found")
			}
			value := bucket.Get([]byte(key))
			if value == nil {
				return fmt.Errorf("bolt: no value for key %q", key)
			}
			ret = make([]byte, len(value))
			copy(ret, value)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}
