package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/openware/kaigara/pkg/storage"
)

var keyName = ""

func kaidelCmd() error {
	if keyName == "" {
		return fmt.Errorf("key name is missing (please pass it via -k)")
	}

	if conf.Scopes == "" {
		return fmt.Errorf("scopes string is empty")
	}

	if conf.AppNames == "" {
		return fmt.Errorf("app names string is empty")
	}

	ss, err := storage.GetStorageService(conf)
	if err != nil {
		return err
	}

	scopes := strings.Split(conf.Scopes, ",")
	appNames := strings.Split(conf.AppNames, ",")

	for _, appName := range appNames {
		for _, scope := range scopes {
			err := ss.Read(appName, scope)
			if err != nil {
				return err
			}

			if err = ss.DeleteEntry(appName, scope, keyName); err != nil {
				return err
			}

			if err = ss.Write(appName, scope); err != nil {
				return err
			}

			log.Printf("INF: deleted %s.%s.%s\n", appName, scope, keyName)
		}
	}

	return err
}
