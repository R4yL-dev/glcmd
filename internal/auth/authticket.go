package auth

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func NewAuthTicket(data []byte) (*authTicket, error) {
	var ticket authTicket

	if err := ticket.parseJSON(data); err != nil {
		return nil, err
	}

	return &ticket, nil
}

func NewAuthTicketFromValues(token string, expires time.Time, duration time.Duration) (*authTicket, error) {
	var ticket authTicket

	if err := ticket.SetToken(token); err != nil {
		return nil, err
	}

	if err := ticket.SetExpires(expires); err != nil {
		return nil, err
	}

	if err := ticket.SetDuration(duration); err != nil {
		return nil, err
	}

	return &ticket, nil
}

func NewAuthTicketFromRawValues(token string, expires int64, duration int64) (*authTicket, error) {
	return NewAuthTicketFromValues(
		token,
		time.Unix(expires, 0),
		time.Duration(duration)*time.Millisecond,
	)
}

func (a *authTicket) Token() string {
	return a.token
}
func (a *authTicket) SetToken(t string) error {
	if t == "" {
		return fmt.Errorf("token cannot be empty")
	}

	parts := strings.Split(t, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT format")
	}

	a.token = t

	return nil
}

func (a *authTicket) Expires() time.Time {
	return a.expires
}
func (a *authTicket) SetExpires(e time.Time) error {
	if e.IsZero() {
		return fmt.Errorf("expiration time cannot be zero")
	}

	if e.Before(time.Now()) {
		return fmt.Errorf("expiration time cannot be in the past")
	}

	a.expires = e

	return nil
}

func (a *authTicket) Duration() time.Duration {
	return a.duration
}
func (a *authTicket) SetDuration(d time.Duration) error {
	if d <= 0 {
		return fmt.Errorf("duration must be positive")
	}

	a.duration = d

	return nil
}

func (a *authTicket) Validate() error {
	if a.token == "" {
		return fmt.Errorf("token is empty")
	}

	if time.Now().After(a.expires) {
		return fmt.Errorf("token has expired")
	}

	return nil
}
func (a *authTicket) IsValid() bool {
	return a.Validate() == nil
}

func (a *authTicket) ToJSON() ([]byte, error) {
	return json.Marshal(struct {
		Token    string `json:"token"`
		Expires  int64  `json:"expires"`
		Duration int64  `json:"duration"`
	}{
		Token:    a.token,
		Expires:  a.expires.Unix(),
		Duration: int64(a.duration / time.Millisecond),
	})
}

func (a *authTicket) parseJSON(data []byte) error {
	var tmp struct {
		Token    string `json:"token"`
		Expires  int64  `json:"expires"`
		Duration int64  `json:"duration"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if err := a.SetToken(tmp.Token); err != nil {
		return err
	}

	if err := a.SetExpires(time.Unix(tmp.Expires, 0)); err != nil {
		return err
	}

	if err := a.SetDuration(time.Duration(tmp.Duration) * time.Millisecond); err != nil {
		return err
	}

	return nil
}
