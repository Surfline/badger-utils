package badgerutils

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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

	_, err = CreateBackup(dbPath, backupPath, backupName)
	require.Nil(t, err)
	require.FileExists(t, path.Join(backupPath, backupName))
}
