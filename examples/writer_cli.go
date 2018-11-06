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

func (r sampleRecord) Key() string {
	return fmt.Sprintf("%v,%v,%v", r.Field1, r.Field2, r.Field3)
}

func lineToKeyed(line string) (badgerutils.Keyed, error) {
	values := strings.Split(line, ",")
	return sampleRecord{values[0], values[1], values[2]}, nil
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

	if err := badgerutils.WriteStream(os.Stdin, *dir, *batchSize, lineToKeyed); err != nil {
		log.Fatal(err)
	}
}
