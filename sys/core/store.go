package core

import (
	"context"
	"os"
	"path/filepath"

	ds "github.com/ipfs/go-datastore"
	badger "github.com/textileio/go-ds-badger3"
)

const (
	ROOT           = ".slater"
	SESSIONS       = "data"
	KEYBYTES       = 64
	KEYKEY         = "k"
	DEVICESKEY     = "d"
	INDEXCACHESIZE = 100 << 20
)

var (
	home        string
	storePath   string
	ErrNotFound error           = ds.ErrNotFound
	whatever    context.Context = context.TODO()
)

type datastore struct {
	store *badger.Datastore
}

func init() {
	root := ROOT

	if len(os.Args) > 1 {
		alt := os.Args[1] // alternate root, to run 2 instances for testing
		if alt != "" {
			root = alt
		}
	}

	home, _ = os.UserHomeDir()
	storePath = filepath.Join(home, root, SESSIONS)

	log.Debug("STORE PATH:", storePath)
}

func findStores() ([]string, error) {
	var stores []string

	if err := os.MkdirAll(storePath, 0700); err != nil {
		return stores, err
	}

	dir, err := os.Open(storePath)
	if err != nil {
		return stores, err
	}

	stores, err = dir.Readdirnames(0)
	if err != nil {
		return stores, err
	}

	return stores, nil
}

func openStore(name string, key string) (datastore, error) {
	path := filepath.Join(storePath, name)

	opts := badger.DefaultOptions
	opts.WithIndexCacheSize(INDEXCACHESIZE)
	opts.WithEncryptionKey([]byte(key))
	//opts.WithEncryptionKeyRotationDuration(...)

	store, err := badger.NewDatastore(path, &opts)

	if err != nil {
		return datastore{nil}, err
	}

	return datastore{store}, nil
}

func deleteStore(name string) {
	path := filepath.Join(storePath, name)
	err := os.RemoveAll(path)
	if err != nil {
		log.Panic("Could not delete store.")
	}
}

func (s datastore) put(k string, v []byte) error {
	key := ds.NewKey(k)
	return s.store.Put(whatever, key, v)
}

func (s datastore) get(key string) (value []byte, err error) {
	k := ds.NewKey(key)
	value, err = s.store.Get(whatever, k)
	if err != nil {
		log.Debugf("%s: %s", k, err)
	}
	return
}
