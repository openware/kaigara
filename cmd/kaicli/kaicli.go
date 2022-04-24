package main

import (
	"os"
	"path"

	"github.com/openware/pkg/kli"

	"github.com/openware/kaigara/pkg/config"
)

var KaiHome = "."
var DefaultConfFile = "kaiconf.yaml"
var conf *config.KaigaraConfig
var Version = "1.0.0"
var SecretsPath = "outputs.yaml"

func main() {
	var err error

	if kaiHome := os.Getenv("KAIGARA_HOME"); kaiHome != "" {
		KaiHome = kaiHome
	}

	SecretsPath = path.Join(KaiHome, SecretsPath)
	confPath := path.Join(KaiHome, DefaultConfFile)
	if _, err := os.Stat(confPath); err == nil {
		config.ConfPath = confPath
	}

	if conf, err = config.NewKaigaraConfig(); err != nil {
		panic(err)
	}

	cli := kli.NewCli("kai", "Kaigara CLI tool for managing secrets", Version)

	kaidump := cli.NewSubCommand("dump", "Get dump of all secrets").Action(kaidumpCmd)
	kaidump.StringFlag("f", "Outputs file path to save secrets", &SecretsPath)
	applyCommonFlags(kaidump)

	kaidel := cli.NewSubCommand("del", "Delete entry by key, app names and scopes").Action(kaidelCmd)
	kaidel.StringFlag("k", "Entry name to delete", &keyName)
	applyCommonFlags(kaidel)

	kaienv := cli.NewSubCommand("env", "Kaigara version of printenv CLI").Action(kaienvCmd)
	kaienvArgs = kaienv.OtherArgs()
	applyCommonFlags(kaienv)

	kaisave := cli.NewSubCommand("save", "Save secrets from file to storage").Action(kaisaveCmd)
	kaisave.StringFlag("f", "Input file to save secrets from", &SecretsPath)
	applyCommonFlags(kaisave)

	kaitail := cli.NewSubCommand("tail", "Get logstream of Redis channel").Action(kaitailCmd)
	kaitail.StringFlag("ch", "Redis channel pattern to subscribe", &redisChannel)
	kaitail.BoolFlag("sn", "Show Redis channel name", &showName)
	applyCommonFlags(kaitail)

	if err := cli.Run(); err != nil {
		panic(err)
	}
}

func applyCommonFlags(kaicmd *kli.Command) {
	kaicmd.StringFlag("a", "Set app names", &conf.AppNames)
	kaicmd.StringFlag("s", "Set scopes", &conf.Scopes)
}
