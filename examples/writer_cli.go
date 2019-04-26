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

type sampleValues struct {
	Field1 string
	Field2 string
}

type sampleRecord struct {
	Key   []string
	Value sampleValues
}

func csvToKeyValue(line string) (*badgerutils.KeyValue, error) {
	values := strings.Split(line, ",")
	if len(values) < 4 {
		return nil, fmt.Errorf("%v has less than 4 values", line)
	}

	return &badgerutils.KeyValue{
		Key:   []interface{}{values[0], values[1]},
		Value: sampleValues{values[2], values[3]},
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
