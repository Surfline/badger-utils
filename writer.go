// Package badgerutils provides functions for interacting with the underlying database.
package badgerutils

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"flag"
	"fmt"
	"github.com/dgraph-io/badger"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

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

func writeInput(reader io.Reader, dir string, batchSize int, lineToKeyed func(string) (Keyed, error)) error {
	log.Printf("Directory: %v", dir)
	log.Printf("Batch Size: %v", batchSize)

	// Open Badger database from directory
	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	db, err := badger.Open(opts)
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

	// Read from stdin and write key/values in batches
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

	// Read and handle errors streaming from stdin
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

// WriteStdin translates stdin into key/value pairs that are written into the Badger.
// lineToKeyed function parameter defines how stdin is translated to a value and how to define a key
// from that value.
func WriteStdin(lineToKeyed func(string) (Keyed, error)) error {
	dir := flag.String("dir", "", "Directory to save DB files")
	batchSize := flag.Int("batch-size", 1000, "Number of records to write per transaction")
	flag.Parse()

	if *dir == "" {
		return errors.New("dir flag is required")
	}

	return writeInput(os.Stdin, *dir, *batchSize, lineToKeyed)
}
