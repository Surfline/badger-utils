package badgerutils

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/badger"
)

// KeyValue struct defines a Key and a Value empty interface to be translated into a record.
type KeyValue struct {
	Key   interface{}
	Value interface{}
}

type kvBytes struct {
	Key   []byte
	Value []byte
}

type count32 int32

func (c *count32) increment(a int32) int32 {
	return atomic.AddInt32((*int32)(c), a)
}

func (c *count32) get() int32 {
	return atomic.LoadInt32((*int32)(c))
}

func stringToKVBytes(str string, lineToKeyValue func(string) (*KeyValue, error)) (*kvBytes, error) {
	record, parseErr := lineToKeyValue(str)
	if parseErr != nil {
		return nil, parseErr
	}

	keyBuf := &bytes.Buffer{}
	if keyErr := gob.NewEncoder(keyBuf).Encode(record.Key); keyErr != nil {
		return nil, keyErr
	}

	valBuf := &bytes.Buffer{}
	if valErr := gob.NewEncoder(valBuf).Encode(record.Value); valErr != nil {
		return nil, valErr
	}

	return &kvBytes{
		Key:   keyBuf.Bytes(),
		Value: valBuf.Bytes(),
	}, nil
}

func writeBatch(kvs []kvBytes, db *badger.DB, cherr chan error, done func(int32)) {
	txn := db.NewTransaction(true)
	defer txn.Discard()

	for _, kv := range kvs {
		if err := txn.Set(kv.Key, kv.Value); err != nil {
			cherr <- err
		}
	}

	txn.Commit(func(err error) {
		if err != nil {
			cherr <- err
		}
		done(int32(len(kvs)))
	})
}

// WriteStream translates io.Reader stream into key/value pairs that are written into the Badger.
// lineToKeyValue function parameter defines how stdin is translated to a value and how to define a key
// from that value.
func WriteStream(reader io.Reader, dir string, batchSize int, lineToKeyValue func(string) (*KeyValue, error)) error {
	if mkdirErr := os.MkdirAll(dir, os.ModePerm); mkdirErr != nil {
		return mkdirErr
	}

	db, dbErr := openDB(dir)
	if dbErr != nil {
		return dbErr
	}
	defer db.Close()

	start := time.Now()

	// Wait group ensures all transactions are committed before reading errors from channel
	var wg sync.WaitGroup
	var kvCount count32
	done := func(processedCount int32) {
		kvCount.increment(processedCount)
		log.Printf("Records: %v\n", int32(kvCount))
		wg.Done()
	}

	kvBatch := make([]kvBytes, 0)
	cherr := make(chan error)

	// Read from stream and write key/values in batches
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		kv, err := stringToKVBytes(scanner.Text(), lineToKeyValue)
		if err != nil {
			return err
		}
		kvBatch = append(kvBatch, *kv)
		if len(kvBatch) == batchSize {
			wg.Add(1)
			go writeBatch(kvBatch, db, cherr, done)
			kvBatch = make([]kvBytes, 0)
		}
	}

	// Write remaining key/values
	if len(kvBatch) > 0 {
		wg.Add(1)
		writeBatch(kvBatch, db, cherr, done)
	}

	// Read and handle errors from stream
	if streamErr := scanner.Err(); streamErr != nil {
		return streamErr
	}

	wg.Wait()
	close(cherr)

	// Read and handle transaction errors
	errs := make([]string, 0)
	for err := range cherr {
		errs = append(errs, fmt.Sprintf("%v", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("Errors inserting records:\n%v", strings.Join(errs, "\n"))
	}

	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("Inserted %v records in %v", kvCount.get(), elapsed)
	return nil
}
