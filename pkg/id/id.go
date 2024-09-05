package id

import (
	"github.com/oklog/ulid/v2"
)

type prefix uint8

const (
	Request = iota
	User
	Session
)

var prefixes = map[prefix]string{
	Request: "req",
	User:    "usr",
	Session: "ses",
}

func New(prefix prefix) string {
	return prefixes[prefix] + "_" + ulid.Make().String()
}
