package configuration

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	MailVersion string
	Data   string `yaml:"data"`
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func SearchFile(fileName string) string {
	// Search Paths
	searchPaths := []string{
		"",
		"./",
		"./conf/",
		"/etc/arquebuse-mail/",
	}

	for _, path := range searchPaths {
		currentPath := path + fileName
		if fileExists(currentPath) {
			return currentPath
		}
	}

	return ""
}

func Load(configFile *string, configuration *Config) {
	// Default values
	configuration.Data = "./data"

	p := SearchFile(*configFile)
	if p != "" {
		c, err := ioutil.ReadFile(p)
		if err != nil {
			log.Printf("ERROR - Unable to read config file '%s'. Error: %s\n", p, err.Error())
		} else {
			err := yaml.Unmarshal(c, configuration)
			if err != nil {
				log.Printf("ERROR - Failed to parse config file '%s'. Error: %s\n", p, err.Error())
			} else {
				log.Printf("Successfully loaded config file '%s'\n", p)
			}
		}
	} else {
		log.Print("No config file found\n")
	}
}
