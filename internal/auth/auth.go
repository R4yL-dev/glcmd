package auth

import (
	"encoding/json"
	"fmt"
)

func NewAuth(data []byte) (*Auth, error) {
	var a Auth

	if err := a.parseJSON(data); err != nil {
		return nil, err
	}

	return &a, nil
}

func (a *Auth) UserID() string {
	return a.userID
}

func (a *Auth) PatientID() string {
	return a.patientID
}
func (a *Auth) SetPatientID(patientID string) error {
	if patientID == "" {
		return fmt.Errorf("patientID cannot be empty")
	}

	a.patientID = patientID

	return nil
}

func (a *Auth) Ticket() *authTicket {
	return a.ticket
}

func (a *Auth) Validate() error {
	if a.userID == "" {
		return fmt.Errorf("userID cannot be empty")
	}

	if a.ticket == nil {
		return fmt.Errorf("ticket cannot be empty")
	}

	if err := a.ticket.Validate(); err != nil {
		return fmt.Errorf("invalid auth ticket: %w", err)
	}

	if a.patientID == "" {
		return fmt.Errorf("patientID cannot be empty")
	}

	return nil
}
func (a *Auth) IsAuth() bool {
	if a == nil {
		return false
	}
	return a.Validate() == nil
}

func (a *Auth) ToJSON() ([]byte, error) {
	ticketJSON, err := a.ticket.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize auth ticket: %w", err)
	}

	var tmp struct {
		UserID     string          `json:"userID"`
		AuthTicket json.RawMessage `json:"authTicket"`
	}

	tmp.UserID = a.userID
	tmp.AuthTicket = ticketJSON

	return json.Marshal(tmp)
}

func (a *Auth) parseJSON(data []byte) error {
	var tmp struct {
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
			AuthTicket struct {
				Token    string `json:"token"`
				Expires  int64  `json:"expires"`
				Duration int64  `json:"duration"`
			} `json:"authTicket"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	ticket, err := NewAuthTicketFromRawValues(
		tmp.Data.AuthTicket.Token,
		tmp.Data.AuthTicket.Expires,
		tmp.Data.AuthTicket.Duration,
	)
	if err != nil {
		return err
	}

	a.userID = tmp.Data.User.ID
	a.ticket = ticket

	return nil
}
