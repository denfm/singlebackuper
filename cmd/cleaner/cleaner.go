package main

import (
	"errors"
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/denfm/singlebackuper/internal/app/service/cleaner"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

func main() {
	config := cfg.NewConfig()
	var clPaths []cleaner.CleanPath

	if config.TargetPath != "" {
		pt := strings.TrimRight(config.TargetPath, "/") + "/"
		clPaths = append(clPaths, cleaner.CleanPath{
			Path:     pt,
			IsRemote: false,
		})
	}

	if config.Remote.SshHost != "" && config.Remote.Path != "" {
		pr := strings.TrimRight(config.Remote.Path, "/") + "/"
		clPaths = append(clPaths, cleaner.CleanPath{
			Path:     pr,
			IsRemote: true,
		})
	}

	if len(clPaths) == 0 {
		logrus.Error(errors.New(`no data to clear`))
		os.Exit(1)
	}

	timeLocation, err := time.LoadLocation(config.TimeZone)

	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	tm := time.Now().In(timeLocation).AddDate(0, 0, config.Rotation)

	cl := cleaner.NewCleaner(config, clPaths, tm)
	err = cl.Clean()

	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	os.Exit(0)
}
