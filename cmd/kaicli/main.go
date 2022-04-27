package main

import (
	"os"

	"github.com/openware/pkg/kli"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/storage"
	"github.com/openware/kaigara/types"
)

var KaiConfPath = "./kaiconf.yaml"
var conf *config.KaigaraConfig
var ss types.Storage
var Version = "1.0.0"
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

	if ss, err = storage.GetStorageService(conf); err != nil {
		panic(err)
	}

	cli := kli.NewCli("kai", "Kaigara CLI tool for managing secrets", Version)

	dump := cli.NewSubCommand("dump", "Get dump of all secrets").Action(dumpCmd)
	dump.StringFlag("o", "Outputs file path to save secrets", &SecretsPath)
	applyCommonFlags(dump)

	del := cli.NewSubCommand("del", "Delete entry by pattern 'app.scope.var'")
	del.Action(delCmd(del))
	del.StringFlag("d", "Set deployment id", &conf.DeploymentID)

	env := cli.NewSubCommand("env", "Kaigara version of printenv CLI")
	env.Action(envCmd(env))
	applyCommonFlags(env)

	save := cli.NewSubCommand("save", "Save secrets from file to storage").Action(saveCmd)
	save.StringFlag("f", "Input file to save secrets from", &SecretsPath)
	applyCommonFlags(save)

	tail := cli.NewSubCommand("tail", "Get logstream of Redis channel").Action(tailCmd)
	tail.StringFlag("ch", "Redis channel pattern to subscribe", &redisChannel)
	tail.BoolFlag("sn", "Show Redis channel name", &showName)

	if err := cli.Run(); err != nil {
		panic(err)
	}
}

func applyCommonFlags(cmd *kli.Command) {
	cmd.StringFlag("a", "Set app names", &conf.AppNames)
	cmd.StringFlag("s", "Set scopes", &conf.Scopes)
	cmd.StringFlag("d", "Set deployment id", &conf.DeploymentID)
}
