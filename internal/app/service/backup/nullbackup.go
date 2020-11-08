package backup

import (
	"errors"
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/denfm/singlebackuper/internal/app/service"
)

type NullBackupModule struct {
	config *cfg.Config
}

func (g *NullBackupModule) Backup() *service.BackupModuleResult {
	return &service.BackupModuleResult{
		Err: errors.New("nothing to backup"),
	}
}

func NewNullBackupModule(config *cfg.Config) *NullBackupModule {
	return &NullBackupModule{
		config: config,
	}
}
