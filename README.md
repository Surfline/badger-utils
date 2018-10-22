# Badger Utils

Go package with utilities for interacting with [Badger](https://github.com/dgraph-io/badger).

## Getting Started

### Stream Stdin to Badger

To create a CLI to stream stdin into Badger, use `badgerutils.WriteStdin`. It takes a function `lineToKeyed` as a parameter, which converts a `string` into a struct that implements the `Keyed` interface.

#### Example

```Go
// examples/writer.go
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
