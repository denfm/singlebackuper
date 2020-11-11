package backup

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ClickHouse/clickhouse-go"
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/denfm/singlebackuper/internal/app/lib"
	"github.com/denfm/singlebackuper/internal/app/service"
	"github.com/denfm/singlebackuper/internal/app/service/state"
	"io"
	"os"
	"strings"
)

type ClickhouseBackupModule struct {
	config     *cfg.Config
	st         *state.State
	connection *sql.DB
}

func (g *ClickhouseBackupModule) Backup() *service.BackupModuleResult {
	res := &service.BackupModuleResult{}
	libPath := strings.TrimSpace(g.config.Clickhouse.LibPath)

	if !HasDir(libPath) {
		res.Err = fmt.Errorf("invalid clickhouse libPath \"%s\"", libPath)
		return res
	}

	if strings.TrimSpace(g.config.Clickhouse.Databases) == "" {
		res.Err = errors.New("you need to specify the database")
		return res
	}

	databases := strings.Split(strings.TrimSpace(g.config.Clickhouse.Databases), ",")

	prepareData := GetPrepareData(g.config.Clickhouse.Prefix, g.config)
	err := CreateDirsByPrepareData(prepareData)

	if err != nil {
		res.Err = err
		return res
	}

	defer CleanTemp(prepareData)
	defer g.st.Clear()

	fileToWrite, err := os.OpenFile(prepareData.TmpArchivePath, os.O_CREATE|os.O_RDWR, os.FileMode(0644))

	if err != nil {
		res.Err = err
		return res
	}

	defer fileToWrite.Close()

	g.connection, err = sql.Open("clickhouse", g.config.Clickhouse.Uri)
	if err != nil {
		res.Err = fmt.Errorf("unable to connect to clickhouse server. Err: %v", err)
		return res
	}

	defer g.connection.Close()

	if err := g.connection.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			res.Err = fmt.Errorf("clickhouse server exception. Err: %s", exception.Message)
		} else {
			res.Err = fmt.Errorf("clickhouse server error. Err: %v", err)
		}

		return res
	}

	for _, database := range databases {
		err = g.freezeDb(database)

		if err != nil {
			res.Err = err
			return res
		}
	}

	metaPath := fmt.Sprintf("%smetadata/", libPath)
	shadowPath := fmt.Sprintf("%sshadow/", libPath)

	if !HasDir(metaPath) {
		res.Err = fmt.Errorf("invalid clickhouse metaPath \"%s\"", metaPath)
		return res
	}

	if !HasDir(shadowPath) {
		res.Err = fmt.Errorf("invalid clickhouse shadowPath \"%s\"", shadowPath)
		return res
	}

	var buf bytes.Buffer
	err = lib.Compress([]string{metaPath, shadowPath}, &buf, []string{})

	if err != nil {
		res.Err = err
		return res
	}

	if _, err := io.Copy(fileToWrite, &buf); err != nil {
		res.Err = err
		return res
	}

	DefArchiveFileSize(res, prepareData)
	err = MoveArchive(prepareData, g.config)

	if err != nil {
		res.Err = err
		return res
	}

	SuccessFinishResult(res, prepareData, g.config)

	return res
}

func (g *ClickhouseBackupModule) freezeDb(database string) error {
	rows, err := g.connection.Query(fmt.Sprintf("show tables from %s", database))
	if err != nil {
		return fmt.Errorf("clickhouse server query error. Err: %v", err)
	}

	defer rows.Close()
	var tableNames []string

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("clickhouse server query error. Err: %v", err)
		}

		tableNames = append(tableNames, tableName)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("clickhouse server query error. Err: %v", err)
	}

	if len(tableNames) == 0 {
		return fmt.Errorf("clickhouse database is empty. Nothing backup")
	}

	for _, tableName := range tableNames {
		_, err = g.connection.Exec(fmt.Sprintf("alter table %s.%s freeze", database, tableName))

		if err != nil {
			return fmt.Errorf("clickhouse table freeze error. Err: %v", err)
		}
	}
	return nil
}

func NewClickhouseBackupModule(config *cfg.Config) *ClickhouseBackupModule {
	st := state.NewState(GetStateUniqueName(config.Clickhouse.Prefix, config), config.TmpPath)

	return &ClickhouseBackupModule{
		config: config,
		st:     st,
	}
}
