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

// Keyed interface defines a Key method for defining a key from a struct.
type Keyed interface {
	Key() string
}

type keyValue struct {
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

func stringToKeyValue(str string, lineToKeyed func(string) (Keyed, error)) (*keyValue, error) {
	record, err := lineToKeyed(str)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if err = gob.NewEncoder(buf).Encode(record); err != nil {
		return nil, err
	}

	return &keyValue{
		Key:   []byte(record.Key()),
		Value: buf.Bytes(),
	}, nil
}

func writeBatch(kvs []keyValue, db *badger.DB, cherr chan error, done func(int32)) {
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
// lineToKeyed function parameter defines how stdin is translated to a value and how to define a key
// from that value.
func WriteStream(reader io.Reader, dir string, batchSize int, lineToKeyed func(string) (Keyed, error)) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	db, err := openDB(dir)
	if err != nil {
		return err
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

	kvBatch := make([]keyValue, 0)
	cherr := make(chan error)

	// Read from stream and write key/values in batches
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		kv, err := stringToKeyValue(scanner.Text(), lineToKeyed)
		if err != nil {
			return err
		}
		kvBatch = append(kvBatch, *kv)
		if len(kvBatch) == batchSize {
			wg.Add(1)
			go writeBatch(kvBatch, db, cherr, done)
			kvBatch = make([]keyValue, 0)
		}
	}

	// Write remaining key/values
	if len(kvBatch) > 0 {
		wg.Add(1)
		writeBatch(kvBatch, db, cherr, done)
	}

	// Read and handle errors from stream
	if err = scanner.Err(); err != nil {
		return err
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
