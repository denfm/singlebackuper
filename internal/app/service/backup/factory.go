package backup

import (
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/denfm/singlebackuper/internal/app/service"
)

func FactoryBackupModule(name string, config *cfg.Config) service.BackupModule {
	switch name {
	case "mongodb":
		return NewMongodbBackupModule(config)

	case "mysqldb":
		return NewMysqlBackupModule(config)

	case "files":
		return NewFilesBackupModule(config)
	}

	return NewNullBackupModule(config)
}
