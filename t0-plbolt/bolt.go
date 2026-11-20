package bolt

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	bolt "go.etcd.io/bbolt"
)

// VIOLATION: state outside State. Owner: t0-plbolt; moves to PackageDBMng in phase 2.
var DBPath string

// Init is init 1 of the package loader tools. It points DBPath at the
// executable directory and creates the program bucket in program_list.db there.
// t3-packageloader main and the test mains call it before Add or Get.
func Init() error {
	path, err := os.Executable()
	if err != nil {
		return fmt.Errorf("bolt: find executable: %w", err)
	}
	DBPath = filepath.Dir(path)
	return withDB(func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte("program"))
			return err
		})
	})
}

// withDB opens the database, runs fn and closes it again. A close error is
// returned when fn itself succeeded.
func withDB(fn func(db *bolt.DB) error) (err error) {
	if DBPath == "" {
		return errors.New("bolt: DBPath is not set, call Init first")
	}
	file := filepath.Join(DBPath, "program_list.db")
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
	if key == "" {
		return errors.New("bolt: put: empty key")
	}
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
	if key == "" {
		return nil, errors.New("bolt: get: empty key")
	}
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
