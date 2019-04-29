package badgerutils

import (
	"fmt"
	"strings"

	"github.com/dgraph-io/badger"
)

type sampleRecord struct {
	Key   string
	Value string
}

func csvToKeyValue(line string) (*KeyValue, error) {
	kv := strings.Split(line, ":")
	if len(kv) < 2 {
		return nil, fmt.Errorf("%v has less than 2 kv", line)
	}

	return &KeyValue{
		Key:   []byte(kv[0]),
		Value: []byte(kv[1]),
	}, nil
}

func readDB(dir string) ([]sampleRecord, error) {
	db, err := openDB(dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	chkv, cherr := make(chan KeyValue), make(chan error)
	go func(chan KeyValue, chan error) {
		err := db.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			it := txn.NewIterator(opts)
			defer it.Close()
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				key := item.Key()
				value, err := item.Value()
				if err != nil {
					return err
				}
				kv := KeyValue{Key: key, Value: value}
				chkv <- kv
			}
			close(chkv)
			return nil
		})
		cherr <- err
	}(chkv, cherr)

	sampleRecords := make([]sampleRecord, 0)
	for kv := range chkv {
		sampleRecords = append(sampleRecords, sampleRecord{
			Key:   string(kv.Key),
			Value: string(kv.Value),
		})
	}

	if err := <-cherr; err != nil {
		return nil, err
	}

	return sampleRecords, nil
}
