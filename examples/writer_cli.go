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
	Field1 string
	Field2 string
	Field3 string
}

func lineToKeyValue(line string) (*badgerutils.KeyValue, error) {
	values := strings.Split(line, ",")
	if len(values) < 3 {
		return nil, fmt.Errorf("%v has less than 3 values", line)
	}

	return &badgerutils.KeyValue{
		Key: line,
		Value: sampleRecord{values[0], values[1], values[2]},
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

	if err := badgerutils.WriteStream(os.Stdin, *dir, *batchSize, lineToKeyValue); err != nil {
		log.Fatal(err)
	}
}
