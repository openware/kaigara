package config

import (
	"log"
	"regexp"
	"strings"
)

// Config is the interface definition of generic config storage
type Config interface {
	ListEntries() map[string]interface{}
}

type Env struct {
	Vars  []string
	Files map[string]*File
}

type File struct {
	Path    string
	Content string
}

var kfile = regexp.MustCompile("(?i)^KFILE_(.*)_(PATH|CONTENT)$")

func BuildCmdEnv(cnf Config, currentEnv []string) *Env {
	env := &Env{
		Vars:  []string{},
		Files: map[string]*File{},
	}

	for _, v := range currentEnv {
		if !strings.HasPrefix(v, "KAIGARA_") {
			env.Vars = append(env.Vars, v)
		}
	}
	if cnf == nil {
		return env
	}
	for k, v := range cnf.ListEntries() {
		m := kfile.FindStringSubmatch(k)
		if m == nil {
			env.Vars = append(env.Vars, strings.ToUpper(k)+"="+v.(string))
			continue
		}
		name := strings.ToUpper(m[1])
		suffix := strings.ToUpper(m[2])

		f, ok := env.Files[name]
		if !ok {
			f = &File{}
			env.Files[name] = f
		}
		switch suffix {
		case "PATH":
			f.Path = v.(string)
		case "CONTENT":
			f.Content = v.(string)
		default:
			log.Printf("ERROR: Unexpected prefix in config key: %s", k)
		}
	}
	return env
}
