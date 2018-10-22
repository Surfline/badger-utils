package badgerutils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"strings"
	"testing"
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

func TestWriteInput(t *testing.T) {
	dir, err := os.Getwd()
	require.Nil(t, err)
	tmpDir, err := ioutil.TempDir(dir, "temp")
	require.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	reader := strings.NewReader(`field11,field12,field13
field21,field22,field23
field31,field32,field33`)
	err = writeInput(reader, tmpDir, 2, csvToSampleRecord)
	require.Nil(t, err)

	writtenSampleRecords, err := readDB(tmpDir)
	require.Nil(t, err)
	require.Equal(t, 3, len(writtenSampleRecords))
	require.EqualValues(t, writtenSampleRecords[0], sampleRecord{"field11", "field12", "field13"})
	require.EqualValues(t, writtenSampleRecords[1], sampleRecord{"field21", "field22", "field23"})
	require.EqualValues(t, writtenSampleRecords[2], sampleRecord{"field31", "field32", "field33"})
}
