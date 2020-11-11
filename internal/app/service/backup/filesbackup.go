package backup

import (
	"bytes"
	"fmt"
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/denfm/singlebackuper/internal/app/lib"
	"github.com/denfm/singlebackuper/internal/app/service"
	"github.com/denfm/singlebackuper/internal/app/service/state"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

type FilesBackupModule struct {
	config *cfg.Config
	st     *state.State
}

func (g *FilesBackupModule) Backup() *service.BackupModuleResult {
	res := &service.BackupModuleResult{}
	path := strings.TrimSpace(g.config.Files.Path)

	if path == "" {
		res.Err = fmt.Errorf("indicate which one needs to be backed up")
		return res
	}

	if !HasDir(g.config.Files.Path) {
		res.Err = fmt.Errorf("the specified backup directory \"%s\" does not exist", path)
		return res
	}

	if path == "/" {
		res.Err = fmt.Errorf("bad path")
		return res
	}

	prepareData := GetPrepareData(g.config.Files.Prefix, g.config)
	err := CreateDirsByPrepareData(prepareData)

	if err != nil {
		res.Err = err
		return res
	}

	defer CleanTemp(prepareData)
	defer g.st.Clear()

	fileToWrite, err := os.OpenFile(prepareData.TmpArchivePath, os.O_CREATE|os.O_RDWR, os.FileMode(0644))

	if err != nil {
		res.Err = err
		return res
	}

	defer fileToWrite.Close()

	logrus.Infof("Start backup directory \"%s\". Symlink support: off.", path)

	var buf bytes.Buffer
	err = lib.Compress(path, &buf, strings.Split(g.config.Files.ExcludesPath, ","))

	if err != nil {
		res.Err = err
		return res
	}

	if _, err := io.Copy(fileToWrite, &buf); err != nil {
		res.Err = err
		return res
	}

	DefArchiveFileSize(res, prepareData)
	err = MoveArchive(prepareData, g.config)

	if err != nil {
		res.Err = err
		return res
	}

	SuccessFinishResult(res, prepareData, g.config)

	return res
}

func NewFilesBackupModule(config *cfg.Config) *FilesBackupModule {
	st := state.NewState(GetStateUniqueName(config.Files.Prefix, config), config.TmpPath)

	return &FilesBackupModule{
		config: config,
		st:     st,
	}
}
