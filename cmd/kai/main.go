package main

import (
	"log"
	"os"

	"github.com/openware/pkg/kli"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/storage"
	"github.com/openware/kaigara/types"
)

var KaiConfPath = "./kaiconf.yaml"
var conf *config.KaigaraConfig
var Version = "dev"
var SecretsPath = "outputs.yaml"

func main() {
	var err error

	if kaiConfPath := os.Getenv("KAICONFIG"); kaiConfPath != "" {
		KaiConfPath = kaiConfPath
	}

	if _, err := os.Stat(KaiConfPath); err == nil {
		config.ConfPath = KaiConfPath
	}

	if conf, err = config.NewKaigaraConfig(); err != nil {
		panic(err)
	}

	cli := kli.NewCli("kai", "Kaigara CLI tool for managing secrets", Version)

	dump := cli.NewSubCommand("dump", "Get dump of all secrets").Action(dumpCmd)
	dump.StringFlag("o", "Outputs file path to save secrets", &SecretsPath)
	applyCommonFlags(dump)

	del := cli.NewSubCommand("del", "Delete entry by pattern 'app.scope.var'").Action(delCmd)
	del.StringFlag("d", "Set deployment id", &conf.DeploymentID)

	env := cli.NewSubCommand("env", "Kaigara version of printenv CLI").Action(envCmd)
	applyCommonFlags(env)

	save := cli.NewSubCommand("save", "Save secrets from file to storage").Action(saveCmd)
	save.StringFlag("f", "Input file to save secrets from", &SecretsPath)
	applyCommonFlags(save)

	if err := cli.Run(); err != nil {
		log.Fatal(err)
	}
}

func loadStorageService() (types.Storage, error) {
	if ss, err := storage.GetStorageService(conf); err != nil {
		return nil, err
	} else {
		return ss, nil
	}
}

func applyCommonFlags(cmd *kli.Command) {
	cmd.StringFlag("a", "Set app names", &conf.AppNames)
	cmd.StringFlag("s", "Set scopes", &conf.Scopes)
	cmd.StringFlag("d", "Set deployment id", &conf.DeploymentID)
}
