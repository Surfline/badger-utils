package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Surfline/badgerutils"
)

type sampleRecord struct {
	Key   string
	Value string
}

func csvToKeyValue(line string) (*badgerutils.KeyValue, error) {
	kv := strings.Split(line, ":")
	if len(kv) < 2 {
		return nil, fmt.Errorf("%v has less than 2 kv", line)
	}

	return &badgerutils.KeyValue{
		Key:   []byte(kv[0]),
		Value: []byte(kv[1]),
	}, nil
}

func main() {
	dir := flag.String("dir", "", "Directory to save DB files")
	batchSize := flag.Int("batch-size", 1000, "Number of records to write per transaction")
	flag.Parse()

	if *dir == "" {
		log.Fatal(errors.New("dir flag is required"))
	}

	log.Printf("Directory: %v", *dir)
	log.Printf("Batch Size: %v", *batchSize)

	if err := badgerutils.WriteStream(os.Stdin, *dir, *batchSize, csvToKeyValue); err != nil {
		log.Fatal(err)
	}
}
