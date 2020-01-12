package sender

import (
	"log"
	"strings"

	"github.com/emersion/go-smtp"
)

func Send() {

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{"recipient@example.net"}
	msg := strings.NewReader("To: recipient@example.net\r\n" +
		"Subject: discount Gophers!\r\n" +
		"\r\n" +
		"This is the email body.\r\n")
	err := smtp.SendMail("127.0.0.1:1025", nil, "sender@example.org", to, msg)
	if err != nil {
		log.Fatal(err)
	}
}