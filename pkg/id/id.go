package id

import (
	"github.com/oklog/ulid/v2"
)

type prefix uint8

const (
	Event prefix = iota
	Ticket
	User
	Session
	Request
)

var prefixes = map[prefix]string{
	Request: "req",
	Event:   "evt",
	Ticket:  "tkt",
	User:    "usr",
}

func New(prefix prefix) string {
	return prefixes[prefix] + "_" + ulid.Make().String()
}
