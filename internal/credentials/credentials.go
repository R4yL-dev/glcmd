package credentials

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func NewCredentials(email string, password string) (*Credentials, error) {
	var newCred Credentials

	if err := newCred.SetEmail(email); err != nil {

		return nil, err
	}

	if err := newCred.SetPassword(password); err != nil {
		return nil, err
	}

	return &newCred, nil
}

func (c *Credentials) Email() string {
	return c.email
}
func (c *Credentials) SetEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if !isValidEmail(email) {
		return fmt.Errorf("invalid emial format")
	}

	c.email = email

	return nil
}

func (c *Credentials) Password() string {
	return c.password
}
func (c *Credentials) SetPassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	c.password = password

	return nil
}

func (c *Credentials) ToJSON() ([]byte, error) {
	type tmp struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return json.Marshal(&tmp{
		Email:    c.email,
		Password: c.password,
	})
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}
