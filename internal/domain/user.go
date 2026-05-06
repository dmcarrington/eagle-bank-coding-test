package domain

import "time"

type Address struct {
	Line1    string
	Line2    string
	Line3    string
	Town     string
	County   string
	Postcode string
}

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
	PhoneNumber  string
	Address      Address
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
