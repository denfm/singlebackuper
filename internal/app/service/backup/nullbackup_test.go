package backup

import (
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNullBackupModule_Create(t *testing.T) {
	assert.IsType(t, FactoryBackupModule("null", &cfg.Config{}), &NullBackupModule{})
}

func TestNullBackupModule_Backup(t *testing.T) {
	module := FactoryBackupModule("null", &cfg.Config{})
	res := module.Backup()

	assert.EqualError(t, res.Err, "nothing to backup")
}
