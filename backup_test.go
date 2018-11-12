package badgerutils

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/stretchr/testify/require"
)

func restoreBackup(dir, backupFilePath string) error {
	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	db, err := badger.Open(opts)
	if err != nil {
		return err
	}
	defer db.Close()
	file, err := os.Open(backupFilePath)
	if err != nil {
		return err
	}
	db.Load(file)
	return nil
}

func TestBackup(t *testing.T) {
	dir, err := os.Getwd()
	require.Nil(t, err)
	tmpDir, err := ioutil.TempDir(dir, "temp")
	require.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := path.Join(tmpDir, "db")
	backupPath := path.Join(tmpDir, "path", "to", "backup")
	backupName := "db.bak"

	reader := strings.NewReader(`field11,field12,field13
field21,field22,field23
field31,field32,field33`)
	err = WriteStream(reader, dbPath, 2, csvToSampleRecord)
	require.Nil(t, err)

	backupFilePath := path.Join(backupPath, backupName)
	_, err = CreateBackup(dbPath, backupPath, backupName)
	require.Nil(t, err)
	require.FileExists(t, backupFilePath)

	restoredDBPath := path.Join(tmpDir, "restored_db")
	err = restoreBackup(restoredDBPath, backupFilePath)
	require.Nil(t, err)

	restoredSampleRecords, err := readDB(restoredDBPath)
	require.Nil(t, err)
	require.Equal(t, 3, len(restoredSampleRecords))
	require.EqualValues(t, restoredSampleRecords[0], sampleRecord{"field11", "field12", "field13"})
	require.EqualValues(t, restoredSampleRecords[1], sampleRecord{"field21", "field22", "field23"})
	require.EqualValues(t, restoredSampleRecords[2], sampleRecord{"field31", "field32", "field33"})
}
