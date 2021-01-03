package cleaner

import (
	"errors"
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/denfm/singlebackuper/internal/app/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

type testData struct {
	config  *cfg.Config
	tmpPath string
	appPath string
}

func makeTestData(t *testing.T) testData {
	tmpPath := os.TempDir() + "/singlebackuper-cleaner-test-files/"
	_ = os.RemoveAll(tmpPath)

	appPath := tmpPath + "backups/"

	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		err := os.MkdirAll(appPath+"tree", os.ModePerm)
		if err != nil {
			t.Errorf("can't create directory \"%s\". Err: %v", appPath, err)
		}
	}

	filesMap := map[string]string{
		"1.txt":      "Hello!",
		"tree/2.txt": "Hello tree!",
	}

	for k, v := range filesMap {
		err := ioutil.WriteFile(appPath+k, []byte(v), 0644)

		if err != nil {
			t.Errorf("can create file \"%s\". Err: %v", k, err)
		}
	}

	logrus.SetOutput(ioutil.Discard)

	config := &cfg.Config{}
	config.TmpPath = tmpPath
	config.TimeZone = "Europe/Moscow"

	return testData{config, tmpPath, appPath}
}

func TestLocalCleanerByNowTime(t *testing.T) {
	tData := makeTestData(t)

	defer func() {
		_ = os.RemoveAll(tData.tmpPath)
	}()

	timeLocation, err := time.LoadLocation(tData.config.TimeZone)

	if err != nil {
		t.Error(err)
	}

	tm := time.Now().In(timeLocation)
	cc := []CleanPath{{tData.appPath, false}}
	c := NewCleaner(tData.config, cc, tm)

	assert.IsType(t, c, &Cleaner{})
	assert.NoError(t, c.Clean())
	assert.False(t, service.HasDir(tData.appPath+"tree"))
}

func TestLocalCleanerByFutureTime(t *testing.T) {
	tData := makeTestData(t)

	defer func() {
		_ = os.RemoveAll(tData.tmpPath)
	}()

	timeLocation, err := time.LoadLocation(tData.config.TimeZone)

	if err != nil {
		t.Error(err)
	}

	tm := time.Now().In(timeLocation).AddDate(0, 0, -1)
	cc := []CleanPath{{tData.appPath, false}}
	c := NewCleaner(tData.config, cc, tm)

	assert.IsType(t, c, &Cleaner{})
	assert.NoError(t, c.Clean())
	assert.NoError(t, checkExistTestFile(tData))
}

// Заполните данные SSH ниже для запуска теста

//func TestRemoteCleanerBySSH(t *testing.T) {
//	tmpPath := os.TempDir() + "/singlebackuper-cleaner-test-files/"
//	_ = os.RemoveAll(tmpPath)
//
//	err := os.MkdirAll(tmpPath, os.ModePerm)
//	if err != nil {
//		t.Errorf("can't create directory \"%s\". Err: %v", tmpPath, err)
//	}
//
//	logrus.SetOutput(ioutil.Discard)
//
//	config := &cfg.Config{}
//	config.TmpPath = tmpPath
//	config.TimeZone = "Europe/Moscow"
//	config.Remote = cfg.Remote{
//		SshHost:     "<host>",
//		SshUser:     "login",
//		SshPort:     22,
//		SshPassword: "pwd",
//		Path:        "/data/test",
//	}
//
//	timeLocation, err := time.LoadLocation(config.TimeZone)
//
//	if err != nil {
//		t.Error(err)
//	}
//
//	tm := time.Now().In(timeLocation)
//	cc := []CleanPath{{"backups", true}}
//	c := NewCleaner(config, cc, tm)
//
//	assert.IsType(t, c, &Cleaner{})
//	assert.NoError(t, c.Clean())
//}

func checkExistTestFile(tData testData) error {
	if _, err := os.Stat(tData.appPath + "1.txt"); os.IsNotExist(err) {
		return errors.New(`files should have remained, but they were deleted`)
	}

	return nil
}
