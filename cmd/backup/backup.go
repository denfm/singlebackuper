package main

import (
	"flag"
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/denfm/singlebackuper/internal/app/service/backup"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	moduleNameArg := flag.String("module", "null", "Name of backup module")
	flag.Parse()

	config := cfg.NewConfig()
	logrus.Infof("Running backup from module \"%s\"", *moduleNameArg)

	module := backup.FactoryBackupModule(*moduleNameArg, config)
	res := module.Backup()

	if res.Err != nil {
		logrus.Error(res.Err)
		os.Exit(1)
	}

	os.Exit(0)
}
