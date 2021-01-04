package backup

import (
	"fmt"
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/denfm/singlebackuper/internal/app/lib"
	"github.com/denfm/singlebackuper/internal/app/service"
	"github.com/denfm/singlebackuper/internal/app/service/command"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

const (
	CmdMongoDump = 0x82
	CmdSftp      = 0x8C
	//CmdMysqlDump = 0x96
)

type PrepareData struct {
	TimeNow           time.Time
	DateLabel         string
	BackupName        string
	LocalPath         string
	LocalArchivePath  string
	RemotePath        string
	RemoteArchivePath string
	TmpPath           string
	TmpArchivePath    string
}

func GetStateUniqueName(component string, config *cfg.Config) string {
	return fmt.Sprintf("singlebackuper_state_%s_%s_%d.tmp", config.Name, component, GetCurrentUnixTime(config.TimeZone))
}

func GetCurrentTime(timeZone string) time.Time {
	timeLocation, err := time.LoadLocation(timeZone)

	if err != nil {
		log.Fatal(err)
	}

	tm := time.Now().In(timeLocation)
	return tm
}

func GetCurrentUnixTime(timeZone string) int64 {
	return GetCurrentTime(timeZone).Unix()
}

func GetPrepareData(prefix string, config *cfg.Config) *PrepareData {
	var ar string

	switch prefix {
	case config.Mysql.Prefix:
		ar = ".sql.tgz"
		break
	default:
		ar = ".tar.gz"
	}

	timeNow := GetCurrentTime(config.TimeZone)

	var tStringMonth, tStringDay, tStringHour, tStringMinute, tStringSecond string

	if timeNow.Month() < 10 {
		tStringMonth = fmt.Sprintf(`0%d`, timeNow.Month())
	} else {
		tStringMonth = fmt.Sprintf(`%d`, timeNow.Month())
	}

	if timeNow.Day() < 10 {
		tStringDay = fmt.Sprintf(`0%d`, timeNow.Day())
	} else {
		tStringDay = fmt.Sprintf(`%d`, timeNow.Day())
	}

	if timeNow.Hour() < 10 {
		tStringHour = fmt.Sprintf(`0%d`, timeNow.Hour())
	} else {
		tStringHour = fmt.Sprintf(`%d`, timeNow.Hour())
	}

	if timeNow.Minute() < 10 {
		tStringMinute = fmt.Sprintf(`0%d`, timeNow.Minute())
	} else {
		tStringMinute = fmt.Sprintf(`%d`, timeNow.Minute())
	}

	if timeNow.Second() < 10 {
		tStringSecond = fmt.Sprintf(`0%d`, timeNow.Second())
	} else {
		tStringSecond = fmt.Sprintf(`%d`, timeNow.Second())
	}

	dateLabel := fmt.Sprintf("%d-%s-%s_%s-%s-%s", timeNow.Year(), tStringMonth, tStringDay,
		tStringHour, tStringMinute, tStringSecond)
	dateLabel2Path := fmt.Sprintf("%d%s%s", timeNow.Year(), tStringMonth, tStringDay)
	backupName := prefix + dateLabel

	var localPath, remotePath, tmpPath, localArchivePath, remoteArchivePath, tmpArchivePath string

	if config.TargetPath != "" {
		localPath = strings.TrimRight(config.TargetPath, "/") + "/" + dateLabel2Path + "/"
		localArchivePath = fmt.Sprintf("%s%s%s", localPath, backupName, ar)
	}

	if config.Remote.Path != "" && config.Remote.SshHost != "" && config.Remote.SshUser != "" {
		remotePath = strings.TrimRight(config.Remote.Path, "/") + "/" + dateLabel2Path + "/"
		remoteArchivePath = fmt.Sprintf("%s%s%s", remotePath, backupName, ar)
	}

	tmpPath = strings.TrimRight(config.TmpPath, "/") + "/singlebackuper/" + prefix + "/" + dateLabel2Path + "/"
	tmpArchivePath = fmt.Sprintf("%s%s%s", tmpPath, backupName, ar)

	return &PrepareData{
		timeNow,
		dateLabel,
		backupName,
		localPath,
		localArchivePath,
		remotePath,
		remoteArchivePath,
		tmpPath,
		tmpArchivePath,
	}
}

func DefArchiveFileSize(res *service.BackupModuleResult, p *PrepareData) {
	err, size := GetFileSize(p.TmpArchivePath)

	if err != nil {
		res.SizeMb = 0
		res.SizeMbLabel = "ERROR!"
		logrus.Errorf("GetFileSize error: %v", err)
	} else {
		res.SizeMb = float64(size / (1024 * 1024))
		res.SizeMbLabel = fmt.Sprintf("%.1fMB", res.SizeMb)
	}
}

func MoveArchive(p *PrepareData, config *cfg.Config) error {
	if p.LocalArchivePath != "" {
		err := lib.MoveFile(p.TmpArchivePath, p.LocalArchivePath)
		if err != nil {
			return err
		}
	}

	if p.RemoteArchivePath != "" {
		var sourceArchivePath string

		if p.LocalArchivePath != "" {
			sourceArchivePath = p.LocalArchivePath
		} else {
			sourceArchivePath = p.TmpArchivePath
		}

		logrus.Infof("Start upload archive to remote server \"%s\".", config.Remote.SshHost)

		err := command.SftpCommand(config, func(sftpClient *sftp.Client) error {
			err := sftpClient.MkdirAll(p.RemotePath)

			if err != nil {
				return fmt.Errorf("mkdir error: %v", err)
			}

			rmFile, err := sftpClient.Create(p.RemoteArchivePath)
			if err != nil {
				return err
			}
			defer rmFile.Close()

			srcFile, err := os.Open(sourceArchivePath)
			if err != nil {
				return err
			}

			_, err = io.Copy(rmFile, srcFile)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("errors: %v. CmdCode: %d", err, CmdSftp)
		}
	}

	return nil
}

func SuccessFinishResult(res *service.BackupModuleResult, p *PrepareData, config *cfg.Config) {
	res.ArchivePath = p.LocalArchivePath
	res.RemoteArchivePath = p.RemoteArchivePath

	rLocalPath := res.ArchivePath
	rRemotePath := res.RemoteArchivePath

	if rLocalPath == "" {
		rLocalPath = "none"
	}

	if rRemotePath == "" {
		rRemotePath = "none"
	}

	duration := GetCurrentTime(config.TimeZone).Sub(p.TimeNow)
	res.DurationSeconds = duration.Seconds()
	res.DurationLabel = fmt.Sprintf("%.1f", duration.Seconds())

	logrus.Infof("Archive size: %s", res.SizeMbLabel)
	logrus.Infof("Path: %s", rLocalPath)
	logrus.Infof("Remote path: %s", rRemotePath)
	logrus.Infof("Elapsed time: %s seconds", res.DurationLabel)
	logrus.Info("Backup success!")
}

func GetFileSize(pathFile string) (error, int64) {
	fi, err := os.Stat(pathFile)
	if err != nil {
		return err, 0
	}

	return nil, fi.Size()
}

func CreateDirs(dirs []string) error {
	for i := range dirs {
		if _, err := os.Stat(dirs[i]); os.IsNotExist(err) {
			err := os.MkdirAll(dirs[i], os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func CreateDirsByPrepareData(p *PrepareData) error {
	localDirs := []string{p.TmpPath}

	if p.LocalPath != "" {
		localDirs = append(localDirs, p.LocalPath)
	}

	err := CreateDirs(localDirs)

	if err != nil {
		return err
	}

	return nil
}

func CleanTemp(p *PrepareData) {
	if _, err := os.Stat(p.TmpPath); !os.IsNotExist(err) {
		err := os.RemoveAll(p.TmpPath)

		if err != nil {
			logrus.Error(err)
		}
	}
}
