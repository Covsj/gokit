package email

import (
	"fmt"
	"testing"
)

func TestWxTemp(t *testing.T) {
	email := NewConfig("", "", "", "")
	wxEmail, err := email.ConfigureEmail(false, "fdsafasfsd")
	if err != nil {
		panic(err)
	}
	fmt.Println(wxEmail)
	emails, err := email.FetchEmails("", "")
	if err != nil {
		panic(err)
	}
	fmt.Println(emails)
}
