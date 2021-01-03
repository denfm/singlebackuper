package main

import (
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

		if config.Files.Path != "" && config.Files.Prefix != "" {
			clPaths = append(clPaths, cleaner.CleanPath{
				Path:     pt + config.Files.Prefix,
				IsRemote: false,
			})
		}

		if config.Mysql.Prefix != "" {
			clPaths = append(clPaths, cleaner.CleanPath{
				Path:     pt + config.Mysql.Prefix,
				IsRemote: false,
			})
		}

		if config.Mongo.Prefix != "" {
			clPaths = append(clPaths, cleaner.CleanPath{
				Path:     pt + config.Mongo.Prefix,
				IsRemote: false,
			})
		}

		if config.Clickhouse.Prefix != "" {
			clPaths = append(clPaths, cleaner.CleanPath{
				Path:     pt + config.Clickhouse.Prefix,
				IsRemote: false,
			})
		}
	}

	if config.Remote.SshHost != "" && config.Remote.Path != "" {
		if config.Files.Path != "" && config.Files.Prefix != "" {
			clPaths = append(clPaths, cleaner.CleanPath{
				Path:     config.Files.Prefix,
				IsRemote: true,
			})
		}

		if config.Mysql.Prefix != "" {
			clPaths = append(clPaths, cleaner.CleanPath{
				Path:     config.Mysql.Prefix,
				IsRemote: true,
			})
		}

		if config.Mongo.Prefix != "" {
			clPaths = append(clPaths, cleaner.CleanPath{
				Path:     config.Mongo.Prefix,
				IsRemote: true,
			})
		}

		if config.Clickhouse.Prefix != "" {
			clPaths = append(clPaths, cleaner.CleanPath{
				Path:     config.Clickhouse.Prefix,
				IsRemote: true,
			})
		}
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
