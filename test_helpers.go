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

func csvToSampleRecord(line string) (*KeyValue, error) {
	values := strings.Split(line, ",")
	if len(values) < 3 {
		return nil, fmt.Errorf("%v has less than 3 values", line)
	}

	return &KeyValue{
		Key: line,
		Value: sampleRecord{values[0], values[1], values[2]},
	}, nil
}

func readDB(dir string) ([]sampleRecord, error) {
	db, err := openDB(dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	chkv, cherr := make(chan kvBytes), make(chan error)
	go func(chan kvBytes, chan error) {
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
				kv := kvBytes{key, value}
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
