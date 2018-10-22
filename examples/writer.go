package main

import (
	"fmt"
	"github.com/Surfline/badgerutils"
	"log"
	"strings"
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
	if err := badgerutils.WriteStdin(lineToKeyed); err != nil {
		log.Fatal(err)
	}
}
