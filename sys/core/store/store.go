package store

import (
	"context"
	"os"
	"path/filepath"

	logging "github.com/ipfs/go-log/v2"

	ds "github.com/ipfs/go-datastore"
	badger "github.com/textileio/go-ds-badger3"
)

const (
	DB             = "db"
	KEYBYTES       = 64
	INDEXCACHESIZE = 100 << 20
)

var (
	ErrNotFound error           = ds.ErrNotFound
	whatever    context.Context = context.TODO()

	log = logging.Logger("slater:store")
)

type Store struct {
	Store *badger.Datastore
}

func FindStores(rootPath string) ([]string, error) {
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

func OpenStore(rootPath, name string, key string) (Store, error) {
	path := filepath.Join(rootPath, name, DB)

	opts := badger.DefaultOptions
	opts.WithIndexCacheSize(INDEXCACHESIZE)
	opts.WithEncryptionKey([]byte(key))
	//opts.WithEncryptionKeyRotationDuration(...)

	store, err := badger.NewDatastore(path, &opts)

	if err != nil {
		return Store{nil}, err
	}

	return Store{store}, nil
}

func RemoveStore(rootPath, name string) {
	storePath := filepath.Join(rootPath, name, DB)
	path := filepath.Join(storePath, name)
	err := os.RemoveAll(path)
	if err != nil {
		log.Panic("Could not delete store.")
	}
}

func (s Store) Put(ns []string, v []byte) error {
	key := ds.KeyWithNamespaces(ns)
	return s.Store.Put(whatever, key, v)
}

func (s Store) Get(key string) (value []byte, err error) {
	k := ds.NewKey(key)
	value, err = s.Store.Get(whatever, k)
	return
}
