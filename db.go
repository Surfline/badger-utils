package badgerutils

import (
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/options"
)

func openDB(dir string) (*badger.DB, error) {
	opts := badger.DefaultOptions
	opts.ValueLogLoadingMode = options.FileIO
	opts.TableLoadingMode = options.FileIO
	opts.Dir = dir
	opts.ValueDir = dir
	return badger.Open(opts)
}
