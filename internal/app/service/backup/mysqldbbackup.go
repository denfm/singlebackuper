package backup

import (
	"fmt"
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/denfm/singlebackuper/internal/app/service"
	"github.com/denfm/singlebackuper/internal/app/service/command"
	"github.com/denfm/singlebackuper/internal/app/service/state"
	"io/ioutil"
	"os"
	"strings"
)

type MysqlBackupModule struct {
	config *cfg.Config
	st     *state.State
}

// https://dev.mysql.com/doc/refman/8.0/en/mysqldump.html
func (g *MysqlBackupModule) Backup() *service.BackupModuleResult {
	res := &service.BackupModuleResult{}
	prepareData := GetPrepareData(g.config.Mysql.Prefix, g.config)
	err := CreateDirsByPrepareData(prepareData)

	if err != nil {
		res.Err = err
		return res
	}

	defer CleanTemp(prepareData)
	defer g.st.Clear()

	myConfPath := fmt.Sprintf("%s/singlebackuper-my-cnf-%s.conf", g.config.TmpPath, service.RandStringRunes(7))
	myConfBody := fmt.Sprintf("[client]\nhost=\"%s\"\nport=\"%d\"\n", g.config.Mysql.Host, g.config.Mysql.Port)

	defer func() {
		_ = os.Remove(myConfPath)
	}()

	if g.config.Mysql.User != "" {
		myConfBody += fmt.Sprintf("user=\"%s\"", g.config.Mysql.User)

		if g.config.Mysql.Password != "" {
			myConfBody += fmt.Sprintf("\npassword=\"%s\"", g.config.Mysql.Password)
		}

		myConfBody += "\n"
	}

	err = ioutil.WriteFile(myConfPath, []byte(myConfBody), 0644)

	if err != nil {
		res.Err = fmt.Errorf("can't create file \"%s\". Error: %v", myConfPath, err)
		return res
	}

	cmd := command.CreateNewCommand(g.config.Mysql.DumpBin, []string{})
	cmd.Add2Arg("--defaults-file="+myConfPath, "")

	if g.config.Mysql.Opt != "" {
		for _, optValue := range strings.Split(g.config.Mysql.Opt, " ") {
			cmd.Add2Arg(optValue, "")
		}
	}

	if g.config.Mysql.Database != "" {
		cmd.Add2Arg(g.config.Mysql.Database, "")
	} else {
		cmd.Add2Arg("--all-databases", "")
	}

	err = cmd.ToGzip(prepareData.TmpArchivePath)

	if err != nil {
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

func NewMysqlBackupModule(config *cfg.Config) *MysqlBackupModule {
	st := state.NewState(GetStateUniqueName(config.Mysql.Prefix, config), config.TmpPath)

	return &MysqlBackupModule{
		config: config,
		st:     st,
	}
}
