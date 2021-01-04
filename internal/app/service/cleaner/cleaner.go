package cleaner

import (
	"fmt"
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/denfm/singlebackuper/internal/app/service"
	"github.com/denfm/singlebackuper/internal/app/service/command"
	"github.com/denfm/singlebackuper/internal/app/service/state"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CleanPath struct {
	Path     string
	IsRemote bool
}

type Cleaner struct {
	config *cfg.Config
	paths  []CleanPath
	until  time.Time
	st     *state.State
}

func NewCleaner(config *cfg.Config, paths []CleanPath, until time.Time) *Cleaner {
	st := state.NewState("cleaner", config.TmpPath)
	return &Cleaner{config, paths, until, st}
}

func (c *Cleaner) Clean() error {
	if c.st.GetStateData().IsBusy {
		return fmt.Errorf(`the process "cleaner" has already started at %s and has not yet been completed`,
			c.st.GetStateData().DateTimeLabel)
	}

	defer c.st.Clear()

	utLabel := c.until.Format("2006-01-02 15:04:05")
	logrus.Infof(`Delete files with a creation date older than "%s".`, utLabel)

	if !c.config.RotationEnabled {
		logrus.Warn(`Real file rotation is disabled in the config. Only logs are output (simulation).`)
	}

	for k, p := range c.paths {
		if p.IsRemote {
			err := c.cleanRemote(k)
			if err != nil {
				return err
			}
		} else {
			err := c.cleanLocal(k)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Cleaner) cleanLocal(k int) error {
	cleanPath := c.paths[k].Path

	logrus.Infof(`Work local dir "%s".`, cleanPath)

	if !service.HasDir(cleanPath) {
		return nil
	}

	timeLocation, err := time.LoadLocation(c.config.TimeZone)

	if err != nil {
		return err
	}

	var dirs []string

	err = filepath.Walk(cleanPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if f.IsDir() {
			dirs = append(dirs, path)
			return nil
		}

		ft := f.ModTime().In(timeLocation)
		ftLabel := ft.Format("2006-01-02 15:04:05")

		if ft.Unix() > c.until.Unix() {
			logrus.Debugf(`Local file "%s" time create as "%s". Skip.`, path, ftLabel)
			return nil
		}

		logrus.Infof(`Remove local file "%s".`, path)

		if c.config.RotationEnabled {
			err = os.Remove(path)

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	if len(dirs) > 0 {
		for _, dir := range dirs {
			files, err := ioutil.ReadDir(dir)

			if err != nil {
				return err
			}

			if len(files) == 0 {
				logrus.Infof(`Remove empty local dir "%s".`, dir)
				if c.config.RotationEnabled {
					err = os.RemoveAll(dir)

					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (c *Cleaner) cleanRemote(k int) error {
	remotePath := c.config.Remote.Path + "/" + c.paths[k].Path

	err := command.SftpCommand(c.config, func(sftpClient *sftp.Client) error {
		logrus.Infof(`Work remote dir "%s".`, remotePath)

		err, dirs := c.remoteRecursiveList(remotePath, sftpClient)

		if err != nil {
			return err
		}

		for _, dir := range dirs {
			files, err := sftpClient.ReadDir(dir)

			if err != nil {
				return err
			}

			if len(files) == 0 {
				logrus.Infof(`Remove empty remote dir "%s".`, dir)
				if c.config.RotationEnabled {
					err = sftpClient.RemoveDirectory(dir)

					if err != nil {
						return err
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Cleaner) remoteRecursiveList(dir string, sftpClient *sftp.Client) (error, []string) {
	dir = strings.TrimRight(dir, "/") + "/"
	var dirs []string
	logrus.Debugf(`Listing remote dir "%s".`, dir)

	files, err := sftpClient.ReadDir(dir)

	if err != nil && os.IsNotExist(err) {
		return nil, dirs
	} else if err != nil {
		return err, dirs
	}

	timeLocation, err := time.LoadLocation(c.config.TimeZone)

	if err != nil {
		return err, dirs
	}

	for _, f := range files {
		if f.IsDir() {
			childDir := dir + f.Name() + "/"
			dirs = append(dirs, childDir)
			err, childDirs := c.remoteRecursiveList(childDir, sftpClient)

			if len(childDirs) > 0 {
				for _, chd := range childDirs {
					dirs = append(dirs, chd)
				}
			}

			if err != nil {
				return err, dirs
			}

			continue
		}

		filePath := dir + f.Name()

		ft := f.ModTime().In(timeLocation)
		ftLabel := ft.Format("2006-01-02 15:04:05")

		if ft.Unix() > c.until.Unix() {
			logrus.Debugf(`Remote file "%s" time create as "%s". Skip.`, filePath, ftLabel)
			continue
		}

		logrus.Infof(`Remove remote file "%s".`, filePath)
		if c.config.RotationEnabled {
			err := sftpClient.Remove(filePath)

			if err != nil {
				return err, dirs
			}
		}
	}

	return nil, dirs
}
