package main

import (
	"os"

	"github.com/openware/pkg/kli"

	"github.com/openware/kaigara/pkg/config"
)

var KaiConfPath = "./kaiconf.yaml"
var conf *config.KaigaraConfig
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

	cli := kli.NewCli("kai", "Kaigara CLI tool for managing secrets", Version)

	kaidump := cli.NewSubCommand("dump", "Get dump of all secrets").Action(kaidumpCmd)
	kaidump.StringFlag("o", "Outputs file path to save secrets", &SecretsPath)
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
