package backup

import (
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/denfm/singlebackuper/internal/app/service"
)

func FactoryBackupModule(name string, config *cfg.Config) service.BackupModule {
	switch name {
	case "mongodb":
		return NewMongodbBackupModule(config)
		break

	case "mysqldb":
		return NewMysqlBackupModule(config)
		break
	}

	return NewNullBackupModule(config)
}
