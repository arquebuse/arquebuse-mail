package main

import (
	"flag"
	"github.com/arquebuse/arquebuse-mail/pkg/configuration"
	"github.com/arquebuse/arquebuse-mail/pkg/receiver"
	"github.com/arquebuse/arquebuse-mail/pkg/sender"
	"time"
)

var mailVersion string
var config configuration.Config

func init() {
	configFile := flag.String("conf", "application.yaml", "Config file to load (default application.yaml.")
	configuration.Load(configFile, &config)
	config.MailVersion = mailVersion
}

func main() {
	go receiver.Start()
	time.Sleep(2 * time.Second)
	sender.Send()
}
