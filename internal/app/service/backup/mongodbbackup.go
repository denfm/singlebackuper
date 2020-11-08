package backup

import (
	"github.com/denfm/singlebackuper/internal/app/cfg"
	"github.com/denfm/singlebackuper/internal/app/service"
	"github.com/denfm/singlebackuper/internal/app/service/command"
	"github.com/denfm/singlebackuper/internal/app/service/state"
	"strconv"
)

type MongodbBackupModule struct {
	config *cfg.Config
	st     *state.State
}

// https://www.mongodb.com/try/download/database-tools
func (g *MongodbBackupModule) Backup() *service.BackupModuleResult {
	res := &service.BackupModuleResult{}
	prepareData := GetPrepareData(g.config.Mongo.Prefix, g.config)
	err := CreateDirsByPrepareData(prepareData)

	if err != nil {
		res.Err = err
		return res
	}

	defer CleanTemp(prepareData)
	defer g.st.Clear()

	cmd := command.CreateNewCommand(g.config.Mongo.DumpBin, []string{
		"--gzip",
		"--quiet",
	})

	if g.config.Mongo.Uri != "" {
		cmd.Add2ArgAsSolo("--uri", g.config.Mongo.Uri)
	} else {
		cmd.Add2Arg("--host", g.config.Mongo.Host)
		cmd.Add2Arg("--port", strconv.Itoa(g.config.Mongo.Port))

		if g.config.Mongo.User != "" {
			cmd.Add2ArgAsSolo("--username", g.config.Mongo.User)
		}

		if g.config.Mongo.Password != "" {
			cmd.Add2ArgAsSolo("--password", g.config.Mongo.Password)
		}

		if g.config.Mongo.Database != "" {
			cmd.Add2ArgAsSolo("--authenticationDatabase", g.config.Mongo.Database)
			cmd.Add2ArgAsSolo("--db", g.config.Mongo.Database)
		}

		if g.config.Mongo.AuthMechanism != "" {
			cmd.Add2ArgAsSolo("--authenticationMechanism", g.config.Mongo.AuthMechanism)
		}
	}

	cmd.Add2ArgAsSolo("--archive", prepareData.TmpArchivePath)
	err = cmd.Run(CmdMongoDump)

	if err != nil {
		res.Err = err
		return res
	}

	err = MoveArchive(prepareData, g.config)

	if err != nil {
		res.Err = err
		return res
	}

	SuccessFinishResult(res, prepareData, g.config)
	return res
}

func NewMongodbBackupModule(config *cfg.Config) *MongodbBackupModule {
	st := state.NewState(GetStateUniqueName(config.Mongo.Prefix, config), config.TmpPath)

	return &MongodbBackupModule{
		config: config,
		st:     st,
	}
}
