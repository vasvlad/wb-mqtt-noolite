package main

import "github.com/contactless/wb-mqtt-noolite/noolite"

type NooliteConnection interface {
	Write(req *noolite.Request) error
	Read() (*noolite.Response, error)
	Close()
}
