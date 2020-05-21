package main

import (
	"flag"
	"github.com/arquebuse/arquebuse-mail/pkg/configuration"
	"github.com/arquebuse/arquebuse-mail/pkg/receiver"
	"github.com/arquebuse/arquebuse-mail/pkg/sender"
)

var config configuration.Config

func init() {
	configFile := flag.String("conf", "application.yaml", "Config file to load (default application.yaml.")
	configuration.Load(configFile, &config)
}

func main() {
	receiver.Start(&config)
	sender.Start(&config)
}
