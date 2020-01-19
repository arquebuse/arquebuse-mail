package sender

import (
	"encoding/json"
	"github.com/arquebuse/arquebuse-mail/pkg/configuration"
	"github.com/emersion/go-smtp"
	"github.com/segmentio/ksuid"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var outboundPath string
var spoolPath string
var failedPath string

type Try struct {
	Timestamp time.Time `json:"timestamp"`
	Result    string    `json:"result"`
}

type spoolMail struct {
	Server   string    `json:"server"`
	From     string    `json:"from"`
	To       string    `json:"to"`
	Data     string    `json:"data"`
	Received time.Time `json:"timestamp"`
	Tries    []Try     `json:"tries"`
	NextTry  time.Time `json:"nextTry"`
	Status   string    `json:"status"`
}

// Create necessary folders
func initDataStructure(dataPath string) {

	outboundPath = path.Join(dataPath, "outbound")
	spoolPath = path.Join(dataPath, "spool")
	failedPath = path.Join(dataPath, "failed")

	pathList := []string{outboundPath, spoolPath, failedPath}

	for _, path := range pathList {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err := os.MkdirAll(path, 0750)

			if err == nil {
				log.Printf("Sender - Created directory '%s'\n", path)
			} else {
				log.Fatalf("Sender - Failed to create directory '%s'. Error: %s\n", path, err.Error())
			}
		}
	}
}

// Start watching for incoming mail in spool
func Start(config *configuration.Config) {
	initDataStructure(config.DataPath)

	// Process files every seconds
	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})

	for {
		select {
		case <-ticker.C:
			err := filepath.Walk(spoolPath, func(filePath string, info os.FileInfo, err error) error {
				if !info.IsDir() && !strings.Contains(filePath, "index.json") {
					processFile(filePath)
				}

				return nil
			})
			if err != nil {
				log.Printf("Sender - Failed to walk folder '%s'. error: %s\n", spoolPath, err.Error())
			}
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

// Process a found JSON file
func processFile(filePath string) {

	mail := spoolMail{}

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Sender - Cannot read JSON file '%s'. Error: %s\n", filePath, err)
	}

	err = json.Unmarshal([]byte(file), &mail)
	if err != nil {
		log.Printf("Sender - Cannot extract data from JSON file '%s'. Error: %s\n", filePath, err)
	}

	if mail.NextTry.Before(time.Now()) {

		try := Try{
			Timestamp: time.Now(),
		}

		err = send(mail)
		if err != nil {
			log.Printf("Sender - Failed to send email. Error: %s\n", err.Error())
			try.Result = err.Error()
			if len(mail.Tries) < 4 {
				mail.NextTry = time.Now().Add(time.Duration(math.Pow(2, float64(len(mail.Tries)))) * time.Second)
				mail.Status = "RETRY"
			} else {
				os.Remove(filePath)
				filePath = path.Join(outboundPath, ksuid.New().String()+".json")
				mail.Status = "FAILED"
			}
		} else {
			log.Println("Sender - Email sent")
			os.Remove(filePath)
			filePath = path.Join(outboundPath, ksuid.New().String()+".json")
			try.Result = "OK"
			mail.Status = "SENT"
		}

		mail.Tries = append(mail.Tries, try)

		file, err = json.MarshalIndent(mail, "", " ")
		if err != nil {
			log.Printf("Sender - Failed to convert to JSON current mail. Error: %s\n", err.Error())
		}

		err = ioutil.WriteFile(filePath, file, 0644)
		if err != nil {
			log.Printf("Sender - Failed to update file in '%s'. Error: %s\n", filePath, err.Error())
		}
	}
}

// Send mail based on json file
func send(mail spoolMail) error {
	to := strings.Split(mail.To, ";")
	return smtp.SendMail(mail.Server, nil, mail.From, to, strings.NewReader(mail.Data))
}
