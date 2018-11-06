# Badger Utils [![GoDoc](https://godoc.org/github.com/Surfline/badgerutils?status.svg)](https://godoc.org/github.com/Surfline/badgerutils) [![Go Report Card](https://goreportcard.com/badge/github.com/Surfline/badgerutils)](https://goreportcard.com/report/github.com/Surfline/badgerutils)

Go package with utilities for interacting with [Badger](https://github.com/dgraph-io/badger).

## Table of Contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Getting Started](#getting-started)
  - [IO Stream to Badger](#io-stream-to-badger)
    - [Example](#example)
- [Development](#development)
  - [Dependency Management](#dependency-management)
  - [Format Code](#format-code)
  - [Unit Tests](#unit-tests)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Getting Started

### IO Stream to Badger

To stream data Badger, use `badgerutils.WriteStream`.

#### Example

Creates a CLI tool that streams data from stdin.

```Go
// examples/writer_cli.go
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
```

The code above can be called with the following flags:

- `-dir` - (required) The path to the directory to persist Badger files.
- `-batch-size` - (default: `1000`) The size of each transaction (or batch of writes). This can be tuned for optimal performance depending on the machine.

For example:

```sh
$ for i in {1..10}; do echo "field${i}1,field${i}2,field${i}3"; done | go run main.go -dir=temp -batch-size=1
Directory: temp
Batch Size: 3
...
Records: 3
Records: 6
Records: 9
Records: 10
Inserted 10 records in 474.69µs
```

## Development

### Dependency Management

[dep](https://github.com/golang/dep) is required for dependency management.

```sh
$ make install
```

### Format Code

Run this before opening pull requests to ensure code is properly formatted.

```sh
$ make fmt
```

### Unit Tests

```sh
$ make test
```
