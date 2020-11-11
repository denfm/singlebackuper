package backup

import (
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestFilesBackupModule_Create(t *testing.T) {
	assert.IsType(t, FactoryBackupModule("files", &cfg.Config{}), &FilesBackupModule{})
}

func TestFilesBackupModule_Backup(t *testing.T) {
	tmpPath := os.TempDir() + "/singlebackuper-backup-test-files/"
	_ = os.RemoveAll(tmpPath)

	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		err := os.MkdirAll(tmpPath+"tree", os.ModePerm)
		if err != nil {
			t.Errorf("can't create directory \"%s\". Err: %v", tmpPath, err)
		}
	}

	filesMap := map[string]string{
		"1.txt":      "Hello!",
		"tree/2.txt": "Hello tree!",
	}

	for k, v := range filesMap {
		err := ioutil.WriteFile(tmpPath+k, []byte(v), 0644)

		if err != nil {
			t.Errorf("can create file \"%s\". Err: %v", k, err)
		}
	}

	logrus.SetOutput(ioutil.Discard)

	config := &cfg.Config{}
	config.TmpPath = os.TempDir()
	config.TargetPath = os.TempDir()
	config.Files.Path = tmpPath
	config.Files.Prefix = "test"

	module := FactoryBackupModule("files", config)
	res := module.Backup()

	assert.Nil(t, res.Err)
}
