package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/openware/kaigara/types"
)

func delCmd() error {
	ss, err := loadStorageService()
	if err != nil {
		return fmt.Errorf("storage service init failed: %s", err)
	}

	if len(os.Args[2:]) == 0 {
		return fmt.Errorf("not enough arguments, please pass the deletion pattern")
	}

	return kaidelRun(ss, os.Args[2])
}

func kaidelRun(ss types.Storage, entryPattern string) error {
	patternValues := strings.Split(entryPattern, ".")
	if len(patternValues) != 3 {
		return fmt.Errorf("string '%s' doesn't match pattern 'app.scope.var'", entryPattern)
	}

	var appNames []string
	if appName := patternValues[0]; appName == "all" {
		var err error
		if appNames, err = ss.ListAppNames(); err != nil {
			return err
		}
	} else {
		appNames = []string{appName}
	}

	var scopes []string
	if scope := patternValues[1]; scope == "all" {
		scopes = []string{
			"public",
			"private",
			"secret",
		}
	} else {
		scopes = []string{scope}
	}

	varName := patternValues[2]

	for _, appName := range appNames {
		for _, scope := range scopes {
			if err := ss.Read(appName, scope); err != nil {
				return err
			}

			if varName == "all" {
				entries, err := ss.ListEntries(appName, scope)
				if err != nil {
					return err
				}

				for _, entry := range entries {
					if err := ss.DeleteEntry(appName, scope, entry); err != nil {
						return err
					}

					log.Printf("INF: deleted %s.%s.%s\n", appName, scope, entry)
				}
			} else {
				if err := ss.DeleteEntry(appName, scope, varName); err != nil {
					return err
				}

				log.Printf("INF: deleted %s.%s.%s\n", appName, scope, varName)
			}

			if err := ss.Write(appName, scope); err != nil {
				return err
			}
		}
	}

	return nil
}
