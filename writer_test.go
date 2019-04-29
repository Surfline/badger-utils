package badgerutils

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteStream(t *testing.T) {
	dir, err := os.Getwd()
	require.Nil(t, err)
	tmpDir, err := ioutil.TempDir(dir, "temp")
	require.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := path.Join(tmpDir, "path", "to", "db")

	reader := strings.NewReader(`key1:value1
key2:value2
key3:value3`)
	err = WriteStream(reader, dbPath, 2, csvToKeyValue)
	require.Nil(t, err)

	writtenSampleRecords, err := readDB(dbPath)
	require.Nil(t, err)
	require.Equal(t, 3, len(writtenSampleRecords))
	require.EqualValues(t, writtenSampleRecords[0], sampleRecord{
		Key:   "key1",
		Value: "value1",
	})
	require.EqualValues(t, writtenSampleRecords[1], sampleRecord{
		Key:   "key2",
		Value: "value2",
	})
	require.EqualValues(t, writtenSampleRecords[2], sampleRecord{
		Key:   "key3",
		Value: "value3",
	})
}
