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

	reader := strings.NewReader(`field11,field12,field13
field21,field22,field23
field31,field32,field33`)
	err = WriteStream(reader, dbPath, 2, csvToSampleRecord)
	require.Nil(t, err)

	writtenSampleRecords, err := readDB(dbPath)
	require.Nil(t, err)
	require.Equal(t, 3, len(writtenSampleRecords))
	require.EqualValues(t, writtenSampleRecords[0], sampleRecord{"field11", "field12", "field13"})
	require.EqualValues(t, writtenSampleRecords[1], sampleRecord{"field21", "field22", "field23"})
	require.EqualValues(t, writtenSampleRecords[2], sampleRecord{"field31", "field32", "field33"})
}
