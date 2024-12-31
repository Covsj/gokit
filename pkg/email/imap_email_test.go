package email

import (
	"fmt"
	"testing"
)

func TestImap(t *testing.T) {
	conn, err := ConnectToIMAP("", "")
	if err != nil {
		panic(err)
	}
	emails, err := FetchLatestIMAPEmails(conn, 10)
	if err != nil {
		panic(err)
	}
	for _, email := range emails {
		fmt.Println(email.TimeStamp, email.From, email.To, email.Subject)
	}

}
