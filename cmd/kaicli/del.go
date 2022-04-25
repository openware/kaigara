package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/openware/pkg/kli"
)

func delCmd(cmd *kli.Command) func() error {
	return func() error {
		args := cmd.OtherArgs()
		if len(args) == 0 {
			return fmt.Errorf("not enough arguments")
		}

		entryPattern := args[0]
		return kaidelRun(entryPattern)
	}
}

func kaidelRun(entryPattern string) error {
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
