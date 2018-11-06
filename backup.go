package badgerutils

import (
	"log"
	"os"
	"path"
	"time"

	"github.com/dgraph-io/badger"
)

// CreateBackup creates a backup from a Badger database directory.
func CreateBackup(dir, backupPath, backupName string) (uint64, error) {
	var version uint64

	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	db, err := badger.Open(opts)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	if err := os.MkdirAll(backupPath, os.ModePerm); err != nil {
		return version, err
	}

	outputPath := path.Join(backupPath, backupName)

	file, err := os.Create(outputPath)
	if err != nil {
		return version, err
	}

	start := time.Now()

	version, err = db.Backup(file, 0)
	if err != nil {
		return version, err
	}

	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("Created backup in %v", elapsed)
	return version, nil
}
