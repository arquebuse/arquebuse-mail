package receiver

import (
	"encoding/json"
	"errors"
	"github.com/arquebuse/arquebuse-mail/pkg/configuration"
	"github.com/emersion/go-smtp"
	"github.com/segmentio/ksuid"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

var inboundPath string

// The Backend implements SMTP server methodserver.
type Backend struct{}

// A Session is returned after successful login.
type Session struct {
	Timestamp time.Time `json:"timestamp"`
	Client    string    `json:"client"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Content   string    `json:"data"`
}

// Login handles a login command with username and password.
func (bkd *Backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	if username != "username" || password != "password" {
		return nil, errors.New("invalid username or password")
	}
	return &Session{Client: state.RemoteAddr.String()}, nil
}

// AnonymousLogin is allowed
func (bkd *Backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return &Session{Client: state.RemoteAddr.String()}, nil
	//return nil, errors.New("Not Today ...")
}

func (s *Session) Mail(from string, opts smtp.MailOptions) error {

	log.Printf("Reciever - Mail from '%s'\n", from)
	s.From = from
	return nil
}

func (s *Session) Rcpt(to string) error {
	log.Printf("Reciever - To '%s'\n", to)
	s.To = to
	return nil
}

func (s *Session) Data(r io.Reader) error {
	if b, err := ioutil.ReadAll(r); err != nil {
		return err
	} else {
		log.Printf("Reciever - %d byte(s) of data\n", len(b))
		s.Content = string(b)
		s.Timestamp = time.Now()

		file, err := json.MarshalIndent(s, "", " ")
		if err != nil {
			return err
		}

		filePath := path.Join(inboundPath, ksuid.New().String()+".json")
		return ioutil.WriteFile(filePath, file, 0644)
	}
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	log.Println("Reciever - Logout")
	return nil
}

func initDataStructure(dataPath string) {

	inboundPath = path.Join(dataPath, "inbound")

	if _, err := os.Stat(inboundPath); os.IsNotExist(err) {
		err := os.MkdirAll(inboundPath, 0750)

		if err == nil {
			log.Printf("Reciever - Created directory '%s'\n", inboundPath)
		} else {
			log.Fatalf("Reciever - Failed to creat directory '%s'. Error: %s\n", inboundPath, err.Error())
		}
	}
}

func Start(config *configuration.Config) {

	initDataStructure(config.DataPath)

	go func() {
		be := &Backend{}

		server := smtp.NewServer(be)

		server.Addr = config.Receiver.ListenOn
		server.Domain = config.Receiver.ListenOn
		server.ReadTimeout = time.Duration(config.Receiver.ReadTimeout) * time.Second
		server.WriteTimeout = time.Duration(config.Receiver.WriteTimeout) * time.Second
		server.MaxMessageBytes = config.Receiver.MaxMessageBytes
		server.MaxRecipients = config.Receiver.MaxMessageBytes
		server.AllowInsecureAuth = config.Receiver.AllowInsecureAuth

		log.Println("Reciever - Starting server at", server.Addr)

		log.Fatalf("Reciever - %s\n", server.ListenAndServe())

	}()
}
