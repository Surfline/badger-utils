package badgerutils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger"
)

type sampleRecord struct {
	Field1 string
	Field2 string
	Field3 string
}

func (r sampleRecord) Key() string {
	return fmt.Sprintf("%v,%v,%v", r.Field1, r.Field2, r.Field3)
}

func csvToSampleRecord(line string) (Keyed, error) {
	values := strings.Split(line, ",")
	return sampleRecord{values[0], values[1], values[2]}, nil
}

func readDB(dir string) ([]sampleRecord, error) {
	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	chkv, cherr := make(chan keyValue), make(chan error)
	go func(chan keyValue, chan error) {
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
				kv := keyValue{key, value}
				chkv <- kv
			}
			close(chkv)
			return nil
		})
		cherr <- err
	}(chkv, cherr)

	sampleRecords := make([]sampleRecord, 0)
	for kv := range chkv {
		var sr sampleRecord
		buf := bytes.NewReader(kv.Value)
		if err := gob.NewDecoder(buf).Decode(&sr); err != nil {
			return nil, err
		}
		sampleRecords = append(sampleRecords, sr)
	}

	if err := <-cherr; err != nil {
		return nil, err
	}

	return sampleRecords, nil
}
