package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var buildVersion string

const delimiter = "="

type Config struct {
	port   int
	appUrl string
	path   string
	name   string
}

func LoadConfig(args []string) *Config {
	c := &Config{port: 5000, name: "lurch"}
	c.setPath("workdir")
	parseArgs(args, func(arg, value string) {
		switch arg {
		case "-p", "--port":
			c.port, _ = strconv.Atoi(value)
		case "-t", "--path":
			c.setPath(value)
		case "-a", "--app-url":
			c.appUrl = value
		case "-n", "--name":
			c.name = value
		case "-h", "--help":
			fmt.Printf("Usage: lurch [options]\nOptions:\n\t-h, --help\t\t\tprint this help\n\t-v, --version\t\t\tprint version\n\t-t, --path [PATH]\t\tabsolute path to work dir\n\t-p, --port [PORT]\t\tsets port for listening\n\t-a, --app-url [APP_URL]\t\tapplication url (if behind proxy)\n\t-n, --name [NAME]\t\tname of application to be displayed\n")
			os.Exit(0)
		case "-v", "--version":
			fmt.Printf("lurch %s\nhttps://github.com/tvrzna/lurch\n\nReleased under the MIT License.\n", c.GetVersion())
			os.Exit(0)
		}
	})
	return c
}

func (c *Config) setPath(value string) {
	if path, err := filepath.Abs(value); err != nil {
		log.Fatal("wrong path", err)
	} else {
		c.path = path
	}
}

func parseArgs(args []string, handler func(arg, nextArg string)) {
	for i, arg := range args {
		nextArg := ""
		if len(args) > i+1 {
			val := strings.TrimSpace(args[i+1])
			if !strings.HasPrefix(val, "-") {
				nextArg = val
			}
		}
		if strings.Contains(arg, delimiter) {
			nextArg = arg[strings.Index(arg, delimiter)+1:]
			arg = arg[0:strings.Index(arg, delimiter)]
			if (strings.HasPrefix(nextArg, "'") && strings.HasSuffix(nextArg, "'")) || (strings.HasPrefix(nextArg, "\"") && strings.HasSuffix(nextArg, "\"")) {
				nextArg = nextArg[1 : len(nextArg)-1]
			}
		}
		handler(arg, nextArg)
	}
}

func (c *Config) getAppUrl() string {
	return c.appUrl
}

func (c *Config) getServerUri() string {
	return "localhost:" + strconv.Itoa(c.port)
}

func (c *Config) GetVersion() string {
	if buildVersion == "" {
		return "develop"
	}
	return buildVersion
}
