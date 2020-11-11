package cfg

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"log"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", "/etc/singlebackuper/singlebackuper.toml", "Path to config singlebackuper.toml file")
}

type Mongodb struct {
	Uri           string `toml:"mongodb_uri"`
	Host          string `toml:"mongodb_host"`
	Port          int    `toml:"mongodb_port"`
	User          string `toml:"mongodb_user"`
	Password      string `toml:"mongodb_password"`
	Database      string `toml:"mongodb_database"`
	AuthMechanism string `toml:"mongodb_auth_mechanism"`
	DumpBin       string `toml:"mongodb_dump_bin"`
	Prefix        string `toml:"mongodb_prefix"`
}

type Remote struct {
	SshHost       string `toml:"remote_ssh_host"`
	SshUser       string `toml:"remote_ssh_user"`
	SshPort       int    `toml:"remote_ssh_port"`
	SshPassword   string `toml:"remote_ssh_password"`
	SshPrivateKey string `toml:"remote_ssh_private_key"`
	Path          string `toml:"remote_path"`
}

type Mysqldb struct {
	Uri      string `toml:"mysqldb_uri"`
	Host     string `toml:"mysqldb_host"`
	Port     int    `toml:"mysqldb_port"`
	User     string `toml:"mysqldb_user"`
	Password string `toml:"mysqldb_password"`
	Database string `toml:"mysqldb_database"`
	DumpBin  string `toml:"mysqldb_dump_bin"`
	Prefix   string `toml:"mysqldb_prefix"`
	Opt      string `toml:"mysqldb_opt"`
	Excludes string `toml:"mysqldb_excludes" `
}

type Files struct {
	Path         string `toml:"files_path"`
	Prefix       string `toml:"files_prefix"`
	ExcludesPath string `toml:"files_exclude_path"`
}

// wait: @feature/backup_clickhouse
//type Clickhouse struct {
//
//}

// wait: @feature/backup_redis
//type Redis struct {
//
//}

// wait: @feature/backup_sphinx
//type Sphinx struct {
//
//}

type Config struct {
	BindAddr   string  `toml:"bind_address"`
	LogLevel   string  `toml:"log_level"`
	Rotation   int     `toml:"rotation"`
	TmpPath    string  `toml:"tmp_path"`
	TargetPath string  `toml:"target_path"`
	TimeZone   string  `toml:"time_zone"`
	GzipBin    string  `toml:"gzip_bin"`
	Remote     Remote  `toml:"remote"`
	Mongo      Mongodb `toml:"mongodb"`
	Mysql      Mysqldb `toml:"mysqldb"`
	Files      Files   `toml:"files"`
}

func NewConfig() *Config {
	flag.Parse()

	config := &Config{
		// wait: @feature/api BindAddr:   "127.0.0.1:8628",
		// wait: @feature/rotation Rotation:   30,
		LogLevel:   "info",
		TmpPath:    "/tmp",
		TargetPath: "/tmp/backup",
		TimeZone:   "Europe/Moscow",
		Remote: Remote{
			SshPort: 22,
		},
		Mongo: Mongodb{
			Uri:           "mongodb://127.0.0.1:27017",
			Host:          "127.0.0.1",
			Port:          27017,
			AuthMechanism: "SCRAM-SHA-256",
			DumpBin:       "/usr/bin/mongodump",
			Prefix:        "mgdb",
		},
		Mysql: Mysqldb{
			Uri:      "mysql://127.0.0.1:3306",
			Host:     "127.0.0.1",
			Port:     3306,
			DumpBin:  "/usr/bin/mysqldump",
			Prefix:   "mysqldb",
			Opt:      "--opt --single-transaction --default-character-set=utf8mb4",
			Excludes: "information_schema,performance_schema",
		},
		Files: Files{
			Prefix: "backup",
		},
	}
	_, err := toml.DecodeFile(configPath, &config)

	if err != nil {
		log.Fatal(err)
	}

	logrusLogLevel, err := logrus.ParseLevel(config.LogLevel)

	if err != nil {
		logrusLogLevel = logrus.InfoLevel
	}

	logrus.SetLevel(logrusLogLevel)

	return config
}
