package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/openware/kaigara/pkg/config"
	"github.com/openware/kaigara/pkg/storage"
	"github.com/openware/pkg/ika"
)

var conf = &config.KaigaraConfig{}

func main() {
	// Parse flags
	scopes := flag.String("s", "public,private,secret", "scopes list")
	appName := flag.String("a", "global", "app name")
	keyName := flag.String("k", "key1", "key name")

	flag.Parse()

	if *keyName == "" {
		panic("ERR: Key name is missing(please pass it via -k)")
	}

	if *appName == "" {
		panic("ERR: App name is missing(please pass it via -a)")
	}

	if *scopes == "" {
		panic("ERR: Scope list is missing(please pass it via -s)")
	}

	// Initialize and write to Vault stores for every component
	if err := ika.ReadConfig("", conf); err != nil {
		panic(err)
	}

	ss, err := storage.GetStorageService(conf)
	if err != nil {
		panic(err)
	}

	// Get the list of scopes by Splitting KAIGARA_SCOPES env
	scopesList := strings.Split(*scopes, ",")
	if len(scopesList) <= 0 {
		panic("Scope list is empty")
	}

	for _, scope := range scopesList {
		err := ss.Read(*appName, scope)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Deleting %s.%s.%s\n", *appName, scope, *keyName)
		err = ss.DeleteEntry(*appName, scope, *keyName)
		if err != nil {
			panic(err)
		}

		err = ss.Write(*appName, scope)
		if err != nil {
			panic(err)
		}
	}
}
