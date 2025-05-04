package auth

import "time"

type Auth struct {
	userID    string
	patientID string
	ticket    *authTicket
}

type authTicket struct {
	token    string
	expires  time.Time
	duration time.Duration
}
