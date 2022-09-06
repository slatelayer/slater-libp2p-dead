package core

import (
	"context"
	"os"
	"path/filepath"

	ds "github.com/ipfs/go-datastore"
	badger "github.com/textileio/go-ds-badger3"
)

const (
	DB             = "db"
	KEYBYTES       = 64
	KEYKEY         = "k"
	DEVICESKEY     = "d"
	INDEXCACHESIZE = 100 << 20
)

var (
	ErrNotFound error           = ds.ErrNotFound
	whatever    context.Context = context.TODO()
)

type datastore struct {
	store *badger.Datastore
}

func findStores(rootPath string) ([]string, error) {
	var stores []string

	dir, err := os.Open(rootPath)
	if err != nil {
		return stores, err
	}

	stores, err = dir.Readdirnames(0)
	if err != nil {
		return stores, err
	}

	return stores, nil
}

func openStore(rootPath, name string, key string) (datastore, error) {
	path := filepath.Join(rootPath, name, DB)

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

func deleteStore(rootPath, name string) {
	storePath := filepath.Join(rootPath, name, DB)
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
	//if err != nil {
	//	log.Debugf("%s: %s", k, err)
	//}
	return
}
